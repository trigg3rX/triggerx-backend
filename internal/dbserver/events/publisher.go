package events

import (
	"context"

	"github.com/trigg3rX/triggerx-backend/internal/dbserver/websocket"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
)

// Publisher handles publishing task events to WebSocket clients
type Publisher struct {
	hub    *websocket.Hub
	logger observability.Logger
	ctx    context.Context
	cancel context.CancelFunc
}

// NewPublisher creates a new task event publisher
func NewPublisher(hub *websocket.Hub, logger observability.Logger) *Publisher {
	ctx, cancel := context.WithCancel(context.Background())

	return &Publisher{
		hub:    hub,
		logger: logger,
		ctx:    ctx,
		cancel: cancel,
	}
}

// PublishTaskCreated publishes a task created event
func (p *Publisher) PublishTaskCreated(ctx context.Context, taskID int64, jobID string, taskDefinitionID int64, isImua bool, userID string) {
	event := NewTaskCreatedEvent(taskID, jobID, taskDefinitionID, isImua, userID)

	taskEventData := &websocket.TaskEventData{
		TaskID:    taskID,
		JobID:     jobID,
		UserID:    userID,
		Changes:   event.Changes,
		Timestamp: event.Timestamp,
	}

	p.hub.BroadcastTaskCreated(ctx, taskEventData)
	p.logger.Info(ctx, "Published task created event for task %d", observability.Int64("task_id", taskID))
}

// PublishTaskUpdated publishes a task updated event
func (p *Publisher) PublishTaskUpdated(ctx context.Context, taskID int64, jobID string, userID string, changes *TaskUpdatedEvent) {
	event := NewTaskUpdatedEvent(taskID, jobID, userID, changes)

	taskEventData := &websocket.TaskEventData{
		TaskID:    taskID,
		JobID:     jobID,
		UserID:    userID,
		Changes:   event.Changes,
		Timestamp: event.Timestamp,
	}

	p.hub.BroadcastTaskUpdated(ctx, taskEventData)
	p.logger.Info(ctx, "Published task updated event for task %d", observability.Int64("task_id", taskID))
}

// PublishTaskStatusChanged publishes a task status changed event
func (p *Publisher) PublishTaskStatusChanged(ctx context.Context, taskID int64, jobID string, oldStatus, newStatus string, userID string, taskNumber *int64, txHash *string) {
	event := NewTaskStatusChangedEvent(taskID, jobID, oldStatus, newStatus, userID, taskNumber, txHash)

	taskEventData := &websocket.TaskEventData{
		TaskID:    taskID,
		JobID:     jobID,
		UserID:    userID,
		Changes:   event.Changes,
		Timestamp: event.Timestamp,
	}

	p.hub.BroadcastTaskStatusChanged(ctx, taskEventData)
	p.logger.Info(ctx, "Published task status changed event for task %d: %s -> %s", observability.Int64("task_id", taskID), observability.String("old_status", oldStatus), observability.String("new_status", newStatus))
}

// PublishTaskFeeUpdated publishes a task fee updated event
func (p *Publisher) PublishTaskFeeUpdated(ctx context.Context, taskID int64, jobID string, oldFee, newFee float64, userID string) {
	event := NewTaskFeeUpdatedEvent(taskID, jobID, oldFee, newFee, userID)

	taskEventData := &websocket.TaskEventData{
		TaskID:    taskID,
		JobID:     jobID,
		UserID:    userID,
		Changes:   event.Changes,
		Timestamp: event.Timestamp,
	}

	p.hub.BroadcastTaskFeeUpdated(ctx, taskEventData)
	p.logger.Info(ctx, "Published task fee updated event for task %d: %.2f -> %.2f", observability.Int64("task_id", taskID), observability.Float64("old_fee", oldFee), observability.Float64("new_fee", newFee))
}

// Shutdown gracefully shuts down the publisher
func (p *Publisher) Shutdown() {
	p.logger.Info(p.ctx, "Shutting down task event publisher")
	p.cancel()
}
