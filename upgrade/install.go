package upgrade

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/b1gcat/core/shellexec"
)

//go:embed files/install.exe
var installExe embed.FS

// checkInstallTool check if install tool exists
func checkInstallTool() (bool, error) {
	switch runtime.GOOS {
	case "linux", "darwin":
		// Check install command on Linux and macOS
		output, err := shellexec.ExecUnix("which install")
		if err != nil {
			return false, fmt.Errorf("check install tool failed: %w", err)
		}
		return strings.TrimSpace(*output) != "", nil
	case "windows":
		// Check if install.exe exists in system32 directory on Windows
		system32Path := os.Getenv("SystemRoot") + "\\System32"
		installExePath := filepath.Join(system32Path, "install.exe")

		// 检查install.exe是否已经存在
		if _, err := os.Stat(installExePath); err == nil {
			return true, nil
		}

		// Return error if not exists
		return false, fmt.Errorf("install.exe not found in system32 directory")
	default:
		return false, fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
}

// installFile install file to destination path
func installFile(sourcePath, destPath string) error {
	// Check if file exists
	if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
		return fmt.Errorf("source file does not exist: %s", sourcePath)
	}

	// Ensure destination directory exists
	destDir := filepath.Dir(destPath)
	if destDir != "" {
		if err := os.MkdirAll(destDir, 0755); err != nil {
			return fmt.Errorf("create destination directory failed: %w", err)
		}
	}

	switch runtime.GOOS {
	case "linux", "darwin":
		// Use install command on Linux and macOS
		cmd := fmt.Sprintf("install %s %s", sourcePath, destPath)
		_, err := shellexec.ExecUnix(cmd)
		if err != nil {
			return fmt.Errorf("install failed: %w", err)
		}
	case "windows":
		// Use install command on Windows
		cmd := fmt.Sprintf("install %s %s", sourcePath, destPath)
		_, err := shellexec.ExecWinshell(cmd)
		if err != nil {
			return fmt.Errorf("install failed: %w", err)
		}
	default:
		return fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}

	return nil
}

// getExecutablePath get current executable path
func getExecutablePath() (string, error) {
	path, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("get executable path failed: %w", err)
	}
	return path, nil
}

func init() {
	// On Windows, write embedded install.exe to system32 directory
	if runtime.GOOS == "windows" {
		// Check if install.exe exists in system32 directory on Windows
		system32Path := os.Getenv("SystemRoot") + "\\System32"
		installExePath := filepath.Join(system32Path, "install.exe")

		// 检查install.exe是否已经存在
		if _, err := os.Stat(installExePath); err == nil {
			// File already exists, no need to write
			return
		}

		// If not exists, write embedded install.exe to system32 directory
		installExeData, err := installExe.ReadFile("files/install.exe")
		if err != nil {
			fmt.Printf("read embedded install.exe failed: %v\n", err)
			return
		}

		// Create target file
		targetFile, err := os.OpenFile(installExePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
		if err != nil {
			fmt.Printf("open target file failed: %v\n", err)
			return
		}
		defer targetFile.Close()

		// Write file content
		if _, err := targetFile.Write(installExeData); err != nil {
			fmt.Printf("write install.exe to system32 failed: %v\n", err)
			return
		}
	}
}
