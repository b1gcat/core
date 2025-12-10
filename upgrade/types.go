package upgrade

// UpgradeOption 升级策略选项
type UpgradeOption int

const (
	// UpgradeOptionImmediate 发现新版本立即升级
	UpgradeOptionImmediate UpgradeOption = iota
	// UpgradeOptionDaily 每天定时升级
	UpgradeOptionDaily
	// UpgradeOptionSpecified 升级到指定版本
	UpgradeOptionSpecified
)

// VersionInfo 版本信息
type VersionInfo struct {
	Version      string `json:"version"`
	ReleaseDate  string `json:"release_date"`
	ReleaseNotes string `json:"release_notes"`
}

// CallbackFunc 升级确认回调函数类型
type CallbackFunc func(newVersion string) bool
