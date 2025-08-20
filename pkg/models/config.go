package models

import (
	"time"
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
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level      string `yaml:"level" json:"level"`
	Format     string `yaml:"format" json:"format"`
	Output     string `yaml:"output" json:"output"`
	EnableFile bool   `yaml:"enable_file" json:"enable_file"`
	FilePath   string `yaml:"file_path" json:"file_path"`
}

// StorageConfig holds database configuration
type StorageConfig struct {
	Type     string `yaml:"type" json:"type"`
	FilePath string `yaml:"file_path" json:"file_path"`
}

// DefaultConfig returns a default configuration suitable for development
func DefaultConfig() *Config {
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
			AutoUnblock: true,
		},
		Logging: LoggingConfig{
			Level:      "debug",
			Format:     "text",
			Output:     "stdout",
			EnableFile: false,
			FilePath:   "/tmp/guardian-dev.log",
		},
		Storage: StorageConfig{
			Type:     "memory",
			FilePath: "/tmp/guardian-dev.db",
		},
		Services: []ServiceConfig{
			{
				Name:            "SSH",
				LogPath:         "/tmp/guardian_test_auth.log",
				LogPattern:      "sshd",
				CustomThreshold: 0,
				Enabled:         true,
			},
		},
	}
}

// ProductionConfig returns a production-ready configuration
func ProductionConfig() *Config {
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
			AutoUnblock: true,
		},
		Logging: LoggingConfig{
			Level:      "info",
			Format:     "text",
			Output:     "stdout",
			EnableFile: false,
			FilePath:   "/var/log/guardian/guardian.log",
		},
		Storage: StorageConfig{
			Type:     "sqlite",
			FilePath: "/var/lib/guardian/guardian.db",
		},
		Services: []ServiceConfig{
			{
				Name:            "SSH",
				LogPath:         "/var/log/auth.log",
				LogPattern:      "sshd",
				CustomThreshold: 0,
				Enabled:         true,
			},
			{
				Name:            "Apache",
				LogPath:         "/var/log/apache2/access.log",
				LogPattern:      "apache",
				CustomThreshold: 10,
				Enabled:         false,
			},
			{
				Name:            "Nginx",
				LogPath:         "/var/log/nginx/access.log",
				LogPattern:      "nginx",
				CustomThreshold: 10,
				Enabled:         false,
			},
		},
	}
}
