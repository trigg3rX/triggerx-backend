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
	UserAddress string      `json:"user_address,omitempty"`
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
	UserAddress      string    `json:"user_address,omitempty"`
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
	UserAddress               string     `json:"user_address,omitempty"`
}

// TaskStatusChangedEvent represents a task status changed event
type TaskStatusChangedEvent struct {
	TaskID               int64   `json:"task_id"`
	JobID                string  `json:"job_id"`
	OldStatus            string  `json:"old_status"`
	NewStatus            string  `json:"new_status"`
	TaskNumber           *int64  `json:"task_number,omitempty"`
	TaskSubmissionTxHash *string `json:"task_submission_tx_hash,omitempty"`
	UserAddress               string  `json:"user_address,omitempty"`
}

// TaskFeeUpdatedEvent represents a task fee updated event
type TaskFeeUpdatedEvent struct {
	TaskID int64   `json:"task_id"`
	JobID  string  `json:"job_id"`
	OldFee float64 `json:"old_fee"`
	NewFee float64 `json:"new_fee"`
	UserAddress string  `json:"user_address,omitempty"`
}

// NewTaskEvent creates a new task event
func NewTaskEvent(eventType TaskEventType, taskID int64, jobID string, userAddress string, changes interface{}) *TaskEvent {
	return &TaskEvent{
		Type:      eventType,
		TaskID:    taskID,
		JobID:     jobID,
		UserAddress: userAddress,
		Changes:   changes,
		Timestamp: time.Now(),
	}
}

// NewTaskCreatedEvent creates a new task created event
func NewTaskCreatedEvent(taskID int64, jobID string, taskDefinitionID int64, isImua bool, userAddress string) *TaskEvent {
	changes := &TaskCreatedEvent{
		TaskID:           taskID,
		JobID:            jobID,
		TaskDefinitionID: taskDefinitionID,
		IsImua:           isImua,
		CreatedAt:        time.Now(),
		UserAddress:      userAddress,
	}

	return NewTaskEvent(TaskEventTypeCreated, taskID, jobID, userAddress, changes)
}

// NewTaskUpdatedEvent creates a new task updated event
func NewTaskUpdatedEvent(taskID int64, jobID string, userAddress string, changes *TaskUpdatedEvent) *TaskEvent {
	changes.TaskID = taskID
	changes.JobID = jobID
	changes.UserAddress = userAddress

	return NewTaskEvent(TaskEventTypeUpdated, taskID, jobID, userAddress, changes)
}

// NewTaskStatusChangedEvent creates a new task status changed event
func NewTaskStatusChangedEvent(taskID int64, jobID string, oldStatus, newStatus string, userAddress string, taskNumber *int64, txHash *string) *TaskEvent {
	changes := &TaskStatusChangedEvent{
		TaskID:               taskID,
		JobID:                jobID,
		OldStatus:            oldStatus,
		NewStatus:            newStatus,
		TaskNumber:           taskNumber,
		TaskSubmissionTxHash: txHash,
		UserAddress:          userAddress,
	}

	return NewTaskEvent(TaskEventTypeStatusChanged, taskID, jobID, userAddress, changes)
}

// NewTaskFeeUpdatedEvent creates a new task fee updated event
func NewTaskFeeUpdatedEvent(taskID int64, jobID string, oldFee, newFee float64, userAddress string) *TaskEvent {
	changes := &TaskFeeUpdatedEvent{
		TaskID: taskID,
		JobID:  jobID,
		OldFee: oldFee,
		NewFee: newFee,
		UserAddress: userAddress,
	}

	return NewTaskEvent(TaskEventTypeFeeUpdated, taskID, jobID, userAddress, changes)
}
