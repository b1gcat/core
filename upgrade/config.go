package upgrade

import (
	"errors"
	"time"
)

// Option configuration option function type
type Option func(*Config)

// WithAppName set application name
func WithAppName(appName string) Option {
	return func(c *Config) {
		c.AppName = appName
	}
}

// WithCurrentVersion set current version
func WithCurrentVersion(version string) Option {
	return func(c *Config) {
		c.CurrentVersion = version
	}
}

// WithOS set operating system
func WithOS(os string) Option {
	return func(c *Config) {
		c.OS = os
	}
}

// WithArch set architecture
func WithArch(arch string) Option {
	return func(c *Config) {
		c.Arch = arch
	}
}

// WithUpgradeServerURL set upgrade server URL
func WithUpgradeServerURL(url string) Option {
	return func(c *Config) {
		c.UpgradeServerURL = url
	}
}

// WithUsername set authentication username
func WithUsername(username string) Option {
	return func(c *Config) {
		c.Username = username
	}
}

// WithPassword set authentication password
func WithPassword(password string) Option {
	return func(c *Config) {
		c.Password = password
	}
}

// WithUpgradeOption set upgrade strategy
func WithUpgradeOption(option UpgradeOption) Option {
	return func(c *Config) {
		c.UpgradeOption = option
	}
}

// WithDailyUpgradeTime set daily upgrade time
func WithDailyUpgradeTime(t time.Time) Option {
	return func(c *Config) {
		c.DailyUpgradeTime = t
	}
}

// WithTargetVersion set target upgrade version
func WithTargetVersion(version string) Option {
	return func(c *Config) {
		c.TargetVersion = version
	}
}

// WithCallback set upgrade confirmation callback
func WithCallback(callback CallbackFunc) Option {
	return func(c *Config) {
		c.Callback = callback
	}
}

// WithLogger set logging function
func WithLogger(logger func(format string, args ...interface{})) Option {
	return func(c *Config) {
		c.Logger = logger
	}
}

// WithProgressCallback set download progress callback
func WithProgressCallback(progressCallback ProgressCallback) Option {
	return func(c *Config) {
		c.ProgressCallback = progressCallback
	}
}

// Config upgrade module configuration struct
type Config struct {
	// AppName application name
	AppName string
	// CurrentVersion current version
	CurrentVersion string
	// OS operating system
	OS string
	// Arch architecture
	Arch string
	// UpgradeServerURL upgrade server URL
	UpgradeServerURL string
	// Username authentication username
	Username string
	// Password authentication password
	Password string
	// UpgradeOption upgrade strategy
	UpgradeOption UpgradeOption
	// DailyUpgradeTime daily upgrade time
	DailyUpgradeTime time.Time
	// TargetVersion target upgrade version
	TargetVersion string
	// Callback upgrade confirmation callback
	Callback CallbackFunc
	// Logger logging function
	Logger func(format string, args ...interface{})
	// ProgressCallback download progress callback
	ProgressCallback ProgressCallback
}

// Validate validate if the configuration is valid
func (c *Config) Validate() error {
	if c.AppName == "" {
		return errors.New("app name is required")
	}
	if c.CurrentVersion == "" {
		return errors.New("current version is required")
	}
	if c.OS == "" {
		return errors.New("os is required")
	}
	if c.Arch == "" {
		return errors.New("arch is required")
	}
	if c.UpgradeServerURL == "" {
		return errors.New("upgrade server URL is required")
	}
	if c.Username == "" || c.Password == "" {
		return errors.New("username and password are required")
	}
	if c.Callback == nil {
		// default callback, always allows upgrade
		c.Callback = func(newVersion string) bool {
			return true
		}
	}
	if c.Logger == nil {
		// default logging function
		c.Logger = func(format string, args ...interface{}) {
			// empty implementation, no log output
		}
	}
	return nil
}
