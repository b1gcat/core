package license

import (
	"strconv"
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

// TestGetRemainingTime 测试获取剩余授权时间的功能
func TestGetRemainingTime(t *testing.T) {
	// Save original buildTime to restore after test
	originalBuildTime := buildTime
	defer func() {
		buildTime = originalBuildTime
	}()

	// Test 1: Permanent license (expirationDays <= 0)
	t.Run("PermanentLicense", func(t *testing.T) {
		cfg := New(WithExpiration(0))
		remaining, err := cfg.GetRemainingTime()
		if err != nil {
			t.Errorf("Expected no error for permanent license, got %v", err)
		}
		if remaining != 0 {
			t.Errorf("Expected 0 remaining time for permanent license, got %v", remaining)
		}
	})

	// Test 2: Valid remaining time (set buildTime to 1 day ago, expiration 30 days)
	t.Run("ValidRemainingTime", func(t *testing.T) {
		// Set buildTime to 24 hours ago
		twoDaysAgo := time.Now().Add(-24 * time.Hour).Unix()
		buildTime = strconv.FormatInt(twoDaysAgo, 10)

		cfg := New(WithExpiration(30))
		remaining, err := cfg.GetRemainingTime()
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		// Should be approximately 29 days remaining
		expectedRemaining := 29 * 24 * time.Hour
		diff := absDuration(remaining - expectedRemaining)
		if diff > 5*time.Minute { // Allow 5 minutes tolerance for NTP time variations
			t.Errorf("Expected approximately %v remaining time, got %v", expectedRemaining, remaining)
		}

		t.Logf("Valid license remaining time: %v", remaining)
	})

	// Test 3: Expired license
	t.Run("ExpiredLicense", func(t *testing.T) {
		// Set buildTime to 31 days ago, expiration 30 days
		thirtyOneDaysAgo := time.Now().Add(-31 * 24 * time.Hour).Unix()
		buildTime = strconv.FormatInt(thirtyOneDaysAgo, 10)

		cfg := New(WithExpiration(30))
		remaining, err := cfg.GetRemainingTime()
		if err != nil {
			t.Errorf("Expected no error for expired license, got %v", err)
		}
		if remaining >= 0 {
			t.Errorf("Expected negative remaining time for expired license, got %v", remaining)
		}

		// Should be approximately 1 day expired
		expectedExpired := -24 * time.Hour
		diff := absDuration(remaining - expectedExpired)
		if diff > 5*time.Minute { // Allow 5 minutes tolerance for NTP time variations
			t.Errorf("Expected approximately %v expired time, got %v", expectedExpired, remaining)
		}

		t.Logf("Expired license time: %v", remaining)
	})

	// Test 4: Invalid buildTime
	t.Run("InvalidBuildTime", func(t *testing.T) {
		buildTime = "invalid_time"

		cfg := New(WithExpiration(30))
		remaining, err := cfg.GetRemainingTime()
		if err == nil {
			t.Error("Expected error for invalid buildTime, got nil")
		}
		if remaining != 0 {
			t.Errorf("Expected 0 remaining time for invalid buildTime, got %v", remaining)
		}

		t.Logf("Expected error occurred for invalid buildTime: %v", err)
	})
}

// absDuration returns the absolute value of a duration
func absDuration(d time.Duration) time.Duration {
	if d < 0 {
		return -d
	}
	return d
}
