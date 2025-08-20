package detector

import (
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/sr-tamim/guardian/internal/core"
	"github.com/sr-tamim/guardian/pkg/models"
)

// BasicThreatDetector implements basic threat detection logic
type BasicThreatDetector struct {
	mu              sync.RWMutex
	whitelistedIPs  map[string]bool
	whitelistedNets []*net.IPNet
	config          ThreatConfig
}

// ThreatConfig holds configuration for threat detection
type ThreatConfig struct {
	FailureThreshold    int
	LookbackDuration    time.Duration
	SeverityMultipliers map[models.Severity]float64
	WhitelistedIPs      []string
}

// NewBasicThreatDetector creates a new basic threat detector
func NewBasicThreatDetector(config ThreatConfig) (*BasicThreatDetector, error) {
	detector := &BasicThreatDetector{
		whitelistedIPs: make(map[string]bool),
		config:         config,
	}

	// Set default severity multipliers if not provided
	if detector.config.SeverityMultipliers == nil {
		detector.config.SeverityMultipliers = map[models.Severity]float64{
			models.SeverityLow:      1.0,
			models.SeverityMedium:   2.0,
			models.SeverityHigh:     3.0,
			models.SeverityCritical: 5.0,
		}
	}

	// Parse whitelisted IPs and networks
	err := detector.updateWhitelist(config.WhitelistedIPs)
	if err != nil {
		return nil, err
	}

	return detector, nil
}

// updateWhitelist parses and updates the whitelist
func (d *BasicThreatDetector) updateWhitelist(ips []string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.whitelistedIPs = make(map[string]bool)
	d.whitelistedNets = nil

	for _, ipStr := range ips {
		if strings.Contains(ipStr, "/") {
			// CIDR network
			_, network, err := net.ParseCIDR(ipStr)
			if err != nil {
				return core.NewError(core.ErrConfigInvalid, 
					"invalid CIDR network: "+ipStr, err)
			}
			d.whitelistedNets = append(d.whitelistedNets, network)
		} else {
			// Single IP
			if net.ParseIP(ipStr) == nil {
				return core.NewError(core.ErrConfigInvalid, 
					"invalid IP address: "+ipStr, nil)
			}
			d.whitelistedIPs[ipStr] = true
		}
	}

	return nil
}

// AnalyzeAttack analyzes a single attack attempt and returns a threat assessment
func (d *BasicThreatDetector) AnalyzeAttack(attempt *models.AttackAttempt) core.ThreatAssessment {
	if attempt == nil {
		return core.ThreatAssessment{
			Severity:          models.SeverityLow,
			Confidence:        0.0,
			ShouldBlock:       false,
			Reason:            "nil attack attempt",
			RecommendedAction: "ignore",
		}
	}

	// Check if IP is whitelisted
	if d.IsWhitelisted(attempt.IP) {
		return core.ThreatAssessment{
			Severity:          models.SeverityLow,
			Confidence:        0.0,
			ShouldBlock:       false,
			Reason:            "IP is whitelisted",
			RecommendedAction: "ignore - whitelisted",
		}
	}

	// Base confidence based on severity
	severityMultiplier := d.config.SeverityMultipliers[attempt.Severity]
	baseConfidence := float64(attempt.Severity+1) / 4.0 // 0.25, 0.5, 0.75, 1.0

	// Adjust confidence based on various factors
	confidence := baseConfidence * severityMultiplier / 5.0 // Normalize

	// Username-based adjustments
	suspiciousUsernames := []string{"root", "admin", "administrator", "test", "oracle", "postgres"}
	for _, suspicious := range suspiciousUsernames {
		if strings.EqualFold(attempt.Username, suspicious) {
			confidence += 0.2
			break
		}
	}

	// Ensure confidence is within bounds
	if confidence > 1.0 {
		confidence = 1.0
	}

	// Determine if should block based on severity and confidence
	shouldBlock := confidence >= 0.7 || attempt.Severity >= models.SeverityHigh

	// Generate reason and recommendation
	reason := fmt.Sprintf("Severity: %s, Confidence: %.2f, User: %s", 
		attempt.Severity, confidence, attempt.Username)

	var recommendedAction string
	if shouldBlock {
		if attempt.Severity == models.SeverityCritical {
			recommendedAction = "immediate block for 24 hours"
		} else if attempt.Severity == models.SeverityHigh {
			recommendedAction = "block for 1 hour"
		} else {
			recommendedAction = "block for 30 minutes"
		}
	} else {
		recommendedAction = "monitor and track"
	}

	return core.ThreatAssessment{
		Severity:          attempt.Severity,
		Confidence:        confidence,
		ShouldBlock:       shouldBlock,
		Reason:            reason,
		RecommendedAction: recommendedAction,
	}
}

// ShouldBlock determines if an IP should be blocked based on attack history
func (d *BasicThreatDetector) ShouldBlock(ip string, attempts []*models.AttackAttempt) bool {
	if d.IsWhitelisted(ip) {
		return false
	}

	if len(attempts) == 0 {
		return false
	}

	// Count recent attempts within lookback period
	now := time.Now()
	recentAttempts := 0
	totalSeverityScore := 0.0

	for _, attempt := range attempts {
		if now.Sub(attempt.Timestamp) <= d.config.LookbackDuration {
			recentAttempts++
			// Add severity-weighted score
			multiplier := d.config.SeverityMultipliers[attempt.Severity]
			totalSeverityScore += multiplier
		}
	}

	// Block if we exceed threshold by count or by severity score
	return recentAttempts >= d.config.FailureThreshold || 
		   totalSeverityScore >= float64(d.config.FailureThreshold)*2.0
}

// IsWhitelisted checks if an IP is whitelisted
func (d *BasicThreatDetector) IsWhitelisted(ip string) bool {
	if ip == "" {
		return false
	}

	d.mu.RLock()
	defer d.mu.RUnlock()

	// Check exact IP match
	if d.whitelistedIPs[ip] {
		return true
	}

	// Check network ranges
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}

	for _, network := range d.whitelistedNets {
		if network.Contains(parsedIP) {
			return true
		}
	}

	return false
}

// UpdateWhitelist updates the whitelist configuration
func (d *BasicThreatDetector) UpdateWhitelist(ips []string) error {
	return d.updateWhitelist(ips)
}

// GetConfig returns the current threat detection configuration
func (d *BasicThreatDetector) GetConfig() ThreatConfig {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.config
}

// SetFailureThreshold updates the failure threshold
func (d *BasicThreatDetector) SetFailureThreshold(threshold int) {
	if threshold <= 0 {
		return
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	d.config.FailureThreshold = threshold
}

// SetLookbackDuration updates the lookback duration
func (d *BasicThreatDetector) SetLookbackDuration(duration time.Duration) {
	if duration <= 0 {
		return
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	d.config.LookbackDuration = duration
}

// GetThreatScore calculates a threat score for an IP based on attack history
func (d *BasicThreatDetector) GetThreatScore(attempts []*models.AttackAttempt) float64 {
	if len(attempts) == 0 {
		return 0.0
	}

	now := time.Now()
	score := 0.0

	for _, attempt := range attempts {
		age := now.Sub(attempt.Timestamp)
		if age > d.config.LookbackDuration {
			continue
		}

		// Decrease score based on age (more recent = higher score)
		ageWeight := 1.0 - (age.Seconds() / d.config.LookbackDuration.Seconds())
		severityMultiplier := d.config.SeverityMultipliers[attempt.Severity]
		
		score += ageWeight * severityMultiplier
	}

	return score
}