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

// checkInstallTool 检查install工具是否存在
func checkInstallTool() (bool, error) {
	switch runtime.GOOS {
	case "linux", "darwin":
		// 在Linux和macOS上检查install命令
		output, err := shellexec.ExecUnix("which install")
		if err != nil {
			return false, fmt.Errorf("check install tool failed: %w", err)
		}
		return strings.TrimSpace(*output) != "", nil
	case "windows":
		// 在Windows上检查system32目录下是否有install.exe
		system32Path := os.Getenv("SystemRoot") + "\\System32"
		installExePath := filepath.Join(system32Path, "install.exe")

		// 检查install.exe是否已经存在
		if _, err := os.Stat(installExePath); err == nil {
			return true, nil
		}

		// 如果不存在，返回错误
		return false, fmt.Errorf("install.exe not found in system32 directory")
	default:
		return false, fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
}

// installFile 安装文件到目标路径
func installFile(sourcePath, destPath string) error {
	// 检查文件是否存在
	if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
		return fmt.Errorf("source file does not exist: %s", sourcePath)
	}

	// 确保目标目录存在
	destDir := strings.TrimSuffix(destPath, "\n")
	destDir = strings.TrimSuffix(destDir, "\r")
	destDir = strings.TrimSpace(destDir)

	if destDir != "" {
		destDir = strings.TrimSuffix(destDir, string(os.PathSeparator))
		destDir = strings.TrimSuffix(destDir, "/")
		destDir = strings.TrimSuffix(destDir, "\\")
		destDir = strings.TrimSpace(destDir)

		if destDir != "" {
			if err := os.MkdirAll(destDir, 0755); err != nil {
				return fmt.Errorf("create destination directory failed: %w", err)
			}
		}
	}

	switch runtime.GOOS {
	case "linux", "darwin":
		// 在Linux和macOS上使用install命令
		cmd := fmt.Sprintf("install %s %s", sourcePath, destPath)
		_, err := shellexec.ExecUnix(cmd)
		if err != nil {
			return fmt.Errorf("install failed: %w", err)
		}
	case "windows":
		// 在Windows上使用install命令
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

// getExecutablePath 获取当前可执行文件的路径
func getExecutablePath() (string, error) {
	path, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("get executable path failed: %w", err)
	}
	return path, nil
}

func init() {
	// 在Windows系统上，将嵌入式install.exe写入system32目录
	if runtime.GOOS == "windows" {
		// 在Windows上检查system32目录下是否有install.exe
		system32Path := os.Getenv("SystemRoot") + "\\System32"
		installExePath := filepath.Join(system32Path, "install.exe")

		// 检查install.exe是否已经存在
		if _, err := os.Stat(installExePath); err == nil {
			// 文件已存在，无需写入
			return
		}

		// 如果不存在，将嵌入的install.exe写入system32目录
		installExeData, err := installExe.ReadFile("files/install.exe")
		if err != nil {
			fmt.Printf("read embedded install.exe failed: %v\n", err)
			return
		}

		// 创建目标文件
		targetFile, err := os.OpenFile(installExePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
		if err != nil {
			fmt.Printf("open target file failed: %v\n", err)
			return
		}
		defer targetFile.Close()

		// 写入文件内容
		if _, err := targetFile.Write(installExeData); err != nil {
			fmt.Printf("write install.exe to system32 failed: %v\n", err)
			return
		}
	}
}
