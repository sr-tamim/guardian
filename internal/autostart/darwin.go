//go:build darwin
// +build darwin

package autostart

import (
	"fmt"
	"os"
	"path/filepath"
)

type darwinAutoStart struct {
	appName  string
	execPath string
}

func (d *darwinAutoStart) plistPath() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, "Library", "LaunchAgents", fmt.Sprintf("com.guardian.%s.plist", d.appName))
}

func (d *darwinAutoStart) Enable() error {
	plistContent := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>Label</key>
	<string>com.guardian.%s</string>
	<key>ProgramArguments</key>
	<array>
		<string>%s</string>
		<string>monitor</string>
	</array>
	<key>RunAtLoad</key>
	<true/>
	<key>KeepAlive</key>
	<true/>
</dict>
</plist>`, d.appName, d.execPath)

	plistPath := d.plistPath()

	// Create directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(plistPath), 0755); err != nil {
		return fmt.Errorf("failed to create LaunchAgents directory: %w", err)
	}

	// Write plist file
	if err := os.WriteFile(plistPath, []byte(plistContent), 0644); err != nil {
		return fmt.Errorf("failed to write plist file: %w", err)
	}

	return nil
}

func (d *darwinAutoStart) Disable() error {
	plistPath := d.plistPath()
	if err := os.Remove(plistPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove plist file: %w", err)
	}
	return nil
}

func (d *darwinAutoStart) IsEnabled() bool {
	_, err := os.Stat(d.plistPath())
	return err == nil
}
