package events

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"strconv"
	"time"

	"github.com/trigg3rX/triggerx-backend/internal/taskmonitor/clients/websocket"
	"github.com/trigg3rX/triggerx-backend/internal/taskmonitor/tasks"
	"github.com/trigg3rX/triggerx-backend/internal/taskmonitor/types"
)

// ProcessTaskEvent processes task-related events
func (h *TaskEventHandler) ProcessTaskEvent(event *websocket.ChainEvent) {
	if eventData, ok := event.Data.(*websocket.ContractEventData); ok {
		// Parse the event data to TaskSubmissionData
		taskData, err := h.parseTaskSubmissionData(eventData.ParsedData, event.TxHash)
		if err != nil {
			h.logger.Errorf("Failed to parse TaskSubmitted event data: %v", err)
			return
		}
		if event.EventName == "TaskRejected" {
			taskData.IsAccepted = false
		}

		if taskData.TaskID != 0 {
			// First, move the task from dispatched to completed based on onchain result
			h.logger.Info("Task accepted onchain, moving to completed stream",
				"task_id", taskData.TaskID,
				"task_number", taskData.TaskNumber,
				"tx_hash", event.TxHash)

			// Move task from dispatched to completed stream
			if err := h.moveTaskToCompleted(taskData.TaskID); err != nil {
				h.logger.Errorf("Failed to move task to completed stream: %v", err)
				return
			}

			// Then update the database with parsed data (like registrar was doing)
			if err := h.db.UpdateTaskSubmissionData(*taskData); err != nil {
				h.logger.Errorf("Failed to update task submission data in database: %v", err)
				return
			}

			// Update keeper points in database
			if err := h.db.UpdateKeeperPointsInDatabase(*taskData); err != nil {
				h.logger.Errorf("Failed to update keeper points in database: %v", err)
				return
			}
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

	// Extract operator address
	performerAddress, ok := parsedData["operator"].(string)
	if !ok {
		return nil, fmt.Errorf("operator not found or invalid type")
	}

	// Extract attesters IDs
	attestersIdsInterface, ok := parsedData["attestersIds"]
	if !ok {
		return nil, fmt.Errorf("attestersIds not found")
	}

	// Convert attestersIds to string slice
	var attestersIds []int64
	switch v := attestersIdsInterface.(type) {
	case []interface{}:
		for _, id := range v {
			switch idVal := id.(type) {
			case *big.Int:
				attestersIds = append(attestersIds, idVal.Int64())
			default:
				return nil, fmt.Errorf("invalid attester ID type: %T", id)
			}
		}
	case []*big.Int:
		for _, id := range v {
			attestersIds = append(attestersIds, id.Int64())
		}
	default:
		return nil, fmt.Errorf("attestersIds is not a slice: %T", v)
	}

	data, ok := parsedData["data"].(string)
	if !ok {
		return nil, fmt.Errorf("data not found or invalid type")
	}

	var decodedData string
	dataBytes, err := hex.DecodeString(data)
	if err != nil {
		return nil, fmt.Errorf("failed to hex-decode data: %v", err)
	}
	decodedData = string(dataBytes)

	ipfsData, err := h.ipfsClient.Fetch(context.Background(), decodedData)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch IPFS content: %v", err)
	}

	taskData := &types.TaskSubmissionData{
		TaskID:               ipfsData.ActionData.TaskID,
		TaskNumber:           taskNumber,
		TaskDefinitionID:     taskDefinitionId,
		IsAccepted:           true,
		TaskSubmissionTxHash: txHash,
		PerformerAddress:     performerAddress,
		AttesterIds:          attestersIds,
		ExecutionTxHash:      ipfsData.ActionData.ActionTxHash,
		ExecutionTimestamp:   ipfsData.ActionData.ExecutionTimestamp,
		TaskOpxCost:          ipfsData.ActionData.TotalFee,
		ProofOfTask:          proofOfTask,
	}

	return taskData, nil
}
