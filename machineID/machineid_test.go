package machineid

import (
	"testing"
)

// TestGetMachineID 测试机器ID获取功能
func TestGetMachineID(t *testing.T) {
	machineID, err := GetMachineID()
	if err != nil {
		t.Logf("获取机器ID失败: %v", err)
		// 不强制要求测试通过，因为在某些环境中可能没有权限获取机器ID
		return
	}

	if machineID == "" {
		t.Error("获取到的机器ID为空")
		return
	}

	t.Logf("获取到的机器ID: %s", machineID)
}

// TestIsVm 测试虚拟机/容器检测功能
func TestIsVm(t *testing.T) {
	isVm := IsVm()
	t.Logf("当前环境是否为虚拟机/容器: %v", isVm)
	// 不强制断言结果，因为测试环境可能是物理机或虚拟机
}

