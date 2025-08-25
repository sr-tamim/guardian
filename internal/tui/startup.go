package tui

// StartupManager defines platform-specific startup functionality
type StartupManager interface {
	Enable() error
	Disable() error
	IsEnabled() bool
	GetDescription() string
}

// CreateStartupManager creates platform-appropriate startup manager
func CreateStartupManager() StartupManager {
	// This will be implemented with build tags for each platform
	return createPlatformStartupManager()
}
