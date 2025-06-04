package redis

import "time"

const (
	// Task Lifecycle Streams (Primary)
	TasksReadyStream     = "tasks:ready"      // Tasks ready for keeper execution
	TasksRetryStream     = "tasks:retry"      // Failed tasks needing retry
	TasksProcessingStream = "tasks:processing" // Tasks currently being processed
	TasksCompletedStream = "tasks:completed"  // Successfully completed tasks
	TasksFailedStream    = "tasks:failed"     // Permanently failed tasks

	// Job Lifecycle Streams (Secondary - for monitoring)
	JobsRunningStream   = "jobs:running"	 // Jobs that are running
	JobsCompletedStream = "jobs:completed"   // Jobs completed their cycle
	JobsFailedStream    = "jobs:failed"      // Jobs that failed

	// Retry Configuration
	MaxRetryAttempts = 3
	RetryBackoffBase = 5 * time.Second
)

// TaskStreamData represents task information for streams
type TaskStreamData struct {
	TaskID               int64                  `json:"task_id"`
	JobID                int64                  `json:"job_id"`
	RetryCount           int                    `json:"retry_count"`
	LastAttemptAt        *time.Time             `json:"last_attempt_at,omitempty"`
	ScheduledFor         *time.Time             `json:"scheduled_for,omitempty"`
	ManagerID            int64                  `json:"manager_id"`
	PerformerID          int64                  `json:"performer_id"`
}

// JobTriggeredData represents job trigger information
type JobStreamData struct {
	JobID          int64                  `json:"job_id"`
	TriggerCount   int                    `json:"trigger_count"`
	TaskIDs        []int64                `json:"task_ids"`
	TaskDefinitionID int                  `json:"task_definition_id"`
	ManagerID        int64                `json:"manager_id"`
}
