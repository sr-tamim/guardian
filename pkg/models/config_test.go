package models

import (
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	// Test monitoring config
	if config.Monitoring.LookbackDuration != 10*time.Minute {
		t.Errorf("expected lookback duration 10m, got %v", config.Monitoring.LookbackDuration)
	}

	if config.Monitoring.CheckInterval != 5*time.Second {
		t.Errorf("expected check interval 5s, got %v", config.Monitoring.CheckInterval)
	}

	if !config.Monitoring.EnableRealTime {
		t.Error("expected real-time monitoring to be enabled")
	}

	if config.Monitoring.LogBufferSize != 100 {
		t.Errorf("expected log buffer size 100, got %d", config.Monitoring.LogBufferSize)
	}

	// Test blocking config
	if config.Blocking.FailureThreshold != 3 {
		t.Errorf("expected failure threshold 3, got %d", config.Blocking.FailureThreshold)
	}

	if config.Blocking.BlockDuration != 2*time.Minute {
		t.Errorf("expected block duration 2m, got %v", config.Blocking.BlockDuration)
	}

	if config.Blocking.MaxConcurrentBlocks != 100 {
		t.Errorf("expected max concurrent blocks 100, got %d", config.Blocking.MaxConcurrentBlocks)
	}

	if !config.Blocking.AutoUnblock {
		t.Error("expected auto-unblock to be enabled")
	}

	// Test whitelisted IPs
	expectedWhitelist := []string{
		"127.0.0.1",
		"::1",
		"192.168.0.0/16",
		"10.0.0.0/8",
	}

	if len(config.Blocking.WhitelistedIPs) != len(expectedWhitelist) {
		t.Errorf("expected %d whitelisted IPs, got %d", 
			len(expectedWhitelist), len(config.Blocking.WhitelistedIPs))
	}

	for i, expected := range expectedWhitelist {
		if i >= len(config.Blocking.WhitelistedIPs) || config.Blocking.WhitelistedIPs[i] != expected {
			t.Errorf("expected whitelisted IP %s at index %d, got %s", 
				expected, i, config.Blocking.WhitelistedIPs[i])
		}
	}

	// Test logging config
	if config.Logging.Level != "debug" {
		t.Errorf("expected log level 'debug', got %s", config.Logging.Level)
	}

	if config.Logging.Format != "text" {
		t.Errorf("expected log format 'text', got %s", config.Logging.Format)
	}

	if config.Logging.Output != "stdout" {
		t.Errorf("expected log output 'stdout', got %s", config.Logging.Output)
	}

	if config.Logging.EnableFile {
		t.Error("expected file logging to be disabled in default config")
	}

	// Test storage config
	if config.Storage.Type != "memory" {
		t.Errorf("expected storage type 'memory', got %s", config.Storage.Type)
	}

	if config.Storage.FilePath != "/tmp/guardian-dev.db" {
		t.Errorf("expected storage file path '/tmp/guardian-dev.db', got %s", config.Storage.FilePath)
	}

	// Test services
	if len(config.Services) != 1 {
		t.Errorf("expected 1 service, got %d", len(config.Services))
	}

	sshService := config.Services[0]
	if sshService.Name != "SSH" {
		t.Errorf("expected service name 'SSH', got %s", sshService.Name)
	}

	if sshService.LogPath != "/tmp/guardian_test_auth.log" {
		t.Errorf("expected log path '/tmp/guardian_test_auth.log', got %s", sshService.LogPath)
	}

	if sshService.LogPattern != "sshd" {
		t.Errorf("expected log pattern 'sshd', got %s", sshService.LogPattern)
	}

	if !sshService.Enabled {
		t.Error("expected SSH service to be enabled")
	}
}

func TestProductionConfig(t *testing.T) {
	config := ProductionConfig()

	// Test that production config has different, more restrictive settings
	if config.Monitoring.LookbackDuration != 1*time.Hour {
		t.Errorf("expected production lookback duration 1h, got %v", config.Monitoring.LookbackDuration)
	}

	if config.Blocking.FailureThreshold != 5 {
		t.Errorf("expected production failure threshold 5, got %d", config.Blocking.FailureThreshold)
	}

	if config.Blocking.BlockDuration != 20*time.Hour {
		t.Errorf("expected production block duration 20h, got %v", config.Blocking.BlockDuration)
	}

	if config.Logging.Level != "info" {
		t.Errorf("expected production log level 'info', got %s", config.Logging.Level)
	}

	if config.Storage.Type != "sqlite" {
		t.Errorf("expected production storage type 'sqlite', got %s", config.Storage.Type)
	}

	// Test production services include more than just SSH
	if len(config.Services) < 3 {
		t.Errorf("expected at least 3 services in production config, got %d", len(config.Services))
	}

	// Find SSH service
	var sshService *ServiceConfig
	for i := range config.Services {
		if config.Services[i].Name == "SSH" {
			sshService = &config.Services[i]
			break
		}
	}

	if sshService == nil {
		t.Fatal("SSH service not found in production config")
	}

	if sshService.LogPath != "/var/log/auth.log" {
		t.Errorf("expected production SSH log path '/var/log/auth.log', got %s", sshService.LogPath)
	}

	// Find Apache service (should be disabled by default)
	var apacheService *ServiceConfig
	for i := range config.Services {
		if config.Services[i].Name == "Apache" {
			apacheService = &config.Services[i]
			break
		}
	}

	if apacheService == nil {
		t.Fatal("Apache service not found in production config")
	}

	if apacheService.Enabled {
		t.Error("expected Apache service to be disabled by default in production config")
	}

	if apacheService.CustomThreshold != 10 {
		t.Errorf("expected Apache custom threshold 10, got %d", apacheService.CustomThreshold)
	}
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name   string
		config *Config
		valid  bool
	}{
		{
			name:   "default config should be valid",
			config: DefaultConfig(),
			valid:  true,
		},
		{
			name:   "production config should be valid",
			config: ProductionConfig(),
			valid:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Basic validation checks
			if tt.config.Monitoring.LookbackDuration <= 0 {
				t.Error("lookback duration should be positive")
			}

			if tt.config.Monitoring.CheckInterval <= 0 {
				t.Error("check interval should be positive")
			}

			if tt.config.Blocking.FailureThreshold <= 0 {
				t.Error("failure threshold should be positive")
			}

			if tt.config.Blocking.BlockDuration <= 0 {
				t.Error("block duration should be positive")
			}

			if tt.config.Blocking.MaxConcurrentBlocks <= 0 {
				t.Error("max concurrent blocks should be positive")
			}

			if tt.config.Storage.Type == "" {
				t.Error("storage type should not be empty")
			}

			if len(tt.config.Services) == 0 {
				t.Error("at least one service should be configured")
			}

			for i, service := range tt.config.Services {
				if service.Name == "" {
					t.Errorf("service %d name should not be empty", i)
				}

				if service.LogPath == "" {
					t.Errorf("service %d log path should not be empty", i)
				}

				if service.LogPattern == "" {
					t.Errorf("service %d log pattern should not be empty", i)
				}
			}
		})
	}
}

// Benchmark the config creation
func BenchmarkDefaultConfig(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = DefaultConfig()
	}
}

func BenchmarkProductionConfig(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = ProductionConfig()
	}
}