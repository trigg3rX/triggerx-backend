package events

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"strconv"
	"time"

	"github.com/trigg3rX/triggerx-backend/internal/taskmonitor/tasks"
	"github.com/trigg3rX/triggerx-backend/internal/taskmonitor/types"
)

// ProcessTaskEvent processes task-related events
func (h *TaskEventHandler) ProcessTaskEvent(event *ChainEvent) {
	if eventData, ok := event.Data.(*ContractEventData); ok {
		// Debug: Log all parsed data keys to see what's available
		// h.logger.Debug("Parsed event data keys", "keys", getMapKeys(eventData.ParsedData))
		// for k, v := range eventData.ParsedData {
		// 	h.logger.Debug("Parsed data field", "key", k, "type", fmt.Sprintf("%T", v), "value", v)
		// }

		// Parse the event data to TaskSubmissionData
		taskData, err := h.parseTaskSubmissionData(eventData.ParsedData, event.TxHash)
		if err != nil {
			h.logger.Errorf("Failed to parse TaskSubmitted event data: %v", err)
			return
		}
		if event.EventName == "TaskRejected" {
			taskData.IsAccepted = false
		}

		// h.logger.Infof("Task data: %+v", taskData)

		switch taskData.TaskDefinitionID {
		case 10001, 10002:
			h.logger.Debugf("Skipping task processing - Task # %d is Internal Task", taskData.TaskNumber)
			return
		case 1, 2, 3, 4, 5, 6:
			dataBytes, err := hex.DecodeString(taskData.Data) // Remove "0x" prefix before decoding
			if err != nil {
				h.logger.Error("Failed to hex-decode data", "error", err)
				return
			}
			ipfsHash := string(dataBytes)
			ipfsData, err := h.ipfsClient.Fetch(context.Background(), ipfsHash)
			if err != nil {
				h.logger.Errorf("Failed to fetch IPFS data: %v", err)
				return
			}

			taskOpxCostFloat, _ := ipfsData.ActionData.TotalFee.Float64()
			taskOpxCostFloat = taskOpxCostFloat / 1e15

			taskData.TaskID = ipfsData.ActionData.TaskID
			taskData.ExecutionTxHash = ipfsData.ActionData.ActionTxHash
			taskData.ExecutionTimestamp = ipfsData.ActionData.ExecutionTimestamp
			taskData.TaskOpxCost = taskOpxCostFloat
			taskData.ProofOfTask = ipfsData.ProofData.ProofOfTask

			// h.logger.Infof("Task data: %+v", taskData)

			// First, move the task from dispatched to completed based on onchain result
			h.logger.Info("Task submitted onchain, moving to completed stream",
				"task_id", taskData.TaskID,
				"task_number", taskData.TaskNumber,
				"tx_hash", event.TxHash,
				"is_accepted", taskData.IsAccepted)

			// Move task from dispatched to completed stream
			if err := h.moveTaskToCompleted(taskData.TaskID); err != nil {
				h.logger.Errorf("Failed to move task to completed stream: %v", err)
			}

			// Update task submission data in database
			if err := h.db.UpdateTaskSubmissionData(*taskData); err != nil {
				h.logger.Errorf("Failed to update task submission data in database: %v", err)
			}

			// Update keeper points in database
			if err := h.db.UpdateKeeperPointsInDatabase(*taskData); err != nil {
				h.logger.Errorf("Failed to update keeper points in database: %v", err)
				return
			}
		default:
			return
		}
	}
}

// moveTaskToCompleted moves a task from dispatched to completed stream
func (h *TaskEventHandler) moveTaskToCompleted(taskID int64) error {
	h.logger.Info("Moving task to completed stream", "task_id", taskID)

	// Find the task in the dispatched stream
	task, err := h.taskStreamManager.FindTaskInDispatched(taskID)
	if err != nil {
		h.logger.Error("Failed to find task in dispatched stream", "task_id", taskID, "error", err)
		return err
	}

	// Mark task as completed
	task.CompletedAt = &[]time.Time{time.Now()}[0]

	// Add to completed stream
	err = h.taskStreamManager.AddTaskToStream(context.Background(), tasks.StreamTaskCompleted, task)
	if err != nil {
		h.logger.Error("Failed to add task to completed stream", "task_id", taskID, "error", err)
		return err
	}

	// Remove from dispatched stream (acknowledge)
	// Note: In a real implementation, we'd need to track the dispatched message ID
	h.logger.Info("Task moved to completed stream successfully", "task_id", taskID)

	return nil
}

// parseTaskSubmissionData parses the event data into TaskSubmissionData
func (h *TaskEventHandler) parseTaskSubmissionData(parsedData map[string]interface{}, txHash string) (*types.TaskSubmissionData, error) {
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
			h.logger.Info("Found attesterIds with alternative spelling")
			attestersIdsInterface = altInterface
		} else if altInterface, altOk := parsedData["attesters"]; altOk {
			h.logger.Info("Found attesters field")
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
				h.logger.Warn("Failed to parse attester ID as string", "value", av, "error", err)
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
					h.logger.Warn("Failed to parse attester ID as string", "index", i, "value", vv, "error", err)
				}
			case *big.Int:
				attestersIds = append(attestersIds, vv.Int64())
			default:
				h.logger.Warn("Unknown attester ID type", "index", i, "type", fmt.Sprintf("%T", vv), "value", vv)
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
						h.logger.Warn("Failed to parse attester ID from string", "index", i, "value", itemVal, "error", err)
					}
				case float64:
					attestersIds = append(attestersIds, int64(itemVal))
				default:
					h.logger.Warn("Unknown attester ID type in slice", "index", i, "type", fmt.Sprintf("%T", itemVal), "value", itemVal)
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
