package main

import (
	"encoding/json"
	"fmt"
	"net"
	"time"
)

type Client struct {
	serverAddr string
	secret     []byte
}

func NewClient(serverAddr string, secret []byte) *Client {
	return &Client{
		serverAddr: serverAddr,
		secret:     secret,
	}
}

func (c *Client) Authenticate() error {
	fmt.Println("0trust client starting authentication...")

	conn, err := net.DialTimeout("udp", c.serverAddr, timeout*time.Second)
	if err != nil {
		return fmt.Errorf("failed to connect to server: %v", err)
	}
	defer conn.Close()

	fmt.Println("✓ Connected to server")

	nonce, err := generateNonce()
	if err != nil {
		return fmt.Errorf("failed to generate nonce: %v", err)
	}

	fmt.Println("✓ Generated nonce")

	signature := signMessage(nonce, c.secret)
	fmt.Println("✓ Signed nonce with secret")

	msg := AuthMessage{
		Nonce:     nonce,
		Signature: signature,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %v", err)
	}

	fmt.Println("✓ Sending authentication request...")
	if _, err := conn.Write(data); err != nil {
		return fmt.Errorf("failed to send message: %v", err)
	}

	conn.SetReadDeadline(time.Now().Add(timeout * time.Second))
	buffer := make([]byte, bufferSize)
	n, err := conn.Read(buffer)
	if err != nil {
		return fmt.Errorf("failed to read response: %v", err)
	}

	var response map[string]string
	if err := json.Unmarshal(buffer[:n], &response); err != nil {
		return fmt.Errorf("failed to unmarshal response: %v", err)
	}

	if response["status"] == "ok" {
		fmt.Println("✅ Authentication successful!")
		fmt.Println("Your IP has been added to the trusted list.")
		fmt.Println("You can now access the protected service.")
		return nil
	}

	return fmt.Errorf("authentication failed: %v", response)
}
