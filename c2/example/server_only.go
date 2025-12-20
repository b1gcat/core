//go:build ignore

package main

import (
	"flag"
	"fmt"

	"github.com/b1gcat/core/c2"
)

func main() {
	// Define flag parameters
	key := flag.String("key", "1234567890123456", "Encryption key (must be 16 characters)")
	address := flag.String("address", "0.0.0.0:123", "Server listen address")
	flag.Parse()

	// Validate key length
	if len(*key) != 16 {
		fmt.Printf("Error: key must be 16 characters long, got %d\n", len(*key))
		return
	}

	// Create server with flag parameters
	server, err := c2.NewServer(
		c2.WithServerKey(*key),
		c2.WithServerAddress(*address),
	)
	if err != nil {
		fmt.Printf("Failed to create server: %v\n", err)
		return
	}

	if err := server.Start(); err != nil {
		fmt.Printf("Server stopped with error: %v\n", err)
	}

}
