package upgrade

import (
	"errors"
	"time"
)

// Config 升级模块配置结构体
type Config struct {
	// AppName 应用名称
	AppName string
	// CurrentVersion 当前版本
	CurrentVersion string
	// OS 操作系统
	OS string
	// Arch 架构
	Arch string
	// UpgradeServerURL 升级服务器URL
	UpgradeServerURL string
	// Username 认证用户名
	Username string
	// Password 认证密码
	Password string
	// UpgradeOption 升级策略
	UpgradeOption UpgradeOption
	// DailyUpgradeTime 每日升级时间
	DailyUpgradeTime time.Time
	// TargetVersion 指定升级版本
	TargetVersion string
	// Callback 升级确认回调
	Callback CallbackFunc
	// Logger 日志记录函数
	Logger func(format string, args ...interface{})
}

// Validate 验证配置是否有效
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
		// 默认回调，总是允许升级
		c.Callback = func(newVersion string) bool {
			return true
		}
	}
	if c.Logger == nil {
		// 默认日志函数
		c.Logger = func(format string, args ...interface{}) {
			// 空实现，不输出日志
		}
	}
	return nil
}
