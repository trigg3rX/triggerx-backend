package events

import (
	"context"
	"sync"

	"github.com/trigg3rX/triggerx-backend/internal/dbserver/websocket"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

// Publisher handles publishing task events to WebSocket clients
type Publisher struct {
	hub    *websocket.Hub
	logger logging.Logger
	ctx    context.Context
	cancel context.CancelFunc
	mu     sync.RWMutex
}

// NewPublisher creates a new task event publisher
func NewPublisher(hub *websocket.Hub, logger logging.Logger) *Publisher {
	ctx, cancel := context.WithCancel(context.Background())

	return &Publisher{
		hub:    hub,
		logger: logger,
		ctx:    ctx,
		cancel: cancel,
	}
}

// PublishTaskCreated publishes a task created event
func (p *Publisher) PublishTaskCreated(taskID int64, jobID string, taskDefinitionID int64, isImua bool, userID string) {
	event := NewTaskCreatedEvent(taskID, jobID, taskDefinitionID, isImua, userID)

	taskEventData := &websocket.TaskEventData{
		TaskID:    taskID,
		JobID:     jobID,
		UserID:    userID,
		Changes:   event.Changes,
		Timestamp: event.Timestamp,
	}

	p.hub.BroadcastTaskCreated(taskEventData)
	p.logger.Infof("Published task created event for task %d", taskID)
}

// PublishTaskUpdated publishes a task updated event
func (p *Publisher) PublishTaskUpdated(taskID int64, jobID string, userID string, changes *TaskUpdatedEvent) {
	event := NewTaskUpdatedEvent(taskID, jobID, userID, changes)

	taskEventData := &websocket.TaskEventData{
		TaskID:    taskID,
		JobID:     jobID,
		UserID:    userID,
		Changes:   event.Changes,
		Timestamp: event.Timestamp,
	}

	p.hub.BroadcastTaskUpdated(taskEventData)
	p.logger.Infof("Published task updated event for task %d", taskID)
}

// PublishTaskStatusChanged publishes a task status changed event
func (p *Publisher) PublishTaskStatusChanged(taskID int64, jobID string, oldStatus, newStatus string, userID string, taskNumber *int64, txHash *string) {
	event := NewTaskStatusChangedEvent(taskID, jobID, oldStatus, newStatus, userID, taskNumber, txHash)

	taskEventData := &websocket.TaskEventData{
		TaskID:    taskID,
		JobID:     jobID,
		UserID:    userID,
		Changes:   event.Changes,
		Timestamp: event.Timestamp,
	}

	p.hub.BroadcastTaskStatusChanged(taskEventData)
	p.logger.Infof("Published task status changed event for task %d: %s -> %s", taskID, oldStatus, newStatus)
}

// PublishTaskFeeUpdated publishes a task fee updated event
func (p *Publisher) PublishTaskFeeUpdated(taskID int64, jobID string, oldFee, newFee float64, userID string) {
	event := NewTaskFeeUpdatedEvent(taskID, jobID, oldFee, newFee, userID)

	taskEventData := &websocket.TaskEventData{
		TaskID:    taskID,
		JobID:     jobID,
		UserID:    userID,
		Changes:   event.Changes,
		Timestamp: event.Timestamp,
	}

	p.hub.BroadcastTaskFeeUpdated(taskEventData)
	p.logger.Infof("Published task fee updated event for task %d: %.2f -> %.2f", taskID, oldFee, newFee)
}

// Shutdown gracefully shuts down the publisher
func (p *Publisher) Shutdown() {
	p.logger.Info("Shutting down task event publisher")
	p.cancel()
}
