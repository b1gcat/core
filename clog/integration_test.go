package clog

import (
	"os"
	"testing"

	"github.com/sirupsen/logrus"
)

// TestInitPlatform 测试平台初始化
func TestInitPlatform(t *testing.T) {
	// 测试平台初始化是否成功
	if logger == nil {
		t.Fatal("Logger should not be nil after initialization")
	}

	// 测试输出设置
	if logger.Out == nil {
		t.Error("Logger output should not be nil")
	}
}

// TestLogLevels 测试各种日志级别的输出
func TestLogLevels(t *testing.T) {
	originalLevel := logger.Level
	defer logger.SetLevel(originalLevel)

	// 设置为debug级别以便测试所有级别
	logger.SetLevel(logrus.DebugLevel)

	// 测试各种日志级别函数是否能正常调用
	Debug("Debug message")
	Debugf("Debug %s", "message")
	Info("Info message")
	Infof("Info %s", "message")
	Warn("Warn message")
	Warnf("Warn %s", "message")
	Error("Error message")
	Errorf("Error %s", "message")
	// 不测试Panic和Fatal，因为它们会导致测试中断
}

// TestNoneLevel 测试none级别是否能禁用所有日志
func TestNoneLevel(t *testing.T) {
	originalLevel := logger.Level
	defer logger.SetLevel(originalLevel)

	// 设置为none级别
	os.Setenv("CLOG_LOGLEVEL", "none")
	setLogLevelFromEnv()

	// 检查日志级别是否高于PanicLevel
	if logger.Level <= logrus.PanicLevel {
		t.Errorf("Expected level > PanicLevel for 'none', got %v", logger.Level)
	}
}
