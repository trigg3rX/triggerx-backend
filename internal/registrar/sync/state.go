package sync

import (
	"context"
	"fmt"
	"strconv"
	"time"

	redisClient "github.com/trigg3rX/triggerx-backend/internal/registrar/clients/redis"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

// Redis keys for storing state
const (
	KeyLastPolledEthBlock  = "registrar:state:last_polled_eth_block"
	KeyLastPolledBaseBlock = "registrar:state:last_polled_base_block"
	KeyLastPolledOptBlock  = "registrar:state:last_polled_opt_block"
	KeyLastRewardsUpdate   = "registrar:state:last_rewards_update"
)

// StateManager manages blockchain synchronization state in Redis
type StateManager struct {
	redis  *redisClient.Client
	logger logging.Logger
}

// BlockchainState represents the current state of blockchain synchronization
type BlockchainState struct {
	LastPolledEthBlock  uint64    `json:"last_polled_eth_block"`
	LastPolledBaseBlock uint64    `json:"last_polled_base_block"`
	LastPolledOptBlock  uint64    `json:"last_polled_opt_block"`
	LastRewardsUpdate   time.Time `json:"last_rewards_update"`
	UpdatedAt           time.Time `json:"updated_at"`
}

// NewStateManager creates a new Redis-backed state manager
func NewStateManager(redis *redisClient.Client, logger logging.Logger) *StateManager {
	return &StateManager{
		redis:  redis,
		logger: logger,
	}
}

// InitializeState initializes the state in Redis with current blockchain data
func (sm *StateManager) InitializeState(ctx context.Context, ethBlock, baseBlock, optBlock uint64) error {
	sm.logger.Info("Initializing blockchain state in Redis")

	// Set initial block numbers if they don't exist
	if err := sm.setBlockIfNotExists(ctx, KeyLastPolledEthBlock, ethBlock); err != nil {
		return fmt.Errorf("failed to initialize ETH block: %w", err)
	}

	if err := sm.setBlockIfNotExists(ctx, KeyLastPolledBaseBlock, baseBlock); err != nil {
		return fmt.Errorf("failed to initialize BASE block: %w", err)
	}

	if err := sm.setBlockIfNotExists(ctx, KeyLastPolledOptBlock, optBlock); err != nil {
		return fmt.Errorf("failed to initialize OPT block: %w", err)
	}

	// Set initial rewards update time if it doesn't exist
	if err := sm.setTimeIfNotExists(ctx, KeyLastRewardsUpdate, time.Now().UTC()); err != nil {
		return fmt.Errorf("failed to initialize rewards update time: %w", err)
	}

	sm.logger.Info("Blockchain state initialized successfully",
		"eth_block", ethBlock,
		"base_block", baseBlock,
		"opt_block", optBlock)

	return nil
}

// GetLastPolledEthBlock gets the last polled Ethereum block number
func (sm *StateManager) GetLastPolledEthBlock(ctx context.Context) (uint64, error) {
	return sm.getBlockNumber(ctx, KeyLastPolledEthBlock)
}

// SetLastPolledEthBlock sets the last polled Ethereum block number
func (sm *StateManager) SetLastPolledEthBlock(ctx context.Context, blockNumber uint64) error {
	return sm.setBlockNumber(ctx, KeyLastPolledEthBlock, blockNumber)
}

// GetLastPolledBaseBlock gets the last polled Base block number
func (sm *StateManager) GetLastPolledBaseBlock(ctx context.Context) (uint64, error) {
	return sm.getBlockNumber(ctx, KeyLastPolledBaseBlock)
}

// SetLastPolledBaseBlock sets the last polled Base block number
func (sm *StateManager) SetLastPolledBaseBlock(ctx context.Context, blockNumber uint64) error {
	return sm.setBlockNumber(ctx, KeyLastPolledBaseBlock, blockNumber)
}

// GetLastPolledOptBlock gets the last polled Optimism block number
func (sm *StateManager) GetLastPolledOptBlock(ctx context.Context) (uint64, error) {
	return sm.getBlockNumber(ctx, KeyLastPolledOptBlock)
}

// SetLastPolledOptBlock sets the last polled Optimism block number
func (sm *StateManager) SetLastPolledOptBlock(ctx context.Context, blockNumber uint64) error {
	return sm.setBlockNumber(ctx, KeyLastPolledOptBlock, blockNumber)
}

// GetLastRewardsUpdate gets the last rewards update timestamp
func (sm *StateManager) GetLastRewardsUpdate(ctx context.Context) (time.Time, error) {
	timestampStr, err := sm.redis.Get(ctx, KeyLastRewardsUpdate)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to get last rewards update: %w", err)
	}

	if timestampStr == "" {
		// Return zero time if not set
		return time.Time{}, nil
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
	ethBlock, err := sm.GetLastPolledEthBlock(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get ETH block: %w", err)
	}

	baseBlock, err := sm.GetLastPolledBaseBlock(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get BASE block: %w", err)
	}

	optBlock, err := sm.GetLastPolledOptBlock(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get OPT block: %w", err)
	}

	rewardsUpdate, err := sm.GetLastRewardsUpdate(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get rewards update: %w", err)
	}

	return &BlockchainState{
		LastPolledEthBlock:  ethBlock,
		LastPolledBaseBlock: baseBlock,
		LastPolledOptBlock:  optBlock,
		LastRewardsUpdate:   rewardsUpdate,
		UpdatedAt:           time.Now().UTC(),
	}, nil
}

// UpdateBlockchainProgress updates multiple blockchain states atomically
func (sm *StateManager) UpdateBlockchainProgress(ctx context.Context, ethBlock, baseBlock, optBlock *uint64) error {
	// Use Redis pipeline for atomic updates
	pipe := sm.redis.Client().Pipeline()

	if ethBlock != nil {
		pipe.Set(ctx, KeyLastPolledEthBlock, strconv.FormatUint(*ethBlock, 10), 0)
	}

	if baseBlock != nil {
		pipe.Set(ctx, KeyLastPolledBaseBlock, strconv.FormatUint(*baseBlock, 10), 0)
	}

	if optBlock != nil {
		pipe.Set(ctx, KeyLastPolledOptBlock, strconv.FormatUint(*optBlock, 10), 0)
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to update blockchain progress: %w", err)
	}

	sm.logger.Debugf("Updated blockchain progress - ETH: %v, BASE: %v, OPT: %v",
		formatBlockPtr(ethBlock), formatBlockPtr(baseBlock), formatBlockPtr(optBlock))

	return nil
}

// ResetState resets all state to initial values
func (sm *StateManager) ResetState(ctx context.Context) error {
	keys := []string{
		KeyLastPolledEthBlock,
		KeyLastPolledBaseBlock,
		KeyLastPolledOptBlock,
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
	if err := sm.redis.CheckConnection(); err != nil {
		health["redis_connected"] = false
		health["error"] = err.Error()
		return health
	}

	// Check each state key
	keys := map[string]string{
		"eth_block":      KeyLastPolledEthBlock,
		"base_block":     KeyLastPolledBaseBlock,
		"opt_block":      KeyLastPolledOptBlock,
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
