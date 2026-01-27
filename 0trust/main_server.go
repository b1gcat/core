// go:build ignore

package main

import (
	"log"
	"os"
	"strconv"
)

func main() {
	if len(os.Args) != 4 && len(os.Args) != 5 {
		log.Println("Usage: 0trust-server <udp-port> <protect-port> <secret> [tcp-mss]")
		log.Println("Example: 0trust-server 12345 8080 mysecret 1300")
		os.Exit(1)
	}

	udpPort := os.Args[1]
	protectPort := os.Args[2]
	secret := []byte(os.Args[3])
	tcpMSS := 0

	if len(os.Args) == 5 {
		var err error
		tcpMSS, err = strconv.Atoi(os.Args[4])
		if err != nil || tcpMSS <= 0 {
			log.Println("Error: tcp-mss must be a positive integer")
			os.Exit(1)
		}
	}

	server := NewServer(udpPort, protectPort, secret, tcpMSS)
	if err := server.Init(); err != nil {
		log.Fatalf("server init failed: %v", err)
	}

	if err := server.Start(); err != nil {
		log.Fatalf("server start failed: %v", err)
	}
}
