package mock

import (
	"context"
	"testing"
	"time"

	"github.com/sr-tamim/guardian/internal/core"
)

func TestProvider_Basic(t *testing.T) {
	provider := New("test-mock")

	if provider.Name() != "test-mock" {
		t.Errorf("expected name 'test-mock', got %s", provider.Name())
	}

	if !provider.IsSupported() {
		t.Error("expected mock provider to be supported")
	}

	if err := provider.RequirementsCheck(); err != nil {
		t.Errorf("unexpected requirements check error: %v", err)
	}
}

func TestProvider_Unsupported(t *testing.T) {
	provider := NewUnsupported("unsupported-mock")

	if provider.IsSupported() {
		t.Error("expected unsupported provider to not be supported")
	}

	err := provider.RequirementsCheck()
	if err == nil {
		t.Error("expected requirements check to fail for unsupported provider")
	}

	if !core.IsErrorCode(err, core.ErrPlatformNotSupported) {
		t.Errorf("expected ErrPlatformNotSupported, got %v", err)
	}
}

func TestProvider_IPBlocking(t *testing.T) {
	provider := New("test")
	testIP := "192.168.1.100"

	// Initially not blocked
	blocked, err := provider.IsBlocked(testIP)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
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
		t.Errorf("unexpected error: %v", err)
	}
	if !blocked {
		t.Error("IP should be blocked")
	}

	// List should include our IP
	blockedIPs, err := provider.ListBlockedIPs()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
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
		t.Errorf("unexpected error: %v", err)
	}

	// Should no longer be blocked
	blocked, err = provider.IsBlocked(testIP)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if blocked {
		t.Error("IP should not be blocked after unblock")
	}
}

func TestProvider_IPBlockingErrors(t *testing.T) {
	provider := New("test")

	// Test empty IP
	err := provider.BlockIP("", time.Hour, "test")
	if !core.IsErrorCode(err, core.ErrInvalidIP) {
		t.Errorf("expected ErrInvalidIP for empty IP, got %v", err)
	}

	// Test double block
	testIP := "10.0.0.1"
	err = provider.BlockIP(testIP, time.Hour, "first block")
	if err != nil {
		t.Errorf("unexpected error on first block: %v", err)
	}

	err = provider.BlockIP(testIP, time.Hour, "second block")
	if !core.IsErrorCode(err, core.ErrIPAlreadyBlocked) {
		t.Errorf("expected ErrIPAlreadyBlocked, got %v", err)
	}

	// Test unblock non-existent IP
	err = provider.UnblockIP("192.168.1.200")
	if !core.IsErrorCode(err, core.ErrIPNotBlocked) {
		t.Errorf("expected ErrIPNotBlocked, got %v", err)
	}

	// Test empty IP for unblock
	err = provider.UnblockIP("")
	if !core.IsErrorCode(err, core.ErrInvalidIP) {
		t.Errorf("expected ErrInvalidIP for empty IP, got %v", err)
	}
}

func TestProvider_BlockExpiration(t *testing.T) {
	provider := New("test")
	testIP := "172.16.0.1"

	// Block IP with very short duration
	duration := 50 * time.Millisecond
	err := provider.BlockIP(testIP, duration, "short block")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Should be blocked initially
	blocked, err := provider.IsBlocked(testIP)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !blocked {
		t.Error("IP should be blocked")
	}

	// Wait for expiration
	time.Sleep(duration + 10*time.Millisecond)

	// Should no longer be blocked
	blocked, err = provider.IsBlocked(testIP)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if blocked {
		t.Error("IP should not be blocked after expiration")
	}

	// Should not appear in blocked list
	blockedIPs, err := provider.ListBlockedIPs()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	for _, ip := range blockedIPs {
		if ip == testIP {
			t.Error("expired IP should not appear in blocked list")
		}
	}
}

func TestProvider_PermanentBlock(t *testing.T) {
	provider := New("test")
	testIP := "10.10.10.10"

	// Block IP permanently (duration = 0)
	err := provider.BlockIP(testIP, 0, "permanent block")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Should be blocked
	blocked, err := provider.IsBlocked(testIP)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !blocked {
		t.Error("IP should be blocked")
	}

	// Wait a bit and check again - should still be blocked
	time.Sleep(10 * time.Millisecond)
	blocked, err = provider.IsBlocked(testIP)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !blocked {
		t.Error("permanently blocked IP should still be blocked")
	}
}

func TestProvider_LogPaths(t *testing.T) {
	provider := New("test")

	// Initially no paths
	paths, err := provider.GetLogPaths("ssh")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(paths) != 0 {
		t.Error("expected no paths initially")
	}

	// Set paths
	expectedPaths := []string{"/var/log/auth.log", "/var/log/secure"}
	provider.SetLogPaths("ssh", expectedPaths)

	// Get paths
	paths, err = provider.GetLogPaths("ssh")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(paths) != len(expectedPaths) {
		t.Errorf("expected %d paths, got %d", len(expectedPaths), len(paths))
	}
	for i, path := range expectedPaths {
		if paths[i] != path {
			t.Errorf("expected path %s at index %d, got %s", path, i, paths[i])
		}
	}
}

func TestProvider_LogMonitoring(t *testing.T) {
	provider := New("test")
	logPath := "/var/log/test.log"

	// Initially not monitoring
	if provider.IsMonitoring(logPath) {
		t.Error("should not be monitoring initially")
	}

	// Start monitoring
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	events := make(chan core.LogEvent, 1)
	err := provider.StartLogMonitoring(ctx, logPath, events)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Should now be monitoring
	if !provider.IsMonitoring(logPath) {
		t.Error("should be monitoring after StartLogMonitoring")
	}

	// Should receive a test event
	select {
	case event := <-events:
		if event.Source != logPath {
			t.Errorf("expected source %s, got %s", logPath, event.Source)
		}
		if event.Service != "mock" {
			t.Errorf("expected service 'mock', got %s", event.Service)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("did not receive test event")
	}
}

func TestProvider_ServiceOperations(t *testing.T) {
	provider := New("test")

	// Test successful operations
	if err := provider.InstallService(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if err := provider.StartService(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	status, err := provider.ServiceStatus()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !status.Running {
		t.Error("expected service to be running")
	}
	if status.PID != 12345 {
		t.Errorf("expected PID 12345, got %d", status.PID)
	}

	if err := provider.StopService(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if err := provider.UninstallService(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestProvider_ServiceErrors(t *testing.T) {
	provider := New("test")
	testErr := core.NewError(core.ErrServicePermission, "permission denied", nil)

	// Set service error
	provider.SetServiceError(testErr)

	// All service operations should return the error
	if err := provider.InstallService(); err != testErr {
		t.Errorf("expected service error, got %v", err)
	}

	if err := provider.StartService(); err != testErr {
		t.Errorf("expected service error, got %v", err)
	}

	if err := provider.StopService(); err != testErr {
		t.Errorf("expected service error, got %v", err)
	}

	if err := provider.UninstallService(); err != testErr {
		t.Errorf("expected service error, got %v", err)
	}

	status, err := provider.ServiceStatus()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if status.Error != testErr {
		t.Errorf("expected status error to be %v, got %v", testErr, status.Error)
	}
}

func TestProvider_Reset(t *testing.T) {
	provider := New("test")

	// Add some data
	provider.BlockIP("192.168.1.1", time.Hour, "test")
	provider.SetLogPaths("ssh", []string{"/var/log/auth.log"})
	provider.SetServiceError(core.NewError(core.ErrServicePermission, "test error", nil))

	// Verify data exists
	blocked, _ := provider.IsBlocked("192.168.1.1")
	if !blocked {
		t.Error("IP should be blocked before reset")
	}

	paths, _ := provider.GetLogPaths("ssh")
	if len(paths) == 0 {
		t.Error("should have log paths before reset")
	}

	// Reset
	provider.Reset()

	// Verify data is cleared
	blocked, _ = provider.IsBlocked("192.168.1.1")
	if blocked {
		t.Error("IP should not be blocked after reset")
	}

	paths, _ = provider.GetLogPaths("ssh")
	if len(paths) != 0 {
		t.Error("should have no log paths after reset")
	}

	if err := provider.InstallService(); err != nil {
		t.Error("service error should be cleared after reset")
	}
}
