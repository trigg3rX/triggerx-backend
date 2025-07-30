package sync

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	redisclient "github.com/redis/go-redis/v9"
	"github.com/trigg3rX/triggerx-backend/pkg/client/redis"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

const (
	// Checkpoint keys
	KeyCheckpointPrefix = "registrar:checkpoint:"
	KeyLatestCheckpoint = "registrar:checkpoint:latest"

	// Default checkpoint interval
	DefaultCheckpointInterval = 10 * time.Minute
)

// CheckpointManager handles periodic state snapshots
type CheckpointManager struct {
	redis    *redis.Client
	logger   logging.Logger
	interval time.Duration
}

// Checkpoint represents a state snapshot
type Checkpoint struct {
	ID                  string                 `json:"id"`
	Timestamp           time.Time              `json:"timestamp"`
	LastPolledEthBlock  uint64                 `json:"last_polled_eth_block"`
	LastPolledBaseBlock uint64                 `json:"last_polled_base_block"`
	// LastPolledOptBlock  uint64                 `json:"last_polled_opt_block"`
	LastRewardsUpdate   time.Time              `json:"last_rewards_update"`
	Version             string                 `json:"version"`
	ServiceInfo         map[string]interface{} `json:"service_info"`
}

// CheckpointSummary represents a lightweight checkpoint summary for listing
type CheckpointSummary struct {
	ID        string    `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	BlockInfo string    `json:"block_info"`
}

// NewCheckpointManager creates a new checkpoint manager
func NewCheckpointManager(redis *redis.Client, logger logging.Logger) *CheckpointManager {
	return &CheckpointManager{
		redis:    redis,
		logger:   logger,
		interval: DefaultCheckpointInterval,
	}
}

// SetInterval sets the checkpoint interval
func (cm *CheckpointManager) SetInterval(interval time.Duration) {
	cm.interval = interval
}

// CreateCheckpoint creates a new checkpoint
func (cm *CheckpointManager) CreateCheckpoint(ctx context.Context, stateManager *StateManager, serviceInfo map[string]interface{}) error {
	// Get current state
	state, err := stateManager.GetFullState(ctx)
	if err != nil {
		return fmt.Errorf("failed to get current state: %w", err)
	}

	// Create checkpoint
	checkpointID := fmt.Sprintf("checkpoint_%d", time.Now().Unix())
	checkpoint := &Checkpoint{
		ID:                  checkpointID,
		Timestamp:           time.Now().UTC(),
		LastPolledEthBlock:  state.LastEthBlockUpdated,
		LastPolledBaseBlock: state.LastBaseBlockUpdated,
		// LastPolledOptBlock:  state.LastOptBlockUpdated,
		LastRewardsUpdate:   state.LastRewardsUpdate,
		Version:             "1.0",
		ServiceInfo:         serviceInfo,
	}

	// Serialize checkpoint
	checkpointData, err := json.Marshal(checkpoint)
	if err != nil {
		return fmt.Errorf("failed to serialize checkpoint: %w", err)
	}

	// Save checkpoint with 7-day expiration
	checkpointKey := KeyCheckpointPrefix + checkpointID
	if err := cm.redis.Set(ctx, checkpointKey, string(checkpointData), 7*24*time.Hour); err != nil {
		return fmt.Errorf("failed to save checkpoint: %w", err)
	}

	// Update latest checkpoint pointer
	if err := cm.redis.Set(ctx, KeyLatestCheckpoint, checkpointID, 0); err != nil {
		return fmt.Errorf("failed to update latest checkpoint: %w", err)
	}

	// Add to sorted set index for efficient listing
	timestamp := float64(checkpoint.Timestamp.Unix())
	if _, err := cm.redis.ZAdd(ctx, "registrar:checkpoints:index", redisclient.Z{Score: timestamp, Member: checkpointID}); err != nil {
		return fmt.Errorf("failed to add checkpoint to index: %w", err)
	}

	cm.logger.Infof("Created checkpoint %s at blocks ETH:%d, BASE:%d",
		checkpointID, state.LastEthBlockUpdated, state.LastBaseBlockUpdated)

	return nil
}

// GetLatestCheckpoint retrieves the latest checkpoint
func (cm *CheckpointManager) GetLatestCheckpoint(ctx context.Context) (*Checkpoint, error) {
	// Get latest checkpoint ID
	latestID, err := cm.redis.Get(ctx, KeyLatestCheckpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest checkpoint ID: %w", err)
	}

	if latestID == "" {
		return nil, fmt.Errorf("no checkpoints found")
	}

	return cm.GetCheckpoint(ctx, latestID)
}

// GetCheckpoint retrieves a specific checkpoint
func (cm *CheckpointManager) GetCheckpoint(ctx context.Context, checkpointID string) (*Checkpoint, error) {
	checkpointKey := KeyCheckpointPrefix + checkpointID
	checkpointData, err := cm.redis.Get(ctx, checkpointKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get checkpoint %s: %w", checkpointID, err)
	}

	if checkpointData == "" {
		return nil, fmt.Errorf("checkpoint %s not found", checkpointID)
	}

	var checkpoint Checkpoint
	if err := json.Unmarshal([]byte(checkpointData), &checkpoint); err != nil {
		return nil, fmt.Errorf("failed to deserialize checkpoint: %w", err)
	}

	return &checkpoint, nil
}

// RestoreFromCheckpoint restores state from a checkpoint
func (cm *CheckpointManager) RestoreFromCheckpoint(ctx context.Context, stateManager *StateManager, checkpointID string) error {
	checkpoint, err := cm.GetCheckpoint(ctx, checkpointID)
	if err != nil {
		return fmt.Errorf("failed to get checkpoint: %w", err)
	}

	// Restore state
	if err := stateManager.SetLastEthBlockUpdated(ctx, checkpoint.LastPolledEthBlock); err != nil {
		return fmt.Errorf("failed to restore ETH block: %w", err)
	}

	if err := stateManager.SetLastBaseBlockUpdated(ctx, checkpoint.LastPolledBaseBlock); err != nil {
		return fmt.Errorf("failed to restore BASE block: %w", err)
	}

	// if err := stateManager.SetLastOptBlockUpdated(ctx, checkpoint.LastPolledOptBlock); err != nil {
	// 	return fmt.Errorf("failed to restore OPT block: %w", err)
	// }

	if !checkpoint.LastRewardsUpdate.IsZero() {
		if err := stateManager.SetLastRewardsUpdate(ctx, checkpoint.LastRewardsUpdate); err != nil {
			return fmt.Errorf("failed to restore rewards update time: %w", err)
		}
	}

	cm.logger.Infof("Restored state from checkpoint %s (timestamp: %s)",
		checkpoint.ID, checkpoint.Timestamp.Format(time.RFC3339))

	return nil
}

// StartPeriodicCheckpoints starts creating periodic checkpoints
func (cm *CheckpointManager) StartPeriodicCheckpoints(ctx context.Context, stateManager *StateManager) {
	ticker := time.NewTicker(cm.interval)
	defer ticker.Stop()

	cm.logger.Infof("Starting periodic checkpoints every %v", cm.interval)

	for {
		select {
		case <-ctx.Done():
			cm.logger.Info("Stopping periodic checkpoints")
			return
		case <-ticker.C:
			serviceInfo := map[string]interface{}{
				"checkpoint_interval": cm.interval.String(),
				"created_by":          "periodic_checkpoint",
			}

			if err := cm.CreateCheckpoint(ctx, stateManager, serviceInfo); err != nil {
				cm.logger.Errorf("Failed to create periodic checkpoint: %v", err)
			}
		}
	}
}

// ListCheckpoints lists all available checkpoints
func (cm *CheckpointManager) ListCheckpoints(ctx context.Context, limit int, offset int) ([]CheckpointSummary, error) {
	// Use Redis Sorted Set for efficient listing
	checkpointIDs, err := cm.redis.ZRevRange(ctx, "registrar:checkpoints:index", int64(offset), int64(offset+limit-1))
	if err != nil {
		return nil, fmt.Errorf("failed to get checkpoint index: %w", err)
	}

	// Fetch metadata for each checkpoint
	var summaries []CheckpointSummary
	for _, id := range checkpointIDs {
		// Get checkpoint metadata (could be cached)
		checkpoint, err := cm.GetCheckpoint(ctx, id)
		if err != nil {
			continue // Skip corrupted checkpoints
		}
		summaries = append(summaries, CheckpointSummary{
			ID:        checkpoint.ID,
			Timestamp: checkpoint.Timestamp,
			BlockInfo: fmt.Sprintf("ETH:%d, BASE:%d",
				checkpoint.LastPolledEthBlock,
				checkpoint.LastPolledBaseBlock),
		})
	}

	return summaries, nil
}

// CleanupOldCheckpoints removes checkpoints older than the specified duration
func (cm *CheckpointManager) CleanupOldCheckpoints(ctx context.Context, maxAge time.Duration) error {
	cm.logger.Infof("Cleaning up checkpoints older than %v", maxAge)

	// Calculate the cutoff timestamp
	cutoffTime := time.Now().Add(-maxAge)
	cutoffScore := float64(cutoffTime.Unix())

	// Remove old checkpoints from the sorted set index
	removedCount, err := cm.redis.ZRemRangeByScore(ctx, "registrar:checkpoints:index", "0", fmt.Sprintf("%.0f", cutoffScore))
	if err != nil {
		return fmt.Errorf("failed to remove old checkpoints from index: %w", err)
	}

	if removedCount > 0 {
		cm.logger.Infof("Removed %d old checkpoints from index", removedCount)
	}

	// Also clean up the actual checkpoint keys that might still exist
	// (in case they weren't automatically expired by TTL)
	pattern := KeyCheckpointPrefix + "*"
	keys, err := cm.redis.Keys(ctx, pattern)
	if err != nil {
		return fmt.Errorf("failed to scan checkpoint keys: %w", err)
	}

	var deletedKeys []string
	for _, key := range keys {
		// Extract checkpoint ID from key
		checkpointID := key[len(KeyCheckpointPrefix):]

		// Get checkpoint to check its timestamp
		checkpoint, err := cm.GetCheckpoint(ctx, checkpointID)
		if err != nil {
			// If we can't get the checkpoint, it might be corrupted, so delete the key
			if err := cm.redis.Del(ctx, key); err != nil {
				cm.logger.Warnf("Failed to delete corrupted checkpoint key %s: %v", key, err)
			} else {
				deletedKeys = append(deletedKeys, key)
			}
			continue
		}

		// Delete checkpoint if it's older than maxAge
		if checkpoint.Timestamp.Before(cutoffTime) {
			if err := cm.redis.Del(ctx, key); err != nil {
				cm.logger.Warnf("Failed to delete old checkpoint key %s: %v", key, err)
			} else {
				deletedKeys = append(deletedKeys, key)
			}
		}
	}

	if len(deletedKeys) > 0 {
		cm.logger.Infof("Deleted %d old checkpoint keys", len(deletedKeys))
	}

	return nil
}

// GetCheckpointCount returns the total number of checkpoints
func (cm *CheckpointManager) GetCheckpointCount(ctx context.Context) (int64, error) {
	return cm.redis.ZCard(ctx, "registrar:checkpoints:index")
}

// GetCheckpointHealth returns health information about checkpoints
func (cm *CheckpointManager) GetCheckpointHealth(ctx context.Context) map[string]interface{} {
	health := map[string]interface{}{
		"checkpoint_interval": cm.interval.String(),
		"latest_checkpoint":   nil,
		"checkpoint_count":    0,
	}

	// Get latest checkpoint info
	latest, err := cm.GetLatestCheckpoint(ctx)
	if err == nil {
		health["latest_checkpoint"] = map[string]interface{}{
			"id":        latest.ID,
			"timestamp": latest.Timestamp.Format(time.RFC3339),
			"age":       time.Since(latest.Timestamp).String(),
		}
	}

	// Get checkpoint count
	count, err := cm.GetCheckpointCount(ctx)
	if err == nil {
		health["checkpoint_count"] = count
	}

	return health
}
