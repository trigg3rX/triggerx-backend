package redis

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

// Client represents a Redis client
type Client struct {
	client *redis.Client
	logger logging.Logger
}

// NewClient creates a new Redis client using Upstash credentials from environment variables
func NewClient(logger logging.Logger) (*Client, error) {
	// Get Redis credentials from environment variables
	redisURL := os.Getenv("UPSTASH_REDIS_URL")
	// If using REST-based API, get the token
	redisToken := os.Getenv("UPSTASH_REDIS_REST_TOKEN")

	if redisURL == "" {
		return nil, errors.New("UPSTASH_REDIS_URL environment variable is not set")
	}

	// Create options for Redis client
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Redis URL: %w", err)
	}

	// If using REST-based API with a token, set the password
	if redisToken != "" {
		opt.Password = redisToken
	}

	// Create Redis client
	client := redis.NewClient(opt)

	// Create our client wrapper
	redisClient := &Client{
		client: client,
		logger: logger,
	}

	// Verify connection
	if err := redisClient.CheckConnection(); err != nil {
		return nil, err
	}

	return redisClient, nil
}

// CheckConnection tests the Redis connection
func (c *Client) CheckConnection() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := c.client.Ping(ctx).Result()
	if err != nil {
		c.logger.Errorf("Failed to connect to Redis: %v", err)
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}

	c.logger.Info("Successfully connected to Redis")
	return nil
}

// Get retrieves a value from Redis by key
func (c *Client) Get(ctx context.Context, key string) (string, error) {
	val, err := c.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", nil // Key does not exist
	} else if err != nil {
		return "", err
	}
	return val, nil
}

// Set stores a key-value pair in Redis with an optional expiration
func (c *Client) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return c.client.Set(ctx, key, value, expiration).Err()
}

// Del removes keys from Redis
func (c *Client) Del(ctx context.Context, keys ...string) error {
	return c.client.Del(ctx, keys...).Err()
}

// Incr increments the given key
func (c *Client) Incr(ctx context.Context, key string) (int64, error) {
	return c.client.Incr(ctx, key).Result()
}

// Expire sets an expiration time on key
func (c *Client) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return c.client.Expire(ctx, key, expiration).Err()
}

// TTL returns the remaining time to live of a key
func (c *Client) TTL(ctx context.Context, key string) (time.Duration, error) {
	return c.client.TTL(ctx, key).Result()
}

// Eval executes a Lua script
func (c *Client) Eval(ctx context.Context, script string, keys []string, args ...interface{}) (interface{}, error) {
	return c.client.Eval(ctx, script, keys, args...).Result()
}

// Client returns the underlying Redis client if direct access is needed
func (c *Client) Client() *redis.Client {
	return c.client
}

// Close closes the Redis client connection
func (c *Client) Close() error {
	return c.client.Close()
} 