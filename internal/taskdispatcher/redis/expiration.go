package redis

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
)

// getStreamExpirationKey returns the Redis key for tracking expiration of a specific stream
func (c *Client) getStreamExpirationKey(stream string) string {
	return StreamExpirationKeyPrefix + stream
}

// AddTaskToTimeoutTracking adds a task to the timeout tracking sorted set
func (c *Client) AddTaskToTimeoutTracking(ctx context.Context, taskID int64, ttl time.Duration) error {
	start := time.Now()

	// Calculate timeout timestamp
	timeoutTimestamp := float64(time.Now().Add(ttl).Unix())
	taskIDStr := strconv.FormatInt(taskID, 10)

	// Add to sorted set with timeout timestamp as score
	_, err := c.client.ZAdd(ctx, DispatchedTimeoutsKey, redis.Z{
		Score:  timeoutTimestamp,
		Member: taskIDStr,
	})
	duration := time.Since(start)

	if err != nil {
		c.logger.Error(ctx, "Failed to add task to timeout tracking",
			observability.Int64("task_id", taskID),
			observability.Float64("timeout_timestamp", timeoutTimestamp),
			observability.Duration("duration", duration),
			observability.Error(err))
		return fmt.Errorf("failed to add task to timeout tracking: %w", err)
	}

	// Set TTL on the sorted set to ensure it expires
	err = c.client.SetTTL(ctx, DispatchedTimeoutsKey, TaskIndexTTL)
	if err != nil {
		c.logger.Warn(ctx, "Failed to set TTL on timeout tracking",
			observability.Int64("task_id", taskID),
			observability.Error(err))
		// Don't return error as the main operation succeeded
	}

	c.logger.Debug(ctx, "Task added to timeout tracking successfully",
		observability.Int64("task_id", taskID),
		observability.Float64("timeout_timestamp", timeoutTimestamp),
		observability.Duration("duration", duration))

	return nil
}

// RemoveTaskFromTimeoutTracking removes a task from the dispatched timeout tracking sorted set
func (c *Client) RemoveTaskFromTimeoutTracking(ctx context.Context, taskID int64) error {
	start := time.Now()
	taskIDStr := strconv.FormatInt(taskID, 10)

	removed, err := c.client.ZRem(ctx, DispatchedTimeoutsKey, taskIDStr)
	duration := time.Since(start)

	if err != nil {
		c.logger.Error(ctx, "Failed to remove task from timeout tracking",
			observability.Int64("task_id", taskID),
			observability.Duration("duration", duration),
			observability.Error(err))
		return fmt.Errorf("failed to remove task from timeout tracking: %w", err)
	}

	if removed == 0 {
		c.logger.Debug(ctx, "Task timeout entry not found for removal",
			observability.Int64("task_id", taskID),
			observability.Duration("duration", duration))
	} else {
		c.logger.Debug(ctx, "Task timeout removed successfully",
			observability.Int64("task_id", taskID),
			observability.Duration("duration", duration))
	}

	return nil
}

// GetExpiredTasks retrieves all tasks that have exceeded their timeout
func (c *Client) GetExpiredTasks(ctx context.Context) ([]int64, error) {
	start := time.Now()

	currentTimestamp := time.Now().Unix()
	expiredTaskIDs, err := c.client.ZRangeByScore(ctx, DispatchedTimeoutsKey, "0", strconv.FormatInt(currentTimestamp, 10))
	duration := time.Since(start)

	if err != nil {
		c.logger.Error(ctx, "Failed to get expired tasks",
			observability.Int64("current_timestamp", currentTimestamp),
			observability.Duration("duration", duration),
			observability.Error(err))
		return nil, fmt.Errorf("failed to get expired tasks: %w", err)
	}

	var taskIDs []int64
	for _, taskIDStr := range expiredTaskIDs {
		taskID, err := strconv.ParseInt(taskIDStr, 10, 64)
		if err != nil {
			c.logger.Error(ctx, "Failed to parse task ID from timeout tracking",
				observability.String("task_id_str", taskIDStr),
				observability.Error(err))
			continue
		}
		taskIDs = append(taskIDs, taskID)
	}

	return taskIDs, nil
}

// AddStreamEntryExpiration adds a stream entry to the expiration tracking sorted set
func (c *Client) AddStreamEntryExpiration(ctx context.Context, stream, messageID string, ttl time.Duration) error {
	start := time.Now()

	expirationTimestamp := float64(time.Now().Add(ttl).Unix())
	expirationKey := c.getStreamExpirationKey(stream)
	member := fmt.Sprintf("%s:%s", stream, messageID)

	_, err := c.client.ZAdd(ctx, expirationKey, redis.Z{
		Score:  expirationTimestamp,
		Member: member,
	})
	duration := time.Since(start)

	if err != nil {
		c.logger.Error(ctx, "Failed to add stream entry expiration",
			observability.String("stream", stream),
			observability.String("message_id", messageID),
			observability.Duration("duration", duration),
			observability.Error(err))
		return fmt.Errorf("failed to add stream entry expiration: %w", err)
	}

	// Set TTL on the sorted set
	if err := c.client.SetTTL(ctx, expirationKey, ExpirationTrackingTTL); err != nil {
		c.logger.Warn(ctx, "Failed to set TTL on expiration tracking",
			observability.String("stream", stream),
			observability.Error(err))
	}

	c.logger.Debug(ctx, "Stream entry expiration added successfully",
		observability.String("stream", stream),
		observability.String("message_id", messageID),
		observability.Duration("duration", duration),
		observability.Duration("ttl", ttl))

	return nil
}

// RemoveStreamEntryExpiration removes a stream entry from expiration tracking
func (c *Client) RemoveStreamEntryExpiration(ctx context.Context, stream, messageID string) error {
	start := time.Now()

	expirationKey := c.getStreamExpirationKey(stream)
	member := fmt.Sprintf("%s:%s", stream, messageID)

	removed, err := c.client.ZRem(ctx, expirationKey, member)
	duration := time.Since(start)

	if err != nil {
		c.logger.Error(ctx, "Failed to remove stream entry expiration",
			observability.String("stream", stream),
			observability.String("message_id", messageID),
			observability.Duration("duration", duration),
			observability.Error(err))
		return fmt.Errorf("failed to remove stream entry expiration: %w", err)
	}

	if removed == 0 {
		c.logger.Debug(ctx, "Stream entry expiration not found for removal",
			observability.String("stream", stream),
			observability.String("message_id", messageID),
			observability.Duration("duration", duration))
	} else {
		c.logger.Debug(ctx, "Stream entry expiration removed successfully",
			observability.String("stream", stream),
			observability.String("message_id", messageID),
			observability.Duration("duration", duration))
	}

	return nil
}

// BatchRemoveExpirations removes multiple entries from expiration tracking in batch using pipeline
func (c *Client) BatchRemoveExpirations(ctx context.Context, stream string, messageIDs []string) error {
	if len(messageIDs) == 0 {
		return nil
	}

	start := time.Now()
	expirationKey := c.getStreamExpirationKey(stream)

	var members []string
	for _, messageID := range messageIDs {
		members = append(members, fmt.Sprintf("%s:%s", stream, messageID))
	}

	_, err := c.client.ExecutePipeline(ctx, func(pipe redis.Pipeliner) error {
		for _, member := range members {
			pipe.ZRem(ctx, expirationKey, member)
		}
		return nil
	})
	duration := time.Since(start)

	if err != nil {
		c.logger.Error(ctx, "Failed to batch remove stream entry expirations",
			observability.String("stream", stream),
			observability.Int("count", len(messageIDs)),
			observability.Duration("duration", duration),
			observability.Error(err))
		return fmt.Errorf("failed to batch remove stream entry expirations: %w", err)
	}

	c.logger.Debug(ctx, "Batch removed stream entry expirations successfully",
		observability.String("stream", stream),
		observability.Int("count", len(messageIDs)),
		observability.Duration("duration", duration))

	return nil
}
