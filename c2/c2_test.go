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
