package sync

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/trigg3rX/triggerx-backend/pkg/client/redis"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

// Redis keys for storing state
const (
	KeyLastEthBlockUpdated  = "registrar:state:last_eth_block_updated"
	KeyLastBaseBlockUpdated = "registrar:state:last_base_block_updated"
	KeyLastRewardsUpdate    = "registrar:state:last_rewards_update"
)

// StateManager manages blockchain synchronization state in Redis
type StateManager struct {
	redis  *redis.Client
	logger logging.Logger
}

// BlockchainState represents the current state of blockchain synchronization
type BlockchainState struct {
	LastEthBlockUpdated  uint64    `json:"last_eth_block_updated"`
	LastBaseBlockUpdated uint64    `json:"last_base_block_updated"`
	LastRewardsUpdate    time.Time `json:"last_rewards_update"`
	UpdatedAt            time.Time `json:"updated_at"`
}

// NewStateManager creates a new Redis-backed state manager
func NewStateManager(redis *redis.Client, logger logging.Logger) *StateManager {
	return &StateManager{
		redis:  redis,
		logger: logger,
	}
}

// InitializeState initializes the state in Redis with current blockchain data
func (sm *StateManager) InitializeState(ctx context.Context, ethBlock uint64, baseBlock uint64, optBlock uint64, rewardsUpdate time.Time) error {
	sm.logger.Info("Initializing blockchain state in Redis")

	// Set initial block numbers if they don't exist
	if err := sm.setBlockIfNotExists(ctx, KeyLastEthBlockUpdated, ethBlock); err != nil {
		return fmt.Errorf("failed to initialize ETH block: %w", err)
	}

	if err := sm.setBlockIfNotExists(ctx, KeyLastBaseBlockUpdated, baseBlock); err != nil {
		return fmt.Errorf("failed to initialize BASE block: %w", err)
	}

	// Set initial rewards update time if it doesn't exist
	if err := sm.setTimeIfNotExists(ctx, KeyLastRewardsUpdate, rewardsUpdate); err != nil {
		return fmt.Errorf("failed to initialize rewards update time: %w", err)
	}

	sm.logger.Info("Blockchain state initialized successfully",
		"eth_block", ethBlock,
		"base_block", baseBlock,
		"last_rewards_update", rewardsUpdate,
	)

	return nil
}

// GetLastPolledEthBlock gets the last polled Ethereum block number
func (sm *StateManager) GetLastEthBlockUpdated(ctx context.Context) (uint64, error) {
	return sm.getBlockNumber(ctx, KeyLastEthBlockUpdated)
}

// SetLastPolledEthBlock sets the last polled Ethereum block number
func (sm *StateManager) SetLastEthBlockUpdated(ctx context.Context, blockNumber uint64) error {
	return sm.setBlockNumber(ctx, KeyLastEthBlockUpdated, blockNumber)
}

// GetLastPolledBaseBlock gets the last polled Base block number
func (sm *StateManager) GetLastBaseBlockUpdated(ctx context.Context) (uint64, error) {
	return sm.getBlockNumber(ctx, KeyLastBaseBlockUpdated)
}

// SetLastPolledBaseBlock sets the last polled Base block number
func (sm *StateManager) SetLastBaseBlockUpdated(ctx context.Context, blockNumber uint64) error {
	return sm.setBlockNumber(ctx, KeyLastBaseBlockUpdated, blockNumber)
}

// GetLastRewardsUpdate gets the last rewards update timestamp
func (sm *StateManager) GetLastRewardsUpdate(ctx context.Context) (time.Time, error) {
	timestampStr, err := sm.redis.Get(ctx, KeyLastRewardsUpdate)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to get last rewards update: %w", err)
	}

	if timestampStr == "" {
		// Return previous day's 6:30 AM UTC
		return time.Now().AddDate(0, 0, -1).Add(6*time.Hour + 30*time.Minute).UTC(), nil
	}

	timestamp, err := time.Parse(time.RFC3339, timestampStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse rewards update timestamp: %w", err)
	}

	return timestamp, nil
}

// SetLastRewardsUpdate sets the last rewards update timestamp
func (sm *StateManager) SetLastRewardsUpdate(ctx context.Context, timestamp time.Time) error {
	timestampStr := timestamp.UTC().Format(time.RFC3339)

	if err := sm.redis.Set(ctx, KeyLastRewardsUpdate, timestampStr, 0); err != nil {
		return fmt.Errorf("failed to set last rewards update: %w", err)
	}

	sm.logger.Debugf("Updated last rewards update to %s", timestampStr)
	return nil
}

// GetFullState gets the complete blockchain state
func (sm *StateManager) GetFullState(ctx context.Context) (*BlockchainState, error) {
	ethBlock, err := sm.GetLastEthBlockUpdated(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get ETH block: %w", err)
	}

	baseBlock, err := sm.GetLastBaseBlockUpdated(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get BASE block: %w", err)
	}

	rewardsUpdate, err := sm.GetLastRewardsUpdate(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get rewards update: %w", err)
	}

	return &BlockchainState{
		LastEthBlockUpdated:  ethBlock,
		LastBaseBlockUpdated: baseBlock,
		LastRewardsUpdate:    rewardsUpdate,
		UpdatedAt:            time.Now().UTC(),
	}, nil
}

// UpdateBlockchainProgress updates multiple blockchain states atomically
func (sm *StateManager) UpdateBlockchainProgress(ctx context.Context, ethBlock, baseBlock *uint64) error {
	// Use Redis pipeline for atomic updates
	pipe := sm.redis.Client().Pipeline()

	if ethBlock != nil {
		pipe.Set(ctx, KeyLastEthBlockUpdated, strconv.FormatUint(*ethBlock, 10), 0)
	}

	if baseBlock != nil {
		pipe.Set(ctx, KeyLastBaseBlockUpdated, strconv.FormatUint(*baseBlock, 10), 0)
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to update blockchain progress: %w", err)
	}

	sm.logger.Debugf("Updated blockchain progress - ETH: %v, BASE: %v",
		formatBlockPtr(ethBlock), formatBlockPtr(baseBlock))

	return nil
}

// ResetState resets all state to initial values
func (sm *StateManager) ResetState(ctx context.Context) error {
	keys := []string{
		KeyLastEthBlockUpdated,
		KeyLastBaseBlockUpdated,
		KeyLastRewardsUpdate,
	}

	if err := sm.redis.Del(ctx, keys...); err != nil {
		return fmt.Errorf("failed to reset state: %w", err)
	}

	sm.logger.Info("Blockchain state reset successfully")
	return nil
}

// GetStateHealth returns health information about the state
func (sm *StateManager) GetStateHealth(ctx context.Context) map[string]interface{} {
	health := map[string]interface{}{
		"redis_connected": true,
		"state_keys":      make(map[string]interface{}),
	}

	// Check if Redis is accessible
	if err := sm.redis.CheckConnection(ctx); err != nil {
		health["redis_connected"] = false
		health["error"] = err.Error()
		return health
	}

	// Check each state key
	keys := map[string]string{
		"eth_block":      KeyLastEthBlockUpdated,
		"base_block":     KeyLastBaseBlockUpdated,
		"rewards_update": KeyLastRewardsUpdate,
	}

	for name, key := range keys {
		value, err := sm.redis.Get(ctx, key)
		ttl, ttlErr := sm.redis.TTL(ctx, key)

		keyHealth := map[string]interface{}{
			"exists": err == nil && value != "",
			"value":  value,
		}

		if ttlErr == nil {
			keyHealth["ttl"] = ttl.String()
		}

		if err != nil {
			keyHealth["error"] = err.Error()
		}

		health["state_keys"].(map[string]interface{})[name] = keyHealth
	}

	return health
}

// Helper methods
func (sm *StateManager) getBlockNumber(ctx context.Context, key string) (uint64, error) {
	blockStr, err := sm.redis.Get(ctx, key)
	if err != nil {
		return 0, fmt.Errorf("failed to get block number for key %s: %w", key, err)
	}

	if blockStr == "" {
		return 0, nil
	}

	blockNumber, err := strconv.ParseUint(blockStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse block number %s: %w", blockStr, err)
	}

	return blockNumber, nil
}

func (sm *StateManager) setBlockNumber(ctx context.Context, key string, blockNumber uint64) error {
	blockStr := strconv.FormatUint(blockNumber, 10)

	if err := sm.redis.Set(ctx, key, blockStr, 0); err != nil {
		return fmt.Errorf("failed to set block number for key %s: %w", key, err)
	}

	sm.logger.Debugf("Updated %s to block %d", key, blockNumber)
	return nil
}

func (sm *StateManager) setBlockIfNotExists(ctx context.Context, key string, blockNumber uint64) error {
	existing, err := sm.redis.Get(ctx, key)
	if err != nil {
		return err
	}

	// Only set if key doesn't exist or is empty
	if existing == "" {
		return sm.setBlockNumber(ctx, key, blockNumber)
	}

	sm.logger.Debugf("Key %s already exists with value %s, skipping initialization", key, existing)
	return nil
}

func (sm *StateManager) setTimeIfNotExists(ctx context.Context, key string, timestamp time.Time) error {
	existing, err := sm.redis.Get(ctx, key)
	if err != nil {
		return err
	}

	// Only set if key doesn't exist or is empty
	if existing == "" {
		timestampStr := timestamp.UTC().Format(time.RFC3339)
		return sm.redis.Set(ctx, key, timestampStr, 0)
	}

	sm.logger.Debugf("Key %s already exists with value %s, skipping initialization", key, existing)
	return nil
}

func formatBlockPtr(block *uint64) interface{} {
	if block == nil {
		return "unchanged"
	}
	return *block
}
