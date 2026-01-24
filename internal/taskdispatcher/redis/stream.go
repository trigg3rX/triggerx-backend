package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// InitializeStreams creates all task streams with their TTLs
func (c *Client) InitializeStreams(ctx context.Context) error {
	streamConfigs := map[string]time.Duration{
		types.StreamTaskDispatched: types.TasksProcessingTTL,
		types.StreamTaskCompleted:  types.TasksCompletedTTL,
		types.StreamTaskFailed:     types.TasksFailedTTL,
		types.StreamTaskRetry:      types.TasksRetryTTL,
	}

	for stream, ttl := range streamConfigs {
		c.logger.Debug(ctx, "Creating stream", observability.String("stream", stream), observability.Duration("ttl", ttl))
		if err := c.client.CreateStreamIfNotExists(ctx, stream, ttl); err != nil {
			c.logger.Error(ctx, "Failed to initialize stream",
				observability.String("stream", stream),
				observability.Error(err),
				observability.Duration("ttl", ttl))
			return fmt.Errorf("failed to initialize stream %s: %w", stream, err)
		}
		c.logger.Debug(ctx, "Stream initialized successfully", observability.String("stream", stream), observability.Duration("ttl", ttl))
	}

	c.logger.Info(ctx, "All task streams initialized successfully")
	return nil
}

// InitializeConsumerGroups registers all consumer groups for task processing
func (c *Client) InitializeConsumerGroups(ctx context.Context) error {
	// Consumer group configurations: stream -> list of groups
	consumerGroupConfigs := map[string][]string{
		types.StreamTaskDispatched: {"task-processors"},
		types.StreamTaskCompleted:  {"task-processors"},
		types.StreamTaskFailed:     {"task-processors"},
		types.StreamTaskRetry:      {"task-processors"},
	}

	for stream, groups := range consumerGroupConfigs {
		for _, group := range groups {
			c.logger.Debug(ctx, "Registering consumer group", observability.String("stream", stream), observability.String("group", group))
			if err := c.client.CreateConsumerGroup(ctx, stream, group); err != nil {
				c.logger.Error(ctx, "Failed to create consumer group",
					observability.String("stream", stream),
					observability.String("group", group),
					observability.Error(err))
				return fmt.Errorf("failed to create consumer group %s for %s: %w", group, stream, err)
			}
			c.logger.Debug(ctx, "Consumer group created successfully", observability.String("stream", stream), observability.String("group", group))
		}
	}

	c.logger.Info(ctx, "All consumer groups initialized successfully")
	return nil
}

// CreateStreamIfNotExists creates a Redis stream if it doesn't exist
func (c *Client) CreateStreamIfNotExists(ctx context.Context, stream string, ttl time.Duration) error {
	return c.client.CreateStreamIfNotExists(ctx, stream, ttl)
}

// CreateConsumerGroup creates a consumer group for a stream
func (c *Client) CreateConsumerGroup(ctx context.Context, stream string, group string) error {
	return c.client.CreateConsumerGroup(ctx, stream, group)
}

// XLen returns the length of a stream
func (c *Client) XLen(ctx context.Context, stream string) (int64, error) {
	return c.client.XLen(ctx, stream)
}

// XAdd adds an entry to a stream
func (c *Client) XAdd(ctx context.Context, args *redis.XAddArgs) (string, error) {
	return c.client.XAdd(ctx, args)
}

// XRange returns entries from a stream in a range
func (c *Client) XRange(ctx context.Context, stream, start, stop string) ([]redis.XMessage, error) {
	return c.client.Client().XRange(ctx, stream, start, stop).Result()
}

// GetStreamLength returns the length of a stream
func (c *Client) GetStreamLength(ctx context.Context, stream string) (int64, error) {
	return c.client.XLen(ctx, stream)
}

// AcknowledgeTask acknowledges a message in a stream consumer group
func (c *Client) AcknowledgeTask(ctx context.Context, stream, group, messageID string) error {
	err := c.client.XAck(ctx, stream, group, messageID)
	if err != nil {
		c.logger.Error(ctx, "Failed to acknowledge task",
			observability.String("stream", stream),
			observability.String("group", group),
			observability.String("message_id", messageID),
			observability.Error(err))
		return fmt.Errorf("failed to acknowledge task: %w", err)
	}

	c.logger.Debug(ctx, "Task acknowledged successfully",
		observability.String("stream", stream),
		observability.String("group", group),
		observability.String("message_id", messageID))

	return nil
}

// DeleteTaskFromStream deletes a message from a stream
func (c *Client) DeleteTaskFromStream(ctx context.Context, stream, messageID string) (int64, error) {
	deleted, err := c.client.XDel(ctx, stream, messageID)
	if err != nil {
		c.logger.Error(ctx, "Failed to delete message from stream",
			observability.String("stream", stream),
			observability.String("message_id", messageID),
			observability.Error(err))
		return 0, fmt.Errorf("failed to delete message from stream: %w", err)
	}

	c.logger.Debug(ctx, "Message deleted from stream",
		observability.String("stream", stream),
		observability.String("message_id", messageID),
		observability.Int64("deleted_count", deleted))

	return deleted, nil
}

// TrimStream trims a stream to the specified max length
func (c *Client) TrimStream(ctx context.Context, stream string, maxLen int64, approx bool) (int64, error) {
	trimmed, err := c.client.XTrim(ctx, stream, maxLen, approx)
	if err != nil {
		c.logger.Error(ctx, "Failed to trim stream",
			observability.String("stream", stream),
			observability.Int64("max_len", maxLen),
			observability.Error(err))
		return 0, fmt.Errorf("failed to trim stream: %w", err)
	}

	if trimmed > 0 {
		c.logger.Debug(ctx, "Stream trimmed successfully",
			observability.String("stream", stream),
			observability.Int64("trimmed_count", trimmed),
			observability.Int64("max_len", maxLen))
	}

	return trimmed, nil
}

// GetTaskFromStream retrieves a task from a stream by taskID
func (c *Client) GetTaskFromStream(ctx context.Context, stream string, taskID int64) (*types.TaskStreamData, error) {
	// Use XRANGE to read recent messages without adding to PEL
	streams, err := c.client.Client().XRange(ctx, stream, "-", "+").Result()
	if err != nil {
		c.logger.Error(ctx, "Failed to read task stream data",
			observability.Int64("task_id", taskID),
			observability.Error(err))
		return nil, fmt.Errorf("failed to read task stream data: %w", err)
	}

	// Limit the search to the most recent 500 messages to avoid performance issues
	startIndex := 0
	if len(streams) > 500 {
		startIndex = len(streams) - 500
	}

	for i := startIndex; i < len(streams); i++ {
		message := streams[i]
		taskJSON, exists := message.Values["task"].(string)
		if !exists {
			continue
		}

		var task types.TaskStreamData
		if err := json.Unmarshal([]byte(taskJSON), &task); err != nil {
			c.logger.Error(ctx, "Failed to unmarshal task data",
				observability.String("message_id", message.ID),
				observability.Error(err))
			continue
		}

		if task.SendTaskDataToKeeper.TaskID[0] == taskID {
			return &task, nil
		}
	}

	return nil, fmt.Errorf("task not found: %d", taskID)
}
