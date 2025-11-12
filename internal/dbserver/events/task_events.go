package events

import (
	"time"
)

// TaskEventType represents the type of task event
type TaskEventType string

const (
	TaskEventTypeCreated       TaskEventType = "TASK_CREATED"
	TaskEventTypeUpdated       TaskEventType = "TASK_UPDATED"
	TaskEventTypeStatusChanged TaskEventType = "TASK_STATUS_CHANGED"
	TaskEventTypeFeeUpdated    TaskEventType = "TASK_FEE_UPDATED"
)

// TaskEvent represents a task-related event
type TaskEvent struct {
	Type      TaskEventType `json:"type"`
	TaskID    int64         `json:"task_id"`
	JobID     string        `json:"job_id"`
	UserID    string        `json:"user_id,omitempty"`
	Changes   interface{}   `json:"changes,omitempty"`
	Timestamp time.Time     `json:"timestamp"`
}

// TaskCreatedEvent represents a task created event
type TaskCreatedEvent struct {
	TaskID           int64     `json:"task_id"`
	JobID            string    `json:"job_id"`
	TaskDefinitionID int64     `json:"task_definition_id"`
	IsImua           bool      `json:"is_imua"`
	CreatedAt        time.Time `json:"created_at"`
	UserID           string    `json:"user_id,omitempty"`
}

// TaskUpdatedEvent represents a task updated event
type TaskUpdatedEvent struct {
	TaskID               int64      `json:"task_id"`
	JobID                string     `json:"job_id"`
	TaskPerformerID      *int64     `json:"task_performer_id,omitempty"`
	ExecutionTimestamp   *time.Time `json:"execution_timestamp,omitempty"`
	ExecutionTxHash      *string    `json:"execution_tx_hash,omitempty"`
	ProofOfTask          *string    `json:"proof_of_task,omitempty"`
	TaskOpXCost          *float64   `json:"task_opx_cost,omitempty"`
	TaskNumber           *int64     `json:"task_number,omitempty"`
	TaskAttesterIDs      *string    `json:"task_attester_ids,omitempty"`
	TpSignature          *string    `json:"tp_signature,omitempty"`
	TaSignature          *string    `json:"ta_signature,omitempty"`
	TaskSubmissionTxHash *string    `json:"task_submission_tx_hash,omitempty"`
	IsSuccessful         *bool      `json:"is_successful,omitempty"`
	TaskStatus           *string    `json:"task_status,omitempty"`
	TaskError            *string    `json:"task_error,omitempty"`
	UserID               string     `json:"user_id,omitempty"`
}

// TaskStatusChangedEvent represents a task status changed event
type TaskStatusChangedEvent struct {
	TaskID               int64   `json:"task_id"`
	JobID                string  `json:"job_id"`
	OldStatus            string  `json:"old_status"`
	NewStatus            string  `json:"new_status"`
	TaskNumber           *int64  `json:"task_number,omitempty"`
	TaskSubmissionTxHash *string `json:"task_submission_tx_hash,omitempty"`
	UserID               string  `json:"user_id,omitempty"`
}

// TaskFeeUpdatedEvent represents a task fee updated event
type TaskFeeUpdatedEvent struct {
	TaskID int64   `json:"task_id"`
	JobID  string  `json:"job_id"`
	OldFee float64 `json:"old_fee"`
	NewFee float64 `json:"new_fee"`
	UserID string  `json:"user_id,omitempty"`
}

// NewTaskEvent creates a new task event
func NewTaskEvent(eventType TaskEventType, taskID int64, jobID string, userID string, changes interface{}) *TaskEvent {
	return &TaskEvent{
		Type:      eventType,
		TaskID:    taskID,
		JobID:     jobID,
		UserID:    userID,
		Changes:   changes,
		Timestamp: time.Now(),
	}
}

// NewTaskCreatedEvent creates a new task created event
func NewTaskCreatedEvent(taskID int64, jobID string, taskDefinitionID int64, isImua bool, userID string) *TaskEvent {
	changes := &TaskCreatedEvent{
		TaskID:           taskID,
		JobID:            jobID,
		TaskDefinitionID: taskDefinitionID,
		IsImua:           isImua,
		CreatedAt:        time.Now(),
		UserID:           userID,
	}

	return NewTaskEvent(TaskEventTypeCreated, taskID, jobID, userID, changes)
}

// NewTaskUpdatedEvent creates a new task updated event
func NewTaskUpdatedEvent(taskID int64, jobID string, userID string, changes *TaskUpdatedEvent) *TaskEvent {
	changes.TaskID = taskID
	changes.JobID = jobID
	changes.UserID = userID

	return NewTaskEvent(TaskEventTypeUpdated, taskID, jobID, userID, changes)
}

// NewTaskStatusChangedEvent creates a new task status changed event
func NewTaskStatusChangedEvent(taskID int64, jobID string, oldStatus, newStatus string, userID string, taskNumber *int64, txHash *string) *TaskEvent {
	changes := &TaskStatusChangedEvent{
		TaskID:               taskID,
		JobID:                jobID,
		OldStatus:            oldStatus,
		NewStatus:            newStatus,
		TaskNumber:           taskNumber,
		TaskSubmissionTxHash: txHash,
		UserID:               userID,
	}

	return NewTaskEvent(TaskEventTypeStatusChanged, taskID, jobID, userID, changes)
}

// NewTaskFeeUpdatedEvent creates a new task fee updated event
func NewTaskFeeUpdatedEvent(taskID int64, jobID string, oldFee, newFee float64, userID string) *TaskEvent {
	changes := &TaskFeeUpdatedEvent{
		TaskID: taskID,
		JobID:  jobID,
		OldFee: oldFee,
		NewFee: newFee,
		UserID: userID,
	}

	return NewTaskEvent(TaskEventTypeFeeUpdated, taskID, jobID, userID, changes)
}
