//go:build linux
// +build linux

package autostart

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

type linuxAutoStart struct {
	appName  string
	execPath string
}

func (l *linuxAutoStart) servicePath() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".config", "systemd", "user", fmt.Sprintf("%s.service", l.appName))
}

func (l *linuxAutoStart) Enable() error {
	serviceContent := fmt.Sprintf(`[Unit]
Description=Guardian Intrusion Prevention System
After=network.target

[Service]
Type=simple
ExecStart=%s monitor
Restart=always
RestartSec=10

[Install]
WantedBy=default.target
`, l.execPath)

	servicePath := l.servicePath()

	// Create directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(servicePath), 0755); err != nil {
		return fmt.Errorf("failed to create systemd user directory: %w", err)
	}

	// Write service file
	if err := os.WriteFile(servicePath, []byte(serviceContent), 0644); err != nil {
		return fmt.Errorf("failed to write service file: %w", err)
	}

	// Reload systemd and enable service
	if err := exec.Command("systemctl", "--user", "daemon-reload").Run(); err != nil {
		return fmt.Errorf("failed to reload systemd: %w", err)
	}

	if err := exec.Command("systemctl", "--user", "enable", fmt.Sprintf("%s.service", l.appName)).Run(); err != nil {
		return fmt.Errorf("failed to enable service: %w", err)
	}

	return nil
}

func (l *linuxAutoStart) Disable() error {
	// Disable and remove service
	exec.Command("systemctl", "--user", "disable", fmt.Sprintf("%s.service", l.appName)).Run()

	servicePath := l.servicePath()
	if err := os.Remove(servicePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove service file: %w", err)
	}

	// Reload systemd
	exec.Command("systemctl", "--user", "daemon-reload").Run()

	return nil
}

func (l *linuxAutoStart) IsEnabled() bool {
	cmd := exec.Command("systemctl", "--user", "is-enabled", fmt.Sprintf("%s.service", l.appName))
	err := cmd.Run()
	return err == nil
}
