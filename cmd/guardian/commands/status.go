package commands

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/sr-tamim/guardian/internal/daemon"
	"github.com/sr-tamim/guardian/pkg/version"
)

// NewStatusCmd creates the status command
func NewStatusCmd(devMode *bool) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show Guardian status and statistics",
		Long:  "Display current Guardian status, active blocks, and monitoring statistics.",
		RunE: func(cmd *cobra.Command, args []string) error {
			versionInfo := version.Get()
			fmt.Printf("🛡️  Guardian Status v%s\n", versionInfo.Version)
			fmt.Println("════════════════════════════════")
			fmt.Printf("📅 Build Time: %s\n", version.GetBuildTime())
			fmt.Printf("🔧 Git Commit: %s\n", version.GetShortCommit())
			fmt.Printf("🐹 Go Version: %s\n", versionInfo.GoVersion)
			fmt.Printf("🖥️  Platform: %s/%s\n", versionInfo.Platform, versionInfo.Arch)
			fmt.Printf("⚙️  Development Mode: %v\n", *devMode)

			// Check daemon status
			pidManager := daemon.NewPIDManager()
			if pid, running := pidManager.GetRunningPID(); running {
				fmt.Printf("📊 Status: ✅ Running (PID: %d)\n", pid)
				fmt.Println("� Monitoring: ✅ Active")
			} else {
				fmt.Println("📊 Status: ⏹️  Stopped")
				fmt.Println("👀 Monitoring: ❌ Not active")
			}

			fmt.Println("🚫 Active Blocks: 0")
			return nil
		},
	}
}
