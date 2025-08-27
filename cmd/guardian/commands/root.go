package commands

import (
	"github.com/spf13/cobra"
	"github.com/sr-tamim/guardian/pkg/models"
	"github.com/sr-tamim/guardian/pkg/version"
)

// NewRootCmd creates the root command
func NewRootCmd(configLoader func() (*models.Config, error), devMode *bool) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "guardian",
		Short:   "Guardian - Modern Cross-Platform Intrusion Prevention System",
		Long:    `Guardian is a modern, cross-platform intrusion prevention system that monitors log files and automatically blocks malicious IP addresses. Built as a contemporary alternative to fail2ban with a beautiful terminal interface and intelligent threat detection.`,
		Version: version.GetVersion(),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Default behavior when no subcommand is provided (e.g., double-click)
			// This automatically launches the TUI interface
			tuiCmd := NewTUICmd(configLoader, devMode)
			return tuiCmd.RunE(cmd, args)
		},
	}

	return cmd
}
