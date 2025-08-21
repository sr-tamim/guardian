package mock

import (
	"context"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/sr-tamim/guardian/internal/core"
	"github.com/sr-tamim/guardian/pkg/models"
)

// MockProvider implements PlatformProvider for development and testing
// It simulates Windows Event Log behavior based on the PowerShell script approach
type MockProvider struct {
	mu            sync.RWMutex
	name          string
	config        *models.Config // Add config reference
	blockedIPs    map[string]*models.BlockRecord
	firewallRules map[string]*FirewallRule
	isRunning     bool
	startTime     time.Time

	// Statistics
	totalAttacks int64
	totalBlocks  int64

	// Channels for communication
	logEvents chan core.LogEvent
	stopChan  chan struct{}
}

// FirewallRule represents a mock Windows firewall rule
// Mimics the structure from PowerShell: "Guardian - $(timestamp) - $IPAddr"
type FirewallRule struct {
	Name      string     // "Guardian - 20250821073000 - 192.168.1.100"
	IP        string     // "192.168.1.100"
	CreatedAt time.Time  // When rule was created
	ExpiresAt *time.Time // When rule expires (like BlockDuration in PS script)
	Action    string     // "Block"
	Direction string     // "Inbound"
	IsActive  bool
}

// NewMockProvider creates a new mock platform provider
func NewMockProvider(config *models.Config) *MockProvider {
	return &MockProvider{
		name:          "MockProvider",
		config:        config,
		blockedIPs:    make(map[string]*models.BlockRecord),
		firewallRules: make(map[string]*FirewallRule),
		logEvents:     make(chan core.LogEvent, 100),
		stopChan:      make(chan struct{}),
		startTime:     time.Now(),
	}
}

// Name returns the provider name
func (m *MockProvider) Name() string {
	return m.name
}

// IsSupported always returns true for mock
func (m *MockProvider) IsSupported() bool {
	return true
}

// RequirementsCheck always passes for mock
func (m *MockProvider) RequirementsCheck() error {
	return nil
}

// BlockIP simulates blocking an IP using Windows Firewall
// Mimics: New-NetFirewallRule -DisplayName $RuleName -Direction Inbound -RemoteAddress $IPAddr -Action Block
func (m *MockProvider) BlockIP(ip string, duration time.Duration, reason string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Validate IP address
	if net.ParseIP(ip) == nil {
		return core.NewError(core.ErrInvalidIP, "invalid IP address", nil)
	}

	// Check if already blocked (like your PowerShell script)
	if existing, exists := m.blockedIPs[ip]; exists && existing.IsActive {
		return core.NewError(core.ErrIPAlreadyBlocked, fmt.Sprintf("IP %s is already blocked", ip), nil)
	}

	// Create firewall rule (using configurable naming convention)
	ruleName := m.config.Blocking.GenerateRuleName(ip, "RDP") // Service can be dynamic too

	var expiresAt *time.Time
	if duration > 0 {
		expiry := time.Now().Add(duration)
		expiresAt = &expiry
	}

	// Create firewall rule
	rule := &FirewallRule{
		Name:      ruleName,
		IP:        ip,
		CreatedAt: time.Now(),
		ExpiresAt: expiresAt,
		Action:    "Block",
		Direction: "Inbound",
		IsActive:  true,
	}
	m.firewallRules[ruleName] = rule

	// Create block record
	blockRecord := &models.BlockRecord{
		IP:          ip,
		BlockedAt:   time.Now(),
		ExpiresAt:   expiresAt,
		Reason:      reason,
		Service:     "RDP", // Since we're focusing on Windows/RDP first
		AttackCount: 1,
		IsActive:    true,
	}
	m.blockedIPs[ip] = blockRecord
	m.totalBlocks++

	fmt.Printf("🚫 [MOCK] Blocked IP %s with rule: %s (expires: %v)\n",
		ip, ruleName, formatExpiry(expiresAt))
	return nil
}

// UnblockIP simulates removing a firewall rule
// Mimics: Remove-NetFirewallRule -Name $Rule.Name
func (m *MockProvider) UnblockIP(ip string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	blockRecord, exists := m.blockedIPs[ip]
	if !exists || !blockRecord.IsActive {
		return core.NewError(core.ErrIPNotBlocked, fmt.Sprintf("IP %s is not blocked", ip), nil)
	}

	// Find and remove firewall rule
	var ruleToRemove string
	for ruleName, rule := range m.firewallRules {
		if rule.IP == ip && rule.IsActive {
			rule.IsActive = false
			ruleToRemove = ruleName
			break
		}
	}

	// Update block record
	blockRecord.IsActive = false
	now := time.Now()
	blockRecord.UnblockedAt = &now

	fmt.Printf("✅ [MOCK] Unblocked IP %s (removed rule: %s)\n", ip, ruleToRemove)
	return nil
}

// IsBlocked checks if an IP is currently blocked
func (m *MockProvider) IsBlocked(ip string) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	blockRecord, exists := m.blockedIPs[ip]
	if !exists {
		return false, nil
	}

	// Check if expired
	if blockRecord.ExpiresAt != nil && time.Now().After(*blockRecord.ExpiresAt) {
		return false, nil
	}

	return blockRecord.IsActive, nil
}

// ListBlockedIPs returns all currently blocked IPs
func (m *MockProvider) ListBlockedIPs() ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var blocked []string
	for ip, record := range m.blockedIPs {
		if record.IsActive {
			// Check if expired
			if record.ExpiresAt == nil || time.Now().Before(*record.ExpiresAt) {
				blocked = append(blocked, ip)
			}
		}
	}

	return blocked, nil
}

// GetLogPaths returns mock log paths
func (m *MockProvider) GetLogPaths(service string) ([]string, error) {
	switch strings.ToLower(service) {
	case "rdp", "windows":
		return []string{"Security"}, nil // Windows Event Log name
	case "ssh":
		return []string{"/tmp/guardian_test_auth.log"}, nil
	default:
		return []string{"/tmp/guardian_test.log"}, nil
	}
}

// StartLogMonitoring simulates Windows Event Log monitoring
// Mimics: Get-WinEvent -FilterHashtable @{LogName='Security'; ID=4625}
func (m *MockProvider) StartLogMonitoring(ctx context.Context, logPath string, events chan<- core.LogEvent) error {
	m.mu.Lock()
	m.isRunning = true
	m.mu.Unlock()

	fmt.Printf("📊 [MOCK] Started monitoring %s (simulating Windows Security Event Log)\n", logPath)

	go m.simulateWindowsSecurityEvents(ctx, events)
	go m.startCleanupScheduler(ctx)

	return nil
}

// Service management methods (not implemented in mock)
func (m *MockProvider) InstallService() error   { return nil }
func (m *MockProvider) UninstallService() error { return nil }
func (m *MockProvider) StartService() error     { return nil }
func (m *MockProvider) StopService() error      { return nil }

func (m *MockProvider) ServiceStatus() (core.ServiceStatus, error) {
	return core.ServiceStatus{
		Running:   m.isRunning,
		PID:       0, // Mock PID
		StartTime: m.startTime,
	}, nil
}

// Helper functions
func formatExpiry(expiresAt *time.Time) string {
	if expiresAt == nil {
		return "never"
	}
	return expiresAt.Format("2006-01-02 15:04:05")
}
