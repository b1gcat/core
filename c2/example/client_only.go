//go:build ignore

package main

import (
	"fmt"
	"time"

	"github.com/b1gcat/core/c2"
)

func main() {
	// Hardcoded parameters
	protocolType := c2.ProtocolType("none") // Options: none, dns, ntp
	domain := "example.com"                 // Domain for DNS protocol
	identifier := "test-client"             // Client identifier
	address := "localhost:9003"             // Server address
	interval := 2 * time.Second             // Probe interval

	// Create client
	options := []c2.Option{
		c2.WithClientKey("1234567890123456"), // 16 characters key
		c2.WithClientAddress(address),
		c2.WithClientIdentifier(identifier),
		c2.WithClientInterval(interval),
		c2.WithClientProtocol(protocolType),
	}

	// Add domain option only for DNS protocol
	if protocolType == c2.ProtocolDNS {
		options = append(options, c2.WithClientDomain(domain))
	}

	client, err := c2.NewClient(options...)
	if err != nil {
		fmt.Printf("Failed to create client: %v\n", err)
		return
	}

	if err := client.Start(); err != nil {
		fmt.Println("Client stopped with error: %v\n", err)
	}
	fmt.Println("Client exited successfully\n")
}
