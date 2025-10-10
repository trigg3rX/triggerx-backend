package tasks

import (
	"time"

	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

const (
	// Task Lifecycle Streams (Redis Managed Internally)
	StreamTaskDispatched = "task:dispatched" // Dispatched to Aggregator
	StreamTaskCompleted  = "task:completed"  // Completed tasks
	StreamTaskFailed     = "task:failed"     // Failed tasks - managed by retry rules
	StreamTaskRetry      = "task:retry"      // Retry tasks - managed by retry rules

	// Expiration Configuration
	TasksProcessingTTL = 1 * time.Hour
	TasksCompletedTTL  = 1 * time.Hour
	TasksFailedTTL     = 1 * time.Hour
	TasksRetryTTL      = 1 * time.Hour

	// Retry Configuration
	MaxRetryAttempts = 3
)

// TaskStreamData represents task information for Redis-managed task streams
type TaskStreamData struct {
	JobID            string    `json:"job_id"`
	TaskDefinitionID int       `json:"task_definition_id"`
	CreatedAt        time.Time `json:"created_at"`
	IsMainnet        bool      `json:"is_mainnet"`

	// Execution tracking
	RetryCount    int        `json:"retry_count"`
	LastAttemptAt *time.Time `json:"last_attempt_at,omitempty"`

	// Core task data from schedulers
	SendTaskDataToKeeper types.SendTaskDataToKeeper `json:"send_task_data_to_keeper"`

	// Processing status (Redis internal use)
	DispatchedAt *time.Time `json:"dispatched_at,omitempty"`
	CompletedAt  *time.Time `json:"completed_at,omitempty"`
	LastError    string     `json:"last_error,omitempty"`
}

// TaskStatusUpdate represents status updates from performers
type TaskStatusUpdate struct {
	TaskID      int64     `json:"task_id"`
	JobID       int64     `json:"job_id"`
	Status      string    `json:"status"` // processing, completed, failed
	PerformerID int64     `json:"performer_id"`
	UpdatedAt   time.Time `json:"updated_at"`
	Error       string    `json:"error,omitempty"`
	Data        []byte    `json:"data,omitempty"`
}

// SchedulerTaskRequest represents the simplified interface for schedulers
type SchedulerTaskRequest struct {
	SendTaskDataToKeeper types.SendTaskDataToKeeper `json:"send_task_data_to_keeper"`
	SchedulerID          int                        `json:"scheduler_id"`
	Source               string                     `json:"source"` // "time_scheduler" or "condition_scheduler"
}
