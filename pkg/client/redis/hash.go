package redis

import (
	"context"

	redis "github.com/redis/go-redis/v9"
)

// Hash operations with retry logic and monitoring
func (c *Client) HSet(ctx context.Context, key string, values ...interface{}) error {
	return c.executeWithRetryAndKey(ctx, func() error {
		return c.redisClient.HSet(ctx, key, values...).Err()
	}, "HSet", key)
}

func (c *Client) HGet(ctx context.Context, key, field string) (string, error) {
	var result string
	err := c.executeWithRetryAndKey(ctx, func() error {
		val, err := c.redisClient.HGet(ctx, key, field).Result()
		if err == redis.Nil {
			return redis.Nil
		}
		if err != nil {
			return err
		}
		result = val
		return nil
	}, "HGet", key)
	return result, err
}

// HGetWithExists returns both the value and existence status for a hash field
// This leverages the HGET command's return value to provide existence information
// without requiring a separate HEXISTS check
func (c *Client) HGetWithExists(ctx context.Context, key, field string) (value string, exists bool, err error) {
	err = c.executeWithRetryAndKey(ctx, func() error {
		val, err := c.redisClient.HGet(ctx, key, field).Result()
		if err == redis.Nil {
			// Field does not exist
			value = ""
			exists = false
			return nil
		}
		if err != nil {
			return err
		}
		// Field exists
		value = val
		exists = true
		return nil
	}, "HGetWithExists", key)
	return value, exists, err
}

func (c *Client) HDel(ctx context.Context, key string, fields ...string) error {
	return c.executeWithRetryAndKey(ctx, func() error {
		return c.redisClient.HDel(ctx, key, fields...).Err()
	}, "HDel", key)
}

// HDelWithCount deletes hash fields and returns the number of fields that were deleted
// This leverages the HDEL command's return value to provide deletion count information
// without requiring separate HEXISTS checks
func (c *Client) HDelWithCount(ctx context.Context, key string, fields ...string) (deletedCount int64, err error) {
	err = c.executeWithRetryAndKey(ctx, func() error {
		count, err := c.redisClient.HDel(ctx, key, fields...).Result()
		if err != nil {
			return err
		}
		deletedCount = count
		return nil
	}, "HDelWithCount", key)
	return deletedCount, err
}