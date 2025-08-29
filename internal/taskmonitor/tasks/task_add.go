package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/trigg3rX/triggerx-backend/internal/taskmonitor/config"
	"github.com/trigg3rX/triggerx-backend/internal/taskmonitor/metrics"
)

// MarkTaskCompleted marks a task as completed
func (tsm *TaskStreamManager) MarkTaskCompleted(ctx context.Context, taskID int64) error {
	tsm.logger.Info("Marking task as completed",
		"task_id", taskID)

	// Find and move task from processing to completed using efficient lookup
	task, messageID, err := tsm.taskIndex.FindTaskByID(ctx, taskID)
	if err != nil {
		tsm.logger.Error("failed to find task in processing", "error", err)
		// return err
	}

	task.CompletedAt = &[]time.Time{time.Now()}[0]

	// Add to completed stream
	err = tsm.addTaskToStream(ctx, StreamTaskCompleted, task)
	if err != nil {
		tsm.logger.Error("failed to add to completed stream", "error", err)
		// return err
	}

	// Remove from processing stream (acknowledge) using the messageID
	if messageID != "" {
		err = tsm.AckTaskProcessed(ctx, StreamTaskDispatched, "task-processors", messageID)
		if err != nil {
			tsm.logger.Error("failed to acknowledge task",
				"task_id", taskID,
				"message_id", messageID,
				"error", err)
		} else {
			// Remove the task from the index since it's been processed
			err = tsm.taskIndex.RemoveTaskIndex(ctx, taskID)
			if err != nil {
				tsm.logger.Warn("failed to remove task from index",
					"task_id", taskID,
					"error", err)
			}
		}
	}

	tsm.logger.Info("Task marked as completed successfully", "task_id", taskID)
	metrics.TasksAddedToStreamTotal.WithLabelValues("completed", "success").Inc()

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
			tsm.logger.Error("Failed to unmarshal task data",
				"message_id", message.ID,
				"error", err)
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
		tsm.logger.Error("Failed to marshal task data",
			"task_id", task.SendTaskDataToKeeper.TaskID[0],
			"stream", stream,
			"error", err)
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
		metrics.TasksAddedToStreamTotal.WithLabelValues(stream, "failure").Inc()
		tsm.logger.Error("Failed to add task to stream",
			"task_id", task.SendTaskDataToKeeper.TaskID[0],
			"stream", stream,
			"duration", duration,
			"error", err)
		return fmt.Errorf("failed to add task to stream: %w", err)
	}

	metrics.TasksAddedToStreamTotal.WithLabelValues(stream, "success").Inc()
	tsm.logger.Debug("Task added to stream successfully",
		"task_id", task.SendTaskDataToKeeper.TaskID[0],
		"stream", stream,
		"stream_id", res,
		"duration", duration,
		"task_json_size", len(taskJSON))

	return nil
}
