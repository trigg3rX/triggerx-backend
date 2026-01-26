package redis

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/trigg3rX/triggerx-backend/internal/health/config"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
)

const (
	// Redis key prefix for storing the last selected operator ID per network
	lastOperatorIDKeyPrefix = "health:last_operator_id:"
)

// Client wraps the Redis client for health service operations
type Client struct {
	client *redis.Client
	logger observability.Logger
}

// NewClient creates a new Redis client for the health service
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

// GetLastOperatorID retrieves the last selected operator ID for a network
// Returns 0 if no operator ID has been stored yet
func (c *Client) GetLastOperatorID(ctx context.Context, network string) (int, error) {
	key := lastOperatorIDKeyPrefix + network

	val, err := c.client.Get(ctx, key).Result()
	if err == redis.Nil {
		// No operator ID stored yet, return 0
		return 0, nil
	} else if err != nil {
		c.logger.Error(ctx, "Failed to get last operator ID",
			observability.Error(err),
			observability.String("network", network),
		)
		return 0, fmt.Errorf("failed to get last operator ID: %w", err)
	}

	operatorID, err := strconv.Atoi(val)
	if err != nil {
		c.logger.Error(ctx, "Failed to parse last operator ID",
			observability.Error(err),
			observability.String("value", val),
		)
		return 0, fmt.Errorf("failed to parse last operator ID: %w", err)
	}

	return operatorID, nil
}

// SetLastOperatorID stores the last selected operator ID for a network
func (c *Client) SetLastOperatorID(ctx context.Context, network string, operatorID int) error {
	key := lastOperatorIDKeyPrefix + network

	err := c.client.Set(ctx, key, strconv.Itoa(operatorID), 0).Err()
	if err != nil {
		c.logger.Error(ctx, "Failed to set last operator ID",
			observability.Error(err),
			observability.String("network", network),
			observability.Int("operator_id", operatorID),
		)
		return fmt.Errorf("failed to set last operator ID: %w", err)
	}

	c.logger.Debug(ctx, "Stored last operator ID",
		observability.String("network", network),
		observability.Int("operator_id", operatorID),
	)
	return nil
}

// Close closes the Redis connection
func (c *Client) Close() error {
	return c.client.Close()
}
