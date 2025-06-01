package repository

import (
	"errors"
	"time"

	"github.com/trigg3rX/triggerx-backend/internal/dbserver/repository/queries"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/types"
	"github.com/trigg3rX/triggerx-backend/pkg/database"
)

type TimeJobRepository interface {
	CreateTimeJob(timeJob *types.TimeJobData) error
	GetTimeJobByJobID(jobID int64) (types.TimeJobData, error)
	CompleteTimeJob(jobID int64) error
	UpdateTimeJobStatus(jobID int64, isActive bool) error
	GetTimeJobsByNextExecutionTimestamp(nextExecutionTimestamp time.Time) ([]types.TimeJobData, error)
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
		timeJob.JobID, timeJob.TimeFrame, timeJob.Recurring, timeJob.ScheduleType, timeJob.TimeInterval, timeJob.CronExpression,
		timeJob.SpecificSchedule, timeJob.NextExecutionTimestamp, timeJob.TargetChainID,
		timeJob.TargetContractAddress, timeJob.TargetFunction, timeJob.ABI, timeJob.ArgType, timeJob.Arguments,
		timeJob.DynamicArgumentsScriptUrl).Exec()

	if err != nil {
		return err
	}

	return nil
}

func (r *timeJobRepository) GetTimeJobByJobID(jobID int64) (types.TimeJobData, error) {
	var timeJob types.TimeJobData
	err := r.db.Session().Query(queries.GetTimeJobDataByJobIDQuery, jobID).Scan(&timeJob)
	if err != nil {
		return types.TimeJobData{}, errors.New("failed to get time job by job ID")
	}

	return timeJob, nil
}

func (r *timeJobRepository) CompleteTimeJob(jobID int64) error {
	err := r.db.Session().Query(queries.CompleteTimeJobStatusQuery, jobID).Exec()
	if err != nil {
		return errors.New("failed to complete time job")
	}

	return nil
}

func (r *timeJobRepository) UpdateTimeJobStatus(jobID int64, isActive bool) error {
	err := r.db.Session().Query(queries.UpdateTimeJobStatusQuery, isActive, jobID).Exec()
	if err != nil {
		return errors.New("failed to update time job status")
	}

	return nil
}

func (r *timeJobRepository) GetTimeJobsByNextExecutionTimestamp(nextExecutionTimestamp time.Time) ([]types.TimeJobData, error) {
	var timeJobs []types.TimeJobData
	err := r.db.Session().Query(queries.GetTimeJobsByNextExecutionTimestampQuery, nextExecutionTimestamp).Scan(&timeJobs)
	if err != nil {
		return []types.TimeJobData{}, errors.New("failed to get time jobs by next execution timestamp")
	}

	return timeJobs, nil
}