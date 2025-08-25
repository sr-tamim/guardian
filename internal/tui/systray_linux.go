//go:build linux
// +build linux

package tui

import (
	"fmt"

	"fyne.io/systray"
	"github.com/sr-tamim/guardian/pkg/version"
)

// setSystrayIcon sets the system tray icon (Linux)
func setSystrayIcon(iconData []byte) {
	if len(iconData) > 0 {
		systray.SetIcon(iconData)
	}
}

// setSystrayTitle sets the system tray title (Linux)
func setSystrayTitle(title string) {
	systray.SetTitle(title)
}

// setSystrayTooltip sets the system tray tooltip (Linux)
func setSystrayTooltip(tooltip string) {
	systray.SetTooltip(tooltip)
}

// initializeTrayDisplay initializes tray appearance for Linux
func initializeTrayDisplay() {
	versionInfo := version.Get()

	// Set tray icon
	iconData := getGuardianIcon()
	setSystrayIcon(iconData)

	// Set title and tooltip
	setSystrayTitle("[G]")
	setSystrayTooltip(fmt.Sprintf("Guardian v%s - Protection Active", versionInfo.Version))
}

// updateTrayTooltip updates the tray tooltip for Linux
func updateTrayTooltip(message string) {
	setSystrayTooltip(message)
}
