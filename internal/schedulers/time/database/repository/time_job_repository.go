package repository

import (
	"errors"
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
	// UpdateTimeJobStatus updates the active status of a time job.
	UpdateTimeJobStatus(jobID string, isActive bool) error
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
	// Exclude expired jobs by checking expiration_time >= currentTime
	iter := r.db.Session().Query(getTimeJobsByNextExecutionTimestampQuery, currentTime, lookAheadTime, currentTime).Iter()

	var timeJobs []types.ScheduleTimeTaskData
	var timeJob types.ScheduleTimeTaskData
	var taskDefinitionID int
	var agentScriptURL, agentScriptLanguage, agentScriptHash *string
	var agentTargetChainID, maxExecutionTime *int
	var challengePeriod *int64

	for iter.Scan(
		&timeJob.TaskTargetData.JobID, &timeJob.LastExecutedAt, &timeJob.ExpirationTime, &timeJob.TimeInterval,
		&timeJob.ScheduleType, &timeJob.CronExpression, &timeJob.SpecificSchedule, &timeJob.NextExecutionTimestamp,
		&timeJob.TaskTargetData.TargetChainID, &timeJob.TaskTargetData.TargetContractAddress, &timeJob.TaskTargetData.TargetFunction, &timeJob.TaskTargetData.ABI, &timeJob.TaskTargetData.ArgType,
		&timeJob.TaskTargetData.Arguments, &timeJob.TaskTargetData.DynamicArgumentsScriptUrl,
		&taskDefinitionID, &agentScriptURL, &agentScriptLanguage, &agentScriptHash, &agentTargetChainID, &maxExecutionTime, &challengePeriod,
	) {
		// Set TaskDefinitionID based on what's in the database
		timeJob.TaskDefinitionID = taskDefinitionID
		timeJob.TaskTargetData.TaskDefinitionID = taskDefinitionID

		// If this is an agent job (TDI 7), populate agent fields
		if taskDefinitionID == types.TaskDefTimeBasedAgent {
			if agentScriptURL != nil {
				timeJob.TaskTargetData.AgentScriptURL = *agentScriptURL
			}
			if agentScriptLanguage != nil {
				timeJob.TaskTargetData.AgentScriptLanguage = *agentScriptLanguage
			}
			if agentScriptHash != nil {
				timeJob.TaskTargetData.AgentScriptHash = *agentScriptHash
			}
			if agentTargetChainID != nil {
				timeJob.TaskTargetData.AgentTargetChainID = *agentTargetChainID
			}
			if maxExecutionTime != nil {
				timeJob.TaskTargetData.MaxExecutionTime = *maxExecutionTime
			} else {
				timeJob.TaskTargetData.MaxExecutionTime = types.DefaultMaxExecutionTime
			}
			if challengePeriod != nil {
				timeJob.TaskTargetData.ChallengePeriod = *challengePeriod
			} else {
				timeJob.TaskTargetData.ChallengePeriod = types.DefaultChallengePeriod
			}
		}

		var isImua bool
		err := r.db.Session().Query(isJobImuaQuery, timeJob.TaskTargetData.JobID).Scan(&isImua)
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
			err = r.completeTimeJob(timeJob.TaskTargetData.JobID)
			if err != nil {
				return nil, err
			}
			err = r.UpdateTimeJobStatus(timeJob.TaskTargetData.JobID, false)
			if err != nil {
				return nil, err
			}
		} else {
			err = r.updateTimeJobNextExecutionTimestamp(timeJob.TaskTargetData.JobID, nextExecutionTime)
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

// completeTimeJob marks a time job as completed by updating the job_data status.
// The is_active field is set to false separately via UpdateTimeJobStatus.
func (r *timeJobRepository) completeTimeJob(jobID string) error {
	err := r.db.Session().Query(updateJobDataToCompletedQuery, jobID).Exec()
	if err != nil {
		return errors.New("failed to update job_data status to completed")
	}

	return nil
}

// UpdateTimeJobStatus updates the active status of a time job.
func (r *timeJobRepository) UpdateTimeJobStatus(jobID string, isActive bool) error {
	err := r.db.Session().Query(updateTimeJobStatusQuery, isActive, jobID).Exec()
	if err != nil {
		return errors.New("failed to update time job status")
	}

	return nil
}

// updateTimeJobNextExecutionTimestamp updates the next execution timestamp for a time job.
func (r *timeJobRepository) updateTimeJobNextExecutionTimestamp(jobID string, nextExecutionTimestamp time.Time) error {
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
			abi, arg_type, arguments, dynamic_arguments_script_url,
			task_definition_id, agent_script_url, agent_script_language, agent_script_hash,
			agent_target_chain_id, max_execution_time, challenge_period
		FROM triggerx.time_job_data
		WHERE next_execution_timestamp >= ? AND next_execution_timestamp <= ? 
			AND expiration_time >= ? AND is_active = true
		ALLOW FILTERING`

	isJobImuaQuery = `
		SELECT is_imua
		FROM triggerx.job_data
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
