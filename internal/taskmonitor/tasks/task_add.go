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
	"github.com/trigg3rX/triggerx-backend/pkg/types"
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

	task.ValidatedAt = &[]time.Time{time.Now()}[0]

	// Add to completed stream
	err = tsm.addTaskToStream(ctx, StreamTaskValidated, task)
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

// MarkTaskExecuted marks a task as executed (moves from dispatched to executed stream)
// This is called when a task has been successfully executed and sent to aggregator
func (tsm *TaskStreamManager) MarkTaskExecuted(ctx context.Context, taskID int64, timeoutDuration time.Duration) error {
	tsm.logger.Info(ctx, "Marking task as executed",
		observability.Int64("task_id", taskID))

	// Try to find task in dispatched stream first
	task, messageID, err := tsm.taskIndex.FindTaskByIDInStream(ctx, taskID, StreamTaskDispatched)
	if err != nil {
		// Task might not be in dispatched stream (e.g., if it was added directly)
		// Create a minimal task entry for executed stream
		tsm.logger.Debug(ctx, "Task not found in dispatched stream, creating minimal entry",
			observability.Int64("task_id", taskID))

		// Create minimal task data - we'll need task definition ID from database
		// For now, we'll create a basic entry
		now := time.Now()
		task = &TaskStreamData{
			CreatedAt:  now,
			ExecutedAt: &now,
			SendTaskDataToKeeper: types.SendTaskDataToKeeper{
				TaskID: []int64{taskID},
			},
		}
		messageID = "" // No message to acknowledge
	} else {
		// Update executed timestamp
		now := time.Now()
		task.ExecutedAt = &now
	}

	// Add to executed stream
	messageIDExecuted, err := tsm.addTaskToStreamWithMessageID(ctx, StreamTaskExecuted, task)
	if err != nil {
		tsm.logger.Error(ctx, "failed to add to executed stream", observability.Error(err))
		return err
	}

	// Store task index for executed stream
	if messageIDExecuted != "" {
		if err := tsm.taskIndex.StoreTaskIndex(ctx, taskID, messageIDExecuted); err != nil {
			tsm.logger.Warn(ctx, "failed to store task index for executed stream",
				observability.Int64("task_id", taskID),
				observability.Error(err))
		}
	}

	// Add timeout tracking for executed tasks
	if timeoutDuration > 0 {
		err = tsm.expirationManager.AddExecutedTaskTimeout(ctx, taskID, timeoutDuration)
		if err != nil {
			tsm.logger.Warn(ctx, "failed to add executed task timeout",
				observability.Int64("task_id", taskID),
				observability.Error(err))
		}
	}

	// Remove from dispatched stream if it was there
	if messageID != "" {
		err = tsm.AckTaskProcessed(ctx, StreamTaskDispatched, "task-processors", messageID)
		if err != nil {
			tsm.logger.Warn(ctx, "failed to acknowledge task from dispatched stream",
				observability.Int64("task_id", taskID),
				observability.String("message_id", messageID),
				observability.Error(err))
		} else {
			// Remove the task from the index since it's been moved
			err = tsm.taskIndex.RemoveTaskIndex(ctx, taskID)
			if err != nil {
				tsm.logger.Warn(ctx, "failed to remove task from index",
					observability.Int64("task_id", taskID),
					observability.Error(err))
			}
		}
	}

	tsm.logger.Info(ctx, "Task marked as executed successfully", observability.Int64("task_id", taskID))
	if metrics.TasksAddedToStreamTotal != nil {
		metrics.TasksAddedToStreamTotal.WithLabelValues("executed", "success").Inc(ctx)
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

// findTaskInStream finds a specific task in a given stream (fallback method when index is not available)
func (tsm *TaskStreamManager) findTaskInStream(taskID int64, stream string) (*TaskStreamData, error) {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, config.GetReadTimeout())
	defer cancel()

	// Use XRANGE to read recent messages without adding to PEL
	// This is a fallback method when the index is not available
	streams, err := tsm.redisClient.Client().XRange(ctx, stream, "-", "+").Result()
	if err != nil {
		return nil, fmt.Errorf("failed to read from stream %s: %w", stream, err)
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
				observability.String("stream", stream),
				observability.Error(err))
			continue
		}

		if task.SendTaskDataToKeeper.TaskID[0] == taskID {
			return &task, nil
		}
	}

	return nil, fmt.Errorf("task %d not found in stream %s", taskID, stream)
}

// addTaskToStreamWithMessageID adds a task to a stream and returns the message ID
func (tsm *TaskStreamManager) addTaskToStreamWithMessageID(ctx context.Context, stream string, task *TaskStreamData) (string, error) {
	start := time.Now()
	ctx, cancel := context.WithTimeout(ctx, config.GetReadTimeout())
	defer cancel()

	taskJSON, err := json.Marshal(task)
	if err != nil {
		tsm.logger.Error(ctx, "Failed to marshal task data",
			observability.Int64("task_id", task.SendTaskDataToKeeper.TaskID[0]),
			observability.String("stream", stream),
			observability.Error(err))
		return "", fmt.Errorf("failed to marshal task data: %w", err)
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
		return "", fmt.Errorf("failed to add task to stream: %w", err)
	}

	// Add entry to expiration tracking with stream-specific TTL
	var entryTTL time.Duration
	switch stream {
	case StreamTaskDispatched:
		entryTTL = TasksProcessingTTL
	case StreamTaskExecuted:
		entryTTL = TasksExecutedTTL
	case StreamTaskValidated:
		entryTTL = TasksValidatedTTL
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

	return res, nil
}

func (tsm *TaskStreamManager) addTaskToStream(ctx context.Context, stream string, task *TaskStreamData) error {
	_, err := tsm.addTaskToStreamWithMessageID(ctx, stream, task)
	return err
}
