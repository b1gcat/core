package upgrade

import (
	"fmt"
	"strconv"
	"strings"
)

// VersionComponents 版本号组件
type VersionComponents struct {
	Major int
	Minor int
	Patch int
	Pre   string
}

// parseVersion 解析版本号字符串为组件
func parseVersion(version string) (*VersionComponents, error) {
	// 移除版本号前缀
	// 先检查是否以"version"或"Version"开头
	if strings.HasPrefix(strings.ToLower(version), "version") {
		version = version[7:] // 跳过"version"这7个字符
	}
	// 再移除"v"或"V"前缀
	version = strings.TrimPrefix(version, "v")
	version = strings.TrimPrefix(version, "V")

	// 处理预发布版本
	pre := ""
	if idx := strings.IndexAny(version, "-_"); idx != -1 {
		pre = version[idx+1:]
		version = version[:idx]
	}

	// 分割版本号组件
	parts := strings.Split(version, ".")
	if len(parts) < 1 || len(parts) > 3 {
		return nil, fmt.Errorf("invalid version format: %s", version)
	}

	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return nil, fmt.Errorf("invalid major version: %s", parts[0])
	}

	minor := 0
	if len(parts) > 1 {
		minor, err = strconv.Atoi(parts[1])
		if err != nil {
			return nil, fmt.Errorf("invalid minor version: %s", parts[1])
		}
	}

	patch := 0
	if len(parts) > 2 {
		// 处理可能的构建信息
		patchPart := parts[2]
		if idx := strings.IndexAny(patchPart, "+_"); idx != -1 {
			patchPart = patchPart[:idx]
		}
		patch, err = strconv.Atoi(patchPart)
		if err != nil {
			return nil, fmt.Errorf("invalid patch version: %s", parts[2])
		}
	}

	return &VersionComponents{
		Major: major,
		Minor: minor,
		Patch: patch,
		Pre:   pre,
	}, nil
}

// compareVersions 比较两个版本号，返回1如果v1 > v2，0如果相等，-1如果v1 < v2
func compareVersions(v1, v2 string) (int, error) {
	vc1, err := parseVersion(v1)
	if err != nil {
		return 0, err
	}

	vc2, err := parseVersion(v2)
	if err != nil {
		return 0, err
	}

	// 比较主版本号
	if vc1.Major > vc2.Major {
		return 1, nil
	} else if vc1.Major < vc2.Major {
		return -1, nil
	}

	// 比较次版本号
	if vc1.Minor > vc2.Minor {
		return 1, nil
	} else if vc1.Minor < vc2.Minor {
		return -1, nil
	}

	// 比较补丁版本号
	if vc1.Patch > vc2.Patch {
		return 1, nil
	} else if vc1.Patch < vc2.Patch {
		return -1, nil
	}

	// 比较预发布版本
	if vc1.Pre == "" && vc2.Pre != "" {
		return 1, nil // 正式版本高于预发布版本
	} else if vc1.Pre != "" && vc2.Pre == "" {
		return -1, nil
	} else if vc1.Pre > vc2.Pre {
		return 1, nil
	} else if vc1.Pre < vc2.Pre {
		return -1, nil
	}

	return 0, nil
}

// needsUpgrade 检查是否需要升级
func needsUpgrade(currentVersion, latestVersion string) (bool, error) {
	cmp, err := compareVersions(latestVersion, currentVersion)
	if err != nil {
		return false, err
	}
	return cmp > 0, nil
}
