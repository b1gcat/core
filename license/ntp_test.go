package license

import (
	"testing"
	"time"
)

// TestNTPTime 获取NTP时间并验证其有效性
func TestNTPTime(t *testing.T) {
	// 创建默认配置实例
	cfg := New()

	// 获取NTP时间
	ntpTime := cfg.getNTPTime()

	// 获取系统时间
	sysTime := time.Now()

	// 验证NTP时间不为零值
	if ntpTime.IsZero() {
		t.Error("NTP time should not be zero value")
	}

	// 验证NTP时间与系统时间相差不超过5分钟（允许网络延迟和时钟偏差）
	if ntpTime.Sub(sysTime) > 5*time.Minute || sysTime.Sub(ntpTime) > 5*time.Minute {
		t.Errorf("NTP time %v differs significantly from system time %v", ntpTime, sysTime)
	}

	t.Logf("NTP time: %v", ntpTime)
	t.Logf("System time: %v", sysTime)
	t.Logf("Time difference: %v", ntpTime.Sub(sysTime))
}

// TestNTPServerList 验证NTP服务器列表配置正确
func TestNTPServerList(t *testing.T) {
	// 验证服务器列表不为空
	if len(globalConfig.ntpServers) == 0 {
		t.Error("NTP server list should not be empty")
	}

	// 验证至少包含一个国内服务器
	hasDomestic := false
	for _, server := range globalConfig.ntpServers {
		// 检查是否包含国内服务器IP特征
		if server == "210.72.145.44:123" || server == "203.107.6.88:123" {
			hasDomestic = true
			break
		}
	}

	if !hasDomestic {
		t.Error("NTP server list should contain at least one domestic server")
	}

	// 验证至少包含一个国际服务器
	hasInternational := false
	for _, server := range globalConfig.ntpServers {
		if server == "pool.ntp.org:123" || server == "time.windows.com:123" {
			hasInternational = true
			break
		}
	}

	if !hasInternational {
		t.Error("NTP server list should contain at least one international server")
	}

	t.Logf("NTP server list contains %d servers", len(globalConfig.ntpServers))
	for i, server := range globalConfig.ntpServers {
		t.Logf("Server %d: %s", i+1, server)
	}
}
