package tui

// getGuardianIcon returns nil to force systray to use title-based display
func getGuardianIcon() []byte {
	// Return nil to force systray to display the title text instead of icon
	// This avoids ICO format issues and ensures visibility
	return nil
}

// getSimpleGuardianIcon returns a Unicode shield character as fallback
func getSimpleGuardianIcon() []byte {
	// If ICO fails, this could be used as text-based fallback
	// But systray.SetIcon expects binary icon data
	return []byte("🛡️")
}
