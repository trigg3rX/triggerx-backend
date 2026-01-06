package tasks

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/trigg3rX/triggerx-backend/internal/taskmonitor/clients/notify"
	"github.com/trigg3rX/triggerx-backend/internal/taskmonitor/metrics"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
)

// StartTimeoutWorker monitors processing tasks for timeouts
func (tsm *TaskStreamManager) StartTimeoutWorker(ctx context.Context) {
	tsm.logger.Info(ctx, "Starting task timeout worker")

	ticker := time.NewTicker(30 * time.Second) // Check timeouts every 30 seconds
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			tsm.logger.Info(ctx, "Timeout worker stopping")
			return
		case <-ticker.C:
			// tsm.logger.Debug("Timeout worker checking for timed out tasks")
			tsm.checkDispatchedTimeouts(ctx)
		}
	}
}

// checkDispatchedTimeouts checks for tasks that have been dispatched too long
func (tsm *TaskStreamManager) checkDispatchedTimeouts(ctx context.Context) {
	// Span: Timeout check cycle
	ctx, span := tsm.tracer.Start(ctx, "worker.timeout_check",
		observability.WithSpanKind(trace.SpanKindInternal),
	)
	defer span.End()

	// Get expired tasks efficiently using sorted set
	expiredTaskIDs, err := tsm.expirationManager.GetExpiredTasks(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to get expired tasks")
		tsm.logger.Error(ctx, "Failed to get expired tasks", observability.Error(err))
		return
	}

	span.SetAttributes(attribute.Int("expired_tasks.count", len(expiredTaskIDs)))

	if len(expiredTaskIDs) == 0 {
		return
	}

	tsm.logger.Info(ctx, "Found expired tasks", observability.Int("expired_count", len(expiredTaskIDs)))

	// Process each expired task
	processedCount := 0
	var processedTaskIDs []int64

	for _, taskID := range expiredTaskIDs {
		// Span: Process individual expired task
		taskCtx, taskSpan := tsm.tracer.Start(ctx, "worker.process_expired_task",
			observability.WithSpanKind(trace.SpanKindInternal),
			observability.WithAttributes(
				attribute.Int64("task.id", taskID),
			),
		)

		// Find the task using the efficient index lookup
		task, messageID, err := tsm.taskIndex.FindTaskByID(taskCtx, taskID)
		if err != nil {
			taskSpan.RecordError(err)
			taskSpan.SetStatus(codes.Error, "failed to find task")
			taskSpan.End()
			tsm.logger.Error(taskCtx, "Failed to find expired task",
				observability.Int64("task_id", taskID),
				observability.Error(err))
			// Still remove from timeout tracking to prevent reprocessing
			processedTaskIDs = append(processedTaskIDs, taskID)
			continue
		}

		taskSpan.AddEvent("task.found")

		// Log task timeout with dispatched_at if available
		logFields := []observability.Field{
			observability.Int64("task_id", taskID),
			observability.Time("created_at", task.CreatedAt),
		}
		if task.DispatchedAt != nil {
			logFields = append(logFields, observability.Time("dispatched_at", *task.DispatchedAt))
			taskSpan.SetAttributes(attribute.String("task.dispatched_at", task.DispatchedAt.Format(time.RFC3339)))
		}
		tsm.logger.Warn(taskCtx, "Task timeout detected", logFields...)

		// Move to failed stream
		if err := tsm.moveTaskToFailed(taskCtx, *task, "dispatched timeout"); err != nil {
			taskSpan.RecordError(err)
			taskSpan.SetStatus(codes.Error, "failed to move task to failed")
			taskSpan.End()
			tsm.logger.Error(taskCtx, "Failed to handle timeout task",
				observability.Int64("task_id", taskID),
				observability.Error(err))
			continue // Don't acknowledge if we failed to move to failed stream
		}

		taskSpan.AddEvent("task.moved_to_failed")

		// Acknowledge the timed-out task if we have the messageID
		if messageID != "" {
			err := tsm.AckTaskProcessed(taskCtx, StreamTaskDispatched, "timeout-checker", messageID)
			if err != nil {
				taskSpan.RecordError(err)
				tsm.logger.Error(taskCtx, "Failed to acknowledge timed-out task",
					observability.Int64("task_id", taskID),
					observability.String("message_id", messageID),
					observability.Error(err))
			} else {
				taskSpan.AddEvent("task.acknowledged")
				tsm.logger.Info(taskCtx, "Task timeout processed and acknowledged successfully",
					observability.Int64("task_id", taskID),
					observability.String("message_id", messageID))
			}
		}

		// Remove from task index since it's been processed
		err = tsm.taskIndex.RemoveTaskIndex(taskCtx, taskID)
		if err != nil {
			tsm.logger.Warn(taskCtx, "failed to remove timed-out task from index",
				observability.Int64("task_id", taskID),
				observability.Error(err))
		}

		processedCount++
		processedTaskIDs = append(processedTaskIDs, taskID)

		err = tsm.dbClient.UpdateTaskFailed(taskCtx, taskID)
		if err != nil {
			taskSpan.RecordError(err)
			tsm.logger.Error(taskCtx, "Failed to update task failed",
				observability.Int64("task_id", taskID),
				observability.Error(err))
		} else {
			taskSpan.AddEvent("db.updated")
		}

		// Notify user on failure
		if tsm.notifier != nil {
			email, e := tsm.dbClient.GetUserEmailByTaskID(taskCtx, taskID)
			if e != nil {
				tsm.logger.Warn(taskCtx, "Could not fetch user email for task failure",
					observability.Int64("task_id", taskID),
					observability.Error(e))
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
					tsm.logger.Warn(taskCtx, "Failed to notify user for task failure",
						observability.String("email", email),
						observability.Int64("task_id", taskID),
						observability.Error(err))
				} else {
					taskSpan.AddEvent("user.notified")
				}
			}
		}

		taskSpan.SetStatus(codes.Ok, "")
		taskSpan.End()
	}

	// Remove all processed tasks from timeout tracking in batch
	if len(processedTaskIDs) > 0 {
		err = tsm.expirationManager.RemoveMultipleTaskTimeouts(ctx, processedTaskIDs)
		if err != nil {
			span.RecordError(err)
			tsm.logger.Error(ctx, "Failed to remove processed tasks from timeout tracking",
				observability.Int("task_count", len(processedTaskIDs)),
				observability.Error(err))
		} else {
			tsm.logger.Info(ctx, "Removed processed tasks from timeout tracking",
				observability.Int("task_count", len(processedTaskIDs)))
		}
	}

	span.SetAttributes(attribute.Int("processed_count", processedCount))
	if processedCount > 0 {
		tsm.logger.Info(ctx, "Processed task timeouts", observability.Int("processed_count", processedCount))
	}
}

// AckTaskProcessed acknowledges that a task has been processed and optionally deletes it from the stream
func (tsm *TaskStreamManager) AckTaskProcessed(ctx context.Context, stream, consumerGroup, messageID string) error {
	return tsm.AckTaskProcessedWithDelete(ctx, stream, consumerGroup, messageID, true)
}

// AckTaskProcessedWithDelete acknowledges that a task has been processed and optionally deletes it from the stream
func (tsm *TaskStreamManager) AckTaskProcessedWithDelete(ctx context.Context, stream, consumerGroup, messageID string, deleteFromStream bool) error {
	tsm.logger.Debug(ctx, "Acknowledging task processed",
		observability.String("stream", stream),
		observability.String("consumer_group", consumerGroup),
		observability.String("message_id", messageID),
		observability.Bool("delete_from_stream", deleteFromStream))

	// Increase timeout for acknowledgment operations
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// First acknowledge the message (removes from PEL)
	err := tsm.redisClient.XAck(ctx, stream, consumerGroup, messageID)
	if err != nil {
		tsm.logger.Error(ctx, "Failed to acknowledge task",
			observability.String("stream", stream),
			observability.String("consumer_group", consumerGroup),
			observability.String("message_id", messageID),
			observability.Error(err))
		return err
	}

	// Remove from expiration tracking since the message has been processed
	err = tsm.expirationManager.RemoveMessageExpiration(ctx, stream, messageID)
	if err != nil {
		tsm.logger.Warn(ctx, "Failed to remove stream entry expiration after acknowledgment",
			observability.String("stream", stream),
			observability.String("consumer_group", consumerGroup),
			observability.String("message_id", messageID),
			observability.Error(err))
		// Don't return error - message was acknowledged, just couldn't remove from expiration tracking
	}

	// If deleteFromStream is true, delete the message from the stream itself
	// This is important because XACK only removes from PEL, not from the stream
	if deleteFromStream {
		deleted, err := tsm.redisClient.XDel(ctx, stream, messageID)
		if err != nil {
			tsm.logger.Warn(ctx, "Failed to delete message from stream after acknowledgment",
				observability.String("stream", stream),
				observability.String("consumer_group", consumerGroup),
				observability.String("message_id", messageID),
				observability.Error(err))
			// Don't return error - message was acknowledged, just couldn't delete
			// The message will be cleaned up by periodic trimming or expiration
		} else if deleted > 0 {
			tsm.logger.Debug(ctx, "Task acknowledged and deleted from stream successfully",
				observability.String("stream", stream),
				observability.String("consumer_group", consumerGroup),
				observability.String("message_id", messageID))
		} else {
			tsm.logger.Debug(ctx, "Task acknowledged but message not found in stream (may have been already deleted)",
				observability.String("stream", stream),
				observability.String("consumer_group", consumerGroup),
				observability.String("message_id", messageID))
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

	tsm.logger.Error(ctx, "Task permanently failed",
		observability.Int64("task_id", task.SendTaskDataToKeeper.TaskID[0]),
		observability.Int("retry_count", task.RetryCount),
		observability.String("error", errorMsg))
	if metrics.TasksAddedToStreamTotal != nil {
		metrics.TasksAddedToStreamTotal.WithLabelValues("failed", "success").Inc(ctx)
	}

	return nil
}
