package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/trigg3rX/triggerx-backend/internal/taskmonitor/config"
	pkgredis "github.com/trigg3rX/triggerx-backend/pkg/client/redis"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
)

const (
	// TaskIndexKey is the Redis hash key that maps taskID to messageID
	TaskIndexKey = "task_id_to_message_id"
	// TaskIndexTTL is the TTL for the task index hash (should be longer than stream TTL)
	TaskIndexTTL = 2 * time.Hour

	// ExecutedTimeoutsKey is the Redis sorted set key for tracking executed task timeouts (pending validation)
	ExecutedTimeoutsKey = "executed_timeouts"
	// StreamExpirationKeyPrefix is the prefix for sorted sets tracking stream entry expiration
	StreamExpirationKeyPrefix = "stream:expiration:"
	// ExpirationTrackingTTL is the TTL for expiration tracking sorted sets
	ExpirationTrackingTTL = 1 * time.Hour
)

// Client wraps the Redis client for TaskMonitor service operations
type Client struct {
	client pkgredis.RedisClientInterface
	logger observability.Logger
}

// NewClient creates a new Redis client wrapper for the TaskMonitor service
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

// GetHealthStatus returns the health status of the Redis connection
func (c *Client) GetHealthStatus(ctx context.Context) *pkgredis.HealthStatus {
	return c.client.GetHealthStatus(ctx)
}

// GetConnectionStatus returns the connection status of the Redis client
func (c *Client) GetConnectionStatus() *pkgredis.ConnectionStatus {
	return c.client.GetConnectionStatus()
}

// GetOperationMetrics returns operation metrics from the Redis client
func (c *Client) GetOperationMetrics() map[string]*pkgredis.OperationMetrics {
	return c.client.GetOperationMetrics()
}
