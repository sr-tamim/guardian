package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/sr-tamim/guardian/internal/platform"
	"github.com/sr-tamim/guardian/pkg/models"
)

var (
	version    string
	buildTime  string
	devMode    bool
	configFile string
	config     *models.Config
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "guardian",
	Short: "Guardian - Modern Cross-Platform Intrusion Prevention System",
	Long: `Guardian is a modern, cross-platform intrusion prevention system that monitors 
log files and automatically blocks malicious IP addresses. Built as a contemporary 
alternative to fail2ban with a beautiful terminal interface and intelligent threat detection.`,
	Version: getVersionString(),
}

var monitorCmd = &cobra.Command{
	Use:   "monitor",
	Short: "Start monitoring logs for intrusion attempts",
	Long:  "Start the Guardian monitoring daemon to watch log files for intrusion attempts and automatically block malicious IPs.",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Load configuration
		if err := loadConfig(); err != nil {
			return fmt.Errorf("failed to load configuration: %w", err)
		}

		fmt.Printf("🛡️  Guardian v%s starting...\n", getVersionString())
		fmt.Printf("⚙️  Development mode: %v\n", devMode)
		fmt.Printf("📊 Monitoring %d services...\n", len(config.Services))

		// Show enabled services
		for _, service := range config.Services {
			if service.Enabled {
				fmt.Printf("   📝 %s: %s\n", service.Name, service.LogPath)
			}
		}

		fmt.Printf("🚫 Failure threshold: %d attempts\n", config.Blocking.FailureThreshold)
		fmt.Printf("⏰ Block duration: %v\n", config.Blocking.BlockDuration)

		if devMode {
			fmt.Println("🧪 Running in development mode - no real blocking will occur")
			fmt.Println("📝 Using mock data and simulation for testing")
		}

		// Create platform provider
		factory := platform.NewFactory()
		provider, err := factory.CreateProvider(devMode, config)
		if err != nil {
			return fmt.Errorf("failed to create platform provider: %w", err)
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

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show Guardian status and statistics",
	Long:  "Display current Guardian status, active blocks, and monitoring statistics.",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("🛡️  Guardian Status v%s\n", getVersionString())
		fmt.Println("════════════════════════════════")
		fmt.Printf("📅 Build Time: %s\n", getBuildTime())
		fmt.Printf("⚙️  Development Mode: %v\n", devMode)
		fmt.Println("📊 Status: Not implemented yet")
		fmt.Println("🚫 Active Blocks: 0")
		fmt.Println("👀 Monitoring: Not active")
		return nil
	},
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().BoolVar(&devMode, "dev", false, "Enable development mode")
	rootCmd.PersistentFlags().StringVar(&configFile, "config", "", "Configuration file path")

	// Add subcommands
	rootCmd.AddCommand(monitorCmd)
	rootCmd.AddCommand(statusCmd)
}

// loadConfig loads configuration from file or uses defaults
func loadConfig() error {
	// If config file is specified, use it
	if configFile != "" {
		viper.SetConfigFile(configFile)
		if err := viper.ReadInConfig(); err != nil {
			return fmt.Errorf("error reading specified config file: %w", err)
		}
	} else if !devMode {
		// In production mode, look for config in standard locations
		viper.SetConfigName("guardian")
		viper.SetConfigType("yaml")
		viper.AddConfigPath("./configs")
		viper.AddConfigPath("$HOME/.config/guardian")
		viper.AddConfigPath("/etc/guardian")

		if err := viper.ReadInConfig(); err != nil {
			return fmt.Errorf("error reading config file: %w", err)
		}
	} else {
		// In dev mode without explicit config, look for development config
		viper.SetConfigName("development")
		viper.SetConfigType("yaml")
		viper.AddConfigPath("./configs")
		viper.AddConfigPath("$HOME/.config/guardian")
		viper.AddConfigPath("/etc/guardian")

		// If development config found, use it; otherwise fall back to defaults
		if err := viper.ReadInConfig(); err != nil {
			// Fallback to platform-aware defaults
			config = models.DefaultConfig()
			return nil
		}
	}

	// Start with platform-aware defaults, then overlay with config file values
	if devMode {
		config = models.DefaultConfig()
	} else {
		config = models.ProductionConfig()
	}

	// Unmarshal config file values on top of defaults (this merges intelligently)
	if err := viper.Unmarshal(config); err != nil {
		return fmt.Errorf("error unmarshaling config: %w", err)
	}

	return nil
}

func getVersionString() string {
	if version == "" {
		return "dev"
	}
	return version
}

func getBuildTime() string {
	if buildTime == "" {
		return "unknown"
	}
	return buildTime
}
