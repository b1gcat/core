//go:build ignore

package main

import (
	"fmt"
	"log"
	"os"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Println("Usage: 0trust-client <server-addr> <secret>")
		fmt.Println("Example: 0trust-client 192.168.1.100:12345 mysecretkey")
		os.Exit(1)
	}

	serverAddr := os.Args[1]
	secret := []byte(os.Args[2])

	client := NewClient(serverAddr, secret)
	if err := client.Authenticate(); err != nil {
		log.Fatalf("Authentication failed: %v", err)
	}

	fmt.Println("\nPress Enter to exit...")
	fmt.Scanln()
}
