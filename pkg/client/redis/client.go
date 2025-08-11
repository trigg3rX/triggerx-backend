package redis

import (
	"context"
	"fmt"
	"sync"
	"time"

	redis "github.com/redis/go-redis/v9"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

// Client represents a Redis client with logging capabilities
type Client struct {
	redisClient      *redis.Client
	config           RedisConfig
	logger           logging.Logger
	mu               sync.Mutex
	retryConfig      *RetryConfig
	recoveryConfig   *ConnectionRecoveryConfig
	isRecovering     bool
	lastHealthCheck  time.Time
	monitoringHooks  *MonitoringHooks
	operationMetrics map[string]*OperationMetrics
}

// NewRedisClient creates a new Redis client instance with enhanced features
func NewRedisClient(logger logging.Logger, config RedisConfig) (*Client, error) {
	var opt *redis.Options

	opt, err := parseRedisConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Redis configuration: %w", err)
	}

	if opt == nil {
		return nil, fmt.Errorf("no valid Redis configuration found")
	}

	client := redis.NewClient(opt)
	redisClient := &Client{
		redisClient:     client,
		config:          config,
		logger:          logger,
		retryConfig:     DefaultRetryConfig(),
		recoveryConfig:  DefaultConnectionRecoveryConfig(),
		isRecovering:    false,
		lastHealthCheck: time.Now(),
	}

	// Use a background context for the initial check. The timeout is handled by the client's config.
	if err := redisClient.CheckConnection(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	// Start connection recovery goroutine if enabled
	if redisClient.recoveryConfig.Enabled {
		go redisClient.connectionRecoveryLoop()
	}

	logger.Infof("Successfully connected to Redis")
	return redisClient, nil
}

// parseRedisConfig parses Redis configuration for both Upstash and local Redis
func parseRedisConfig(config RedisConfig) (*redis.Options, error) {
	var opt *redis.Options
	var err error

	opt, err = redis.ParseURL(config.UpstashConfig.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Upstash Redis URL: %w", err)
	}

	if config.UpstashConfig.Token != "" {
		opt.Password = config.UpstashConfig.Token
	}

	applyConnectionSettings(opt, config)
	return opt, nil
}

// applyConnectionSettings applies common connection settings
func applyConnectionSettings(opt *redis.Options, config RedisConfig) {
	opt.PoolSize = config.ConnectionSettings.PoolSize
	opt.MinIdleConns = config.ConnectionSettings.MinIdleConns
	opt.MaxRetries = config.ConnectionSettings.MaxRetries
	opt.DialTimeout = config.ConnectionSettings.DialTimeout
	opt.ReadTimeout = config.ConnectionSettings.ReadTimeout
	opt.WriteTimeout = config.ConnectionSettings.WriteTimeout
	opt.PoolTimeout = config.ConnectionSettings.PoolTimeout
}

// CheckConnection validates the Redis connection.
// It now uses the passed context and relies on the client's Read/Write timeouts.
func (c *Client) CheckConnection(ctx context.Context) error {
	return c.executeWithRetry(ctx, func() error {
		_, err := c.redisClient.Ping(ctx).Result()
		if err != nil {
			c.logger.Errorf("Redis connection failed: %v", err)
			return fmt.Errorf("redis connection failed: %w", err)
		}
		return nil
	}, "CheckConnection")
}

// Ping checks if Redis is reachable.
// It now uses a background context, relying on the client's configured Read/Write timeouts
// instead of creating its own timeout context.
func (c *Client) Ping(ctx context.Context) error {
	start := time.Now()
	// The timeout for the actual Ping command is now governed by the client's ReadTimeout.
	err := c.executeWithRetry(ctx, func() error {
		return c.redisClient.Ping(ctx).Err()
	}, "Ping")

	// Track connection status
	latency := time.Since(start)
	c.trackConnectionStatus(err == nil, latency)

	return err
}

// ExecutePipeline executes a series of commands in a single network round-trip.
// It wraps the entire pipeline execution within the client's retry logic,
// making batch operations resilient.
// The provided PipelineFunc is responsible for queuing commands on the pipeline.
func (c *Client) ExecutePipeline(ctx context.Context, fn PipelineFunc) ([]redis.Cmder, error) {
	var cmds []redis.Cmder

	// Wrap the entire pipeline execution in our robust retry mechanism.
	// If the connection drops, it will retry the entire batch of commands.
	err := c.executeWithRetry(ctx, func() error {
		pipe := c.redisClient.Pipeline()

		// The user's function queues the commands.
		if err := fn(pipe); err != nil {
			// This is a non-retryable error from the user's logic, not a Redis error.
			return fmt.Errorf("pipeline setup function failed: %w", err)
		}

		// Exec sends all queued commands to Redis and returns their results.
		var execErr error
		cmds, execErr = pipe.Exec(ctx)
		// This is the error we check for retry eligibility (e.g., network issues).
		return execErr
	}, "ExecutePipeline")

	return cmds, err
}

// Core Redis operations with retry logic and monitoring
func (c *Client) Get(ctx context.Context, key string) (string, error) {
	var result string
	err := c.executeWithRetryAndKey(ctx, func() error {
		val, err := c.redisClient.Get(ctx, key).Result()
		if err == redis.Nil {
			return redis.Nil
		}
		if err != nil {
			return err
		}
		result = val
		return nil
	}, "Get", key)
	return result, err
}

// GetWithExists returns both the value and existence status for a key
// This leverages the GET command's return value to provide existence information
// without requiring a separate EXISTS check
func (c *Client) GetWithExists(ctx context.Context, key string) (value string, exists bool, err error) {
	err = c.executeWithRetryAndKey(ctx, func() error {
		val, err := c.redisClient.Get(ctx, key).Result()
		if err == redis.Nil {
			// Key does not exist
			value = ""
			exists = false
			return nil
		}
		if err != nil {
			return err
		}
		// Key exists
		value = val
		exists = true
		return nil
	}, "GetWithExists", key)
	return value, exists, err
}

func (c *Client) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return c.executeWithRetryAndKey(ctx, func() error {
		return c.redisClient.Set(ctx, key, value, expiration).Err()
	}, "Set", key)
}

func (c *Client) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	var result bool
	err := c.executeWithRetry(ctx, func() error {
		val, err := c.redisClient.SetNX(ctx, key, value, expiration).Result()
		if err != nil {
			return err
		}
		result = val
		return nil
	}, "SetNX")
	return result, err
}

func (c *Client) Del(ctx context.Context, keys ...string) error {
	return c.executeWithRetry(ctx, func() error {
		return c.redisClient.Del(ctx, keys...).Err()
	}, "Del")
}

// DelWithCount deletes keys and returns the number of keys that were deleted
// This leverages the DEL command's return value to provide deletion count information
// without requiring separate EXISTS checks
func (c *Client) DelWithCount(ctx context.Context, keys ...string) (deletedCount int64, err error) {
	err = c.executeWithRetry(ctx, func() error {
		count, err := c.redisClient.Del(ctx, keys...).Result()
		if err != nil {
			return err
		}
		deletedCount = count
		return nil
	}, "DelWithCount")
	return deletedCount, err
}

// TTL returns the time-to-live for a key with retry logic
func (c *Client) TTL(ctx context.Context, key string) (time.Duration, error) {
	var result time.Duration
	err := c.executeWithRetry(ctx, func() error {
		val, err := c.redisClient.TTL(ctx, key).Result()
		if err != nil {
			return err
		}
		result = val
		return nil
	}, "TTL")
	return result, err
}

// Eval executes a Lua script on the server.
func (c *Client) Eval(ctx context.Context, script string, keys []string, args ...interface{}) (interface{}, error) {
	var result interface{}
	err := c.executeWithRetry(ctx, func() error {
		val, err := c.redisClient.Eval(ctx, script, keys, args...).Result()
		if err != nil {
			return err
		}
		result = val
		return nil
	}, "Eval")
	return result, err
}

// Client returns the underlying Redis client for advanced operations like pipelines
func (c *Client) Client() *redis.Client {
	return c.redisClient
}

func (c *Client) Close() error {
	return c.redisClient.Close()
}
