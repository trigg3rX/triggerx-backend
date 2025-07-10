package events

import (
	"context"

	"github.com/trigg3rX/triggerx-backend/internal/registrar/events/websocket"
)

// ProcessTaskEvent processes task-related events
func (h *TaskEventHandler) ProcessTaskEvent(ctx context.Context, event *websocket.ChainEvent) {
	switch event.EventName {
	case "TaskSubmitted":
		h.handleTaskSubmitted(ctx, event)
	case "TaskRejected":
		h.handleTaskRejected(ctx, event)
	default:
		h.logger.Warnf("Unknown task event: %s", event.EventName)
	}
}

// handleTaskSubmitted handles TaskSubmitted events
func (h *TaskEventHandler) handleTaskSubmitted(ctx context.Context, event *websocket.ChainEvent) {
	h.logger.Infof("Processing TaskSubmitted event on chain %s", event.ChainID)

	// Extract event data
	if eventData, ok := event.Data.(*websocket.ContractEventData); ok {
		h.logger.Infof("Task submitted: %+v", eventData.ParsedData)

		// TODO: Add your business logic here
		// - Store task in database
		// - Schedule task execution
		// - Send notifications to operators
		// - Update metrics

		h.logger.Infof("Successfully processed TaskSubmitted event at block %d", event.BlockNumber)
	} else {
		h.logger.Errorf("Failed to parse TaskSubmitted event data")
	}
}

// handleTaskRejected handles TaskRejected events
func (h *TaskEventHandler) handleTaskRejected(ctx context.Context, event *websocket.ChainEvent) {
	h.logger.Infof("Processing TaskRejected event on chain %s", event.ChainID)

	// Extract event data
	if eventData, ok := event.Data.(*websocket.ContractEventData); ok {
		h.logger.Infof("Task rejected: %+v", eventData.ParsedData)

		// TODO: Add your business logic here
		// - Update task status in database
		// - Send notifications
		// - Update metrics
		// - Handle refunds if applicable

		h.logger.Infof("Successfully processed TaskRejected event at block %d", event.BlockNumber)
	} else {
		h.logger.Errorf("Failed to parse TaskRejected event data")
	}
}
