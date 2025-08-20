package core

import (
	"context"
	"testing"
	"time"

	"github.com/sr-tamim/guardian/pkg/models"
)

// Mock implementations for testing

type mockPlatformProvider struct {
	name        string
	supported   bool
	blockedIPs  map[string]bool
	logPaths    map[string][]string
	serviceErr  error
}

func newMockPlatformProvider(name string, supported bool) *mockPlatformProvider {
	return &mockPlatformProvider{
		name:       name,
		supported:  supported,
		blockedIPs: make(map[string]bool),
		logPaths:   make(map[string][]string),
	}
}

func (m *mockPlatformProvider) Name() string {
	return m.name
}

func (m *mockPlatformProvider) IsSupported() bool {
	return m.supported
}

func (m *mockPlatformProvider) RequirementsCheck() error {
	return nil
}

func (m *mockPlatformProvider) BlockIP(ip string, duration time.Duration, reason string) error {
	m.blockedIPs[ip] = true
	return nil
}

func (m *mockPlatformProvider) UnblockIP(ip string) error {
	delete(m.blockedIPs, ip)
	return nil
}

func (m *mockPlatformProvider) IsBlocked(ip string) (bool, error) {
	return m.blockedIPs[ip], nil
}

func (m *mockPlatformProvider) ListBlockedIPs() ([]string, error) {
	var ips []string
	for ip := range m.blockedIPs {
		ips = append(ips, ip)
	}
	return ips, nil
}

func (m *mockPlatformProvider) GetLogPaths(service string) ([]string, error) {
	return m.logPaths[service], nil
}

func (m *mockPlatformProvider) StartLogMonitoring(ctx context.Context, logPath string, events chan<- LogEvent) error {
	return nil
}

func (m *mockPlatformProvider) InstallService() error {
	return m.serviceErr
}

func (m *mockPlatformProvider) UninstallService() error {
	return m.serviceErr
}

func (m *mockPlatformProvider) StartService() error {
	return m.serviceErr
}

func (m *mockPlatformProvider) StopService() error {
	return m.serviceErr
}

func (m *mockPlatformProvider) ServiceStatus() (ServiceStatus, error) {
	return ServiceStatus{
		Running:   true,
		PID:       1234,
		StartTime: time.Now(),
		Error:     m.serviceErr,
	}, nil
}

func TestMockPlatformProvider(t *testing.T) {
	provider := newMockPlatformProvider("mock", true)

	t.Run("basic properties", func(t *testing.T) {
		if provider.Name() != "mock" {
			t.Errorf("expected name 'mock', got %s", provider.Name())
		}
		
		if !provider.IsSupported() {
			t.Error("expected platform to be supported")
		}
		
		if err := provider.RequirementsCheck(); err != nil {
			t.Errorf("unexpected requirements error: %v", err)
		}
	})

	t.Run("IP blocking", func(t *testing.T) {
		testIP := "192.168.1.100"
		
		// Initially not blocked
		blocked, err := provider.IsBlocked(testIP)
		if err != nil {
			t.Errorf("unexpected error checking block status: %v", err)
		}
		if blocked {
			t.Error("IP should not be blocked initially")
		}
		
		// Block IP
		err = provider.BlockIP(testIP, time.Hour, "test block")
		if err != nil {
			t.Errorf("unexpected error blocking IP: %v", err)
		}
		
		// Should now be blocked
		blocked, err = provider.IsBlocked(testIP)
		if err != nil {
			t.Errorf("unexpected error checking block status: %v", err)
		}
		if !blocked {
			t.Error("IP should be blocked after BlockIP call")
		}
		
		// List blocked IPs should include our IP
		blockedIPs, err := provider.ListBlockedIPs()
		if err != nil {
			t.Errorf("unexpected error listing blocked IPs: %v", err)
		}
		found := false
		for _, ip := range blockedIPs {
			if ip == testIP {
				found = true
				break
			}
		}
		if !found {
			t.Error("blocked IP not found in list")
		}
		
		// Unblock IP
		err = provider.UnblockIP(testIP)
		if err != nil {
			t.Errorf("unexpected error unblocking IP: %v", err)
		}
		
		// Should no longer be blocked
		blocked, err = provider.IsBlocked(testIP)
		if err != nil {
			t.Errorf("unexpected error checking block status: %v", err)
		}
		if blocked {
			t.Error("IP should not be blocked after UnblockIP call")
		}
	})
}

func TestLogEvent(t *testing.T) {
	event := LogEvent{
		Timestamp: time.Now(),
		Source:    "/var/log/auth.log",
		Line:      "test log line",
		Service:   "ssh",
	}
	
	if event.Source == "" {
		t.Error("source should not be empty")
	}
	
	if event.Service == "" {
		t.Error("service should not be empty")
	}
}

func TestThreatAssessment(t *testing.T) {
	assessment := ThreatAssessment{
		Severity:          models.SeverityHigh,
		Confidence:        0.95,
		ShouldBlock:       true,
		Reason:            "Multiple failed login attempts",
		RecommendedAction: "Block IP for 1 hour",
	}
	
	if assessment.Confidence < 0 || assessment.Confidence > 1 {
		t.Error("confidence should be between 0 and 1")
	}
	
	if assessment.Severity == models.SeverityHigh && !assessment.ShouldBlock {
		t.Error("high severity attacks should typically be blocked")
	}
}

func TestServiceStatus(t *testing.T) {
	status := ServiceStatus{
		Running:   true,
		PID:       1234,
		StartTime: time.Now().Add(-time.Hour),
		Error:     nil,
	}
	
	if !status.Running && status.PID > 0 {
		t.Error("running status and PID are inconsistent")
	}
	
	if status.StartTime.After(time.Now()) {
		t.Error("start time cannot be in the future")
	}
}

func TestGuardianStatus(t *testing.T) {
	status := GuardianStatus{
		Running:           true,
		StartTime:         time.Now().Add(-time.Hour),
		Platform:          "linux",
		MonitoredServices: []string{"ssh", "apache"},
		ActiveBlocks:      5,
		TotalAttacks:      123,
		Version:           "1.0.0",
		ConfigPath:        "/etc/guardian/config.yaml",
	}
	
	if len(status.MonitoredServices) == 0 {
		t.Error("should have at least one monitored service")
	}
	
	if status.ActiveBlocks < 0 {
		t.Error("active blocks cannot be negative")
	}
	
	if status.TotalAttacks < 0 {
		t.Error("total attacks cannot be negative")
	}
	
	if status.Platform == "" {
		t.Error("platform should not be empty")
	}
}