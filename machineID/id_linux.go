//go:build linux

package machineid

import (
	"crypto/md5"
	"errors"
	"fmt"
	"os"
	"strings"
)

func init() {
	getMachineID = linuxGetMachineID
}

// isContainer 检查当前是否运行在容器环境中
func isContainer() bool {
	// 检查Docker容器特征
	if _, err := os.Stat("/.dockerenv"); err == nil {
		return true
	}

	// 检查Kubernetes容器特征
	if _, err := os.Stat("/var/run/secrets/kubernetes.io/serviceaccount"); err == nil {
		return true
	}

	// 检查/proc/self/cgroup内容
	cgroupContent, err := os.ReadFile("/proc/self/cgroup")
	if err == nil {
		contentStr := string(cgroupContent)
		// 检查是否包含容器运行时特征
		if strings.Contains(contentStr, "docker") ||
			strings.Contains(contentStr, "containerd") ||
			strings.Contains(contentStr, "cri-o") ||
			strings.Contains(contentStr, "podman") {
			return true
		}
	}

	// 检查/proc/1/sched进程名
	schedContent, err := os.ReadFile("/proc/1/sched")
	if err == nil {
		contentStr := string(schedContent)
		// 容器中init进程通常不是systemd或init
		if !strings.HasPrefix(contentStr, "systemd (") &&
			!strings.HasPrefix(contentStr, "init (") {
			return true
		}
	}

	return false
}

// linuxGetMachineID 获取Linux系统的机器ID
func linuxGetMachineID() (string, error) {
	// 检查是否运行在容器环境中
	if isContainer() {
		return "", errors.New("无法获取机器ID")
	}

	// 需要计算MD5的文件路径
	md5Paths := []string{
		"/etc/machine-id",
		"/var/lib/dbus/machine-id",
		"/sys/devices/virtual/dmi/id/product_uuid", // 添加对DMI product_uuid的支持
	}

	// 处理需要计算MD5的文件
	for _, path := range md5Paths {
		if _, err := os.Stat(path); err == nil {
			content, err := os.ReadFile(path)
			if err != nil {
				continue
			}

			machineID := strings.TrimSpace(string(content))
			if machineID != "" {
				// 计算MD5哈希值
				hash := md5.Sum([]byte(machineID))
				return fmt.Sprintf("%x", hash), nil
			}
		}
	}

	// 如果上述位置都不存在，返回错误
	return "", errors.New("无法获取机器ID")
}
