package common

import (
	"fmt"
	"net"
	"regexp"
	"time"

	"github.com/sr-tamim/guardian/internal/core"
	"github.com/sr-tamim/guardian/pkg/models"
)

// SSHParser parses SSH authentication logs
type SSHParser struct {
	patterns []*regexp.Regexp
}

// NewSSHParser creates a new SSH log parser
func NewSSHParser() *SSHParser {
	// Common SSH attack patterns
	patterns := []*regexp.Regexp{
		// Failed password attempts (IPv4 and IPv6)
		regexp.MustCompile(`sshd\[\d+\]: Failed password for (?:invalid user )?(\w+) from ((?:\d{1,3}\.){3}\d{1,3}|[a-fA-F\d:]+) port \d+`),
		// Invalid user attempts
		regexp.MustCompile(`sshd\[\d+\]: Invalid user (\w+) from ((?:\d{1,3}\.){3}\d{1,3}|[a-fA-F\d:]+) port \d+`),
		// Connection closed due to preauth
		regexp.MustCompile(`sshd\[\d+\]: Connection closed by (?:invalid user )?(?:(\w+) )?((?:\d{1,3}\.){3}\d{1,3}|[a-fA-F\d:]+) port \d+ \[preauth\]`),
		// Authentication failure
		regexp.MustCompile(`sshd\[\d+\]: authentication failure; .* rhost=((?:\d{1,3}\.){3}\d{1,3}|[a-fA-F\d:]+)`),
		// Bad protocol version
		regexp.MustCompile(`sshd\[\d+\]: Bad protocol version identification .* from ((?:\d{1,3}\.){3}\d{1,3}|[a-fA-F\d:]+)`),
		// Illegal user
		regexp.MustCompile(`sshd\[\d+\]: Illegal user (\w+) from ((?:\d{1,3}\.){3}\d{1,3}|[a-fA-F\d:]+)`),
	}

	return &SSHParser{
		patterns: patterns,
	}
}

// ParseLine parses a single log line and returns an attack attempt if detected
func (p *SSHParser) ParseLine(line string) (*models.AttackAttempt, error) {
	if line == "" {
		return nil, nil // Not an error, just no attack detected
	}

	for i, pattern := range p.patterns {
		matches := pattern.FindStringSubmatch(line)
		if matches == nil {
			continue
		}

		var ip, username string
		var severity models.Severity

		switch i {
		case 0: // Failed password
			if len(matches) >= 3 {
				username = matches[1]
				ip = matches[2]
				severity = models.SeverityMedium
			}
		case 1: // Invalid user
			if len(matches) >= 3 {
				username = matches[1]
				ip = matches[2]
				severity = models.SeverityHigh
			}
		case 2: // Connection closed preauth
			if len(matches) >= 3 {
				if matches[1] != "" {
					username = matches[1]
				} else {
					username = "unknown"
				}
				ip = matches[2]
				severity = models.SeverityLow
			}
		case 3: // Authentication failure
			if len(matches) >= 2 {
				ip = matches[1]
				username = "unknown"
				severity = models.SeverityMedium
			}
		case 4: // Bad protocol version
			if len(matches) >= 2 {
				ip = matches[1]
				username = "unknown"
				severity = models.SeverityHigh
			}
		case 5: // Illegal user
			if len(matches) >= 3 {
				username = matches[1]
				ip = matches[2]
				severity = models.SeverityHigh
			}
		}

		// Validate IP address
		if ip != "" && net.ParseIP(ip) == nil {
			return nil, core.NewError(core.ErrLogParseError, 
				fmt.Sprintf("invalid IP address in log: %s", ip), nil)
		}

		// Skip if no IP was extracted
		if ip == "" {
			continue
		}

		return &models.AttackAttempt{
			Timestamp: time.Now(), // In real implementation, would parse from log
			IP:        ip,
			Service:   "ssh",
			Username:  username,
			Message:   line,
			Severity:  severity,
			Blocked:   false,
		}, nil
	}

	// No attack pattern matched
	return nil, nil
}

// ServiceName returns the service name this parser handles
func (p *SSHParser) ServiceName() string {
	return "ssh"
}

// Patterns returns the regex patterns used by this parser
func (p *SSHParser) Patterns() []string {
	patterns := make([]string, len(p.patterns))
	for i, pattern := range p.patterns {
		patterns[i] = pattern.String()
	}
	return patterns
}

// GetSeverityForPattern returns the severity level for a specific pattern
func (p *SSHParser) GetSeverityForPattern(patternIndex int) models.Severity {
	switch patternIndex {
	case 0: // Failed password
		return models.SeverityMedium
	case 1: // Invalid user
		return models.SeverityHigh
	case 2: // Connection closed preauth
		return models.SeverityLow
	case 3: // Authentication failure
		return models.SeverityMedium
	case 4: // Bad protocol version
		return models.SeverityHigh
	case 5: // Illegal user
		return models.SeverityHigh
	default:
		return models.SeverityLow
	}
}

// IsValidLogFormat checks if a log line appears to be in a supported format
func (p *SSHParser) IsValidLogFormat(line string) bool {
	// Check for common syslog format with sshd
	syslogPattern := regexp.MustCompile(`^\w+\s+\d+\s+\d+:\d+:\d+\s+\w+\s+sshd\[`)
	return syslogPattern.MatchString(line)
}