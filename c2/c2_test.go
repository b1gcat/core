package c2

import (
	"bytes"
	"encoding/gob"
	"testing"
	"time"

	"github.com/b1gcat/core/pki"
)

// TestMessageEncoding tests message encoding and decoding
func TestMessageEncoding(t *testing.T) {
	// Test message
	msg := Message{
		Type:       MessageTypeProbe,
		Identifier: "test-client",
		Payload:    []byte("test-payload"),
	}

	// Encode message
	var buf bytes.Buffer
	en := gob.NewEncoder(&buf)
	if err := en.Encode(msg); err != nil {
		t.Errorf("Failed to encode message: %v", err)
		return
	}

	// Decode message
	var decodedMsg Message
	dec := gob.NewDecoder(&buf)
	if err := dec.Decode(&decodedMsg); err != nil {
		t.Errorf("Failed to decode message: %v", err)
		return
	}

	// Verify decoded message
	if decodedMsg.Type != msg.Type {
		t.Errorf("Expected message type %d, got %d", msg.Type, decodedMsg.Type)
	}
	if decodedMsg.Identifier != msg.Identifier {
		t.Errorf("Expected message identifier %s, got %s", msg.Identifier, decodedMsg.Identifier)
	}
	if !bytes.Equal(decodedMsg.Payload, msg.Payload) {
		t.Errorf("Expected message payload %s, got %s", string(msg.Payload), string(decodedMsg.Payload))
	}
}

// TestEncryption tests XTEA encryption and decryption
func TestEncryption(t *testing.T) {
	// Create a test key
	key := []byte("1234567890123456")

	// Test data
	testData := []byte("test command to execute")

	// Encrypt data
	encrypted, err := pki.Encrypt(key, testData)
	if err != nil {
		t.Errorf("Failed to encrypt data: %v", err)
		return
	}

	// Decrypt data
	decrypted, err := pki.Decrypt(key, encrypted)
	if err != nil {
		t.Errorf("Failed to decrypt data: %v", err)
		return
	}

	// Verify data
	if !bytes.Equal(decrypted, testData) {
		t.Errorf("Expected decrypted data %s, got %s", string(testData), string(decrypted))
	}
}

// TestServerClientInteraction tests basic server-client interaction
func TestServerClientInteraction(t *testing.T) {

	// Create server
	server, err := NewServer(
		WithServerKey("1234567890123456"),
		WithServerAddress("0.0.0.0:9002"),
	)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}
	defer server.Stop()

	// Start server (includes console processing automatically)
	go server.Start()

	// Give server time to start
	time.Sleep(500 * time.Millisecond)

	// Create client
	client, err := NewClient(
		WithClientKey("1234567890123456"),
		WithClientAddress("localhost:9002"),
		WithClientIdentifier("test-client-001"),
		WithClientInterval(1*time.Second),
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer client.Stop()

	// Start client
	go client.Start()

	// Wait for client to send probe
	time.Sleep(2 * time.Second)

	// Check if client is registered on server
	server.clientsMu.RLock()
	clientInfo, exists := server.clients["test-client-001"]
	server.clientsMu.RUnlock()

	if !exists {
		t.Error("Expected client to be registered on server")
		return
	}

	if clientInfo.Identifier != "test-client-001" {
		t.Errorf("Expected client identifier 'test-client-001', got '%s'", clientInfo.Identifier)
	}

	// Give client time to process
	time.Sleep(1 * time.Second)
}

// TestDNSProtocolInteraction tests server-client interaction using DNS protocol obfuscation
func TestDNSProtocolInteraction(t *testing.T) {
	// Create server
	server, err := NewServer(
		WithServerKey("1234567890123456"),
		WithServerAddress("0.0.0.0:9003"),
	)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}
	defer server.Stop()

	// Start server (includes console processing automatically)
	go server.Start()

	// Give server time to start
	time.Sleep(500 * time.Millisecond)

	// Create client with DNS protocol
	client, err := NewClient(
		WithClientKey("1234567890123456"),
		WithClientAddress("localhost:9003"),
		WithClientIdentifier("test-dns-client-001"),
		WithClientInterval(1*time.Second),
		WithClientProtocol(ProtocolDNS),
		WithClientDomain("example.com"),
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer client.Stop()

	// Start client
	go client.Start()

	// Wait for client to send probe with DNS protocol
	time.Sleep(2 * time.Second)

	// Check if client is registered on server with DNS protocol
	server.clientsMu.RLock()
	clientInfo, exists := server.clients["test-dns-client-001"]
	server.clientsMu.RUnlock()

	if !exists {
		t.Error("Expected client to be registered on server")
		return
	}

	if clientInfo.Identifier != "test-dns-client-001" {
		t.Errorf("Expected client identifier 'test-dns-client-001', got '%s'", clientInfo.Identifier)
	}

	if clientInfo.Protocol != ProtocolDNS {
		t.Errorf("Expected client protocol '%s', got '%s'", ProtocolDNS, clientInfo.Protocol)
	}

	// Test command execution over DNS protocol
	command := "echo test-dns-command"
	server.executeCommand("test-dns-client-001", command)

	// Give client time to process
	time.Sleep(1 * time.Second)
}

// TestDNSWrapper tests DNS protocol wrapping and unwrapping functionality
func TestDNSWrapper(t *testing.T) {
	wrapper := NewDNSWrapper("example.com")

	// Test data to wrap
	testData := []byte("test-payload-for-dns")

	// Wrap the data
	wrappedData, err := wrapper.Wrap(testData)
	if err != nil {
		t.Fatalf("Failed to wrap data: %v", err)
	}

	// Verify wrapped data is different from original
	if bytes.Equal(wrappedData, testData) {
		t.Error("Wrapped data should be different from original data")
	}

	// Unwrap the data
	unwrappedData, err := wrapper.Unwrap(wrappedData)
	if err != nil {
		t.Fatalf("Failed to unwrap data: %v", err)
	}

	// Verify unwrapped data matches original
	if !bytes.Equal(unwrappedData, testData) {
		t.Errorf("Expected unwrapped data %v, got %v", testData, unwrappedData)
	}

	// Test IsValid functionality
	if !wrapper.IsValid(wrappedData) {
		t.Error("Valid DNS packet should pass IsValid check")
	}

	// Test invalid data
	if wrapper.IsValid([]byte("invalid")) {
		t.Error("Invalid data should not pass IsValid check")
	}
}
