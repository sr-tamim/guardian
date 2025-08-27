package commands

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/sr-tamim/guardian/internal/platform"
	"github.com/sr-tamim/guardian/internal/tui"
	"github.com/sr-tamim/guardian/pkg/models"
	"github.com/sr-tamim/guardian/pkg/version"
)

// NewTUICmd creates the tui command
func NewTUICmd(configLoader func() (*models.Config, error), devMode *bool) *cobra.Command {
	return &cobra.Command{
		Use:   "tui",
		Short: "Launch interactive Guardian dashboard",
		Long:  "Launch the Guardian interactive terminal user interface to view daemon status, logs, and system information. This is a read-only dashboard that shows the status of running Guardian daemons.",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Load configuration first
			config, err := configLoader()
			if err != nil {
				return fmt.Errorf("failed to load configuration: %w", err)
			}

			// Create platform provider
			factory := platform.NewFactory()
			provider, err := factory.CreateProvider(*devMode, config)
			if err != nil {
				return fmt.Errorf("failed to create platform provider: %w", err)
			}

			fmt.Printf("🛡️  Guardian v%s Dashboard\n", version.GetVersion())
			fmt.Printf("⚙️  Development mode: %v\n", *devMode)
			fmt.Printf("🖥️  Platform: %s\n", provider.Name())
			fmt.Println("📊 Starting daemon status viewer...")

			// Launch TUI dashboard directly (no system tray)
			return tui.StartDashboard(provider, *devMode)
		},
	}
}
