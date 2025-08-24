//go:build windows
// +build windows

package parser

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/sr-tamim/guardian/pkg/models"
)

// WindowsEventLogParser parses Windows Security Event Log entries
// Specifically designed to process Event ID 4625 (Failed Logon) events
// This mirrors the functionality of your production PowerShell script
type WindowsEventLogParser struct {
	name     string
	patterns []string

	// Regex patterns matching your PowerShell script
	ipRegex        *regexp.Regexp
	usernameRegex  *regexp.Regexp
	eventIDRegex   *regexp.Regexp
	timestampRegex *regexp.Regexp
}

// NewWindowsEventLogParser creates a new Windows Event Log parser
// Uses the exact same regex pattern as your PowerShell script: "Source Network Address:\s+([\d\.]+)"
func NewWindowsEventLogParser() *WindowsEventLogParser {
	return &WindowsEventLogParser{
		name: "Windows Security Event Log",
		patterns: []string{
			"4625", // Failed logon event ID
			"Source Network Address",
			"Account Name",
			"Logon Type",
		},

		// These regex patterns match your PowerShell script exactly
		ipRegex:        regexp.MustCompile(`Source Network Address:\s+([\d\.]+)`),
		usernameRegex:  regexp.MustCompile(`Account Name:\s+([^\r\n\s]+)`),
		eventIDRegex:   regexp.MustCompile(`Event ID:\s+(\d+)`),
		timestampRegex: regexp.MustCompile(`Time Created:\s+([^\r\n]+)`),
	}
}

// ParseLine processes a Windows Event Log entry and extracts attack attempt information
// This is the Go equivalent of your PowerShell script's event processing logic
func (p *WindowsEventLogParser) ParseLine(line string) (*models.AttackAttempt, error) {
	// First, check if this is a relevant event (Event ID 4625)
	eventMatches := p.eventIDRegex.FindStringSubmatch(line)
	if len(eventMatches) < 2 {
		return nil, fmt.Errorf("no event ID found")
	}

	eventID, err := strconv.Atoi(eventMatches[1])
	if err != nil || eventID != 4625 {
		return nil, fmt.Errorf("not a failed logon event (ID: %d)", eventID)
	}

	// Extract IP address using the same regex as your PowerShell script
	ipMatches := p.ipRegex.FindStringSubmatch(line)
	if len(ipMatches) < 2 {
		return nil, fmt.Errorf("could not extract IP address from event")
	}

	sourceIP := strings.TrimSpace(ipMatches[1])

	// Skip invalid or local IPs (like your PowerShell script does)
	if sourceIP == "" || sourceIP == "-" || sourceIP == "127.0.0.1" || sourceIP == "::1" {
		return nil, fmt.Errorf("invalid or local IP address: %s", sourceIP)
	}

	// Extract username
	usernameMatches := p.usernameRegex.FindStringSubmatch(line)
	username := "unknown"
	if len(usernameMatches) >= 2 {
		username = strings.TrimSpace(usernameMatches[1])
		// Filter out system accounts (like your PowerShell script)
		if username == "-" || username == "" || strings.HasSuffix(username, "$") {
			username = "system_account"
		}
	}

	// Extract timestamp
	timestamp := time.Now() // Default to current time
	timestampMatches := p.timestampRegex.FindStringSubmatch(line)
	if len(timestampMatches) >= 2 {
		if parsedTime, err := time.Parse("2006-01-02T15:04:05.000000000Z", timestampMatches[1]); err == nil {
			timestamp = parsedTime
		}
	}

	// Determine severity based on username patterns (similar to your PS logic)
	severity := p.determineSeverity(username, sourceIP)

	return &models.AttackAttempt{
		Timestamp: timestamp,
		IP:        sourceIP,
		Service:   "RDP",
		Username:  username,
		Message:   p.formatLogMessage(sourceIP, username),
		Severity:  severity,
		Source:    "Security", // Windows Event Log name
		Blocked:   false,      // Will be set later if blocking occurs
	}, nil
}

// determineSeverity assesses threat level like your PowerShell script might
func (p *WindowsEventLogParser) determineSeverity(username, ip string) models.Severity {
	// High severity for admin accounts
	adminAccounts := []string{"administrator", "admin", "root", "sa"}
	for _, admin := range adminAccounts {
		if strings.EqualFold(username, admin) {
			return models.SeverityHigh
		}
	}

	// Medium severity for service accounts
	if strings.HasSuffix(strings.ToLower(username), "service") ||
		strings.HasSuffix(strings.ToLower(username), "svc") {
		return models.SeverityMedium
	}

	// Check for common dictionary attack patterns
	commonTargets := []string{"user", "test", "guest", "demo"}
	for _, target := range commonTargets {
		if strings.EqualFold(username, target) {
			return models.SeverityMedium
		}
	}

	return models.SeverityLow
}

// formatLogMessage creates a readable log message
func (p *WindowsEventLogParser) formatLogMessage(ip, username string) string {
	return fmt.Sprintf("Failed RDP logon attempt from %s for user '%s'", ip, username)
}

// ServiceName returns the service name this parser handles
func (p *WindowsEventLogParser) ServiceName() string {
	return "RDP"
}

// Patterns returns the regex patterns this parser uses
func (p *WindowsEventLogParser) Patterns() []string {
	return p.patterns
}

// IsRDPEvent checks if the log line is an RDP-related event
// This helps filter events efficiently like your PowerShell script
func (p *WindowsEventLogParser) IsRDPEvent(line string) bool {
	// Look for RDP-specific indicators
	rdpIndicators := []string{
		`Logon Type:\s+3`,  // Network logon (typical for RDP)
		`Logon Type:\s+10`, // RemoteInteractive logon (RDP)
		"Source Network Address",
		"Event ID: 4625",
	}

	for _, indicator := range rdpIndicators {
		if matched, _ := regexp.MatchString(indicator, line); matched {
			return true
		}
	}

	return false
}

// ParseEventXML handles structured Windows Event Log XML format
// For future enhancement when we integrate with Windows Event Log API
func (p *WindowsEventLogParser) ParseEventXML(xmlData string) (*models.AttackAttempt, error) {
	// TODO: Implement XML parsing for direct Windows Event Log API integration
	// This would allow reading events directly instead of parsing exported text
	return nil, fmt.Errorf("XML parsing not implemented yet")
}
