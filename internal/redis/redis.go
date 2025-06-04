package redis

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/trigg3rX/triggerx-backend/internal/redis/config"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

// Client represents a Redis client with logging capabilities
type Client struct {
	redisClient *redis.Client
	logger      logging.Logger
}

var (
	redisClient *redis.Client
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

	// Try configurations in priority order: Upstash first, then local
	var opt *redis.Options
	var err error
	var redisType string

	// Priority 1: Try Upstash Redis (cloud)
	if config.IsUpstashEnabled() {
		opt, err = parseUpstashConfig()
		if err == nil {
			redisType = "upstash"
			logger.Infof("Using Upstash Redis configuration")
		} else {
			logger.Warnf("Failed to parse Upstash config: %v", err)
		}
	}

	// Priority 2: Try local Redis if Upstash failed or not available
	if opt == nil && config.IsLocalRedisEnabled() {
		opt, err = parseLocalRedisConfig()
		if err == nil {
			redisType = "local"
			logger.Infof("Using local Redis configuration")
		} else {
			logger.Warnf("Failed to parse local Redis config: %v", err)
		}
	}

	if opt == nil {
		return nil, fmt.Errorf("no valid Redis configuration found")
	}

	client := redis.NewClient(opt)

	redisClient := &Client{
		redisClient: client,
		logger:      logger,
	}

	if err := redisClient.CheckConnection(); err != nil {
		return nil, fmt.Errorf("failed to connect to %s Redis: %w", redisType, err)
	}

	logger.Infof("Successfully connected to %s Redis", redisType)
	return redisClient, nil
}

// parseUpstashConfig parses Upstash Redis configuration
func parseUpstashConfig() (*redis.Options, error) {
	if !config.IsUpstashEnabled() {
		return nil, fmt.Errorf("upstash not enabled")
	}

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

// parseLocalRedisConfig parses local Redis configuration
func parseLocalRedisConfig() (*redis.Options, error) {
	if !config.IsLocalRedisEnabled() {
		return nil, fmt.Errorf("local Redis not enabled")
	}

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
		"job_stream_ttl": config.GetJobStreamTTL().String(),
		"task_stream_ttl": config.GetTaskStreamTTL().String(),
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
