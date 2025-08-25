//go:build linux
// +build linux

package tui

import (
	"fmt"
	"os"
	"path/filepath"
)

// NoOpStartup provides a fallback implementation for error cases
type NoOpStartup struct {
	err error
}

func (nos *NoOpStartup) Enable() error          { return nos.err }
func (nos *NoOpStartup) Disable() error         { return nos.err }
func (nos *NoOpStartup) IsEnabled() bool        { return false }
func (nos *NoOpStartup) GetDescription() string { return "not supported" }

// LinuxStartup manages Linux systemd user service startup
type LinuxStartup struct {
	executablePath string
	serviceName    string
	userDir        string
}

// createPlatformStartupManager creates Linux-specific startup manager
func createPlatformStartupManager() StartupManager {
	startup, err := NewLinuxStartup()
	if err != nil {
		return &NoOpStartup{err: err}
	}
	return startup
}

// NewLinuxStartup creates a new Linux startup manager
func NewLinuxStartup() (*LinuxStartup, error) {
	execPath, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("failed to get executable path: %w", err)
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home directory: %w", err)
	}

	return &LinuxStartup{
		executablePath: execPath,
		serviceName:    "guardian",
		userDir:        filepath.Join(homeDir, ".config", "systemd", "user"),
	}, nil
}

// Enable adds Guardian to Linux systemd user services
func (ls *LinuxStartup) Enable() error {
	// Implementation would create systemd user service file
	// and enable it with systemctl --user
	return fmt.Errorf("Linux auto-startup not yet implemented")
}

// Disable removes Guardian from Linux systemd user services
func (ls *LinuxStartup) Disable() error {
	// Implementation would disable and remove systemd user service
	return fmt.Errorf("Linux auto-startup not yet implemented")
}

// IsEnabled checks if Guardian systemd user service is enabled
func (ls *LinuxStartup) IsEnabled() bool {
	// Implementation would check systemctl --user status
	return false
}

// GetDescription returns platform-specific description
func (ls *LinuxStartup) GetDescription() string {
	return "systemd user service (~/.config/systemd/user/)"
}
