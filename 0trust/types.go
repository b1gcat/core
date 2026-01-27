package main

import (
	"net"
)

const (
	configFile = "0trust.conf"
	ipsetName  = "trustIP"
	bufferSize = 1024
	timeout    = 5
)

type Config struct {
	TrustedIPs []net.IP `json:"trusted_ips"`
	TCPMSS     int      `json:"tcp_mss,omitempty"`
}

type AuthMessage struct {
	Nonce     []byte `json:"nonce"`
	Signature []byte `json:"signature"`
}
