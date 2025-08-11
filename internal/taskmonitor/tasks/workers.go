package tasks

import (
	"context"
	"fmt"
	"time"

	"github.com/trigg3rX/triggerx-backend/internal/taskmonitor/metrics"
)

// StartTimeoutWorker monitors processing tasks for timeouts
func (tsm *TaskStreamManager) StartTimeoutWorker(ctx context.Context) {
	tsm.logger.Info("Starting task timeout worker")

	ticker := time.NewTicker(30 * time.Second) // Check timeouts every 30 seconds
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			tsm.logger.Info("Timeout worker stopping")
			return
		case <-ticker.C:
			tsm.checkDispatchedTimeouts(ctx)
		}
	}
}

// checkDispatchedTimeouts checks for tasks that have been dispatched too long
func (tsm *TaskStreamManager) checkDispatchedTimeouts(ctx context.Context) {
	tsm.logger.Debug("Checking for dispatched timeouts")

	// Read all dispatched tasks (simplified - in production would use consumer groups)
	tasks, messageIDs, err := tsm.ReadTasksFromStream(StreamTaskDispatched, "timeout-checker", "timeout-worker", 100)
	if err != nil {
		tsm.logger.Error("Failed to read dispatched tasks for timeout check", "error", err)
		return
	}

	now := time.Now()
	timeoutCount := 0

	for i, task := range tasks {
		if task.DispatchedAt != nil {
			dispatchedDuration := now.Sub(*task.DispatchedAt)
			if dispatchedDuration > TasksProcessingTTL {
				tsm.logger.Warn("Task dispatched timeout detected",
					"task_id", task.SendTaskDataToKeeper.TaskID[0],
					"dispatched_duration", dispatchedDuration)

				// Move to failed/retry
				if err := tsm.moveTaskToFailed(ctx, task, "dispatched timeout"); err != nil {
					tsm.logger.Error("Failed to handle timeout task",
						"task_id", task.SendTaskDataToKeeper.TaskID[0],
						"error", err)
				} else {
					// Acknowledge the timed-out task
					err := tsm.AckTaskProcessed(ctx, StreamTaskDispatched, "timeout-checker", messageIDs[i])
					if err != nil {
						tsm.logger.Error("Failed to acknowledge timed-out task",
							"task_id", task.SendTaskDataToKeeper.TaskID[0],
							"error", err)
					}
					timeoutCount++
				}
			}
		}
	}

	if timeoutCount > 0 {
		tsm.logger.Info("Processed task timeouts", "timeout_count", timeoutCount)
	}
}

// AckTaskProcessed acknowledges that a task has been processed
func (tsm *TaskStreamManager) AckTaskProcessed(ctx context.Context, stream, consumerGroup, messageID string) error {
	tsm.logger.Debug("Acknowledging task processed",
		"stream", stream,
		"consumer_group", consumerGroup,
		"message_id", messageID)

	// Increase timeout for acknowledgment operations
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	err := tsm.redisClient.XAck(ctx, stream, consumerGroup, messageID)
	if err != nil {
		tsm.logger.Error("Failed to acknowledge task",
			"stream", stream,
			"consumer_group", consumerGroup,
			"message_id", messageID,
			"error", err)
		return err
	}

	tsm.logger.Debug("Task acknowledged successfully",
		"stream", stream,
		"message_id", messageID)

	return nil
}

// moveTaskToFailed moves a task to the failed stream or retry stream
func (tsm *TaskStreamManager) moveTaskToFailed(ctx context.Context, task TaskStreamData, errorMsg string) error {
	task.LastError = errorMsg
	task.RetryCount++

	// Move to failed stream permanently
	err := tsm.addTaskToStream(ctx, StreamTaskFailed, &task)
	if err != nil {
		return fmt.Errorf("failed to add to failed stream: %w", err)
	}

	tsm.logger.Error("Task permanently failed",
		"task_id", task.SendTaskDataToKeeper.TaskID[0],
		"retry_count", task.RetryCount,
		"error", errorMsg)

	metrics.TasksAddedToStreamTotal.WithLabelValues("failed", "success").Inc()

	return nil
}
