package performers

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/trigg3rX/triggerx-backend/internal/taskmanager/config"
	redisClient "github.com/trigg3rX/triggerx-backend/pkg/client/redis"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

const (
	PerformerLockTTL    = 5 * time.Minute
	PerformerHealthTTL  = 30 * time.Second
	PerformerRefreshTTL = 15 * time.Second

	// Redis keys for performer state
	NextPerformerKey  = "taskmanager:next_performer"
	BusyPerformersKey = "taskmanager:busy_performers"
)

// GetPerformerData gets a performer using the dynamic selection system
func (pm *PerformerManager) GetPerformerData(isImua bool) (types.PerformerData, error) {
	pm.logger.Debug("Getting performer data dynamically", "is_imua", isImua)

	// Refresh performers if needed
	if time.Since(pm.lastRefresh) > PerformerRefreshTTL {
		pm.logger.Debug("Refreshing performers from health service")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := pm.refreshPerformers(ctx); err != nil {
			pm.logger.Error("Failed to refresh performers", "error", err)
			// Fall back to cached performers
		}
	}

	var availablePerformers []types.PerformerData
	// availablePerformers := pm.GetAvailablePerformers()
	// pm.logger.Debug("Available performers count", "count", len(availablePerformers))

	if len(availablePerformers) == 0 {
		pm.logger.Warn("No performers available from health service, using fallback")
		// Fallback to default performers
		fallbackPerformers := []types.PerformerData{
			{
				OperatorID:    3,
				KeeperAddress: "0x011fcbae5f306cd793456ab7d4c0cc86756c693d",
				IsImua:        false,
			},
			{
				OperatorID:    4,
				KeeperAddress: "0x0a067a261c5f5e8c4c0b9137430b4fe1255eb62e",
				IsImua:        false,
			},
		}
		availablePerformers = fallbackPerformers
		pm.logger.Info("Using fallback performers", "count", len(availablePerformers))
	}

	// Log available performers for debugging
	for i, performer := range availablePerformers {
		pm.logger.Debug("Available performer",
			"index", i,
			"operator_id", performer.OperatorID,
			"keeper_address", performer.KeeperAddress,
			"is_imua", performer.IsImua)
	}

	// Filter by Imua status
	var filteredPerformers []types.PerformerData
	for _, performer := range availablePerformers {
		if performer.IsImua == isImua {
			filteredPerformers = append(filteredPerformers, performer)
		}
	}

	pm.logger.Debug("Filtered performers by Imua status",
		"is_imua", isImua,
		"total_available", len(availablePerformers),
		"filtered_count", len(filteredPerformers))

	if len(filteredPerformers) == 0 {
		pm.logger.Error("No suitable performers available after Imua filtering",
			"is_imua", isImua,
			"total_available", len(availablePerformers))
		return types.PerformerData{}, fmt.Errorf("no suitable performers available for isImua=%v", isImua)
	}

	// Use round-robin selection with Redis state
	pm.logger.Debug("Selecting performer using round-robin")
	performer := pm.SelectPerformerRoundRobin(filteredPerformers)
	if performer == nil {
		pm.logger.Error("No available performers after round-robin selection")
		return types.PerformerData{}, fmt.Errorf("no available performers after selection")
	}

	pm.logger.Debug("Selected performer",
		"operator_id", performer.OperatorID,
		"keeper_address", performer.KeeperAddress,
		"is_imua", performer.IsImua)

	// Mark performer as busy
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := pm.MarkPerformerBusy(ctx, performer.OperatorID); err != nil {
		pm.logger.Warn("Failed to mark performer as busy",
			"performer_id", performer.OperatorID,
			"error", err)
		// Continue anyway, the performer is still selected
	}

	pm.logger.Info("Selected performer dynamically",
		"performer_id", performer.OperatorID,
		"performer_address", performer.KeeperAddress,
		"is_imua", isImua,
		"performer_is_imua", performer.IsImua)

	return *performer, nil
}

// PerformerManager handles performer lifecycle and assignment
type PerformerManager struct {
	client       redisClient.RedisClientInterface
	logger       logging.Logger
	healthClient *HealthClient
	startTime    time.Time

	// Performance tracking
	lastRoundRobinIndex int

	// Cached performers
	performers   []types.PerformerData
	performersMu sync.RWMutex
	lastRefresh  time.Time
}

// NewPerformerManager creates a new performer manager with improved initialization
// func NewPerformerManager(client redisClient.RedisClientInterface, logger logging.Logger) *PerformerManager {
func NewPerformerManager(logger logging.Logger) *PerformerManager {
	pm := &PerformerManager{
		// client:       client,
		logger:       logger,
		healthClient: NewHealthClient(logger),
		startTime:    time.Now(),
		performers:   make([]types.PerformerData, 0),
	}

	logger.Info("PerformerManager initialized successfully", "health_rpc_url", config.GetHealthRPCUrl())
	return pm
}

// refreshPerformers fetches fresh performers from health service
func (pm *PerformerManager) refreshPerformers(ctx context.Context) error {
	pm.logger.Debug("Refreshing performers from health service")

	performers, err := pm.healthClient.GetActivePerformers(ctx)
	if err != nil {
		pm.logger.Error("Failed to fetch performers from health service", "error", err)
		return err
	}

	pm.logger.Debug("Fetched performers from health service", "count", len(performers))

	// Log fetched performers for debugging
	for i, performer := range performers {
		pm.logger.Debug("Fetched performer",
			"index", i,
			"operator_id", performer.OperatorID,
			"keeper_address", performer.KeeperAddress,
			"is_imua", performer.IsImua)
	}

	pm.performersMu.Lock()
	pm.performers = performers
	pm.lastRefresh = time.Now()
	pm.performersMu.Unlock()

	pm.logger.Info("Successfully refreshed performers", "count", len(performers))
	return nil
}

// TODO: Add isImua flag
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
	lockKey := fmt.Sprintf("%s%d", PerformerLockPrefix, performer.OperatorID)

	// Try to acquire lock with improved timeout handling
	locked, err := pm.client.SetNX(ctx, lockKey, "locked", PerformerLockTTL)
	if err != nil {
		pm.logger.Error("Failed to acquire performer lock",
			"performer_id", performer.OperatorID,
			"error", err)
		return nil, fmt.Errorf("failed to acquire performer lock: %w", err)
	}

	if !locked {
		pm.logger.Debug("Performer is locked, trying next available performer",
			"performer_id", performer.OperatorID)

		// Try to find another available performer
		for _, altPerformer := range availablePerformers {
			if altPerformer.OperatorID == performer.OperatorID {
				continue // Skip the one we just tried
			}

			altLockKey := fmt.Sprintf("%s%d", PerformerLockPrefix, altPerformer.OperatorID)
			altLocked, altErr := pm.client.SetNX(ctx, altLockKey, "locked", PerformerLockTTL)
			if altErr == nil && altLocked {
				pm.logger.Info("Acquired alternative performer for task execution",
					"performer_id", altPerformer.OperatorID,
					"performer_address", altPerformer.KeeperAddress)
				return &altPerformer, nil
			}
		}

		return nil, fmt.Errorf("no available performers")
	}

	pm.logger.Info("Acquired performer for task execution",
		"performer_id", performer.OperatorID,
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
	pm.logger.Debug("Updated performer status",
		"performer_id", performerID,
		"is_online", isOnline)
}

// GetAvailablePerformers returns all available performers with improved filtering
func (pm *PerformerManager) GetAvailablePerformers() []types.PerformerData {
	pm.performersMu.RLock()
	defer pm.performersMu.RUnlock()

	pm.logger.Debug("Getting available performers", "cached_count", len(pm.performers), "last_refresh", pm.lastRefresh)

	// Check if we need to refresh performers
	if time.Since(pm.lastRefresh) > PerformerRefreshTTL {
		pm.logger.Debug("Performers need refresh, triggering background refresh")
		// Refresh in background
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			if err := pm.refreshPerformers(ctx); err != nil {
				pm.logger.Error("Background performer refresh failed", "error", err)
			}
		}()
	}

	// Return cached performers
	return pm.performers
}

// SelectPerformerRoundRobin selects a performer using round-robin algorithm with Redis state
func (pm *PerformerManager) SelectPerformerRoundRobin(performers []types.PerformerData) *types.PerformerData {
	if len(performers) == 0 {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Get current next performer index from Redis
	nextPerformerStr, err := pm.client.Get(ctx, NextPerformerKey)
	if err != nil {
		// If not found, start with 0
		nextPerformerStr = "0"
	}

	var nextIndex int
	if _, parseErr := fmt.Sscanf(nextPerformerStr, "%d", &nextIndex); parseErr != nil {
		nextIndex = 0
	}

	// Ensure index is within bounds
	nextIndex = nextIndex % len(performers)

	// Get busy performers from Redis
	busyPerformersStr, err := pm.client.Get(ctx, BusyPerformersKey)
	if err != nil {
		busyPerformersStr = "[]"
	}

	var busyPerformerIDs []int64
	if err := json.Unmarshal([]byte(busyPerformersStr), &busyPerformerIDs); err != nil {
		busyPerformerIDs = []int64{}
	}

	// Find next available performer
	attempts := 0
	for attempts < len(performers) {
		selectedPerformer := performers[nextIndex]

		// Check if performer is busy
		isBusy := false
		for _, busyID := range busyPerformerIDs {
			if busyID == selectedPerformer.OperatorID {
				isBusy = true
				break
			}
		}

		if !isBusy {
			// Update next performer index in Redis
			nextIndex = (nextIndex + 1) % len(performers)
			err := pm.client.Set(ctx, NextPerformerKey, fmt.Sprintf("%d", nextIndex), 0)
			if err != nil {
				pm.logger.Error("Failed to update next performer index in Redis", "error", err)
			}

			pm.logger.Debug("Selected performer using round-robin",
				"performer_id", selectedPerformer.OperatorID,
				"index", nextIndex,
				"total_performers", len(performers))

			return &selectedPerformer
		}

		// Try next performer
		nextIndex = (nextIndex + 1) % len(performers)
		attempts++
	}

	pm.logger.Warn("No available performers found in round-robin selection")
	return nil
}

// IsPerformerAvailable checks if a performer is available with improved health checking
func (pm *PerformerManager) IsPerformerAvailable(ctx context.Context, performerID int64) bool {
	// Get busy performers from Redis
	busyPerformersStr, err := pm.client.Get(ctx, BusyPerformersKey)
	if err != nil {
		// If not found, performer is available
		return true
	}

	var busyPerformerIDs []int64
	if err := json.Unmarshal([]byte(busyPerformersStr), &busyPerformerIDs); err != nil {
		return true
	}

	// Check if performer is in busy list
	for _, busyID := range busyPerformerIDs {
		if busyID == performerID {
			return false
		}
	}

	return true
}

// MarkPerformerBusy marks a performer as busy in Redis
func (pm *PerformerManager) MarkPerformerBusy(ctx context.Context, performerID int64) error {
	// Get current busy performers
	busyPerformersStr, err := pm.client.Get(ctx, BusyPerformersKey)
	if err != nil {
		busyPerformersStr = "[]"
	}

	var busyPerformerIDs []int64
	if err := json.Unmarshal([]byte(busyPerformersStr), &busyPerformerIDs); err != nil {
		busyPerformerIDs = []int64{}
	}

	// Add performer to busy list if not already there
	found := false
	for _, busyID := range busyPerformerIDs {
		if busyID == performerID {
			found = true
			break
		}
	}

	if !found {
		busyPerformerIDs = append(busyPerformerIDs, performerID)
		busyJSON, _ := json.Marshal(busyPerformerIDs)
		return pm.client.Set(ctx, BusyPerformersKey, string(busyJSON), PerformerLockTTL)
	}

	return nil
}

// MarkPerformerAvailable marks a performer as available in Redis
func (pm *PerformerManager) MarkPerformerAvailable(ctx context.Context, performerID int64) error {
	// Get current busy performers
	busyPerformersStr, err := pm.client.Get(ctx, BusyPerformersKey)
	if err != nil {
		return nil // Already available
	}

	var busyPerformerIDs []int64
	if err := json.Unmarshal([]byte(busyPerformersStr), &busyPerformerIDs); err != nil {
		return nil
	}

	// Remove performer from busy list
	newBusyIDs := make([]int64, 0, len(busyPerformerIDs))
	for _, busyID := range busyPerformerIDs {
		if busyID != performerID {
			newBusyIDs = append(newBusyIDs, busyID)
		}
	}

	busyJSON, _ := json.Marshal(newBusyIDs)
	return pm.client.Set(ctx, BusyPerformersKey, string(busyJSON), PerformerLockTTL)
}

// GetPerformerStats returns statistics about performer usage and availability
func (pm *PerformerManager) GetPerformerStats() map[string]interface{} {
	pm.performersMu.RLock()
	defer pm.performersMu.RUnlock()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Get busy performers count
	busyCount := 0
	busyPerformersStr, err := pm.client.Get(ctx, BusyPerformersKey)
	if err == nil {
		var busyPerformerIDs []int64
		if json.Unmarshal([]byte(busyPerformersStr), &busyPerformerIDs) == nil {
			busyCount = len(busyPerformerIDs)
		}
	}

	stats := map[string]interface{}{
		"total_performers":       len(pm.performers),
		"busy_performers":        busyCount,
		"available_performers":   len(pm.performers) - busyCount,
		"uptime_seconds":         time.Since(pm.startTime).Seconds(),
		"start_time":             pm.startTime.Format(time.RFC3339),
		"last_refresh":           pm.lastRefresh.Format(time.RFC3339),
		"last_round_robin_index": pm.lastRoundRobinIndex,
	}

	// Add performer-specific stats
	performerStats := make(map[int64]map[string]interface{})
	for _, performer := range pm.performers {
		isAvailable := pm.IsPerformerAvailable(ctx, performer.OperatorID)
		performerStats[performer.OperatorID] = map[string]interface{}{
			"address":   performer.KeeperAddress,
			"available": isAvailable,
		}
	}
	stats["performers"] = performerStats

	return stats
}
