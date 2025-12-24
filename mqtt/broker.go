package mqtt

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	mqtt "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/hooks/auth"
	"github.com/mochi-mqtt/server/v2/listeners"
	"github.com/mochi-mqtt/server/v2/packets"
)

// Options holds the configuration for the broker
type Options struct {
	Address string
	Port    int
	Context context.Context
}

// Option defines the function type for configuring the broker
type Option func(*Options)

// Broker represents the MQTT broker instance
type Broker struct {
	*mqtt.Server
	config       *Options
	Policy       *Policy
	Clients      map[string]*ClientInfo
	clientsMutex sync.RWMutex
	ctx          context.Context
}

// customHook implements the mqtt.Hook interface to listen for client events
type customHook struct {
	mqtt.HookBase
	broker *Broker
}

// ID returns the hook ID
func (h *customHook) ID() string {
	return "custom-broker-hook"
}

// Provides returns the hook capabilities
func (h *customHook) Provides(b byte) bool {
	return b == mqtt.OnConnect || b == mqtt.OnDisconnect || b == mqtt.OnPublish
}

// OnConnect handles client connect events
func (h *customHook) OnConnect(client *mqtt.Client, packet packets.Packet) error {
	clientID := client.ID
	ipAddress := client.Net.Remote

	// Update client information
	clientInfo := &ClientInfo{
		ClientID:   clientID,
		IPAddress:  ipAddress,
		Connected:  true,
		LastSeen:   time.Now().Unix(),
		Metadata:   make(map[string]any),
		PolicyInfo: nil,
	}

	h.broker.clientsMutex.Lock()
	h.broker.Clients[clientID] = clientInfo
	h.broker.clientsMutex.Unlock()

	fmt.Printf("Client connected: %s (%s)\n", clientID, ipAddress)

	// If policy exists, push it to the newly connected client (条件1：哪个客户端连上就推送)
	if h.broker.Policy != nil {
		go func() {
			if err := h.broker.PushPolicyToClient(clientID); err != nil {
				fmt.Printf("Failed to push policy to client %s: %v\n", clientID, err)
			}
		}()
	}
	return nil
}

// OnDisconnect handles client disconnect events
func (h *customHook) OnDisconnect(client *mqtt.Client, err error, expire bool) {
	clientID := client.ID

	h.broker.clientsMutex.Lock()
	if clientInfo, exists := h.broker.Clients[clientID]; exists {
		clientInfo.Connected = false
		clientInfo.LastSeen = time.Now().Unix()
		h.broker.Clients[clientID] = clientInfo
	}
	h.broker.clientsMutex.Unlock()

	fmt.Printf("Client disconnected: %s\n", clientID)
}

// OnPublish handles publish events (for client info collection)
func (h *customHook) OnPublish(client *mqtt.Client, packet packets.Packet) (packets.Packet, error) {
	// Check if this is a client info message
	if packet.TopicName == TopicClientInfo {
		var clientInfo ClientInfo
		if err := json.Unmarshal(packet.Payload, &clientInfo); err != nil {
			fmt.Printf("Failed to unmarshal client info: %v\n", err)
		}

		h.broker.clientsMutex.Lock()
		defer h.broker.clientsMutex.Unlock()

		// Update client information
		h.broker.Clients[clientInfo.ClientID] = &clientInfo
	}
	return packet, nil
}

// WithAddress sets the address for the broker
func WithAddress(addr string) Option {
	return func(o *Options) {
		o.Address = addr
	}
}

// WithPort sets the port for the broker
func WithPort(port int) Option {
	return func(o *Options) {
		o.Port = port
	}
}

// WithContext sets the context for the broker
func WithContext(ctx context.Context) Option {
	return func(o *Options) {
		o.Context = ctx
	}
}

// NewBroker creates a new MQTT broker instance using the with-options pattern
func NewBroker(opts ...Option) (*Broker, error) {
	// Default options
	options := &Options{
		Address: "0.0.0.0",
		Port:    1883,
		Context: context.Background(),
	}

	// Apply provided options
	for _, opt := range opts {
		opt(options)
	}

	// Get the context from options
	brokerCtx := options.Context

	s := mqtt.New(&mqtt.Options{
		// Default options
	})

	// Add anonymous authentication hook
	err := s.AddHook(new(auth.AllowHook), nil)
	if err != nil {
		return nil, err
	}

	broker := &Broker{
		Server:  s,
		config:  options,
		Clients: make(map[string]*ClientInfo),
		ctx:     brokerCtx,
	}

	// Add custom hook for client events
	customHook := &customHook{
		broker: broker,
	}

	err = s.AddHook(customHook, nil)
	if err != nil {
		return nil, err
	}

	return broker, nil
}

// Start starts the MQTT broker
func (b *Broker) Start() error {
	listenerID := "tcp"
	address := fmt.Sprintf("%s:%d", b.config.Address, b.config.Port)

	tcp := listeners.NewTCP(listeners.Config{
		ID:      listenerID,
		Address: address,
	})

	if err := b.AddListener(tcp); err != nil {
		return fmt.Errorf("failed to add listener: %w", err)
	}

	fmt.Printf("MQTT broker started on %s\n", address)

	// Start the server in a goroutine
	go func() {
		if err := b.Serve(); err != nil {
			fmt.Printf("Broker serve error: %v\n", err)
		}
	}()

	// Wait for context to be done
	<-b.ctx.Done()

	// Stop the server
	if err := b.Close(); err != nil {
		return fmt.Errorf("failed to stop broker: %w", err)
	}

	fmt.Println("MQTT broker stopped")
	return nil
}

// CreatePolicy creates a new policy
func (b *Broker) CreatePolicy(name, description string, settings map[string]any) *Policy {
	policy := &Policy{
		ID:          fmt.Sprintf("policy-%s", uuid.New().String()),
		Name:        name,
		Description: description,
		Settings:    settings,
		Timestamp:   time.Now().Unix(),
	}

	return policy
}

// SetPolicy sets the current policy and pushes it to all clients
func (b *Broker) SetPolicy(policy *Policy) error {
	b.Policy = policy

	// Push to all clients (主动触发条件)
	return b.PushPolicyToAllClients()
}

// PushPolicyToClient pushes the current policy to a specific client
func (b *Broker) PushPolicyToClient(clientID string) error {
	if b.Policy == nil {
		return fmt.Errorf("no policy set")
	}

	// Marshal policy to JSON
	payload, err := json.Marshal(b.Policy)
	if err != nil {
		return fmt.Errorf("failed to marshal policy: %w", err)
	}

	// Create a publish packet
	packet := packets.Packet{
		FixedHeader: packets.FixedHeader{
			Type:   packets.Publish,
			Qos:    0, // 使用QoS 0避免需要packet id
			Retain: true,
		},
		TopicName: TopicPolicy,
		Payload:   payload,
	}

	// Get client by ID
	client, exists := b.Server.Clients.Get(clientID)
	if !exists {
		return fmt.Errorf("client not found: %s", clientID)
	}

	// Send the packet to the specific client
	if err := b.Server.InjectPacket(client, packet); err != nil {
		return fmt.Errorf("failed to send policy to client: %w", err)
	}

	fmt.Printf("Pushed policy %s to client %s\n", b.Policy.Name, clientID)
	return nil
}

// PushPolicyToAllClients pushes the current policy to all connected clients
func (b *Broker) PushPolicyToAllClients() error {
	if b.Policy == nil {
		return fmt.Errorf("no policy set")
	}

	// Marshal policy to JSON
	payload, err := json.Marshal(b.Policy)
	if err != nil {
		return fmt.Errorf("failed to marshal policy: %w", err)
	}

	// Create a publish packet
	packet := packets.Packet{
		FixedHeader: packets.FixedHeader{
			Type:   packets.Publish,
			Qos:    0, // 使用QoS 0避免需要packet id
			Retain: true,
		},
		TopicName: TopicPolicy,
		Payload:   payload,
	}

	// Get all connected clients and send the policy
	b.clientsMutex.RLock()
	for clientID, clientInfo := range b.Clients {
		if clientInfo.Connected {
			// Get client by ID
			client, exists := b.Server.Clients.Get(clientID)
			if exists {
				// Send to each connected client
				if err := b.Server.InjectPacket(client, packet); err != nil {
					fmt.Printf("Failed to send policy to client %s: %v\n", clientID, err)
					// Continue with other clients
				}
			}
		}
	}
	b.clientsMutex.RUnlock()

	fmt.Printf("Pushed policy %s to all connected clients\n", b.Policy.Name)
	return nil
}

// GetConnectedClients returns the number of connected clients
func (b *Broker) GetConnectedClients() int {
	b.clientsMutex.RLock()
	defer b.clientsMutex.RUnlock()

	count := 0
	for _, clientInfo := range b.Clients {
		if clientInfo.Connected {
			count++
		}
	}

	return count
}

// GetClientInfo returns information about a specific client
func (b *Broker) GetClientInfo(clientID string) (*ClientInfo, bool) {
	b.clientsMutex.RLock()
	defer b.clientsMutex.RUnlock()

	clientInfo, exists := b.Clients[clientID]
	return clientInfo, exists
}

// GetAllClients returns information about all clients
func (b *Broker) GetAllClients() []*ClientInfo {
	b.clientsMutex.RLock()
	defer b.clientsMutex.RUnlock()

	clients := make([]*ClientInfo, 0, len(b.Clients))
	for _, clientInfo := range b.Clients {
		clients = append(clients, clientInfo)
	}

	return clients
}

// SendCommandToClient sends a command to a specific client
func (b *Broker) SendCommandToClient(clientID string, cmd *SystemCommand) error {
	// Marshal command to JSON
	payload, err := json.Marshal(cmd)
	if err != nil {
		return fmt.Errorf("failed to marshal command: %w", err)
	}

	// Create a publish packet
	packet := packets.Packet{
		FixedHeader: packets.FixedHeader{
			Type:   packets.Publish,
			Qos:    0, // 使用QoS 0避免需要packet id
			Retain: false,
		},
		TopicName: TopicSystemCmd,
		Payload:   payload,
	}

	// Get client by ID
	client, exists := b.Server.Clients.Get(clientID)
	if !exists {
		return fmt.Errorf("client not found: %s", clientID)
	}

	// Send the packet to the specific client
	if err := b.Server.InjectPacket(client, packet); err != nil {
		return fmt.Errorf("failed to send command to client: %w", err)
	}

	fmt.Printf("Sent command %s to client %s\n", cmd.Type, clientID)
	return nil
}

// SendCommandToAllClients sends a command to all connected clients
func (b *Broker) SendCommandToAllClients(cmd *SystemCommand) error {
	// Marshal command to JSON
	payload, err := json.Marshal(cmd)
	if err != nil {
		return fmt.Errorf("failed to marshal command: %w", err)
	}

	// Create a publish packet
	packet := packets.Packet{
		FixedHeader: packets.FixedHeader{
			Type:   packets.Publish,
			Qos:    0, // 使用QoS 0避免需要packet id
			Retain: false,
		},
		TopicName: TopicSystemCmd,
		Payload:   payload,
	}

	// Get all connected clients and send the command
	b.clientsMutex.RLock()
	for clientID, clientInfo := range b.Clients {
		if clientInfo.Connected {
			// Get client by ID
			client, exists := b.Server.Clients.Get(clientID)
			if exists {
				// Send to each connected client
				if err := b.Server.InjectPacket(client, packet); err != nil {
					fmt.Printf("Failed to send command to client %s: %v\n", clientID, err)
					// Continue with other clients
				}
			}
		}
	}
	b.clientsMutex.RUnlock()

	fmt.Printf("Sent command %s to all clients\n", cmd.Type)
	return nil
}
