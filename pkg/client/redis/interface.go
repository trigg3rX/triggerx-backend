package redis

import (
	"context"
	"time"

	redis "github.com/redis/go-redis/v9"
)

// RedisClientInterface defines the interface for Redis client operations
type RedisClientInterface interface {
	// Connection management
	CheckConnection() error
	Ping() error
	Close() error

	// Core Redis operations
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error)
	Del(ctx context.Context, keys ...string) error
	TTL(ctx context.Context, key string) (time.Duration, error)

	// Stream operations
	CreateStreamIfNotExists(ctx context.Context, stream string, ttl time.Duration) error
	CreateConsumerGroup(ctx context.Context, stream, group string) error
	XAdd(ctx context.Context, args *redis.XAddArgs) (string, error)
	XLen(ctx context.Context, stream string) (int64, error)
	XReadGroup(ctx context.Context, args *redis.XReadGroupArgs) ([]redis.XStream, error)
	XAck(ctx context.Context, stream, group, id string) error
	XPending(ctx context.Context, stream, group string) (*redis.XPending, error)
	XPendingExt(ctx context.Context, args *redis.XPendingExtArgs) ([]redis.XPendingExt, error)
	XClaim(ctx context.Context, args *redis.XClaimArgs) *redis.XMessageSliceCmd

	// TTL management
	RefreshTTL(ctx context.Context, key string, ttl time.Duration) error
	RefreshStreamTTL(ctx context.Context, stream string, ttl time.Duration) error
	SetTTL(ctx context.Context, key string, ttl time.Duration) error
	GetTTLStatus(ctx context.Context, key string) (time.Duration, bool, error)

	// Health and monitoring
	GetHealthStatus(ctx context.Context) *HealthStatus
	IsHealthy(ctx context.Context) bool
	PerformHealthCheck(ctx context.Context) (map[string]interface{}, error)
	GetConnectionStatus() map[string]interface{}
	GetOperationMetrics() map[string]*OperationMetrics
	ResetOperationMetrics()

	// Configuration and monitoring
	SetMonitoringHooks(hooks *MonitoringHooks)
	SetRetryConfig(config *RetryConfig)
	SetConnectionRecoveryConfig(config *ConnectionRecoveryConfig)

	// Utility methods
	GetRedisInfo() map[string]interface{}
	Client() *redis.Client
}
