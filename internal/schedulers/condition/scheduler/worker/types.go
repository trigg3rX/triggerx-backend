// worker.go
package worker

import (
	"math/big"
	"time"
)

const (
	PerformerLockTTL         = 15 * time.Minute // Lock duration for condition monitoring
	DuplicateConditionWindow = 10 * time.Second // Window to prevent duplicate condition processing

	// Event-specific constants
	ConditionPollInterval = 1 * time.Second  // Poll every 1 second as requested
	EventPollInterval     = 2 * time.Second  // Poll every 2 seconds for new blocks
	DuplicateEventWindow  = 30 * time.Second // Window to prevent duplicate event processing
)

// Supported condition types
const (
	ConditionGreaterThan  = "greater_than"
	ConditionLessThan     = "less_than"
	ConditionBetween      = "between"
	ConditionEquals       = "equals"
	ConditionNotEquals    = "not_equals"
	ConditionGreaterEqual = "greater_equal"
	ConditionLessEqual    = "less_equal"
)

// Supported value source types
const (
	SourceTypeAPI    = "api"
	SourceTypeOracle = "oracle"
	SourceTypeStatic = "static"
)

// ValueResponse represents a generic response structure for fetching values
type ValueResponse struct {
	Value            float64 `json:"value"`
	Price            float64 `json:"price"`              // Common for price APIs
	USD              float64 `json:"usd"`                // Common for CoinGecko-style APIs
	Rate             float64 `json:"rate"`               // Common for exchange rate APIs
	Result           float64 `json:"result"`             // Generic result field
	Data             float64 `json:"data"`               // Generic data field
	Timestamp        int64   `json:"timestamp"`          // Optional timestamp
	SelectedKeyRoute string  `json:"selected_key_route"` // Dot-notation path to value in JSON
}

// ConditionTriggerNotification represents a notification from a worker when a condition is satisfied
type TriggerNotification struct {
	JobID         *big.Int  `json:"job_id"`
	TriggerTxHash string    `json:"trigger_tx_hash"`
	TriggerValue  float64   `json:"trigger_value"`
	TriggeredAt   time.Time `json:"triggered_at"`
}

// WorkerTriggerCallback is the interface that workers use to notify the scheduler
type WorkerTriggerCallback func(notification *TriggerNotification) error

// WorkerCleanupCallback is a callback function to clean up job data when worker stops
type WorkerCleanupCallback func(*big.Int) error
