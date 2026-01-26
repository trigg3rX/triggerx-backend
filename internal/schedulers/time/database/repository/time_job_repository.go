package repository

import (
	"errors"
	"time"

	"github.com/gocql/gocql"
	"github.com/trigg3rX/triggerx-backend/pkg/database"
	"github.com/trigg3rX/triggerx-backend/pkg/parser"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// TimeJobRepository defines the interface for time-based job operations.
type TimeJobRepository interface {
	// GetTimeJobsByNextExecutionTimestamp retrieves time-based jobs that are due for execution
	// within the specified look-ahead window, calculates the next_execution_timestamp for eligible jobs
	// and creates the task data for the eligible jobs.
	GetTimeJobsByNextExecutionTimestamp(lookAheadTime time.Time) ([]types.SendTaskDataToKeeper, error)
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
func (r *timeJobRepository) GetTimeJobsByNextExecutionTimestamp(lookAheadTime time.Time) ([]types.SendTaskDataToKeeper, error) {
	currentTime := time.Now()
	// Fetch time jobs by checking next_execution_timestamp <= lookAheadTime
	// This includes both past-due jobs and jobs scheduled within the look-ahead window
	// Expired jobs (expiration_time <= currentTime) will be filtered out and completed in the loop
	iter := r.db.Session().Query(getTimeJobsByNextExecutionTimestampQuery, lookAheadTime).Iter()

	var eligibleTimeJobs [3]types.SendTaskDataToKeeper

	eligibleTimeJobs[0].Network = types.NetworkMainnet
	eligibleTimeJobs[1].Network = types.NetworkSepolia
	eligibleTimeJobs[2].Network = types.NetworkImua

	var timeJobData types.TimeJobDataEntity

	for iter.Scan(
		&timeJobData.JobID, &timeJobData.TaskDefinitionID, &timeJobData.Network, &timeJobData.ScheduleType, &timeJobData.TimeInterval,
		&timeJobData.CronExpression, &timeJobData.SpecificSchedule, &timeJobData.Timezone, &timeJobData.NextExecutionTimestamp,
		&timeJobData.TargetChainID, &timeJobData.TargetContractAddress, &timeJobData.TargetFunction, &timeJobData.ABI, 
		&timeJobData.ArgType, &timeJobData.Arguments, &timeJobData.ExecutionScriptURL, &timeJobData.ExecutionScriptLanguage, 
		&timeJobData.ExecutionScriptHash, &timeJobData.MaxExecutionTime, &timeJobData.ChallengePeriod,
		&timeJobData.IsActive, &timeJobData.LastExecutedAt, &timeJobData.ExpirationTime,
	) {
		// If the job has expiration_time <= currentTime, complete the job
		if timeJobData.ExpirationTime.Before(currentTime) {
			err := r.completeTimeJob(timeJobData.JobID)
			if err != nil {
				return nil, err
			}
			continue // Skip to the next job
		}

		// Calculate next execution time after the current execution time
		nextExecutionTime, err := parser.CalculateNextExecutionTime(timeJobData.NextExecutionTimestamp, timeJobData.ScheduleType, timeJobData.TimeInterval, timeJobData.CronExpression, timeJobData.SpecificSchedule)
		if err != nil {
			return nil, err
		}

		// Update the DB with the next execution time
		err = r.updateTimeJobNextExecutionTimestamp(timeJobData.JobID, nextExecutionTime)
		if err != nil {
			return nil, err
		}
		
		// Create the task for this job
		taskID, err := r.createTaskDataInDB(&types.CreateTaskDataRequest{
			JobID:            timeJobData.JobID,
			TaskDefinitionID: timeJobData.TaskDefinitionID,
			Network:          timeJobData.Network,
		})
		if err != nil {
			return nil, err
		}

		// Add the task ID to the job
		err = r.addTaskIDToJob(timeJobData.JobID, taskID)
		if err != nil {
			return nil, err
		}

		// Fetch the script storage for the job if TDI = 7
		var scriptStorage []types.ScriptStorageDTO
		if timeJobData.TaskDefinitionID == 7 {
			scriptStorage, err = r.getStorageByJobID(timeJobData.JobID)
			if err != nil {
				return nil, err
			}
		} else {
			scriptStorage = []types.ScriptStorageDTO{}
		}

		targetData := types.TaskTargetData{
			JobID:            timeJobData.JobID,
			TaskID:           taskID,
			TaskDefinitionID: timeJobData.TaskDefinitionID,
			TargetChainID:    timeJobData.TargetChainID,
			TargetContractAddress: timeJobData.TargetContractAddress,
			TargetFunction: timeJobData.TargetFunction,
			ABI: timeJobData.ABI,
			ArgType: timeJobData.ArgType,
			Arguments: timeJobData.Arguments,
			ExecutionScriptURL: timeJobData.ExecutionScriptURL,
			ExecutionScriptLanguage: timeJobData.ExecutionScriptLanguage,
			ExecutionScriptHash: timeJobData.ExecutionScriptHash,
			MaxExecutionTime: timeJobData.MaxExecutionTime,
			ChallengePeriod: timeJobData.ChallengePeriod,
			ScriptStorage: scriptStorage,
		}

		triggerData := types.TaskTriggerData{
			TaskID:                  taskID,
			TaskDefinitionID:        timeJobData.TaskDefinitionID,
			Recurring: false,
			ExpirationTime:          timeJobData.ExpirationTime,
			CurrentTriggerTimestamp: timeJobData.LastExecutedAt,
			NextTriggerTimestamp:    timeJobData.NextExecutionTimestamp, // the original one, not the calculated one
			TimeScheduleType:        timeJobData.ScheduleType,
			TimeCronExpression:      timeJobData.CronExpression,
			TimeSpecificSchedule:    timeJobData.SpecificSchedule,
			TimeInterval:            timeJobData.TimeInterval,
		}

		switch timeJobData.Network {
		case string(types.NetworkMainnet):
			eligibleTimeJobs[0].TaskID = append(eligibleTimeJobs[0].TaskID, taskID)
			eligibleTimeJobs[0].TargetData = append(eligibleTimeJobs[0].TargetData, targetData)
			eligibleTimeJobs[0].TriggerData = append(eligibleTimeJobs[0].TriggerData, triggerData)
		case string(types.NetworkSepolia):
			eligibleTimeJobs[1].TaskID = append(eligibleTimeJobs[1].TaskID, taskID)
			eligibleTimeJobs[1].TargetData = append(eligibleTimeJobs[1].TargetData, targetData)
			eligibleTimeJobs[1].TriggerData = append(eligibleTimeJobs[1].TriggerData, triggerData)
		case string(types.NetworkImua):
			eligibleTimeJobs[2].TaskID = append(eligibleTimeJobs[2].TaskID, taskID)
			eligibleTimeJobs[2].TargetData = append(eligibleTimeJobs[2].TargetData, targetData)
			eligibleTimeJobs[2].TriggerData = append(eligibleTimeJobs[2].TriggerData, triggerData)
		default:
			return nil, errors.New("invalid network")
		}

	}
	return eligibleTimeJobs[:], nil
}

// completeTimeJob marks a time job as completed by updating the job_data status
// and time_job_data is_active field to false
func (r *timeJobRepository) completeTimeJob(jobID string) error {
	err := r.db.Session().Query(updateJobDataToCompletedQuery, jobID).Exec()
	if err != nil {
		return errors.New("failed to update job_data status to completed")
	}
	err = r.db.Session().Query(updateTimeJobStatusQuery, false, jobID).Exec()
	if err != nil {
		return errors.New("failed to update time_job_data is_active to false")
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

// createTaskDataInDB creates a new task record in the database.
// It generates a new task ID by getting the max task ID and incrementing it.
// It also fetches the job_cost_prediction from job_data and sets it as task_opx_predicted_cost.
func (r *timeJobRepository) createTaskDataInDB(task *types.CreateTaskDataRequest) (int64, error) {
	var maxTaskID int64
	err := r.db.Session().Query(getMaxTaskIDQuery).Scan(&maxTaskID)
	if err != nil {
		return -1, errors.New("error getting max task ID")
	}

	// Fetch job_cost_prediction from job_data
	var jobCostPrediction string
	err = r.db.Session().Query(getJobCostPredictionQuery, task.JobID).Scan(&jobCostPrediction)
	if err != nil {
		return -1, errors.New("error getting job cost prediction")
	}

	taskID := maxTaskID + 1
	err = r.db.Session().Query(createTaskDataQuery, taskID, task.JobID, task.TaskDefinitionID, task.Network, string(types.TaskStatusCreated), time.Now().UTC(), jobCostPrediction).Exec()
	if err != nil {
		return -1, errors.New("error creating task data")
	}

	return taskID, nil
}

// addTaskIDToJob adds a task ID to the job's task_ids list.
// It first retrieves existing task IDs, appends the new one, and updates the job.
func (r *timeJobRepository) addTaskIDToJob(jobID string, taskID int64) error {
	var existingTaskIDs []int64
	err := r.db.Session().Query(getTaskIDsByJobIDQuery, jobID).Scan(&existingTaskIDs)
	if err != nil {
		// If no task IDs exist yet (ErrNotFound) or if the field is null, start with an empty slice
		if err == gocql.ErrNotFound {
			existingTaskIDs = []int64{}
		} else {
			// For other errors, return the error
			return errors.New("error getting task IDs by job ID")
		}
	}

	// Append the new task ID
	taskIDs := append(existingTaskIDs, taskID)
	err = r.db.Session().Query(addTaskIDToJobQuery, taskIDs, jobID).Exec()
	if err != nil {
		return errors.New("error adding task ID to job")
	}

	return nil
}

// getStorageByJobID retrieves all storage key-value pairs for a job.
func (r *timeJobRepository) getStorageByJobID(jobID string) ([]types.ScriptStorageDTO, error) {
	iter := r.db.Session().Query(getStorageByJobIDQuery, jobID).Iter()

	storage := []types.ScriptStorageDTO{}
	var key, value string
	var updatedAt time.Time

	for iter.Scan(&key, &value, &updatedAt) {
		storage = append(storage, types.ScriptStorageDTO{
			StorageKey: key,
			StorageValue: value,
			UpdatedAt: updatedAt,
		})
	}

	if err := iter.Close(); err != nil {
		return nil, err
	}

	return storage, nil
}
