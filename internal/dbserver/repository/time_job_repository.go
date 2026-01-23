package repository

import (
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/trigg3rX/triggerx-backend/internal/dbserver/repository/queries"
	"github.com/trigg3rX/triggerx-backend/pkg/database"
	commonTypes "github.com/trigg3rX/triggerx-backend/pkg/types"
)

type TimeJobRepository interface {
	CreateTimeJob(timeJob *commonTypes.TimeJobData) error
	GetTimeJobByJobID(jobID *big.Int) (commonTypes.TimeJobData, error)
	UpdateTimeJobStatus(jobID *big.Int, isActive bool) error
	UpdateTimeJobInterval(jobID *big.Int, timeInterval int64) error
}

type timeJobRepository struct {
	db *database.Connection
}

func NewTimeJobRepository(db *database.Connection) TimeJobRepository {
	return &timeJobRepository{
		db: db,
	}
}

func (r *timeJobRepository) CreateTimeJob(timeJob *commonTypes.TimeJobData) error {
	err := r.db.Session().Query(queries.CreateTimeJobDataQuery,
		timeJob.JobID.ToBigInt(), timeJob.TaskDefinitionID, timeJob.ExpirationTime, timeJob.NextExecutionTimestamp,
		timeJob.ScheduleType, timeJob.TimeInterval, timeJob.CronExpression, timeJob.SpecificSchedule,
		timeJob.Timezone, timeJob.TargetChainID, timeJob.TargetContractAddress, timeJob.TargetFunction,
		timeJob.ABI, timeJob.ArgType, timeJob.Arguments, timeJob.DynamicArgumentsScriptUrl,
		timeJob.IsCompleted, timeJob.IsActive, time.Now(), time.Now()).Exec()

	if err != nil {
		return err
	}

	return nil
}

func (r *timeJobRepository) GetTimeJobByJobID(jobID *big.Int) (commonTypes.TimeJobData, error) {
	var timeJob commonTypes.TimeJobData
	var temp *big.Int
	err := r.db.Session().Query(queries.GetTimeJobDataByJobIDQuery, jobID).Scan(
		&temp, &timeJob.ExpirationTime, &timeJob.NextExecutionTimestamp,
		&timeJob.ScheduleType, &timeJob.TimeInterval, &timeJob.CronExpression,
		&timeJob.SpecificSchedule, &timeJob.Timezone, &timeJob.TargetChainID,
		&timeJob.TargetContractAddress, &timeJob.TargetFunction, &timeJob.ABI, &timeJob.ArgType,
		&timeJob.Arguments, &timeJob.DynamicArgumentsScriptUrl, &timeJob.IsCompleted, &timeJob.IsActive)
	if err != nil {
		return commonTypes.TimeJobData{}, fmt.Errorf("failed to get time job by job ID: %v", err)
	}
	timeJob.JobID = commonTypes.NewBigInt(jobID)
	return timeJob, nil
}

func (r *timeJobRepository) UpdateTimeJobStatus(jobID *big.Int, isActive bool) error {
	err := r.db.Session().Query(queries.UpdateTimeJobStatusQuery, isActive, jobID).Exec()
	if err != nil {
		return errors.New("failed to update time job status")
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
