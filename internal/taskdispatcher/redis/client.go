package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/trigg3rX/triggerx-backend/internal/taskdispatcher/config"
	pkgredis "github.com/trigg3rX/triggerx-backend/pkg/client/redis"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
)

const (
	// TaskIndexKey is the Redis hash key that maps taskID to messageID
	TaskIndexKey = "task_id_to_message_id"
	// TaskIndexTTL is the TTL for the task index hash
	TaskIndexTTL = 2 * time.Hour

	// DispatchedTimeoutsKey is the Redis sorted set key for tracking dispatched task timeouts
	DispatchedTimeoutsKey = "dispatched_timeouts"
	// StreamExpirationKeyPrefix is the prefix for sorted sets tracking stream entry expiration
	StreamExpirationKeyPrefix = "stream:expiration:"
	// ExpirationTrackingTTL is the TTL for expiration tracking sorted sets
	ExpirationTrackingTTL = 48 * time.Hour
)

// Client wraps the Redis client for task stream management operations
type Client struct {
	client pkgredis.RedisClientInterface
	logger observability.Logger
}

// NewClient creates a new Redis client wrapper for the taskdispatcher service
func NewClient(ctx context.Context, logger observability.Logger) (*Client, error) {
	redisConfig := config.GetRedisClientConfig()
	client, err := pkgredis.NewRedisClient(ctx, logger, redisConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Redis client: %w", err)
	}

	if err := client.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping Redis: %w", err)
	}

	return &Client{
		client: client,
		logger: logger,
	}, nil
}

// GetUnderlyingClient returns the underlying Redis client interface for advanced operations
func (c *Client) GetUnderlyingClient() pkgredis.RedisClientInterface {
	return c.client
}

// SetMonitoringHooks sets monitoring hooks for metrics integration
func (c *Client) SetMonitoringHooks(hooks *pkgredis.MonitoringHooks) {
	c.client.SetMonitoringHooks(hooks)
}

// Close closes the Redis connection
func (c *Client) Close() error {
	return c.client.Close()
}

// SetTTL sets a TTL on a key
func (c *Client) SetTTL(ctx context.Context, key string, ttl time.Duration) error {
	return c.client.SetTTL(ctx, key, ttl)
}

// ZAdd adds a member to a sorted set
func (c *Client) ZAdd(ctx context.Context, key string, members ...redis.Z) (int64, error) {
	return c.client.ZAdd(ctx, key, members...)
}

// ZRangeByScore returns members of a sorted set by score range
func (c *Client) ZRangeByScore(ctx context.Context, key string, min string, max string) ([]string, error) {
	return c.client.ZRangeByScore(ctx, key, min, max)
}
