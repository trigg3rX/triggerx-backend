package redis

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/trigg3rX/triggerx-backend/internal/redis/config"
	"github.com/trigg3rX/triggerx-backend/pkg/client/aggregator"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

// Client represents a Redis client with logging capabilities
type Client struct {
	redisClient *redis.Client
	aggClient   *aggregator.AggregatorClient
	logger      logging.Logger
}

var (
	redisClient *redis.Client
	aggClient   *aggregator.AggregatorClient
	once        sync.Once
)

// GetRedisClient returns a singleton Redis client
func GetRedisClient() *redis.Client {
	once.Do(func() {
		redisClient = redis.NewClient(&redis.Options{
			Addr:     config.GetRedisAddr(),
			Password: config.GetRedisPassword(),
			DB:       config.GetRedisDB(),
		})
	})
	return redisClient
}

// NewClient creates a new Redis client instance with enhanced features
func NewRedisClient(logger logging.Logger) (*Client, error) {
	if !config.IsRedisAvailable() {
		return nil, fmt.Errorf("redis is not configured. Please set UPSTASH_REDIS_URL or enable local Redis with REDIS_LOCAL_ENABLED=true")
	}

	opt, err := parseRedisConfigFromSettings()
	if err != nil {
		return nil, err
	}

	aggClient, err := aggregator.NewAggregatorClient(
		logger,
		aggregator.AggregatorClientConfig{
			AggregatorRPCUrl: config.GetAggregatorRPCURL(),
			SenderPrivateKey: config.GetDispatcherPrivateKey(),
			SenderAddress:    config.GetDispatcherAddress(),
			RetryAttempts:    3,
			RetryDelay:       1 * time.Second,
			RequestTimeout:   10 * time.Second,
		},
	)
	if err != nil {
		return nil, err
	}

	client := redis.NewClient(opt)

	redisClient := &Client{
		redisClient: client,
		aggClient: aggClient,
		logger: logger,
	}

	if err := redisClient.CheckConnection(); err != nil {
		return nil, err
	}

	logger.Infof("Connected to Redis (%s)", config.GetRedisType())
	return redisClient, nil
}

// parseRedisConfigFromSettings determines Redis configuration from config settings
func parseRedisConfigFromSettings() (*redis.Options, error) {
	// Try Upstash configuration first (cloud Redis - preferred)
	if config.IsUpstashEnabled() {
		opt, err := redis.ParseURL(config.GetUpstashURL())
		if err != nil {
			return nil, fmt.Errorf("failed to parse Upstash Redis URL: %w", err)
		}

		// Set token as password if provided
		if config.GetUpstashToken() != "" {
			opt.Password = config.GetUpstashToken()
		}

		// Apply additional settings
		opt.PoolSize = config.GetPoolSize()
		opt.MinIdleConns = config.GetMinIdleConns()
		opt.MaxRetries = config.GetMaxRetries()
		opt.DialTimeout = config.GetDialTimeout()
		opt.ReadTimeout = config.GetReadTimeout()
		opt.WriteTimeout = config.GetWriteTimeout()
		opt.PoolTimeout = config.GetPoolTimeout()

		return opt, nil
	}

	// Fall back to local Redis configuration (if explicitly enabled)
	if config.IsLocalRedisEnabled() {
		return &redis.Options{
			Addr:         config.GetRedisAddr(),
			Password:     config.GetRedisPassword(),
			DB:           config.GetRedisDB(),
			PoolSize:     config.GetPoolSize(),
			MinIdleConns: config.GetMinIdleConns(),
			MaxRetries:   config.GetMaxRetries(),
			DialTimeout:  config.GetDialTimeout(),
			ReadTimeout:  config.GetReadTimeout(),
			WriteTimeout: config.GetWriteTimeout(),
			PoolTimeout:  config.GetPoolTimeout(),
		}, nil
	}

	return nil, fmt.Errorf("no Redis configuration available")
}

// CheckConnection validates the Redis connection
func (c *Client) CheckConnection() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := c.redisClient.Ping(ctx).Result()
	if err != nil {
		c.logger.Errorf("Failed to connect to Redis (%s): %v", config.GetRedisType(), err)
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}

	c.logger.Infof("Successfully connected to Redis (%s)", config.GetRedisType())
	return nil
}

// Ping checks if Redis is reachable (instance method)
func (c *Client) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	return c.redisClient.Ping(ctx).Err()
}

// IsAvailable checks if Redis is configured and available
func IsAvailable() bool {
	return config.IsRedisAvailable()
}

// GetRedisInfo returns information about the current Redis configuration
func GetRedisInfo() map[string]interface{} {
	return map[string]interface{}{
		"available":     config.IsRedisAvailable(),
		"type":          config.GetRedisType(),
		"upstash":       config.IsUpstashEnabled(),
		"local":         config.IsLocalRedisEnabled(),
		"stream_ttl":    config.GetStreamTTL().String(),
		"stream_maxlen": config.GetStreamMaxLen(),
	}
}

// Enhanced Redis operations with context support
func (c *Client) Get(ctx context.Context, key string) (string, error) {
	val, err := c.redisClient.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", nil
	} else if err != nil {
		return "", err
	}
	return val, nil
}

func (c *Client) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return c.redisClient.Set(ctx, key, value, expiration).Err()
}

// SetNX sets key to hold string value if key does not exist (atomic operation)
func (c *Client) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	return c.redisClient.SetNX(ctx, key, value, expiration).Result()
}

func (c *Client) Del(ctx context.Context, keys ...string) error {
	return c.redisClient.Del(ctx, keys...).Err()
}

func (c *Client) Incr(ctx context.Context, key string) (int64, error) {
	return c.redisClient.Incr(ctx, key).Result()
}

func (c *Client) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return c.redisClient.Expire(ctx, key, expiration).Err()
}

func (c *Client) TTL(ctx context.Context, key string) (time.Duration, error) {
	return c.redisClient.TTL(ctx, key).Result()
}

func (c *Client) Eval(ctx context.Context, script string, keys []string, args ...interface{}) (interface{}, error) {
	return c.redisClient.Eval(ctx, script, keys, args...).Result()
}

func (c *Client) EvalScript(ctx context.Context, script string, keys []string, args []interface{}) (interface{}, error) {
	return c.redisClient.Eval(ctx, script, keys, args...).Result()
}

// Info returns Redis server information
func (c *Client) Info(ctx context.Context, section string) (string, error) {
	return c.redisClient.Info(ctx, section).Result()
}

// XAdd adds an entry to a Redis stream
func (c *Client) XAdd(ctx context.Context, args *redis.XAddArgs) (string, error) {
	return c.redisClient.XAdd(ctx, args).Result()
}

// XLen returns the length of a Redis stream
func (c *Client) XLen(ctx context.Context, stream string) (int64, error) {
	return c.redisClient.XLen(ctx, stream).Result()
}

// Client returns the underlying Redis client
func (c *Client) Client() *redis.Client {
	return c.redisClient
}

// Close closes the Redis connection
func (c *Client) Close() error {
	return c.redisClient.Close()
}
