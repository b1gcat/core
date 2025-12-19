//go:build windows

package machineid

import (
	"errors"
	"os"
	"syscall"
	"unsafe"
)

var (
	modkernel32            = syscall.NewLazyDLL("kernel32.dll")
	procGetComputerNameExW = modkernel32.NewProc("GetComputerNameExW")
	modadvapi32            = syscall.NewLazyDLL("advapi32.dll")
	procRegOpenKeyExW      = modadvapi32.NewProc("RegOpenKeyExW")
	procRegQueryValueExW   = modadvapi32.NewProc("RegQueryValueExW")
	procRegCloseKey        = modadvapi32.NewProc("RegCloseKey")
)

const (
	ComputerNamePhysicalDnsHostname = 1 // 完全限定的DNS名称

	// Registry constants
	hkeyLocalMachine = 0x80000002
	keyQueryValue    = 0x0001
	regSZ            = 1
)

func init() {
	getMachineID = windowsGetMachineID
	isVm = windowsIsVm
}

// windowsIsVm 检查Windows系统是否运行在虚拟机或容器中
func windowsIsVm() bool {
	// 检查容器环境
	containerChecks := []string{
		// WSL容器特征
		"c:\\windows\\system32\\config\\systemprofile\\.wslconfig",
		// Docker容器特征
		"c:\\windows\\system32\\config\\systemprofile\\.dockerinit",
		"c:\\.dockerenv",
	}

	for _, path := range containerChecks {
		if _, err := os.Stat(path); err == nil {
			return true
		}
	}

	// 检查容器相关环境变量
	containerEnvVars := []string{
		"WSL_DISTRO_NAME",
		"DOCKER_CONTAINER",
		"CONTAINER_NAME",
		"PODMAN_CONTAINER",
		"KUBERNETES_SERVICE_HOST",
	}

	for _, envVar := range containerEnvVars {
		if os.Getenv(envVar) != "" {
			return true
		}
	}

	// 检查注册表中的虚拟机特征
	vmRegKeys := []string{
		"SOFTWARE\\Microsoft\\Windows NT\\CurrentVersion\\Virtualization",
		"SYSTEM\\CurrentControlSet\\Services\\VBoxGuest",
		"SYSTEM\\CurrentControlSet\\Services\\vmhgfs",
		"SYSTEM\\CurrentControlSet\\Services\\xenevtchn",
		"SYSTEM\\CurrentControlSet\\Services\\xennet",
		"SYSTEM\\CurrentControlSet\\Services\\hvboot",
		"SYSTEM\\CurrentControlSet\\Services\\hvdetection",
		"SYSTEM\\CurrentControlSet\\Services\\hvservice",
		"SYSTEM\\CurrentControlSet\\Services\\hyperkbd",
		"SYSTEM\\CurrentControlSet\\Services\\hypervreg",
		"SYSTEM\\CurrentControlSet\\Services\\hypervstorvsc",
		"SYSTEM\\CurrentControlSet\\Services\\vm3dmp",
		"SYSTEM\\CurrentControlSet\\Services\\vmicrdv",
		"SYSTEM\\CurrentControlSet\\Services\\vmicheartbeat",
		"SYSTEM\\CurrentControlSet\\Services\\vmickvpexchange",
		"SYSTEM\\CurrentControlSet\\Services\\vmicshutdown",
		"SYSTEM\\CurrentControlSet\\Services\\vmicvmsession",
		"SYSTEM\\CurrentControlSet\\Services\\vmicvss",
	}

	for _, keyPath := range vmRegKeys {
		if regKeyExists(keyPath) {
			return true
		}
	}

	// 检查虚拟机相关的设备和驱动
	vmDrivers := []string{
		"VBoxGuest.sys",
		"VBoxMouse.sys",
		"VBoxService.exe",
		"VBoxTray.exe",
		"vmhgfs.sys",
		"vmxnet3.sys",
		"xenevtchn.sys",
		"xennet.sys",
		"hyperkbd.sys",
		"hvboot.sys",
		"hvdetection.sys",
		"hvservice.sys",
		"hypervstorvsc.sys",
	}

	for _, driver := range vmDrivers {
		if fileExists("c:\\windows\\system32\\drivers\\"+driver) ||
			fileExists("c:\\windows\\system32\\"+driver) {
			return true
		}
	}

	return false
}

// regKeyExists 检查Windows注册表键是否存在
func regKeyExists(keyPath string) bool {
	var hKey syscall.Handle

	// 将Go字符串转换为UTF-16
	keyPathUTF16, _ := syscall.UTF16PtrFromString(keyPath)

	// 打开注册表键
	r1, _, _ := procRegOpenKeyExW.Call(
		uintptr(hkeyLocalMachine),
		uintptr(unsafe.Pointer(keyPathUTF16)),
		0,
		uintptr(keyQueryValue),
		uintptr(unsafe.Pointer(&hKey)),
	)

	// 检查是否成功打开
	if r1 != 0 {
		return false
	}

	// 关闭注册表键
	procRegCloseKey.Call(uintptr(hKey))
	return true
}

// fileExists 检查文件是否存在
func fileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return err == nil
}

// windowsGetMachineID 获取Windows系统的机器ID
func windowsGetMachineID() (string, error) {
	// 在Windows中，机器ID使用计算机名
	var bufSize uint32 = 1024
	buf := make([]uint16, bufSize)

	r1, _, _ := procGetComputerNameExW.Call(
		uintptr(ComputerNamePhysicalDnsHostname),
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(unsafe.Pointer(&bufSize)),
	)

	if r1 == 0 {
		return "", errors.New("无法获取计算机名")
	}

	machineID := syscall.UTF16ToString(buf[:bufSize])
	if machineID == "" {
		return "", errors.New("计算机名为空")
	}

	return machineID, nil
}
