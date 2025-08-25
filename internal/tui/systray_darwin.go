//go:build darwin
// +build darwin

package tui

import (
	"fmt"
)

// setSystrayIcon sets the system tray icon (macOS - no-op if unsupported)
func setSystrayIcon(iconData []byte) {
	// macOS might not support these functions in this systray version
	// Provide no-op implementation
}

// setSystrayTitle sets the system tray title (macOS - no-op if unsupported)
func setSystrayTitle(title string) {
	// macOS might not support these functions in this systray version
	// Provide no-op implementation
}

// setSystrayTooltip sets the system tray tooltip (macOS - no-op if unsupported)
func setSystrayTooltip(tooltip string) {
	// macOS might not support these functions in this systray version
	// Provide no-op implementation
}

// initializeTrayDisplay initializes tray appearance for macOS
func initializeTrayDisplay() {
	// macOS systray initialization - simplified
	fmt.Println("Guardian system tray active (macOS)")
}

// updateTrayTooltip updates the tray tooltip for macOS
func updateTrayTooltip(message string) {
	// No-op for macOS compatibility
}
