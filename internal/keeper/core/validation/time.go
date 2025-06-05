package validation

import (
	"fmt"
	"time"

	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// ValidateTimeBasedJob checks if a time-based job (task definitions 1 and 2) should be executed
// based on its time interval, timeframe, and last execution time
func (v *TaskValidator) ValidateTimeBasedTask(job *types.ScheduleTimeJobData) (bool, error) {
	// Define tolerance constant (3 seconds)
	const timeTolerance = 1500 * time.Millisecond

	// Ensure this is a time-based job
	if job.TaskDefinitionID != 1 && job.TaskDefinitionID != 2 {
		return false, fmt.Errorf("not a time-based job: task definition ID %d", job.TaskDefinitionID)
	}

	v.logger.Infof("Validating time-based job %d (taskDefID: %d)", job.JobID, job.TaskDefinitionID)

	// Check for zero or negative values
	if job.TimeInterval <= 0 {
		return false, fmt.Errorf("invalid time interval: %d (must be positive)", job.TimeInterval)
	}

	// Check if job last executed is before expiration time
	if job.LastExecutedAt.Before(job.ExpirationTime) {
		v.logger.Infof("Job %d was eligible for execution", job.JobID)
		return true, nil
	}

	// Check if next execution - last executed = time interval +- tolerance
	nextExecution := job.LastExecutedAt.Add(time.Duration(job.TimeInterval) * time.Second).Add(-timeTolerance)
	if nextExecution.After(job.ExpirationTime) {
		v.logger.Infof("Job %d is eligible for execution", job.JobID)
		return true, nil
	}

	return false, nil
}
