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

const guardianRuleTag = "GuardianTag=Guardian"

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

	// Check firewall for existing Guardian rule (avoid duplicate rules after restarts)
	if w.guardianRuleExists(ip) {
		return core.NewError(core.ErrIPAlreadyBlocked, fmt.Sprintf("IP %s is already blocked (firewall rule exists)", ip), nil)
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
		fmt.Sprintf("description=Guardian IPS - Blocked due to failed login attempts (%s)", guardianRuleTag))

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

	// Use structured logging for firewall action
	logger.LogIPBlocked(w.config, ip, reason, ruleName, duration)
	logger.Info("IP blocked with Windows Firewall",
		"ip", ip,
		"rule", ruleName,
		"reason", reason,
		"duration", duration)

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

	// Use structured logging for firewall action
	logger.LogIPUnblocked(w.config, ip, ruleName, activeTime)
	logger.Info("IP unblocked from Windows Firewall",
		"ip", ip,
		"rule", ruleName,
		"activeTime", activeTime)

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

	// Use structured logging for monitoring events
	logger.LogMonitoringStart(w.config, "RDP", logPath, "WindowsProvider")
	logger.Info("Started monitoring Windows Event Log for RDP failures",
		"logPath", logPath,
		"provider", "WindowsProvider")

	// Start the event monitoring goroutine
	go w.monitorWindowsEventLog(ctx, events)

	// Start cleanup scheduler (like your PowerShell script's Remove-ExpiredRules)
	go w.startCleanupScheduler(ctx)

	return nil
}

// monitorWindowsEventLog monitors Windows Security Event Log for Event ID 4625
// This is the Go equivalent of your PowerShell Get-WinEvent command
func (w *WindowsProvider) monitorWindowsEventLog(ctx context.Context, events chan<- core.LogEvent) {
	// Use configurable check interval from configuration
	checkInterval := w.config.Monitoring.CheckInterval
	if checkInterval <= 0 {
		// Fallback to 10 seconds if not configured
		checkInterval = 10 * time.Second
		logger.Warn("CheckInterval not configured, using fallback", "fallback", "10s")
	}

	// Use configurable lookback duration with safety check
	lookbackDuration := w.config.Monitoring.LookbackDuration
	if lookbackDuration <= 0 {
		// Fallback to 1 hour if not configured
		lookbackDuration = 1 * time.Hour
		logger.Warn("LookbackDuration not configured or zero, using fallback", "fallback", "1h")
	}

	ticker := time.NewTicker(checkInterval)
	defer ticker.Stop()

	var lastEventTime time.Time = time.Now().Add(-lookbackDuration)

	// Log startup information
	logger.Info("Windows Event Log monitoring started",
		"checkInterval", checkInterval.String(),
		"eventID", "4625",
		"lookbackDuration", lookbackDuration.String(),
		"initialLookback", lastEventTime.Format("2006-01-02 15:04:05"))

	// Log configuration details for troubleshooting
	logger.Info("Monitoring configuration details",
		"configCheckInterval", w.config.Monitoring.CheckInterval.String(),
		"configLookbackDuration", w.config.Monitoring.LookbackDuration.String(),
		"configEnableRealTime", w.config.Monitoring.EnableRealTime)

	for {
		select {
		case <-ctx.Done():
			logger.Info("Windows Event Log monitoring stopped")
			return
		case <-ticker.C:
			// Use sliding window approach - always look back the full configured duration
			// This ensures we analyze attack patterns over the complete time window
			currentTime := time.Now()
			windowStart := currentTime.Add(-lookbackDuration)

			logger.Info("Checking Windows Event Log with sliding window",
				"windowStart", windowStart.Format("15:04:05"),
				"windowEnd", currentTime.Format("15:04:05"),
				"windowDuration", lookbackDuration.String(),
				"timezone", currentTime.Location().String())

			w.queryRecentEvents(windowStart, lookbackDuration, events)
		}
	}
}

// queryRecentEvents queries Windows Event Log for recent failed logon events
// Equivalent to your PowerShell: Get-WinEvent -FilterHashtable @{LogName='Security'; ID=4625}
func (w *WindowsProvider) queryRecentEvents(since time.Time, windowDuration time.Duration, events chan<- core.LogEvent) {
	// Windows Event Log @SystemTime queries require UTC format with Z suffix
	// This was confirmed by manual testing: UTC works, local time doesn't
	sinceUTC := since.UTC().Format("2006-01-02T15:04:05.000Z")

	// Keep local time for logging purposes
	sinceLocal := since.Format("2006-01-02T15:04:05.000")

	// Use UTC format (this is what @SystemTime expects)
	sinceStr := sinceUTC

	cmd := exec.Command("wevtutil", "qe", "Security",
		"/q:*[System[EventID=4625 and TimeCreated[@SystemTime>='"+sinceStr+"']]]",
		"/f:text",
		"/c:50") // Limit to 50 events per query

	logger.Info("Executing Windows Event Log query",
		"command", "wevtutil qe Security",
		"filter", fmt.Sprintf("EventID=4625 and TimeCreated>='%s'", sinceStr),
		"limit", 50,
		"localTimeQuery", sinceLocal,
		"utcTimeQuery", sinceUTC,
		"usingFormat", sinceStr)

	output, err := cmd.Output()
	if err != nil {
		// Log error with detailed information
		logger.Error("Failed to query Windows Event Log",
			"error", err,
			"command", cmd.String(),
			"sinceTime", sinceStr)

		// Try to get more detailed error info
		if exitError, ok := err.(*exec.ExitError); ok {
			logger.Error("Windows Event Log query stderr",
				"stderr", string(exitError.Stderr),
				"exitCode", exitError.ExitCode())
		}
		return
	}

	logger.Info("Event Log query completed",
		"outputBytes", len(output),
		"hasEvents", len(output) > 0,
		"since", since.Format("15:04:05"))

	if len(output) > 0 {
		// Log sample of output for debugging (first 200 chars)
		sampleOutput := string(output)
		if len(sampleOutput) > 200 {
			sampleOutput = sampleOutput[:200] + "..."
		}
		logger.Info("Event Log query found data",
			"sampleOutput", sampleOutput,
			"totalBytes", len(output))
	} else {
		logger.Info("No events found in specified time range",
			"timeRange", fmt.Sprintf("since %s", since.Format("15:04:05")))
	}

	// Parse the output and send events
	w.parseEventLogOutput(string(output), windowDuration, events)
}

// parseEventLogOutput processes wevtutil output and creates LogEvent entries
func (w *WindowsProvider) parseEventLogOutput(output string, windowDuration time.Duration, events chan<- core.LogEvent) {
	// Split output into individual events
	eventBlocks := strings.Split(output, "Event[")

	logger.Info("Parsing Event Log output",
		"totalBlocks", len(eventBlocks),
		"outputLength", len(output))

	parsedEvents := 0
	rdpEvents := 0
	attackCounts := make(map[string]int)
	uniqueIPs := make(map[string]struct{})
	for i, eventBlock := range eventBlocks {
		if strings.TrimSpace(eventBlock) == "" {
			continue
		}

		// Log each event block for debugging
		logger.Info("Processing event block",
			"blockNumber", i,
			"blockLength", len(eventBlock),
			"preview", func() string {
				if len(eventBlock) > 200 {
					return eventBlock[:200] + "..."
				}
				return eventBlock
			}())

		// Check if this looks like an RDP event
		isRDP := w.eventParser.IsRDPEvent(eventBlock)
		logger.Info("RDP event check result",
			"blockNumber", i,
			"isRDPEvent", isRDP)

		if isRDP {
			rdpEvents++
			event := core.LogEvent{
				Timestamp: time.Now(),
				Source:    "Security",
				Line:      eventBlock,
				Service:   "RDP",
			}

			if events != nil {
				select {
				case events <- event:
					w.mu.Lock()
					w.totalAttacks++
					w.mu.Unlock()
					parsedEvents++
					logger.Info("RDP attack event detected and sent for processing",
						"eventNumber", parsedEvents,
						"totalAttacks", w.totalAttacks)
				default:
					logger.Warn("Event channel full, dropping RDP attack event")
				}
			} else {
				attempt, err := w.eventParser.ParseLine(eventBlock)
				if err != nil {
					continue
				}

				attackCounts[attempt.IP]++
				uniqueIPs[attempt.IP] = struct{}{}

				logger.LogAttackAttempt(w.config, attempt.IP, attempt.Service, attempt.Username, attempt.Severity.String())
			}
		}
	}

	if events == nil {
		ips := make([]string, 0, len(uniqueIPs))
		for ip := range uniqueIPs {
			ips = append(ips, ip)
		}

		logger.LogEventLookup(w.config, "RDP", "Security", rdpEvents, ips)

		threshold := w.config.Blocking.FailureThreshold
		if threshold <= 0 {
			threshold = 1
		}

		for ip, count := range attackCounts {
			if count < threshold {
				continue
			}
			if w.isWhitelistedIP(ip) {
				logger.Info("Skipping whitelisted IP", "ip", ip)
				continue
			}

			reason := fmt.Sprintf("Failed logon threshold exceeded: %d attempts in %s", count, windowDuration.Truncate(time.Second))
			if err := w.BlockIP(ip, w.config.Blocking.BlockDuration, reason); err != nil {
				logger.Warn("Failed to block IP after threshold exceeded", "ip", ip, "error", err)
			}
		}
	}

	if parsedEvents > 0 {
		logger.Info("Parsed and processed RDP attack events",
			"eventCount", parsedEvents,
			"totalAttacks", w.totalAttacks)
	} else if rdpEvents == 0 && len(eventBlocks) > 1 {
		logger.Info("Event blocks found but no RDP events detected",
			"totalBlocks", len(eventBlocks),
			"rdpEvents", 0)
	}
}

func (w *WindowsProvider) isWhitelistedIP(ip string) bool {
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return false
	}

	for _, entry := range w.config.Blocking.WhitelistedIPs {
		candidate := strings.TrimSpace(entry)
		if candidate == "" {
			continue
		}
		if strings.Contains(candidate, "/") {
			_, network, err := net.ParseCIDR(candidate)
			if err == nil && network.Contains(parsed) {
				return true
			}
			continue
		}
		if candidate == ip {
			return true
		}
	}

	return false
}

func (w *WindowsProvider) guardianRuleExists(ip string) bool {
	cmd := exec.Command("netsh", "advfirewall", "firewall", "show", "rule", "name=all")
	output, err := cmd.Output()
	if err != nil {
		logger.Warn("Failed to query firewall rules", "error", err)
		return false
	}

	lines := strings.Split(string(output), "\n")
	currentHasTag := false
	currentMatchesIP := false

	checkRule := func() bool {
		return currentHasTag && currentMatchesIP
	}

	for _, raw := range lines {
		line := strings.TrimSpace(raw)
		if line == "" {
			if checkRule() {
				return true
			}
			currentHasTag = false
			currentMatchesIP = false
			continue
		}

		lower := strings.ToLower(line)
		if strings.HasPrefix(lower, "rule name") {
			if checkRule() {
				return true
			}
			currentHasTag = false
			currentMatchesIP = false
			continue
		}

		if strings.HasPrefix(lower, "description") && strings.Contains(line, guardianRuleTag) {
			currentHasTag = true
		}

		if strings.HasPrefix(lower, "remoteip") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				for _, candidate := range strings.Split(parts[1], ",") {
					if strings.TrimSpace(candidate) == ip {
						currentMatchesIP = true
						break
					}
				}
			}
		}
	}

	return checkRule()
}

// startCleanupScheduler runs periodic cleanup like your PowerShell script
func (w *WindowsProvider) startCleanupScheduler(ctx context.Context) {
	// Use configurable cleanup interval from configuration
	cleanupInterval := w.config.Blocking.CleanupInterval
	if cleanupInterval <= 0 {
		// Fallback to 5 minutes if not configured
		cleanupInterval = 5 * time.Minute
	}

	ticker := time.NewTicker(cleanupInterval)
	defer ticker.Stop()

	// Log startup using structured logging
	logger.Info("Started Windows Firewall cleanup scheduler",
		"interval", cleanupInterval.String(),
		"configurable", true)

	for {
		select {
		case <-ctx.Done():
			logger.Info("Windows Firewall cleanup scheduler stopped")
			return
		case <-w.stopCleanup:
			logger.Info("Windows Firewall cleanup scheduler stopped by request")
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
	var removedIPs []string

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
					removedIPs = append(removedIPs, ip)

					elapsed := currentTime.Sub(record.BlockedAt)
					logger.Info("Removed expired Windows Firewall rule",
						"ip", ip,
						"rule", ruleName,
						"activeTime", elapsed.Truncate(time.Second))
				} else {
					logger.Error("Failed to remove expired firewall rule",
						"ip", ip,
						"rule", ruleName,
						"error", err)
				}
			}
		}
	}

	if removedCount > 0 {
		logger.Info("Cleanup operation completed",
			"removedRules", removedCount,
			"totalBlocked", len(w.blockedIPs),
			"removedIPs", removedIPs)

		// Use structured logging for cleanup events
		logger.LogCleanupOperation(w.config, removedCount, len(w.blockedIPs))
	} else {
		logger.Debug("Cleanup operation completed - no expired rules found")
	}
}
