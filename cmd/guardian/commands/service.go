package commands

import (
	"fmt"

	"github.com/kardianos/service"
	"github.com/spf13/cobra"
	appservice "github.com/sr-tamim/guardian/internal/service"
	"github.com/sr-tamim/guardian/pkg/models"
)

// NewServiceCmd creates the service command
func NewServiceCmd(configLoader func() (*models.Config, error), devMode *bool, configPath *string) *cobra.Command {
	manager := appservice.NewManager(configLoader, devMode, configPath)

	cmd := &cobra.Command{
		Use:   "service",
		Short: "Manage Guardian as a system service",
		Long:  "Install, start, stop, and remove the Guardian service.",
	}

	installCmd := &cobra.Command{
		Use:   "install",
		Short: "Install the Guardian service",
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, err := manager.NewService()
			if err != nil {
				return err
			}
			return svc.Install()
		},
	}

	uninstallCmd := &cobra.Command{
		Use:   "uninstall",
		Short: "Uninstall the Guardian service",
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, err := manager.NewService()
			if err != nil {
				return err
			}
			return svc.Uninstall()
		},
	}

	startCmd := &cobra.Command{
		Use:   "start",
		Short: "Start the Guardian service",
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, err := manager.NewService()
			if err != nil {
				return err
			}
			return svc.Start()
		},
	}

	stopCmd := &cobra.Command{
		Use:   "stop",
		Short: "Stop the Guardian service",
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, err := manager.NewService()
			if err != nil {
				return err
			}
			return svc.Stop()
		},
	}

	statusCmd := &cobra.Command{
		Use:   "status",
		Short: "Check the Guardian service status",
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, err := manager.NewService()
			if err != nil {
				return err
			}
			status, err := svc.Status()
			if err != nil {
				return err
			}

			switch status {
			case service.StatusRunning:
				fmt.Println("🟢 Service status: running")
			case service.StatusStopped:
				fmt.Println("🔴 Service status: stopped")
			default:
				fmt.Println("🟡 Service status: unknown")
			}
			return nil
		},
	}

	runCmd := &cobra.Command{
		Use:    "run",
		Short:  "Run the service (internal)",
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, err := manager.NewService()
			if err != nil {
				return err
			}
			return svc.Run()
		},
	}

	cmd.AddCommand(installCmd, uninstallCmd, startCmd, stopCmd, statusCmd, runCmd)
	return cmd
}
