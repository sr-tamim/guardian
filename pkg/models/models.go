package models

import (
	"net"
	"time"
)

// AttackAttempt represents a detected intrusion attempt
type AttackAttempt struct {
	ID        int64     `json:"id" db:"id"`
	Timestamp time.Time `json:"timestamp" db:"timestamp"`
	IP        string    `json:"ip" db:"ip"`
	Service   string    `json:"service" db:"service"`
	Username  string    `json:"username" db:"username"`
	Message   string    `json:"message" db:"message"`
	Severity  Severity  `json:"severity" db:"severity"`
	Source    string    `json:"source" db:"source"` // log file path
	Blocked   bool      `json:"blocked" db:"blocked"`
}

// BlockRecord represents an IP that has been blocked
type BlockRecord struct {
	ID          int64      `json:"id" db:"id"`
	IP          string     `json:"ip" db:"ip"`
	BlockedAt   time.Time  `json:"blocked_at" db:"blocked_at"`
	ExpiresAt   *time.Time `json:"expires_at" db:"expires_at"`
	Reason      string     `json:"reason" db:"reason"`
	Service     string     `json:"service" db:"service"`
	AttackCount int        `json:"attack_count" db:"attack_count"`
	IsActive    bool       `json:"is_active" db:"is_active"`
	UnblockedAt *time.Time `json:"unblocked_at" db:"unblocked_at"`
}

// ServiceConfig represents configuration for a monitored service
type ServiceConfig struct {
	Name            string `yaml:"name" json:"name"`
	LogPath         string `yaml:"log_path" json:"log_path"`
	LogPattern      string `yaml:"log_pattern" json:"log_pattern"`
	CustomThreshold int    `yaml:"custom_threshold" json:"custom_threshold"`
	Enabled         bool   `yaml:"enabled" json:"enabled"`
}

// Statistics holds monitoring and blocking statistics
type Statistics struct {
	TotalAttacks      int64     `json:"total_attacks"`
	BlockedIPs        int64     `json:"blocked_ips"`
	ActiveBlocks      int64     `json:"active_blocks"`
	ServicesMonitored int       `json:"services_monitored"`
	UptimeSeconds     int64     `json:"uptime_seconds"`
	LastActivity      time.Time `json:"last_activity"`
}

// Severity levels for attack attempts
type Severity int

const (
	SeverityLow Severity = iota
	SeverityMedium
	SeverityHigh
	SeverityCritical
)

func (s Severity) String() string {
	switch s {
	case SeverityLow:
		return "low"
	case SeverityMedium:
		return "medium"
	case SeverityHigh:
		return "high"
	case SeverityCritical:
		return "critical"
	default:
		return "unknown"
	}
}

// IsValidIP checks if the given string is a valid IP address
func (a *AttackAttempt) IsValidIP() bool {
	return net.ParseIP(a.IP) != nil
}

// IsExpired checks if a block record has expired
func (b *BlockRecord) IsExpired() bool {
	if b.ExpiresAt == nil {
		return false // permanent block
	}
	return time.Now().After(*b.ExpiresAt)
}

// TimeUntilExpiry returns the duration until the block expires
func (b *BlockRecord) TimeUntilExpiry() time.Duration {
	if b.ExpiresAt == nil {
		return 0 // permanent block
	}
	return time.Until(*b.ExpiresAt)
}
