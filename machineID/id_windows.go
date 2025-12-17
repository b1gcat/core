//go:build windows

package machineid

import (
	"errors"
	"syscall"
	"unsafe"
)

var (
	modkernel32            = syscall.NewLazyDLL("kernel32.dll")
	procGetComputerNameExW = modkernel32.NewProc("GetComputerNameExW")
)

const (
	ComputerNamePhysicalDnsHostname = 1 // 完全限定的DNS名称
)

func init() {
	getMachineID = windowsGetMachineID
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
