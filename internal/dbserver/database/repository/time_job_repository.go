package repository

import (
	"errors"
	"fmt"

	"github.com/trigg3rX/triggerx-backend/pkg/database"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

type TimeJobRepository interface {
	CreateTimeJob(timeJob *types.TimeJobDataEntity) error
	GetTimeJobByJobID(jobID string) (*types.TimeJobDataDTO, error)
	UpdateTimeJobStatus(jobID string, isActive bool) error
	UpdateTimeJobInterval(jobID string, timeInterval int64) error
}

type timeJobRepository struct {
	db *database.Connection
}

func NewTimeJobRepository(db *database.Connection) TimeJobRepository {
	return &timeJobRepository{
		db: db,
	}
}

func (r *timeJobRepository) CreateTimeJob(timeJob *types.TimeJobDataEntity) error {
	err := r.db.Session().Query(CreateTimeJobDataQuery,
		timeJob.JobID, timeJob.TaskDefinitionID, timeJob.ScheduleType, timeJob.TimeInterval,
		timeJob.CronExpression, timeJob.SpecificSchedule, timeJob.Timezone, timeJob.NextExecutionTimestamp,
		timeJob.TargetChainID, timeJob.TargetContractAddress, timeJob.TargetFunction,
		timeJob.ABI, timeJob.ArgType, timeJob.Arguments, timeJob.DynamicArgumentsScriptURL,
		timeJob.AgentScriptURL, timeJob.AgentScriptLanguage, timeJob.AgentScriptHash,
		timeJob.AgentTargetChainID, timeJob.MaxExecutionTime, timeJob.ChallengePeriod,
		timeJob.IsActive, timeJob.LastExecutedAt, timeJob.ExpirationTime).Exec()

	if err != nil {
		return err
	}

	return nil
}

func (r *timeJobRepository) GetTimeJobByJobID(jobID string) (*types.TimeJobDataDTO, error) {
	var entity types.TimeJobDataEntity
	err := r.db.Session().Query(GetTimeJobDataByJobIDQuery, jobID).Scan(
		&entity.JobID, &entity.TaskDefinitionID, &entity.ScheduleType, &entity.TimeInterval,
		&entity.CronExpression, &entity.SpecificSchedule, &entity.Timezone, &entity.NextExecutionTimestamp,
		&entity.TargetChainID, &entity.TargetContractAddress, &entity.TargetFunction,
		&entity.ABI, &entity.ArgType, &entity.Arguments, &entity.DynamicArgumentsScriptURL,
		&entity.AgentScriptURL, &entity.AgentScriptLanguage, &entity.AgentScriptHash,
		&entity.AgentTargetChainID, &entity.MaxExecutionTime, &entity.ChallengePeriod,
		&entity.IsActive, &entity.LastExecutedAt, &entity.ExpirationTime)
	if err != nil {
		return nil, fmt.Errorf("failed to get time job by job ID: %v", err)
	}

	dto := types.TimeJobDataEntityToDTO(&entity)
	return dto, nil
}

func (r *timeJobRepository) UpdateTimeJobStatus(jobID string, isActive bool) error {
	err := r.db.Session().Query(UpdateTimeJobStatusQuery, isActive, jobID).Exec()
	if err != nil {
		return errors.New("failed to update time job status")
	}

	return nil
}

func (r *timeJobRepository) UpdateTimeJobInterval(jobID string, timeInterval int64) error {
	err := r.db.Session().Query(UpdateTimeJobIntervalQuery, timeInterval, jobID).Exec()
	if err != nil {
		return errors.New("failed to update time_interval in time_job_data")
	}
	return nil
}
