package models

import (
	"net"
	"testing"
	"time"
)

func TestAttackAttempt(t *testing.T) {
	now := time.Now()
	
	tests := []struct {
		name    string
		attempt AttackAttempt
		isValid bool
	}{
		{
			name: "valid IPv4",
			attempt: AttackAttempt{
				ID:        1,
				Timestamp: now,
				IP:        "192.168.1.100",
				Service:   "ssh",
				Username:  "admin",
				Message:   "Failed password for admin",
				Severity:  SeverityHigh,
				Source:    "/var/log/auth.log",
				Blocked:   false,
			},
			isValid: true,
		},
		{
			name: "valid IPv6",
			attempt: AttackAttempt{
				ID:        2,
				Timestamp: now,
				IP:        "2001:db8::1",
				Service:   "ssh",
				Username:  "root",
				Message:   "Failed password for root",
				Severity:  SeverityMedium,
				Source:    "/var/log/auth.log",
				Blocked:   true,
			},
			isValid: true,
		},
		{
			name: "invalid IP",
			attempt: AttackAttempt{
				ID:        3,
				Timestamp: now,
				IP:        "invalid-ip",
				Service:   "ssh",
				Username:  "user",
				Message:   "Authentication failure",
				Severity:  SeverityLow,
				Source:    "/var/log/auth.log",
				Blocked:   false,
			},
			isValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.attempt.IsValidIP() != tt.isValid {
				t.Errorf("IsValidIP() = %v, want %v for IP %s", 
					tt.attempt.IsValidIP(), tt.isValid, tt.attempt.IP)
			}
		})
	}
}

func TestBlockRecord(t *testing.T) {
	now := time.Now()
	future := now.Add(time.Hour)
	past := now.Add(-time.Hour)

	tests := []struct {
		name       string
		record     BlockRecord
		isExpired  bool
		hasExpiry  bool
	}{
		{
			name: "active block with future expiry",
			record: BlockRecord{
				ID:          1,
				IP:          "192.168.1.100",
				BlockedAt:   now,
				ExpiresAt:   &future,
				Reason:      "Multiple failed logins",
				Service:     "ssh",
				AttackCount: 5,
				IsActive:    true,
			},
			isExpired: false,
			hasExpiry: true,
		},
		{
			name: "expired block",
			record: BlockRecord{
				ID:          2,
				IP:          "10.0.0.50",
				BlockedAt:   past,
				ExpiresAt:   &past,
				Reason:      "Brute force attack",
				Service:     "ftp",
				AttackCount: 10,
				IsActive:    false,
			},
			isExpired: true,
			hasExpiry: true,
		},
		{
			name: "permanent block",
			record: BlockRecord{
				ID:          3,
				IP:          "172.16.0.1",
				BlockedAt:   now,
				ExpiresAt:   nil,
				Reason:      "Malicious activity",
				Service:     "web",
				AttackCount: 50,
				IsActive:    true,
			},
			isExpired: false,
			hasExpiry: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.record.IsExpired() != tt.isExpired {
				t.Errorf("IsExpired() = %v, want %v", tt.record.IsExpired(), tt.isExpired)
			}

			if tt.hasExpiry {
				duration := tt.record.TimeUntilExpiry()
				if tt.isExpired && duration > 0 {
					t.Error("expired blocks should have non-positive time until expiry")
				}
				if !tt.isExpired && duration <= 0 {
					t.Error("active blocks should have positive time until expiry")
				}
			} else {
				// Permanent blocks should return 0 duration
				if tt.record.TimeUntilExpiry() != 0 {
					t.Error("permanent blocks should return 0 time until expiry")
				}
			}
		})
	}
}

func TestSeverity(t *testing.T) {
	tests := []struct {
		severity Severity
		expected string
	}{
		{SeverityLow, "low"},
		{SeverityMedium, "medium"},
		{SeverityHigh, "high"},
		{SeverityCritical, "critical"},
		{Severity(99), "unknown"}, // Invalid severity
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if tt.severity.String() != tt.expected {
				t.Errorf("Severity.String() = %v, want %v", tt.severity.String(), tt.expected)
			}
		})
	}
}

func TestServiceConfig(t *testing.T) {
	config := ServiceConfig{
		Name:            "SSH",
		LogPath:         "/var/log/auth.log",
		LogPattern:      "sshd",
		CustomThreshold: 5,
		Enabled:         true,
	}

	if config.Name == "" {
		t.Error("service name should not be empty")
	}

	if config.LogPath == "" {
		t.Error("log path should not be empty")
	}

	if config.CustomThreshold < 0 {
		t.Error("custom threshold should not be negative")
	}
}

func TestStatistics(t *testing.T) {
	stats := Statistics{
		TotalAttacks:      1000,
		BlockedIPs:        50,
		ActiveBlocks:      25,
		ServicesMonitored: 3,
		UptimeSeconds:     3600,
		LastActivity:      time.Now(),
	}

	if stats.TotalAttacks < 0 {
		t.Error("total attacks should not be negative")
	}

	if stats.BlockedIPs < 0 {
		t.Error("blocked IPs should not be negative")
	}

	if stats.ActiveBlocks < 0 {
		t.Error("active blocks should not be negative")
	}

	if stats.ActiveBlocks > stats.BlockedIPs {
		t.Error("active blocks should not exceed total blocked IPs")
	}

	if stats.ServicesMonitored < 0 {
		t.Error("monitored services should not be negative")
	}

	if stats.UptimeSeconds < 0 {
		t.Error("uptime should not be negative")
	}
}

func TestIPValidation(t *testing.T) {
	validIPs := []string{
		"127.0.0.1",
		"192.168.1.1",
		"10.0.0.1",
		"172.16.0.1",
		"8.8.8.8",
		"::1",
		"2001:db8::1",
	}

	invalidIPs := []string{
		"",
		"invalid",
		"256.256.256.256",
		"192.168.1",
		"192.168.1.1.1",
		":::",
		"gggg::1",
	}

	for _, ip := range validIPs {
		t.Run("valid_"+ip, func(t *testing.T) {
			attempt := AttackAttempt{IP: ip}
			if !attempt.IsValidIP() {
				t.Errorf("IP %s should be valid", ip)
			}
		})
	}

	for _, ip := range invalidIPs {
		t.Run("invalid_"+ip, func(t *testing.T) {
			attempt := AttackAttempt{IP: ip}
			if attempt.IsValidIP() {
				t.Errorf("IP %s should be invalid", ip)
			}
		})
	}
}

// Test that net.ParseIP works as expected in our validation
func TestNetParseIPBehavior(t *testing.T) {
	testCases := []struct {
		ip       string
		expected bool
	}{
		{"192.168.1.1", true},
		{"::1", true},
		{"", false},
		{"invalid", false},
		{"256.1.1.1", false},
	}

	for _, tc := range testCases {
		t.Run(tc.ip, func(t *testing.T) {
			result := net.ParseIP(tc.ip) != nil
			if result != tc.expected {
				t.Errorf("net.ParseIP(%q) != nil = %v, want %v", tc.ip, result, tc.expected)
			}
		})
	}
}