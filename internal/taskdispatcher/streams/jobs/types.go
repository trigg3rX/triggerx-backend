package jobs

import (
	"time"

	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

const (
	// Job Lifecycle Streams (Condition Scheduler Only)
	JobsRunningStream   = "jobs:running"   // Active jobs - NO EXPIRATION
	JobsCompletedStream = "jobs:completed" // Completed jobs - EXPIRE IN 24 HOURS

	// Expiration Configuration
	JobsCompletedTTL   = 24 * time.Hour     // 24 hours for completed jobs
)

// JobStreamData represents job information for condition scheduler streams
type JobStreamData struct {
	JobID            int64      `json:"job_id"`
	TaskDefinitionID int        `json:"task_definition_id"`
	CreatedAt        time.Time  `json:"created_at"`
	ExpirationTime   time.Time  `json:"expiration_time"`
	LastExecutedAt   *time.Time `json:"last_executed_at,omitempty"`
	Recurring        bool       `json:"recurring"`
	IsActive         bool       `json:"is_active"`

	// Task generation data
	TaskTargetData types.TaskTargetData `json:"task_target_data"`

	// Type-specific trigger data (only one will be populated)
	EventData     *types.EventWorkerData     `json:"event_data,omitempty"`
	ConditionData *types.ConditionWorkerData `json:"condition_data,omitempty"`

	// Execution tracking
	TriggerCount int    `json:"trigger_count"`
	LastError    string `json:"last_error,omitempty"`
}