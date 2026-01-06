package repository

import (
	"errors"
	"math/big"
	"time"

	"github.com/trigg3rX/triggerx-backend/pkg/database"
	"github.com/trigg3rX/triggerx-backend/pkg/parser"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// TimeJobRepository defines the interface for time-based job operations.
type TimeJobRepository interface {
	// GetTimeJobsByNextExecutionTimestamp retrieves time-based jobs that are due for execution
	// within the specified look-ahead window.
	GetTimeJobsByNextExecutionTimestamp(lookAheadTime time.Time) ([]types.ScheduleTimeTaskData, error)
}

type timeJobRepository struct {
	db *database.Connection
}

// NewTimeJobRepository creates a new time job repository instance.
func NewTimeJobRepository(db *database.Connection) TimeJobRepository {
	return &timeJobRepository{
		db: db,
	}
}

// GetTimeJobsByNextExecutionTimestamp retrieves time-based jobs due for execution.
func (r *timeJobRepository) GetTimeJobsByNextExecutionTimestamp(lookAheadTime time.Time) ([]types.ScheduleTimeTaskData, error) {
	currentTime := time.Now()
	iter := r.db.Session().Query(getTimeJobsByNextExecutionTimestampQuery, currentTime, lookAheadTime).Iter()

	var timeJobs []types.ScheduleTimeTaskData
	var timeJob types.ScheduleTimeTaskData
	var jobIDBigInt *big.Int
	for iter.Scan(
		&jobIDBigInt, &timeJob.LastExecutedAt, &timeJob.ExpirationTime, &timeJob.TimeInterval,
		&timeJob.ScheduleType, &timeJob.CronExpression, &timeJob.SpecificSchedule, &timeJob.NextExecutionTimestamp,
		&timeJob.TaskTargetData.TargetChainID, &timeJob.TaskTargetData.TargetContractAddress, &timeJob.TaskTargetData.TargetFunction, &timeJob.TaskTargetData.ABI, &timeJob.TaskTargetData.ArgType,
		&timeJob.TaskTargetData.Arguments, &timeJob.TaskTargetData.DynamicArgumentsScriptUrl,
	) {
		timeJob.TaskTargetData.JobID = types.NewBigInt(jobIDBigInt)
		if timeJob.TaskTargetData.DynamicArgumentsScriptUrl != "" {
			timeJob.TaskDefinitionID = 2
			timeJob.TaskTargetData.TaskDefinitionID = 2
		} else {
			timeJob.TaskDefinitionID = 1
			timeJob.TaskTargetData.TaskDefinitionID = 1
		}

		var isImua bool
		err := r.db.Session().Query(isJobImuaQuery, jobIDBigInt).Scan(&isImua)
		if err != nil {
			return nil, err
		}
		timeJob.IsImua = isImua

		// Calculate next execution time after the current execution time
		nextExecutionTime, err := parser.CalculateNextExecutionTime(timeJob.NextExecutionTimestamp, timeJob.ScheduleType, timeJob.TimeInterval, timeJob.CronExpression, timeJob.SpecificSchedule)
		if err != nil {
			return nil, err
		}

		// If the next execution time is after the expiration time, complete the job
		if nextExecutionTime.After(timeJob.ExpirationTime) {
			err = r.completeTimeJob(timeJob.TaskTargetData.JobID.Int)
			if err != nil {
				return nil, err
			}
			err = r.updateTimeJobStatus(timeJob.TaskTargetData.JobID.Int, false)
			if err != nil {
				return nil, err
			}
		} else {
			err = r.updateTimeJobNextExecutionTimestamp(timeJob.TaskTargetData.JobID.Int, nextExecutionTime)
			if err != nil {
				return nil, err
			}
		}

		timeJobs = append(timeJobs, timeJob)
	}
	if err := iter.Close(); err != nil {
		return nil, err
	}

	return timeJobs, nil
}

// completeTimeJob marks a time job as completed.
func (r *timeJobRepository) completeTimeJob(jobID *big.Int) error {
	err := r.db.Session().Query(completeTimeJobStatusQuery, jobID).Exec()
	if err != nil {
		return errors.New("failed to complete time job")
	}

	err = r.db.Session().Query(updateJobDataToCompletedQuery, jobID).Exec()
	if err != nil {
		return errors.New("failed to update job_data status to completed")
	}

	return nil
}

// updateTimeJobStatus updates the active status of a time job.
func (r *timeJobRepository) updateTimeJobStatus(jobID *big.Int, isActive bool) error {
	err := r.db.Session().Query(updateTimeJobStatusQuery, isActive, jobID).Exec()
	if err != nil {
		return errors.New("failed to update time job status")
	}

	return nil
}

// updateTimeJobNextExecutionTimestamp updates the next execution timestamp for a time job.
func (r *timeJobRepository) updateTimeJobNextExecutionTimestamp(jobID *big.Int, nextExecutionTimestamp time.Time) error {
	err := r.db.Session().Query(updateTimeJobNextExecutionTimestampQuery, nextExecutionTimestamp, jobID).Exec()
	if err != nil {
		return errors.New("failed to update time job next execution timestamp")
	}

	return nil
}

// Query constants
const (
	getTimeJobsByNextExecutionTimestampQuery = `
		SELECT job_id, last_executed_at, expiration_time, time_interval,
			schedule_type, cron_expression, specific_schedule, next_execution_timestamp,
			target_chain_id, target_contract_address, target_function, 
			abi, arg_type, arguments, dynamic_arguments_script_url
		FROM triggerx.time_job_data
		WHERE next_execution_timestamp >= ? AND next_execution_timestamp <= ? AND is_active = true
		ALLOW FILTERING`

	isJobImuaQuery = `
		SELECT is_imua
		FROM triggerx.job_data
		WHERE job_id = ?`

	completeTimeJobStatusQuery = `
		UPDATE triggerx.time_job_data
		SET is_completed = true
		WHERE job_id = ?`

	updateJobDataToCompletedQuery = `
		UPDATE triggerx.job_data
		SET status = 'completed'
		WHERE job_id = ?`

	updateTimeJobStatusQuery = `
		UPDATE triggerx.time_job_data
		SET is_active = ?
		WHERE job_id = ?`

	updateTimeJobNextExecutionTimestampQuery = `
		UPDATE triggerx.time_job_data
		SET next_execution_timestamp = ?
		WHERE job_id = ?`
)
