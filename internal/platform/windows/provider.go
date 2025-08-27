//go:build windows
// +build windows

package windows

import (
	"context"
	"fmt"
	"net"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/sr-tamim/guardian/internal/core"
	"github.com/sr-tamim/guardian/internal/parser"
	"github.com/sr-tamim/guardian/pkg/logger"
	"github.com/sr-tamim/guardian/pkg/models"
)

// WindowsProvider implements PlatformProvider for Windows systems
// This mirrors the functionality of your production PowerShell script
// Uses Windows Firewall (netsh/New-NetFirewallRule) and Windows Event Log
type WindowsProvider struct {
	mu         sync.RWMutex
	name       string
	config     *models.Config
	blockedIPs map[string]*models.BlockRecord
	isRunning  bool
	startTime  time.Time

	// Statistics
	totalAttacks int64
	totalBlocks  int64

	// Event log parser
	eventParser *parser.WindowsEventLogParser

	// Cleanup scheduler
	stopCleanup chan struct{}
}

// NewWindowsProvider creates a new Windows platform provider
func NewWindowsProvider(config *models.Config) *WindowsProvider {
	return &WindowsProvider{
		name:        "Windows Provider",
		config:      config,
		blockedIPs:  make(map[string]*models.BlockRecord),
		startTime:   time.Now(),
		eventParser: parser.NewWindowsEventLogParser(),
		stopCleanup: make(chan struct{}),
	}
}

// Name returns the provider name
func (w *WindowsProvider) Name() string {
	return w.name
}

// IsSupported checks if Windows firewall and event log are available
func (w *WindowsProvider) IsSupported() bool {
	// Check if we can access Windows Firewall
	cmd := exec.Command("netsh", "advfirewall", "firewall", "show", "rule", "name=all")
	return cmd.Run() == nil
}

// RequirementsCheck validates Windows-specific requirements
func (w *WindowsProvider) RequirementsCheck() error {
	// Check if running with appropriate privileges
	cmd := exec.Command("net", "session")
	if err := cmd.Run(); err != nil {
		return core.NewError(core.ErrPlatformRequirements,
			"Administrator privileges required for Windows Firewall management", err)
	}

	// Check Windows Firewall service
	cmd = exec.Command("sc", "query", "MpsSvc")
	if err := cmd.Run(); err != nil {
		return core.NewError(core.ErrPlatformRequirements,
			"Windows Firewall service not available", err)
	}

	return nil
}

// BlockIP creates a Windows Firewall rule to block an IP
// This mirrors your PowerShell script's New-NetFirewallRule command
func (w *WindowsProvider) BlockIP(ip string, duration time.Duration, reason string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Validate IP address
	if net.ParseIP(ip) == nil {
		return core.NewError(core.ErrInvalidIP, "invalid IP address", nil)
	}

	// Check if already blocked
	if existing, exists := w.blockedIPs[ip]; exists && existing.IsActive {
		return core.NewError(core.ErrIPAlreadyBlocked, fmt.Sprintf("IP %s is already blocked", ip), nil)
	}

	// Generate rule name using the configurable template (like your PS script)
	ruleName := w.config.Blocking.GenerateRuleName(ip, "RDP")

	// Create Windows Firewall rule using netsh (PowerShell equivalent)
	// Your script uses: New-NetFirewallRule -DisplayName $RuleName -Direction Inbound -RemoteAddress $IPAddr -Action Block
	cmd := exec.Command("netsh", "advfirewall", "firewall", "add", "rule",
		fmt.Sprintf("name=%s", ruleName),
		"dir=in",
		"action=block",
		fmt.Sprintf("remoteip=%s", ip),
		"description=Guardian IPS - Blocked due to failed login attempts")

	if err := cmd.Run(); err != nil {
		return core.NewError(core.ErrFirewallOperation,
			fmt.Sprintf("failed to create firewall rule for %s", ip), err)
	}

	// Calculate expiration time
	var expiresAt *time.Time
	if duration > 0 {
		expiry := time.Now().Add(duration)
		expiresAt = &expiry
	}

	// Create block record
	blockRecord := &models.BlockRecord{
		IP:          ip,
		BlockedAt:   time.Now(),
		ExpiresAt:   expiresAt,
		Reason:      reason,
		Service:     "RDP",
		AttackCount: 1,
		IsActive:    true,
	}
	w.blockedIPs[ip] = blockRecord
	w.totalBlocks++

	// Keep the console output for immediate feedback
	fmt.Printf("🚫 [WIN] Blocked IP %s with Windows Firewall rule: %s\n", ip, ruleName)

	// Use structured logging for firewall action if configured
	logger.LogIPBlocked(w.config, ip, reason, ruleName, duration)

	return nil
}

// UnblockIP removes a Windows Firewall rule
// Mirrors your PowerShell: Remove-NetFirewallRule -Name $Rule.Name
func (w *WindowsProvider) UnblockIP(ip string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	blockRecord, exists := w.blockedIPs[ip]
	if !exists || !blockRecord.IsActive {
		return core.NewError(core.ErrIPNotBlocked, fmt.Sprintf("IP %s is not blocked", ip), nil)
	}

	// Find the rule name (we stored it in our records)
	ruleName := w.config.Blocking.GenerateRuleName(ip, "RDP")

	// Remove Windows Firewall rule
	cmd := exec.Command("netsh", "advfirewall", "firewall", "delete", "rule",
		fmt.Sprintf("name=%s", ruleName))

	if err := cmd.Run(); err != nil {
		return core.NewError(core.ErrFirewallOperation,
			fmt.Sprintf("failed to remove firewall rule for %s", ip), err)
	}

	// Update block record
	blockRecord.IsActive = false
	now := time.Now()
	blockRecord.UnblockedAt = &now
	activeTime := now.Sub(blockRecord.BlockedAt)

	// Keep the console output for immediate feedback
	fmt.Printf("✅ [WIN] Unblocked IP %s (removed Windows Firewall rule: %s)\n", ip, ruleName)

	// Use structured logging for firewall action if configured
	logger.LogIPUnblocked(w.config, ip, ruleName, activeTime)

	return nil
}

// IsBlocked checks if an IP is currently blocked
func (w *WindowsProvider) IsBlocked(ip string) (bool, error) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	blockRecord, exists := w.blockedIPs[ip]
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
func (w *WindowsProvider) ListBlockedIPs() ([]string, error) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	var blocked []string
	for ip, record := range w.blockedIPs {
		if record.IsActive {
			// Check if expired
			if record.ExpiresAt == nil || time.Now().Before(*record.ExpiresAt) {
				blocked = append(blocked, ip)
			}
		}
	}

	return blocked, nil
}

// GetLogPaths returns Windows Event Log paths
func (w *WindowsProvider) GetLogPaths(service string) ([]string, error) {
	switch strings.ToLower(service) {
	case "rdp", "windows":
		return []string{"Security"}, nil // Windows Event Log name
	default:
		return []string{}, fmt.Errorf("unsupported service: %s", service)
	}
}

// StartLogMonitoring begins monitoring Windows Event Log
// This mimics your PowerShell script's Get-WinEvent -FilterHashtable approach
func (w *WindowsProvider) StartLogMonitoring(ctx context.Context, logPath string, events chan<- core.LogEvent) error {
	w.mu.Lock()
	w.isRunning = true
	w.mu.Unlock()

	// Keep the console output for immediate feedback
	fmt.Printf("📊 [WIN] Started monitoring Windows %s Event Log for RDP failures\n", logPath)

	// Use structured logging for monitoring events if configured
	logger.LogMonitoringStart(w.config, "RDP", logPath, "WindowsProvider")

	// Start the event monitoring goroutine
	go w.monitorWindowsEventLog(ctx, events)

	// Start cleanup scheduler (like your PowerShell script's Remove-ExpiredRules)
	go w.startCleanupScheduler(ctx)

	return nil
}

// monitorWindowsEventLog monitors Windows Security Event Log for Event ID 4625
// This is the Go equivalent of your PowerShell Get-WinEvent command
func (w *WindowsProvider) monitorWindowsEventLog(ctx context.Context, events chan<- core.LogEvent) {
	// Use wevtutil to query Windows Event Log (alternative to Get-WinEvent)
	// Query for failed logon events (Event ID 4625) in real-time
	ticker := time.NewTicker(10 * time.Second) // Check every 10 seconds
	defer ticker.Stop()

	var lastEventTime time.Time = time.Now().Add(-1 * time.Hour) // Look back 1 hour initially

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.queryRecentEvents(lastEventTime, events)
			lastEventTime = time.Now()
		}
	}
}

// queryRecentEvents queries Windows Event Log for recent failed logon events
// Equivalent to your PowerShell: Get-WinEvent -FilterHashtable @{LogName='Security'; ID=4625}
func (w *WindowsProvider) queryRecentEvents(since time.Time, events chan<- core.LogEvent) {
	// Use wevtutil to query events since the last check
	sinceStr := since.Format("2006-01-02T15:04:05.000Z")

	cmd := exec.Command("wevtutil", "qe", "Security",
		"/q:*[System[EventID=4625 and TimeCreated[@SystemTime>='"+sinceStr+"']]]",
		"/f:text",
		"/c:50") // Limit to 50 events per query

	output, err := cmd.Output()
	if err != nil {
		// Keep the console output for immediate feedback
		fmt.Printf("⚠️  [WIN] Error querying Windows Event Log: %v\n", err)

		// Log error using structured logging
		logger.Error("Failed to query Windows Event Log", "error", err)
		return
	}

	// Parse the output and send events
	w.parseEventLogOutput(string(output), events)
}

// parseEventLogOutput processes wevtutil output and creates LogEvent entries
func (w *WindowsProvider) parseEventLogOutput(output string, events chan<- core.LogEvent) {
	// Split output into individual events
	eventBlocks := strings.Split(output, "Event[")

	for _, eventBlock := range eventBlocks {
		if strings.TrimSpace(eventBlock) == "" {
			continue
		}

		// Only process if this looks like an RDP event
		if w.eventParser.IsRDPEvent(eventBlock) {
			event := core.LogEvent{
				Timestamp: time.Now(),
				Source:    "Security",
				Line:      eventBlock,
				Service:   "RDP",
			}

			select {
			case events <- event:
				w.mu.Lock()
				w.totalAttacks++
				w.mu.Unlock()
			default:
				// Channel is full, skip
			}
		}
	}
}

// startCleanupScheduler runs periodic cleanup like your PowerShell script
func (w *WindowsProvider) startCleanupScheduler(ctx context.Context) {
	// Your PowerShell script runs cleanup every 5 minutes
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	// Keep the console output for immediate feedback
	fmt.Println("🧹 [WIN] Started Windows Firewall cleanup scheduler (runs every 5 minutes)")

	// Log using structured logging
	logger.Info("Started Windows Firewall cleanup scheduler", "interval", "5 minutes")

	for {
		select {
		case <-ctx.Done():
			return
		case <-w.stopCleanup:
			return
		case <-ticker.C:
			w.cleanupExpiredRules()
		}
	}
}

// cleanupExpiredRules removes expired firewall rules
// This mirrors your PowerShell Remove-ExpiredRules function
func (w *WindowsProvider) cleanupExpiredRules() {
	w.mu.Lock()
	defer w.mu.Unlock()

	currentTime := time.Now()
	removedCount := 0

	for ip, record := range w.blockedIPs {
		if record.IsActive && record.ExpiresAt != nil {
			if currentTime.After(*record.ExpiresAt) {
				// Remove expired rule
				ruleName := w.config.Blocking.GenerateRuleName(ip, "RDP")

				cmd := exec.Command("netsh", "advfirewall", "firewall", "delete", "rule",
					fmt.Sprintf("name=%s", ruleName))

				if err := cmd.Run(); err == nil {
					record.IsActive = false
					record.UnblockedAt = &currentTime
					removedCount++

					elapsed := currentTime.Sub(record.BlockedAt)
					// Keep the console output for immediate feedback
					fmt.Printf("🧹 [WIN] Removed expired Windows Firewall rule for %s (was active for %v)\n",
						ip, elapsed.Truncate(time.Second))
				}
			}
		}
	}

	if removedCount > 0 {
		fmt.Printf("✅ [WIN] Cleanup completed: removed %d expired Windows Firewall rules\n", removedCount)

		// Use structured logging for cleanup events if configured
		logger.LogCleanupOperation(w.config, removedCount, len(w.blockedIPs))
	}
}
