package upgrade

import (
	"fmt"
	"os"
	"strings"
	"time"
)

// Upgrader self-upgrade core structure
type Upgrader struct {
	config *Config
}

// NewUpgrader create new self-upgrade instance
func NewUpgrader(options ...Option) (*Upgrader, error) {
	// Create default configuration
	config := &Config{}

	// Apply all options
	for _, option := range options {
		option(config)
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &Upgrader{
		config: config,
	}, nil
}

// CheckUpgrade check if there's a new version
func (u *Upgrader) CheckUpgrade() (*VersionInfo, error) {
	u.config.Logger("Start checking for updates...")

	// Check if install tool exists
	hasInstallTool, err := checkInstallTool()
	if err != nil {
		return nil, fmt.Errorf("check install tool failed: %w", err)
	}
	if !hasInstallTool {
		return nil, fmt.Errorf("install tool not found")
	}

	// Construct version info URL
	versionURL := fmt.Sprintf("%s/list.txt", u.config.UpgradeServerURL)
	u.config.Logger("Fetch version info: %s", versionURL)

	// Fetch version info
	versionContent, err := fetchWithAuth(versionURL, u.config.Username, u.config.Password)
	if err != nil {
		return nil, fmt.Errorf("fetch version info failed: %w", err)
	}

	// Parse upgrade list file and extract all version numbers
	upgradeMap, err := parseUpgradeList(versionContent)
	if err != nil {
		return nil, fmt.Errorf("parse upgrade list failed: %w", err)
	}

	// Extract latest version number from upgrade list
	var latestVersion string
	for filename := range upgradeMap {
		// Extract version number from filename (format: {os}-{arch}-{name}.v1.0.0)
		versionPrefix := ".v"
		versionIndex := strings.LastIndex(filename, versionPrefix)
		if versionIndex == -1 {
			continue
		}

		// Extract version number part
		version := filename[versionIndex+len(versionPrefix):]
		if version == "" {
			continue
		}

		// Compare versions to find the latest one
		if latestVersion == "" {
			latestVersion = version
		} else {
			cmp, err := compareVersions(version, latestVersion)
			if err == nil && cmp > 0 {
				latestVersion = version
			}
		}
	}

	if latestVersion == "" {
		return nil, fmt.Errorf("no valid version found in upgrade list")
	}

	// Check if upgrade is needed
	needUpgrade, err := needsUpgrade(u.config.CurrentVersion, latestVersion)
	if err != nil {
		return nil, fmt.Errorf("compare versions failed: %w", err)
	}

	if !needUpgrade {
		u.config.Logger("Current is already the latest version: %s", u.config.CurrentVersion)
		return nil, nil
	}

	// Create version info
	versionInfo := &VersionInfo{
		Version:    latestVersion,
		UpgradeMap: upgradeMap,
	}

	u.config.Logger("Found new version: %s", versionInfo.Version)
	return versionInfo, nil
}

// StartUpgrade start the upgrade process
func (u *Upgrader) StartUpgrade() error {
	u.config.Logger("Start upgrading...")

	// Check if there's a new version
	versionInfo, err := u.CheckUpgrade()
	if err != nil {
		return fmt.Errorf("check upgrade failed: %w", err)
	}
	if versionInfo == nil {
		return nil // Already the latest version, no need to upgrade
	}

	// Call callback to confirm upgrade
	if !u.config.Callback(versionInfo.Version) {
		u.config.Logger("User cancelled upgrade")
		return nil
	}

	// Get current executable path
	execPath, err := getExecutablePath()
	if err != nil {
		return fmt.Errorf("get executable path failed: %w", err)
	}

	// Construct upgrade package URL
	appName := u.config.AppName
	upgradePackageName := fmt.Sprintf("%s-%s-%s.v%s", u.config.OS, u.config.Arch, appName, versionInfo.Version)
	upgradePackageURL := fmt.Sprintf("%s/%s", u.config.UpgradeServerURL, upgradePackageName)
	u.config.Logger("Download upgrade package: %s", upgradePackageURL)

	// Download upgrade package to temporary directory
	tempFilePath := getTempFilePath(upgradePackageName)
	defer os.Remove(tempFilePath)

	u.config.Logger("Temporary file path: %s", tempFilePath)

	if err := downloadFileWithAuth(upgradePackageURL, u.config.Username, u.config.Password, tempFilePath, u.config.ProgressCallback); err != nil {
		return fmt.Errorf("download upgrade package failed: %s, %w",
			upgradePackageURL, err)
	}

	// Verify upgrade package integrity
	u.config.Logger("Verify upgrade package integrity...")
	expectedHash, exists := versionInfo.UpgradeMap[upgradePackageName]
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

	// Install upgrade package
	u.config.Logger("Install upgrade package...")
	if err := installFile(tempFilePath, execPath); err != nil {
		return fmt.Errorf("install upgrade package failed: %w", err)
	}

	u.config.Logger("Upgrade completed: %s -> %s", u.config.CurrentVersion, versionInfo.Version)
	return nil
}

// SetUpgradeOption set upgrade strategy
func (u *Upgrader) SetUpgradeOption(option UpgradeOption) error {
	u.config.UpgradeOption = option
	return nil
}

// SetDailyUpgradeTime set daily upgrade time
func (u *Upgrader) SetDailyUpgradeTime(hour, minute int) error {
	if hour < 0 || hour > 23 || minute < 0 || minute > 59 {
		return fmt.Errorf("invalid time: hour must be 0-23, minute must be 0-59")
	}
	u.config.DailyUpgradeTime = time.Date(0, 1, 1, hour, minute, 0, 0, time.UTC)
	return nil
}

// UpgradeToVersion upgrade to specified version
func (u *Upgrader) UpgradeToVersion(version string) error {
	u.config.TargetVersion = version
	u.config.UpgradeOption = UpgradeOptionSpecified
	return u.StartUpgrade()
}
