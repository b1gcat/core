package pki

import (
	"testing"
)

func TestXTEAEncryptDecrypt(t *testing.T) {
	// Test key (16 bytes)
	key := []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f}
	
	// Test data
	testData := []byte("Hello, XTEA!")
	
	// Encrypt data
	encrypted, err := Encrypt(key, testData)
	if err != nil {
		t.Fatalf("Failed to encrypt data: %v", err)
	}
	
	// Ensure encrypted data is different from original
	if string(encrypted) == string(testData) {
		t.Error("Encrypted data should be different from original data")
	}
	
	// Decrypt data
	decrypted, err := Decrypt(key, encrypted)
	if err != nil {
		t.Fatalf("Failed to decrypt data: %v", err)
	}
	
	// Ensure decrypted data matches original
	if string(decrypted) != string(testData) {
		t.Errorf("Decrypted data does not match original data. Expected: %q, Got: %q", testData, decrypted)
	}
}

func TestXTEAInvalidKey(t *testing.T) {
	// Test data
	testData := []byte("Hello, XTEA!")
	
	// Test with invalid key length (less than 16 bytes)
	invalidKey := []byte{0x00, 0x01, 0x02, 0x03}
	_, err := Encrypt(invalidKey, testData)
	if err == nil {
		t.Error("Encrypt should return error for invalid key length")
	}
	
	// Test with invalid key length (more than 16 bytes)
	invalidKey2 := []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10}
	_, err = Encrypt(invalidKey2, testData)
	if err == nil {
		t.Error("Encrypt should return error for invalid key length")
	}
}

func TestXTEAEmptyData(t *testing.T) {
	// Test key (16 bytes)
	key := []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f}
	
	// Test with empty data
	emptyData := []byte{}
	
	// Encrypt empty data
	encrypted, err := Encrypt(key, emptyData)
	if err != nil {
		t.Fatalf("Failed to encrypt empty data: %v", err)
	}
	
	// Decrypt empty data
	decrypted, err := Decrypt(key, encrypted)
	if err != nil {
		t.Fatalf("Failed to decrypt empty data: %v", err)
	}
	
	// Ensure decrypted data is empty
	if len(decrypted) != 0 {
		t.Errorf("Decrypted data should be empty. Got: %q", decrypted)
	}
}
