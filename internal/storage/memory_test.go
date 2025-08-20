package storage

import (
	"fmt"
	"testing"
	"time"

	"github.com/sr-tamim/guardian/internal/core"
	"github.com/sr-tamim/guardian/pkg/models"
)

func TestMemoryStorage_SaveAndGetAttacks(t *testing.T) {
	storage := NewMemoryStorage()
	defer storage.Close()

	// Test saving an attack
	attack := &models.AttackAttempt{
		Timestamp: time.Now(),
		IP:        "192.168.1.100",
		Service:   "ssh",
		Username:  "admin",
		Message:   "Failed password",
		Severity:  models.SeverityHigh,
		Source:    "/var/log/auth.log",
		Blocked:   false,
	}

	err := storage.SaveAttack(attack)
	if err != nil {
		t.Errorf("unexpected error saving attack: %v", err)
	}

	if attack.ID == 0 {
		t.Error("attack ID should be assigned")
	}

	// Test getting attacks
	attacks, err := storage.GetAttacks(10, 0)
	if err != nil {
		t.Errorf("unexpected error getting attacks: %v", err)
	}

	if len(attacks) != 1 {
		t.Errorf("expected 1 attack, got %d", len(attacks))
	}

	retrievedAttack := attacks[0]
	if retrievedAttack.IP != attack.IP {
		t.Errorf("expected IP %s, got %s", attack.IP, retrievedAttack.IP)
	}

	if retrievedAttack.Service != attack.Service {
		t.Errorf("expected service %s, got %s", attack.Service, retrievedAttack.Service)
	}
}

func TestMemoryStorage_SaveAttackValidation(t *testing.T) {
	storage := NewMemoryStorage()
	defer storage.Close()

	// Test nil attack
	err := storage.SaveAttack(nil)
	if !core.IsErrorCode(err, core.ErrStorageOperation) {
		t.Errorf("expected ErrStorageOperation for nil attack, got %v", err)
	}

	// Test invalid IP
	attack := &models.AttackAttempt{
		IP:      "invalid-ip",
		Service: "ssh",
	}

	err = storage.SaveAttack(attack)
	if !core.IsErrorCode(err, core.ErrInvalidIP) {
		t.Errorf("expected ErrInvalidIP for invalid IP, got %v", err)
	}
}

func TestMemoryStorage_GetAttacksByIP(t *testing.T) {
	storage := NewMemoryStorage()
	defer storage.Close()

	testIP := "10.0.0.50"
	now := time.Now()

	// Save attacks for the test IP
	attacks := []*models.AttackAttempt{
		{
			Timestamp: now.Add(-2 * time.Hour),
			IP:        testIP,
			Service:   "ssh",
			Message:   "Old attack",
		},
		{
			Timestamp: now.Add(-30 * time.Minute),
			IP:        testIP,
			Service:   "ssh",
			Message:   "Recent attack",
		},
		{
			Timestamp: now.Add(-1 * time.Hour),
			IP:        "192.168.1.1", // Different IP
			Service:   "ssh",
			Message:   "Other IP attack",
		},
	}

	for _, attack := range attacks {
		err := storage.SaveAttack(attack)
		if err != nil {
			t.Errorf("unexpected error saving attack: %v", err)
		}
	}

	// Get attacks for test IP since 1 hour ago
	since := now.Add(-1 * time.Hour)
	ipAttacks, err := storage.GetAttacksByIP(testIP, since)
	if err != nil {
		t.Errorf("unexpected error getting attacks by IP: %v", err)
	}

	// Should only return the recent attack
	if len(ipAttacks) != 1 {
		t.Errorf("expected 1 attack, got %d", len(ipAttacks))
	}

	if ipAttacks[0].Message != "Recent attack" {
		t.Errorf("expected 'Recent attack', got %s", ipAttacks[0].Message)
	}

	// Test getting all attacks for the IP
	allIPAttacks, err := storage.GetAttacksByIP(testIP, time.Time{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if len(allIPAttacks) != 2 {
		t.Errorf("expected 2 attacks for IP, got %d", len(allIPAttacks))
	}

	// Should be sorted by timestamp (newest first)
	if allIPAttacks[0].Timestamp.Before(allIPAttacks[1].Timestamp) {
		t.Error("attacks should be sorted by timestamp (newest first)")
	}
}

func TestMemoryStorage_GetAttacksByIPValidation(t *testing.T) {
	storage := NewMemoryStorage()
	defer storage.Close()

	// Test empty IP
	_, err := storage.GetAttacksByIP("", time.Now())
	if !core.IsErrorCode(err, core.ErrInvalidIP) {
		t.Errorf("expected ErrInvalidIP for empty IP, got %v", err)
	}

	// Test non-existent IP
	attacks, err := storage.GetAttacksByIP("1.2.3.4", time.Now())
	if err != nil {
		t.Errorf("unexpected error for non-existent IP: %v", err)
	}
	if len(attacks) != 0 {
		t.Errorf("expected 0 attacks for non-existent IP, got %d", len(attacks))
	}
}

func TestMemoryStorage_Blocks(t *testing.T) {
	storage := NewMemoryStorage()
	defer storage.Close()

	testIP := "172.16.0.100"
	now := time.Now()
	expiresAt := now.Add(time.Hour)

	// Test saving a block
	block := &models.BlockRecord{
		IP:          testIP,
		BlockedAt:   now,
		ExpiresAt:   &expiresAt,
		Reason:      "Multiple failed logins",
		Service:     "ssh",
		AttackCount: 5,
		IsActive:    true,
	}

	err := storage.SaveBlock(block)
	if err != nil {
		t.Errorf("unexpected error saving block: %v", err)
	}

	if block.ID == 0 {
		t.Error("block ID should be assigned")
	}

	// Test getting block
	retrievedBlock, err := storage.GetBlock(testIP)
	if err != nil {
		t.Errorf("unexpected error getting block: %v", err)
	}

	if retrievedBlock.IP != block.IP {
		t.Errorf("expected IP %s, got %s", block.IP, retrievedBlock.IP)
	}

	if retrievedBlock.Reason != block.Reason {
		t.Errorf("expected reason %s, got %s", block.Reason, retrievedBlock.Reason)
	}

	// Test updating block
	block.AttackCount = 10
	block.Reason = "Updated reason"

	err = storage.UpdateBlock(block)
	if err != nil {
		t.Errorf("unexpected error updating block: %v", err)
	}

	updatedBlock, err := storage.GetBlock(testIP)
	if err != nil {
		t.Errorf("unexpected error getting updated block: %v", err)
	}

	if updatedBlock.AttackCount != 10 {
		t.Errorf("expected attack count 10, got %d", updatedBlock.AttackCount)
	}

	if updatedBlock.Reason != "Updated reason" {
		t.Errorf("expected updated reason, got %s", updatedBlock.Reason)
	}
}

func TestMemoryStorage_BlockValidation(t *testing.T) {
	storage := NewMemoryStorage()
	defer storage.Close()

	// Test nil block
	err := storage.SaveBlock(nil)
	if !core.IsErrorCode(err, core.ErrStorageOperation) {
		t.Errorf("expected ErrStorageOperation for nil block, got %v", err)
	}

	// Test empty IP
	block := &models.BlockRecord{
		IP:      "",
		Reason:  "test",
		Service: "ssh",
	}

	err = storage.SaveBlock(block)
	if !core.IsErrorCode(err, core.ErrInvalidIP) {
		t.Errorf("expected ErrInvalidIP for empty IP, got %v", err)
	}

	// Test getting non-existent block
	_, err = storage.GetBlock("192.168.1.200")
	if !core.IsErrorCode(err, core.ErrRecordNotFound) {
		t.Errorf("expected ErrRecordNotFound, got %v", err)
	}

	// Test updating non-existent block
	validBlock := &models.BlockRecord{
		IP:      "192.168.1.201",
		Reason:  "test",
		Service: "ssh",
	}

	err = storage.UpdateBlock(validBlock)
	if !core.IsErrorCode(err, core.ErrRecordNotFound) {
		t.Errorf("expected ErrRecordNotFound for update, got %v", err)
	}
}

func TestMemoryStorage_GetActiveBlocks(t *testing.T) {
	storage := NewMemoryStorage()
	defer storage.Close()

	now := time.Now()

	// Create blocks with different states
	blocks := []*models.BlockRecord{
		{
			IP:        "10.0.0.1",
			BlockedAt: now.Add(-2 * time.Hour),
			ExpiresAt: nil, // Permanent block
			Reason:    "Permanent ban",
			Service:   "ssh",
			IsActive:  true,
		},
		{
			IP:        "10.0.0.2",
			BlockedAt: now.Add(-30 * time.Minute),
			ExpiresAt: &[]time.Time{now.Add(30 * time.Minute)}[0], // Active block
			Reason:    "Temporary ban",
			Service:   "ssh",
			IsActive:  true,
		},
		{
			IP:        "10.0.0.3",
			BlockedAt: now.Add(-2 * time.Hour),
			ExpiresAt: &[]time.Time{now.Add(-30 * time.Minute)}[0], // Expired block
			Reason:    "Expired ban",
			Service:   "ssh",
			IsActive:  true,
		},
		{
			IP:        "10.0.0.4",
			BlockedAt: now.Add(-1 * time.Hour),
			ExpiresAt: &[]time.Time{now.Add(time.Hour)}[0], // Active but manually deactivated
			Reason:    "Manually unblocked",
			Service:   "ssh",
			IsActive:  false,
		},
	}

	for _, block := range blocks {
		err := storage.SaveBlock(block)
		if err != nil {
			t.Errorf("unexpected error saving block: %v", err)
		}
	}

	// Get active blocks
	activeBlocks, err := storage.GetActiveBlocks()
	if err != nil {
		t.Errorf("unexpected error getting active blocks: %v", err)
	}

	// Should only return the first two blocks (permanent and active temporary)
	if len(activeBlocks) != 2 {
		t.Errorf("expected 2 active blocks, got %d", len(activeBlocks))
	}

	// Check that they are sorted by blocked time (newest first)
	if len(activeBlocks) >= 2 {
		if activeBlocks[0].BlockedAt.Before(activeBlocks[1].BlockedAt) {
			t.Error("active blocks should be sorted by blocked time (newest first)")
		}
	}
}

func TestMemoryStorage_GetStatistics(t *testing.T) {
	storage := NewMemoryStorage()
	defer storage.Close()

	// Add some test data
	attack := &models.AttackAttempt{
		IP:      "192.168.1.1",
		Service: "ssh",
	}
	err := storage.SaveAttack(attack)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	block := &models.BlockRecord{
		IP:       "192.168.1.1",
		Reason:   "test",
		Service:  "ssh",
		IsActive: true,
	}
	err = storage.SaveBlock(block)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Get statistics
	stats, err := storage.GetStatistics()
	if err != nil {
		t.Errorf("unexpected error getting statistics: %v", err)
	}

	if stats.TotalAttacks != 1 {
		t.Errorf("expected 1 total attack, got %d", stats.TotalAttacks)
	}

	if stats.BlockedIPs != 1 {
		t.Errorf("expected 1 blocked IP, got %d", stats.BlockedIPs)
	}

	if stats.ActiveBlocks != 1 {
		t.Errorf("expected 1 active block, got %d", stats.ActiveBlocks)
	}

	if stats.ServicesMonitored != 1 {
		t.Errorf("expected 1 monitored service, got %d", stats.ServicesMonitored)
	}
}

func TestMemoryStorage_Pagination(t *testing.T) {
	storage := NewMemoryStorage()
	defer storage.Close()

	// Add multiple attacks
	for i := 0; i < 25; i++ {
		attack := &models.AttackAttempt{
			IP:      "192.168.1.1",
			Service: "ssh",
			Message: fmt.Sprintf("Attack %d", i),
		}
		err := storage.SaveAttack(attack)
		if err != nil {
			t.Errorf("unexpected error saving attack %d: %v", i, err)
		}
	}

	// Test first page
	attacks, err := storage.GetAttacks(10, 0)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(attacks) != 10 {
		t.Errorf("expected 10 attacks, got %d", len(attacks))
	}

	// Test second page
	attacks, err = storage.GetAttacks(10, 10)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(attacks) != 10 {
		t.Errorf("expected 10 attacks, got %d", len(attacks))
	}

	// Test partial last page
	attacks, err = storage.GetAttacks(10, 20)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(attacks) != 5 {
		t.Errorf("expected 5 attacks, got %d", len(attacks))
	}

	// Test offset beyond data
	attacks, err = storage.GetAttacks(10, 100)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(attacks) != 0 {
		t.Errorf("expected 0 attacks, got %d", len(attacks))
	}
}

func TestMemoryStorage_Clear(t *testing.T) {
	storage := NewMemoryStorage()
	defer storage.Close()

	// Add some data
	attack := &models.AttackAttempt{IP: "192.168.1.1", Service: "ssh"}
	storage.SaveAttack(attack)

	block := &models.BlockRecord{IP: "192.168.1.1", Service: "ssh", IsActive: true}
	storage.SaveBlock(block)

	// Verify data exists
	if storage.GetAttackCount() != 1 {
		t.Error("expected 1 attack before clear")
	}
	if storage.GetBlockCount() != 1 {
		t.Error("expected 1 block before clear")
	}

	// Clear and verify
	storage.Clear()

	if storage.GetAttackCount() != 0 {
		t.Error("expected 0 attacks after clear")
	}
	if storage.GetBlockCount() != 0 {
		t.Error("expected 0 blocks after clear")
	}
}