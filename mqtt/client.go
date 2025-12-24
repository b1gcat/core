package mqtt

import (
	"encoding/json"
	"fmt"
	"net"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/google/uuid"
)

// Client represents an MQTT client instance
type Client struct {
	mqtt.Client
	Config            *ClientConfig
	Policy            *Policy
	ClientInfo        *ClientInfo
	OnPolicyReceived  func(*Policy) error
	OnCommandReceived func(*SystemCommand) error
}

// ClientConfig contains client configuration
type ClientConfig struct {
	BrokerAddress string
	ClientID      string
	ClientName    string
	ClientType    string
	Metadata      map[string]any
}

// NewClient creates a new MQTT client instance
func NewClient(config *ClientConfig) (*Client, error) {
	if config == nil {
		config = &ClientConfig{
			BrokerAddress: DefaultBrokerAddress,
			ClientID:      fmt.Sprintf("client-%s", uuid.New().String()),
			ClientName:    "mqtt-client",
			ClientType:    "default",
			Metadata:      make(map[string]any),
		}
	}

	if config.ClientID == "" {
		config.ClientID = fmt.Sprintf("client-%s", uuid.New().String())
	}

	if config.Metadata == nil {
		config.Metadata = make(map[string]any)
	}

	// Set default broker address if not provided
	if config.BrokerAddress == "" {
		config.BrokerAddress = DefaultBrokerAddress
	}

	// Set MQTT client options
	options := mqtt.NewClientOptions()
	options.AddBroker(config.BrokerAddress)
	options.SetClientID(config.ClientID)
	options.SetCleanSession(true)
	options.SetAutoReconnect(true)
	options.SetConnectTimeout(5 * time.Second)
	options.SetKeepAlive(30 * time.Second)
	options.SetMaxReconnectInterval(1 * time.Minute)

	// Create client instance
	client := mqtt.NewClient(options)

	// Create client info
	clientInfo := &ClientInfo{
		ClientID:  config.ClientID,
		Connected: false,
		Metadata:  config.Metadata,
		LastSeen:  time.Now().Unix(),
	}

	mqttClient := &Client{
		Client:     client,
		Config:     config,
		ClientInfo: clientInfo,
	}

	return mqttClient, nil
}

// Connect connects the client to the MQTT broker
func (c *Client) Connect() error {
	if token := c.Client.Connect(); token.Wait() && token.Error() != nil {
		return fmt.Errorf("failed to connect to broker: %w", token.Error())
	}

	c.ClientInfo.Connected = true
	c.ClientInfo.LastSeen = time.Now().Unix()
	c.ClientInfo.IPAddress = getLocalIP()

	// Subscribe to policy topic
	if token := c.Client.Subscribe(TopicPolicy, 1, c.onPolicyMessage); token.Wait() && token.Error() != nil {
		return fmt.Errorf("failed to subscribe to policy topic: %w", token.Error())
	}

	// Subscribe to system command topic
	if token := c.Client.Subscribe(TopicSystemCmd, 1, c.onSystemCommand); token.Wait() && token.Error() != nil {
		return fmt.Errorf("failed to subscribe to system command topic: %w", token.Error())
	}

	// Publish initial client info
	if err := c.PublishClientInfo(); err != nil {
		return fmt.Errorf("failed to publish client info: %w", err)
	}

	fmt.Printf("Client %s connected to broker %s\n", c.Config.ClientID, c.Config.BrokerAddress)
	return nil
}

// Disconnect disconnects the client from the MQTT broker
func (c *Client) Disconnect() {
	c.ClientInfo.Connected = false
	c.ClientInfo.LastSeen = time.Now().Unix()
	c.PublishClientInfo()
	c.Client.Disconnect(250)
	fmt.Printf("Client %s disconnected\n", c.Config.ClientID)
}

// PublishClientInfo publishes client information to the broker
func (c *Client) PublishClientInfo() error {
	c.ClientInfo.LastSeen = time.Now().Unix()
	c.ClientInfo.PolicyInfo = c.Policy

	payload, err := json.Marshal(c.ClientInfo)
	if err != nil {
		return fmt.Errorf("failed to marshal client info: %w", err)
	}

	if token := c.Client.Publish(TopicClientInfo, 1, false, payload); token.Wait() && token.Error() != nil {
		return fmt.Errorf("failed to publish client info: %w", token.Error())
	}

	return nil
}

// onPolicyMessage handles incoming policy messages
func (c *Client) onPolicyMessage(client mqtt.Client, msg mqtt.Message) {
	var policy Policy
	if err := json.Unmarshal(msg.Payload(), &policy); err != nil {
		fmt.Printf("Failed to unmarshal policy: %v\n", err)
		return
	}

	c.Policy = &policy

	// Update client info with new policy
	c.ClientInfo.PolicyInfo = &policy

	// Call callback if registered
	if c.OnPolicyReceived != nil {
		if err := c.OnPolicyReceived(&policy); err != nil {
			fmt.Printf("Error handling policy: %v\n", err)
		}
	}

	fmt.Printf("Client %s received policy: %s\n", c.Config.ClientID, policy.Name)
}

// onSystemCommand handles incoming system commands
func (c *Client) onSystemCommand(client mqtt.Client, msg mqtt.Message) {
	var cmd SystemCommand
	if err := json.Unmarshal(msg.Payload(), &cmd); err != nil {
		fmt.Printf("Failed to unmarshal system command: %v\n", err)
		return
	}

	// Call callback if registered
	if c.OnCommandReceived != nil {
		if err := c.OnCommandReceived(&cmd); err != nil {
			fmt.Printf("Error handling system command: %v\n", err)
		}
	}

	fmt.Printf("Client %s received command: %s\n", c.Config.ClientID, cmd.Type)
}

// getLocalIP returns the local IP address of the machine
func getLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "unknown"
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}

	return "127.0.0.1"
}
