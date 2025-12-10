package upgrade

import (
	"fmt"
	"runtime"
	"testing"
)

// TestParseVersion 测试版本号解析
func TestParseVersion(t *testing.T) {
	testCases := []struct {
		version     string
		expected    *VersionComponents
		expectError bool
	}{
		{"v1.0.0", &VersionComponents{1, 0, 0, ""}, false},
		{"1.0.0", &VersionComponents{1, 0, 0, ""}, false},
		{"V1.2.3", &VersionComponents{1, 2, 3, ""}, false},
		{"version1.2.3", &VersionComponents{1, 2, 3, ""}, false},
		{"1.2.3-alpha", &VersionComponents{1, 2, 3, "alpha"}, false},
		{"1.2.3+build123", &VersionComponents{1, 2, 3, ""}, false},
		{"1.2", &VersionComponents{1, 2, 0, ""}, false},
		{"1", &VersionComponents{1, 0, 0, ""}, false},
		{"invalid", nil, true},
		{"1.0.0.0", nil, true},
		{"1.a.0", nil, true},
	}

	for _, tc := range testCases {
		t.Run(tc.version, func(t *testing.T) {
			result, err := parseVersion(tc.version)
			if tc.expectError {
				if err == nil {
					t.Errorf("expected error for version %s, got none", tc.version)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error for version %s: %v", tc.version, err)
				return
			}

			if result.Major != tc.expected.Major || result.Minor != tc.expected.Minor || result.Patch != tc.expected.Patch || result.Pre != tc.expected.Pre {
				t.Errorf("version %s parsed incorrectly: expected %+v, got %+v", tc.version, tc.expected, result)
			}
		})
	}
}

// TestCompareVersions 测试版本号比较
func TestCompareVersions(t *testing.T) {
	testCases := []struct {
		v1          string
		v2          string
		expected    int
		expectError bool
	}{
		{"v1.0.0", "v1.0.0", 0, false},
		{"v1.0.1", "v1.0.0", 1, false},
		{"v1.0.0", "v1.0.1", -1, false},
		{"v1.1.0", "v1.0.10", 1, false},
		{"v2.0.0", "v1.999.999", 1, false},
		{"v1.0.0-alpha", "v1.0.0", -1, false},
		{"v1.0.0", "v1.0.0-alpha", 1, false},
		{"v1.0.0-beta", "v1.0.0-alpha", 1, false},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%s vs %s", tc.v1, tc.v2), func(t *testing.T) {
			result, err := compareVersions(tc.v1, tc.v2)
			if tc.expectError {
				if err == nil {
					t.Errorf("expected error for comparison %s vs %s, got none", tc.v1, tc.v2)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error for comparison %s vs %s: %v", tc.v1, tc.v2, err)
				return
			}

			if result != tc.expected {
				t.Errorf("comparison %s vs %s incorrect: expected %d, got %d", tc.v1, tc.v2, tc.expected, result)
			}
		})
	}
}

// TestNeedsUpgrade 测试是否需要升级
func TestNeedsUpgrade(t *testing.T) {
	testCases := []struct {
		current     string
		latest      string
		expected    bool
		expectError bool
	}{
		{"v1.0.0", "v1.0.1", true, false},
		{"v1.0.1", "v1.0.0", false, false},
		{"v1.0.0", "v1.0.0", false, false},
		{"v1.0.0", "v1.1.0", true, false},
		{"v1.0.0", "v2.0.0", true, false},
		{"v1.0.0-alpha", "v1.0.0", true, false},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%s -> %s", tc.current, tc.latest), func(t *testing.T) {
			result, err := needsUpgrade(tc.current, tc.latest)
			if tc.expectError {
				if err == nil {
					t.Errorf("expected error for upgrade check %s -> %s, got none", tc.current, tc.latest)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error for upgrade check %s -> %s: %v", tc.current, tc.latest, err)
				return
			}

			if result != tc.expected {
				t.Errorf("upgrade check %s -> %s incorrect: expected %t, got %t", tc.current, tc.latest, tc.expected, result)
			}
		})
	}
}

// Example 自升级模块使用示例
func Example() {
	// 创建升级配置
	config := &Config{
		AppName:          "myapp",
		CurrentVersion:   "v1.0.0",
		OS:               runtime.GOOS,
		Arch:             runtime.GOARCH,
		UpgradeServerURL: "https://example.com/upgrades",
		Username:         "admin",
		Password:         "password",
		UpgradeOption:    UpgradeOptionImmediate,
		Callback: func(newVersion string) bool {
			fmt.Printf("发现新版本 %s，是否升级？(y/n): ", newVersion)
			// 示例中直接返回true，实际应用中可以让用户确认
			return true
		},
		Logger: func(format string, args ...interface{}) {
			fmt.Printf("[UPGRADE] "+format+"\n", args...)
		},
	}

	// 创建升级器
	upgrader, err := NewUpgrader(config)
	if err != nil {
		fmt.Printf("创建升级器失败: %v\n", err)
		return
	}

	// 检查升级
	versionInfo, err := upgrader.CheckUpgrade()
	if err != nil {
		fmt.Printf("检查升级失败: %v\n", err)
		return
	}

	if versionInfo != nil {
		fmt.Printf("发现新版本: %s\n", versionInfo.Version)
		// 开始升级
		if err := upgrader.StartUpgrade(); err != nil {
			fmt.Printf("升级失败: %v\n", err)
			return
		}
		fmt.Println("升级成功！")
	} else {
		fmt.Println("当前已是最新版本")
	}

	// 输出:
	// [UPGRADE] 开始检查升级...
	// [UPGRADE] 获取版本信息: https://example.com/upgrades/list.txt
	// 当前已是最新版本
}
