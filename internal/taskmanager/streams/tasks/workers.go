package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/trigg3rX/triggerx-backend/internal/taskmanager/metrics"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// StartTaskProcessor starts the main task processing loop with consumer groups
func (tsm *TaskStreamManager) StartTaskProcessor(ctx context.Context, consumerName string) {
	tsm.logger.Info("Starting task processor", "consumer_name", consumerName)

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
	tasks, messageIDs, err := tsm.readTasksFromStreamWithIDs(TasksReadyStream, "task-processors", consumerName, 10)
	if err != nil {
		tsm.logger.Error("Failed to read tasks from ready stream", "error", err)
		return
	}

	if len(tasks) == 0 {
		return // No tasks available
	}

	tsm.logger.Debug("Processing tasks from ready stream", "count", len(tasks))

	for i, task := range tasks {
		messageID := messageIDs[i]

		// Move task to processing stream
		if err := tsm.moveTaskToProcessing(task, messageID); err != nil {
			tsm.logger.Error("Failed to move task to processing",
				"task_id", task.SendTaskDataToKeeper.TaskID[0],
				"error", err)
			continue
		}

		// Send task to performer
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

	// Acknowledge in ready stream
	if err := tsm.AckTaskProcessed(TasksReadyStream, "task-processors", messageID); err != nil {
		tsm.logger.Warn("Failed to acknowledge task in ready stream",
			"task_id", task.SendTaskDataToKeeper.TaskID[0],
			"message_id", messageID,
			"error", err)
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

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := tsm.client.XAck(ctx, stream, consumerGroup, messageID)
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

// sendTaskToPerformer sends the task to the assigned performer
func (tsm *TaskStreamManager) sendTaskToPerformer(task TaskStreamData) {
	taskID := task.SendTaskDataToKeeper.TaskID[0]

	tsm.logger.Info("Sending task to performer", "task_id", taskID)

	// Send to aggregator/performer using existing method
	jsonData, err := json.Marshal(task.SendTaskDataToKeeper)
	if err != nil {
		tsm.logger.Errorf("Failed to marshal batch task data: %v", err)
		return
	}
	dataBytes := []byte(jsonData)

	broadcastDataForPerformer := types.BroadcastDataForPerformer{
		TaskID:           task.SendTaskDataToKeeper.TaskID[0],
		TaskDefinitionID: task.SendTaskDataToKeeper.TargetData[0].TaskDefinitionID,
		PerformerAddress: task.SendTaskDataToKeeper.PerformerData.KeeperAddress,
		Data:             dataBytes,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	success, err := tsm.aggClient.SendTaskToPerformer(ctx, &broadcastDataForPerformer)
	if err != nil {
		tsm.logger.Error("Failed to send task to performer",
			"task_id", taskID,
			"error", err)

		// Move task to failed stream
		if moveErr := tsm.moveTaskToFailed(task, err.Error()); moveErr != nil {
			tsm.logger.Error("Failed to move task to failed stream",
				"task_id", taskID,
				"error", moveErr)
		}
		return
	}

	if success {
		tsm.logger.Info("Task sent to performer successfully", "task_id", taskID)
		metrics.TasksAddedToStreamTotal.WithLabelValues("processing", "success").Inc()
	} else {
		tsm.logger.Warn("Task sending to performer was not successful", "task_id", taskID)
		if moveErr := tsm.moveTaskToFailed(task, "performer send failed"); moveErr != nil {
			tsm.logger.Error("Failed to move task to failed stream",
				"task_id", taskID,
				"error", moveErr)
		}
	}
}

// moveTaskToFailed moves a task to the failed stream or retry stream
func (tsm *TaskStreamManager) moveTaskToFailed(task TaskStreamData, errorMsg string) error {
	task.LastError = errorMsg
	task.RetryCount++

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
