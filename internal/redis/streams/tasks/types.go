package tasks

import (
	"time"

	"github.com/trigg3rX/triggerx-backend-imua/pkg/types"
)

const (
	// Task Lifecycle Streams (Redis Managed Internally)
	TasksReadyStream      = "tasks:ready"      // Ready tasks - NO EXPIRATION until moved
	TasksProcessingStream = "tasks:processing" // Processing tasks - 1 HOUR timeout → auto-move to failed
	TasksCompletedStream  = "tasks:completed"  // Completed tasks - EXPIRE IN 1 HOUR
	TasksFailedStream     = "tasks:failed"     // Failed tasks - managed by retry rules
	TasksRetryStream      = "tasks:retry"      // Retry tasks - managed by retry rules

	// Expiration Configuration
	TasksProcessingTTL = 1 * time.Hour      // 1 hour timeout for processing tasks
	TasksCompletedTTL  = 1 * time.Hour      // 1 hour for completed tasks
	TasksFailedTTL     = 7 * 24 * time.Hour // 7 days for failed tasks (debugging)
	TasksRetryTTL      = 24 * time.Hour     // 24 hours for retry tasks

	// Retry Configuration
	MaxRetryAttempts = 3
	RetryBackoffBase = 5 * time.Second
)

// TaskStreamData represents task information for Redis-managed task streams
type TaskStreamData struct {
	JobID            int64     `json:"job_id"`
	TaskDefinitionID int       `json:"task_definition_id"`
	CreatedAt        time.Time `json:"created_at"`

	// Execution tracking
	RetryCount    int        `json:"retry_count"`
	LastAttemptAt *time.Time `json:"last_attempt_at,omitempty"`
	ScheduledFor  *time.Time `json:"scheduled_for,omitempty"`

	// Core task data from schedulers
	SendTaskDataToKeeper types.SendTaskDataToKeeper `json:"send_task_data_to_keeper"`

	// Processing status (Redis internal use)
	ProcessingStartedAt *time.Time `json:"processing_started_at,omitempty"`
	CompletedAt         *time.Time `json:"completed_at,omitempty"`
	LastError           string     `json:"last_error,omitempty"`
}

// TaskProcessingTimeout represents a task that has timed out in processing
type TaskProcessingTimeout struct {
	TaskID              int64     `json:"task_id"`
	JobID               int64     `json:"job_id"`
	ProcessingStartedAt time.Time `json:"processing_started_at"`
	TimeoutAt           time.Time `json:"timeout_at"`
	PerformerID         int64     `json:"performer_id,omitempty"`
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
