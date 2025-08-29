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

	// Read dispatched tasks with a smaller limit for better performance
	// The index will help us find specific tasks efficiently when needed
	tasks, messageIDs, err := tsm.ReadTasksFromStream(StreamTaskDispatched, "timeout-checker", "timeout-worker", 100)
	if err != nil {
		tsm.logger.Error("Failed to read dispatched tasks for timeout check", "error", err)
		return
	}

	if len(tasks) > 0 {
		tsm.logger.Debug("Timeout worker found tasks to check", "task_count", len(tasks))
	}

	now := time.Now()
	timeoutCount := 0
	acknowledgedCount := 0

	for i, task := range tasks {
		shouldAcknowledge := false
		acknowledgeReason := ""

		if task.DispatchedAt != nil {
			dispatchedDuration := now.Sub(*task.DispatchedAt)
			if dispatchedDuration > TasksProcessingTTL {
				tsm.logger.Warn("Task dispatched timeout detected",
					"task_id", task.SendTaskDataToKeeper.TaskID[0],
					"dispatched_duration", dispatchedDuration,
					"timeout_threshold", TasksProcessingTTL)

				// Move to failed stream
				if err := tsm.moveTaskToFailed(ctx, task, "dispatched timeout"); err != nil {
					tsm.logger.Error("Failed to handle timeout task",
						"task_id", task.SendTaskDataToKeeper.TaskID[0],
						"error", err)
					continue // Don't acknowledge if we failed to move to failed stream
				}

				shouldAcknowledge = true
				acknowledgeReason = "timeout"
			}
		} else {
			// If task has no DispatchedAt timestamp, it might be a stale task
			// Check if it's been in the stream for too long based on CreatedAt
			if task.CreatedAt.Add(TasksProcessingTTL).Before(now) {
				tsm.logger.Warn("Task without dispatched timestamp timed out",
					"task_id", task.SendTaskDataToKeeper.TaskID[0],
					"created_at", task.CreatedAt,
					"age", now.Sub(task.CreatedAt))

				// Move to failed stream
				if err := tsm.moveTaskToFailed(ctx, task, "stale task timeout"); err != nil {
					tsm.logger.Error("Failed to handle stale timeout task",
						"task_id", task.SendTaskDataToKeeper.TaskID[0],
						"error", err)
					continue
				}

				shouldAcknowledge = true
				acknowledgeReason = "stale timeout"
			}
		}

		// Always acknowledge the message to prevent PEL growth
		// If the task was processed (moved to failed stream), acknowledge it
		// If the task is still valid, acknowledge it to remove from PEL
		if shouldAcknowledge {
			// Acknowledge the timed-out task
			err := tsm.AckTaskProcessed(ctx, StreamTaskDispatched, "timeout-checker", messageIDs[i])
			if err != nil {
				tsm.logger.Error("Failed to acknowledge timed-out task",
					"task_id", task.SendTaskDataToKeeper.TaskID[0],
					"message_id", messageIDs[i],
					"error", err)
			} else {
				// Remove the task from the index since it's been processed
				err = tsm.taskIndex.RemoveTaskIndex(ctx, task.SendTaskDataToKeeper.TaskID[0])
				if err != nil {
					tsm.logger.Warn("failed to remove timed-out task from index",
						"task_id", task.SendTaskDataToKeeper.TaskID[0],
						"error", err)
				}

				timeoutCount++
				acknowledgedCount++
				tsm.logger.Info("Task timeout processed and acknowledged successfully",
					"task_id", task.SendTaskDataToKeeper.TaskID[0],
					"message_id", messageIDs[i],
					"reason", acknowledgeReason)
			}
		} else {
			// Acknowledge valid tasks to prevent PEL growth
			err := tsm.AckTaskProcessed(ctx, StreamTaskDispatched, "timeout-checker", messageIDs[i])
			if err != nil {
				tsm.logger.Error("Failed to acknowledge valid task",
					"task_id", task.SendTaskDataToKeeper.TaskID[0],
					"message_id", messageIDs[i],
					"error", err)
			} else {
				acknowledgedCount++
				tsm.logger.Debug("Valid task acknowledged successfully",
					"task_id", task.SendTaskDataToKeeper.TaskID[0],
					"message_id", messageIDs[i])
			}
		}
	}

	if timeoutCount > 0 {
		tsm.logger.Info("Processed task timeouts", "timeout_count", timeoutCount)
	}
	if acknowledgedCount > 0 {
		tsm.logger.Debug("Acknowledged messages", "acknowledged_count", acknowledgedCount)
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
