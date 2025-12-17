package upgrade

// UpgradeOption upgrade strategy option
type UpgradeOption int

const (
	// UpgradeOptionImmediate upgrade immediately when new version is found
	UpgradeOptionImmediate UpgradeOption = iota
	// UpgradeOptionDaily upgrade daily at scheduled time
	UpgradeOptionDaily
	// UpgradeOptionSpecified upgrade to specified version
	UpgradeOptionSpecified
)

// VersionInfo version information
type VersionInfo struct {
	Version      string            `json:"version"`
	ReleaseDate  string            `json:"release_date"`
	ReleaseNotes string            `json:"release_notes"`
	UpgradeMap   map[string]string `json:"upgrade_map"` // map of upgrade package names and their hash values
}

// CallbackFunc upgrade confirmation callback function type
type CallbackFunc func(newVersion string) bool

// ProgressCallback download progress callback function type
type ProgressCallback func(downloaded, total int64)
