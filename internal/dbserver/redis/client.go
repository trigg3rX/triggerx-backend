package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/trigg3rX/triggerx-backend-imua/internal/dbserver/config"
	"github.com/trigg3rX/triggerx-backend-imua/pkg/logging"
)

type Client struct {
	client *redis.Client
	logger logging.Logger
}

func NewClient(logger logging.Logger) (*Client, error) {
	opt, err := redis.ParseURL(config.GetUpstashRedisUrl())
	if err != nil {
		return nil, fmt.Errorf("failed to parse Redis URL: %w", err)
	}

	if config.GetUpstashRedisRestToken() != "" {
		opt.Password = config.GetUpstashRedisRestToken()
	}

	client := redis.NewClient(opt)

	redisClient := &Client{
		client: client,
		logger: logger,
	}

	if err := redisClient.CheckConnection(); err != nil {
		return nil, err
	}

	return redisClient, nil
}

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

func (c *Client) Get(ctx context.Context, key string) (string, error) {
	val, err := c.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", nil
	} else if err != nil {
		return "", err
	}
	return val, nil
}

func (c *Client) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return c.client.Set(ctx, key, value, expiration).Err()
}

func (c *Client) Del(ctx context.Context, keys ...string) error {
	return c.client.Del(ctx, keys...).Err()
}

func (c *Client) Incr(ctx context.Context, key string) (int64, error) {
	return c.client.Incr(ctx, key).Result()
}

func (c *Client) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return c.client.Expire(ctx, key, expiration).Err()
}

func (c *Client) TTL(ctx context.Context, key string) (time.Duration, error) {
	return c.client.TTL(ctx, key).Result()
}

func (c *Client) Eval(ctx context.Context, script string, keys []string, args ...interface{}) (interface{}, error) {
	return c.client.Eval(ctx, script, keys, args...).Result()
}

func (c *Client) EvalScript(ctx context.Context, script string, keys []string, args []interface{}) (interface{}, error) {
	return c.client.Eval(ctx, script, keys, args...).Result()
}

func (c *Client) Client() *redis.Client {
	return c.client
}

func (c *Client) Close() error {
	return c.client.Close()
}
