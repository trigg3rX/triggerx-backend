package redis

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// getStreamExpirationKey returns the Redis key for tracking expiration of a specific stream
func (c *Client) getStreamExpirationKey(stream string) string {
	return StreamExpirationKeyPrefix + stream
}

// AddTaskToTimeoutTracking adds a task to the executed timeout tracking sorted set
func (c *Client) AddTaskToTimeoutTracking(ctx context.Context, taskID int64, timeoutDuration time.Duration) error {
	start := time.Now()

	timeoutTimestamp := float64(time.Now().Add(timeoutDuration).Unix())
	taskIDStr := strconv.FormatInt(taskID, 10)

	_, err := c.client.ZAdd(ctx, ExecutedTimeoutsKey, redis.Z{
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

	// Set TTL on the sorted set
	if err := c.client.SetTTL(ctx, ExecutedTimeoutsKey, ExpirationTrackingTTL); err != nil {
		c.logger.Warn(ctx, "Failed to set TTL on timeout tracking",
			observability.Int64("task_id", taskID),
			observability.Error(err))
	}

	c.logger.Debug(ctx, "Task added to timeout tracking successfully",
		observability.Int64("task_id", taskID),
		observability.Float64("timeout_timestamp", timeoutTimestamp),
		observability.Duration("duration", duration))

	return nil
}

// RemoveTaskFromTimeoutTracking removes a task from the timeout tracking sorted set
func (c *Client) RemoveTaskFromTimeoutTracking(ctx context.Context, taskID int64) error {
	start := time.Now()
	taskIDStr := strconv.FormatInt(taskID, 10)

	removed, err := c.client.ZRem(ctx, ExecutedTimeoutsKey, taskIDStr)
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
	expiredTaskIDs, err := c.client.ZRangeByScore(ctx, ExecutedTimeoutsKey, "0", strconv.FormatInt(currentTimestamp, 10))
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

// RemoveMultipleTaskTimeouts removes multiple tasks from timeout tracking in batch
func (c *Client) RemoveMultipleTaskTimeouts(ctx context.Context, taskIDs []int64) error {
	if len(taskIDs) == 0 {
		return nil
	}

	start := time.Now()

	var taskIDStrs []string
	for _, taskID := range taskIDs {
		taskIDStrs = append(taskIDStrs, strconv.FormatInt(taskID, 10))
	}

	_, err := c.client.ExecutePipeline(ctx, func(pipe redis.Pipeliner) error {
		for _, taskIDStr := range taskIDStrs {
			pipe.ZRem(ctx, ExecutedTimeoutsKey, taskIDStr)
		}
		return nil
	})
	duration := time.Since(start)

	if err != nil {
		c.logger.Error(ctx, "Failed to remove multiple task timeouts",
			observability.Int("task_count", len(taskIDs)),
			observability.Duration("duration", duration),
			observability.Error(err))
		return fmt.Errorf("failed to remove multiple task timeouts: %w", err)
	}

	c.logger.Debug(ctx, "Multiple task timeouts removed successfully",
		observability.Int("task_count", len(taskIDs)),
		observability.Duration("duration", duration))

	return nil
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

// GetExpiredStreamEntries retrieves all expired entries for a specific stream
func (c *Client) GetExpiredStreamEntries(ctx context.Context, stream string) ([]string, error) {
	start := time.Now()

	currentTimestamp := time.Now().Unix()
	expirationKey := c.getStreamExpirationKey(stream)

	expiredMembers, err := c.client.ZRangeByScore(ctx, expirationKey, "0", strconv.FormatInt(currentTimestamp, 10))
	duration := time.Since(start)

	if err != nil {
		c.logger.Error(ctx, "Failed to get expired stream entries",
			observability.String("stream", stream),
			observability.Duration("duration", duration),
			observability.Error(err))
		return nil, fmt.Errorf("failed to get expired stream entries: %w", err)
	}

	// Extract message IDs from members (format: "stream:messageID")
	var messageIDs []string
	streamPrefix := stream + ":"
	for _, member := range expiredMembers {
		if strings.HasPrefix(member, streamPrefix) {
			messageID := member[len(streamPrefix):]
			messageIDs = append(messageIDs, messageID)
		}
	}

	return messageIDs, nil
}

// GetExpiredEntriesForAllStreams retrieves expired entries for all tracked streams
func (c *Client) GetExpiredEntriesForAllStreams(ctx context.Context) (map[string][]string, error) {
	streams := []string{
		types.StreamTaskDispatched,
		types.StreamTaskExecuted,
		types.StreamTaskValidated,
		types.StreamTaskFailed,
		types.StreamTaskRetry,
	}

	expiredEntries := make(map[string][]string)
	for _, stream := range streams {
		messageIDs, err := c.GetExpiredStreamEntries(ctx, stream)
		if err != nil {
			c.logger.Warn(ctx, "Failed to get expired entries for stream",
				observability.String("stream", stream),
				observability.Error(err))
			continue
		}
		if len(messageIDs) > 0 {
			expiredEntries[stream] = messageIDs
		}
	}

	return expiredEntries, nil
}

// RemoveMultipleStreamExpirations removes multiple entries from expiration tracking in batch
func (c *Client) RemoveMultipleStreamExpirations(ctx context.Context, stream string, messageIDs []string) error {
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
		c.logger.Error(ctx, "Failed to remove multiple stream entry expirations",
			observability.String("stream", stream),
			observability.Int("count", len(messageIDs)),
			observability.Duration("duration", duration),
			observability.Error(err))
		return fmt.Errorf("failed to remove multiple stream entry expirations: %w", err)
	}

	c.logger.Debug(ctx, "Multiple stream entry expirations removed successfully",
		observability.String("stream", stream),
		observability.Int("count", len(messageIDs)),
		observability.Duration("duration", duration))

	return nil
}
