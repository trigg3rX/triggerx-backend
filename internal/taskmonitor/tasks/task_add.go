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

	// Find and move task from processing to completed
	task, err := tsm.findTaskInDispatched(taskID)
	if err != nil {
		return fmt.Errorf("failed to find task in processing: %w", err)
	}

	task.CompletedAt = &[]time.Time{time.Now()}[0]

	// Add to completed stream
	err = tsm.addTaskToStream(ctx, StreamTaskCompleted, task)
	if err != nil {
		return fmt.Errorf("failed to add to completed stream: %w", err)
	}

	// Remove from processing stream (acknowledge)
	// Note: In a real implementation, we'd need to track the processing message ID
	tsm.logger.Info("Task marked as completed successfully", "task_id", taskID)
	metrics.TasksAddedToStreamTotal.WithLabelValues("completed", "success").Inc()

	return nil
}

// findTaskInDispatched finds a specific task in the dispatched stream
func (tsm *TaskStreamManager) findTaskInDispatched(taskID int64) (*TaskStreamData, error) {
	tasks, _, err := tsm.ReadTasksFromStream(StreamTaskDispatched, "task-finder", "finder", 1000)
	if err != nil {
		return nil, err
	}

	for _, task := range tasks {
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
