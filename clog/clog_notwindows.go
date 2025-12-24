//go:build !windows

package clog

import "github.com/sirupsen/logrus"

// 非Windows平台的空实现
func initEventLog() {}

func writeToEventLog(level logrus.Level, message string) {}

func closeEventLog() {}

type eventLogHook struct{}

func (h *eventLogHook) Fire(entry *logrus.Entry) error {
	return nil
}

func (h *eventLogHook) Levels() []logrus.Level {
	return nil
}
