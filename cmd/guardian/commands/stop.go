package commands

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/sr-tamim/guardian/internal/daemon"
	"github.com/sr-tamim/guardian/internal/platform"
	"github.com/sr-tamim/guardian/pkg/models"
)

// NewStopCmd creates the stop command
func NewStopCmd(configLoader func() (*models.Config, error), devMode *bool) *cobra.Command {
	return &cobra.Command{
		Use:   "stop",
		Short: "Stop the Guardian daemon",
		Long:  "Stop the Guardian daemon that is running in the background.",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Load configuration
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

			// Create daemon manager
			daemonManager := daemon.NewManager(config, provider, *devMode)

			// Stop the daemon
			return daemonManager.StopDaemon()
		},
	}
}
