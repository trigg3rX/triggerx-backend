package redis

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/trigg3rX/triggerx-backend/internal/eventmonitor/config"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
)

const (
	// Redis key prefix for storing the last polled block number
	lastBlockKeyPrefix = "event_monitor:last_block"
)

// Client wraps the Redis client for event monitor operations
type Client struct {
	client *redis.Client
	logger observability.Logger
}

// NewClient creates a new Redis client for the event monitor service
func NewClient(logger observability.Logger) (*Client, error) {
	opt, err := redis.ParseURL(config.GetUpstashRedisUrl())
	if err != nil {
		return nil, fmt.Errorf("failed to parse Redis URL: %w", err)
	}

	if config.GetUpstashRedisRestToken() != "" {
		opt.Password = config.GetUpstashRedisRestToken()
	}

	client := redis.NewClient(opt)

	redisClient := &Client{
		client: client,
		logger: logger,
	}

	if err := redisClient.CheckConnection(); err != nil {
		return nil, err
	}

	return redisClient, nil
}

// CheckConnection validates the Redis connection
func (c *Client) CheckConnection() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := c.client.Ping(ctx).Result()
	if err != nil {
		c.logger.Error(ctx, "Failed to connect to Redis", observability.Error(err))
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}

	c.logger.Info(ctx, "Successfully connected to Redis")
	return nil
}

// GetLastBlock retrieves the last polled block number for a specific chain
// Returns 0 if no block number has been stored yet
func (c *Client) GetLastBlock(ctx context.Context, chainID string) (uint64, error) {
	key := fmt.Sprintf("%s:%s", lastBlockKeyPrefix, chainID)

	val, err := c.client.Get(ctx, key).Result()
	if err == redis.Nil {
		// No block stored yet, return 0
		return 0, nil
	} else if err != nil {
		c.logger.Error(ctx, "Failed to get last block",
			observability.Error(err),
			observability.String("chain_id", chainID),
		)
		return 0, fmt.Errorf("failed to get last block: %w", err)
	}

	blockNumber, err := strconv.ParseUint(val, 10, 64)
	if err != nil {
		c.logger.Error(ctx, "Failed to parse last block number",
			observability.Error(err),
			observability.String("value", val),
		)
		return 0, fmt.Errorf("failed to parse last block number: %w", err)
	}

	return blockNumber, nil
}

// SetLastBlock stores the last polled block number for a specific chain
func (c *Client) SetLastBlock(ctx context.Context, chainID string, blockNumber uint64) error {
	key := fmt.Sprintf("%s:%s", lastBlockKeyPrefix, chainID)

	// Store with TTL to auto-expire old entries (24 hours)
	err := c.client.Set(ctx, key, strconv.FormatUint(blockNumber, 10), 24*time.Hour).Err()
	if err != nil {
		c.logger.Error(ctx, "Failed to set last block",
			observability.Error(err),
			observability.String("chain_id", chainID),
			observability.Uint64("block_number", blockNumber),
		)
		return fmt.Errorf("failed to set last block: %w", err)
	}

	c.logger.Debug(ctx, "Stored last block",
		observability.String("chain_id", chainID),
		observability.Uint64("block_number", blockNumber),
	)
	return nil
}

// GetLastBlockForWorker retrieves the last polled block for a specific worker key
// Worker key format: chainID:contractAddr:eventSig
func (c *Client) GetLastBlockForWorker(ctx context.Context, workerKey string) (uint64, error) {
	key := fmt.Sprintf("%s:worker:%s", lastBlockKeyPrefix, workerKey)

	val, err := c.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return 0, nil
	} else if err != nil {
		c.logger.Error(ctx, "Failed to get last block for worker",
			observability.Error(err),
			observability.String("worker_key", workerKey),
		)
		return 0, fmt.Errorf("failed to get last block for worker: %w", err)
	}

	blockNumber, err := strconv.ParseUint(val, 10, 64)
	if err != nil {
		c.logger.Error(ctx, "Failed to parse last block for worker",
			observability.Error(err),
			observability.String("value", val),
		)
		return 0, fmt.Errorf("failed to parse last block for worker: %w", err)
	}

	return blockNumber, nil
}

// SetLastBlockForWorker stores the last polled block for a specific worker key
func (c *Client) SetLastBlockForWorker(ctx context.Context, workerKey string, blockNumber uint64) error {
	key := fmt.Sprintf("%s:worker:%s", lastBlockKeyPrefix, workerKey)

	// Store with TTL to auto-expire old entries (24 hours)
	err := c.client.Set(ctx, key, strconv.FormatUint(blockNumber, 10), 24*time.Hour).Err()
	if err != nil {
		c.logger.Error(ctx, "Failed to set last block for worker",
			observability.Error(err),
			observability.String("worker_key", workerKey),
			observability.Uint64("block_number", blockNumber),
		)
		return fmt.Errorf("failed to set last block for worker: %w", err)
	}

	c.logger.Debug(ctx, "Stored last block for worker",
		observability.String("worker_key", workerKey),
		observability.Uint64("block_number", blockNumber),
	)
	return nil
}

// Close closes the Redis connection
func (c *Client) Close() error {
	return c.client.Close()
}
