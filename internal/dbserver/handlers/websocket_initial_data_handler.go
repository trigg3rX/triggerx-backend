package handlers

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"github.com/trigg3rX/triggerx-backend/internal/dbserver/repository"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/websocket"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
)

// InitialDataHandler handles fetching initial data for WebSocket subscriptions
type InitialDataHandler struct {
	taskRepository repository.TaskRepository
	logger         observability.Logger
}

// NewInitialDataHandler creates a new initial data handler
func NewInitialDataHandler(taskRepository repository.TaskRepository, logger observability.Logger) *InitialDataHandler {
	return &InitialDataHandler{
		taskRepository: taskRepository,
		logger:         logger,
	}
}

// HandleInitialData fetches and sends initial data when a client subscribes to a room
func (h *InitialDataHandler) HandleInitialData(ctx context.Context, room string, client *websocket.Client) error {
	// Check if this is a job room subscription
	if strings.HasPrefix(room, "job:") {
		return h.handleJobRoomSubscription(ctx, room, client)
	}

	// For other room types, we don't need initial data
	return nil
}

// handleJobRoomSubscription handles initial data for job room subscriptions
func (h *InitialDataHandler) handleJobRoomSubscription(ctx context.Context, room string, client *websocket.Client) error {
	// Extract job ID from room name (e.g., "job:123" -> "123")
	jobIDStr := strings.TrimPrefix(room, "job:")
	if jobIDStr == "" {
		h.logger.Error(ctx, "Invalid job room format: %s", observability.String("room", room))
		return nil
	}

	// Convert job ID string to big.Int
	jobID, ok := new(big.Int).SetString(jobIDStr, 10)
	if !ok {
		h.logger.Error(ctx, "Invalid job ID format: %s", observability.String("job_id", jobIDStr))
		return nil
	}

	h.logger.Info(ctx, "Fetching initial tasks for job ID: %s", observability.String("job_id", jobIDStr))

	// Fetch all tasks for this job
	tasks, err := h.taskRepository.GetTasksByJobID(jobID)
	if err != nil {
		h.logger.Error(ctx, "Error fetching tasks for job %s: %v", observability.String("job_id", jobIDStr), observability.Error(err))
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
			TaskError:          task.TaskError,
			ConvertedArguments: task.ConvertedArguments,
		}
	}

	//find the created_chain id for the job using jobIDBig from database
	var createdChainID string
	createdChainID, err = h.taskRepository.GetCreatedChainIDByJobID(jobID)
	if err != nil {
		h.logger.Error(ctx, "Error retrieving created_chain_id for jobID %s: %v", observability.String("job_id", jobID.String()), observability.Error(err))
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

	// Send the message to the client safely (handle closed channel)
	func() {
		defer func() {
			if r := recover(); r != nil {
				// Channel is closed, client disconnected
				h.logger.Warn(ctx, "Client %s disconnected while sending initial snapshot for job %s: %v", observability.String("client_id", client.ID), observability.String("job_id", jobIDStr), observability.Error(r.(error)))
			}
		}()

		select {
		case client.Send <- snapshotMessage:
			h.logger.Info(ctx, "Sent initial snapshot with %d tasks for job %s to client %s", observability.Int("tasks_count", len(snapshotTasks)), observability.String("job_id", jobIDStr), observability.String("client_id", client.ID))
		default:
			h.logger.Error(ctx, "Failed to send initial snapshot to client %s - channel full", observability.String("client_id", client.ID))
		}
	}()

	return nil
}
