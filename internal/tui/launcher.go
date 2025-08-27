package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/sr-tamim/guardian/internal/core"
)

// StartDashboard launches the Guardian TUI dashboard without system tray support
func StartDashboard(provider core.PlatformProvider, devMode bool) error {
	// Create dashboard
	dashboard := NewDashboard()
	dashboard.SetProvider(provider, devMode)

	// Create and run the TUI program
	program := tea.NewProgram(dashboard, tea.WithAltScreen())

	// Run the program
	if _, err := program.Run(); err != nil {
		return fmt.Errorf("failed to run TUI dashboard: %w", err)
	}

	return nil
}
