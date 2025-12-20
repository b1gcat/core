package c2

import (
	"io"
	"net"
	"time"
)

// ProtocolType defines the supported protocol obfuscation types
type ProtocolType string

const (
	// ProtocolNone means no obfuscation
	ProtocolNone ProtocolType = "none"
	// ProtocolDNS means DNS protocol obfuscation
	ProtocolDNS ProtocolType = "dns"
	// ProtocolNTP means NTP protocol obfuscation
	ProtocolNTP ProtocolType = "ntp"
)

// ClientInfo represents a connected client's information
type ClientInfo struct {
	Identifier string       `json:"identifier"`
	SourceIP   net.Addr     `json:"source_ip"`
	LastSeen   time.Time    `json:"last_seen"`
	PendingCmd string       `json:"pending_cmd,omitempty"`
	Protocol   ProtocolType `json:"protocol,omitempty"`
}

// MessageType defines the type of message
type MessageType uint8

const (
	MessageTypeProbe   MessageType = 0x01
	MessageTypeCommand MessageType = 0x02
	MessageTypeResult  MessageType = 0x03
)

// Message represents a UDP message structure
type Message struct {
	Type       MessageType `json:"type"`
	Identifier string      `json:"identifier"`
	Payload    []byte      `json:"payload,omitempty"`
}

// Config defines the configuration for client and server
type Config struct {
	Key        []byte
	Address    string
	Identifier string
	Interval   time.Duration
	Protocol   ProtocolType
	Domain     string
	Logger     io.Writer // Logger to use for output
}

// Option is a function type for configuring client/server
type Option func(*Config)
