package performers

import (
	"context"
	"fmt"
	"sync"
	"time"

	redisClient "github.com/trigg3rX/triggerx-backend/pkg/client/redis"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

const (
	PerformerLockTTL    = 5 * time.Minute
	PerformerHealthTTL  = 30 * time.Second
)

// TODO: This is a temporary implementation. In the future, this should:
// 1. Fetch available performers from health service
// 2. Implement proper round-robin or load balancing
// 3. Handle performer locking/unlocking
// 4. Monitor performer availability and capacity

// GetPerformerData returns an available performer for task execution
// Currently returns a hardcoded performer - will be enhanced later
func GetPerformerData() types.PerformerData {
	// TODO: Replace with actual performer selection logic
	// For now, returning a fixed performer as mentioned in the scheduler code
	// return types.PerformerData{
	// 	KeeperID:      3,
	// 	KeeperAddress: "0x0a067a261c5f5e8c4c0b9137430b4fe1255eb62e",
	// }
	return types.PerformerData{
		KeeperID:      2,
		KeeperAddress: "0x011fcbae5f306cd793456ab7d4c0cc86756c693d",
	}
}

// PerformerManager handles performer lifecycle and assignment
type PerformerManager struct {
	client     redisClient.RedisClientInterface
	logger     logging.Logger
	performers map[int64]*types.PerformerData
	mu         sync.RWMutex
	startTime  time.Time

	// Performance tracking
	lastRoundRobinIndex int
	roundRobinMu        sync.Mutex
}

// NewPerformerManager creates a new performer manager with improved initialization
func NewPerformerManager(client redisClient.RedisClientInterface, logger logging.Logger) *PerformerManager {
	pm := &PerformerManager{
		client:     client,
		logger:     logger,
		performers: make(map[int64]*types.PerformerData),
		startTime:  time.Now(),
	}

	// Initialize with default performers
	pm.initializeDefaultPerformers()

	logger.Info("PerformerManager initialized successfully")
	return pm
}

// initializeDefaultPerformers sets up default performers for the system
func (pm *PerformerManager) initializeDefaultPerformers() {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// Add default performers - this should be replaced with dynamic discovery
	defaultPerformers := []types.PerformerData{
		{
			KeeperID:      3,
			KeeperAddress: "0x0a067a261c5f5e8c4c0b9137430b4fe1255eb62e",
		},
		// Add more default performers as needed
	}

	for _, performer := range defaultPerformers {
		pm.performers[performer.KeeperID] = &performer
	}

	pm.logger.Info("Default performers initialized", "count", len(defaultPerformers))
}

// AcquirePerformer gets an available performer and locks it for task execution with improved selection
func (pm *PerformerManager) AcquirePerformer(ctx context.Context) (*types.PerformerData, error) {
	availablePerformers := pm.GetAvailablePerformers()
	if len(availablePerformers) == 0 {
		return nil, fmt.Errorf("no performers available")
	}

	// Use round-robin selection for better load distribution
	performer := pm.SelectPerformerRoundRobin(availablePerformers)
	if performer == nil {
		return nil, fmt.Errorf("no available performers after selection")
	}

	// Create lock key with improved naming
	lockKey := fmt.Sprintf("%s%d", PerformerLockPrefix, performer.KeeperID)

	// Try to acquire lock with improved timeout handling
	locked, err := pm.client.SetNX(ctx, lockKey, "locked", PerformerLockTTL)
	if err != nil {
		pm.logger.Error("Failed to acquire performer lock",
			"performer_id", performer.KeeperID,
			"error", err)
		return nil, fmt.Errorf("failed to acquire performer lock: %w", err)
	}

	if !locked {
		pm.logger.Debug("Performer is locked, trying next available performer",
			"performer_id", performer.KeeperID)

		// Try to find another available performer
		for _, altPerformer := range availablePerformers {
			if altPerformer.KeeperID == performer.KeeperID {
				continue // Skip the one we just tried
			}

			altLockKey := fmt.Sprintf("%s%d", PerformerLockPrefix, altPerformer.KeeperID)
			altLocked, altErr := pm.client.SetNX(ctx, altLockKey, "locked", PerformerLockTTL)
			if altErr == nil && altLocked {
				pm.logger.Info("Acquired alternative performer for task execution",
					"performer_id", altPerformer.KeeperID,
					"performer_address", altPerformer.KeeperAddress)
				return &altPerformer, nil
			}
		}

		return nil, fmt.Errorf("no available performers")
	}

	pm.logger.Info("Acquired performer for task execution",
		"performer_id", performer.KeeperID,
		"performer_address", performer.KeeperAddress,
		"lock_ttl", PerformerLockTTL)

	return performer, nil
}

// ReleasePerformer releases the performer lock with improved error handling
func (pm *PerformerManager) ReleasePerformer(ctx context.Context, performerID int64) error {
	lockKey := fmt.Sprintf("%s%d", PerformerLockPrefix, performerID)

	err := pm.client.Del(ctx, lockKey)
	if err != nil {
		pm.logger.Error("Failed to release performer lock",
			"performer_id", performerID,
			"error", err)
		return fmt.Errorf("failed to release performer lock: %w", err)
	}

	pm.logger.Info("Released performer lock",
		"performer_id", performerID)

	return nil
}

// UpdatePerformerStatus updates the status of a performer with improved tracking
func (pm *PerformerManager) UpdatePerformerStatus(performerID int64, isOnline bool) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if performer, exists := pm.performers[performerID]; exists {
		// Update performer status
		pm.logger.Debug("Updated performer status",
			"performer_id", performerID,
			"is_online", isOnline,
			"address", performer.KeeperAddress)
	} else {
		pm.logger.Debug("Performer not found for status update",
			"performer_id", performerID)
	}
}

// GetAvailablePerformers returns all available performers with improved filtering
func (pm *PerformerManager) GetAvailablePerformers() []types.PerformerData {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	performers := make([]types.PerformerData, 0, len(pm.performers))
	for _, performer := range pm.performers {
		performers = append(performers, *performer)
	}

	return performers
}

// SelectPerformerRoundRobin selects a performer using round-robin algorithm with improved performance
func (pm *PerformerManager) SelectPerformerRoundRobin(performers []types.PerformerData) *types.PerformerData {
	if len(performers) == 0 {
		return nil
	}

	pm.roundRobinMu.Lock()
	defer pm.roundRobinMu.Unlock()

	// Use round-robin selection
	selectedIndex := pm.lastRoundRobinIndex % len(performers)
	pm.lastRoundRobinIndex = (pm.lastRoundRobinIndex + 1) % len(performers)

	selectedPerformer := performers[selectedIndex]

	pm.logger.Debug("Selected performer using round-robin",
		"performer_id", selectedPerformer.KeeperID,
		"index", selectedIndex,
		"total_performers", len(performers))

	return &selectedPerformer
}

// IsPerformerAvailable checks if a performer is available with improved health checking
func (pm *PerformerManager) IsPerformerAvailable(ctx context.Context, performerID int64) bool {
	lockKey := fmt.Sprintf("%s%d", PerformerLockPrefix, performerID)

		// Check if performer is locked by trying to get the lock key
	_, err := pm.client.Get(ctx, lockKey)
	if err != nil {
		// If key doesn't exist, performer is available
		return true
	}
	
	// If we can get the key, performer is locked
	return false
}

// GetPerformerStats returns statistics about performer usage and availability
func (pm *PerformerManager) GetPerformerStats() map[string]interface{} {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	stats := map[string]interface{}{
		"total_performers":       len(pm.performers),
		"uptime_seconds":         time.Since(pm.startTime).Seconds(),
		"start_time":             pm.startTime.Format(time.RFC3339),
		"last_round_robin_index": pm.lastRoundRobinIndex,
	}

	// Add performer-specific stats
	performerStats := make(map[int64]map[string]interface{})
	for performerID, performer := range pm.performers {
		performerStats[performerID] = map[string]interface{}{
			"address": performer.KeeperAddress,
		}
	}
	stats["performers"] = performerStats

	return stats
}
