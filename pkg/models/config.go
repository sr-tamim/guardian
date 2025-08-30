package models

import (
	"runtime"
	"strings"
	"time"

	"github.com/sr-tamim/guardian/pkg/utils"
)

// Config represents the complete Guardian configuration
type Config struct {
	Monitoring MonitoringConfig `yaml:"monitoring" json:"monitoring"`
	Blocking   BlockingConfig   `yaml:"blocking" json:"blocking"`
	Logging    LoggingConfig    `yaml:"logging" json:"logging"`
	Storage    StorageConfig    `yaml:"storage" json:"storage"`
	Services   []ServiceConfig  `yaml:"services" json:"services"`
}

// MonitoringConfig holds monitoring-related settings
type MonitoringConfig struct {
	LookbackDuration time.Duration `yaml:"lookback_duration" json:"lookback_duration"`
	CheckInterval    time.Duration `yaml:"check_interval" json:"check_interval"`
	EnableRealTime   bool          `yaml:"enable_real_time" json:"enable_real_time"`
	LogBufferSize    int           `yaml:"log_buffer_size" json:"log_buffer_size"`
}

// BlockingConfig holds IP blocking settings
type BlockingConfig struct {
	FailureThreshold    int           `yaml:"failure_threshold" json:"failure_threshold"`
	BlockDuration       time.Duration `yaml:"block_duration" json:"block_duration"`
	MaxConcurrentBlocks int           `yaml:"max_concurrent_blocks" json:"max_concurrent_blocks"`
	WhitelistedIPs      []string      `yaml:"whitelisted_ips" json:"whitelisted_ips"`
	AutoUnblock         bool          `yaml:"auto_unblock" json:"auto_unblock"`
	CleanupInterval     time.Duration `yaml:"cleanup_interval" json:"cleanup_interval"`
	RuleNameTemplate    string        `yaml:"rule_name_template" json:"rule_name_template"`
}

// GenerateRuleName creates a firewall rule name from the template
// Supports placeholders: {app}, {timestamp}, {ip}, {service}
// Default template: "Guardian - {timestamp} - {ip}"
func (b *BlockingConfig) GenerateRuleName(ip, service string) string {
	template := b.RuleNameTemplate
	if template == "" {
		template = "Guardian - {timestamp} - {ip}" // Default fallback
	}

	timestamp := time.Now().Format("20060102150405") // yyyyMMddHHmmss format

	// Replace placeholders
	template = strings.ReplaceAll(template, "{app}", "Guardian")
	template = strings.ReplaceAll(template, "{timestamp}", timestamp)
	template = strings.ReplaceAll(template, "{ip}", ip)
	template = strings.ReplaceAll(template, "{service}", service)

	return template
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level               string `yaml:"level" json:"level"`
	Format              string `yaml:"format" json:"format"`
	Output              string `yaml:"output" json:"output"`
	EnableFile          bool   `yaml:"enable_file" json:"enable_file"`
	FilePath            string `yaml:"file_path" json:"file_path"`
	EnableContextual    bool   `yaml:"enable_contextual" json:"enable_contextual"`
	LogEventLookups     bool   `yaml:"log_event_lookups" json:"log_event_lookups"`
	LogFirewallActions  bool   `yaml:"log_firewall_actions" json:"log_firewall_actions"`
	LogAttackAttempts   bool   `yaml:"log_attack_attempts" json:"log_attack_attempts"`
	LogMonitoringEvents bool   `yaml:"log_monitoring_events" json:"log_monitoring_events"`
	LogCleanupEvents    bool   `yaml:"log_cleanup_events" json:"log_cleanup_events"`
}

// StorageConfig holds database configuration
type StorageConfig struct {
	Type     string `yaml:"type" json:"type"`
	FilePath string `yaml:"file_path" json:"file_path"`
}

// DefaultConfig returns a default configuration suitable for development
func DefaultConfig() *Config {
	paths := utils.NewPlatformPaths()

	// Development-friendly paths
	var logPath, dbPath, serviceLogPath string

	if runtime.GOOS == "windows" {
		// Windows development paths
		logPath = paths.GetDefaultGuardianLogPath()
		dbPath = paths.GetDefaultGuardianDatabasePath()
		serviceLogPath = paths.GetDefaultServiceLogPaths("SSH")[0] // First available path
	} else {
		// Unix-like development paths (use temp for safety)
		logPath = "/tmp/guardian-dev.log"
		dbPath = "/tmp/guardian-dev.db"
		serviceLogPath = "/tmp/guardian_test_auth.log"
	}

	return &Config{
		Monitoring: MonitoringConfig{
			LookbackDuration: 10 * time.Minute,
			CheckInterval:    5 * time.Second,
			EnableRealTime:   true,
			LogBufferSize:    100,
		},
		Blocking: BlockingConfig{
			FailureThreshold:    3,
			BlockDuration:       2 * time.Minute,
			MaxConcurrentBlocks: 100,
			WhitelistedIPs: []string{
				"127.0.0.1",
				"::1",
				"192.168.0.0/16",
				"10.0.0.0/8",
			},
			AutoUnblock:      true,
			CleanupInterval:  30 * time.Second,                // Fast cleanup for development
			RuleNameTemplate: "Guardian - {timestamp} - {ip}", // Default template
		},
		Logging: LoggingConfig{
			Level:               "debug",
			Format:              "text",
			Output:              "stdout",
			EnableFile:          false,
			FilePath:            logPath, // Platform-aware path
			EnableContextual:    true,
			LogEventLookups:     true,
			LogFirewallActions:  true,
			LogAttackAttempts:   true,
			LogMonitoringEvents: true,
			LogCleanupEvents:    true,
		},
		Storage: StorageConfig{
			Type:     "memory",
			FilePath: dbPath, // Platform-aware path
		},
		Services: []ServiceConfig{
			{
				Name:            "SSH",
				LogPath:         serviceLogPath, // Platform-aware path
				LogPattern:      "sshd",
				CustomThreshold: 0,
				Enabled:         true,
			},
		},
	}
}

// ProductionConfig returns a production-ready configuration
func ProductionConfig() *Config {
	paths := utils.NewPlatformPaths()

	return &Config{
		Monitoring: MonitoringConfig{
			LookbackDuration: 1 * time.Hour,
			CheckInterval:    30 * time.Second,
			EnableRealTime:   true,
			LogBufferSize:    1000,
		},
		Blocking: BlockingConfig{
			FailureThreshold:    5,
			BlockDuration:       20 * time.Hour,
			MaxConcurrentBlocks: 1000,
			WhitelistedIPs: []string{
				"127.0.0.1",
				"::1",
				"192.168.1.0/24",
			},
			AutoUnblock:      true,
			CleanupInterval:  5 * time.Minute,                 // Production cleanup every 5 minutes
			RuleNameTemplate: "Guardian - {timestamp} - {ip}", // Default template
		},
		Logging: LoggingConfig{
			Level:               "info",
			Format:              "text",
			Output:              "stdout",
			EnableFile:          true,
			FilePath:            paths.GetDefaultGuardianLogPath(), // Platform-aware path
			EnableContextual:    true,
			LogEventLookups:     false, // Too verbose for production
			LogFirewallActions:  true,
			LogAttackAttempts:   true,
			LogMonitoringEvents: false, // Too verbose for production
			LogCleanupEvents:    true,
		},
		Storage: StorageConfig{
			Type:     "sqlite",
			FilePath: paths.GetDefaultGuardianDatabasePath(), // Platform-aware path
		},
		Services: createPlatformServices(paths),
	}
}

// createPlatformServices creates platform-specific service configurations
func createPlatformServices(paths *utils.PlatformPaths) []ServiceConfig {
	var services []ServiceConfig

	// SSH service (available on all platforms but different log locations)
	sshPaths := paths.GetDefaultServiceLogPaths("SSH")
	if len(sshPaths) > 0 {
		services = append(services, ServiceConfig{
			Name:            "SSH",
			LogPath:         sshPaths[0], // Use first available path
			LogPattern:      "sshd",
			CustomThreshold: 0,
			Enabled:         true,
		})
	}

	// Platform-specific services
	switch runtime.GOOS {
	case "windows":
		// Windows RDP service
		services = append(services, ServiceConfig{
			Name:            "RDP",
			LogPath:         "Security", // Windows Event Log
			LogPattern:      "4625",     // Failed logon event ID
			CustomThreshold: 0,
			Enabled:         true,
		})

		// Windows IIS (if available)
		iisPaths := paths.GetDefaultServiceLogPaths("IIS")
		if len(iisPaths) > 0 {
			services = append(services, ServiceConfig{
				Name:            "IIS",
				LogPath:         iisPaths[0],
				LogPattern:      "iis",
				CustomThreshold: 10,
				Enabled:         false,
			})
		}

	case "linux", "darwin":
		// Apache service
		apachePaths := paths.GetDefaultServiceLogPaths("Apache")
		if len(apachePaths) > 0 {
			services = append(services, ServiceConfig{
				Name:            "Apache",
				LogPath:         apachePaths[0],
				LogPattern:      "apache",
				CustomThreshold: 10,
				Enabled:         false,
			})
		}

		// Nginx service
		nginxPaths := paths.GetDefaultServiceLogPaths("Nginx")
		if len(nginxPaths) > 0 {
			services = append(services, ServiceConfig{
				Name:            "Nginx",
				LogPath:         nginxPaths[0],
				LogPattern:      "nginx",
				CustomThreshold: 10,
				Enabled:         false,
			})
		}
	}

	return services
}
