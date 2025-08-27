package commands

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/sr-tamim/guardian/pkg/version"
)

// NewVersionCmd creates the version command
func NewVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show detailed version information",
		Long:  "Display detailed version information including build time and git commit.",
		RunE: func(cmd *cobra.Command, args []string) error {
			versionInfo := version.Get()
			fmt.Println(versionInfo.String())
			fmt.Printf("\n📦 Build Details:\n")
			fmt.Printf("   Version: %s\n", versionInfo.Version)
			fmt.Printf("   Git Commit: %s\n", versionInfo.GitCommit)
			fmt.Printf("   Build Time: %s\n", version.GetBuildTime())
			fmt.Printf("   Go Version: %s\n", versionInfo.GoVersion)
			fmt.Printf("   Platform: %s/%s\n", versionInfo.Platform, versionInfo.Arch)
			if version.IsDevelopment() {
				fmt.Printf("   Build Type: Development\n")
			} else {
				fmt.Printf("   Build Type: Release\n")
			}
			return nil
		},
	}
}
