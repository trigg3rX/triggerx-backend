package redis

import (
	"context"

	redis "github.com/redis/go-redis/v9"
)

// ZAdd adds one or more members to a sorted set, or updates its score if it already exists
func (c *Client) ZAdd(ctx context.Context, key string, members ...redis.Z) (int64, error) {
	var result int64
	err := c.executeWithRetryAndKey(ctx, func() error {
		val, err := c.redisClient.ZAdd(ctx, key, members...).Result()
		if err != nil {
			return err
		}
		result = val
		return nil
	}, "ZAdd", key)
	return result, err
}

// ZAddWithExists adds members to a sorted set and returns both the number of new elements added
// and whether the key existed before the operation
func (c *Client) ZAddWithExists(ctx context.Context, key string, members ...redis.Z) (newElements int64, keyExisted bool, err error) {
	err = c.executeWithRetryAndKey(ctx, func() error {
		// First check if the key exists
		exists, err := c.redisClient.Exists(ctx, key).Result()
		if err != nil {
			return err
		}
		keyExisted = exists > 0

		// Add the members
		val, err := c.redisClient.ZAdd(ctx, key, members...).Result()
		if err != nil {
			return err
		}
		newElements = val
		return nil
	}, "ZAddWithExists", key)
	return newElements, keyExisted, err
}

// ZRevRange returns a range of members in a sorted set, by index, with scores ordered from high to low
func (c *Client) ZRevRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	var result []string
	err := c.executeWithRetryAndKey(ctx, func() error {
		val, err := c.redisClient.ZRevRange(ctx, key, start, stop).Result()
		if err != nil {
			return err
		}
		result = val
		return nil
	}, "ZRevRange", key)
	return result, err
}

// ZRangeByScore returns a range of members in a sorted set, by score
func (c *Client) ZRangeByScore(ctx context.Context, key, min, max string) ([]string, error) {
	var result []string
	err := c.executeWithRetryAndKey(ctx, func() error {
		val, err := c.redisClient.ZRangeByScore(ctx, key, &redis.ZRangeBy{
			Min: min,
			Max: max,
		}).Result()
		if err != nil {
			return err
		}
		result = val
		return nil
	}, "ZRangeByScore", key)
	return result, err
}

// ZRemRangeByScore removes all members in a sorted set within the given scores
func (c *Client) ZRemRangeByScore(ctx context.Context, key, min, max string) (int64, error) {
	var result int64
	err := c.executeWithRetryAndKey(ctx, func() error {
		val, err := c.redisClient.ZRemRangeByScore(ctx, key, min, max).Result()
		if err != nil {
			return err
		}
		result = val
		return nil
	}, "ZRemRangeByScore", key)
	return result, err
}

// ZCard returns the sorted set cardinality (number of elements) of the sorted set stored at key
func (c *Client) ZCard(ctx context.Context, key string) (int64, error) {
	var result int64
	err := c.executeWithRetryAndKey(ctx, func() error {
		val, err := c.redisClient.ZCard(ctx, key).Result()
		if err != nil {
			return err
		}
		result = val
		return nil
	}, "ZCard", key)
	return result, err
}
