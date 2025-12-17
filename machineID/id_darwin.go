//go:build darwin

package machineid

import (
	"bytes"
	"errors"
	"os/exec"
	"strings"
)

func init() {
	getMachineID = darwinGetMachineID
}

// darwinGetMachineID 获取macOS系统的机器ID
func darwinGetMachineID() (string, error) {
	// 在macOS中，可以使用ioreg命令获取硬件UUID
	cmd := exec.Command("ioreg", "-rd1", "-c", "IOPlatformExpertDevice")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	// 查找Hardware UUID行
	lines := bytes.Split(output, []byte("\n"))
	for _, line := range lines {
		lineStr := string(line)
		if strings.Contains(lineStr, "IOPlatformUUID") {
			// 提取UUID值
			parts := strings.Split(lineStr, " = ")
			if len(parts) != 2 {
				continue
			}

			uuid := strings.Trim(parts[1], `"`)
			uuid = strings.TrimSpace(uuid)
			if uuid != "" {
				return uuid, nil
			}
		}
	}

	return "", errors.New("无法获取macOS机器ID")
}
