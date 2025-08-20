package common

import (
	"fmt"
	"testing"

	"github.com/sr-tamim/guardian/internal/core"
	"github.com/sr-tamim/guardian/pkg/models"
)

func TestSSHParser_ParseLine(t *testing.T) {
	parser := NewSSHParser()

	tests := []struct {
		name         string
		logLine      string
		expectAttack bool
		expectedIP   string
		expectedUser string
		expectedSev  models.Severity
		expectError  bool
	}{
		{
			name:         "failed password valid user",
			logLine:      "Jan 15 14:20:33 server sshd[12345]: Failed password for admin from 192.168.1.100 port 22",
			expectAttack: true,
			expectedIP:   "192.168.1.100",
			expectedUser: "admin",
			expectedSev:  models.SeverityMedium,
		},
		{
			name:         "failed password invalid user",
			logLine:      "Jan 15 14:20:33 server sshd[12345]: Failed password for invalid user root from 10.0.0.50 port 22",
			expectAttack: true,
			expectedIP:   "10.0.0.50",
			expectedUser: "root",
			expectedSev:  models.SeverityMedium,
		},
		{
			name:         "invalid user",
			logLine:      "Jan 15 14:20:33 server sshd[12345]: Invalid user hacker from 172.16.0.1 port 22",
			expectAttack: true,
			expectedIP:   "172.16.0.1",
			expectedUser: "hacker",
			expectedSev:  models.SeverityHigh,
		},
		{
			name:         "connection closed preauth",
			logLine:      "Jan 15 14:20:33 server sshd[12345]: Connection closed by invalid user test 203.0.113.1 port 22 [preauth]",
			expectAttack: true,
			expectedIP:   "203.0.113.1",
			expectedUser: "test",
			expectedSev:  models.SeverityLow,
		},
		{
			name:         "authentication failure",
			logLine:      "Jan 15 14:20:33 server sshd[12345]: authentication failure; logname= uid=0 euid=0 tty=ssh ruser= rhost=198.51.100.1",
			expectAttack: true,
			expectedIP:   "198.51.100.1",
			expectedUser: "unknown",
			expectedSev:  models.SeverityMedium,
		},
		{
			name:         "bad protocol version",
			logLine:      "Jan 15 14:20:33 server sshd[12345]: Bad protocol version identification 'random_string' from 203.0.113.50",
			expectAttack: true,
			expectedIP:   "203.0.113.50",
			expectedUser: "unknown",
			expectedSev:  models.SeverityHigh,
		},
		{
			name:         "illegal user",
			logLine:      "Jan 15 14:20:33 server sshd[12345]: Illegal user admin from 192.168.1.200",
			expectAttack: true,
			expectedIP:   "192.168.1.200",
			expectedUser: "admin",
			expectedSev:  models.SeverityHigh,
		},
		{
			name:         "IPv6 address",
			logLine:      "Jan 15 14:20:33 server sshd[12345]: Failed password for user from 2001:db8::1 port 22",
			expectAttack: true,
			expectedIP:   "2001:db8::1",
			expectedUser: "user",
			expectedSev:  models.SeverityMedium,
		},
		{
			name:         "successful login - no attack",
			logLine:      "Jan 15 14:20:33 server sshd[12345]: Accepted password for admin from 192.168.1.100 port 22",
			expectAttack: false,
		},
		{
			name:         "empty line",
			logLine:      "",
			expectAttack: false,
		},
		{
			name:         "non-ssh log line",
			logLine:      "Jan 15 14:20:33 server kernel: USB disconnect",
			expectAttack: false,
		},
		{
			name:         "invalid IP in log - should return nil",
			logLine:      "Jan 15 14:20:33 server sshd[12345]: Failed password for admin from invalid-ip port 22",
			expectAttack: false,
			expectError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			attack, err := parser.ParseLine(tt.logLine)

			if tt.expectError {
				if err == nil {
					t.Error("expected error but got none")
				}
				if !core.IsErrorCode(err, core.ErrLogParseError) {
					t.Errorf("expected ErrLogParseError, got %v", err)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if tt.expectAttack {
				if attack == nil {
					t.Error("expected attack to be detected but got nil")
					return
				}

				if attack.IP != tt.expectedIP {
					t.Errorf("expected IP %s, got %s", tt.expectedIP, attack.IP)
				}

				if attack.Username != tt.expectedUser {
					t.Errorf("expected username %s, got %s", tt.expectedUser, attack.Username)
				}

				if attack.Severity != tt.expectedSev {
					t.Errorf("expected severity %s, got %s", tt.expectedSev, attack.Severity)
				}

				if attack.Service != "ssh" {
					t.Errorf("expected service 'ssh', got %s", attack.Service)
				}

				if attack.Message != tt.logLine {
					t.Errorf("expected message to be the log line")
				}

				if attack.Blocked {
					t.Error("newly detected attacks should not be marked as blocked")
				}

				if !attack.IsValidIP() {
					t.Errorf("attack IP should be valid: %s", attack.IP)
				}
			} else {
				if attack != nil {
					t.Errorf("expected no attack but got: %+v", attack)
				}
			}
		})
	}
}

func TestSSHParser_ServiceName(t *testing.T) {
	parser := NewSSHParser()
	if parser.ServiceName() != "ssh" {
		t.Errorf("expected service name 'ssh', got %s", parser.ServiceName())
	}
}

func TestSSHParser_Patterns(t *testing.T) {
	parser := NewSSHParser()
	patterns := parser.Patterns()

	if len(patterns) == 0 {
		t.Error("expected at least one pattern")
	}

	for i, pattern := range patterns {
		if pattern == "" {
			t.Errorf("pattern %d should not be empty", i)
		}
	}
}

func TestSSHParser_GetSeverityForPattern(t *testing.T) {
	parser := NewSSHParser()

	tests := []struct {
		patternIndex     int
		expectedSeverity models.Severity
	}{
		{0, models.SeverityMedium}, // Failed password
		{1, models.SeverityHigh},   // Invalid user
		{2, models.SeverityLow},    // Connection closed preauth
		{3, models.SeverityMedium}, // Authentication failure
		{4, models.SeverityHigh},   // Bad protocol version
		{5, models.SeverityHigh},   // Illegal user
		{99, models.SeverityLow},   // Unknown pattern
	}

	for _, tt := range tests {
		t.Run(string(rune(tt.patternIndex+48)), func(t *testing.T) {
			severity := parser.GetSeverityForPattern(tt.patternIndex)
			if severity != tt.expectedSeverity {
				t.Errorf("expected severity %s for pattern %d, got %s",
					tt.expectedSeverity, tt.patternIndex, severity)
			}
		})
	}
}

func TestSSHParser_IsValidLogFormat(t *testing.T) {
	parser := NewSSHParser()

	tests := []struct {
		name    string
		logLine string
		isValid bool
	}{
		{
			name:    "valid syslog format",
			logLine: "Jan 15 14:20:33 server sshd[12345]: Failed password for admin",
			isValid: true,
		},
		{
			name:    "valid with different month",
			logLine: "Dec 31 23:59:59 host sshd[999]: Invalid user test",
			isValid: true,
		},
		{
			name:    "invalid - no sshd",
			logLine: "Jan 15 14:20:33 server kernel: USB disconnect",
			isValid: false,
		},
		{
			name:    "invalid - wrong format",
			logLine: "sshd: Failed password",
			isValid: false,
		},
		{
			name:    "empty line",
			logLine: "",
			isValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := parser.IsValidLogFormat(tt.logLine)
			if isValid != tt.isValid {
				t.Errorf("expected IsValidLogFormat(%s) = %v, got %v",
					tt.logLine, tt.isValid, isValid)
			}
		})
	}
}

func TestSSHParser_RealWorldLogs(t *testing.T) {
	parser := NewSSHParser()

	// Real-world SSH log samples
	realLogs := []string{
		"Jan 15 14:20:33 ubuntu sshd[1234]: Failed password for admin from 203.0.113.1 port 22 ssh2",
		"Feb  3 10:15:42 debian sshd[5678]: Invalid user oracle from 198.51.100.50 port 55432",
		"Mar 10 08:30:15 centos sshd[9999]: Illegal user postgres from 203.0.113.100",
		"Apr 22 16:45:30 server sshd[1111]: Connection closed by unknown 192.168.1.200 port 33445 [preauth]",
		"May  5 12:00:00 host sshd[2222]: Bad protocol version identification '\\003' from 10.0.0.100",
	}

	for i, logLine := range realLogs {
		t.Run(fmt.Sprintf("real_log_%d", i), func(t *testing.T) {
			attack, err := parser.ParseLine(logLine)
			if err != nil {
				t.Errorf("unexpected error parsing real log: %v", err)
				return
			}

			if attack == nil {
				t.Error("expected attack to be detected from real log")
				return
			}

			// Basic validation
			if attack.IP == "" {
				t.Error("attack IP should not be empty")
			}

			if attack.Service != "ssh" {
				t.Errorf("expected service 'ssh', got %s", attack.Service)
			}

			if !attack.IsValidIP() {
				t.Errorf("invalid IP detected: %s", attack.IP)
			}
		})
	}
}

// Benchmark the parser performance
func BenchmarkSSHParser_ParseLine(b *testing.B) {
	parser := NewSSHParser()
	logLine := "Jan 15 14:20:33 server sshd[12345]: Failed password for admin from 192.168.1.100 port 22"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = parser.ParseLine(logLine)
	}
}

func BenchmarkSSHParser_ParseLineNoMatch(b *testing.B) {
	parser := NewSSHParser()
	logLine := "Jan 15 14:20:33 server kernel: USB disconnect detected"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = parser.ParseLine(logLine)
	}
}

func TestSSHParser_ConcurrentParsing(t *testing.T) {
	parser := NewSSHParser()
	logLine := "Jan 15 14:20:33 server sshd[12345]: Failed password for admin from 192.168.1.100 port 22"

	// Test concurrent parsing to ensure thread safety
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func() {
			defer func() { done <- true }()
			for j := 0; j < 100; j++ {
				attack, err := parser.ParseLine(logLine)
				if err != nil {
					t.Errorf("unexpected error in concurrent parsing: %v", err)
					return
				}
				if attack == nil {
					t.Error("expected attack to be detected")
					return
				}
			}
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}
