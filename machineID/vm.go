package machineid

// IsVm 检查当前系统是否运行在虚拟机或容器环境中
//
// 该函数是一个跨平台的虚拟机/容器检测接口，会根据当前操作系统调用相应的实现
// - Linux: 检查/proc/cpuinfo、/proc/meminfo、/proc/scsi/scsi、/sys/class/dmi/id等文件的VM特征
// - Windows: 使用WMI查询系统硬件信息和虚拟设备特征
// - macOS: 检查系统设备树和虚拟环境特征
//
// 返回值：
// - bool: 如果运行在虚拟机或容器中，返回true；否则返回false
func IsVm() bool {
	// 该函数会根据编译时的操作系统自动选择对应的实现
	return isVm()
}

// isVm 是一个内部函数，由具体的操作系统实现
var isVm func() bool
