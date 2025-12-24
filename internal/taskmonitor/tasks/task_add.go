package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/trigg3rX/triggerx-backend/internal/taskmonitor/config"
	"github.com/trigg3rX/triggerx-backend/internal/taskmonitor/metrics"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
)

// MarkTaskCompleted marks a task as completed
func (tsm *TaskStreamManager) MarkTaskCompleted(ctx context.Context, taskID int64) error {
	tsm.logger.Info(ctx, "Marking task as completed",
		observability.Int64("task_id", taskID))

	// Find and move task from processing to completed using efficient lookup
	task, messageID, err := tsm.taskIndex.FindTaskByID(ctx, taskID)
	if err != nil {
		tsm.logger.Error(ctx, "failed to find task in processing", observability.Error(err))
		// return err
	}

	task.CompletedAt = &[]time.Time{time.Now()}[0]

	// Add to completed stream
	err = tsm.addTaskToStream(ctx, StreamTaskCompleted, task)
	if err != nil {
		tsm.logger.Error(ctx, "failed to add to completed stream", observability.Error(err))
		// return err
	}

	// Remove from processing stream (acknowledge) using the messageID
	if messageID != "" {
		err = tsm.AckTaskProcessed(ctx, StreamTaskDispatched, "task-processors", messageID)
		if err != nil {
			tsm.logger.Error(ctx, "failed to acknowledge task",
				observability.Int64("task_id", taskID),
				observability.String("message_id", messageID),
				observability.Error(err))
		} else {
			// Remove the task from the index since it's been processed
			err = tsm.taskIndex.RemoveTaskIndex(ctx, taskID)
			if err != nil {
				tsm.logger.Warn(ctx, "failed to remove task from index",
					observability.Int64("task_id", taskID),
					observability.Error(err))
			}

			// Remove the task from timeout tracking since it's been completed
			err = tsm.expirationManager.RemoveTaskTimeout(ctx, taskID)
			if err != nil {
				tsm.logger.Warn(ctx, "failed to remove task from timeout tracking",
					observability.Int64("task_id", taskID),
					observability.Error(err))
			}
		}
	}

	tsm.logger.Info(ctx, "Task marked as completed successfully", observability.Int64("task_id", taskID))
	if metrics.TasksAddedToStreamTotal != nil {
		metrics.TasksAddedToStreamTotal.WithLabelValues("completed", "success").Inc(ctx)
	}

	return nil
}

// MarkTaskFailed marks a task as failed with an error message
func (tsm *TaskStreamManager) MarkTaskFailed(ctx context.Context, taskID int64, errorMsg string) error {
	tsm.logger.Info(ctx, "Marking task as failed",
		observability.Int64("task_id", taskID),
		observability.String("error", errorMsg))

	// Find task in dispatched stream
	task, messageID, err := tsm.taskIndex.FindTaskByID(ctx, taskID)
	if err != nil {
		tsm.logger.Warn(ctx, "Failed to find task in dispatched stream, may already be processed",
			observability.Int64("task_id", taskID),
			observability.Error(err))
		// Don't return error - task may have already been processed
		return nil
	}

	// Move to failed stream
	if err := tsm.moveTaskToFailed(ctx, *task, errorMsg); err != nil {
		tsm.logger.Error(ctx, "Failed to move task to failed stream",
			observability.Int64("task_id", taskID),
			observability.Error(err))
		return err
	}

	// Acknowledge the task if we have the messageID
	if messageID != "" {
		err := tsm.AckTaskProcessed(ctx, StreamTaskDispatched, "task-processors", messageID)
		if err != nil {
			tsm.logger.Error(ctx, "Failed to acknowledge failed task",
				observability.Int64("task_id", taskID),
				observability.String("message_id", messageID),
				observability.Error(err))
		} else {
			// Remove the task from the index since it's been processed
			err = tsm.taskIndex.RemoveTaskIndex(ctx, taskID)
			if err != nil {
				tsm.logger.Warn(ctx, "failed to remove failed task from index",
					observability.Int64("task_id", taskID),
					observability.Error(err))
			}

			// Remove the task from timeout tracking since it's been failed
			err = tsm.expirationManager.RemoveTaskTimeout(ctx, taskID)
			if err != nil {
				tsm.logger.Warn(ctx, "failed to remove failed task from timeout tracking",
					observability.Int64("task_id", taskID),
					observability.Error(err))
			}
		}
	}

	tsm.logger.Info(ctx, "Task marked as failed successfully", observability.Int64("task_id", taskID))
	if metrics.TasksAddedToStreamTotal != nil {
		metrics.TasksAddedToStreamTotal.WithLabelValues("failed", "success").Inc(ctx)
	}

	return nil
}

// findTaskInDispatched finds a specific task in the dispatched stream
func (tsm *TaskStreamManager) findTaskInDispatched(taskID int64) (*TaskStreamData, error) {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, config.GetReadTimeout())
	defer cancel()

	// Use XRANGE to read recent messages without adding to PEL
	// This is a fallback method when the index is not available
	streams, err := tsm.redisClient.Client().XRange(ctx, StreamTaskDispatched, "-", "+").Result()
	if err != nil {
		return nil, fmt.Errorf("failed to read from stream: %w", err)
	}

	// Limit the search to the most recent 1000 messages to avoid performance issues
	startIndex := 0
	if len(streams) > 1000 {
		startIndex = len(streams) - 1000
	}

	for i := startIndex; i < len(streams); i++ {
		message := streams[i]
		taskJSON, exists := message.Values["task"].(string)
		if !exists {
			continue
		}

		var task TaskStreamData
		if err := json.Unmarshal([]byte(taskJSON), &task); err != nil {
			tsm.logger.Error(ctx, "Failed to unmarshal task data",
				observability.String("message_id", message.ID),
				observability.Error(err))
			continue
		}

		if task.SendTaskDataToKeeper.TaskID[0] == taskID {
			return &task, nil
		}
	}

	return nil, fmt.Errorf("task %d not found in processing stream", taskID)
}

func (tsm *TaskStreamManager) addTaskToStream(ctx context.Context, stream string, task *TaskStreamData) error {
	start := time.Now()
	ctx, cancel := context.WithTimeout(ctx, config.GetReadTimeout())
	defer cancel()

	taskJSON, err := json.Marshal(task)
	if err != nil {
		tsm.logger.Error(ctx, "Failed to marshal task data",
			observability.Int64("task_id", task.SendTaskDataToKeeper.TaskID[0]),
			observability.String("stream", stream),
			observability.Error(err))
		return fmt.Errorf("failed to marshal task data: %w", err)
	}

	res, err := tsm.redisClient.XAdd(ctx, &redis.XAddArgs{
		Stream: stream,
		MaxLen: int64(config.GetStreamMaxLen()),
		Approx: true,
		Values: map[string]interface{}{
			"task":       taskJSON,
			"created_at": time.Now().Unix(),
		},
	})
	duration := time.Since(start)

	if err != nil {
		if metrics.TasksAddedToStreamTotal != nil {
			metrics.TasksAddedToStreamTotal.WithLabelValues(stream, "failure").Inc(ctx)
		}
		tsm.logger.Error(ctx, "Failed to add task to stream",
			observability.Int64("task_id", task.SendTaskDataToKeeper.TaskID[0]),
			observability.String("stream", stream),
			observability.Duration("duration", duration),
			observability.Error(err))
		return fmt.Errorf("failed to add task to stream: %w", err)
	}

	// Add entry to expiration tracking with stream-specific TTL
	var entryTTL time.Duration
	switch stream {
	case StreamTaskDispatched:
		entryTTL = TasksProcessingTTL
	case StreamTaskCompleted:
		entryTTL = TasksCompletedTTL
	case StreamTaskFailed:
		entryTTL = TasksFailedTTL
	case StreamTaskRetry:
		entryTTL = TasksRetryTTL
	default:
		entryTTL = TasksProcessingTTL // Default fallback
	}

	// Track expiration for this stream entry
	err = tsm.expirationManager.AddMessageExpiration(ctx, stream, res, entryTTL)
	if err != nil {
		tsm.logger.Warn(ctx, "Failed to add stream entry expiration, but task was added to stream",
			observability.Int64("task_id", task.SendTaskDataToKeeper.TaskID[0]),
			observability.String("stream", stream),
			observability.String("message_id", res),
			observability.Duration("duration", duration),
			observability.Error(err))
		// Don't fail the entire operation if expiration tracking fails
	}

	if metrics.TasksAddedToStreamTotal != nil {
		metrics.TasksAddedToStreamTotal.WithLabelValues(stream, "success").Inc(ctx)
	}
	tsm.logger.Debug(ctx, "Task added to stream successfully",
		observability.Int64("task_id", task.SendTaskDataToKeeper.TaskID[0]),
		observability.String("stream", stream),
		observability.String("message_id", res),
		observability.Duration("duration", duration),
		observability.Int("task_json_size", len(taskJSON)),
		observability.Duration("entry_ttl", entryTTL))

	return nil
}
