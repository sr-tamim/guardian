package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/sr-tamim/guardian/cmd/guardian/commands"
	"github.com/sr-tamim/guardian/internal/tui"
	"github.com/sr-tamim/guardian/pkg/logger"
	"github.com/sr-tamim/guardian/pkg/models"
	"github.com/sr-tamim/guardian/pkg/version"
)

var (
	devMode    bool
	configFile string
	config     *models.Config
)

func main() {
	// Handle no-args scenario (double-click)
	if len(os.Args) == 1 {
		// Use fmt.Println for TUI launch message since logger isn't initialized yet
		fmt.Println("🛡️  Guardian - No command specified, launching TUI...")
		if err := launchTUI(); err != nil {
			fmt.Printf("❌ Failed to launch TUI: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Execute CLI commands
	if err := newRootCommand().Execute(); err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		os.Exit(1)
	}
}

// newRootCommand creates the root command with all subcommands
func newRootCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:     "guardian",
		Short:   "Guardian - Modern Cross-Platform Intrusion Prevention System",
		Long:    `Guardian is a modern, cross-platform intrusion prevention system that monitors log files and automatically blocks malicious IP addresses. Built as a contemporary alternative to fail2ban with a beautiful terminal interface and intelligent threat detection.`,
		Version: version.GetVersion(),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Default to TUI when no subcommand
			return launchTUI()
		},
	}

	// Global flags
	rootCmd.PersistentFlags().BoolVar(&devMode, "dev", false, "Enable development mode")
	rootCmd.PersistentFlags().StringVar(&configFile, "config", "", "Configuration file path")

	// Add subcommands
	rootCmd.AddCommand(commands.NewMonitorCmd(getConfig, &devMode))
	rootCmd.AddCommand(commands.NewStopCmd(getConfig, &devMode))
	rootCmd.AddCommand(commands.NewStatusCmd(&devMode))
	rootCmd.AddCommand(commands.NewVersionCmd())
	rootCmd.AddCommand(commands.NewTUICmd(getConfig, &devMode))
	rootCmd.AddCommand(commands.NewAutostartCmd(getConfig, &devMode))

	return rootCmd
}

// getConfig loads and returns configuration (lazy loading with caching)
func getConfig() (*models.Config, error) {
	if config != nil {
		return config, nil
	}

	if err := loadConfig(); err != nil {
		return nil, err
	}
	return config, nil
}

// launchTUI starts the terminal user interface
func launchTUI() error {
	if err := loadConfig(); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	program := tea.NewProgram(
		tui.NewDashboard(),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	_, err := program.Run()
	return err
}

// loadConfig loads configuration from file or sets defaults
func loadConfig() error {
	if config != nil {
		return nil // Already loaded
	}

	// Set sensible defaults
	viper.SetDefault("blocking.failure_threshold", 3)
	viper.SetDefault("blocking.block_duration", "2m")
	viper.SetDefault("blocking.cleanup_interval", "1m")

	// Configure config file paths
	if configFile != "" {
		viper.SetConfigFile(configFile)
	} else {
		viper.SetConfigName("guardian")
		viper.SetConfigType("yaml")
		viper.AddConfigPath("./configs/")
		viper.AddConfigPath(".")

		// Platform-specific paths
		if userConfigDir, err := os.UserConfigDir(); err == nil {
			viper.AddConfigPath(userConfigDir + "/Guardian")
		}
	}

	// Load config file (non-fatal if missing)
	if err := viper.ReadInConfig(); err != nil {
		// Use fmt.Printf before logger is initialized
		fmt.Printf("⚠️  Could not read config file: %v\n", err)
		fmt.Println("📝 Using default configuration...")
	}

	// Unmarshal into struct
	config = &models.Config{}
	if err := viper.Unmarshal(config); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Initialize logger with configuration
	if err := logger.InitializeLogger(&config.Logging); err != nil {
		// Fallback to fmt.Printf if logger initialization fails
		fmt.Printf("⚠️  Failed to initialize logger: %v\n", err)
		// Continue with default logging via fmt.Printf
	} else {
		logger.Info("Guardian configuration loaded successfully",
			"config_file", viper.ConfigFileUsed(),
			"log_level", config.Logging.Level,
			"log_format", config.Logging.Format,
		)
	}

	// Ensure default services exist
	if len(config.Services) == 0 {
		config.Services = []models.ServiceConfig{
			{
				Name:    "SSH",
				LogPath: "C:\\ProgramData\\ssh\\logs\\sshd.log",
				Enabled: true,
			},
		}
	}

	return nil
}
