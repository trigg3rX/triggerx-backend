package handlers

import (
	"strings"

	"github.com/trigg3rX/triggerx-backend/internal/dbserver/websocket"
	"github.com/trigg3rX/triggerx-backend/pkg/datastore/interfaces"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// InitialDataHandler handles fetching initial data for WebSocket subscriptions
type InitialDataHandler struct {
	taskRepo interfaces.GenericRepository[types.TaskDataEntity]
	logger   logging.Logger
}

// NewInitialDataHandler creates a new initial data handler
func NewInitialDataHandler(taskRepository interfaces.GenericRepository[types.TaskDataEntity], logger logging.Logger) *InitialDataHandler {
	return &InitialDataHandler{
		taskRepo: taskRepository,
		logger:   logger,
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

	h.logger.Infof("Fetching initial tasks for job ID: %s", jobIDStr)

	// Note: We need jobRepository to get job details. This handler should be updated to take it
	// For now, we'll return an empty snapshot since we can't get tasks without job repository
	// TODO: Update InitialDataHandler to take jobRepository as well

	h.logger.Warnf("WebSocket initial data handler needs job repository to fetch tasks for job %s", jobIDStr)

	// Return empty snapshot for now
	snapshotMessage := websocket.NewJobTasksSnapshotMessage(jobIDStr, []websocket.JobTaskSnapshotData{})

	select {
	case client.Send <- snapshotMessage:
		h.logger.Infof("Sent empty initial snapshot for job %s to client %s", jobIDStr, client.ID)
	default:
		h.logger.Errorf("Failed to send initial snapshot to client %s - channel full", client.ID)
	}

	return nil
}
