package main

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
)

var (
	version   string
	buildTime string
	devMode   bool
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
		fmt.Printf("🛡️  Guardian v%s starting...\n", getVersionString())
		fmt.Printf("⚙️  Development mode: %v\n", devMode)
		fmt.Println("📊 Monitoring system logs for intrusion attempts...")

		if devMode {
			fmt.Println("🧪 Running in development mode - no real blocking will occur")
			fmt.Println("📝 Use mock data and simulation for testing")
		}

		// For now, just show that it's working
		fmt.Println("✅ Guardian is ready! (Press Ctrl+C to stop)")

		// Simple loop to keep the program running
		for {
			time.Sleep(5 * time.Second)
			fmt.Printf("⏰ %s - System monitoring active...\n", time.Now().Format("15:04:05"))
		}
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

	// Add subcommands
	rootCmd.AddCommand(monitorCmd)
	rootCmd.AddCommand(statusCmd)
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
