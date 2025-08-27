package commands

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/sr-tamim/guardian/internal/daemon"
	"github.com/sr-tamim/guardian/internal/platform"
	"github.com/sr-tamim/guardian/pkg/models"
	"github.com/sr-tamim/guardian/pkg/version"
)

// NewMonitorCmd creates the monitor command
func NewMonitorCmd(configLoader func() (*models.Config, error), devMode *bool) *cobra.Command {
	var daemonMode bool
	var daemonInternal bool

	cmd := &cobra.Command{
		Use:   "monitor",
		Short: "Start monitoring logs for intrusion attempts",
		Long:  "Start the Guardian monitoring daemon to watch log files for intrusion attempts and automatically block malicious IPs.",
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

			// Handle daemon-internal mode (when spawned by daemon)
			if daemonInternal {
				// This is the actual daemon process, run monitoring directly
				ctx := context.Background()
				return daemonManager.RunMonitorInCurrentProcess(ctx)
			}

			// Check if daemon mode is requested
			if daemonMode {
				return daemonManager.StartDaemon()
			}

			// Check if daemon is already running (non-daemon mode)
			if daemonManager.IsRunning() {
				fmt.Printf("🛡️  Guardian daemon is already running!\n")
				fmt.Println("════════════════════════════════════════")

				status := daemonManager.GetStatus()
				fmt.Printf("📊 Status: Running\n")
				fmt.Printf("🔢 PID: %d\n", status.PID)
				fmt.Printf("📄 PID File: %s\n", status.PIDFile)
				fmt.Printf("🖥️  Platform: %s\n", provider.Name())
				fmt.Printf("⚙️  Development Mode: %v\n", *devMode)

				fmt.Println("\n💡 Tips:")
				fmt.Println("   • Use 'guardian status' to check detailed status")
				fmt.Println("   • Use 'guardian stop' to stop the daemon")

				return nil
			}

			fmt.Printf("🛡️  Guardian v%s starting...\n", version.GetVersion())
			fmt.Printf("⚙️  Development mode: %v\n", *devMode)
			fmt.Printf("📊 Monitoring %d services...\n", len(config.Services))

			// Show enabled services
			for _, service := range config.Services {
				if service.Enabled {
					fmt.Printf("   📝 %s: %s\n", service.Name, service.LogPath)
				}
			}

			fmt.Printf("🚫 Failure threshold: %d attempts\n", config.Blocking.FailureThreshold)
			fmt.Printf("⏰ Block duration: %v\n", config.Blocking.BlockDuration)

			if *devMode {
				fmt.Println("🧪 Running in development mode - no real blocking will occur")
				fmt.Println("📝 Using mock data and simulation for testing")
			}

			fmt.Printf("🖥️  Platform: %s\n", provider.Name())
			fmt.Println("✅ Guardian is ready! (Press Ctrl+C to stop)")

			// Set up context for graceful shutdown
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			// Handle shutdown signals
			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

			// Start monitoring for enabled services
			for _, service := range config.Services {
				if service.Enabled {
					logPaths, err := provider.GetLogPaths(service.Name)
					if err != nil {
						fmt.Printf("❌ Failed to get log paths for %s: %v\n", service.Name, err)
						continue
					}

					for _, logPath := range logPaths {
						go func(path, serviceName string) {
							// Start log monitoring (this will spawn background goroutines)
							if err := provider.StartLogMonitoring(ctx, path, nil); err != nil {
								fmt.Printf("❌ Failed to start monitoring %s: %v\n", path, err)
							}
						}(logPath, service.Name)
					}
				}
			}

			// Wait for shutdown signal
			select {
			case <-sigChan:
				fmt.Println("\n🛑 Received shutdown signal...")
			case <-ctx.Done():
				fmt.Println("\n🛑 Context cancelled...")
			}

			cancel()
			fmt.Println("👋 Guardian stopped gracefully")
			return nil
		},
	}

	// Add flags
	cmd.Flags().BoolVarP(&daemonMode, "daemon", "d", false, "Run in daemon mode (background)")
	cmd.Flags().BoolVar(&daemonInternal, "daemon-internal", false, "Internal flag for daemon process (do not use directly)")
	cmd.Flags().MarkHidden("daemon-internal") // Hide from help

	return cmd
}
