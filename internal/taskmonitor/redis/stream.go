package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/trigg3rX/triggerx-backend/internal/taskmonitor/config"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// InitializeStreams creates all task streams with their TTLs
func (c *Client) InitializeStreams(ctx context.Context) error {
	streamConfigs := map[string]time.Duration{
		types.StreamTaskDispatched: types.TasksProcessingTTL,
		types.StreamTaskExecuted:   types.TasksExecutedTTL,
		types.StreamTaskValidated:  types.TasksValidatedTTL,
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
		types.StreamTaskDispatched: {"task-processors", "task-finder"},
		types.StreamTaskExecuted:   {"task-processors", "timeout-checker", "task-finder"},
		types.StreamTaskValidated:  {"task-processors"},
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

// AddTaskToStream adds a task to a stream and returns the message ID
func (c *Client) AddTaskToStream(ctx context.Context, stream string, task *types.TaskStreamData) (string, error) {
	start := time.Now()

	taskJSON, err := json.Marshal(task)
	if err != nil {
		c.logger.Error(ctx, "Failed to marshal task data",
			observability.Int64("task_id", task.SendTaskDataToKeeper.TaskID[0]),
			observability.String("stream", stream),
			observability.Error(err))
		return "", fmt.Errorf("failed to marshal task data: %w", err)
	}

	messageID, err := c.client.XAdd(ctx, &redis.XAddArgs{
		Stream: stream,
		MaxLen: int64(config.GetStreamMaxLen()),
		Approx: true,
		Values: map[string]interface{}{
			"task":       string(taskJSON),
			"created_at": time.Now().Unix(),
		},
	})
	duration := time.Since(start)

	if err != nil {
		c.logger.Error(ctx, "Failed to add task to stream",
			observability.Int64("task_id", task.SendTaskDataToKeeper.TaskID[0]),
			observability.String("stream", stream),
			observability.Duration("duration", duration),
			observability.Error(err))
		return "", fmt.Errorf("failed to add task to stream: %w", err)
	}

	c.logger.Debug(ctx, "Task added to stream successfully",
		observability.Int64("task_id", task.SendTaskDataToKeeper.TaskID[0]),
		observability.String("stream", stream),
		observability.String("message_id", messageID),
		observability.Duration("duration", duration))

	return messageID, nil
}

// ReadTasksFromStream reads tasks from a stream using consumer groups
func (c *Client) ReadTasksFromStream(ctx context.Context, stream, consumerGroup, consumerName string, count int64, blockDuration time.Duration) ([]types.TaskStreamData, []string, error) {
	start := time.Now()

	streams, err := c.client.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group:    consumerGroup,
		Consumer: consumerName,
		Streams:  []string{stream, ">"},
		Count:    count,
		Block:    blockDuration,
	})

	duration := time.Since(start)

	if err != nil {
		if err == redis.Nil {
			return []types.TaskStreamData{}, []string{}, nil
		}
		c.logger.Error(ctx, "Failed to read from stream",
			observability.String("stream", stream),
			observability.String("consumer_group", consumerGroup),
			observability.Duration("duration", duration),
			observability.Error(err))
		return nil, nil, fmt.Errorf("failed to read from stream: %w", err)
	}

	var tasks []types.TaskStreamData
	var messageIDs []string

	for _, s := range streams {
		for _, message := range s.Messages {
			taskJSON, exists := message.Values["task"].(string)
			if !exists {
				c.logger.Warn(ctx, "Message missing task data",
					observability.String("stream", s.Stream),
					observability.String("message_id", message.ID))
				continue
			}

			var task types.TaskStreamData
			if err := json.Unmarshal([]byte(taskJSON), &task); err != nil {
				c.logger.Error(ctx, "Failed to unmarshal task data",
					observability.String("stream", s.Stream),
					observability.String("message_id", message.ID),
					observability.Error(err))
				continue
			}

			tasks = append(tasks, task)
			messageIDs = append(messageIDs, message.ID)
		}
	}

	c.logger.Debug(ctx, "Tasks read from stream successfully",
		observability.String("stream", stream),
		observability.Int("task_count", len(tasks)),
		observability.Duration("duration", duration))

	return tasks, messageIDs, nil
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

// GetStreamLength returns the length of a stream
func (c *Client) GetStreamLength(ctx context.Context, stream string) (int64, error) {
	return c.client.XLen(ctx, stream)
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

// GetPendingTasks returns pending entries info for a consumer group
func (c *Client) GetPendingTasks(ctx context.Context, stream, group string) (*redis.XPending, error) {
	return c.client.XPending(ctx, stream, group)
}

// GetPendingTasksExt returns detailed pending entries for a consumer group
func (c *Client) GetPendingTasksExt(ctx context.Context, stream, group string, count int64) ([]redis.XPendingExt, error) {
	return c.client.XPendingExt(ctx, &redis.XPendingExtArgs{
		Stream: stream,
		Group:  group,
		Start:  "-",
		End:    "+",
		Count:  count,
	})
}

// GetTaskByMessageID retrieves a specific task by its message ID using XRANGE
func (c *Client) GetTaskByMessageID(ctx context.Context, stream, messageID string) (*types.TaskStreamData, error) {
	start := time.Now()

	messages, err := c.client.Client().XRange(ctx, stream, messageID, messageID).Result()
	if err != nil {
		c.logger.Error(ctx, "Failed to get task by messageID",
			observability.String("stream", stream),
			observability.String("message_id", messageID),
			observability.Error(err))
		return nil, fmt.Errorf("failed to get task by messageID: %w", err)
	}

	if len(messages) == 0 {
		return nil, fmt.Errorf("no message found with ID %s in stream %s", messageID, stream)
	}

	message := messages[0]
	taskJSON, exists := message.Values["task"].(string)
	if !exists {
		c.logger.Error(ctx, "Message missing task data",
			observability.String("stream", stream),
			observability.String("message_id", messageID))
		return nil, fmt.Errorf("message %s missing task data", messageID)
	}

	var task types.TaskStreamData
	if err := json.Unmarshal([]byte(taskJSON), &task); err != nil {
		c.logger.Error(ctx, "Failed to unmarshal task data",
			observability.String("stream", stream),
			observability.String("message_id", messageID),
			observability.Error(err))
		return nil, fmt.Errorf("failed to unmarshal task data: %w", err)
	}

	duration := time.Since(start)
	c.logger.Debug(ctx, "Task retrieved by messageID successfully",
		observability.Int64("task_id", task.SendTaskDataToKeeper.TaskID[0]),
		observability.String("stream", stream),
		observability.String("message_id", messageID),
		observability.Duration("duration", duration))

	return &task, nil
}

// ScanStreamForTask scans a stream for a task by ID (fallback when index is unavailable)
func (c *Client) ScanStreamForTask(ctx context.Context, stream string, taskID int64, limit int) (*types.TaskStreamData, error) {
	messages, err := c.client.Client().XRange(ctx, stream, "-", "+").Result()
	if err != nil {
		return nil, fmt.Errorf("failed to read from stream %s: %w", stream, err)
	}

	// Limit the search to the most recent messages
	startIndex := 0
	if len(messages) > limit {
		startIndex = len(messages) - limit
	}

	for i := startIndex; i < len(messages); i++ {
		message := messages[i]
		taskJSON, exists := message.Values["task"].(string)
		if !exists {
			continue
		}

		var task types.TaskStreamData
		if err := json.Unmarshal([]byte(taskJSON), &task); err != nil {
			continue
		}

		if task.SendTaskDataToKeeper.TaskID[0] == taskID {
			return &task, nil
		}
	}

	return nil, fmt.Errorf("task %d not found in stream %s", taskID, stream)
}
