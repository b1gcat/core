package clog

import (
	"os"
	"runtime"
	"time"

	"github.com/sirupsen/logrus"
)

// Config 日志配置结构体
type Config struct {
	// EventSourceName Windows事件日志的事件源名称，默认为"clog"
	EventSourceName string
	// Logger 自定义logrus.Logger实例，默认为nil（使用内部创建的实例）
	Logger *logrus.Logger
}

// Option 日志配置选项类型
type Option func(*Config)

// WithEventSourceName 设置Windows事件日志的事件源名称
func WithEventSourceName(name string) Option {
	return func(c *Config) {
		c.EventSourceName = name
	}
}

// WithLogger 设置自定义logrus.Logger实例
func WithLogger(l *logrus.Logger) Option {
	return func(c *Config) {
		c.Logger = l
	}
}

var (
	logger     *logrus.Logger
	lastLevel  logrus.Level
	levelCheck time.Duration = time.Second * 5
	config     *Config
)

// Init 使用自定义配置初始化日志系统
// 注意：如果要使用自定义配置，应在程序开始时调用此函数
// 如果不调用此函数，将使用默认配置进行初始化
// 支持函数选项模式，例如：clog.Init(clog.WithEventSourceName("myapp"), clog.WithLogger(customLogger))
func Init(opts ...Option) {
	// 创建默认配置
	cfg := Config{
		EventSourceName: "clog",
	}

	// 应用选项函数
	for _, opt := range opts {
		opt(&cfg)
	}

	// 保存配置
	config = &cfg

	// 如果提供了自定义logger，则使用它
	if cfg.Logger != nil {
		logger = cfg.Logger
	} else {
		// 创建默认logger
		logger = logrus.New()
		logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp: true,
			DisableColors: false,
		})
	}

	// 初始化日志级别
	setLogLevelFromEnv()

	// 根据平台进行初始化
	initPlatform()

	// 启动环境变量监控协程
	go monitorEnvChanges()
}

func init() {
	// 如果用户没有调用Init函数，则使用默认配置
	if logger == nil {
		Init()
	}
}

func initPlatform() {
	platform := runtime.GOOS

	switch platform {
	case "linux", "darwin":
		// Unix-like系统，仅输出到终端
		logger.SetOutput(os.Stdout)
	case "windows":
		// Windows系统，输出到终端和eventlog
		logger.SetOutput(os.Stdout)
		// 添加eventlog hook
		eventSourceName := "clog"
		if config != nil && config.EventSourceName != "" {
			eventSourceName = config.EventSourceName
		}
		initEventLog(eventSourceName)
		logger.AddHook(&eventLogHook{})
	default:
		// 其他平台，默认输出到终端
		logger.SetOutput(os.Stdout)
	}
}

func setLogLevelFromEnv() {
	levelStr := os.Getenv("CLOG_LOGLEVEL")
	var level logrus.Level
	var err error

	switch levelStr {
	case "":
		fallthrough
	case "none":
		// 设置一个高于PanicLevel的级别，这样所有日志都不会输出
		level = logrus.PanicLevel + 1
	default:
		level, err = logrus.ParseLevel(levelStr)
		if err != nil {
			level = logrus.InfoLevel
		}
	}

	logger.SetLevel(level)
	lastLevel = level
}

func monitorEnvChanges() {
	ticker := time.NewTicker(levelCheck)
	defer ticker.Stop()

	for {
		<-ticker.C
		currentLevelStr := os.Getenv("CLOG_LOGLEVEL")
		var currentLevel logrus.Level
		var err error

		switch currentLevelStr {
		case "":
			fallthrough
		case "none":
			currentLevel = logrus.PanicLevel + 1
		default:
			currentLevel, err = logrus.ParseLevel(currentLevelStr)
			if err != nil {
				currentLevel = logrus.InfoLevel
			}
		}

		if currentLevel != lastLevel {
			logger.SetLevel(currentLevel)
			lastLevel = currentLevel
			// 记录日志级别变更
			logger.WithFields(logrus.Fields{
				"old_level": lastLevel,
				"new_level": currentLevel,
			}).Info("日志级别已变更")
		}
	}
}

// Debug logs a message at debug level
func Debug(args ...interface{}) {
	logger.Debug(args...)
}

// Debugf logs a message at debug level with format
func Debugf(format string, args ...interface{}) {
	logger.Debugf(format, args...)
}

// Info logs a message at info level
func Info(args ...interface{}) {
	logger.Info(args...)
}

// Infof logs a message at info level with format
func Infof(format string, args ...interface{}) {
	logger.Infof(format, args...)
}

// Warn logs a message at warn level
func Warn(args ...interface{}) {
	logger.Warn(args...)
}

// Warnf logs a message at warn level with format
func Warnf(format string, args ...interface{}) {
	logger.Warnf(format, args...)
}

// Error logs a message at error level
func Error(args ...interface{}) {
	logger.Error(args...)
}

// Errorf logs a message at error level with format
func Errorf(format string, args ...interface{}) {
	logger.Errorf(format, args...)
}

// Fatal logs a message at fatal level and exits
func Fatal(args ...interface{}) {
	logger.Fatal(args...)
}

// Fatalf logs a message at fatal level with format and exits
func Fatalf(format string, args ...interface{}) {
	logger.Fatalf(format, args...)
}

// Panic logs a message at panic level and panics
func Panic(args ...interface{}) {
	logger.Panic(args...)
}

// Panicf logs a message at panic level with format and panics
func Panicf(format string, args ...interface{}) {
	logger.Panicf(format, args...)
}
