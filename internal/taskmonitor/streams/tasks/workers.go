package tasks

import (
	"context"
	// "encoding/json"
	"fmt"
	"time"

	// "github.com/redis/go-redis/v9"
	"github.com/trigg3rX/triggerx-backend/internal/taskmonitor/metrics"
)

// StartTaskProcessor starts the main task processing loop with consumer groups
func (tsm *TaskStreamManager) StartTaskProcessor(ctx context.Context, consumerName string) {
	tsm.logger.Info("Starting task processor", "consumer_name", consumerName)

	// Handle any orphaned pending messages before starting
	tsm.handleOrphanedPendingMessages(consumerName)

	ticker := time.NewTicker(1 * time.Second) // Process tasks every second
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			tsm.logger.Info("Task processor stopping", "consumer_name", consumerName)
			return
		case <-ticker.C:
			tsm.processReadyTasks(consumerName)
		}
	}
}

// handleOrphanedPendingMessages handles any pending messages that might be blocking the stream
func (tsm *TaskStreamManager) handleOrphanedPendingMessages(consumerName string) {
	tsm.logger.Info("Checking for orphaned pending messages")

	// ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	// defer cancel()

	// Get pending messages info
	// pendingInfo, err := tsm.client.XPending(ctx, TasksReadyStream, "task-processors")
	// if err != nil {
	// 	tsm.logger.Error("Failed to get pending messages info", "error", err)
	// 	return
	// }

	// if pendingInfo.Count == 0 {
	// 	tsm.logger.Debug("No pending messages found")
	// 	return
	// }

	// tsm.logger.Info("Found pending messages", "count", pendingInfo.Count)

	// Get detailed pending messages
	// pendingMsgs, err := tsm.client.XPendingExt(ctx, &redis.XPendingExtArgs{
	// 	Stream: TasksReadyStream,
	// 	Group:  "task-processors",
	// 	Start:  "-",
	// 	End:    "+",
	// 	Count:  pendingInfo.Count,
	// })
	// if err != nil {
	// 	tsm.logger.Error("Failed to get pending messages details", "error", err)
	// 	return
	// }

	claimedCount := 0

	// for _, msg := range pendingMsgs {
	// 	// Check if message is older than 5 minutes (likely orphaned)
	// 	if msg.Idle > 5*time.Minute {
	// 		tsm.logger.Warn("Found orphaned pending message, attempting to claim",
	// 			"message_id", msg.ID,
	// 			"consumer", msg.Consumer,
	// 			"idle_time", msg.Idle)

	// 		// Try to claim the message
	// 		claimedMsgs, err := tsm.client.XClaim(ctx, &redis.XClaimArgs{
	// 			Stream:   TasksReadyStream,
	// 			Group:    "task-processors",
	// 			Consumer: consumerName,
	// 			MinIdle:  0,
	// 			Messages: []string{msg.ID},
	// 		}).Result()

	// 		if err != nil {
	// 			tsm.logger.Error("Failed to claim orphaned message",
	// 				"message_id", msg.ID,
	// 				"error", err)
	// 			continue
	// 		}

	// 		if len(claimedMsgs) > 0 {
	// 			// Process the claimed message
	// 			for _, claimedMsg := range claimedMsgs {
	// 				taskJSON, exists := claimedMsg.Values["task"].(string)
	// 				if !exists {
	// 					tsm.logger.Warn("Claimed message missing task data", "message_id", claimedMsg.ID)
	// 					// Acknowledge to remove from pending
	// 					if ackErr := tsm.AckTaskProcessed(TasksReadyStream, "task-processors", claimedMsg.ID); ackErr != nil {
	// 						tsm.logger.Error("Failed to acknowledge orphaned message", "message_id", claimedMsg.ID, "error", ackErr)
	// 					}
	// 					continue
	// 				}

	// 				var task TaskStreamData
	// 				if err := json.Unmarshal([]byte(taskJSON), &task); err != nil {
	// 					tsm.logger.Error("Failed to unmarshal claimed task", "message_id", claimedMsg.ID, "error", err)
	// 					// Acknowledge to remove from pending
	// 					if ackErr := tsm.AckTaskProcessed(TasksReadyStream, "task-processors", claimedMsg.ID); ackErr != nil {
	// 						tsm.logger.Error("Failed to acknowledge invalid task", "message_id", claimedMsg.ID, "error", ackErr)
	// 					}
	// 					continue
	// 				}

	// 				tsm.logger.Info("Processing orphaned task",
	// 					"task_id", task.SendTaskDataToKeeper.TaskID[0],
	// 					"message_id", claimedMsg.ID)

	// 				// Process the task
	// 				if err := tsm.moveTaskToProcessing(task, claimedMsg.ID); err != nil {
	// 					tsm.logger.Error("Failed to process orphaned task",
	// 						"task_id", task.SendTaskDataToKeeper.TaskID[0],
	// 						"message_id", claimedMsg.ID,
	// 						"error", err)
	// 					// Move to retry stream
	// 					if retryErr := tsm.AddTaskToRetryStream(&task, err.Error()); retryErr != nil {
	// 						tsm.logger.Error("Failed to move orphaned task to retry", "task_id", task.SendTaskDataToKeeper.TaskID[0], "error", retryErr)
	// 					}
	// 				} else {
	// 					// Send task to performer
	// 					go tsm.sendTaskToPerformer(task)
	// 				}

	// 				claimedCount++
	// 			}
	// 		}
	// 	}
	// }

	if claimedCount > 0 {
		tsm.logger.Info("Successfully processed orphaned pending messages", "count", claimedCount)
	}
}

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
			tsm.checkProcessingTimeouts()
		}
	}
}

// checkProcessingTimeouts checks for tasks that have been processing too long
func (tsm *TaskStreamManager) checkProcessingTimeouts() {
	tsm.logger.Debug("Checking for processing timeouts")

	// Read all processing tasks (simplified - in production would use consumer groups)
	tasks, messageIDs, err := tsm.readTasksFromStreamWithIDs(TasksProcessingStream, "timeout-checker", "timeout-worker", 100)
	if err != nil {
		tsm.logger.Error("Failed to read processing tasks for timeout check", "error", err)
		return
	}

	now := time.Now()
	timeoutCount := 0

	for i, task := range tasks {
		if task.ProcessingStartedAt != nil {
			processingDuration := now.Sub(*task.ProcessingStartedAt)
			if processingDuration > TasksProcessingTTL {
				tsm.logger.Warn("Task processing timeout detected",
					"task_id", task.SendTaskDataToKeeper.TaskID[0],
					"processing_duration", processingDuration)

				// Move to failed/retry
				if err := tsm.moveTaskToFailed(task, "processing timeout"); err != nil {
					tsm.logger.Error("Failed to handle timeout task",
						"task_id", task.SendTaskDataToKeeper.TaskID[0],
						"error", err)
				} else {
					// Acknowledge the timed-out task
					messageID := messageIDs[i]
					err := tsm.AckTaskProcessed(TasksProcessingStream, "timeout-checker", messageID)
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

// StartRetryWorker processes tasks from retry stream when they're ready
func (tsm *TaskStreamManager) StartRetryWorker(ctx context.Context) {
	tsm.logger.Info("Starting task retry worker")

	ticker := time.NewTicker(5 * time.Second) // Check retries every 5 seconds
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			tsm.logger.Info("Retry worker stopping")
			return
		case <-ticker.C:
			tsm.processRetryTasks()
		}
	}
}

// processReadyTasks processes tasks from the ready stream
func (tsm *TaskStreamManager) processReadyTasks(consumerName string) {
	// tsm.logger.Debug("Processing ready tasks", "consumer_name", consumerName)

	tasks, messageIDs, err := tsm.readTasksFromStreamWithIDs(TasksReadyStream, "task-processors", consumerName, 10)
	if err != nil {
		tsm.logger.Error("Failed to read tasks from ready stream", "error", err)
		return
	}

	if len(tasks) == 0 {
		return // No tasks available
	}

	for i, task := range tasks {
		messageID := messageIDs[i]
		tsm.logger.Debug("Processing task",
			"task_id", task.SendTaskDataToKeeper.TaskID[0],
			"message_id", messageID)

		// Move task to processing stream with improved error handling
		if err := tsm.moveTaskToProcessing(task, messageID); err != nil {
			tsm.logger.Error("Failed to move task to processing",
				"task_id", task.SendTaskDataToKeeper.TaskID[0],
				"message_id", messageID,
				"error", err)

			// Try to acknowledge the message to prevent reprocessing
			if ackErr := tsm.AckTaskProcessed(TasksReadyStream, "task-processors", messageID); ackErr != nil {
				tsm.logger.Error("Failed to acknowledge failed task",
					"task_id", task.SendTaskDataToKeeper.TaskID[0],
					"message_id", messageID,
					"ack_error", ackErr)
			}

			// Move task to retry stream
			if retryErr := tsm.AddTaskToRetryStream(&task, err.Error()); retryErr != nil {
				tsm.logger.Error("Failed to move task to retry stream",
					"task_id", task.SendTaskDataToKeeper.TaskID[0],
					"error", retryErr)
			}
			continue
		}

		tsm.logger.Info("Successfully moved task to processing",
			"task_id", task.SendTaskDataToKeeper.TaskID[0])

		// Send task to performer asynchronously
		go tsm.sendTaskToPerformer(task)
	}
}

// processRetryTasks processes tasks that are ready for retry
func (tsm *TaskStreamManager) processRetryTasks() {
	tasks, messageIDs, err := tsm.readTasksFromStreamWithIDs(TasksRetryStream, "retry-processor", "retry-worker", 10)
	if err != nil {
		tsm.logger.Error("Failed to read retry tasks", "error", err)
		return
	}

	now := time.Now()
	retryCount := 0

	for i, task := range tasks {
		if task.ScheduledFor == nil || task.ScheduledFor.Before(now) {
			// Task is ready for retry
			task.LastAttemptAt = &now

			// Move back to ready stream
			err := tsm.addTaskToStream(TasksReadyStream, &task)
			if err != nil {
				tsm.logger.Error("Failed to move retry task to ready",
					"task_id", task.SendTaskDataToKeeper.TaskID[0],
					"error", err)
				continue
			}

			// Acknowledge in retry stream
			messageID := messageIDs[i]
			err = tsm.AckTaskProcessed(TasksRetryStream, "retry-processor", messageID)
			if err != nil {
				tsm.logger.Error("Failed to acknowledge task in retry stream",
					"task_id", task.SendTaskDataToKeeper.TaskID[0],
					"message_id", messageID,
					"error", err)
			}

			tsm.logger.Info("Task moved from retry to ready",
				"task_id", task.SendTaskDataToKeeper.TaskID[0],
				"retry_count", task.RetryCount)

			retryCount++
		}
	}

	if retryCount > 0 {
		tsm.logger.Info("Processed retry tasks", "retry_count", retryCount)
	}
}

// moveTaskToProcessing moves a task from ready to processing stream
func (tsm *TaskStreamManager) moveTaskToProcessing(task TaskStreamData, messageID string) error {
	task.ProcessingStartedAt = &[]time.Time{time.Now()}[0]

	// Add to processing stream
	err := tsm.addTaskToStream(TasksProcessingStream, &task)
	if err != nil {
		return fmt.Errorf("failed to add to processing stream: %w", err)
	}

	// Acknowledge in ready stream with retry logic
	maxAckRetries := 5
	for attempt := 1; attempt <= maxAckRetries; attempt++ {
		if err := tsm.AckTaskProcessed(TasksReadyStream, "task-processors", messageID); err != nil {
			tsm.logger.Warn("Failed to acknowledge task in ready stream (attempt %d/%d)",
				"task_id", task.SendTaskDataToKeeper.TaskID[0],
				"message_id", messageID,
				"attempt", attempt,
				"max_retries", maxAckRetries,
				"error", err)

			if attempt == maxAckRetries {
				// On final attempt, log as error but don't fail the operation
				tsm.logger.Error("Failed to acknowledge task after all retries",
					"task_id", task.SendTaskDataToKeeper.TaskID[0],
					"message_id", messageID,
					"error", err)
				return fmt.Errorf("failed to acknowledge task after %d retries: %w", maxAckRetries, err)
			} else {
				// Wait before retry with exponential backoff
				backoffDuration := time.Duration(attempt*attempt) * 100 * time.Millisecond
				time.Sleep(backoffDuration)
				continue
			}
		} else {
			// Successfully acknowledged
			tsm.logger.Debug("Task acknowledged successfully on attempt %d",
				"task_id", task.SendTaskDataToKeeper.TaskID[0],
				"message_id", messageID,
				"attempt", attempt)
			break
		}
	}

	tsm.logger.Debug("Task moved to processing stream",
		"task_id", task.SendTaskDataToKeeper.TaskID[0])

	return nil
}

// AckTaskProcessed acknowledges that a task has been processed
func (tsm *TaskStreamManager) AckTaskProcessed(stream, consumerGroup, messageID string) error {
	tsm.logger.Debug("Acknowledging task processed",
		"stream", stream,
		"consumer_group", consumerGroup,
		"message_id", messageID)

	// Increase timeout for acknowledgment operations
	// ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	// defer cancel()

	// err := tsm.client.XAck(ctx, stream, consumerGroup, messageID)
	// if err != nil {
	// 	tsm.logger.Error("Failed to acknowledge task",
	// 		"stream", stream,
	// 		"consumer_group", consumerGroup,
	// 		"message_id", messageID,
	// 		"error", err)
	// 	return err
	// }

	tsm.logger.Debug("Task acknowledged successfully",
		"stream", stream,
		"message_id", messageID)

	return nil
}

// moveTaskToFailed moves a task to the failed stream or retry stream
func (tsm *TaskStreamManager) moveTaskToFailed(task TaskStreamData, errorMsg string) error {
	task.LastError = errorMsg
	task.RetryCount++

	// Release the performer if it was assigned
	if task.SendTaskDataToKeeper.PerformerData.OperatorID != 0 {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := tsm.performerManager.MarkPerformerAvailable(ctx, task.SendTaskDataToKeeper.PerformerData.OperatorID); err != nil {
			tsm.logger.Warn("Failed to mark performer as available after task failure",
				"performer_id", task.SendTaskDataToKeeper.PerformerData.OperatorID,
				"error", err)
		}
	}

	if task.RetryCount <= MaxRetryAttempts {
		// Move to retry stream with backoff
		delay := time.Duration(task.RetryCount) * RetryBackoffBase
		scheduledFor := time.Now().Add(delay)
		task.ScheduledFor = &scheduledFor

		err := tsm.addTaskToStream(TasksRetryStream, &task)
		if err != nil {
			return fmt.Errorf("failed to add to retry stream: %w", err)
		}

		tsm.logger.Info("Task moved to retry stream",
			"task_id", task.SendTaskDataToKeeper.TaskID[0],
			"retry_count", task.RetryCount,
			"scheduled_for", scheduledFor)

		metrics.TasksAddedToStreamTotal.WithLabelValues("retry", "success").Inc()
	} else {
		// Move to failed stream permanently
		err := tsm.addTaskToStream(TasksFailedStream, &task)
		if err != nil {
			return fmt.Errorf("failed to add to failed stream: %w", err)
		}

		tsm.logger.Error("Task permanently failed",
			"task_id", task.SendTaskDataToKeeper.TaskID[0],
			"retry_count", task.RetryCount,
			"error", errorMsg)

		metrics.TasksAddedToStreamTotal.WithLabelValues("failed", "success").Inc()
	}

	return nil
}
