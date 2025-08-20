package storage

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/sr-tamim/guardian/internal/core"
	"github.com/sr-tamim/guardian/pkg/models"
)

// MemoryStorage implements in-memory storage for Guardian data
type MemoryStorage struct {
	mu sync.RWMutex

	// Auto-incrementing IDs
	nextAttackID int64
	nextBlockID  int64

	// Data storage
	attacks     map[int64]*models.AttackAttempt
	blocks      map[string]*models.BlockRecord // keyed by IP
	attacksByIP map[string][]*models.AttackAttempt
}

// NewMemoryStorage creates a new in-memory storage instance
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		nextAttackID: 1,
		nextBlockID:  1,
		attacks:      make(map[int64]*models.AttackAttempt),
		blocks:       make(map[string]*models.BlockRecord),
		attacksByIP:  make(map[string][]*models.AttackAttempt),
	}
}

// SaveAttack saves an attack attempt
func (m *MemoryStorage) SaveAttack(attempt *models.AttackAttempt) error {
	if attempt == nil {
		return core.NewError(core.ErrStorageOperation, "attack attempt cannot be nil", nil)
	}

	if !attempt.IsValidIP() {
		return core.NewError(core.ErrInvalidIP,
			fmt.Sprintf("invalid IP address: %s", attempt.IP), nil)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Assign ID if not set
	if attempt.ID == 0 {
		attempt.ID = m.nextAttackID
		m.nextAttackID++
	}

	// Set timestamp if not set
	if attempt.Timestamp.IsZero() {
		attempt.Timestamp = time.Now()
	}

	// Store attack
	attackCopy := *attempt
	m.attacks[attempt.ID] = &attackCopy

	// Update IP index
	m.attacksByIP[attempt.IP] = append(m.attacksByIP[attempt.IP], &attackCopy)

	return nil
}

// GetAttacks retrieves attack attempts with pagination
func (m *MemoryStorage) GetAttacks(limit int, offset int) ([]*models.AttackAttempt, error) {
	if limit <= 0 {
		limit = 100 // Default limit
	}
	if offset < 0 {
		offset = 0
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	// Collect all attacks and sort by timestamp (newest first)
	attacks := make([]*models.AttackAttempt, 0, len(m.attacks))
	for _, attack := range m.attacks {
		attacks = append(attacks, attack)
	}

	sort.Slice(attacks, func(i, j int) bool {
		return attacks[i].Timestamp.After(attacks[j].Timestamp)
	})

	// Apply pagination
	start := offset
	if start >= len(attacks) {
		return []*models.AttackAttempt{}, nil
	}

	end := start + limit
	if end > len(attacks) {
		end = len(attacks)
	}

	result := make([]*models.AttackAttempt, end-start)
	for i := start; i < end; i++ {
		// Return copies to prevent external modifications
		copy := *attacks[i]
		result[i-start] = &copy
	}

	return result, nil
}

// GetAttacksByIP retrieves attacks for a specific IP since a given time
func (m *MemoryStorage) GetAttacksByIP(ip string, since time.Time) ([]*models.AttackAttempt, error) {
	if ip == "" {
		return nil, core.NewError(core.ErrInvalidIP, "IP address cannot be empty", nil)
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	ipAttacks, exists := m.attacksByIP[ip]
	if !exists {
		return []*models.AttackAttempt{}, nil
	}

	var result []*models.AttackAttempt
	for _, attack := range ipAttacks {
		if attack.Timestamp.After(since) || attack.Timestamp.Equal(since) {
			// Return copy to prevent external modifications
			copy := *attack
			result = append(result, &copy)
		}
	}

	// Sort by timestamp (newest first)
	sort.Slice(result, func(i, j int) bool {
		return result[i].Timestamp.After(result[j].Timestamp)
	})

	return result, nil
}

// SaveBlock saves a block record
func (m *MemoryStorage) SaveBlock(block *models.BlockRecord) error {
	if block == nil {
		return core.NewError(core.ErrStorageOperation, "block record cannot be nil", nil)
	}

	if block.IP == "" {
		return core.NewError(core.ErrInvalidIP, "block IP cannot be empty", nil)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Assign ID if not set
	if block.ID == 0 {
		block.ID = m.nextBlockID
		m.nextBlockID++
	}

	// Set blocked time if not set
	if block.BlockedAt.IsZero() {
		block.BlockedAt = time.Now()
	}

	// Store block
	blockCopy := *block
	m.blocks[block.IP] = &blockCopy

	return nil
}

// GetBlock retrieves a block record by IP
func (m *MemoryStorage) GetBlock(ip string) (*models.BlockRecord, error) {
	if ip == "" {
		return nil, core.NewError(core.ErrInvalidIP, "IP address cannot be empty", nil)
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	block, exists := m.blocks[ip]
	if !exists {
		return nil, core.NewError(core.ErrRecordNotFound,
			fmt.Sprintf("no block record found for IP %s", ip), nil)
	}

	// Return copy to prevent external modifications
	copy := *block
	return &copy, nil
}

// GetActiveBlocks retrieves all active (non-expired) block records
func (m *MemoryStorage) GetActiveBlocks() ([]*models.BlockRecord, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var activeBlocks []*models.BlockRecord
	now := time.Now()

	for _, block := range m.blocks {
		if block.IsActive && !block.IsExpired() {
			// Double-check expiration with current time
			if block.ExpiresAt != nil && now.After(*block.ExpiresAt) {
				continue // Skip expired blocks
			}
			// Return copy to prevent external modifications
			copy := *block
			activeBlocks = append(activeBlocks, &copy)
		}
	}

	// Sort by blocked time (newest first)
	sort.Slice(activeBlocks, func(i, j int) bool {
		return activeBlocks[i].BlockedAt.After(activeBlocks[j].BlockedAt)
	})

	return activeBlocks, nil
}

// UpdateBlock updates an existing block record
func (m *MemoryStorage) UpdateBlock(block *models.BlockRecord) error {
	if block == nil {
		return core.NewError(core.ErrStorageOperation, "block record cannot be nil", nil)
	}

	if block.IP == "" {
		return core.NewError(core.ErrInvalidIP, "block IP cannot be empty", nil)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	_, exists := m.blocks[block.IP]
	if !exists {
		return core.NewError(core.ErrRecordNotFound,
			fmt.Sprintf("no block record found for IP %s", block.IP), nil)
	}

	// Update block
	blockCopy := *block
	m.blocks[block.IP] = &blockCopy

	return nil
}

// GetStatistics returns storage statistics
func (m *MemoryStorage) GetStatistics() (*models.Statistics, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	totalAttacks := int64(len(m.attacks))
	totalBlocks := int64(len(m.blocks))

	// Count active blocks
	var activeBlocks int64
	now := time.Now()
	for _, block := range m.blocks {
		if block.IsActive && !block.IsExpired() {
			// Double-check expiration
			if block.ExpiresAt == nil || now.Before(*block.ExpiresAt) {
				activeBlocks++
			}
		}
	}

	// Count monitored services (simplified for memory storage)
	servicesMonitored := 1 // Default assumption

	// Find last activity
	var lastActivity time.Time
	for _, attack := range m.attacks {
		if attack.Timestamp.After(lastActivity) {
			lastActivity = attack.Timestamp
		}
	}

	return &models.Statistics{
		TotalAttacks:      totalAttacks,
		BlockedIPs:        totalBlocks,
		ActiveBlocks:      activeBlocks,
		ServicesMonitored: servicesMonitored,
		UptimeSeconds:     0, // Would need to track separately
		LastActivity:      lastActivity,
	}, nil
}

// Close closes the storage (no-op for memory storage)
func (m *MemoryStorage) Close() error {
	return nil
}

// Clear removes all stored data (for testing)
func (m *MemoryStorage) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.nextAttackID = 1
	m.nextBlockID = 1
	m.attacks = make(map[int64]*models.AttackAttempt)
	m.blocks = make(map[string]*models.BlockRecord)
	m.attacksByIP = make(map[string][]*models.AttackAttempt)
}

// GetAttackCount returns the total number of stored attacks
func (m *MemoryStorage) GetAttackCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.attacks)
}

// GetBlockCount returns the total number of stored blocks
func (m *MemoryStorage) GetBlockCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.blocks)
}
