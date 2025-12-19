package pki

import (
	"bytes"
	"crypto/cipher"
	"encoding/binary"
	"errors"
)

const (
	xteaBlockSize = 8
	xteaNumRounds = 32
	xteaDelta     = 0x9E3779B9
)

// XTEA implements the XTEA encryption algorithm
// XTEA uses a 128-bit key and encrypts 64-bit blocks

type XTEA struct {
	key [4]uint32
}

// NewXTEA creates a new XTEA cipher with the given 16-byte key
func NewXTEA(key []byte) (*XTEA, error) {
	if len(key) != 16 {
		return nil, errors.New("xtea: key must be 16 bytes long")
	}

	tea := &XTEA{}
	tea.key[0] = binary.BigEndian.Uint32(key[0:4])
	tea.key[1] = binary.BigEndian.Uint32(key[4:8])
	tea.key[2] = binary.BigEndian.Uint32(key[8:12])
	tea.key[3] = binary.BigEndian.Uint32(key[12:16])

	return tea, nil
}

// BlockSize returns the XTEA block size in bytes
func (x *XTEA) BlockSize() int {
	return xteaBlockSize
}

// Encrypt encrypts a single 64-bit block
func (x *XTEA) Encrypt(dst, src []byte) {
	if len(src) < xteaBlockSize {
		panic("xtea: input block too short")
	}
	if len(dst) < xteaBlockSize {
		panic("xtea: output block too short")
	}

	v0 := binary.BigEndian.Uint32(src[0:4])
	v1 := binary.BigEndian.Uint32(src[4:8])
	key := x.key
	var sum uint32 = 0

	for i := 0; i < xteaNumRounds; i++ {
		v0 += (((v1 << 4) ^ (v1 >> 5)) + v1) ^ (sum + key[sum&3])
		sum += xteaDelta
		v1 += (((v0 << 4) ^ (v0 >> 5)) + v0) ^ (sum + key[(sum>>11)&3])
	}

	binary.BigEndian.PutUint32(dst[0:4], v0)
	binary.BigEndian.PutUint32(dst[4:8], v1)
}

// Decrypt decrypts a single 64-bit block
func (x *XTEA) Decrypt(dst, src []byte) {
	if len(src) < xteaBlockSize {
		panic("xtea: input block too short")
	}
	if len(dst) < xteaBlockSize {
		panic("xtea: output block too short")
	}

	v0 := binary.BigEndian.Uint32(src[0:4])
	v1 := binary.BigEndian.Uint32(src[4:8])
	key := x.key
	// XTEA decryption starts with sum = delta * rounds, but we need to use the correct uint32 value
	// For delta = 0x9E3779B9 and rounds = 32, the value is 0xC6EF3720
	var sum uint32 = 0xC6EF3720

	for i := 0; i < xteaNumRounds; i++ {
		v1 -= (((v0 << 4) ^ (v0 >> 5)) + v0) ^ (sum + key[(sum>>11)&3])
		sum -= xteaDelta
		v0 -= (((v1 << 4) ^ (v1 >> 5)) + v1) ^ (sum + key[sum&3])
	}

	binary.BigEndian.PutUint32(dst[0:4], v0)
	binary.BigEndian.PutUint32(dst[4:8], v1)
}

// EncryptCFB encrypts data using CFB mode with XTEA
func (x *XTEA) EncryptCFB(plaintext []byte) ([]byte, error) {
	// CFB mode requires a nonce/IV
	iv := make([]byte, xteaBlockSize)
	for i := range iv {
		iv[i] = 0 // Zero IV for simplicity, in production use a random IV
	}

	stream := cipher.NewCFBEncrypter(x, iv)
	ciphertext := make([]byte, len(plaintext))
	stream.XORKeyStream(ciphertext, plaintext)

	return ciphertext, nil
}

// DecryptCFB decrypts data using CFB mode with XTEA
func (x *XTEA) DecryptCFB(ciphertext []byte) ([]byte, error) {
	// CFB mode requires a nonce/IV
	iv := make([]byte, xteaBlockSize)
	for i := range iv {
		iv[i] = 0 // Zero IV for simplicity, in production use a random IV
	}

	stream := cipher.NewCFBDecrypter(x, iv)
	plaintext := make([]byte, len(ciphertext))
	stream.XORKeyStream(plaintext, ciphertext)

	return plaintext, nil
}

// Pad pads data to a multiple of block size using PKCS#7
func pad(data []byte, blockSize int) []byte {
	padding := blockSize - (len(data) % blockSize)
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padtext...)
}

// Unpad removes PKCS#7 padding from data
func unpad(data []byte, blockSize int) ([]byte, error) {
	if len(data) == 0 {
		return nil, errors.New("xtea: unpad: empty data")
	}

	padding := int(data[len(data)-1])
	if padding > blockSize || padding == 0 {
		return nil, errors.New("xtea: unpad: invalid padding")
	}

	for i := 0; i < padding; i++ {
		if data[len(data)-1-i] != byte(padding) {
			return nil, errors.New("xtea: unpad: invalid padding")
		}
	}

	return data[:len(data)-padding], nil
}

// Encrypt encrypts data using XTEA with CBC mode and PKCS#7 padding
func Encrypt(key []byte, data []byte) ([]byte, error) {
	xtea, err := NewXTEA(key)
	if err != nil {
		return nil, err
	}

	// Pad data to block size
	paddedData := pad(data, xteaBlockSize)

	// CBC mode requires an IV
	iv := make([]byte, xteaBlockSize)
	for i := range iv {
		iv[i] = 0 // Zero IV for simplicity
	}

	mode := cipher.NewCBCEncrypter(xtea, iv)
	ciphertext := make([]byte, len(paddedData))
	mode.CryptBlocks(ciphertext, paddedData)

	return ciphertext, nil
}

// Decrypt decrypts data using XTEA with CBC mode and PKCS#7 padding
func Decrypt(key []byte, ciphertext []byte) ([]byte, error) {
	xtea, err := NewXTEA(key)
	if err != nil {
		return nil, err
	}

	// CBC mode requires an IV
	iv := make([]byte, xteaBlockSize)
	for i := range iv {
		iv[i] = 0 // Zero IV for simplicity
	}

	mode := cipher.NewCBCDecrypter(xtea, iv)
	plaintext := make([]byte, len(ciphertext))
	mode.CryptBlocks(plaintext, ciphertext)

	// Unpad data
	unpaddedData, err := unpad(plaintext, xteaBlockSize)
	if err != nil {
		return nil, err
	}

	return unpaddedData, nil
}
