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
	isVm = linuxIsVm
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

	return false
}

// linuxIsVm 检查Linux系统是否运行在虚拟机中
func linuxIsVm() bool {
	// 首先检查是否运行在容器中
	if isContainer() {
		return true
	}

	// 检查/proc/cpuinfo中的虚拟机特征
	cpuInfo, err := os.ReadFile("/proc/cpuinfo")
	if err == nil {
		contentStr := string(cpuInfo)
		// 检查常见的VM CPU特征
		vmCpuPatterns := []string{
			"QEMU Virtual CPU",
			"VMware Virtual CPU",
			"VirtualBox CPU",
			"Xen Virtual CPU",
			"Hyper-V",
			"KVM",
			"Microsoft Hv",
		}

		for _, pattern := range vmCpuPatterns {
			if strings.Contains(contentStr, pattern) {
				return true
			}
		}
	}

	// 检查/proc/scsi/scsi中的虚拟机特征
	scsiInfo, err := os.ReadFile("/proc/scsi/scsi")
	if err == nil {
		contentStr := string(scsiInfo)
		// 检查常见的VM SCSI设备特征
		vmScsiPatterns := []string{
			"VMware",
			"VirtualBox",
			"QEMU",
			"Xen",
			"Hyper-V",
		}

		for _, pattern := range vmScsiPatterns {
			if strings.Contains(contentStr, pattern) {
				return true
			}
		}
	}

	// 检查/sys/class/dmi/id中的虚拟机特征
	dmiPaths := []string{
		"/sys/class/dmi/id/board_vendor",
		"/sys/class/dmi/id/board_name",
		"/sys/class/dmi/id/chassis_vendor",
		"/sys/class/dmi/id/chassis_name",
	}

	vmDmiPatterns := []string{
		"VMware",
		"VirtualBox",
		"QEMU",
		"Xen",
		"Hyper-V",
		"Microsoft Hv",
	}

	for _, path := range dmiPaths {
		content, err := os.ReadFile(path)
		if err == nil {
			contentStr := strings.TrimSpace(string(content))
			for _, pattern := range vmDmiPatterns {
				if strings.Contains(contentStr, pattern) {
					return true
				}
			}
		}
	}

	// 检查/proc/meminfo中的虚拟机特征
	memInfo, err := os.ReadFile("/proc/meminfo")
	if err == nil {
		contentStr := string(memInfo)
		// 检查是否包含Hyper-V内存管理特征
		if strings.Contains(contentStr, "Hyper-V Memory") {
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
