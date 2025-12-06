package tasks

import (
	"context"
	"fmt"
	"time"

	"github.com/trigg3rX/triggerx-backend/internal/taskmonitor/clients/notify"
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
	expiredTaskIDs, err := tsm.expirationManager.GetExpiredTasks(ctx)
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

		err = tsm.dbClient.UpdateTaskFailed(taskID)
		if err != nil {
			tsm.logger.Error("Failed to update task failed",
				"task_id", taskID,
				"error", err)
		}

		// Notify user on failure
		if tsm.notifier != nil {
			email, e := tsm.dbClient.GetUserEmailByTaskID(taskID)
			if e != nil {
				tsm.logger.Warn("Could not fetch user email for task failure", "task_id", taskID, "error", e)
			} else if email != "" {
				payload := notify.TaskStatusPayload{
					TaskID:     taskID,
					JobID:      0,
					Status:     "failed",
					IsAccepted: false,
					Error:      "dispatched timeout",
					OccurredAt: time.Now(),
				}
				if err := tsm.notifier.NotifyTaskStatus(context.Background(), email, payload); err != nil {
					tsm.logger.Warn("Failed to notify user for task failure", "email", email, "task_id", taskID, "error", err)
				}
			}
		}
	}

	// Remove all processed tasks from timeout tracking in batch
	if len(processedTaskIDs) > 0 {
		err = tsm.expirationManager.RemoveMultipleTaskTimeouts(ctx, processedTaskIDs)
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

// AckTaskProcessed acknowledges that a task has been processed and optionally deletes it from the stream
func (tsm *TaskStreamManager) AckTaskProcessed(ctx context.Context, stream, consumerGroup, messageID string) error {
	return tsm.AckTaskProcessedWithDelete(ctx, stream, consumerGroup, messageID, true)
}

// AckTaskProcessedWithDelete acknowledges that a task has been processed and optionally deletes it from the stream
func (tsm *TaskStreamManager) AckTaskProcessedWithDelete(ctx context.Context, stream, consumerGroup, messageID string, deleteFromStream bool) error {
	tsm.logger.Debug("Acknowledging task processed",
		"stream", stream,
		"consumer_group", consumerGroup,
		"message_id", messageID,
		"delete_from_stream", deleteFromStream)

	// Increase timeout for acknowledgment operations
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// First acknowledge the message (removes from PEL)
	err := tsm.redisClient.XAck(ctx, stream, consumerGroup, messageID)
	if err != nil {
		tsm.logger.Error("Failed to acknowledge task",
			"stream", stream,
			"consumer_group", consumerGroup,
			"message_id", messageID,
			"error", err)
		return err
	}

	// Remove from expiration tracking since the message has been processed
	err = tsm.expirationManager.RemoveMessageExpiration(ctx, stream, messageID)
	if err != nil {
		tsm.logger.Warn("Failed to remove stream entry expiration after acknowledgment",
			"stream", stream,
			"message_id", messageID,
			"error", err)
		// Don't return error - message was acknowledged, just couldn't remove from expiration tracking
	}

	// If deleteFromStream is true, delete the message from the stream itself
	// This is important because XACK only removes from PEL, not from the stream
	if deleteFromStream {
		deleted, err := tsm.redisClient.XDel(ctx, stream, messageID)
		if err != nil {
			tsm.logger.Warn("Failed to delete message from stream after acknowledgment",
				"stream", stream,
				"message_id", messageID,
				"error", err)
			// Don't return error - message was acknowledged, just couldn't delete
			// The message will be cleaned up by periodic trimming or expiration
		} else if deleted > 0 {
			tsm.logger.Debug("Task acknowledged and deleted from stream successfully",
				"stream", stream,
				"message_id", messageID)
		} else {
			tsm.logger.Debug("Task acknowledged but message not found in stream (may have been already deleted)",
				"stream", stream,
				"message_id", messageID)
		}
	}

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
