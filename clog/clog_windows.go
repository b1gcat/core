//go:build windows

package clog

import (
	"github.com/sirupsen/logrus"
	"golang.org/x/sys/windows/svc/eventlog"
)

var eventLogHandle *eventlog.Log

func initEventLog() {
	// 注册事件源
	log, err := eventlog.Open("clog")
	if err == nil {
		eventLogHandle = log
	}
}

func writeToEventLog(level logrus.Level, message string) {
	if eventLogHandle == nil {
		return
	}

	// Event ID范围1-1000（符合EventCreate.exe的要求）
	var eventID uint32 = 1

	switch level {
	case logrus.PanicLevel, logrus.FatalLevel, logrus.ErrorLevel:
		eventLogHandle.Error(eventID, message)
	case logrus.WarnLevel:
		eventLogHandle.Warning(eventID, message)
	default:
		eventLogHandle.Info(eventID, message)
	}
}

func closeEventLog() {
	if eventLogHandle != nil {
		eventLogHandle.Close()
		eventLogHandle = nil
	}
}

// Windows平台的EventLog Hook

type eventLogHook struct{}

func (h *eventLogHook) Fire(entry *logrus.Entry) error {
	message, err := entry.String()
	if err != nil {
		return err
	}
	writeToEventLog(entry.Level, message)
	return nil
}

func (h *eventLogHook) Levels() []logrus.Level {
	return logrus.AllLevels
}
