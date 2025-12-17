package upgrade

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// httpClient HTTP client with authentication
var httpClient = &http.Client{
	// No global timeout, use fine-grained timeout control in Transport
	Transport: &http.Transport{
		// Timeout for TLS handshake
		TLSHandshakeTimeout: 30 * time.Second,
		// Timeout for waiting response headers
		ResponseHeaderTimeout: 30 * time.Second,
		// Maximum idle time between two read operations
		IdleConnTimeout: 90 * time.Second,
		// Enable HTTP/2 support (optional)
		ForceAttemptHTTP2: true,
	},
}

// fetchWithAuth fetch URL content with authentication
func fetchWithAuth(url, username, password string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request failed: %w", err)
	}

	req.SetBasicAuth(username, password)
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body failed: %w", err)
	}

	return body, nil
}

// ProgressWriter implements io.Writer interface for reporting download progress
type ProgressWriter struct {
	Writer     io.Writer
	Total      int64
	Downloaded int64
	Callback   ProgressCallback
}

// Write implements io.Writer interface, records downloaded bytes and calls progress callback
func (pw *ProgressWriter) Write(p []byte) (int, error) {
	n, err := pw.Writer.Write(p)
	if err != nil {
		return n, err
	}
	pw.Downloaded += int64(n)
	if pw.Callback != nil {
		pw.Callback(pw.Downloaded, pw.Total)
	}
	return n, nil
}

// downloadFileWithAuth download file to specified path with authentication
func downloadFileWithAuth(url, username, password, destPath string, progressCallback ProgressCallback) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("create request failed: %w", err)
	}

	req.SetBasicAuth(username, password)
	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Create destination directory
	destDir := filepath.Dir(destPath)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("create destination directory failed: %w", err)
	}

	// Create destination file
	destFile, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("create destination file failed: %w", err)
	}
	defer destFile.Close()

	// Create progress writer
	progressWriter := &ProgressWriter{
		Writer:   destFile,
		Total:    resp.ContentLength,
		Callback: progressCallback,
	}

	// Download file content
	if _, err := io.Copy(progressWriter, resp.Body); err != nil {
		return fmt.Errorf("download file failed: %w", err)
	}

	return nil
}

// calculateSHA256 calculate SHA256 hash of a file
func calculateSHA256(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("open file failed: %w", err)
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", fmt.Errorf("calculate hash failed: %w", err)
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

// parseUpgradeList parse upgrade list file
func parseUpgradeList(content []byte) (map[string]string, error) {
	lines := strings.Split(string(content), "\n")
	upgradeMap := make(map[string]string)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Parse format: sha256sum upgrade package name
		parts := strings.SplitN(line, " ", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid upgrade list line: %s", line)
		}

		hash := strings.TrimSpace(parts[0])
		filename := strings.TrimSpace(parts[1])
		upgradeMap[filename] = hash
	}

	return upgradeMap, nil
}

// getTempFilePath get temporary file path
func getTempFilePath(filename string) string {
	tempDir := os.TempDir()
	timestamp := time.Now().UnixNano() / int64(time.Millisecond)
	return filepath.Join(tempDir, fmt.Sprintf("%s.%d", filename, timestamp))
}
