package redis

import (
	"context"
	"fmt"
	"sync"
	"time"

	redis "github.com/redis/go-redis/v9" // Add alias here
	"github.com/trigg3rX/triggerx-backend/internal/redis/config"
	"github.com/trigg3rX/triggerx-backend/internal/redis/metrics"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/client/dbserver"
)

// Client represents a Redis client with logging capabilities
type Client struct {
	redisClient *redis.Client
	dbserverClient *dbserver.DBServerClient
	logger      logging.Logger
	mu          sync.Mutex
}

func IsRedisAvailable() bool {
	return config.IsRedisAvailable()
}

// NewRedisClient creates a new Redis client instance with enhanced features
func NewRedisClient(logger logging.Logger) (*Client, error) {
	if !config.IsRedisAvailable() {
		metrics.IsRedisUpstashAvailable.Set(float64(0))
		return nil, fmt.Errorf("redis is not configured")
	}

	var opt *redis.Options
	var isUpstash bool
	var err error

	// Priority 1: Try Upstash Redis (cloud)
	if config.IsUpstashEnabled() {
		opt, err = parseUpstashConfig()
		if err == nil {
			isUpstash = true
			logger.Infof("Using Upstash Redis configuration")
			metrics.IsRedisUpstashAvailable.Set(float64(1))
		} else {
			logger.Warnf("Failed to parse Upstash config: %v", err)
			metrics.ClientConnectionErrorsTotal.WithLabelValues("config_parse").Inc()
		}
	}

	// Priority 2: Try local Redis if Upstash failed or not available
	if opt == nil && config.IsLocalRedisEnabled() {
		opt, err = parseLocalRedisConfig()
		if err == nil {
			isUpstash = false
			logger.Infof("Using local Redis configuration")
			metrics.IsRedisUpstashAvailable.Set(float64(0))
		} else {
			logger.Warnf("Failed to parse local Redis config: %v", err)
			metrics.ClientConnectionErrorsTotal.WithLabelValues("config_parse").Inc()
		}
	}

	if opt == nil {
		metrics.ServiceStatus.WithLabelValues("client").Set(float64(0))
		return nil, fmt.Errorf("no valid Redis configuration found")
	}

	// Initialize dbserver client
	dbserverClient, err := dbserver.NewDBServerClient(logger, config.GetDBServerRPCUrl())
	if err != nil {
		logger.Fatal("Failed to initialize dbserver client", "error", err)
	}
	logger.Info("DBServer client Initialised")

	client := redis.NewClient(opt)
	redisClient := &Client{
		redisClient: client,
		dbserverClient: dbserverClient,
		logger:      logger,
	}

	if err := redisClient.CheckConnection(); err != nil {
		metrics.ClientConnectionsTotal.WithLabelValues("failure").Inc()
		metrics.ClientConnectionErrorsTotal.WithLabelValues("connection_failed").Inc()
		metrics.ServiceStatus.WithLabelValues("client").Set(float64(0))
		redisType := "local"
		if isUpstash {
			redisType = "upstash"
		}
		return nil, fmt.Errorf("failed to connect to %s Redis: %w", redisType, err)
	}

	// Record successful connection
	metrics.ClientConnectionsTotal.WithLabelValues("success").Inc()
	metrics.ServiceStatus.WithLabelValues("client").Set(float64(1))

	redisType := "local"
	if isUpstash {
		redisType = "upstash"
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

	if config.GetUpstashToken() != "" {
		opt.Password = config.GetUpstashToken()
	}

	applyConnectionSettings(opt)
	return opt, nil
}

// parseLocalRedisConfig parses local Redis configuration
func parseLocalRedisConfig() (*redis.Options, error) {
	if !config.IsLocalRedisEnabled() {
		return nil, fmt.Errorf("local Redis not enabled")
	}

	opt := &redis.Options{
		Addr:     config.GetRedisAddr(),
		Password: config.GetRedisPassword(),
		DB:       config.GetRedisDB(),
	}

	applyConnectionSettings(opt)
	return opt, nil
}

// applyConnectionSettings applies common connection settings
func applyConnectionSettings(opt *redis.Options) {
	opt.PoolSize = config.GetPoolSize()
	opt.MinIdleConns = config.GetMinIdleConns()
	opt.MaxRetries = config.GetMaxRetries()
	opt.DialTimeout = config.GetDialTimeout()
	opt.ReadTimeout = config.GetReadTimeout()
	opt.WriteTimeout = config.GetWriteTimeout()
	opt.PoolTimeout = config.GetPoolTimeout()
}

// CheckConnection validates the Redis connection
func (c *Client) CheckConnection() error {
	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := c.redisClient.Ping(ctx).Result()
	duration := time.Since(start)

	if err != nil {
		c.logger.Errorf("Redis connection failed (%s): %v", config.GetRedisType(), err)
		metrics.ConnectionChecksTotal.WithLabelValues("failure").Inc()
		return fmt.Errorf("redis connection failed: %w", err)
	}

	metrics.ConnectionChecksTotal.WithLabelValues("success").Inc()
	metrics.PingOperationsTotal.WithLabelValues("success").Inc()
	metrics.PingDuration.Observe(duration.Seconds())
	return nil
}

// Ping checks if Redis is reachable
func (c *Client) Ping() error {
	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := c.redisClient.Ping(ctx).Err()
	duration := time.Since(start)

	if err != nil {
		metrics.PingOperationsTotal.WithLabelValues("failure").Inc()
		return err
	}

	metrics.PingOperationsTotal.WithLabelValues("success").Inc()
	metrics.PingDuration.Observe(duration.Seconds())
	return nil
}

// GetRedisInfo returns information about the current Redis configuration
func GetRedisInfo() map[string]interface{} {
	return map[string]interface{}{
		"available":       config.IsRedisAvailable(),
		"type":            config.GetRedisType(),
		"upstash":         config.IsUpstashEnabled(),
		"local":           config.IsLocalRedisEnabled(),
		"job_stream_ttl":  config.GetJobStreamTTL().String(),
		"task_stream_ttl": config.GetTaskStreamTTL().String(),
		"stream_maxlen":   config.GetStreamMaxLen(),
	}
}

// Stream management functions
func (c *Client) CreateStreamIfNotExists(ctx context.Context, stream string, ttl time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

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
}

func (c *Client) CreateConsumerGroup(ctx context.Context, stream, group string) error {
	err := c.redisClient.XGroupCreateMkStream(ctx, stream, group, "0").Err()
	if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
		return err
	}
	return nil
}

// Core Redis operations
func (c *Client) Get(ctx context.Context, key string) (string, error) {
	val, err := c.redisClient.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", nil
	}
	return val, err
}

func (c *Client) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return c.redisClient.Set(ctx, key, value, expiration).Err()
}

func (c *Client) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	return c.redisClient.SetNX(ctx, key, value, expiration).Result()
}

func (c *Client) Del(ctx context.Context, keys ...string) error {
	return c.redisClient.Del(ctx, keys...).Err()
}

func (c *Client) XAdd(ctx context.Context, args *redis.XAddArgs) (string, error) {
	return c.redisClient.XAdd(ctx, args).Result()
}

func (c *Client) XLen(ctx context.Context, stream string) (int64, error) {
	return c.redisClient.XLen(ctx, stream).Result()
}

func (c *Client) XReadGroup(ctx context.Context, args *redis.XReadGroupArgs) ([]redis.XStream, error) {
	return c.redisClient.XReadGroup(ctx, args).Result()
}

func (c *Client) XAck(ctx context.Context, stream, group, id string) error {
	return c.redisClient.XAck(ctx, stream, group, id).Err()
}

func (c *Client) Close() error {
	c.dbserverClient.Close()
	return c.redisClient.Close()
}
