package mock

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/sr-tamim/guardian/internal/core"
)

// Provider implements a mock platform provider for testing
type Provider struct {
	mu          sync.RWMutex
	name        string
	supported   bool
	blockedIPs  map[string]blockInfo
	logPaths    map[string][]string
	serviceErr  error
	monitoring  map[string]bool // logPath -> monitoring status
}

type blockInfo struct {
	blockedAt time.Time
	duration  time.Duration
	reason    string
}

// New creates a new mock platform provider
func New(name string) *Provider {
	return &Provider{
		name:       name,
		supported:  true,
		blockedIPs: make(map[string]blockInfo),
		logPaths:   make(map[string][]string),
		monitoring: make(map[string]bool),
	}
}

// NewUnsupported creates a mock provider that reports as unsupported
func NewUnsupported(name string) *Provider {
	return &Provider{
		name:       name,
		supported:  false,
		blockedIPs: make(map[string]blockInfo),
		logPaths:   make(map[string][]string),
		monitoring: make(map[string]bool),
	}
}

// Name returns the platform name
func (p *Provider) Name() string {
	return p.name
}

// IsSupported returns whether this platform is supported
func (p *Provider) IsSupported() bool {
	return p.supported
}

// RequirementsCheck checks if all requirements are met
func (p *Provider) RequirementsCheck() error {
	if !p.supported {
		return core.NewError(core.ErrPlatformNotSupported, "mock platform not supported", nil)
	}
	return nil
}

// BlockIP blocks an IP address
func (p *Provider) BlockIP(ip string, duration time.Duration, reason string) error {
	if ip == "" {
		return core.NewError(core.ErrInvalidIP, "IP address cannot be empty", nil)
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	// Check if already blocked
	if _, exists := p.blockedIPs[ip]; exists {
		return core.NewError(core.ErrIPAlreadyBlocked, 
			fmt.Sprintf("IP %s is already blocked", ip), nil)
	}

	p.blockedIPs[ip] = blockInfo{
		blockedAt: time.Now(),
		duration:  duration,
		reason:    reason,
	}

	return nil
}

// UnblockIP unblocks an IP address
func (p *Provider) UnblockIP(ip string) error {
	if ip == "" {
		return core.NewError(core.ErrInvalidIP, "IP address cannot be empty", nil)
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	if _, exists := p.blockedIPs[ip]; !exists {
		return core.NewError(core.ErrIPNotBlocked, 
			fmt.Sprintf("IP %s is not blocked", ip), nil)
	}

	delete(p.blockedIPs, ip)
	return nil
}

// IsBlocked checks if an IP is blocked
func (p *Provider) IsBlocked(ip string) (bool, error) {
	if ip == "" {
		return false, core.NewError(core.ErrInvalidIP, "IP address cannot be empty", nil)
	}

	p.mu.RLock()
	defer p.mu.RUnlock()

	info, exists := p.blockedIPs[ip]
	if !exists {
		return false, nil
	}

	// Check if block has expired (if not permanent)
	if info.duration > 0 && time.Since(info.blockedAt) > info.duration {
		// Clean up expired block
		p.mu.RUnlock()
		p.mu.Lock()
		delete(p.blockedIPs, ip)
		p.mu.Unlock()
		p.mu.RLock()
		return false, nil
	}

	return true, nil
}

// ListBlockedIPs returns list of blocked IPs
func (p *Provider) ListBlockedIPs() ([]string, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	var ips []string
	now := time.Now()

	for ip, info := range p.blockedIPs {
		// Skip expired blocks
		if info.duration > 0 && now.Sub(info.blockedAt) > info.duration {
			continue
		}
		ips = append(ips, ip)
	}

	return ips, nil
}

// GetLogPaths returns log paths for a service
func (p *Provider) GetLogPaths(service string) ([]string, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	paths, exists := p.logPaths[service]
	if !exists {
		return []string{}, nil
	}

	return paths, nil
}

// StartLogMonitoring starts monitoring a log file
func (p *Provider) StartLogMonitoring(ctx context.Context, logPath string, events chan<- core.LogEvent) error {
	p.mu.Lock()
	p.monitoring[logPath] = true
	p.mu.Unlock()

	// In a real implementation, this would start file monitoring
	// For the mock, we'll just simulate by sending a test event
	go func() {
		select {
		case <-ctx.Done():
			return
		case events <- core.LogEvent{
			Timestamp: time.Now(),
			Source:    logPath,
			Line:      "Mock log entry for testing",
			Service:   "mock",
		}:
		}
	}()

	return nil
}

// InstallService installs Guardian as a system service
func (p *Provider) InstallService() error {
	return p.serviceErr
}

// UninstallService uninstalls Guardian system service
func (p *Provider) UninstallService() error {
	return p.serviceErr
}

// StartService starts the Guardian service
func (p *Provider) StartService() error {
	return p.serviceErr
}

// StopService stops the Guardian service
func (p *Provider) StopService() error {
	return p.serviceErr
}

// ServiceStatus returns the current service status
func (p *Provider) ServiceStatus() (core.ServiceStatus, error) {
	return core.ServiceStatus{
		Running:   true,
		PID:       12345,
		StartTime: time.Now().Add(-time.Hour),
		Error:     p.serviceErr,
	}, nil
}

// SetServiceError sets an error to be returned by service operations
func (p *Provider) SetServiceError(err error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.serviceErr = err
}

// SetLogPaths sets the log paths for a service
func (p *Provider) SetLogPaths(service string, paths []string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.logPaths[service] = paths
}

// IsMonitoring returns whether a log path is being monitored
func (p *Provider) IsMonitoring(logPath string) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.monitoring[logPath]
}

// GetBlockInfo returns block information for an IP (for testing)
func (p *Provider) GetBlockInfo(ip string) (blockInfo, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	info, exists := p.blockedIPs[ip]
	return info, exists
}

// Reset clears all state (for testing)
func (p *Provider) Reset() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.blockedIPs = make(map[string]blockInfo)
	p.logPaths = make(map[string][]string)
	p.monitoring = make(map[string]bool)
	p.serviceErr = nil
}