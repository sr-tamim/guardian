//go:build !linux
// +build !linux

package autostart

import "fmt"

// linuxAutoStart stub for non-Linux platforms
type linuxAutoStart struct {
	appName  string
	execPath string
}

func (l *linuxAutoStart) Enable() error {
	return fmt.Errorf("auto-startup not supported on this platform")
}

func (l *linuxAutoStart) Disable() error {
	return fmt.Errorf("auto-startup not supported on this platform")
}

func (l *linuxAutoStart) IsEnabled() bool {
	return false
}
