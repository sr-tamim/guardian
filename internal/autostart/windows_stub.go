//go:build !windows
// +build !windows

package autostart

import "fmt"

// windowsAutoStart stub for non-Windows platforms
type windowsAutoStart struct {
	appName  string
	execPath string
}

func (w *windowsAutoStart) Enable() error {
	return fmt.Errorf("auto-startup not supported on this platform")
}

func (w *windowsAutoStart) Disable() error {
	return fmt.Errorf("auto-startup not supported on this platform")
}

func (w *windowsAutoStart) IsEnabled() bool {
	return false
}
