package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/b1gcat/core/c2"
)

func main() {
	// Parse command line arguments
	protocol := flag.String("protocol", "none", "Protocol type: none, dns, ntp")
	domain := flag.String("domain", "example.com", "Domain for DNS protocol")
	identifier := flag.String("id", "test-client", "Client identifier")
	address := flag.String("address", "localhost:9003", "Server address")
	interval := flag.Duration("interval", 2*time.Second, "Probe interval")
	flag.Parse()

	// Validate protocol
	protocolType := c2.ProtocolType(*protocol)
	switch protocolType {
	case c2.ProtocolNone, c2.ProtocolDNS, c2.ProtocolNTP:
		// Valid protocol
	default:
		fmt.Printf("Invalid protocol: %s. Use 'none', 'dns', or 'ntp'.\n", *protocol)
		os.Exit(1)
	}

	// Create a shared encryption key (must be 16 bytes)
	key := []byte("1234567890123456")

	// Create client
	options := []c2.Option{
		c2.WithClientKey(key),
		c2.WithClientAddress(*address),
		c2.WithClientIdentifier(*identifier),
		c2.WithClientInterval(*interval),
		c2.WithClientProtocol(protocolType),
		c2.WithClientLogger(os.Stdout), // Output logs to standard output
	}

	// Add domain option only for DNS protocol
	if protocolType == c2.ProtocolDNS {
		options = append(options, c2.WithClientDomain(*domain))
	}

	client, err := c2.NewClient(options...)
	if err != nil {
		fmt.Printf("Failed to create client: %v\n", err)
		return
	}

	// Start client in a goroutine
	go func() {
		fmt.Printf("Client started, connecting to %s using protocol: %s\n", *address, protocolType)
		if protocolType == c2.ProtocolDNS {
			fmt.Printf("Using DNS domain: %s\n", *domain)
		}
		if err := client.Start(); err != nil {
			fmt.Printf("Client stopped with error: %v\n", err)
		}
		fmt.Println("Client exited successfully")
	}()

	// Wait for interrupt signal to stop
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	fmt.Println("Client running. Press Ctrl+C to stop.")
	<-sigCh

	// Stop client gracefully
	client.Stop()
	fmt.Println("Client stopped gracefully")
}
