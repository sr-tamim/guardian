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
		Short: "Launch interactive TUI dashboard with system tray support",
		Long:  "Launch the Guardian interactive terminal user interface with background monitoring and system tray integration. Minimizes to tray when closed to keep protection active.",
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

			// Create service manager with TUI and tray support
			serviceManager := tui.NewServiceManager(provider, *devMode)

			fmt.Printf("🛡️  Guardian v%s with TUI & System Tray\n", version.GetVersion())
			fmt.Printf("⚙️  Development mode: %v\n", *devMode)
			fmt.Printf("🖥️  Platform: %s\n", provider.Name())
			fmt.Println("✨ Starting with system tray integration...")

			return serviceManager.StartWithTraySupport()
		},
	}
}
