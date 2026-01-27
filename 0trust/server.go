package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os/exec"
	"strconv"
)

type Server struct {
	udpPort     string
	protectPort string
	secret      []byte
	tcpMSS      int
	config      *Config
}

func NewServer(udpPort, protectPort string, secret []byte, tcpMSS int) *Server {
	return &Server{
		udpPort:     udpPort,
		protectPort: protectPort,
		secret:      secret,
		tcpMSS:      tcpMSS,
	}
}

func (s *Server) Init() error {
	var err error
	s.config, err = loadConfig()
	if err != nil {
		return err
	}

	if err := s.initIPSet(); err != nil {
		return err
	}

	if err := s.loadTrustedIPs(); err != nil {
		return err
	}

	if err := s.setupFirewall(); err != nil {
		return err
	}

	return nil
}

func (s *Server) initIPSet() error {
	cmd := exec.Command("ipset", "list", ipsetName)
	if err := cmd.Run(); err != nil {
		cmd = exec.Command("ipset", "create", ipsetName, "hash:ip")
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to create ipset: %v", err)
		}
	}
	return nil
}

func (s *Server) loadTrustedIPs() error {
	for _, ip := range s.config.TrustedIPs {
		if err := s.addIPToIPSet(ip); err != nil {
			log.Printf("failed to add IP %s to ipset: %v", ipToString(ip), err)
		}
	}
	return nil
}

func (s *Server) setupFirewall() error {
	cmd := exec.Command("iptables", "-C", "INPUT", "-p", "tcp", "--dport", s.protectPort, "-m", "set", "--match-set", ipsetName, "src", "-j", "ACCEPT")
	if err := cmd.Run(); err != nil {
		cmd = exec.Command("iptables", "-A", "INPUT", "-p", "tcp", "--dport", s.protectPort, "-m", "set", "--match-set", ipsetName, "src", "-j", "ACCEPT")
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to setup iptables rule: %v", err)
		}
	}

	cmd = exec.Command("iptables", "-C", "INPUT", "-p", "tcp", "--dport", s.protectPort, "-j", "REJECT")
	if err := cmd.Run(); err != nil {
		cmd = exec.Command("iptables", "-A", "INPUT", "-p", "tcp", "--dport", s.protectPort, "-j", "REJECT")
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to setup iptables drop rule: %v", err)
		}
	}

	if s.tcpMSS > 0 {
		if err := s.setupTCPMSS(); err != nil {
			return fmt.Errorf("failed to setup TCP MSS: %v", err)
		}
	}

	return nil
}

func (s *Server) setupTCPMSS() error {
	// 清除可能存在的旧规则
	cmd := exec.Command("iptables", "-t", "mangle", "-D", "POSTROUTING", "-p", "tcp", "--tcp-flags", "SYN,RST", "SYN", "-j", "TCPMSS", "--set-mss", strconv.Itoa(s.tcpMSS))
	cmd.Run() // 忽略错误，因为规则可能不存在

	// 添加新规则
	cmd = exec.Command("iptables", "-t", "mangle", "-A", "POSTROUTING", "-p", "tcp", "--tcp-flags", "SYN,RST", "SYN", "-j", "TCPMSS", "--set-mss", strconv.Itoa(s.tcpMSS))
	if err := cmd.Run(); err != nil {
		return err
	}

	log.Printf("✓ TCP MSS set to %d", s.tcpMSS)
	return nil
}

func (s *Server) addIPToIPSet(ip net.IP) error {
	cmd := exec.Command("ipset", "add", ipsetName, ipToString(ip), "-exist")
	return cmd.Run()
}

func (s *Server) Start() error {
	addr, err := net.ResolveUDPAddr("udp", ":"+s.udpPort)
	if err != nil {
		return err
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return err
	}
	defer conn.Close()

	log.Printf("0trust server started on UDP port %s, protecting TCP port %s", s.udpPort, s.protectPort)

	buffer := make([]byte, bufferSize)
	for {
		n, remoteAddr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			log.Printf("read error: %v", err)
			continue
		}

		go s.handleClient(conn, remoteAddr, buffer[:n])
	}
}

func (s *Server) handleClient(conn *net.UDPConn, remoteAddr *net.UDPAddr, data []byte) {
	var msg AuthMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		log.Printf("invalid message from %s: %v", remoteAddr.IP, err)
		return
	}

	if !verifySignature(msg.Nonce, msg.Signature, s.secret) {
		log.Printf("invalid signature from %s", remoteAddr.IP)
		return
	}

	log.Printf("successful authentication from %s", remoteAddr.IP)

	if addIPToConfig(remoteAddr.IP, s.config) {
		if err := saveConfig(s.config); err != nil {
			log.Printf("failed to save config: %v", err)
		}
	}

	if err := s.addIPToIPSet(remoteAddr.IP); err != nil {
		log.Printf("failed to add IP %s to ipset: %v", remoteAddr.IP, err)
		return
	}

	response := map[string]string{"status": "ok"}
	responseData, _ := json.Marshal(response)
	conn.WriteToUDP(responseData, remoteAddr)
}
