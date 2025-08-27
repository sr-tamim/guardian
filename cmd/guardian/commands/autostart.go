package commands

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/sr-tamim/guardian/internal/autostart"
	"github.com/sr-tamim/guardian/pkg/models"
)

// NewAutostartCmd creates the autostart command
func NewAutostartCmd(configLoader func() (*models.Config, error), devMode *bool) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "autostart",
		Short: "Manage automatic startup settings",
		Long:  `Configure Guardian to automatically start when the system boots.`,
	}

	enableCmd := &cobra.Command{
		Use:   "enable",
		Short: "Enable automatic startup",
		RunE: func(cmd *cobra.Command, args []string) error {
			execPath, err := autostart.GetExecutablePath()
			if err != nil {
				return fmt.Errorf("failed to get executable path: %w", err)
			}

			autoStart := autostart.New("Guardian", execPath)
			if err := autoStart.Enable(); err != nil {
				return fmt.Errorf("failed to enable auto-startup: %w", err)
			}

			fmt.Println("✅ Auto-startup enabled successfully")
			return nil
		},
	}

	disableCmd := &cobra.Command{
		Use:   "disable",
		Short: "Disable automatic startup",
		RunE: func(cmd *cobra.Command, args []string) error {
			execPath, err := autostart.GetExecutablePath()
			if err != nil {
				return fmt.Errorf("failed to get executable path: %w", err)
			}

			autoStart := autostart.New("Guardian", execPath)
			if err := autoStart.Disable(); err != nil {
				return fmt.Errorf("failed to disable auto-startup: %w", err)
			}

			fmt.Println("✅ Auto-startup disabled successfully")
			return nil
		},
	}

	statusCmd := &cobra.Command{
		Use:   "status",
		Short: "Check automatic startup status",
		RunE: func(cmd *cobra.Command, args []string) error {
			execPath, err := autostart.GetExecutablePath()
			if err != nil {
				return fmt.Errorf("failed to get executable path: %w", err)
			}

			autoStart := autostart.New("Guardian", execPath)
			if autoStart.IsEnabled() {
				fmt.Println("🟢 Auto-startup: enabled")
			} else {
				fmt.Println("🔴 Auto-startup: disabled")
			}
			return nil
		},
	}

	cmd.AddCommand(enableCmd, disableCmd, statusCmd)
	return cmd
}
