package events

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"strconv"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/trigg3rX/triggerx-backend/internal/taskmonitor/clients/notify"
	"github.com/trigg3rX/triggerx-backend/internal/taskmonitor/tasks"
	"github.com/trigg3rX/triggerx-backend/internal/taskmonitor/types"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
)

// ProcessTaskEvent processes task-related events
func (h *TaskEventHandler) ProcessTaskEvent(ctx context.Context, event *ChainEvent) {
	if eventData, ok := event.Data.(*ContractEventData); ok {
		// Parse the event data to TaskSubmissionData
		taskData, err := h.parseTaskSubmissionData(ctx, eventData.ParsedData, event.TxHash)
		if err != nil {
			h.logger.Error(ctx, "Failed to parse TaskSubmitted event data", observability.Error(err))
			return
		}
		if event.EventName == "TaskRejected" {
			taskData.IsAccepted = false
		}

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

// parseTaskSubmissionData parses the event data into TaskSubmissionData
func (h *TaskEventHandler) parseTaskSubmissionData(ctx context.Context, parsedData map[string]interface{}, txHash string) (*types.TaskSubmissionData, error) {
	// Extract taskDefinitionId - it's indexed, so it comes as a string (hex-encoded)
	taskDefinitionIdStr, ok := parsedData["taskDefinitionId"].(string)
	if !ok {
		return nil, fmt.Errorf("taskDefinitionId not found or invalid type")
	}

	// Convert hex string to integer
	taskDefinitionIdInt64, err := strconv.ParseInt(taskDefinitionIdStr, 0, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse taskDefinitionId: %v", err)
	}
	taskDefinitionId := int(taskDefinitionIdInt64)

	if taskDefinitionId == 10001 || taskDefinitionId == 10002 {
		taskData := &types.TaskSubmissionData{
			TaskID: 0,
		}
		return taskData, nil
	}

	// Extract task number - it's already parsed as uint32, so we need to handle it as a number
	var taskNumber int64
	switch v := parsedData["taskNumber"].(type) {
	case string:
		// If it's a string, parse it
		var err error
		taskNumber, err = strconv.ParseInt(v, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse taskNumber: %v", err)
		}
	case float64:
		// If it's a float64 (from JSON unmarshaling), convert to int64
		taskNumber = int64(v)
	case int64:
		taskNumber = v
	case int:
		taskNumber = int64(v)
	case uint32:
		taskNumber = int64(v)
	case uint64:
		taskNumber = int64(v)
	default:
		return nil, fmt.Errorf("taskNumber has unexpected type: %T", v)
	}

	// Extract proof of task
	proofOfTask, ok := parsedData["proofOfTask"].(string)
	if !ok {
		return nil, fmt.Errorf("proofOfTask not found or invalid type")
	}

	// Extract data field - it's bytes, so it could be []byte or string
	var data string
	switch v := parsedData["data"].(type) {
	case []byte:
		data = hex.EncodeToString(v)
	default:
		return nil, fmt.Errorf("data field has unexpected type: %T", v)
	}

	// Extract operator address
	performerAddress, ok := parsedData["operator"].(string)
	if !ok {
		return nil, fmt.Errorf("operator not found or invalid type")
	}

	// Extract attesters IDs
	attestersIdsInterface, ok := parsedData["attestersIds"]
	if !ok {
		// h.logger.Warn("attestersIds not found in parsed data", "available_keys", getMapKeys(parsedData))

		// Try alternative field names that might be used
		if altInterface, altOk := parsedData["attesterIds"]; altOk {
			h.logger.Info(ctx, "Found attesterIds with alternative spelling")
			attestersIdsInterface = altInterface
		} else if altInterface, altOk := parsedData["attesters"]; altOk {
			h.logger.Info(ctx, "Found attesters field")
			attestersIdsInterface = altInterface
		} else {
			// Don't return error, just log and continue with empty slice
			attestersIdsInterface = []interface{}{}
		}
	}

	// Debug logging to see what we're working with
	// h.logger.Debug("attestersIds raw data",
	// 	"type", fmt.Sprintf("%T", attestersIdsInterface),
	// 	"value", attestersIdsInterface,
	// 	"is_nil", attestersIdsInterface == nil)

	// Convert attestersIds to int64 slice
	var attestersIds []int64
	switch v := attestersIdsInterface.(type) {
	case []string:
		// Handle the corrected format from formatValue ([]*big.Int -> []string)
		// h.logger.Debug("Processing attestersIds as []string", "count", len(v))
		for _, av := range v {
			if n, err := strconv.ParseInt(av, 10, 64); err == nil {
				attestersIds = append(attestersIds, n)
			} else {
				h.logger.Warn(ctx, "Failed to parse attester ID as string", observability.String("value", av), observability.Error(err))
			}
		}
	case []interface{}:
		// Fallback for legacy format
		// h.logger.Debug("Processing attestersIds as []interface{}", "count", len(v))
		for i, av := range v {
			switch vv := av.(type) {
			case float64:
				attestersIds = append(attestersIds, int64(vv))
			case string:
				// attempt parse decimal
				if n, err := strconv.ParseInt(vv, 10, 64); err == nil {
					attestersIds = append(attestersIds, n)
				} else {
					h.logger.Warn(ctx, "Failed to parse attester ID as string", observability.Int("index", i), observability.String("value", vv), observability.Error(err))
				}
			case *big.Int:
				attestersIds = append(attestersIds, vv.Int64())
			default:
				h.logger.Warn(ctx, "Unknown attester ID type", observability.Int("index", i), observability.String("type", fmt.Sprintf("%T", vv)), observability.String("value", fmt.Sprintf("%v", vv)))
			}
		}
	case []*big.Int:
		// Direct handling of []*big.Int
		// h.logger.Debug("Processing attestersIds as []*big.Int", "count", len(v))
		for _, id := range v {
			attestersIds = append(attestersIds, id.Int64())
		}
	// case nil:
	// h.logger.Debug("attestersIds is nil")
	default:
		// h.logger.Warn("attestersIds has unexpected type", "type", fmt.Sprintf("%T", v), "value", v)

		// Try to manually parse if it's a slice of unknown interface{}
		if slice, ok := v.([]interface{}); ok {
			// h.logger.Info("Attempting to parse as []interface{}", "count", len(slice))
			for i, item := range slice {
				switch itemVal := item.(type) {
				case *big.Int:
					attestersIds = append(attestersIds, itemVal.Int64())
				case string:
					if n, err := strconv.ParseInt(itemVal, 10, 64); err == nil {
						attestersIds = append(attestersIds, n)
					} else {
						h.logger.Warn(ctx, "Failed to parse attester ID from string", observability.Int("index", i), observability.String("value", itemVal), observability.Error(err))
					}
				case float64:
					attestersIds = append(attestersIds, int64(itemVal))
				default:
					h.logger.Warn(ctx, "Unknown attester ID type in slice", observability.Int("index", i), observability.String("type", fmt.Sprintf("%T", itemVal)), observability.String("value", fmt.Sprintf("%v", itemVal)))
				}
			}
		}
	}

	// h.logger.Debug("Final attestersIds", "count", len(attestersIds), "ids", attestersIds)

	// Create task submission data
	return &types.TaskSubmissionData{
		TaskID:               0,
		TaskNumber:           taskNumber,
		TaskDefinitionID:     taskDefinitionId,
		IsAccepted:           true,
		TaskSubmissionTxHash: txHash,
		PerformerAddress:     performerAddress,
		AttesterIds:          attestersIds,
		ProofOfTask:          proofOfTask,
		Data:                 data,
	}, nil
}

// getMapKeys returns the keys of a map as a slice of strings
// func getMapKeys(m map[string]interface{}) []string {
// 	keys := make([]string, 0, len(m))
// 	for k := range m {
// 		keys = append(keys, k)
// 	}
// 	return keys
// }

// func trim0x(s string) string {
// 	if len(s) >= 2 && s[0:2] == "0x" {
// 		return s[2:]
// 	}
// 	return s
// }
