package repository

import (
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/trigg3rX/triggerx-backend/pkg/datastore/connection"
	"github.com/trigg3rX/triggerx-backend/pkg/datastore/repository/queries"
	"github.com/trigg3rX/triggerx-backend/pkg/parser"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

type TimeJobRepository interface {
	CreateTimeJob(timeJob *types.TimeJobData) error
	GetTimeJobByJobID(jobID *big.Int) (types.TimeJobData, error)
	CompleteTimeJob(jobID *big.Int) error
	UpdateTimeJobStatus(jobID *big.Int, isActive bool) error
	GetTimeJobsByNextExecutionTimestamp(lookAheadTime time.Time) ([]types.ScheduleTimeTaskData, error)
	UpdateTimeJobNextExecutionTimestamp(jobID *big.Int, nextExecutionTimestamp time.Time) error
	UpdateTimeJobInterval(jobID *big.Int, timeInterval int64) error
	GetActiveTimeJobs() ([]types.TimeJobData, error)
}

type timeJobRepository struct {
	db connection.ConnectionManager
}

func NewTimeJobRepository(db connection.ConnectionManager) TimeJobRepository {
	return &timeJobRepository{
		db: db,
	}
}

func (r *timeJobRepository) CreateTimeJob(timeJob *types.TimeJobData) error {
	err := r.db.GetSession().Query(queries.CreateTimeJobDataQuery,
		timeJob.JobID, timeJob.TaskDefinitionID, timeJob.ExpirationTime, timeJob.NextExecutionTimestamp,
		timeJob.ScheduleType, timeJob.TimeInterval, timeJob.CronExpression, timeJob.SpecificSchedule,
		timeJob.TargetChainID, timeJob.TargetContractAddress, timeJob.TargetFunction,
		timeJob.ABI, timeJob.ArgType, timeJob.Arguments, timeJob.DynamicArgumentsScriptUrl,
		timeJob.IsCompleted, timeJob.ExpirationTime, time.Now()).Exec()

	if err != nil {
		return err
	}

	return nil
}

func (r *timeJobRepository) GetTimeJobByJobID(jobID *big.Int) (types.TimeJobData, error) {
	var timeJob types.TimeJobData
	var temp *big.Int
	err := r.db.GetSession().Query(queries.GetTimeJobDataByJobIDQuery, jobID).Scan(
		&temp, &timeJob.ExpirationTime, &timeJob.NextExecutionTimestamp,
		&timeJob.ScheduleType, &timeJob.TimeInterval, &timeJob.CronExpression,
		&timeJob.SpecificSchedule, &timeJob.TargetChainID,
		&timeJob.TargetContractAddress, &timeJob.TargetFunction, &timeJob.ABI, &timeJob.ArgType,
		&timeJob.Arguments, &timeJob.DynamicArgumentsScriptUrl, &timeJob.IsCompleted)
	if err != nil {
		return types.TimeJobData{}, fmt.Errorf("failed to get time job by job ID: %v", err)
	}
	timeJob.JobID = jobID
	return timeJob, nil
}

func (r *timeJobRepository) CompleteTimeJob(jobID *big.Int) error {
	err := r.db.GetSession().Query(queries.CompleteTimeJobStatusQuery, jobID).Exec()
	if err != nil {
		return errors.New("failed to complete time job")
	}

	err = r.db.GetSession().Query(queries.UpdateJobDataToCompletedQuery, jobID).Exec()
	if err != nil {
		return errors.New("failed to update job_data status to completed")
	}

	return nil
}

func (r *timeJobRepository) UpdateTimeJobStatus(jobID *big.Int, isActive bool) error {
	err := r.db.GetSession().Query(queries.UpdateTimeJobStatusQuery, isActive, jobID).Exec()
	if err != nil {
		return errors.New("failed to update time job status")
	}

	return nil
}

func (r *timeJobRepository) GetTimeJobsByNextExecutionTimestamp(lookAheadTime time.Time) ([]types.ScheduleTimeTaskData, error) {
	currentTime := time.Now()
	iter := r.db.GetSession().Query(queries.GetTimeJobsByNextExecutionTimestampQuery, currentTime, lookAheadTime).Iter()

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
		err := r.db.GetSession().Query(queries.IsJobImuaQuery, jobIDBigInt).Scan(&isImua)
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
			err = r.CompleteTimeJob(timeJob.TaskTargetData.JobID.Int)
			if err != nil {
				return nil, err
			}
			err = r.UpdateTimeJobStatus(timeJob.TaskTargetData.JobID.Int, false)
			if err != nil {
				return nil, err
			}
		} else {
			err = r.UpdateTimeJobNextExecutionTimestamp(timeJob.TaskTargetData.JobID.Int, nextExecutionTime)
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
	err := r.db.GetSession().Query(queries.UpdateTimeJobNextExecutionTimestampQuery, nextExecutionTimestamp, jobID).Exec()
	if err != nil {
		return errors.New("failed to update time job next execution timestamp")
	}

	return nil
}

func (r *timeJobRepository) UpdateTimeJobInterval(jobID *big.Int, timeInterval int64) error {
	err := r.db.GetSession().Query(queries.UpdateTimeJobIntervalQuery, timeInterval, jobID).Exec()
	if err != nil {
		return errors.New("failed to update time_interval in time_job_data")
	}
	return nil
}

func (r *timeJobRepository) GetActiveTimeJobs() ([]types.TimeJobData, error) {
	var timeJobs []types.TimeJobData
	iter := r.db.GetSession().Query(queries.GetActiveTimeJobsQuery).Iter()
	var timeJob types.TimeJobData
	var jobIDBigInt *big.Int
	for iter.Scan(
		&jobIDBigInt, &timeJob.ExpirationTime, &timeJob.NextExecutionTimestamp, &timeJob.ScheduleType,
		&timeJob.TimeInterval, &timeJob.CronExpression, &timeJob.SpecificSchedule,
		&timeJob.TargetChainID, &timeJob.TargetContractAddress, &timeJob.TargetFunction, &timeJob.ABI, &timeJob.ArgType,
		&timeJob.Arguments, &timeJob.DynamicArgumentsScriptUrl, &timeJob.IsCompleted) {
		timeJob.JobID = jobIDBigInt
		timeJobs = append(timeJobs, timeJob)
	}
	if err := iter.Close(); err != nil {
		return nil, err
	}
	return timeJobs, nil
}
