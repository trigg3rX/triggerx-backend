package handlers

import (
	"math/big"
	"fmt"
	"strings"

	"github.com/trigg3rX/triggerx-backend/internal/dbserver/repository"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/websocket"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

// InitialDataHandler handles fetching initial data for WebSocket subscriptions
type InitialDataHandler struct {
	taskRepository repository.TaskRepository
	logger         logging.Logger
}

// NewInitialDataHandler creates a new initial data handler
func NewInitialDataHandler(taskRepository repository.TaskRepository, logger logging.Logger) *InitialDataHandler {
	return &InitialDataHandler{
		taskRepository: taskRepository,
		logger:         logger,
	}
}

// HandleInitialData fetches and sends initial data when a client subscribes to a room
func (h *InitialDataHandler) HandleInitialData(room string, client *websocket.Client) error {
	// Check if this is a job room subscription
	if strings.HasPrefix(room, "job:") {
		return h.handleJobRoomSubscription(room, client)
	}

	// For other room types, we don't need initial data
	return nil
}

// handleJobRoomSubscription handles initial data for job room subscriptions
func (h *InitialDataHandler) handleJobRoomSubscription(room string, client *websocket.Client) error {
	// Extract job ID from room name (e.g., "job:123" -> "123")
	jobIDStr := strings.TrimPrefix(room, "job:")
	if jobIDStr == "" {
		h.logger.Errorf("Invalid job room format: %s", room)
		return nil
	}

	// Convert job ID string to big.Int
	jobID, ok := new(big.Int).SetString(jobIDStr, 10)
	if !ok {
		h.logger.Errorf("Invalid job ID format: %s", jobIDStr)
		return nil
	}

	h.logger.Infof("Fetching initial tasks for job ID: %s", jobIDStr)

	// Fetch all tasks for this job
	tasks, err := h.taskRepository.GetTasksByJobID(jobID)
	if err != nil {
		h.logger.Errorf("Error fetching tasks for job %s: %v", jobIDStr, err)
		return err
	}

	// Convert repository tasks to snapshot format
	snapshotTasks := make([]websocket.JobTaskSnapshotData, len(tasks))
	for i, task := range tasks {
		snapshotTasks[i] = websocket.JobTaskSnapshotData{
			TaskID:             task.TaskID,
			TaskNumber:         task.TaskNumber,
			TaskOpXCost:        task.TaskOpXCost,
			ExecutionTimestamp: task.ExecutionTimestamp,
			ExecutionTxHash:    task.ExecutionTxHash,
			TaskPerformerID:    task.TaskPerformerID,
			TaskAttesterIDs:    task.TaskAttesterIDs,
			IsAccepted:         task.IsAccepted,
			TxURL:              task.TxURL,
			TaskStatus:         task.TaskStatus,
			ConvertedArguments:  task.ConvertedArguments ,
		}
	}

	//find the created_chain id for the job using jobIDBig from database
	var createdChainID string
	createdChainID, err = h.taskRepository.GetCreatedChainIDByJobID(jobID)
	if err != nil {
		h.logger.Errorf("Error retrieving created_chain_id for jobID %s: %v", jobID.String(), err)
		return err
	}

	// Set tx_url for each task
	explorerBaseURL := getExplorerBaseURL(createdChainID)
	for i := range snapshotTasks {
		if snapshotTasks[i].ExecutionTxHash != "" {
			snapshotTasks[i].TxURL = fmt.Sprintf("%s%s", explorerBaseURL, snapshotTasks[i].ExecutionTxHash)
		}
	}

	// Create and send snapshot message
	snapshotMessage := websocket.NewJobTasksSnapshotMessage(jobIDStr, snapshotTasks)

	// Send the message to the client
	select {
	case client.Send <- snapshotMessage:
		h.logger.Infof("Sent initial snapshot with %d tasks for job %s to client %s", len(snapshotTasks), jobIDStr, client.ID)
	default:
		h.logger.Errorf("Failed to send initial snapshot to client %s - channel full", client.ID)
	}

	return nil
}
