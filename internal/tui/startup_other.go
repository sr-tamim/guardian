//go:build !windows && !linux
// +build !windows,!linux

package tui

import "fmt"

// NoOpStartup provides no-op startup functionality for unsupported platforms
type NoOpStartup struct {
	err error
}

// createPlatformStartupManager creates no-op startup manager for unsupported platforms
func createPlatformStartupManager() StartupManager {
	return &NoOpStartup{
		err: fmt.Errorf("auto-startup not supported on this platform"),
	}
}

// Enable does nothing on unsupported platforms
func (nos *NoOpStartup) Enable() error {
	return nos.err
}

// Disable does nothing on unsupported platforms
func (nos *NoOpStartup) Disable() error {
	return nos.err
}

// IsEnabled returns false on unsupported platforms
func (nos *NoOpStartup) IsEnabled() bool {
	return false
}

// GetDescription returns platform-specific description
func (nos *NoOpStartup) GetDescription() string {
	return "Not supported on this platform"
}
