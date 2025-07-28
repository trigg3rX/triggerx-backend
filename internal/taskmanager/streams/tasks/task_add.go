package tasks

import (
	// "context"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	// "github.com/redis/go-redis/v9"
	"github.com/trigg3rX/triggerx-backend/internal/taskmanager/config"
	"github.com/trigg3rX/triggerx-backend/internal/taskmanager/metrics"
	"github.com/trigg3rX/triggerx-backend/pkg/cryptography"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// Ready to be sent to the performer with improved error handling and performance
func (tsm *TaskStreamManager) AddTaskToReadyStream(task TaskStreamData) (types.PerformerData, error) {
	// Use dynamic performer selection instead of hardcoded selection
	performerData, err := tsm.performerManager.GetPerformerData(task.SendTaskDataToKeeper.TargetData[0].IsImua)
	if err != nil {
		tsm.logger.Error("Failed to get performer data dynamically",
			"task_id", task.SendTaskDataToKeeper.TaskID[0],
			"error", err)
		return types.PerformerData{}, fmt.Errorf("failed to get performer: %w", err)
	}

	// Update task with performer information
	task.SendTaskDataToKeeper.PerformerData = performerData

	// Sign the task data with improved error handling
	signature, err := cryptography.SignJSONMessage(task.SendTaskDataToKeeper, config.GetRedisSigningKey())
	if err != nil {
		tsm.logger.Error("Failed to sign batch task data",
			"task_id", task.SendTaskDataToKeeper.TaskID[0],
			"error", err)
		return types.PerformerData{}, fmt.Errorf("failed to sign task data: %w", err)
	}
	task.SendTaskDataToKeeper.ManagerSignature = signature

	// Add task to stream with improved performance
	err = tsm.addTaskToStream(TasksReadyStream, &task)
	if err != nil {
		tsm.logger.Error("Failed to add task to ready stream",
			"task_id", task.SendTaskDataToKeeper.TaskID,
			"performer_id", performerData.OperatorID,
			"error", err)
		return types.PerformerData{}, fmt.Errorf("failed to add task to ready stream: %w", err)
	}

	// Send task to performer asynchronously for better performance
	go func() {
		tsm.sendTaskToPerformer(task)
	}()

	tsm.logger.Info("Task added to ready stream successfully",
		"task_id", task.SendTaskDataToKeeper.TaskID[0],
		"performer_id", performerData.OperatorID,
		"performer_address", performerData.KeeperAddress)

	return performerData, nil
}

// Failed to send to the performer, sent to the retry stream with improved backoff strategy
func (tsm *TaskStreamManager) AddTaskToRetryStream(task *TaskStreamData, retryReason string) error {
	tsm.logger.Warn("Adding task to retry stream",
		"task_id", task.SendTaskDataToKeeper.TaskID[0],
		"job_id", task.JobID,
		"retry_count", task.RetryCount,
		"retry_reason", retryReason)

	task.RetryCount++
	now := time.Now()
	task.LastAttemptAt = &now

	// Improved exponential backoff with jitter and maximum cap
	baseBackoff := time.Duration(task.RetryCount) * RetryBackoffBase
	maxBackoff := config.GetMaxRetryBackoff()
	if baseBackoff > maxBackoff {
		baseBackoff = maxBackoff
	}

	jitter := time.Duration(rand.Int63n(int64(RetryBackoffBase))) // Up to 1 base backoff
	backoffDuration := baseBackoff + jitter

	scheduledFor := now.Add(backoffDuration)
	task.ScheduledFor = &scheduledFor

	tsm.logger.Info("Task retry scheduled",
		"task_id", task.SendTaskDataToKeeper.TaskID[0],
		"retry_count", task.RetryCount,
		"scheduled_for", scheduledFor.Format(time.RFC3339),
		"backoff_duration", backoffDuration)

	if task.RetryCount >= MaxRetryAttempts {
		tsm.logger.Error("Task exceeded max retry attempts, moving to failed stream",
			"task_id", task.SendTaskDataToKeeper.TaskID[0],
			"retry_count", task.RetryCount,
			"max_attempts", MaxRetryAttempts)

		metrics.TaskMaxRetriesExceededTotal.Inc()
		metrics.TasksMovedToFailedStreamTotal.Inc()

		err := tsm.addTaskToStream(TasksFailedStream, task)
		if err != nil {
			tsm.logger.Error("Failed to add task to failed stream",
				"task_id", task.SendTaskDataToKeeper.TaskID[0],
				"error", err)
			return fmt.Errorf("failed to add task to failed stream: %w", err)
		}

		// Notify scheduler about task failure
		// tsm.notifySchedulerTaskComplete(task, false)

		tsm.logger.Error("Task permanently failed",
			"task_id", task.SendTaskDataToKeeper.TaskID[0],
			"final_retry_count", task.RetryCount)

		return nil
	}

	err := tsm.addTaskToStream(TasksRetryStream, task)
	if err != nil {
		tsm.logger.Error("Failed to add task to retry stream",
			"task_id", task.SendTaskDataToKeeper.TaskID[0],
			"error", err)
		metrics.TasksAddedToStreamTotal.WithLabelValues("retry", "failure").Inc()
		return fmt.Errorf("failed to add task to retry stream: %w", err)
	}

	metrics.TasksAddedToStreamTotal.WithLabelValues("retry", "success").Inc()
	return nil
}

func (tsm *TaskStreamManager) addTaskToStream(stream string, task *TaskStreamData) error {
	start := time.Now()
	// ctx, cancel := context.WithTimeout(context.Background(), config.GetStreamOperationTimeout())
	// defer cancel()

	taskJSON, err := json.Marshal(task)
	if err != nil {
		tsm.logger.Error("Failed to marshal task data",
			"task_id", task.SendTaskDataToKeeper.TaskID[0],
			"stream", stream,
			"error", err)
		return fmt.Errorf("failed to marshal task data: %w", err)
	}

	// res, err := tsm.client.XAdd(ctx, &redis.XAddArgs{
	// 	Stream: stream,
	// 	MaxLen: int64(config.GetStreamMaxLen()),
	// 	Approx: true,
	// 	Values: map[string]interface{}{
	// 		"task":       taskJSON,
	// 		"created_at": time.Now().Unix(),
	// 	},
	// })
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
		// "stream_id", res,
		"duration", duration,
		"task_json_size", len(taskJSON))

	return nil
}
