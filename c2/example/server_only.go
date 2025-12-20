// tag: server_only
package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/b1gcat/core/c2"
)

func main() {
	// Create a shared encryption key (must be 16 bytes)
	key := []byte("1234567890123456")

	// Create server
	server, err := c2.NewServer(
		c2.WithServerKey(key),
		c2.WithServerAddress("0.0.0.0:9003"),
	)
	if err != nil {
		fmt.Printf("Failed to create server: %v\n", err)
		return
	}

	// Handle interrupt signals for graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// Start server in a goroutine
	go func() {
		if err := server.Start(); err != nil {
			fmt.Printf("Server stopped with error: %v\n", err)
		}
	}()

	// Print server information
	fmt.Println("C2 Server started on 0.0.0.0:9003")
	fmt.Println("Type 'help' for available commands")
	fmt.Println("Press Ctrl+C to stop the server")
	fmt.Println()

	// Wait for interrupt signal
	<-sigCh

	// Stop server gracefully
	fmt.Println("\nStopping server...")
	server.Stop()
	fmt.Println("Server exited successfully")
}
