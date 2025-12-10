package upgrade

import (
	"fmt"
	"os"
	"time"
)

// Upgrader 自升级核心结构体
type Upgrader struct {
	config *Config
}

// NewUpgrader 创建新的自升级实例
func NewUpgrader(config *Config) (*Upgrader, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &Upgrader{
		config: config,
	}, nil
}

// CheckUpgrade 检查是否有新版本
func (u *Upgrader) CheckUpgrade() (*VersionInfo, error) {
	u.config.Logger("开始检查升级...")

	// 检查install工具是否存在
	hasInstallTool, err := checkInstallTool()
	if err != nil {
		return nil, fmt.Errorf("check install tool failed: %w", err)
	}
	if !hasInstallTool {
		return nil, fmt.Errorf("install tool not found")
	}

	// 构造版本信息URL
	versionURL := fmt.Sprintf("%s/list.txt", u.config.UpgradeServerURL)
	u.config.Logger("获取版本信息: %s", versionURL)

	// 获取版本信息
	versionContent, err := fetchWithAuth(versionURL, u.config.Username, u.config.Password)
	if err != nil {
		return nil, fmt.Errorf("fetch version info failed: %w", err)
	}

	// 解析版本信息
	versionInfo := &VersionInfo{
		Version: string(versionContent),
	}

	// 检查是否需要升级
	needUpgrade, err := needsUpgrade(u.config.CurrentVersion, versionInfo.Version)
	if err != nil {
		return nil, fmt.Errorf("compare versions failed: %w", err)
	}

	if !needUpgrade {
		u.config.Logger("当前已是最新版本: %s", u.config.CurrentVersion)
		return nil, nil
	}

	u.config.Logger("发现新版本: %s", versionInfo.Version)
	return versionInfo, nil
}

// StartUpgrade 开始升级
func (u *Upgrader) StartUpgrade() error {
	u.config.Logger("开始升级...")

	// 检查是否有新版本
	versionInfo, err := u.CheckUpgrade()
	if err != nil {
		return fmt.Errorf("check upgrade failed: %w", err)
	}
	if versionInfo == nil {
		return nil // 已是最新版本，无需升级
	}

	// 调用回调确认是否升级
	if !u.config.Callback(versionInfo.Version) {
		u.config.Logger("用户取消升级")
		return nil
	}

	// 获取当前可执行文件路径
	execPath, err := getExecutablePath()
	if err != nil {
		return fmt.Errorf("get executable path failed: %w", err)
	}

	// 构造升级包URL
	upgradePackageName := fmt.Sprintf("%s-%s-%s%s", u.config.AppName, u.config.OS, u.config.Arch, versionInfo.Version)
	if u.config.OS == "windows" {
		upgradePackageName += ".exe"
	}
	upgradePackageURL := fmt.Sprintf("%s/%s", u.config.UpgradeServerURL, upgradePackageName)
	u.config.Logger("下载升级包: %s", upgradePackageURL)

	// 下载升级包到临时目录
	tempFilePath := getTempFilePath(upgradePackageName)
	u.config.Logger("临时文件路径: %s", tempFilePath)

	if err := downloadFileWithAuth(upgradePackageURL, u.config.Username, u.config.Password, tempFilePath); err != nil {
		return fmt.Errorf("download upgrade package failed: %w", err)
	}

	// 验证升级包完整性
	u.config.Logger("验证升级包完整性...")
	upgradeListURL := fmt.Sprintf("%s/upgrade-list.txt", u.config.UpgradeServerURL)
	upgradeListContent, err := fetchWithAuth(upgradeListURL, u.config.Username, u.config.Password)
	if err != nil {
		return fmt.Errorf("fetch upgrade list failed: %w", err)
	}

	upgradeMap, err := parseUpgradeList(upgradeListContent)
	if err != nil {
		return fmt.Errorf("parse upgrade list failed: %w", err)
	}

	expectedHash, exists := upgradeMap[upgradePackageName]
	if !exists {
		return fmt.Errorf("upgrade package not found in upgrade list: %s", upgradePackageName)
	}

	actualHash, err := calculateSHA256(tempFilePath)
	if err != nil {
		return fmt.Errorf("calculate SHA256 failed: %w", err)
	}

	if actualHash != expectedHash {
		return fmt.Errorf("SHA256 check failed: expected %s, got %s", expectedHash, actualHash)
	}

	u.config.Logger("升级包验证通过")

	// 安装升级包
	u.config.Logger("安装升级包...")
	if err := installFile(tempFilePath, execPath); err != nil {
		return fmt.Errorf("install upgrade package failed: %w", err)
	}

	// 清理临时文件
	if err := os.Remove(tempFilePath); err != nil {
		u.config.Logger("清理临时文件失败: %v", err)
	}

	u.config.Logger("升级完成: %s -> %s", u.config.CurrentVersion, versionInfo.Version)
	return nil
}

// SetUpgradeOption 设置升级策略
func (u *Upgrader) SetUpgradeOption(option UpgradeOption) error {
	u.config.UpgradeOption = option
	return nil
}

// SetDailyUpgradeTime 设置每日升级时间
func (u *Upgrader) SetDailyUpgradeTime(hour, minute int) error {
	if hour < 0 || hour > 23 || minute < 0 || minute > 59 {
		return fmt.Errorf("invalid time: hour must be 0-23, minute must be 0-59")
	}
	u.config.DailyUpgradeTime = time.Date(0, 1, 1, hour, minute, 0, 0, time.UTC)
	return nil
}

// UpgradeToVersion 升级到指定版本
func (u *Upgrader) UpgradeToVersion(version string) error {
	u.config.TargetVersion = version
	u.config.UpgradeOption = UpgradeOptionSpecified
	return u.StartUpgrade()
}
