package redis

import (
	"context"
	"time"
)

// RefreshTTL refreshes the TTL for a key if it exists
func (c *Client) RefreshTTL(ctx context.Context, key string, ttl time.Duration) error {
	return c.executeWithRetry(ctx, func() error {
		exists, err := c.redisClient.Exists(ctx, key).Result()
		if err != nil {
			return err
		}
		if exists > 0 {
			return c.redisClient.Expire(ctx, key, ttl).Err()
		}
		return nil
	}, "RefreshTTL")
}

// RefreshStreamTTL refreshes the TTL for a stream and ensures it stays active
func (c *Client) RefreshStreamTTL(ctx context.Context, stream string, ttl time.Duration) error {
	return c.executeWithRetry(ctx, func() error {
		exists, err := c.redisClient.Exists(ctx, stream).Result()
		if err != nil {
			return err
		}
		if exists > 0 {
			return c.redisClient.Expire(ctx, stream, ttl).Err()
		}
		return nil
	}, "RefreshStreamTTL")
}

// SetTTL sets TTL for a key
func (c *Client) SetTTL(ctx context.Context, key string, ttl time.Duration) error {
	return c.executeWithRetry(ctx, func() error {
		return c.redisClient.Expire(ctx, key, ttl).Err()
	}, "SetTTL")
}

// GetTTLStatus returns TTL status for a key
func (c *Client) GetTTLStatus(ctx context.Context, key string) (time.Duration, bool, error) {
	var ttl time.Duration
	var exists bool
	err := c.executeWithRetry(ctx, func() error {
		existsResult, err := c.redisClient.Exists(ctx, key).Result()
		if err != nil {
			return err
		}
		exists = existsResult > 0

		if exists {
			ttlResult, err := c.redisClient.TTL(ctx, key).Result()
			if err != nil {
				return err
			}
			ttl = ttlResult
		}
		return nil
	}, "GetTTLStatus")

	return ttl, exists, err
}
