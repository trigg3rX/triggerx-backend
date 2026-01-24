package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/time/config"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
)

const (
	// Redis key prefix for storing the last poll timestamp
	lastPollTimestampKey = "time_scheduler:last_poll_timestamp"
)

// Client wraps the Redis client for time scheduler operations
type Client struct {
	client *redis.Client
	logger observability.Logger
}

// NewClient creates a new Redis client for the time scheduler service
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

// GetLastPollTimestamp retrieves the last poll timestamp for crash recovery
// Returns zero time if no timestamp has been stored yet
func (c *Client) GetLastPollTimestamp(ctx context.Context, schedulerID string) (time.Time, error) {
	key := fmt.Sprintf("%s:%s", lastPollTimestampKey, schedulerID)

	val, err := c.client.Get(ctx, key).Result()
	if err == redis.Nil {
		// No timestamp stored yet, return zero time
		return time.Time{}, nil
	} else if err != nil {
		c.logger.Error(ctx, "Failed to get last poll timestamp",
			observability.Error(err),
			observability.String("scheduler_id", schedulerID),
		)
		return time.Time{}, fmt.Errorf("failed to get last poll timestamp: %w", err)
	}

	timestamp, err := time.Parse(time.RFC3339Nano, val)
	if err != nil {
		c.logger.Error(ctx, "Failed to parse last poll timestamp",
			observability.Error(err),
			observability.String("value", val),
		)
		return time.Time{}, fmt.Errorf("failed to parse last poll timestamp: %w", err)
	}

	return timestamp, nil
}

// SetLastPollTimestamp stores the last poll timestamp for crash recovery
func (c *Client) SetLastPollTimestamp(ctx context.Context, schedulerID string, timestamp time.Time) error {
	key := fmt.Sprintf("%s:%s", lastPollTimestampKey, schedulerID)

	// Store with TTL to auto-expire old entries (24 hours)
	err := c.client.Set(ctx, key, timestamp.Format(time.RFC3339Nano), 24*time.Hour).Err()
	if err != nil {
		c.logger.Error(ctx, "Failed to set last poll timestamp",
			observability.Error(err),
			observability.String("scheduler_id", schedulerID),
			observability.Time("timestamp", timestamp),
		)
		return fmt.Errorf("failed to set last poll timestamp: %w", err)
	}

	c.logger.Debug(ctx, "Stored last poll timestamp",
		observability.String("scheduler_id", schedulerID),
		observability.Time("timestamp", timestamp),
	)
	return nil
}

// Close closes the Redis connection
func (c *Client) Close() error {
	return c.client.Close()
}
