package redis

import (
	"context"

	redis "github.com/redis/go-redis/v9"
)

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

func (c *Client) XPending(ctx context.Context, stream, group string) (*redis.XPending, error) {
	var result *redis.XPending
	err := c.executeWithRetry(ctx, func() error {
		val, err := c.redisClient.XPending(ctx, stream, group).Result()
		if err != nil {
			return err
		}
		result = val
		return nil
	}, "XPending")
	return result, err
}

func (c *Client) XPendingExt(ctx context.Context, args *redis.XPendingExtArgs) ([]redis.XPendingExt, error) {
	var result []redis.XPendingExt
	err := c.executeWithRetry(ctx, func() error {
		val, err := c.redisClient.XPendingExt(ctx, args).Result()
		if err != nil {
			return err
		}
		result = val
		return nil
	}, "XPendingExt")
	return result, err
}

func (c *Client) XClaim(ctx context.Context, args *redis.XClaimArgs) *redis.XMessageSliceCmd {
	return c.redisClient.XClaim(ctx, args)
}
