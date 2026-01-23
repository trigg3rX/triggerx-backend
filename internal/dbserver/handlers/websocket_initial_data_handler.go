package handlers

import (
	"context"
	"fmt"
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
	jobID := strings.TrimPrefix(room, "job:")
	if jobID == "" {
		h.logger.Error(ctx, "Invalid job room format", observability.String("room", room))
		return nil
	}

	// Fetch all tasks for this job
	tasks, err := h.taskRepository.GetTasksByJobID(jobID)
	if err != nil {
		h.logger.Error(ctx, "Error fetching tasks for job", observability.String("job_id", jobID), observability.Error(err))
		return err
	}

	// Convert repository tasks to snapshot format
	snapshotTasks := make([]websocket.JobTaskSnapshotData, len(tasks))
	for i, task := range tasks {
		snapshotTasks[i] = websocket.JobTaskSnapshotData{
			TaskID:               task.TaskID,
			TaskNumber:           task.TaskNumber,
			TaskOpXCost:          task.TaskOpXCost,
			ExecutionTimestamp:   task.ExecutionTimestamp,
			ExecutionTxHash:      task.ExecutionTxHash,
			TaskPerformerAddress: task.TaskPerformerAddress,
			TaskAttesterAddress:  task.TaskAttesterAddress,
			IsAccepted:           task.IsAccepted,
			TxURL:                task.TxURL,
			TaskStatus:           task.TaskStatus,
			TaskError:            task.TaskError,
			ConvertedArguments:   task.ConvertedArguments,
		}
	}

	// Get the created_chain_id for the job using jobID from database
	createdChainID, err := h.taskRepository.GetCreatedChainIDByJobID(jobID)
	if err != nil {
		h.logger.Error(ctx, "Error retrieving created_chain_id for jobID", observability.String("job_id", jobID), observability.Error(err))
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
	snapshotMessage := websocket.NewJobTasksSnapshotMessage(jobID, snapshotTasks)

	// Send the message to the client safely (handle closed channel)
	func() {
		defer func() {
			if r := recover(); r != nil {
				// Channel is closed, client disconnected
				h.logger.Warn(ctx, "Client disconnected while sending initial snapshot for job", observability.String("client_id", client.ID), observability.String("job_id", jobID), observability.Error(r.(error)))
			}
		}()

		select {
		case client.Send <- snapshotMessage:
			h.logger.Info(ctx, "Sent initial snapshot with tasks for job to client", observability.Int("tasks_count", len(snapshotTasks)), observability.String("job_id", jobID), observability.String("client_id", client.ID))
		default:
			h.logger.Error(ctx, "Failed to send initial snapshot to client - channel full", observability.String("client_id", client.ID))
		}
	}()

	return nil
}
