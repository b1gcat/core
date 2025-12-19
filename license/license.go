package license

import (
	"encoding/binary"
	"net"
	"strconv"
	"time"
)

// Config holds all configurable parameters for the license system
type Config struct {
	ntpServers         []string
	expirationDays     int
	expirationCallback func()
}

// WithOption is a function type for configuring the license system
type WithOption func(*Config)

// Default configuration
var defaultConfig = Config{
	ntpServers: []string{
		// 国内权威节点
		"210.72.145.44:123",  // 中科院国家授时中心（陕西临潼）核心节点
		"203.107.6.88:123",   // 国家授时中心备用节点 / 阿里云
		"139.224.11.107:123", // 中国计量科学研究院
		"120.25.115.20:123",  // 阿里云备用
		"103.75.192.24:123",  // 腾讯云
		"219.141.136.10:123", // 电信公用
		"202.106.0.20:123",   // 联通公用
		// 国际备用节点
		"pool.ntp.org:123",     // 全球分布式NTP池
		"time.windows.com:123", // 微软公共NTP
	},
	expirationDays: 0, // Default to permanent
}

// Global config instance
var globalConfig = defaultConfig

// WithExpiration sets the license expiration days
func WithExpiration(days int) WithOption {
	return func(cfg *Config) {
		cfg.expirationDays = days
	}
}

// WithExpirationCallback sets the callback function to be executed when license expires
func WithExpirationCallback(callback func()) WithOption {
	return func(cfg *Config) {
		cfg.expirationCallback = callback
	}
}

// WithNTPServers sets custom NTP servers
func WithNTPServers(servers []string) WithOption {
	return func(cfg *Config) {
		cfg.ntpServers = servers
	}
}

// New creates a new Config with the specified options
func New(opts ...WithOption) Config {
	cfg := defaultConfig
	for _, opt := range opts {
		opt(&cfg)
	}
	return cfg
}

// getNTPTime 从NTP服务器获取当前时间
// 如果所有NTP服务器都不可用，返回系统时间
func getNTPTime() time.Time {
	// 尝试从每个NTP服务器获取时间
	for _, server := range globalConfig.ntpServers {
		ntpTime, err := getTimeFromNTP(server)
		if err == nil {
			return ntpTime
		}
	}

	// 所有NTP服务器都不可用，返回系统时间
	return time.Now()
}

// getTimeFromNTP 从单个NTP服务器获取时间
func getTimeFromNTP(server string) (time.Time, error) {
	// 创建UDP连接
	conn, err := net.DialUDP("udp", nil, &net.UDPAddr{IP: net.ParseIP(server[:len(server)-4]), Port: 123})
	if err != nil {
		return time.Time{}, err
	}
	defer conn.Close()

	// 设置超时
	conn.SetDeadline(time.Now().Add(5 * time.Second))

	// 创建NTP请求包
	// NTP包结构：48字节，第一字节为版本号3，模式3（客户端）
	req := make([]byte, 48)
	req[0] = 0x1B // 00 011 011 (LI=0, VN=3, Mode=3)

	// 发送请求
	_, err = conn.Write(req)
	if err != nil {
		return time.Time{}, err
	}

	// 接收响应
	resp := make([]byte, 48)
	_, err = conn.Read(resp)
	if err != nil {
		return time.Time{}, err
	}

	// 解析响应
	// NTP时间戳是从1900年1月1日开始的秒数
	// 转换为Unix时间戳（从1970年1月1日开始的秒数）
	secs := binary.BigEndian.Uint32(resp[40:44])
	frac := binary.BigEndian.Uint32(resp[44:48])

	// 转换为Unix时间
	ntpTime := time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC)
	ntpTime = ntpTime.Add(time.Duration(secs) * time.Second)
	// 使用float64避免整数溢出
	fracNano := float64(frac) * 1e9 / float64(0x100000000)
	ntpTime = ntpTime.Add(time.Duration(fracNano) * time.Nanosecond)

	// 转换为Unix时间戳
	timestamp := ntpTime.Unix()

	// 返回当前时间
	return time.Unix(timestamp, 0), nil
}

// buildTime 是编译时注入的时间戳，格式为Unix秒
// 通过构建参数设置: go build -ldflags "-X github.com/b1gcat/core/license.buildTime=$(date +%s)"
var buildTime string

// SetExpiration 设置授权的天数 (backward compatible)
func SetExpiration(days int) {
	globalConfig.expirationDays = days
}

// CheckLicense 检查授权是否有效
// 如果当前时间超过编译时间加上授权天数，则执行过期回调并退出程序
func CheckLicense() {
	if globalConfig.expirationDays <= 0 {
		// 授权天数未设置，默认为永久有效
		return
	}

	// 解析编译时间
	buildSec, err := strconv.ParseInt(buildTime, 10, 64)
	if err != nil {
		// 无法解析编译时间，在测试环境中跳过检查
		// 生产环境中应该确保buildTime被正确注入
		return
	}

	// 获取当前时间（优先使用NTP时间）
	now := getNTPTime()
	buildTimeObj := time.Unix(buildSec, 0)

	// 计算过期时间
	expirationTime := buildTimeObj.AddDate(0, 0, globalConfig.expirationDays)

	// 检查是否过期
	if now.After(expirationTime) {
		// 执行过期回调（如果已设置）
		if globalConfig.expirationCallback != nil {
			globalConfig.expirationCallback()
		}
	}
}
