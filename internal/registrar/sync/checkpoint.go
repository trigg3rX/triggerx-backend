package sync

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	redisClient "github.com/trigg3rX/triggerx-backend-imua/internal/registrar/clients/redis"
	"github.com/trigg3rX/triggerx-backend-imua/pkg/logging"
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
	redis    *redisClient.Client
	logger   logging.Logger
	interval time.Duration
}

// Checkpoint represents a state snapshot
type Checkpoint struct {
	ID                  string                 `json:"id"`
	Timestamp           time.Time              `json:"timestamp"`
	LastPolledEthBlock  uint64                 `json:"last_polled_eth_block"`
	LastPolledBaseBlock uint64                 `json:"last_polled_base_block"`
	LastPolledOptBlock  uint64                 `json:"last_polled_opt_block"`
	LastRewardsUpdate   time.Time              `json:"last_rewards_update"`
	Version             string                 `json:"version"`
	ServiceInfo         map[string]interface{} `json:"service_info"`
}

// NewCheckpointManager creates a new checkpoint manager
func NewCheckpointManager(redis *redisClient.Client, logger logging.Logger) *CheckpointManager {
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
		LastPolledEthBlock:  state.LastPolledEthBlock,
		LastPolledBaseBlock: state.LastPolledBaseBlock,
		LastPolledOptBlock:  state.LastPolledOptBlock,
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

	cm.logger.Infof("Created checkpoint %s at blocks ETH:%d, BASE:%d, OPT:%d",
		checkpointID, state.LastPolledEthBlock, state.LastPolledBaseBlock, state.LastPolledOptBlock)

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
	if err := stateManager.SetLastPolledEthBlock(ctx, checkpoint.LastPolledEthBlock); err != nil {
		return fmt.Errorf("failed to restore ETH block: %w", err)
	}

	if err := stateManager.SetLastPolledBaseBlock(ctx, checkpoint.LastPolledBaseBlock); err != nil {
		return fmt.Errorf("failed to restore BASE block: %w", err)
	}

	if err := stateManager.SetLastPolledOptBlock(ctx, checkpoint.LastPolledOptBlock); err != nil {
		return fmt.Errorf("failed to restore OPT block: %w", err)
	}

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
func (cm *CheckpointManager) ListCheckpoints(ctx context.Context) ([]string, error) {
	// This would need a more sophisticated implementation in production
	// For now, we'll use a simple pattern match
	// In Redis, you'd typically maintain an index of checkpoint IDs

	latestID, err := cm.redis.Get(ctx, KeyLatestCheckpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest checkpoint: %w", err)
	}

	if latestID == "" {
		return []string{}, nil
	}

	// For simplicity, return just the latest checkpoint
	// In production, you'd maintain a sorted set of checkpoint IDs
	return []string{latestID}, nil
}

// CleanupOldCheckpoints removes checkpoints older than the specified duration
func (cm *CheckpointManager) CleanupOldCheckpoints(ctx context.Context, maxAge time.Duration) error {
	// This is a simplified implementation
	// In production, you'd maintain an index of checkpoints with timestamps
	cm.logger.Infof("Checkpoint cleanup would remove checkpoints older than %v", maxAge)

	// The TTL on checkpoint keys handles automatic cleanup
	// This method could be enhanced to manually clean up based on business logic

	return nil
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
	checkpoints, err := cm.ListCheckpoints(ctx)
	if err == nil {
		health["checkpoint_count"] = len(checkpoints)
	}

	return health
}
