package redis

import (
	"context"
	"time"

	redis "github.com/redis/go-redis/v9"
)

func (c *Client) CreateStreamIfNotExists(ctx context.Context, stream string, ttl time.Duration) error {
	// This Lua script checks for the stream's existence. If it doesn't exist,
	// it creates it with an initial message and sets its TTL.
	// It returns 1 if it created the stream, and 0 if it already existed.
	// The use of PEXPIRE ensures we can handle TTLs with millisecond precision if needed.
	script := `
	if redis.call("exists", KEYS[1]) == 0 then
		redis.call("xadd", KEYS[1], "*", "init", "stream_created")
		redis.call("pexpire", KEYS[1], ARGV[1])
		return 1
	else
		return 0
	end`

	// TTL needs to be in milliseconds for PEXPIRE.
	ttlMs := ttl.Milliseconds()

	return c.executeWithRetry(ctx, func() error {
		// Eval runs the script atomically on the server.
		_, err := c.redisClient.Eval(ctx, script, []string{stream}, ttlMs).Result()
		return err
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

// CreateConsumerGroupAtomic creates a consumer group atomically using Lua script
// This handles the "already exists" case more efficiently and provides better error handling
func (c *Client) CreateConsumerGroupAtomic(ctx context.Context, stream, group string) (created bool, err error) {
	err = c.executeWithRetry(ctx, func() error {
		err := c.redisClient.XGroupCreateMkStream(ctx, stream, group, "0").Err()
		if err != nil {
			if err.Error() == "BUSYGROUP Consumer Group name already exists" {
				created = false
				return nil
			}
			return err
		}
		created = true
		return nil
	}, "CreateConsumerGroupAtomic")

	return created, err
}

// CreateStreamWithConsumerGroup creates a stream and consumer group in a single operation
// This is more efficient than calling CreateStreamIfNotExists + CreateConsumerGroup separately
func (c *Client) CreateStreamWithConsumerGroup(ctx context.Context, stream, group string, ttl time.Duration) error {
	return c.executeWithRetry(ctx, func() error {
		// Use pipeline to combine all operations into a single network round trip
		pipe := c.redisClient.Pipeline()

		// Create stream and consumer group in one operation
		groupCmd := pipe.XGroupCreateMkStream(ctx, stream, group, "0")

		// Add EXPIRE command if TTL is specified
		var expireCmd *redis.BoolCmd
		if ttl > 0 {
			expireCmd = pipe.Expire(ctx, stream, ttl)
		}

		// Execute all commands in a single network round trip
		_, err := pipe.Exec(ctx)
		if err != nil {
			// Check if it's just a "group already exists" error, which is not a problem
			if err.Error() == "BUSYGROUP Consumer Group name already exists" {
				// Group exists, but we still need to set TTL if specified
				if ttl > 0 {
					return c.redisClient.Expire(ctx, stream, ttl).Err()
				}
				return nil
			}
			return err
		}

		// Check if group creation was successful
		if err := groupCmd.Err(); err != nil {
			return err
		}

		// Check if EXPIRE was successful (if it was executed)
		if expireCmd != nil {
			if _, err := expireCmd.Result(); err != nil {
				return err
			}
		}

		return nil
	}, "CreateStreamWithConsumerGroup")
}
