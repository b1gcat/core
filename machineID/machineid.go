package machineid

// GetMachineID 获取当前操作系统的机器ID
//
// 该函数是一个跨平台的机器ID获取接口，会根据当前操作系统调用相应的实现：
// - Linux: 按优先级从/etc/machine-id、/var/lib/dbus/machine-id或/sys/devices/virtual/dmi/id/product_uuid获取内容的MD5哈希值
// - Windows: 获取计算机名（完全限定的DNS名称）
// - macOS: 使用ioreg命令获取硬件UUID
//
// 返回值：
// - string: 机器ID字符串
// - error: 如果获取失败，返回相应的错误信息
func GetMachineID() (string, error) {
	// 该函数会根据编译时的操作系统自动选择对应的实现
	// 具体实现见id_linux.go、id_windows.go和id_darwin.go文件
	return getMachineID()
}

// getMachineID 是一个内部函数，由具体的操作系统实现
var getMachineID func() (string, error)
