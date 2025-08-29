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
			// tsm.logger.Debug("Timeout worker checking for timed out tasks")
			tsm.checkDispatchedTimeouts(ctx)
		}
	}
}

// checkDispatchedTimeouts checks for tasks that have been dispatched too long
func (tsm *TaskStreamManager) checkDispatchedTimeouts(ctx context.Context) {
	// tsm.logger.Debug("Checking for dispatched timeouts")

	// Get expired tasks efficiently using sorted set
	expiredTaskIDs, err := tsm.timeoutManager.GetExpiredTasks(ctx)
	if err != nil {
		tsm.logger.Error("Failed to get expired tasks", "error", err)
		return
	}

	if len(expiredTaskIDs) == 0 {
		// tsm.logger.Debug("No expired tasks found")
		return
	}

	tsm.logger.Info("Found expired tasks", "expired_count", len(expiredTaskIDs))

	// Process each expired task
	processedCount := 0
	var processedTaskIDs []int64

	for _, taskID := range expiredTaskIDs {
		// Find the task using the efficient index lookup
		task, messageID, err := tsm.taskIndex.FindTaskByID(ctx, taskID)
		if err != nil {
			tsm.logger.Error("Failed to find expired task",
				"task_id", taskID,
				"error", err)
			// Still remove from timeout tracking to prevent reprocessing
			processedTaskIDs = append(processedTaskIDs, taskID)
			continue
		}

		tsm.logger.Warn("Task timeout detected",
			"task_id", taskID,
			"dispatched_at", task.DispatchedAt,
			"created_at", task.CreatedAt)

		// Move to failed stream
		if err := tsm.moveTaskToFailed(ctx, *task, "dispatched timeout"); err != nil {
			tsm.logger.Error("Failed to handle timeout task",
				"task_id", taskID,
				"error", err)
			continue // Don't acknowledge if we failed to move to failed stream
		}

		// Acknowledge the timed-out task if we have the messageID
		if messageID != "" {
			err := tsm.AckTaskProcessed(ctx, StreamTaskDispatched, "timeout-checker", messageID)
			if err != nil {
				tsm.logger.Error("Failed to acknowledge timed-out task",
					"task_id", taskID,
					"message_id", messageID,
					"error", err)
			} else {
				tsm.logger.Info("Task timeout processed and acknowledged successfully",
					"task_id", taskID,
					"message_id", messageID)
			}
		}

		// Remove from task index since it's been processed
		err = tsm.taskIndex.RemoveTaskIndex(ctx, taskID)
		if err != nil {
			tsm.logger.Warn("failed to remove timed-out task from index",
				"task_id", taskID,
				"error", err)
		}

		processedCount++
		processedTaskIDs = append(processedTaskIDs, taskID)
	}

	// Remove all processed tasks from timeout tracking in batch
	if len(processedTaskIDs) > 0 {
		err = tsm.timeoutManager.RemoveMultipleTaskTimeouts(ctx, processedTaskIDs)
		if err != nil {
			tsm.logger.Error("Failed to remove processed tasks from timeout tracking",
				"task_count", len(processedTaskIDs),
				"error", err)
		} else {
			tsm.logger.Info("Removed processed tasks from timeout tracking",
				"task_count", len(processedTaskIDs))
		}
	}

	if processedCount > 0 {
		tsm.logger.Info("Processed task timeouts", "processed_count", processedCount)
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
