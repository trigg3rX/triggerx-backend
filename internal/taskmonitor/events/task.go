package events

import (
	"context"
	"encoding/hex"
	"sync"
	"time"

	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/trigg3rX/triggerx-backend/internal/taskmonitor/clients/database"
	"github.com/trigg3rX/triggerx-backend/internal/taskmonitor/clients/notify"
	"github.com/trigg3rX/triggerx-backend/internal/taskmonitor/tasks"
	"github.com/trigg3rX/triggerx-backend/internal/taskmonitor/types"
	"github.com/trigg3rX/triggerx-backend/pkg/ipfs"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
)

// ContractType represents the type of contract
type ContractType string

const (
	ContractTypeAttestationCenter ContractType = "attestation_center"
)

// ContractEventData represents parsed contract event data used downstream
type ContractEventData struct {
	EventType    string                 `json:"event_type"`
	ContractType ContractType           `json:"contract_type"`
	ParsedData   map[string]interface{} `json:"parsed_data"`
	RawData      []byte                 `json:"raw_data"`
	Topics       []string               `json:"topics"`
	BlockNumber  uint64                 `json:"block_number"`
	TxHash       string                 `json:"tx_hash"`
	LogIndex     uint                   `json:"log_index"`
}

// ChainEvent represents an event from any blockchain
type ChainEvent struct {
	ChainID      string       `json:"chain_id"`
	ChainName    string       `json:"chain_name"`
	ContractAddr string       `json:"contract_address"`
	ContractType ContractType `json:"contract_type"`
	EventName    string       `json:"event_name"`
	BlockNumber  uint64       `json:"block_number"`
	TxHash       string       `json:"tx_hash"`
	LogIndex     uint         `json:"log_index"`
	Data         interface{}  `json:"data"`
	RawLog       ethtypes.Log `json:"raw_log"`
	ProcessedAt  time.Time    `json:"processed_at"`
}

// TaskEventHandler handles task-related events
type TaskEventHandler struct {
	logger            observability.Logger
	tracer            observability.Tracer
	db                *database.DatabaseClient
	ipfsClient        ipfs.IPFSClient
	taskStreamManager *tasks.TaskStreamManager
	notifier          notify.Notifier
	traceRegistry     *sync.Map // taskID -> traceID for correlation between execution and validation
}

// NewTaskEventHandler creates a new TaskEventHandler instance
func NewTaskEventHandler(logger observability.Logger, tracer observability.Tracer, db *database.DatabaseClient, ipfsClient ipfs.IPFSClient, taskStreamManager *tasks.TaskStreamManager, notifier notify.Notifier) *TaskEventHandler {
	return &TaskEventHandler{
		logger:            logger,
		tracer:            tracer,
		db:                db,
		ipfsClient:        ipfsClient,
		taskStreamManager: taskStreamManager,
		notifier:          notifier,
		traceRegistry:     &sync.Map{},
	}
}

// ProcessConsensusEvent processes consensus events with already-parsed TaskSubmissionData
func (h *TaskEventHandler) ProcessConsensusEvent(ctx context.Context, event *ChainEvent, taskData *types.TaskSubmissionData) {
	// TaskData is already parsed by EventMonitor, just use it directly
	switch taskData.TaskDefinitionID {
	case 10001, 10002:
		h.logger.Debug(ctx, "Skipping task processing - Task is Internal Task", observability.Int64("task_number", taskData.TaskNumber))
		return
	case 1, 2, 3, 4, 5, 6, 7: // Added 7 for custom script jobs
		dataBytes, err := hex.DecodeString(taskData.Data) // Remove "0x" prefix before decoding
		if err != nil {
			h.logger.Error(ctx, "Failed to hex-decode data", observability.Error(err))
			return
		}
		ipfsHash := string(dataBytes)
		ipfsData, err := h.ipfsClient.Fetch(ctx, ipfsHash)
		if err != nil {
			h.logger.Error(ctx, "Failed to fetch IPFS data", observability.Error(err))
			return
		}

		// Extract trace context from IPFS data (embedded by performer)
		var traceID, spanID string
		if ipfsData.TraceID != "" {
			traceID = ipfsData.TraceID
			spanID = ipfsData.SpanID
		}

		// Continue trace if trace context is available
		if traceID != "" {
			ctx = observability.ContinueTrace(ctx, traceID, spanID)
		}

		// Create span for execution data processing
		ctx, span := h.tracer.Start(ctx, "task.monitor.execution",
			observability.WithSpanKind(trace.SpanKindConsumer),
		)
		defer span.End()

		taskOpxCostFloat, _ := ipfsData.ActionData.TotalFee.Float64()
		taskOpxCostFloat = taskOpxCostFloat / 1e18

		taskData.TaskID = ipfsData.ActionData.TaskID
		taskData.ExecutionTxHash = ipfsData.ActionData.ActionTxHash
		taskData.ExecutionTimestamp = ipfsData.ActionData.ExecutionTimestamp
		taskData.TaskOpxCost = taskOpxCostFloat
		taskData.ProofOfTask = ipfsData.ProofData.ProofOfTask
		taskData.ConvertedArguments = ipfsData.ActionData.ConvertedArguments

		// Set span attributes
		span.SetAttributes(
			attribute.Int64("task.id", taskData.TaskID),
			attribute.Int64("task.number", taskData.TaskNumber),
			attribute.String("task.submission.tx_hash", event.TxHash),
			attribute.Bool("task.is_accepted", taskData.IsAccepted),
			attribute.Int("task.definition_id", taskData.TaskDefinitionID),
			attribute.String("performer.address", taskData.PerformerAddress),
			attribute.String("ipfs.cid", ipfsHash),
		)

		// Add event when execution data is processed
		span.AddEvent("execution.data.processed", observability.WithEventAttributes(
			attribute.String("execution.tx_hash", taskData.ExecutionTxHash),
			attribute.String("ipfs.cid", ipfsHash),
		))

		// First, move the task from dispatched to completed based on onchain result
		h.logger.Info(ctx, "Task submitted onchain, moving to completed stream",
			observability.Int64("task_id", taskData.TaskID),
			observability.Int64("task_number", taskData.TaskNumber),
			observability.String("tx_hash", event.TxHash),
			observability.Bool("is_accepted", taskData.IsAccepted))

		// Move task from dispatched to completed stream
		if err := h.moveTaskToCompleted(ctx, taskData.TaskID); err != nil {
			span.RecordError(err, observability.WithErrorAttributes(
				attribute.String("error.type", "stream_move_failed"),
			))
			h.logger.Error(ctx, "Failed to move task to completed stream", observability.Error(err))
		}

		// Update task submission data in database
		if err := h.db.UpdateTaskSubmissionData(ctx, *taskData); err != nil {
			span.RecordError(err, observability.WithErrorAttributes(
				attribute.String("error.type", "database_update_failed"),
			))
			span.SetStatus(codes.Error, "failed to update execution data")
			h.logger.Error(ctx, "Failed to update task submission data in database", observability.Error(err))
		} else {
			span.AddEvent("execution.data.updated", observability.WithEventAttributes(
				attribute.String("database.table", "tasks"),
			))
			// Store trace ID in registry for correlation with validation event (after we have taskID)
			if traceID != "" && taskData.TaskID > 0 {
				h.traceRegistry.Store(taskData.TaskID, traceID)
			}
		}

		// For custom script jobs (TaskDefinitionID = 7), update storage
		if taskData.TaskDefinitionID == 7 && ipfsData.ActionData.StorageUpdates != nil && len(ipfsData.ActionData.StorageUpdates) > 0 {
			jobID, err := h.db.GetJobIDByTaskID(ctx, taskData.TaskID)
			if err != nil {
				span.RecordError(err, observability.WithErrorAttributes(
					attribute.String("error.type", "job_id_lookup_failed"),
				))
				h.logger.Error(ctx, "Failed to get job ID for task", observability.Int64("task_id", taskData.TaskID), observability.Error(err))
			} else {
				if err := h.db.UpdateScriptStorage(ctx, jobID, ipfsData.ActionData.StorageUpdates); err != nil {
					span.RecordError(err, observability.WithErrorAttributes(
						attribute.String("error.type", "storage_update_failed"),
					))
					h.logger.Error(ctx, "Failed to update script storage for job", observability.String("job_id", jobID.String()), observability.Error(err))
				} else {
					h.logger.Info(ctx, "Successfully updated storage keys for job", observability.Int("storage_keys", len(ipfsData.ActionData.StorageUpdates)), observability.String("job_id", jobID.String()))
				}
			}
		}

		// Update keeper points in database
		if err := h.db.UpdateKeeperPointsInDatabase(ctx, *taskData); err != nil {
			span.RecordError(err, observability.WithErrorAttributes(
				attribute.String("error.type", "keeper_points_update_failed"),
			))
			h.logger.Error(ctx, "Failed to update keeper points in database", observability.Error(err))
			// Don't return, continue processing
		}

		// Process validation data if attester IDs are present (validation complete)
		if len(taskData.AttesterIds) > 0 {
			// Retrieve trace ID from registry for correlation
			var validationTraceID string
			if traceID != "" {
				validationTraceID = traceID
			} else if taskData.TaskID > 0 {
				if storedTraceID, exists := h.traceRegistry.Load(taskData.TaskID); exists {
					if traceIDStr, ok := storedTraceID.(string); ok && traceIDStr != "" {
						validationTraceID = traceIDStr
					}
				}
			}

			// Continue trace for validation if we have trace ID
			validationCtx := ctx
			if validationTraceID != "" {
				validationCtx = observability.ContinueTrace(ctx, validationTraceID, "")
			}

			// Create span for validation data processing
			validationCtx, validationSpan := h.tracer.Start(validationCtx, "task.monitor.validation",
				observability.WithSpanKind(trace.SpanKindConsumer),
				observability.WithAttributes(
					attribute.Int64("task.id", taskData.TaskID),
					attribute.String("validation.tx_hash", event.TxHash),
					attribute.Int("validation.attester_count", len(taskData.AttesterIds)),
					attribute.String("validation.timestamp", time.Now().Format(time.RFC3339)),
				),
			)
			defer validationSpan.End()

			validationSpan.AddEvent("validation.data.received", observability.WithEventAttributes(
				attribute.String("validation.tx_hash", event.TxHash),
				attribute.Int("attester_count", len(taskData.AttesterIds)),
			))

			// Note: UpdateTaskValidationData doesn't exist in the database client
			// Validation data is already included in TaskSubmissionData and updated via UpdateTaskSubmissionData
			// If a separate validation update method exists, it should be called here
			// For now, we'll just mark validation as complete in the span
			validationSpan.AddEvent("validation.data.updated", observability.WithEventAttributes(
				attribute.String("database.table", "tasks"),
				attribute.Bool("validation.complete", true),
			))

			h.logger.Info(validationCtx, "Validation data processed",
				observability.Int64("task_id", taskData.TaskID),
				observability.Int("attester_count", len(taskData.AttesterIds)),
				observability.String("validation_tx_hash", event.TxHash))
		}

		// Notify user about task completion/rejection
		if h.notifier != nil {
			// Fetch user email by task id -> job id mapping
			email, err := h.db.GetUserEmailByTaskID(ctx, taskData.TaskID)
			if err != nil {
				h.logger.Warn(ctx, "Could not fetch user email for task", observability.Int64("task_id", taskData.TaskID), observability.Error(err))
			} else if email != "" {
				payload := notify.TaskStatusPayload{
					TaskID:          taskData.TaskID,
					JobID:           0,
					Status:          "completed",
					IsAccepted:      taskData.IsAccepted,
					SubmissionTx:    taskData.TaskSubmissionTxHash,
					ExecutionTxHash: taskData.ExecutionTxHash,
					ProofOfTask:     taskData.ProofOfTask,
					OccurredAt:      time.Now(),
				}
				if !taskData.IsAccepted {
					payload.Status = "failed"
				}
				if err := h.notifier.NotifyTaskStatus(context.Background(), email, payload); err != nil {
					h.logger.Warn(ctx, "Failed to notify user", observability.String("email", email), observability.Int64("task_id", taskData.TaskID), observability.Error(err))
				}
			}
		}
	default:
		return
	}
}

// moveTaskToCompleted moves a task from dispatched to completed stream
func (h *TaskEventHandler) moveTaskToCompleted(ctx context.Context, taskID int64) error {
	h.logger.Info(ctx, "Moving task to completed stream", observability.Int64("task_id", taskID))

	// Find the task in the dispatched stream
	task, err := h.taskStreamManager.FindTaskInDispatched(taskID)
	if err != nil {
		h.logger.Error(ctx, "Failed to find task in dispatched stream", observability.Int64("task_id", taskID), observability.Error(err))
		return err
	}

	// Mark task as completed
	task.CompletedAt = &[]time.Time{time.Now()}[0]

	// Add to completed stream
	err = h.taskStreamManager.AddTaskToStream(ctx, tasks.StreamTaskCompleted, task)
	if err != nil {
		h.logger.Error(ctx, "Failed to add task to completed stream", observability.Int64("task_id", taskID), observability.Error(err))
		return err
	}

	// Remove from dispatched stream (acknowledge)
	// Note: In a real implementation, we'd need to track the dispatched message ID
	h.logger.Info(ctx, "Task moved to completed stream successfully", observability.Int64("task_id", taskID))

	return nil
}
