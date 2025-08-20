package detector

import (
	"testing"
	"time"

	"github.com/sr-tamim/guardian/pkg/models"
)

func TestNewBasicThreatDetector(t *testing.T) {
	config := ThreatConfig{
		FailureThreshold: 3,
		LookbackDuration: 10 * time.Minute,
		WhitelistedIPs:   []string{"127.0.0.1", "192.168.0.0/16"},
	}

	detector, err := NewBasicThreatDetector(config)
	if err != nil {
		t.Errorf("unexpected error creating detector: %v", err)
	}

	if detector == nil {
		t.Fatal("detector should not be nil")
	}

	// Test default severity multipliers
	retrievedConfig := detector.GetConfig()
	if len(retrievedConfig.SeverityMultipliers) == 0 {
		t.Error("default severity multipliers should be set")
	}
}

func TestBasicThreatDetector_IsWhitelisted(t *testing.T) {
	config := ThreatConfig{
		WhitelistedIPs: []string{
			"127.0.0.1",
			"::1",
			"192.168.0.0/16",
			"10.0.0.0/8",
		},
	}

	detector, err := NewBasicThreatDetector(config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tests := []struct {
		name        string
		ip          string
		whitelisted bool
	}{
		{"localhost IPv4", "127.0.0.1", true},
		{"localhost IPv6", "::1", true},
		{"private network", "192.168.1.100", true},
		{"private network 10", "10.0.0.50", true},
		{"public IP", "8.8.8.8", false},
		{"empty IP", "", false},
		{"invalid IP", "invalid", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detector.IsWhitelisted(tt.ip)
			if result != tt.whitelisted {
				t.Errorf("IsWhitelisted(%s) = %v, want %v", tt.ip, result, tt.whitelisted)
			}
		})
	}
}

func TestBasicThreatDetector_AnalyzeAttack(t *testing.T) {
	config := ThreatConfig{
		FailureThreshold: 3,
		LookbackDuration: 10 * time.Minute,
		WhitelistedIPs:   []string{"127.0.0.1"},
	}

	detector, err := NewBasicThreatDetector(config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tests := []struct {
		name           string
		attack         *models.AttackAttempt
		expectBlock    bool
		minConfidence  float64
		maxConfidence  float64
	}{
		{
			name: "critical severity attack",
			attack: &models.AttackAttempt{
				IP:       "203.0.113.1",
				Username: "root",
				Severity: models.SeverityCritical,
			},
			expectBlock:   true,
			minConfidence: 0.7,
			maxConfidence: 1.0,
		},
		{
			name: "high severity attack",
			attack: &models.AttackAttempt{
				IP:       "198.51.100.1",
				Username: "admin",
				Severity: models.SeverityHigh,
			},
			expectBlock:   true,
			minConfidence: 0.5,
			maxConfidence: 1.0,
		},
		{
			name: "low severity attack",
			attack: &models.AttackAttempt{
				IP:       "203.0.113.50",
				Username: "user",
				Severity: models.SeverityLow,
			},
			expectBlock:   false,
			minConfidence: 0.0,
			maxConfidence: 0.7,
		},
		{
			name: "whitelisted IP",
			attack: &models.AttackAttempt{
				IP:       "127.0.0.1",
				Username: "root",
				Severity: models.SeverityCritical,
			},
			expectBlock:   false,
			minConfidence: 0.0,
			maxConfidence: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assessment := detector.AnalyzeAttack(tt.attack)

			if assessment.ShouldBlock != tt.expectBlock {
				t.Errorf("expected ShouldBlock = %v, got %v", tt.expectBlock, assessment.ShouldBlock)
			}

			if assessment.Confidence < tt.minConfidence || assessment.Confidence > tt.maxConfidence {
				t.Errorf("confidence %f not in expected range [%f, %f]", 
					assessment.Confidence, tt.minConfidence, tt.maxConfidence)
			}

			if assessment.Reason == "" {
				t.Error("reason should not be empty")
			}

			if assessment.RecommendedAction == "" {
				t.Error("recommended action should not be empty")
			}
		})
	}
}

func TestBasicThreatDetector_AnalyzeAttackNil(t *testing.T) {
	config := ThreatConfig{
		FailureThreshold: 3,
		LookbackDuration: 10 * time.Minute,
	}

	detector, err := NewBasicThreatDetector(config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assessment := detector.AnalyzeAttack(nil)
	
	if assessment.ShouldBlock {
		t.Error("nil attack should not result in blocking")
	}
	
	if assessment.Confidence != 0.0 {
		t.Errorf("expected confidence 0.0 for nil attack, got %f", assessment.Confidence)
	}
}

func TestBasicThreatDetector_ShouldBlock(t *testing.T) {
	config := ThreatConfig{
		FailureThreshold: 3,
		LookbackDuration: 10 * time.Minute,
		WhitelistedIPs:   []string{"127.0.0.1"},
	}

	detector, err := NewBasicThreatDetector(config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	now := time.Now()

	tests := []struct {
		name        string
		ip          string
		attempts    []*models.AttackAttempt
		shouldBlock bool
	}{
		{
			name: "threshold exceeded by count",
			ip:   "203.0.113.1",
			attempts: []*models.AttackAttempt{
				{IP: "203.0.113.1", Timestamp: now.Add(-1 * time.Minute), Severity: models.SeverityLow},
				{IP: "203.0.113.1", Timestamp: now.Add(-2 * time.Minute), Severity: models.SeverityLow},
				{IP: "203.0.113.1", Timestamp: now.Add(-3 * time.Minute), Severity: models.SeverityLow},
			},
			shouldBlock: true,
		},
		{
			name: "threshold exceeded by severity score",
			ip:   "198.51.100.1",
			attempts: []*models.AttackAttempt{
				{IP: "198.51.100.1", Timestamp: now.Add(-1 * time.Minute), Severity: models.SeverityCritical},
				{IP: "198.51.100.1", Timestamp: now.Add(-2 * time.Minute), Severity: models.SeverityHigh},
			},
			shouldBlock: true,
		},
		{
			name: "below threshold",
			ip:   "203.0.113.50",
			attempts: []*models.AttackAttempt{
				{IP: "203.0.113.50", Timestamp: now.Add(-1 * time.Minute), Severity: models.SeverityLow},
				{IP: "203.0.113.50", Timestamp: now.Add(-2 * time.Minute), Severity: models.SeverityLow},
			},
			shouldBlock: false,
		},
		{
			name: "old attempts outside lookback",
			ip:   "203.0.113.75",
			attempts: []*models.AttackAttempt{
				{IP: "203.0.113.75", Timestamp: now.Add(-20 * time.Minute), Severity: models.SeverityCritical},
				{IP: "203.0.113.75", Timestamp: now.Add(-30 * time.Minute), Severity: models.SeverityCritical},
				{IP: "203.0.113.75", Timestamp: now.Add(-40 * time.Minute), Severity: models.SeverityCritical},
			},
			shouldBlock: false,
		},
		{
			name: "whitelisted IP",
			ip:   "127.0.0.1",
			attempts: []*models.AttackAttempt{
				{IP: "127.0.0.1", Timestamp: now.Add(-1 * time.Minute), Severity: models.SeverityCritical},
				{IP: "127.0.0.1", Timestamp: now.Add(-2 * time.Minute), Severity: models.SeverityCritical},
				{IP: "127.0.0.1", Timestamp: now.Add(-3 * time.Minute), Severity: models.SeverityCritical},
			},
			shouldBlock: false,
		},
		{
			name:        "no attempts",
			ip:          "203.0.113.99",
			attempts:    []*models.AttackAttempt{},
			shouldBlock: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detector.ShouldBlock(tt.ip, tt.attempts)
			if result != tt.shouldBlock {
				t.Errorf("ShouldBlock(%s) = %v, want %v", tt.ip, result, tt.shouldBlock)
			}
		})
	}
}

func TestBasicThreatDetector_Configuration(t *testing.T) {
	config := ThreatConfig{
		FailureThreshold: 5,
		LookbackDuration: 30 * time.Minute,
	}

	detector, err := NewBasicThreatDetector(config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Test getting config
	retrievedConfig := detector.GetConfig()
	if retrievedConfig.FailureThreshold != 5 {
		t.Errorf("expected threshold 5, got %d", retrievedConfig.FailureThreshold)
	}

	if retrievedConfig.LookbackDuration != 30*time.Minute {
		t.Errorf("expected duration 30m, got %v", retrievedConfig.LookbackDuration)
	}

	// Test updating threshold
	detector.SetFailureThreshold(10)
	updatedConfig := detector.GetConfig()
	if updatedConfig.FailureThreshold != 10 {
		t.Errorf("expected updated threshold 10, got %d", updatedConfig.FailureThreshold)
	}

	// Test updating duration
	detector.SetLookbackDuration(1 * time.Hour)
	updatedConfig = detector.GetConfig()
	if updatedConfig.LookbackDuration != 1*time.Hour {
		t.Errorf("expected updated duration 1h, got %v", updatedConfig.LookbackDuration)
	}

	// Test invalid updates (should be ignored)
	detector.SetFailureThreshold(0)
	detector.SetLookbackDuration(0)
	finalConfig := detector.GetConfig()
	if finalConfig.FailureThreshold != 10 {
		t.Error("threshold should not change for invalid value")
	}
	if finalConfig.LookbackDuration != 1*time.Hour {
		t.Error("duration should not change for invalid value")
	}
}

func TestBasicThreatDetector_InvalidWhitelist(t *testing.T) {
	tests := []struct {
		name        string
		whitelist   []string
		expectError bool
	}{
		{
			name:        "valid whitelist",
			whitelist:   []string{"127.0.0.1", "192.168.0.0/16"},
			expectError: false,
		},
		{
			name:        "invalid IP",
			whitelist:   []string{"invalid-ip"},
			expectError: true,
		},
		{
			name:        "invalid CIDR",
			whitelist:   []string{"192.168.0.0/99"},
			expectError: true,
		},
		{
			name:        "empty whitelist",
			whitelist:   []string{},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := ThreatConfig{
				FailureThreshold: 3,
				LookbackDuration: 10 * time.Minute,
				WhitelistedIPs:   tt.whitelist,
			}

			_, err := NewBasicThreatDetector(config)
			
			if tt.expectError && err == nil {
				t.Error("expected error but got none")
			}
			
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestBasicThreatDetector_GetThreatScore(t *testing.T) {
	config := ThreatConfig{
		FailureThreshold: 3,
		LookbackDuration: 10 * time.Minute,
	}

	detector, err := NewBasicThreatDetector(config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	now := time.Now()

	tests := []struct {
		name     string
		attempts []*models.AttackAttempt
		minScore float64
	}{
		{
			name:     "no attempts",
			attempts: []*models.AttackAttempt{},
			minScore: 0.0,
		},
		{
			name: "recent high severity",
			attempts: []*models.AttackAttempt{
				{Timestamp: now.Add(-1 * time.Minute), Severity: models.SeverityHigh},
			},
			minScore: 2.0,
		},
		{
			name: "old attempts",
			attempts: []*models.AttackAttempt{
				{Timestamp: now.Add(-20 * time.Minute), Severity: models.SeverityCritical},
			},
			minScore: 0.0,
		},
		{
			name: "mixed recent and old",
			attempts: []*models.AttackAttempt{
				{Timestamp: now.Add(-1 * time.Minute), Severity: models.SeverityMedium},
				{Timestamp: now.Add(-20 * time.Minute), Severity: models.SeverityCritical},
			},
			minScore: 1.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := detector.GetThreatScore(tt.attempts)
			if score < tt.minScore {
				t.Errorf("expected score >= %f, got %f", tt.minScore, score)
			}
		})
	}
}

func TestBasicThreatDetector_UpdateWhitelist(t *testing.T) {
	config := ThreatConfig{
		FailureThreshold: 3,
		LookbackDuration: 10 * time.Minute,
		WhitelistedIPs:   []string{"127.0.0.1"},
	}

	detector, err := NewBasicThreatDetector(config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Initially whitelisted
	if !detector.IsWhitelisted("127.0.0.1") {
		t.Error("127.0.0.1 should be whitelisted initially")
	}

	// Update whitelist
	newWhitelist := []string{"192.168.1.0/24", "10.0.0.1"}
	err = detector.UpdateWhitelist(newWhitelist)
	if err != nil {
		t.Errorf("unexpected error updating whitelist: %v", err)
	}

	// Old IP should no longer be whitelisted
	if detector.IsWhitelisted("127.0.0.1") {
		t.Error("127.0.0.1 should not be whitelisted after update")
	}

	// New IPs should be whitelisted
	if !detector.IsWhitelisted("192.168.1.100") {
		t.Error("192.168.1.100 should be whitelisted")
	}

	if !detector.IsWhitelisted("10.0.0.1") {
		t.Error("10.0.0.1 should be whitelisted")
	}
}

// Benchmark threat analysis performance
func BenchmarkBasicThreatDetector_AnalyzeAttack(b *testing.B) {
	config := ThreatConfig{
		FailureThreshold: 3,
		LookbackDuration: 10 * time.Minute,
	}

	detector, err := NewBasicThreatDetector(config)
	if err != nil {
		b.Fatalf("unexpected error: %v", err)
	}

	attack := &models.AttackAttempt{
		IP:       "203.0.113.1",
		Username: "admin",
		Severity: models.SeverityHigh,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = detector.AnalyzeAttack(attack)
	}
}

func BenchmarkBasicThreatDetector_ShouldBlock(b *testing.B) {
	config := ThreatConfig{
		FailureThreshold: 3,
		LookbackDuration: 10 * time.Minute,
	}

	detector, err := NewBasicThreatDetector(config)
	if err != nil {
		b.Fatalf("unexpected error: %v", err)
	}

	now := time.Now()
	attempts := []*models.AttackAttempt{
		{IP: "203.0.113.1", Timestamp: now.Add(-1 * time.Minute), Severity: models.SeverityMedium},
		{IP: "203.0.113.1", Timestamp: now.Add(-2 * time.Minute), Severity: models.SeverityMedium},
		{IP: "203.0.113.1", Timestamp: now.Add(-3 * time.Minute), Severity: models.SeverityMedium},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = detector.ShouldBlock("203.0.113.1", attempts)
	}
}