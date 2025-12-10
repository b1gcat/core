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

// httpClient 带认证的HTTP客户端
var httpClient = &http.Client{
	Timeout: 30 * time.Second,
}

// fetchWithAuth 使用认证获取URL内容
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

// downloadFileWithAuth 使用认证下载文件到指定路径
func downloadFileWithAuth(url, username, password, destPath string) error {
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

	// 创建目标目录
	destDir := filepath.Dir(destPath)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("create destination directory failed: %w", err)
	}

	// 创建目标文件
	destFile, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("create destination file failed: %w", err)
	}
	defer destFile.Close()

	// 下载文件内容
	if _, err := io.Copy(destFile, resp.Body); err != nil {
		return fmt.Errorf("download file failed: %w", err)
	}

	return nil
}

// calculateSHA256 计算文件的SHA256哈希值
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

// parseUpgradeList 解析升级列表文件
func parseUpgradeList(content []byte) (map[string]string, error) {
	lines := strings.Split(string(content), "\n")
	upgradeMap := make(map[string]string)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.Split(line, ":")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid upgrade list line: %s", line)
		}

		filename := strings.TrimSpace(parts[0])
		hash := strings.TrimSpace(parts[1])
		upgradeMap[filename] = hash
	}

	return upgradeMap, nil
}

// getTempFilePath 获取临时文件路径
func getTempFilePath(filename string) string {
	tempDir := os.TempDir()
	timestamp := time.Now().UnixNano() / int64(time.Millisecond)
	return filepath.Join(tempDir, fmt.Sprintf("%s.%d", filename, timestamp))
}
