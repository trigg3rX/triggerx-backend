package validation

import (
	"fmt"
	"time"

	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// ValidateTimeBasedJob checks if a time-based job (task definitions 1 and 2) should be executed
// based on its time interval, timeframe, and last execution time
func (v *TaskValidator) ValidateTimeBasedTask(job *types.HandleCreateJobData) (bool, error) {
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

	// For non-recurring jobs, check if job has already been executed and shouldn't run again
	if !job.Recurring && !job.LastExecutedAt.IsZero() {
		v.logger.Infof("Job %d is non-recurring and has already been executed on %s",
			job.JobID, job.LastExecutedAt.Format(time.RFC3339))
		return false, nil
	}

	now := time.Now().UTC()

	// Check if this is the job's first execution
	if job.LastExecutedAt.IsZero() {
		// For first execution, check if it's within the timeframe from creation
		if job.TimeFrame > 0 {
			v.logger.Infof("Job %d is within its timeframe (created: %s, timeframe: %d seconds)",
				job.JobID, job.CreatedAt.Format(time.RFC3339), job.TimeFrame)
			// Add tolerance to timeframe check
			// endTime := job.CreatedAt.Add(time.Duration(job.TimeFrame) * time.Second).Add(timeTolerance)
			// if now.After(endTime) {
			// 	v.logger.Infof("Job %d is outside its timeframe (created: %s, timeframe: %d seconds, with %v tolerance)",
			// 		job.JobID, job.CreatedAt.Format(time.RFC3339), job.TimeFrame, timeTolerance)
			// 	return false, nil
			// }
		}

		v.logger.Infof("Job %d is eligible for first execution", job.JobID)
		return true, nil
	}

	// Calculate the next scheduled execution time with tolerance
	nextExecution := job.LastExecutedAt.Add(time.Duration(job.TimeInterval) * time.Second).Add(-timeTolerance)

	// Store current time for logging but don't update job.LastExecutedAt yet
	// (this should be done by the caller after successful execution)
	//currentTime := now

	// Check if enough time has passed since the last execution (with tolerance)
	if now.Before(nextExecution) {
		v.logger.Infof("Not enough time has passed for job %d. Last executed: %s, next execution: %s (with %v tolerance)",
			job.JobID, job.LastExecutedAt.Format(time.RFC3339), nextExecution.Add(timeTolerance).Format(time.RFC3339), timeTolerance)
		return false, nil
	}

	// If timeframe is set, check if job is still within its timeframe (with tolerance)
	if job.TimeFrame > 0 {
		endTime := job.CreatedAt.Add(time.Duration(job.TimeFrame) * time.Second).Add(timeTolerance)
		if job.LastExecutedAt.After(endTime) {
			v.logger.Infof("Job %d is outside its timeframe (created: %s, timeframe: %d seconds, with %v tolerance)",
				job.JobID, job.CreatedAt.Format(time.RFC3339), job.TimeFrame, timeTolerance)
			return false, nil
		}
	}

	v.logger.Infof("Job %d is eligible for execution", job.JobID)
	return true, nil
}
