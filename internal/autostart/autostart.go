package autostart

import (
	"fmt"
	"os"
	"runtime"
)

// AutoStart handles automatic startup configuration
type AutoStart interface {
	Enable() error
	Disable() error
	IsEnabled() bool
}

// New creates a new AutoStart manager for the current platform
func New(appName string, execPath string) AutoStart {
	switch runtime.GOOS {
	case "windows":
		return &windowsAutoStart{appName: appName, execPath: execPath}
	case "darwin":
		return &darwinAutoStart{appName: appName, execPath: execPath}
	case "linux":
		return &linuxAutoStart{appName: appName, execPath: execPath}
	default:
		return &unsupportedAutoStart{}
	}
}

// GetExecutablePath returns the current executable path
func GetExecutablePath() (string, error) {
	execPath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("failed to get executable path: %w", err)
	}
	return execPath, nil
}

// unsupportedAutoStart provides a no-op implementation for unsupported platforms
type unsupportedAutoStart struct{}

func (u *unsupportedAutoStart) Enable() error {
	return fmt.Errorf("auto-startup not supported on %s", runtime.GOOS)
}

func (u *unsupportedAutoStart) Disable() error {
	return fmt.Errorf("auto-startup not supported on %s", runtime.GOOS)
}

func (u *unsupportedAutoStart) IsEnabled() bool {
	return false
}
