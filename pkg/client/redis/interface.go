package redis

import (
	"context"
	"time"

	redis "github.com/redis/go-redis/v9"
)

// RedisClientInterface defines the interface for Redis client operations
type RedisClientInterface interface {
	// Connection management
	CheckConnection(ctx context.Context) error
	Ping(ctx context.Context) error
	Close() error

	// Distributed Locking
	NewLock(key string, ttl time.Duration, retryStrategy *RetryStrategy) (*Lock, error)

	// Pipeline and Scripting
	ExecutePipeline(ctx context.Context, fn PipelineFunc) ([]redis.Cmder, error)
	Eval(ctx context.Context, script string, keys []string, args ...interface{}) (interface{}, error)

	// Core Redis operations
	Get(ctx context.Context, key string) (string, error)
	GetWithExists(ctx context.Context, key string) (value string, exists bool, err error)
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error)
	Del(ctx context.Context, keys ...string) error
	DelWithCount(ctx context.Context, keys ...string) (deletedCount int64, err error)

	// Hash operations
	HSet(ctx context.Context, key string, values ...interface{}) error
	HGet(ctx context.Context, key, field string) (string, error)
	HGetWithExists(ctx context.Context, key, field string) (value string, exists bool, err error)
	HDel(ctx context.Context, key string, fields ...string) error
	HDelWithCount(ctx context.Context, key string, fields ...string) (deletedCount int64, err error)

	// Safe key scanning operations (alternatives to KEYS command)
	Scan(ctx context.Context, cursor uint64, options *ScanOptions) (*ScanResult, error)
	ScanAll(ctx context.Context, options *ScanOptions) ([]string, error)
	ScanKeysByPattern(ctx context.Context, pattern string, count int64) (*ScanResult, error)
	ScanKeysByType(ctx context.Context, keyType string, count int64) (*ScanResult, error)

	// Stream operations
	CreateStreamIfNotExists(ctx context.Context, stream string, ttl time.Duration) error
	CreateConsumerGroup(ctx context.Context, stream, group string) error
	CreateConsumerGroupAtomic(ctx context.Context, stream, group string) (created bool, err error)
	CreateStreamWithConsumerGroup(ctx context.Context, stream, group string, ttl time.Duration) error
	XAdd(ctx context.Context, args *redis.XAddArgs) (string, error)
	XLen(ctx context.Context, stream string) (int64, error)
	XReadGroup(ctx context.Context, args *redis.XReadGroupArgs) ([]redis.XStream, error)
	XAck(ctx context.Context, stream, group, id string) error
	XPending(ctx context.Context, stream, group string) (*redis.XPending, error)
	XPendingExt(ctx context.Context, args *redis.XPendingExtArgs) ([]redis.XPendingExt, error)
	XClaim(ctx context.Context, args *redis.XClaimArgs) *redis.XMessageSliceCmd

	// Sorted Set operations
	ZAdd(ctx context.Context, key string, members ...redis.Z) (int64, error)
	ZAddWithExists(ctx context.Context, key string, members ...redis.Z) (newElements int64, keyExisted bool, err error)
	ZRevRange(ctx context.Context, key string, start, stop int64) ([]string, error)
	ZRangeByScore(ctx context.Context, key, min, max string) ([]string, error)
	ZRemRangeByScore(ctx context.Context, key, min, max string) (int64, error)
	ZCard(ctx context.Context, key string) (int64, error)

	// TTL management
	TTL(ctx context.Context, key string) (time.Duration, error)
	RefreshTTL(ctx context.Context, key string, ttl time.Duration) error
	RefreshStreamTTL(ctx context.Context, stream string, ttl time.Duration) error
	SetTTL(ctx context.Context, key string, ttl time.Duration) error
	GetTTLStatus(ctx context.Context, key string) (time.Duration, bool, error)

	// Health and monitoring
	GetHealthStatus(ctx context.Context) *HealthStatus
	IsHealthy(ctx context.Context) bool
	PerformHealthCheck(ctx context.Context) (*HealthCheckResult, error)
	GetConnectionStatus() *ConnectionStatus
	GetOperationMetrics() map[string]*OperationMetrics
	ResetOperationMetrics()

	// Configuration and monitoring hooks
	SetMonitoringHooks(hooks *MonitoringHooks)
	SetRetryConfig(config *RetryConfig)
	SetConnectionRecoveryConfig(config *ConnectionRecoveryConfig)

	// Utility methods
	Client() *redis.Client
}
