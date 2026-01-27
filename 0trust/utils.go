package _0trust

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"net"
	"os"
)

func generateNonce() ([]byte, error) {
	nonce := make([]byte, 16)
	_, err := rand.Read(nonce)
	return nonce, err
}

func signMessage(message []byte, key []byte) []byte {
	h := hmac.New(sha256.New, key)
	h.Write(message)
	return h.Sum(nil)
}

func verifySignature(message, signature, key []byte) bool {
	expected := signMessage(message, key)
	return hmac.Equal(signature, expected)
}

func saveConfig(config *Config) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(configFile, data, 0600)
}

func loadConfig() (*Config, error) {
	var config Config
	data, err := os.ReadFile(configFile)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{TrustedIPs: []net.IP{}}, nil
		}
		return nil, err
	}
	err = json.Unmarshal(data, &config)
	return &config, err
}

func addIPToConfig(ip net.IP, config *Config) bool {
	for _, existing := range config.TrustedIPs {
		if existing.Equal(ip) {
			return false
		}
	}
	config.TrustedIPs = append(config.TrustedIPs, ip)
	return true
}

func ipToString(ip net.IP) string {
	if ip4 := ip.To4(); ip4 != nil {
		return ip4.String()
	}
	return ip.String()
}
