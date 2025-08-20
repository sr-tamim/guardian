package core

import (
	"context"
	"time"

	"github.com/sr-tamim/guardian/pkg/models"
)

// PlatformProvider defines the interface that each platform (Linux, Windows, macOS) must implement
type PlatformProvider interface {
	// Platform identification
	Name() string
	IsSupported() bool
	RequirementsCheck() error

	// Firewall operations
	BlockIP(ip string, duration time.Duration, reason string) error
	UnblockIP(ip string) error
	IsBlocked(ip string) (bool, error)
	ListBlockedIPs() ([]string, error)

	// Log monitoring
	GetLogPaths(service string) ([]string, error)
	StartLogMonitoring(ctx context.Context, logPath string, events chan<- LogEvent) error

	// Service management
	InstallService() error
	UninstallService() error
	StartService() error
	StopService() error
	ServiceStatus() (ServiceStatus, error)
}

// LogMonitor handles real-time log file monitoring
type LogMonitor interface {
	Start(ctx context.Context) error
	Stop() error
	AddLogFile(path string, parser LogParser) error
	RemoveLogFile(path string) error
	Events() <-chan LogEvent
}

// LogParser parses log entries to detect attack attempts
type LogParser interface {
	ParseLine(line string) (*models.AttackAttempt, error)
	ServiceName() string
	Patterns() []string
}

// ThreatDetector analyzes attack patterns and makes blocking decisions
type ThreatDetector interface {
	AnalyzeAttack(attempt *models.AttackAttempt) ThreatAssessment
	ShouldBlock(ip string, attempts []*models.AttackAttempt) bool
	IsWhitelisted(ip string) bool
}

// FirewallManager handles IP blocking across platforms
type FirewallManager interface {
	Block(ip string, duration time.Duration, reason string) error
	Unblock(ip string) error
	IsBlocked(ip string) (bool, error)
	ListBlocked() ([]*models.BlockRecord, error)
	Cleanup() error // Remove expired blocks
}

// Storage handles persistent data storage
type Storage interface {
	// Attack attempts
	SaveAttack(attempt *models.AttackAttempt) error
	GetAttacks(limit int, offset int) ([]*models.AttackAttempt, error)
	GetAttacksByIP(ip string, since time.Time) ([]*models.AttackAttempt, error)

	// Block records
	SaveBlock(block *models.BlockRecord) error
	GetBlock(ip string) (*models.BlockRecord, error)
	GetActiveBlocks() ([]*models.BlockRecord, error)
	UpdateBlock(block *models.BlockRecord) error

	// Statistics
	GetStatistics() (*models.Statistics, error)

	// Cleanup
	Close() error
}

// LogEvent represents a log entry event
type LogEvent struct {
	Timestamp time.Time
	Source    string // log file path
	Line      string
	Service   string
}

// ThreatAssessment represents the result of threat analysis
type ThreatAssessment struct {
	Severity          models.Severity
	Confidence        float64 // 0.0 to 1.0
	ShouldBlock       bool
	Reason            string
	RecommendedAction string
}

// ServiceStatus represents the status of Guardian service
type ServiceStatus struct {
	Running   bool
	PID       int
	StartTime time.Time
	Error     error
}

// Application represents the main Guardian application
type Application interface {
	Start(ctx context.Context) error
	Stop() error
	Reload() error
	Status() (*GuardianStatus, error)
}

// GuardianStatus represents the current status of Guardian
type GuardianStatus struct {
	Running           bool
	StartTime         time.Time
	Platform          string
	MonitoredServices []string
	ActiveBlocks      int
	TotalAttacks      int64
	Version           string
	ConfigPath        string
}
