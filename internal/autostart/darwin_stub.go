//go:build !darwin
// +build !darwin

package autostart

import "fmt"

// darwinAutoStart stub for non-Darwin platforms
type darwinAutoStart struct {
	appName  string
	execPath string
}

func (d *darwinAutoStart) Enable() error {
	return fmt.Errorf("auto-startup not supported on this platform")
}

func (d *darwinAutoStart) Disable() error {
	return fmt.Errorf("auto-startup not supported on this platform")
}

func (d *darwinAutoStart) IsEnabled() bool {
	return false
}
