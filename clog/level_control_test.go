package clog

import (
	"os"
	"testing"

	"github.com/sirupsen/logrus"
)

// TestSetLogLevelFromEnv 测试从环境变量设置日志级别
func TestSetLogLevelFromEnv(t *testing.T) {
	// 测试默认情况
	originalLevel := logger.Level
	defer logger.SetLevel(originalLevel)

	// 测试空环境变量
	os.Unsetenv("CLOG_LOGLEVEL")
	setLogLevelFromEnv()
	if logger.Level != logrus.InfoLevel {
		t.Errorf("Expected InfoLevel, got %v", logger.Level)
	}

	// 测试debug级别
	os.Setenv("CLOG_LOGLEVEL", "debug")
	setLogLevelFromEnv()
	if logger.Level != logrus.DebugLevel {
		t.Errorf("Expected DebugLevel, got %v", logger.Level)
	}

	// 测试info级别
	os.Setenv("CLOG_LOGLEVEL", "info")
	setLogLevelFromEnv()
	if logger.Level != logrus.InfoLevel {
		t.Errorf("Expected InfoLevel, got %v", logger.Level)
	}

	// 测试warn级别
	os.Setenv("CLOG_LOGLEVEL", "warn")
	setLogLevelFromEnv()
	if logger.Level != logrus.WarnLevel {
		t.Errorf("Expected WarnLevel, got %v", logger.Level)
	}

	// 测试error级别
	os.Setenv("CLOG_LOGLEVEL", "error")
	setLogLevelFromEnv()
	if logger.Level != logrus.ErrorLevel {
		t.Errorf("Expected ErrorLevel, got %v", logger.Level)
	}

	// 测试none级别
	os.Setenv("CLOG_LOGLEVEL", "none")
	setLogLevelFromEnv()
	if logger.Level <= logrus.PanicLevel {
		t.Errorf("Expected level > PanicLevel, got %v", logger.Level)
	}

	// 测试无效级别
	os.Setenv("CLOG_LOGLEVEL", "invalid")
	setLogLevelFromEnv()
	if logger.Level != logrus.InfoLevel {
		t.Errorf("Expected InfoLevel for invalid, got %v", logger.Level)
	}
}
