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

	// Utility methods
	GetRedisInfo() map[string]interface{}
	Client() *redis.Client
}