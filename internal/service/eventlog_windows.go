//go:build windows
// +build windows

package service

import "golang.org/x/sys/windows/svc/eventlog"

func writeEventLog(level string, message string) {
	log, err := eventlog.Open("Guardian")
	if err != nil {
		_ = eventlog.InstallAsEventCreate("Guardian", eventlog.Error|eventlog.Warning|eventlog.Info)
		log, err = eventlog.Open("Guardian")
		if err != nil {
			return
		}
	}
	defer log.Close()

	switch level {
	case "error":
		_ = log.Error(1, message)
	case "warn":
		_ = log.Warning(1, message)
	default:
		_ = log.Info(1, message)
	}
}
