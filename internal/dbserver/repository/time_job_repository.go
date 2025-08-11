package repository

import (
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/trigg3rX/triggerx-backend/internal/dbserver/repository/queries"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/types"
	"github.com/trigg3rX/triggerx-backend/pkg/database"
	"github.com/trigg3rX/triggerx-backend/pkg/parser"
	commonTypes "github.com/trigg3rX/triggerx-backend/pkg/types"
)

type TimeJobRepository interface {
	CreateTimeJob(timeJob *types.TimeJobData) error
	GetTimeJobByJobID(jobID *big.Int) (types.TimeJobData, error)
	CompleteTimeJob(jobID *big.Int) error
	UpdateTimeJobStatus(jobID *big.Int, isActive bool) error
	GetTimeJobsByNextExecutionTimestamp(lookAheadTime time.Time) ([]commonTypes.ScheduleTimeTaskData, error)
	UpdateTimeJobNextExecutionTimestamp(jobID *big.Int, nextExecutionTimestamp time.Time) error
	UpdateTimeJobInterval(jobID *big.Int, timeInterval int64) error
	GetActiveTimeJobs() ([]types.TimeJobData, error)
}

type timeJobRepository struct {
	db *database.Connection
}

func NewTimeJobRepository(db *database.Connection) TimeJobRepository {
	return &timeJobRepository{
		db: db,
	}
}

func (r *timeJobRepository) CreateTimeJob(timeJob *types.TimeJobData) error {
	err := r.db.Session().Query(queries.CreateTimeJobDataQuery,
		timeJob.JobID, timeJob.TaskDefinitionID, timeJob.ExpirationTime, timeJob.NextExecutionTimestamp,
		timeJob.ScheduleType, timeJob.TimeInterval, timeJob.CronExpression, timeJob.SpecificSchedule,
		timeJob.Timezone, timeJob.TargetChainID, timeJob.TargetContractAddress, timeJob.TargetFunction,
		timeJob.ABI, timeJob.ArgType, timeJob.Arguments, timeJob.DynamicArgumentsScriptUrl,
		timeJob.IsCompleted, timeJob.IsActive, time.Now(), time.Now()).Exec()

	if err != nil {
		return err
	}

	return nil
}

func (r *timeJobRepository) GetTimeJobByJobID(jobID *big.Int) (types.TimeJobData, error) {
	var timeJob types.TimeJobData
	err := r.db.Session().Query(queries.GetTimeJobDataByJobIDQuery, jobID).Scan(
		&timeJob.JobID, &timeJob.ExpirationTime, &timeJob.NextExecutionTimestamp,
		&timeJob.ScheduleType, &timeJob.TimeInterval, &timeJob.CronExpression,
		&timeJob.SpecificSchedule, &timeJob.Timezone, &timeJob.TargetChainID,
		&timeJob.TargetContractAddress, &timeJob.TargetFunction, &timeJob.ABI, &timeJob.ArgType,
		&timeJob.Arguments, &timeJob.DynamicArgumentsScriptUrl, &timeJob.IsCompleted, &timeJob.IsActive)
	if err != nil {
		return types.TimeJobData{}, fmt.Errorf("failed to get time job by job ID: %v", err)
	}

	return timeJob, nil
}

func (r *timeJobRepository) CompleteTimeJob(jobID *big.Int) error {
	err := r.db.Session().Query(queries.CompleteTimeJobStatusQuery, jobID).Exec()
	if err != nil {
		return errors.New("failed to complete time job")
	}

	err = r.db.Session().Query(queries.UpdateJobDataToCompletedQuery, jobID).Exec()
	if err != nil {
		return errors.New("failed to update job_data status to completed")
	}

	return nil
}

func (r *timeJobRepository) UpdateTimeJobStatus(jobID *big.Int, isActive bool) error {
	err := r.db.Session().Query(queries.UpdateTimeJobStatusQuery, isActive, jobID).Exec()
	if err != nil {
		return errors.New("failed to update time job status")
	}

	return nil
}

func (r *timeJobRepository) GetTimeJobsByNextExecutionTimestamp(lookAheadTime time.Time) ([]commonTypes.ScheduleTimeTaskData, error) {
	currentTime := time.Now()
	iter := r.db.Session().Query(queries.GetTimeJobsByNextExecutionTimestampQuery, currentTime, lookAheadTime).Iter()

	var timeJobs []commonTypes.ScheduleTimeTaskData
	var timeJob commonTypes.ScheduleTimeTaskData

	for iter.Scan(
		&timeJob.TaskTargetData.JobID, &timeJob.LastExecutedAt, &timeJob.ExpirationTime, &timeJob.TimeInterval,
		&timeJob.ScheduleType, &timeJob.CronExpression, &timeJob.SpecificSchedule, &timeJob.NextExecutionTimestamp,
		&timeJob.TaskTargetData.TargetChainID, &timeJob.TaskTargetData.TargetContractAddress, &timeJob.TaskTargetData.TargetFunction, &timeJob.TaskTargetData.ABI, &timeJob.TaskTargetData.ArgType,
		&timeJob.TaskTargetData.Arguments, &timeJob.TaskTargetData.DynamicArgumentsScriptUrl,
	) {
		if timeJob.TaskTargetData.DynamicArgumentsScriptUrl != "" {
			timeJob.TaskDefinitionID = 2
			timeJob.TaskTargetData.TaskDefinitionID = 2
		} else {
			timeJob.TaskDefinitionID = 1
			timeJob.TaskTargetData.TaskDefinitionID = 1
		}

		var isImua bool
		err := r.db.Session().Query(queries.IsJobImuaQuery, timeJob.TaskTargetData.JobID).Scan(&isImua)
		if err != nil {
			return nil, err
		}
		timeJob.IsImua = isImua

		// Calculate next execution time after the current execution time
		nextExecutionTime, err := parser.CalculateNextExecutionTime(timeJob.NextExecutionTimestamp, timeJob.ScheduleType, timeJob.TimeInterval, timeJob.CronExpression, timeJob.SpecificSchedule)
		if err != nil {
			return nil, err
		}

		// If the next execution time is after the expiration time, That means the job will be completed after current execution time that is being passed
		if nextExecutionTime.After(timeJob.ExpirationTime) {
			err = r.CompleteTimeJob(timeJob.TaskTargetData.JobID)
			if err != nil {
				return nil, err
			}
			err = r.UpdateTimeJobStatus(timeJob.TaskTargetData.JobID, false)
			if err != nil {
				return nil, err
			}
		} else {
			err = r.UpdateTimeJobNextExecutionTimestamp(timeJob.TaskTargetData.JobID, nextExecutionTime)
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

func (r *timeJobRepository) UpdateTimeJobNextExecutionTimestamp(jobID *big.Int, nextExecutionTimestamp time.Time) error {
	err := r.db.Session().Query(queries.UpdateTimeJobNextExecutionTimestampQuery, nextExecutionTimestamp, jobID).Exec()
	if err != nil {
		return errors.New("failed to update time job next execution timestamp")
	}

	return nil
}

func (r *timeJobRepository) UpdateTimeJobInterval(jobID *big.Int, timeInterval int64) error {
	err := r.db.Session().Query(queries.UpdateTimeJobIntervalQuery, timeInterval, jobID).Exec()
	if err != nil {
		return errors.New("failed to update time_interval in time_job_data")
	}
	return nil
}

func (r *timeJobRepository) GetActiveTimeJobs() ([]types.TimeJobData, error) {
	var timeJobs []types.TimeJobData
	iter := r.db.Session().Query(queries.GetActiveTimeJobsQuery).Iter()
	var timeJob types.TimeJobData
	for iter.Scan(
		&timeJob.JobID, &timeJob.ExpirationTime, &timeJob.NextExecutionTimestamp, &timeJob.ScheduleType,
		&timeJob.TimeInterval, &timeJob.CronExpression, &timeJob.SpecificSchedule, &timeJob.Timezone,
		&timeJob.TargetChainID, &timeJob.TargetContractAddress, &timeJob.TargetFunction, &timeJob.ABI, &timeJob.ArgType,
		&timeJob.Arguments, &timeJob.DynamicArgumentsScriptUrl, &timeJob.IsCompleted, &timeJob.IsActive) {
		timeJobs = append(timeJobs, timeJob)
	}
	if err := iter.Close(); err != nil {
		return nil, err
	}
	return timeJobs, nil
}
