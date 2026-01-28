package tasks

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/trigg3rX/triggerx-backend/internal/taskmonitor/metrics"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
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
			tsm.checkExecutedTimeouts(ctx)
		}
	}
}

// checkExecutedTimeouts checks for executed tasks that have been pending validation too long
func (tsm *TaskStreamManager) checkExecutedTimeouts(ctx context.Context) {
	// Span: Timeout check cycle
	ctx, span := tsm.tracer.Start(ctx, "worker.timeout_check",
		observability.WithSpanKind(trace.SpanKindInternal),
	)
	defer span.End()

	// Get expired tasks efficiently using sorted set
	expiredTaskIDs, err := tsm.expirationManager.GetExpiredExecutedTasks(ctx)
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

		// CRITICAL: Check if task is already validated FIRST - if so, don't rebroadcast
		// Task in validated stream should never be rebroadcasted, regardless of timeout tracking state
		validatedTask, _, validatedErr := tsm.taskIndex.FindTaskByIDInStream(taskCtx, taskID, types.StreamTaskValidated)
		if validatedErr == nil && validatedTask != nil {
			// Task was already validated, just clean up timeout tracking
			taskSpan.AddEvent("task.already_validated")
			tsm.logger.Debug(taskCtx, "Expired task was already validated, cleaning up timeout tracking (should not rebroadcast)",
				observability.Int64("task_id", taskID))
			// Remove from timeout tracking since task is validated
			if err := tsm.expirationManager.RemoveExecutedTaskTimeout(taskCtx, taskID); err != nil {
				tsm.logger.Warn(taskCtx, "Failed to remove validated task from timeout tracking",
					observability.Int64("task_id", taskID),
					observability.Error(err))
			}
			processedTaskIDs = append(processedTaskIDs, taskID)
			taskSpan.SetStatus(codes.Ok, "task already validated")
			taskSpan.End()
			continue
		}

		// Find the task in executed stream using the efficient index lookup
		task, messageID, err := tsm.taskIndex.FindTaskByIDInStream(taskCtx, taskID, types.StreamTaskExecuted)
		if err != nil {

			// Check if task is in failed stream (already processed)
			failedTask, _, failedErr := tsm.taskIndex.FindTaskByIDInStream(taskCtx, taskID, types.StreamTaskFailed)
			if failedErr == nil && failedTask != nil {
				// Task was already moved to failed, just clean up timeout tracking
				taskSpan.AddEvent("task.already_failed")
				tsm.logger.Debug(taskCtx, "Expired task was already moved to failed, cleaning up timeout tracking",
					observability.Int64("task_id", taskID))
				processedTaskIDs = append(processedTaskIDs, taskID)
				taskSpan.SetStatus(codes.Ok, "task already failed")
				taskSpan.End()
				continue
			}

			// Task not found in any stream - might have been cleaned up or never existed
			// This could happen if:
			// 1. Task was validated/moved before timeout worker processed it
			// 2. Task index is out of sync
			// 3. Task was never actually added to executed stream
			taskSpan.RecordError(err)
			taskSpan.SetStatus(codes.Error, "task not found in executed stream")
			tsm.logger.Warn(taskCtx, "Failed to find expired executed task in any stream, cleaning up timeout tracking",
				observability.Int64("task_id", taskID),
				observability.Error(err))
			// Still remove from timeout tracking to prevent reprocessing
			processedTaskIDs = append(processedTaskIDs, taskID)
			taskSpan.End()
			continue
		}

		taskSpan.AddEvent("task.found")

		// Log task timeout with executed_at if available
		logFields := []observability.Field{
			observability.Int64("task_id", taskID),
			observability.Time("created_at", task.CreatedAt),
		}
		if task.ExecutedAt != nil {
			logFields = append(logFields, observability.Time("executed_at", *task.ExecutedAt))
			taskSpan.SetAttributes(attribute.String("task.executed_at", task.ExecutedAt.Format(time.RFC3339)))
		}
		tsm.logger.Warn(taskCtx, "Executed task validation timeout detected", logFields...)

		// Check rebroadcast limit (max 3 attempts to prevent infinite rebroadcasts)
		maxRebroadcastAttempts := 3
		rebroadcastCount := task.RebroadcastCount

		if tsm.keeperClient != nil && rebroadcastCount < maxRebroadcastAttempts {
			rebroadcastErr := tsm.keeperClient.RebroadcastTask(taskCtx, taskID)
			if rebroadcastErr != nil {
				taskSpan.RecordError(rebroadcastErr)
				tsm.logger.Error(taskCtx, "Failed to rebroadcast task",
					observability.Int64("task_id", taskID),
					observability.Int("rebroadcast_count", rebroadcastCount),
					observability.Error(rebroadcastErr))
				// Continue to move to failed even if rebroadcast fails
			} else {
				// Update rebroadcast count and timestamp
				now := time.Now()
				task.RebroadcastCount = rebroadcastCount + 1
				task.LastRebroadcastAt = &now

				// Acknowledge the old message to remove it from PEL
				if messageID != "" {
					if err := tsm.AckTaskProcessed(taskCtx, types.StreamTaskExecuted, "timeout-checker", messageID); err != nil {
						tsm.logger.Warn(taskCtx, "Failed to acknowledge old message after rebroadcast",
							observability.Int64("task_id", taskID),
							observability.String("message_id", messageID),
							observability.Error(err))
					}

					// Add new entry to stream with updated rebroadcast info
					newMessageID, err := tsm.addTaskToStreamWithMessageID(taskCtx, types.StreamTaskExecuted, task)
					if err != nil {
						tsm.logger.Warn(taskCtx, "Failed to add task with rebroadcast info to stream",
							observability.Int64("task_id", taskID),
							observability.Error(err))
					} else if newMessageID != "" {
						// Update task index to point to new message
						if err := tsm.taskIndex.StoreTaskIndex(taskCtx, taskID, newMessageID); err != nil {
							tsm.logger.Warn(taskCtx, "Failed to update task index after rebroadcast",
								observability.Int64("task_id", taskID),
								observability.Error(err))
						}
					}
				}

				taskSpan.AddEvent("task.rebroadcast.initiated")
				tsm.logger.Info(taskCtx, "Task rebroadcast initiated",
					observability.Int64("task_id", taskID),
					observability.Int("rebroadcast_count", task.RebroadcastCount),
					observability.Int("max_attempts", maxRebroadcastAttempts))
				// Remove from current timeout tracking
				processedTaskIDs = append(processedTaskIDs, taskID)
				// Re-add to timeout tracking with a new timeout
				if err := tsm.expirationManager.AddExecutedTaskTimeout(taskCtx, taskID, types.TasksExecutedTTL); err != nil {
					tsm.logger.Warn(taskCtx, "Failed to re-add task to timeout tracking after rebroadcast",
						observability.Int64("task_id", taskID),
						observability.Error(err))
				} else {
					taskSpan.AddEvent("task.timeout_reset")
				}
				taskSpan.SetStatus(codes.Ok, "rebroadcast initiated")
				taskSpan.End()
				continue
			}
		} else if rebroadcastCount >= maxRebroadcastAttempts {
			tsm.logger.Warn(taskCtx, "Task rebroadcast limit reached, moving to failed",
				observability.Int64("task_id", taskID),
				observability.Int("rebroadcast_count", rebroadcastCount),
				observability.Int("max_attempts", maxRebroadcastAttempts))
			taskSpan.AddEvent("task.rebroadcast_limit_reached")
		}

		// Move to failed stream (task didn't get validated on-chain in time and rebroadcast failed or wasn't attempted)
		if err := tsm.moveTaskToFailed(taskCtx, *task, "validation timeout - task not confirmed on-chain"); err != nil {
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
			err := tsm.AckTaskProcessed(taskCtx, types.StreamTaskExecuted, "timeout-checker", messageID)
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

		err = tsm.taskRepo.UpdateTaskAttestationTimeoutFailure(taskCtx, taskID, "validation timeout - task not confirmed on-chain")
		if err != nil {
			taskSpan.RecordError(err)
			tsm.logger.Error(taskCtx, "Failed to update task failed",
				observability.Int64("task_id", taskID),
				observability.Error(err))
		} else {
			taskSpan.AddEvent("db.updated")
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

	// First acknowledge the message (removes from PEL)
	err := tsm.redisClient.AcknowledgeTask(ctx, stream, consumerGroup, messageID)
	if err != nil {
		tsm.logger.Error(ctx, "Failed to acknowledge task",
			observability.String("stream", stream),
			observability.String("consumer_group", consumerGroup),
			observability.String("message_id", messageID),
			observability.Error(err))
		return err
	}

	// Remove from expiration tracking since the message has been processed
	err = tsm.redisClient.RemoveStreamEntryExpiration(ctx, stream, messageID)
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
		deleted, err := tsm.redisClient.DeleteTaskFromStream(ctx, stream, messageID)
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
func (tsm *TaskStreamManager) moveTaskToFailed(ctx context.Context, task types.TaskStreamData, errorMsg string) error {
	task.LastError = errorMsg
	task.RetryCount++

	// Move to failed stream permanently
	err := tsm.addTaskToStream(ctx, types.StreamTaskFailed, &task)
	if err != nil {
		return fmt.Errorf("failed to add to failed stream: %w", err)
	}

	tsm.logger.Error(ctx, "Task permanently failed",
		observability.Int64("task_id", task.SendTaskDataToKeeper.TaskID),
		observability.Int("retry_count", task.RetryCount),
		observability.String("error", errorMsg))
	if metrics.TasksAddedToStreamTotal != nil {
		metrics.TasksAddedToStreamTotal.WithLabelValues("failed", "success").Inc(ctx)
	}

	return nil
}
