//go:build darwin

package machineid

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

func init() {
	getMachineID = darwinGetMachineID
	isVm = darwinIsVm
}

// darwinIsVm 检查macOS系统是否运行在虚拟机或容器中
func darwinIsVm() bool {
	// 检查容器环境
	if _, err := os.Stat("/.dockerenv"); err == nil {
		return true
	}

	// 检查Docker容器特征
	if _, err := os.Stat("/var/run/secrets/kubernetes.io/serviceaccount"); err == nil {
		return true
	}

	// 检查cgroup内容
	if cgroupContent, err := os.ReadFile("/proc/self/cgroup"); err == nil {
		contentStr := string(cgroupContent)
		if strings.Contains(contentStr, "docker") ||
			strings.Contains(contentStr, "containerd") ||
			strings.Contains(contentStr, "cri-o") ||
			strings.Contains(contentStr, "podman") {
			return true
		}
	}

	// 检查虚拟机相关的内核扩展
	vmKexts := []string{
		"com.vmware.kext.vmhgfs",
		"com.vmware.kext.vmci",
		"com.vmware.kext.vmx86",
		"org.virtualbox.kext.VBoxDrv",
		"org.virtualbox.kext.VBoxNetFlt",
		"org.virtualbox.kext.VBoxNetAdp",
		"org.virtualbox.kext.VBoxUSB",
	}

	for _, kext := range vmKexts {
		cmd := exec.Command("kextstat", "-l", "-b", kext)
		output, err := cmd.CombinedOutput()
		if err == nil {
			// 检查输出是否真正包含内核扩展信息，而不仅仅是命令的状态提示
			outputStr := string(output)
			if len(outputStr) > 0 && !strings.Contains(outputStr, "kmutil showloaded") && !strings.Contains(outputStr, "No variant specified") {
				return true
			}
		}
	}

	// 检查system_profiler
	spCmd := exec.Command("system_profiler", "SPHardwareDataType")
	spOutput, err := spCmd.Output()
	if err == nil {
		spStr := string(spOutput)
		vmKeywords := []string{"VMware", "VirtualBox", "Parallels", "QEMU", "Xen"}
		for _, kw := range vmKeywords {
			if strings.Contains(spStr, kw) {
				return true
			}
		}
	}

	// 检查sysctl CPU品牌
	sysctlCmd := exec.Command("sysctl", "machdep.cpu.brand_string")
	sysctlOutput, err := sysctlCmd.Output()
	if err == nil {
		sysctlStr := string(sysctlOutput)
		vmCpuKeywords := []string{"QEMU Virtual CPU", "VMware Virtual CPU", "VirtualBox CPU", "Xen Virtual CPU"}
		for _, kw := range vmCpuKeywords {
			if strings.Contains(sysctlStr, kw) {
				return true
			}
		}
	}

	// 检查IOPlatformExpertDevice的硬件特性
	ioregDeviceCmd := exec.Command("ioreg", "-c", "IOPlatformExpertDevice", "-l")
	ioregDeviceOutput, err := ioregDeviceCmd.Output()
	if err == nil {
		ioregDeviceStr := string(ioregDeviceOutput)

		// 检查更精确的虚拟机硬件标识，避免匹配到VM软件进程
		hwVmPatterns := []string{
			"VMware.*Hardware",
			"VMware.*Virtual",
			"Parallels.*Hardware",
			"Parallels.*Virtual",
			"QEMU.*Hardware",
			"QEMU.*Virtual",
			"Xen.*Hardware",
			"Xen.*Virtual",
		}

		isVmHardware := false
		for _, pattern := range hwVmPatterns {
			if matched, _ := regexp.MatchString(pattern, ioregDeviceStr); matched {
				// 确保这个匹配不是来自IOUserClientCreator中的VM软件进程
				if !strings.Contains(ioregDeviceStr, "IOUserClientCreator") || !strings.Contains(ioregDeviceStr, "VirtualBox") {
					isVmHardware = true
					break
				}
			}
		}

		if isVmHardware {
			return true
		}
	}

	// 检查PCI设备
	pciCmd := exec.Command("ioreg", "-l", "-r", "-c", "IOPCIDevice")
	pciOutput, err := pciCmd.Output()
	if err == nil {
		pciStr := string(pciOutput)

		// 只检查特定的虚拟显卡PCI ID
		vmPciIds := []string{
			"pci15ad,0405", // VMware SVGA II
			"pci8086,2918", // VirtualBox Graphics
			"pci1ab8,4005", // Parallels Display
		}

		for _, id := range vmPciIds {
			if strings.Contains(pciStr, id) {
				return true
			}
		}
	}

	// 检查硬件相关的模式
	hwPatterns := []string{"VMware.*Hardware", "VirtualBox.*Hardware", "Parallels.*Hardware"}
	fullIoregCmd := exec.Command("ioreg", "-l")
	fullIoregOutput, err := fullIoregCmd.Output()
	if err == nil {
		fullIoregStr := string(fullIoregOutput)

		for _, pattern := range hwPatterns {
			if matched, _ := regexp.MatchString(pattern, fullIoregStr); matched {
				return true
			}
		}
	}

	// 注意：不再检查VirtualBox进程，因为安装了VM软件不等于在虚拟机中运行

	return false
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
