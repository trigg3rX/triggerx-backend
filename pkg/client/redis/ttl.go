package redis

import (
	"context"
	"time"
)

// RefreshTTL refreshes the TTL for a key if it exists
// Optimized to use EXPIRE XX option to reduce network round trips
func (c *Client) RefreshTTL(ctx context.Context, key string, ttl time.Duration) error {
	return c.executeWithRetry(ctx, func() error {
		// Use EXPIRE XX to only set expiry if the key exists
		// This eliminates the need for a separate EXISTS check
		result, err := c.redisClient.ExpireXX(ctx, key, ttl).Result()
		if err != nil {
			return err
		}

		// If result is false, the key doesn't exist, which is not an error
		// for this operation (we only want to refresh TTL if it exists)
		if !result {
			// Key doesn't exist, which is fine for refresh operations
			return nil
		}

		return nil
	}, "RefreshTTL")
}

// RefreshStreamTTL refreshes the TTL for a stream and ensures it stays active
// Optimized to use EXPIRE XX option to reduce network round trips
func (c *Client) RefreshStreamTTL(ctx context.Context, stream string, ttl time.Duration) error {
	return c.executeWithRetry(ctx, func() error {
		// Use EXPIRE XX to only set expiry if the stream exists
		// This eliminates the need for a separate EXISTS check
		result, err := c.redisClient.ExpireXX(ctx, stream, ttl).Result()
		if err != nil {
			return err
		}

		// If result is false, the stream doesn't exist, which is not an error
		// for this operation (we only want to refresh TTL if it exists)
		if !result {
			// Stream doesn't exist, which is fine for refresh operations
			return nil
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
func (c *Client) GetTTLStatus(ctx context.Context, key string) (ttl time.Duration, exists bool, err error) {
	err = c.executeWithRetry(ctx, func() error {
		// TTL command is sufficient to determine both existence and expiry.
		ttlResult, err := c.redisClient.TTL(ctx, key).Result()
		if err != nil {
			return err
		}

		if ttlResult == -2 {
			// Key does not exist.
			exists = false
			ttl = 0
		} else {
			// Key exists. It may or may not have an expiry.
			exists = true
			ttl = ttlResult
		}
		return nil
	}, "GetTTLStatus")

	return ttl, exists, err
}
