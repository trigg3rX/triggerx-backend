package tasks

import (
	"math/big"
	"time"

	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

const (
	// Task Lifecycle Streams (Redis Managed Internally)
	StreamTaskDispatched = "task:dispatched" // Dispatched to Aggregator (pending execution)
	StreamTaskExecuted   = "task:executed"   // Executed, pending validation (waiting for on-chain confirmation)
	StreamTaskValidated  = "task:validated"  // Validated on-chain (completed)
	StreamTaskFailed     = "task:failed"     // Failed tasks - managed by retry rules
	StreamTaskRetry      = "task:retry"      // Retry tasks - managed by retry rules

	// Expiration Configuration
	TasksProcessingTTL = 1 * time.Hour
	TasksExecutedTTL   = 15 * time.Minute // Timeout for executed tasks waiting validation
	TasksValidatedTTL  = 1 * time.Hour
	TasksCompletedTTL  = 1 * time.Hour
	TasksFailedTTL     = 1 * time.Hour
	TasksRetryTTL      = 1 * time.Hour

	// Retry Configuration
	MaxRetryAttempts = 3
)

// TaskStreamData represents task information for Redis-managed task streams
type TaskStreamData struct {
	JobID            *big.Int  `json:"job_id"`
	TaskDefinitionID int       `json:"task_definition_id"`
	CreatedAt        time.Time `json:"created_at"`

	// Execution tracking
	RetryCount    int        `json:"retry_count"`
	LastAttemptAt *time.Time `json:"last_attempt_at,omitempty"`

	// Core task data from schedulers
	SendTaskDataToKeeper types.SendTaskDataToKeeper `json:"send_task_data_to_keeper"`

	// Processing status (Redis internal use)
	DispatchedAt      *time.Time `json:"dispatched_at,omitempty"`
	ExecutedAt        *time.Time `json:"executed_at,omitempty"`         // When task was executed and sent to aggregator
	ValidatedAt       *time.Time `json:"validated_at,omitempty"`        // When task was validated on-chain
	RebroadcastCount  int        `json:"rebroadcast_count,omitempty"`   // Number of rebroadcast attempts
	LastRebroadcastAt *time.Time `json:"last_rebroadcast_at,omitempty"` // Last rebroadcast attempt time
	LastError         string     `json:"last_error,omitempty"`
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
