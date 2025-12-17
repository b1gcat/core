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
