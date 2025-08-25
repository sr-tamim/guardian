//go:build windows
// +build windows

package tui

import (
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/sys/windows/registry"
)

// WindowsStartup manages Windows startup registry entries
type WindowsStartup struct {
	executablePath string
	appName        string
}

// createPlatformStartupManager creates Windows-specific startup manager
func createPlatformStartupManager() StartupManager {
	startup, err := NewWindowsStartup()
	if err != nil {
		return &NoOpStartup{err: err}
	}
	return startup
}

// NoOpStartup provides fallback functionality when Windows startup fails
type NoOpStartup struct {
	err error
}

// Enable does nothing when startup manager fails to initialize
func (nos *NoOpStartup) Enable() error {
	return nos.err
}

// Disable does nothing when startup manager fails to initialize
func (nos *NoOpStartup) Disable() error {
	return nos.err
}

// IsEnabled returns false when startup manager fails to initialize
func (nos *NoOpStartup) IsEnabled() bool {
	return false
}

// GetDescription returns error description
func (nos *NoOpStartup) GetDescription() string {
	return fmt.Sprintf("Startup unavailable: %v", nos.err)
}

// GetDescription returns platform-specific description
func (ws *WindowsStartup) GetDescription() string {
	return "Windows Registry (HKEY_CURRENT_USER\\...\\Run)"
}

// NewWindowsStartup creates a new Windows startup manager
func NewWindowsStartup() (*WindowsStartup, error) {
	execPath, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("failed to get executable path: %w", err)
	}

	return &WindowsStartup{
		executablePath: execPath,
		appName:        "Guardian",
	}, nil
}

// Enable adds Guardian to Windows startup
func (ws *WindowsStartup) Enable() error {
	key, err := registry.OpenKey(registry.CURRENT_USER, `SOFTWARE\Microsoft\Windows\CurrentVersion\Run`, registry.SET_VALUE)
	if err != nil {
		return fmt.Errorf("failed to open registry key: %w", err)
	}
	defer key.Close()

	// Add the executable with --tui flag for automatic TUI startup
	commandLine := fmt.Sprintf(`"%s" --tui`, ws.executablePath)

	err = key.SetStringValue(ws.appName, commandLine)
	if err != nil {
		return fmt.Errorf("failed to set registry value: %w", err)
	}

	return nil
}

// Disable removes Guardian from Windows startup
func (ws *WindowsStartup) Disable() error {
	key, err := registry.OpenKey(registry.CURRENT_USER, `SOFTWARE\Microsoft\Windows\CurrentVersion\Run`, registry.SET_VALUE)
	if err != nil {
		return fmt.Errorf("failed to open registry key: %w", err)
	}
	defer key.Close()

	err = key.DeleteValue(ws.appName)
	if err != nil {
		// If value doesn't exist, that's fine
		if err == registry.ErrNotExist {
			return nil
		}
		return fmt.Errorf("failed to delete registry value: %w", err)
	}

	return nil
}

// IsEnabled checks if Guardian is set to start with Windows
func (ws *WindowsStartup) IsEnabled() bool {
	key, err := registry.OpenKey(registry.CURRENT_USER, `SOFTWARE\Microsoft\Windows\CurrentVersion\Run`, registry.QUERY_VALUE)
	if err != nil {
		return false
	}
	defer key.Close()

	_, _, err = key.GetStringValue(ws.appName)
	return err == nil
}

// GetExecutablePath returns the current executable path
func (ws *WindowsStartup) GetExecutablePath() string {
	return ws.executablePath
}

// GetInstallLocation returns suggested install location for Guardian
func (ws *WindowsStartup) GetInstallLocation() string {
	programFiles := os.Getenv("PROGRAMFILES")
	if programFiles == "" {
		programFiles = "C:\\Program Files"
	}
	return filepath.Join(programFiles, "Guardian")
}
