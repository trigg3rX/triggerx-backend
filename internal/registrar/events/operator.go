package events

import (
	"context"

	"github.com/trigg3rX/triggerx-backend/internal/registrar/events/websocket"
)

// ProcessOperatorEvent processes operator-related events
func (h *OperatorEventHandler) ProcessOperatorEvent(ctx context.Context, event *websocket.ChainEvent) {
	switch event.EventName {
	case "OperatorRegistered":
		h.handleOperatorRegistered(ctx, event)
	case "OperatorUnregistered":
		h.handleOperatorUnregistered(ctx, event)
	default:
		h.logger.Warnf("Unknown operator event: %s", event.EventName)
	}
}

// handleOperatorRegistered handles OperatorRegistered events
func (h *OperatorEventHandler) handleOperatorRegistered(ctx context.Context, event *websocket.ChainEvent) {
	h.logger.Infof("Processing OperatorRegistered event on chain %s", event.ChainID)

	// Extract event data
	if eventData, ok := event.Data.(*websocket.ContractEventData); ok {
		h.logger.Infof("Operator registered: %+v", eventData.ParsedData)

		// TODO: Add your business logic here
		// - Update operator registry in database
		// - Send notifications
		// - Update metrics

		h.logger.Infof("Successfully processed OperatorRegistered event at block %d", event.BlockNumber)
	} else {
		h.logger.Errorf("Failed to parse OperatorRegistered event data")
	}
}

// handleOperatorUnregistered handles OperatorUnregistered events
func (h *OperatorEventHandler) handleOperatorUnregistered(ctx context.Context, event *websocket.ChainEvent) {
	h.logger.Infof("Processing OperatorUnregistered event on chain %s", event.ChainID)

	// Extract event data
	if eventData, ok := event.Data.(*websocket.ContractEventData); ok {
		h.logger.Infof("Operator unregistered: %+v", eventData.ParsedData)

		// TODO: Add your business logic here
		// - Update operator registry in database
		// - Send notifications
		// - Update metrics

		h.logger.Infof("Successfully processed OperatorUnregistered event at block %d", event.BlockNumber)
	} else {
		h.logger.Errorf("Failed to parse OperatorUnregistered event data")
	}
}
