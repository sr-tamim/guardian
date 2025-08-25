//go:build windows
// +build windows

package tui

import (
	"fmt"

	"fyne.io/systray"
	"github.com/sr-tamim/guardian/pkg/version"
)

// setSystrayIcon sets the system tray icon (Windows)
func setSystrayIcon(iconData []byte) {
	if len(iconData) > 0 {
		systray.SetIcon(iconData)
	}
}

// setSystrayTitle sets the system tray title (Windows)
func setSystrayTitle(title string) {
	systray.SetTitle(title)
}

// setSystrayTooltip sets the system tray tooltip (Windows)
func setSystrayTooltip(tooltip string) {
	systray.SetTooltip(tooltip)
}

// initializeTrayDisplay initializes tray appearance for Windows
func initializeTrayDisplay() {
	versionInfo := version.Get()

	// Set tray icon
	iconData := getGuardianIcon()
	setSystrayIcon(iconData)

	// Set title and tooltip
	setSystrayTitle("[G]")
	setSystrayTooltip(fmt.Sprintf("Guardian v%s - Protection Active", versionInfo.Version))
}

// updateTrayTooltip updates the tray tooltip for Windows
func updateTrayTooltip(message string) {
	setSystrayTooltip(message)
}
