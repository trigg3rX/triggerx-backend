package redis

import (
	"context"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"
	"github.com/trigg3rX/triggerx-backend/internal/registrar/config"
	"github.com/trigg3rX/triggerx-backend/pkg/client/redis"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

// Client wraps the pkg Redis client for registrar-specific usage
type Client struct {
	client *redis.Client
}

// Delegate methods to the embedded Redis client

// Get retrieves a value by key
func (c *Client) Get(ctx context.Context, key string) (string, error) {
	return c.client.Get(ctx, key)
}

// Set stores a key-value pair with expiration
func (c *Client) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return c.client.Set(ctx, key, value, expiration)
}

// Del deletes one or more keys
func (c *Client) Del(ctx context.Context, keys ...string) error {
	return c.client.Del(ctx, keys...)
}

// CheckConnection validates the Redis connection
func (c *Client) CheckConnection() error {
	return c.client.CheckConnection()
}

// TTL returns the time-to-live for a key
func (c *Client) TTL(ctx context.Context, key string) (time.Duration, error) {
	return c.client.TTL(ctx, key)
}

// Client returns the underlying Redis client for operations like pipelines
func (c *Client) Client() *goredis.Client {
	return c.client.Client()
}

// NewClient creates a new Redis client instance using registrar configuration
func NewClient(logger logging.Logger) (*Client, error) {
	// Create Redis configuration from registrar config
	redisConfig := redis.RedisConfig{
		IsUpstash: true, // Always use Upstash for registrar
		UpstashConfig: redis.UpstashConfig{
			URL:   config.GetUpstashRedisUrl(),
			Token: config.GetUpstashRedisRestToken(),
		},
		ConnectionSettings: redis.ConnectionSettings{
			PoolSize:     10,
			MaxIdleConns: 10,
			MinIdleConns: 2,
			MaxRetries:   3,
			DialTimeout:  5 * time.Second,
			ReadTimeout:  3 * time.Second,
			WriteTimeout: 3 * time.Second,
			PoolTimeout:  4 * time.Second,
		},
		StreamsConfig: redis.StreamsConfig{
			JobStreamTTL:       120 * time.Hour,
			TaskStreamTTL:      1 * time.Hour,
			KeeperStreamTTL:    24 * time.Hour,
			RegistrarStreamTTL: 48 * time.Hour,
		},
	}

	// Create the Redis client
	client, err := redis.NewRedisClient(logger, redisConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Redis client: %w", err)
	}

	return &Client{client: client}, nil
}
