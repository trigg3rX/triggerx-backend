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

	if err := redisClient.CheckConnection(); err != nil {
		redisType := "local"
		if config.IsUpstash {
			redisType = "upstash"
		}
		return nil, fmt.Errorf("failed to connect to %s Redis: %w", redisType, err)
	}

	// Start connection recovery goroutine if enabled
	if redisClient.recoveryConfig.Enabled {
		go redisClient.connectionRecoveryLoop()
	}

	redisType := "local"
	if config.IsUpstash {
		redisType = "upstash"
	}
	logger.Infof("Successfully connected to %s Redis", redisType)
	return redisClient, nil
}

// parseRedisConfig parses Redis configuration for both Upstash and local Redis
func parseRedisConfig(config RedisConfig) (*redis.Options, error) {
	var opt *redis.Options
	var err error

	if config.IsUpstash {
		opt, err = redis.ParseURL(config.UpstashConfig.URL)
		if err != nil {
			return nil, fmt.Errorf("failed to parse Upstash Redis URL: %w", err)
		}

		if config.UpstashConfig.Token != "" {
			opt.Password = config.UpstashConfig.Token
		}
	} else {
		opt = &redis.Options{
			Addr:     config.LocalRedisConfig.Addr,
			Password: config.LocalRedisConfig.Password,
			DB:       config.LocalRedisConfig.DB,
		}
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

// GetRedisInfo returns information about the current Redis configuration
func (c *Client) GetRedisInfo() map[string]interface{} {
	return map[string]interface{}{
		"available":            c.config.IsUpstash,
		"type":                 c.config.IsUpstash,
		"upstash":              c.config.IsUpstash,
		"local":                !c.config.IsUpstash,
		"job_stream_ttl":       c.config.StreamsConfig.JobStreamTTL,
		"task_stream_ttl":      c.config.StreamsConfig.TaskStreamTTL,
		"keeper_stream_ttl":    c.config.StreamsConfig.KeeperStreamTTL,
		"registrar_stream_ttl": c.config.StreamsConfig.RegistrarStreamTTL,
		"retry_config":         c.retryConfig,
	}
}

// CheckConnection validates the Redis connection
func (c *Client) CheckConnection() error {
	timeout := c.config.ConnectionSettings.HealthTimeout
	if timeout == 0 {
		timeout = 5 * time.Second
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	return c.executeWithRetry(ctx, func() error {
		_, err := c.redisClient.Ping(ctx).Result()
		if err != nil {
			c.logger.Errorf("Redis connection failed: %v", err)
			return fmt.Errorf("redis connection failed: %w", err)
		}
		return nil
	}, "CheckConnection")
}

// Ping checks if Redis is reachable
func (c *Client) Ping() error {
	timeout := c.config.ConnectionSettings.PingTimeout
	if timeout == 0 {
		timeout = 2 * time.Second
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	start := time.Now()
	err := c.executeWithRetry(ctx, func() error {
		return c.redisClient.Ping(ctx).Err()
	}, "Ping")

	// Track connection status
	latency := time.Since(start)
	c.trackConnectionStatus(err == nil, latency)

	return err
}

// Stream management functions
func (c *Client) CreateStreamIfNotExists(ctx context.Context, stream string, ttl time.Duration) error {
	return c.executeWithRetry(ctx, func() error {
		exists, err := c.redisClient.Exists(ctx, stream).Result()
		if err != nil {
			return err
		}

		if exists == 0 {
			// Create empty stream
			if _, err := c.redisClient.XAdd(ctx, &redis.XAddArgs{
				Stream: stream,
				ID:     "*",
				Values: map[string]interface{}{"init": "stream_created"},
			}).Result(); err != nil {
				return err
			}

			// Set TTL only once at creation
			if err := c.redisClient.Expire(ctx, stream, ttl).Err(); err != nil {
				return err
			}
		}
		return nil
	}, "CreateStreamIfNotExists")
}

func (c *Client) CreateConsumerGroup(ctx context.Context, stream, group string) error {
	return c.executeWithRetry(ctx, func() error {
		err := c.redisClient.XGroupCreateMkStream(ctx, stream, group, "0").Err()
		if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
			return err
		}
		return nil
	}, "CreateConsumerGroup")
}

// Core Redis operations with retry logic and monitoring
func (c *Client) Get(ctx context.Context, key string) (string, error) {
	var result string
	err := c.executeWithRetryAndKey(ctx, func() error {
		val, err := c.redisClient.Get(ctx, key).Result()
		if err == redis.Nil {
			result = ""
			return nil
		}
		if err != nil {
			return err
		}
		result = val
		return nil
	}, "Get", key)
	return result, err
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

func (c *Client) XAdd(ctx context.Context, args *redis.XAddArgs) (string, error) {
	var result string
	err := c.executeWithRetry(ctx, func() error {
		val, err := c.redisClient.XAdd(ctx, args).Result()
		if err != nil {
			return err
		}
		result = val
		return nil
	}, "XAdd")
	return result, err
}

func (c *Client) XLen(ctx context.Context, stream string) (int64, error) {
	var result int64
	err := c.executeWithRetry(ctx, func() error {
		val, err := c.redisClient.XLen(ctx, stream).Result()
		if err != nil {
			return err
		}
		result = val
		return nil
	}, "XLen")
	return result, err
}

func (c *Client) XReadGroup(ctx context.Context, args *redis.XReadGroupArgs) ([]redis.XStream, error) {
	var result []redis.XStream
	err := c.executeWithRetry(ctx, func() error {
		val, err := c.redisClient.XReadGroup(ctx, args).Result()
		if err != nil {
			return err
		}
		result = val
		return nil
	}, "XReadGroup")
	return result, err
}

func (c *Client) XAck(ctx context.Context, stream, group, id string) error {
	return c.executeWithRetry(ctx, func() error {
		return c.redisClient.XAck(ctx, stream, group, id).Err()
	}, "XAck")
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

// Client returns the underlying Redis client for advanced operations like pipelines
func (c *Client) Client() *redis.Client {
	return c.redisClient
}

func (c *Client) Close() error {
	return c.redisClient.Close()
}
