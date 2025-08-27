package mock

import (
	"context"
	"fmt"
	"math/rand"
	"regexp"
	"time"

	"github.com/sr-tamim/guardian/internal/core"
	"github.com/sr-tamim/guardian/pkg/logger"
	"github.com/sr-tamim/guardian/pkg/models"
)

// simulateWindowsSecurityEvents generates fake Windows Security Event Log entries
// Mimics the events that your PowerShell script processes (Event ID 4625)
func (m *MockProvider) simulateWindowsSecurityEvents(ctx context.Context, events chan<- core.LogEvent) {
	// Common attacking IPs for simulation
	attackerIPs := []string{
		"192.168.1.100",
		"10.0.0.45",
		"172.16.1.200",
		"203.0.113.15",
		"198.51.100.23",
	}

	// Common usernames attackers try
	usernames := []string{
		"administrator",
		"admin",
		"user",
		"test",
		"guest",
		"root",
	}

	ticker := time.NewTicker(2 * time.Second) // Generate attack every 2 seconds
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Simulate Windows Security Event Log entry (Event ID 4625 - Logon Failure)
			ip := attackerIPs[rand.Intn(len(attackerIPs))]
			username := usernames[rand.Intn(len(usernames))]

			// This mimics the Windows Event Log message format that your PowerShell script parses
			eventMessage := m.generateWindowsSecurityEventMessage(ip, username)

			event := core.LogEvent{
				Timestamp: time.Now(),
				Source:    "Security", // Windows Event Log name
				Line:      eventMessage,
				Service:   "RDP",
			}

			select {
			case events <- event:
				m.mu.Lock()
				m.totalAttacks++
				m.mu.Unlock()
				fmt.Printf("🚨 [MOCK] Generated Windows Security Event: Failed RDP logon from %s (user: %s)\n", ip, username)

				// Use structured logging for attack attempts if configured
				logger.LogAttackAttempt(m.config, ip, "RDP", username, "medium")
			default:
				// Channel is full, skip
			}
		}
	}
}

// generateWindowsSecurityEventMessage creates a realistic Windows Event Log message
// This matches the format that your PowerShell regex parses: "Source Network Address:\s+([\d\.]+)"
func (m *MockProvider) generateWindowsSecurityEventMessage(ip, username string) string {
	// This is a simplified version of a real Windows Security Event 4625 message
	return fmt.Sprintf(`Event ID: 4625
Task Category: Logon
Level: Information
Description: An account failed to log on.

Subject:
	Security ID: S-1-0-0
	Account Name: -
	Account Domain: -
	Logon ID: 0x0

Logon Type: 3

Account For Which Logon Failed:
	Security ID: S-1-0-0
	Account Name: %s
	Account Domain: WORKGROUP

Failure Information:
	Failure Reason: Unknown user name or bad password.
	Status: 0xC000006D
	Sub Status: 0xC000006A

Process Information:
	Caller Process ID: 0x0
	Caller Process Name: -

Network Information:
	Workstation Name: -
	Source Network Address: %s
	Source Port: 52341

Detailed Authentication Information:
	Logon Process: NtLmSsp
	Authentication Package: NTLM
	Transited Services: -
	Package Name (NTLM only): -
	Key Length: 0`, username, ip)
}

// startCleanupScheduler runs periodic cleanup of expired blocks
// Mimics the Remove-ExpiredRules function from your PowerShell script
func (m *MockProvider) startCleanupScheduler(ctx context.Context) {
	// Run cleanup every minute (faster than your 5-minute PowerShell interval for demo purposes)
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	fmt.Println("🧹 [MOCK] Started cleanup scheduler (runs every 1 minute)")

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.cleanupExpiredBlocks()
		}
	}
}

// cleanupExpiredBlocks removes expired firewall rules
// Mimics the Remove-ExpiredRules function from your PowerShell script
func (m *MockProvider) cleanupExpiredBlocks() {
	m.mu.Lock()
	defer m.mu.Unlock()

	currentTime := time.Now()
	removedCount := 0

	// Check firewall rules for expiration (like your PowerShell script)
	for ruleName, rule := range m.firewallRules {
		if rule.IsActive && rule.ExpiresAt != nil {
			// Calculate elapsed time (like your PowerShell: ($CurrentTime - $RuleTime).TotalMinutes)
			elapsed := currentTime.Sub(rule.CreatedAt)

			if currentTime.After(*rule.ExpiresAt) {
				// Remove expired rule
				rule.IsActive = false

				// Update corresponding block record
				if blockRecord, exists := m.blockedIPs[rule.IP]; exists {
					blockRecord.IsActive = false
					blockRecord.UnblockedAt = &currentTime
				}

				removedCount++
				fmt.Printf("🧹 [MOCK] Removed expired rule: %s (was active for %v)\n",
					ruleName, elapsed.Truncate(time.Second))
			}
		}
	}

	if removedCount > 0 {
		fmt.Printf("✅ [MOCK] Cleanup completed: removed %d expired rules\n", removedCount)

		// Use structured logging for cleanup events if configured
		logger.LogCleanupOperation(m.config, removedCount, len(m.firewallRules))
	}
}

// parseWindowsSecurityEvent simulates parsing a Windows Security Event
// Uses the same regex pattern as your PowerShell script
func (m *MockProvider) parseWindowsSecurityEvent(eventMessage string) (*models.AttackAttempt, error) {
	// This is the exact regex from your PowerShell script!
	ipRegex := regexp.MustCompile(`Source Network Address:\s+([\d\.]+)`)
	usernameRegex := regexp.MustCompile(`Account Name:\s+([^\r\n]+)`)

	ipMatches := ipRegex.FindStringSubmatch(eventMessage)
	if len(ipMatches) < 2 {
		return nil, fmt.Errorf("could not extract IP address from event")
	}

	usernameMatches := usernameRegex.FindStringSubmatch(eventMessage)
	username := "unknown"
	if len(usernameMatches) >= 2 {
		username = usernameMatches[1]
	}

	return &models.AttackAttempt{
		Timestamp: time.Now(),
		IP:        ipMatches[1],
		Service:   "RDP",
		Username:  username,
		Message:   eventMessage,
		Severity:  models.SeverityMedium,
		Source:    "Security", // Windows Event Log
		Blocked:   false,
	}, nil
}

// GetStatistics returns current mock statistics
func (m *MockProvider) GetStatistics() (*models.Statistics, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	activeBlocks := int64(0)
	for _, record := range m.blockedIPs {
		if record.IsActive {
			activeBlocks++
		}
	}

	return &models.Statistics{
		TotalAttacks:      m.totalAttacks,
		BlockedIPs:        m.totalBlocks,
		ActiveBlocks:      activeBlocks,
		ServicesMonitored: 1, // RDP
		UptimeSeconds:     int64(time.Since(m.startTime).Seconds()),
		LastActivity:      time.Now(),
	}, nil
}
