package events

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/trigg3rX/triggerx-backend/internal/taskmonitor/core/tasks"
	"github.com/trigg3rX/triggerx-backend/internal/taskmonitor/database/repository"
	"github.com/trigg3rX/triggerx-backend/internal/taskmonitor/rpc/clients/notify"
	"github.com/trigg3rX/triggerx-backend/pkg/ipfs"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// TaskEventHandler handles task-related events
type TaskEventHandler struct {
	logger            observability.Logger
	tracer            observability.Tracer
	taskRepo          repository.TaskRepository
	ipfsClient        ipfs.IPFSClient
	taskStreamManager *tasks.TaskStreamManager
	notifier          notify.Notifier
	traceRegistry     *sync.Map // taskID -> traceID for correlation between execution and validation
}

// NewTaskEventHandler creates a new TaskEventHandler instance
func NewTaskEventHandler(logger observability.Logger, tracer observability.Tracer, taskRepo repository.TaskRepository, ipfsClient ipfs.IPFSClient, taskStreamManager *tasks.TaskStreamManager, notifier notify.Notifier) *TaskEventHandler {
	return &TaskEventHandler{
		logger:            logger,
		tracer:            tracer,
		taskRepo:          taskRepo,
		ipfsClient:        ipfsClient,
		taskStreamManager: taskStreamManager,
		notifier:          notifier,
		traceRegistry:     &sync.Map{},
	}
}

// ProcessConsensusEventFromIPFS processes consensus events with IPFS data directly from eventmonitor
// This is the new flow where eventmonitor fetches IPFS data and passes it with trace context
func (h *TaskEventHandler) ProcessConsensusEventFromIPFS(ctx context.Context, txHash string, isAccepted bool, ipfsData *types.IPFSData, ipfsCID string) error {
	if ipfsData == nil {
		return fmt.Errorf("ipfs data is nil")
	}

	// Get task ID from ActionData (single task ID per execution)
	taskID := int64(0)
	if ipfsData.ActionData != nil {
		taskID = ipfsData.ActionData.TaskID
	}

	// Get task definition ID from PerformerData in TaskData
	taskDefinitionID := 0
	if ipfsData.TaskData != nil && len(ipfsData.TaskData.TargetData) > 0 {
		taskDefinitionID = ipfsData.TaskData.TargetData[0].TaskDefinitionID
	}

	h.logger.Info(ctx, "Processing consensus event from IPFS data",
		observability.Int64("task_id", taskID),
		observability.String("tx_hash", txHash),
		observability.Bool("is_accepted", isAccepted))

	// Use TotalFee directly as string (Wei) - no conversion needed
	taskOpxCostWei := "0"
	if ipfsData.ActionData != nil && ipfsData.ActionData.TotalFee != "" {
		taskOpxCostWei = ipfsData.ActionData.TotalFee
	}

	// Build TaskSubmissionData from IPFS data
	taskData := &types.TaskSubmissionData{
		TaskID:               taskID,
		TaskDefinitionID:     taskDefinitionID,
		IsAccepted:           isAccepted,
		TaskSubmissionTxHash: txHash,
		TaskOpxActualCost:    taskOpxCostWei,
	}

	// Populate from ActionData
	if ipfsData.ActionData != nil {
		taskData.ExecutionTxHash = ipfsData.ActionData.ActionTxHash
		taskData.ExecutedAt = ipfsData.ActionData.ExecutionTimestamp
		taskData.ConvertedArguments = ipfsData.ActionData.ConvertedArguments
	}

	// Populate from ProofData
	if ipfsData.ProofData != nil {
		taskData.ProofOfTask = ipfsData.ProofData.ProofOfTask
	}

	// Populate from PerformerSignature
	if ipfsData.PerformerSignature != nil {
		taskData.PerformerAddress = ipfsData.PerformerSignature.PerformerSigningAddress
	}

	// Create span for task processing
	ctx, span := h.tracer.Start(ctx, "task.monitor.process",
		observability.WithSpanKind(trace.SpanKindInternal),
		observability.WithAttributes(
			attribute.Int64("task.id", taskID),
			attribute.String("task.submission.tx_hash", txHash),
			attribute.Bool("task.is_accepted", isAccepted),
		),
	)
	defer span.End()

	span.AddEvent("ipfs.data.received", observability.WithEventAttributes(
		attribute.String("tx.hash", txHash),
	))

	// Move task from dispatched to completed stream
	streamHandled := true
	if err := h.moveTaskToCompleted(ctx, taskID); err != nil {
		span.RecordError(err, observability.WithErrorAttributes(
			attribute.String("error.type", "stream_move_failed"),
		))
		h.logger.Error(ctx, "Failed to move task to validated stream", observability.Error(err))
		streamHandled = false
		// Stream move failure is critical - task should be in validated stream
		// But continue to attempt DB update anyway
	}

	// Update task submission data in database
	if err := h.taskRepo.UpdateTaskSubmissionData(ctx, *taskData); err != nil {
		span.RecordError(err, observability.WithErrorAttributes(
			attribute.String("error.type", "database_update_failed"),
		))
		h.logger.Error(ctx, "Failed to update task submission data in database", observability.Error(err))
		// Don't return error - task is already validated in stream
		// DB update can be retried later if needed, but task should not be rebroadcasted
	} else {
		span.SetStatus(codes.Ok, "task validated and database updated")

		// Schedule IPFS file deletion after 6 hours delay
		// This happens after data is downloaded, DB is updated, and stream is handled
		// Only schedule deletion if both stream handling and DB update succeeded
		if streamHandled && ipfsCID != "" {
			h.scheduleIPFSDeletion(ctx, ipfsCID, taskID)
		} else if !streamHandled {
			h.logger.Warn(ctx, "Skipping IPFS deletion scheduling - stream handling failed",
				observability.String("ipfs_cid", ipfsCID),
				observability.Int64("task_id", taskID))
		}
	}

	span.AddEvent("task.data.updated", observability.WithEventAttributes(
		attribute.String("database.table", "tasks"),
	))

	// For agent script jobs (TaskDefinitionID = 7, 8, 9), update storage
	if (taskData.TaskDefinitionID == 7 || taskData.TaskDefinitionID == 8 || taskData.TaskDefinitionID == 9) && ipfsData.ActionData != nil && ipfsData.ActionData.StorageUpdates != nil && len(ipfsData.ActionData.StorageUpdates) > 0 {
		jobID, err := h.taskRepo.GetJobIDByTaskID(ctx, taskID)
		if err != nil {
			span.RecordError(err, observability.WithErrorAttributes(
				attribute.String("error.type", "job_id_lookup_failed"),
			))
			h.logger.Error(ctx, "Failed to get job ID for task", observability.Int64("task_id", taskID), observability.Error(err))
		} else {
			if err := h.taskRepo.UpdateScriptStorage(ctx, jobID, ipfsData.ActionData.StorageUpdates); err != nil {
				span.RecordError(err, observability.WithErrorAttributes(
					attribute.String("error.type", "storage_update_failed"),
				))
				h.logger.Error(ctx, "Failed to update script storage for job", observability.String("job_id", jobID), observability.Error(err))
			} else {
				h.logger.Info(ctx, "Successfully updated storage keys for job", observability.Int("storage_keys", len(ipfsData.ActionData.StorageUpdates)), observability.String("job_id", jobID))
			}
		}
	}

	// Update keeper points in database
	if err := h.taskRepo.UpdateKeeperPointsInDatabase(ctx, *taskData); err != nil {
		span.RecordError(err, observability.WithErrorAttributes(
			attribute.String("error.type", "keeper_points_update_failed"),
		))
		h.logger.Error(ctx, "Failed to update keeper points in database", observability.Error(err))
		// Don't return, continue processing
	}

	// Notify user about task completion/rejection
	if h.notifier != nil {
		email, err := h.taskRepo.GetUserEmailByTaskID(ctx, taskID)
		if err != nil {
			h.logger.Warn(ctx, "Could not fetch user email for task", observability.Int64("task_id", taskID), observability.Error(err))
		} else if email != "" {
			payload := notify.TaskStatusPayload{
				TaskID:          taskID,
				JobID:           0,
				Status:          "completed",
				IsAccepted:      isAccepted,
				SubmissionTx:    txHash,
				ExecutionTxHash: taskData.ExecutionTxHash,
				ProofOfTask:     taskData.ProofOfTask,
				OccurredAt:      time.Now(),
			}
			if !isAccepted {
				payload.Status = "failed"
			}
			if err := h.notifier.NotifyTaskStatus(context.Background(), email, payload); err != nil {
				h.logger.Warn(ctx, "Failed to notify user", observability.String("email", email), observability.Int64("task_id", taskID), observability.Error(err))
			}
		}
	}

	span.SetStatus(codes.Ok, "consensus event processed")
	h.logger.Info(ctx, "Consensus event processed successfully",
		observability.Int64("task_id", taskID),
		observability.String("tx_hash", txHash),
		observability.Bool("is_accepted", isAccepted))

	return nil
}

// moveTaskToCompleted moves a task from executed (or dispatched) to validated stream
func (h *TaskEventHandler) moveTaskToCompleted(ctx context.Context, taskID int64) error {
	h.logger.Info(ctx, "Moving task to validated stream", observability.Int64("task_id", taskID))

	// Try to find task in executed stream first (most common case after execution)
	var task *types.TaskStreamData
	var messageID string
	var err error
	task, messageID, err = h.taskStreamManager.FindTaskByIDInStream(ctx, taskID, types.StreamTaskExecuted)
	if err != nil {
		// Fallback to dispatched stream (for tasks that were validated before execution completed)
		h.logger.Debug(ctx, "Task not found in executed stream, checking dispatched stream",
			observability.Int64("task_id", taskID))
		task, err = h.taskStreamManager.FindTaskInDispatched(taskID)
		if err != nil {
			h.logger.Error(ctx, "Failed to find task in executed or dispatched stream",
				observability.Int64("task_id", taskID),
				observability.Error(err))
			return err
		}
		messageID = "" // No messageID for dispatched stream lookup
	}

	// Mark task as validated
	now := time.Now()
	task.ValidatedAt = &now

	// Add to validated stream
	err = h.taskStreamManager.AddTaskToStream(ctx, types.StreamTaskValidated, task)
	if err != nil {
		h.logger.Error(ctx, "Failed to add task to validated stream", observability.Int64("task_id", taskID), observability.Error(err))
		return err
	}

	// Remove from executed stream if it was there (acknowledge)
	if messageID != "" {
		if err := h.taskStreamManager.AckTaskProcessed(ctx, types.StreamTaskExecuted, "task-processors", messageID); err != nil {
			h.logger.Warn(ctx, "Failed to acknowledge task from executed stream",
				observability.Int64("task_id", taskID),
				observability.String("message_id", messageID),
				observability.Error(err))
		}
		// Always remove from task index and timeout tracking, even if ack failed
		// The task is already validated, so it should not be in timeout tracking
		if err := h.taskStreamManager.RemoveTaskIndex(ctx, taskID); err != nil {
			h.logger.Warn(ctx, "Failed to remove task from index after validation",
				observability.Int64("task_id", taskID),
				observability.Error(err))
		}
		// CRITICAL: Always remove from timeout tracking - task is validated, don't rebroadcast
		if err := h.taskStreamManager.RemoveExecutedTaskTimeout(ctx, taskID); err != nil {
			h.logger.Warn(ctx, "Failed to remove task from timeout tracking after validation",
				observability.Int64("task_id", taskID),
				observability.Error(err))
		}
	} else {
		// Task was in dispatched stream (no timeout tracking), but still clean up index if present
		if err := h.taskStreamManager.RemoveTaskIndex(ctx, taskID); err != nil {
			h.logger.Debug(ctx, "Task index not found (expected for dispatched stream tasks)",
				observability.Int64("task_id", taskID))
		}
	}

	h.logger.Info(ctx, "Task moved to validated stream successfully", observability.Int64("task_id", taskID))

	return nil
}

// scheduleIPFSDeletion schedules IPFS file deletion after a 6-hour delay
// This is called after data is downloaded, DB is updated, and stream is handled
func (h *TaskEventHandler) scheduleIPFSDeletion(ctx context.Context, ipfsCID string, taskID int64) {
	const deletionDelay = 6 * time.Hour

	// Start a goroutine to handle delayed deletion
	go func() {
		// Create a new context for the deletion operation
		// We use a background context since the original context might be cancelled
		deleteCtx := context.Background()

		// Wait for the delay
		select {
		case <-time.After(deletionDelay):
			// Attempt to delete the IPFS file
			if err := h.ipfsClient.Delete(deleteCtx, ipfsCID); err != nil {
				h.logger.Error(deleteCtx, "Failed to delete IPFS file",
					observability.String("ipfs_cid", ipfsCID),
					observability.Int64("task_id", taskID),
					observability.Error(err))
			} else {
				h.logger.Info(deleteCtx, "Successfully deleted IPFS file",
					observability.String("ipfs_cid", ipfsCID),
					observability.Int64("task_id", taskID))
			}
		case <-ctx.Done():
			// Original context cancelled, abort deletion
			h.logger.Debug(deleteCtx, "IPFS deletion cancelled due to context cancellation",
				observability.String("ipfs_cid", ipfsCID),
				observability.Int64("task_id", taskID))
		}
	}()
}
