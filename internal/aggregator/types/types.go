package types

import (
	"time"

	"github.com/ethereum/go-ethereum/common"
)

// Task management constants
const (
	// Quorum threshold configuration
	QUORUM_THRESHOLD_NUMERATOR   = uint8(100)
	QUORUM_THRESHOLD_DENOMINATOR = uint8(100)

	// Task settings
	TASK_CHALLENGE_WINDOW_BLOCKS = 100
	BLOCK_TIME_SECONDS           = 12 * time.Second
	TASK_EXPIRY_TIMEOUT          = TASK_CHALLENGE_WINDOW_BLOCKS * BLOCK_TIME_SECONDS

	// Retry and timeout settings
	MAX_RETRY_ATTEMPTS = 3
	RPC_TIMEOUT        = 30 * time.Second
)

// Task management types
type TaskIndex = uint32
type BlockNumber = uint32

// TaskStatus represents the current state of a task
type TaskStatus string

const (
	TaskStatusPending    TaskStatus = "pending"
	TaskStatusProcessing TaskStatus = "processing"
	TaskStatusCompleted  TaskStatus = "completed"
	TaskStatusFailed     TaskStatus = "failed"
	TaskStatusExpired    TaskStatus = "expired"
)

// Task represents a task submitted to the aggregator
type Task struct {
	Index          TaskIndex      `json:"task_index"`
	Data           string         `json:"data"`
	CreatedAt      time.Time      `json:"created_at"`
	ExpiresAt      time.Time      `json:"expires_at"`
	Status         TaskStatus     `json:"status"`
	RequiredQuorum uint8          `json:"required_quorum"`
	SubmitterAddr  common.Address `json:"submitter_address"`
	ResponseCount  int            `json:"response_count"`
	BlockNumber    BlockNumber    `json:"block_number"`
}

// TaskResponse represents a response from an operator
type TaskResponse struct {
	TaskIndex    TaskIndex      `json:"task_index"`
	OperatorID   string         `json:"operator_id"`
	OperatorAddr common.Address `json:"operator_address"`
	Response     string         `json:"response"`
	Signature    string         `json:"signature"`
	SubmittedAt  time.Time      `json:"submitted_at"`
	IsValid      bool           `json:"is_valid"`
}

// OperatorInfo holds information about registered operators
type OperatorInfo struct {
	ID           string         `json:"id"`
	Address      common.Address `json:"address"`
	PublicKey    string         `json:"public_key"`
	Stake        string         `json:"stake"`
	RegisteredAt time.Time      `json:"registered_at"`
	LastActivity time.Time      `json:"last_activity"`
	IsActive     bool           `json:"is_active"`
}

// AggregatorStats holds aggregator performance metrics
type AggregatorStats struct {
	TotalTasks          uint64        `json:"total_tasks"`
	CompletedTasks      uint64        `json:"completed_tasks"`
	FailedTasks         uint64        `json:"failed_tasks"`
	ActiveOperators     int           `json:"active_operators"`
	LastTaskCreated     time.Time     `json:"last_task_created"`
	AverageResponseTime time.Duration `json:"average_response_time"`
}

// NewTaskRequest represents a request to create a new task
type NewTaskRequest struct {
	Data           string         `json:"data" binding:"required"`
	RequiredQuorum uint8          `json:"required_quorum"`
	Timeout        time.Duration  `json:"timeout,omitempty"`
	SubmitterAddr  common.Address `json:"submitter_address" binding:"required"`
}

// TaskListResponse represents the response for task listing
type TaskListResponse struct {
	Tasks      []Task `json:"tasks"`
	TotalCount int    `json:"total_count"`
	Page       int    `json:"page"`
	PageSize   int    `json:"page_size"`
}

// AggregatorConfig holds configuration specific to the aggregator
type AggregatorConfig struct {
	MaxConcurrentTasks int           `json:"max_concurrent_tasks"`
	DefaultTimeout     time.Duration `json:"default_timeout"`
	MinOperators       int           `json:"min_operators"`
	MaxOperators       int           `json:"max_operators"`
}
