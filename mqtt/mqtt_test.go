package mqtt

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"
)

// TestMQTTIntegration tests the complete MQTT integration
func TestMQTTIntegration(t *testing.T) {
	// Create and start broker using with-options pattern
	broker, err := NewBroker(
		WithAddress("0.0.0.0"),
		WithPort(1883),
	)
	if err != nil {
		t.Fatalf("Failed to create broker: %v", err)
	}

	// Start broker in a goroutine
	brokerDone := make(chan error, 1)
	go func() {
		brokerDone <- broker.Start()
	}()

	// Give broker time to start
	time.Sleep(2 * time.Second)

	// Cleanup function
	defer func() {
		// Stop the broker
		if broker != nil {
			if err := broker.Close(); err != nil {
				t.Logf("Error stopping broker: %v", err)
			}
		}

		// Wait for broker to finish
		select {
		case err := <-brokerDone:
			if err != nil && err.Error() != "context canceled" {
				t.Logf("Broker exited with error: %v", err)
			}
		case <-time.After(2 * time.Second):
			// Timeout
		}
	}()

	// Test basic functionality
	t.Run("BasicBrokerSetup", func(t *testing.T) {
		if broker.GetConnectedClients() != 0 {
			t.Errorf("Expected 0 connected clients, got %d", broker.GetConnectedClients())
		}
	})

	// Test server and client functionality
	t.Run("ServerClientInteraction", func(t *testing.T) {
		testServerClientInteraction(t)
	})

	// Test policy pushing scenarios
	t.Run("PolicyPushing", func(t *testing.T) {
		testPolicyPushing(t, broker)
	})

	// Test client info collection
	t.Run("ClientInfoCollection", func(t *testing.T) {
		testClientInfoCollection(t, broker)
	})
}

// testServerClientInteraction tests server and client basic interaction
func testServerClientInteraction(t *testing.T) {
	// Create client
	clientConfig := &ClientConfig{
		ClientID:   "test-client-1",
		ClientName: "Test Client",
		Metadata: map[string]any{
			"app":     "test",
			"version": "1.0.0",
		},
	}

	client, err := NewClient(clientConfig)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Connect client
	if err := client.Connect(); err != nil {
		t.Fatalf("Failed to connect client: %v", err)
	}
	defer client.Disconnect()

	// Give time for client to connect and send info
	time.Sleep(2 * time.Second)

	// Verify client is connected
	if !client.IsConnected() {
		t.Error("Client should be connected")
	}
}

// testPolicyPushing tests both policy pushing scenarios
func testPolicyPushing(t *testing.T, broker *Broker) {
	// Create a test policy
	testPolicy := broker.CreatePolicy(
		"test-policy",
		"Test Policy Description",
		map[string]interface{}{
			"interval":        30,
			"retry_count":     3,
			"log_level":       "info",
			"max_connections": 100,
		},
	)

	// Set the policy (active trigger - pushes to all clients)
	if err := broker.SetPolicy(testPolicy); err != nil {
		t.Fatalf("Failed to set policy: %v", err)
	}

	// Give time for policy to be processed
	time.Sleep(1 * time.Second)

	// Create client 1 - should receive policy when connecting (scenario 1)
	policyReceived1 := make(chan *Policy, 1)

	client1Config := &ClientConfig{
		ClientID:   "test-client-1",
		ClientName: "Test Client 1",
	}

	client1, err := NewClient(client1Config)
	if err != nil {
		t.Fatalf("Failed to create client 1: %v", err)
	}

	client1.OnPolicyReceived = func(p *Policy) error {
		policyReceived1 <- p
		return nil
	}

	// Connect client 1
	if err := client1.Connect(); err != nil {
		t.Fatalf("Failed to connect client 1: %v", err)
	}

	// Wait for client 1 to receive policy
	select {
	case policy := <-policyReceived1:
		if policy == nil || policy.ID != testPolicy.ID {
			t.Error("Client 1 did not receive correct policy")
		}
		t.Logf("Client 1 received policy: %s", policy.Name)
	case <-time.After(5 * time.Second):
		t.Error("Client 1 timed out waiting for policy")
	}

	// Create client 2 - should receive policy when connecting (scenario 1)
	policyReceived2 := make(chan *Policy, 1)

	client2Config := &ClientConfig{
		ClientID:   "test-client-2",
		ClientName: "Test Client 2",
	}

	client2, err := NewClient(client2Config)
	if err != nil {
		t.Fatalf("Failed to create client 2: %v", err)
	}

	client2.OnPolicyReceived = func(p *Policy) error {
		policyReceived2 <- p
		return nil
	}

	// Connect client 2
	if err := client2.Connect(); err != nil {
		t.Fatalf("Failed to connect client 2: %v", err)
	}

	// Wait for client 2 to receive policy
	select {
	case policy := <-policyReceived2:
		if policy == nil || policy.ID != testPolicy.ID {
			t.Error("Client 2 did not receive correct policy")
		}
		t.Logf("Client 2 received policy: %s", policy.Name)
	case <-time.After(5 * time.Second):
		t.Error("Client 2 timed out waiting for policy")
	}

	// Create a new policy and push to all clients (active trigger - scenario 2)
	newPolicy := broker.CreatePolicy(
		"new-test-policy",
		"New Test Policy Description",
		map[string]interface{}{
			"interval":        60,
			"retry_count":     5,
			"log_level":       "debug",
			"max_connections": 200,
		},
	)

	// Push to all clients
	if err := broker.SetPolicy(newPolicy); err != nil {
		t.Fatalf("Failed to set new policy: %v", err)
	}

	// Wait for both clients to receive new policy
	clientsUpdated := 0
	timeout := time.After(5 * time.Second)

	for clientsUpdated < 2 {
		select {
		case policy := <-policyReceived1:
			if policy != nil && policy.ID == newPolicy.ID {
				clientsUpdated++
				t.Logf("Client 1 received updated policy: %s", policy.Name)
			}
		case policy := <-policyReceived2:
			if policy != nil && policy.ID == newPolicy.ID {
				clientsUpdated++
				t.Logf("Client 2 received updated policy: %s", policy.Name)
			}
		case <-timeout:
			t.Error("Timeout waiting for clients to receive updated policy")
			goto cleanup
		}
	}

cleanup:
	// Disconnect clients
	if client1 != nil {
		client1.Disconnect()
	}
	if client2 != nil {
		client2.Disconnect()
	}
}

// testClientInfoCollection tests client info collection by broker
func testClientInfoCollection(t *testing.T, broker *Broker) {

	// Create test client
	clientConfig := &ClientConfig{
		ClientID:   "info-client",
		ClientName: "Info Client",
		Metadata: map[string]any{
			"app":     "info-collector",
			"version": "1.0.0",
			"os":      "linux",
		},
	}

	client, err := NewClient(clientConfig)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Connect client
	if err := client.Connect(); err != nil {
		t.Fatalf("Failed to connect client: %v", err)
	}
	defer client.Disconnect()

	// Give time for client info to be processed
	time.Sleep(3 * time.Second)

	// Get connected clients count from broker
	clientCount := broker.GetConnectedClients()
	if clientCount != 1 {
		t.Errorf("Expected 1 connected client, got %d", clientCount)
	}

	// Check client info
	clientInfo, exists := broker.GetClientInfo("info-client")
	if !exists {
		t.Error("Client info not found in broker's client list")
		return
	}

	if clientInfo.Metadata["app"] != "info-collector" {
		t.Errorf("Expected app metadata to be 'info-collector', got '%s'", clientInfo.Metadata["app"])
	}
	if clientInfo.Metadata["version"] != "1.0.0" {
		t.Errorf("Expected version metadata to be '1.0.0', got '%s'", clientInfo.Metadata["version"])
	}
}

// ExampleUsage shows a complete usage example
func Example() {
	// This is a complete example showing how to use the MQTT system
	fmt.Println("MQTT System Usage Example")

	// Step 1: Create and start broker using with-options pattern
	broker, err := NewBroker(
		WithAddress("0.0.0.0"),
		WithPort(1883),
	)
	if err != nil {
		fmt.Printf("Failed to create broker: %v\n", err)
		os.Exit(1)
	}

	// Start broker in a goroutine
	go func() {
		if err := broker.Start(); err != nil {
			fmt.Printf("Broker error: %v\n", err)
		}
	}()

	// Give broker time to start
	time.Sleep(2 * time.Second)
	fmt.Println("Broker started")

	// Step 2: Create client with policy handler
	policyChan := make(chan *Policy, 1)

	clientConfig := &ClientConfig{
		ClientID:   "example-client",
		ClientName: "Example Client",
		Metadata: map[string]any{
			"app":     "example-app",
			"version": "1.0.0",
		},
	}

	client, err := NewClient(clientConfig)
	if err != nil {
		fmt.Printf("Failed to create client: %v\n", err)
		os.Exit(1)
	}

	client.OnPolicyReceived = func(p *Policy) error {
		policyChan <- p
		return nil
	}

	// Step 3: Connect client
	if err := client.Connect(); err != nil {
		fmt.Printf("Failed to connect client: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Client connected")

	// Step 4: Create and set test policy (broker handles policy management)
	testPolicy := broker.CreatePolicy(
		"example-policy",
		"Example Policy Description",
		map[string]interface{}{
			"interval":    30,
			"retry_count": 3,
			"log_level":   "info",
		},
	)

	// Set policy (pushes to all connected clients)
	if err := broker.SetPolicy(testPolicy); err != nil {
		fmt.Printf("Failed to set policy: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Policy set and pushed to all clients")

	// Step 5: Wait for client to receive policy
	select {
	case policy := <-policyChan:
		fmt.Printf("Client received policy: %s\n", policy.Name)
		// Print policy details
		if data, err := json.MarshalIndent(policy, "", "  "); err == nil {
			fmt.Println("Policy details:")
			fmt.Println(string(data))
		}
	case <-time.After(5 * time.Second):
		fmt.Println("Timed out waiting for policy")
	}

	// Step 6: Get and display connected clients from broker
	clientInfo, exists := broker.GetClientInfo("example-client")
	if exists && clientInfo != nil {
		fmt.Printf("Connected clients: %d\n", broker.GetConnectedClients())
		fmt.Printf("- %s (Connected: %v)\n", clientInfo.ClientID, clientInfo.Connected)
	}

	// Step 7: Cleanup
	client.Disconnect()
	if err := broker.Close(); err != nil {
		fmt.Printf("Error stopping broker: %v\n", err)
	}

	fmt.Println("Example completed")

}

// Helper function to print policy details
func printPolicy(policy *Policy) {
	if policy == nil {
		fmt.Println("No policy")
		return
	}

	fmt.Printf("Policy: %s\n", policy.Name)
	fmt.Printf("ID: %s\n", policy.ID)
	fmt.Printf("Description: %s\n", policy.Description)
	fmt.Println("Settings:")
	for k, v := range policy.Settings {
		fmt.Printf("  %s: %v\n", k, v)
	}
	fmt.Printf("Timestamp: %d\n", policy.Timestamp)
}
