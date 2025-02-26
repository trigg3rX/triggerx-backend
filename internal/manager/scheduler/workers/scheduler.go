package workers

import (
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// JobScheduler interface defines the methods that workers need from the scheduler
type JobScheduler interface {
	// RemoveJob removes a job from the scheduler
	RemoveJob(jobID int64)

	// GetJobDetails retrieves job details by ID
	GetJobDetails(jobID int64) (*types.HandleCreateJobData, error)

	// StartTimeBasedJob starts a time-based job
	StartTimeBasedJob(jobData types.HandleCreateJobData) error

	// StartEventBasedJob starts an event-based job
	StartEventBasedJob(jobData types.HandleCreateJobData) error

	// StartConditionBasedJob starts a condition-based job
	StartConditionBasedJob(jobData types.HandleCreateJobData) error

	// UpdateJobChainStatus updates the status of a job in a chain
	UpdateJobChainStatus(jobID int64, status string)

	// Logger returns the logger instance
	Logger() logging.Logger
}