//go:build !windows
// +build !windows

package service

func writeEventLog(level string, message string) {
	_ = level
	_ = message
}
