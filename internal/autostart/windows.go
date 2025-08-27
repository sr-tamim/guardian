//go:build windows
// +build windows

package autostart

import (
	"fmt"

	"golang.org/x/sys/windows/registry"
)

type windowsAutoStart struct {
	appName  string
	execPath string
}

func (w *windowsAutoStart) Enable() error {
	key, err := registry.OpenKey(registry.CURRENT_USER, `SOFTWARE\Microsoft\Windows\CurrentVersion\Run`, registry.SET_VALUE)
	if err != nil {
		return fmt.Errorf("failed to open registry key: %w", err)
	}
	defer key.Close()

	command := fmt.Sprintf(`"%s" monitor`, w.execPath)
	err = key.SetStringValue(w.appName, command)
	if err != nil {
		return fmt.Errorf("failed to set registry value: %w", err)
	}

	return nil
}

func (w *windowsAutoStart) Disable() error {
	key, err := registry.OpenKey(registry.CURRENT_USER, `SOFTWARE\Microsoft\Windows\CurrentVersion\Run`, registry.SET_VALUE)
	if err != nil {
		return nil // Key doesn't exist, already disabled
	}
	defer key.Close()

	err = key.DeleteValue(w.appName)
	if err != nil && err != registry.ErrNotExist {
		return fmt.Errorf("failed to delete registry value: %w", err)
	}

	return nil
}

func (w *windowsAutoStart) IsEnabled() bool {
	key, err := registry.OpenKey(registry.CURRENT_USER, `SOFTWARE\Microsoft\Windows\CurrentVersion\Run`, registry.QUERY_VALUE)
	if err != nil {
		return false
	}
	defer key.Close()

	_, _, err = key.GetStringValue(w.appName)
	return err == nil
}
