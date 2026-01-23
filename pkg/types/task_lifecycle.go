package types

import (
	"encoding/json"
	"fmt"
	"time"
)

// Task Lifecycle Stream Names (Redis Managed Internally)
const (
	// Task Lifecycle Streams
	StreamTaskDispatched = "task:dispatched" // Dispatched to Aggregator (pending execution)
	StreamTaskExecuted   = "task:executed"   // Executed, pending validation (waiting for on-chain confirmation)
	StreamTaskValidated  = "task:validated"  // Validated on-chain (completed)
	StreamTaskCompleted  = "task:completed"  // Completed tasks
	StreamTaskFailed     = "task:failed"     // Failed tasks - managed by retry rules
	StreamTaskRetry      = "task:retry"      // Retry tasks - managed by retry rules

	// Expiration Configuration
	TasksProcessingTTL = 30 * time.Minute
	TasksExecutedTTL   = 30 * time.Minute
	TasksValidatedTTL  = 30 * time.Minute
	TasksCompletedTTL  = 30 * time.Minute
	TasksFailedTTL     = 30 * time.Minute
	TasksRetryTTL      = 30 * time.Minute

	// Retry Configuration
	MaxRetryAttempts = 3
)

// TaskStreamData represents task information for Redis-managed task streams
// Used by both taskmonitor and taskdispatcher services
type TaskStreamData struct {
	JobID            string        `json:"job_id"`             // Job identifier (string to match standard across codebase)
	TaskDefinitionID int           `json:"task_definition_id"` // Task definition ID
	CreatedAt        time.Time     `json:"created_at"`         // Task creation timestamp
	Network          KeeperNetwork `json:"network"`            // Network: mainnet, sepolia, or imua

	// Execution tracking
	RetryCount    int        `json:"retry_count"`               // Number of retry attempts
	LastAttemptAt *time.Time `json:"last_attempt_at,omitempty"` // Last retry attempt timestamp

	// Core task data from schedulers
	SendTaskDataToKeeper SendTaskDataToKeeper `json:"send_task_data_to_keeper"` // Task data to send to keeper

	// Processing status (Redis internal use)
	DispatchedAt      *time.Time `json:"dispatched_at,omitempty"`       // When task was dispatched
	ExecutedAt        *time.Time `json:"executed_at,omitempty"`         // When task was executed and sent to aggregator
	ValidatedAt       *time.Time `json:"validated_at,omitempty"`        // When task was validated on-chain
	CompletedAt       *time.Time `json:"completed_at,omitempty"`        // When task was completed
	RebroadcastCount  int        `json:"rebroadcast_count,omitempty"`   // Number of rebroadcast attempts
	LastRebroadcastAt *time.Time `json:"last_rebroadcast_at,omitempty"` // Last rebroadcast attempt time
	LastError         string     `json:"last_error,omitempty"`          // Last error message if any
}

// UnmarshalJSON implements custom JSON unmarshaling to handle backward compatibility
// with old messages where job_id was stored as a number instead of a string
func (t *TaskStreamData) UnmarshalJSON(data []byte) error {
	// Use an alias type to avoid infinite recursion
	type Alias TaskStreamData
	aux := &struct {
		JobID interface{} `json:"job_id"` // Accept both string and number
		*Alias
	}{
		Alias: (*Alias)(t),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Convert job_id to string regardless of input type
	switch v := aux.JobID.(type) {
	case string:
		t.JobID = v
	case float64:
		// JSON numbers are unmarshaled as float64
		t.JobID = fmt.Sprintf("%.0f", v)
	case int64:
		t.JobID = fmt.Sprintf("%d", v)
	case int:
		t.JobID = fmt.Sprintf("%d", v)
	default:
		return fmt.Errorf("job_id must be a string or number, got %T", v)
	}

	return nil
}

// TaskStatusUpdate represents status updates from performers
// Used for communication between keepers and taskmonitor
type TaskStatusUpdate struct {
	TaskID      int64     `json:"task_id"`         // Task identifier
	JobID       int64     `json:"job_id"`          // Job identifier
	Status      string    `json:"status"`          // Status: processing, completed, failed
	PerformerID int64     `json:"performer_id"`    // Performer identifier
	UpdatedAt   time.Time `json:"updated_at"`      // Update timestamp
	Error       string    `json:"error,omitempty"` // Error message if status is failed
	Data        []byte    `json:"data,omitempty"`  // Additional data
}

// TaskSubmissionData represents task submission data for consensus events
// Used by taskmonitor for tracking task submissions
type TaskSubmissionData struct {
	TaskID               int64         `json:"task_id"`                 // Task identifier
	TaskNumber           int64         `json:"task_number"`             // Task number
	TaskDefinitionID     int           `json:"task_definition_id"`      // Task definition ID
	IsAccepted           bool          `json:"is_accepted"`             // Whether task was accepted
	TaskSubmissionTxHash string        `json:"task_submission_tx_hash"` // Transaction hash of task submission
	PerformerAddress     string        `json:"performer_address"`       // Performer address
	AttesterIds          []int64       `json:"attester_ids"`            // Attester IDs
	ExecutionTxHash      string        `json:"execution_tx_hash"`       // Execution transaction hash
	ExecutionTimestamp   time.Time     `json:"execution_timestamp"`     // Execution timestamp
	TaskOpxCost          string        `json:"task_opx_cost"`           // Task OPX cost in Wei (as string)
	ProofOfTask          string        `json:"proof_of_task"`           // Proof of task
	Data                 string        `json:"data"`                    // Task data
	ConvertedArguments   []interface{} `json:"converted_arguments"`     // Converted arguments
}
