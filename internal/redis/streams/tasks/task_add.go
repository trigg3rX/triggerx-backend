package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/trigg3rX/triggerx-backend/internal/redis/config"
	"github.com/trigg3rX/triggerx-backend/internal/redis/metrics"
	"github.com/trigg3rX/triggerx-backend/internal/redis/streams/performers"
	"github.com/trigg3rX/triggerx-backend/pkg/cryptography"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// Ready to be sent to the performer
func (tsm *TaskStreamManager) AddTaskToReadyStream(task TaskStreamData) (types.PerformerData, error) {
	performerData := performers.GetPerformerData()
	if performerData.KeeperID == 0 {
		tsm.logger.Error("No performers available for task", "task_id", task.SendTaskDataToKeeper.TaskID)
		return types.PerformerData{}, fmt.Errorf("no performers available")
	}

	// Update task with performer information
	task.SendTaskDataToKeeper.PerformerData = performerData
	task.SendTaskDataToKeeper.SchedulerSignature = &types.SchedulerSignatureData{
		TaskID:                  task.SendTaskDataToKeeper.TaskID,
		// TODO: add this before keeper v0.1.5
		// SchedulerID:             task.SendTaskDataToKeeper.SchedulerSignature.SchedulerID,
		SchedulerSigningAddress: config.GetRedisSigningAddress(),
	}

	// Sign the task data
	signature, err := cryptography.SignJSONMessage(task.SendTaskDataToKeeper, config.GetRedisSigningKey())
	if err != nil {
		tsm.logger.Errorf("Failed to sign batch task data: %v", err)
		return types.PerformerData{}, err
	}
	task.SendTaskDataToKeeper.SchedulerSignature.SchedulerSignature = signature

	// err = tsm.addTaskToStream(TasksReadyStream, &task)
	// if err != nil {
	// 	tsm.logger.Error("Failed to add task to ready stream",
	// 		"task_id", task.SendTaskDataToKeeper.TaskID,
	// 		"performer_id", performerData.KeeperID,
	// 		"error", err)
	// 	return types.PerformerData{}, err
	// }

	tsm.sendTaskToPerformer(task)

	tsm.logger.Info("Task added to ready stream successfully",
		"task_id", task.SendTaskDataToKeeper.TaskID,
		"performer_id", performerData.KeeperID,
		"performer_address", performerData.KeeperAddress)

	return performerData, nil
}

// Failed to send to the performer, sent to the retry stream
func (tsm *TaskStreamManager) AddTaskToRetryStream(task *TaskStreamData, retryReason string) error {
	tsm.logger.Warn("Adding task to retry stream",
		"task_id", task.SendTaskDataToKeeper.TaskID,
		"job_id", task.JobID,
		"retry_count", task.RetryCount,
		"retry_reason", retryReason)

	task.RetryCount++
	now := time.Now()
	task.LastAttemptAt = &now

	// Exponential backoff with jitter
	baseBackoff := time.Duration(task.RetryCount) * RetryBackoffBase
	jitter := time.Duration(rand.Int63n(int64(RetryBackoffBase))) // Up to 1 base backoff
	backoffDuration := baseBackoff + jitter

	scheduledFor := now.Add(backoffDuration)
	task.ScheduledFor = &scheduledFor

	tsm.logger.Info("Task retry scheduled",
		"task_id", task.SendTaskDataToKeeper.TaskID,
		"retry_count", task.RetryCount,
		"scheduled_for", scheduledFor.Format(time.RFC3339),
		"backoff_duration", backoffDuration)

	if task.RetryCount >= MaxRetryAttempts {
		tsm.logger.Error("Task exceeded max retry attempts, moving to failed stream",
			"task_id", task.SendTaskDataToKeeper.TaskID,
			"retry_count", task.RetryCount,
			"max_attempts", MaxRetryAttempts)

		metrics.TaskMaxRetriesExceededTotal.Inc()
		metrics.TasksMovedToFailedStreamTotal.Inc()

		err := tsm.addTaskToStream(TasksFailedStream, task)
		if err != nil {
			tsm.logger.Error("Failed to add task to failed stream",
				"task_id", task.SendTaskDataToKeeper.TaskID,
				"error", err)
			return err
		}

		// Notify scheduler about task failure
		// tsm.notifySchedulerTaskComplete(task, false)

		tsm.logger.Error("Task permanently failed",
			"task_id", task.SendTaskDataToKeeper.TaskID,
			"final_retry_count", task.RetryCount)

		return nil
	}

	err := tsm.addTaskToStream(TasksRetryStream, task)
	if err != nil {
		tsm.logger.Error("Failed to add task to retry stream",
			"task_id", task.SendTaskDataToKeeper.TaskID,
			"error", err)
		metrics.TasksAddedToStreamTotal.WithLabelValues("retry", "failure").Inc()
		return err
	}

	metrics.TasksAddedToStreamTotal.WithLabelValues("retry", "success").Inc()
	return nil
}

func (tsm *TaskStreamManager) addTaskToStream(stream string, task *TaskStreamData) error {
	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	taskJSON, err := json.Marshal(task)
	if err != nil {
		tsm.logger.Error("Failed to marshal task data",
			"task_id", task.SendTaskDataToKeeper.TaskID,
			"stream", stream,
			"error", err)
		return err
	}

	res, err := tsm.client.XAdd(ctx, &redis.XAddArgs{
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
			"task_id", task.SendTaskDataToKeeper.TaskID,
			"stream", stream,
			"duration", duration,
			"error", err)
		return err
	}

	metrics.TasksAddedToStreamTotal.WithLabelValues(stream, "success").Inc()
	tsm.logger.Debug("Task added to stream successfully",
		"task_id", task.SendTaskDataToKeeper.TaskID,
		"stream", stream,
		"stream_id", res,
		"duration", duration,
		"task_json_size", len(taskJSON))

	return nil
}
