package repository

import (
	"errors"
	"fmt"
	"strings"

	"github.com/trigg3rX/triggerx-backend/pkg/database"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

type SpecificJobRepository interface {
	CreateTimeJob(timeJob *types.TimeJobDataEntity) error
	CreateEventJob(eventJob *types.EventJobDataEntity) error
	CreateConditionJob(conditionJob *types.ConditionJobDataEntity) error

	GetTimeJobByJobID(jobID string) (*types.TimeJobDataDTO, error)
	GetEventJobByJobID(jobID string) (*types.EventJobDataDTO, error)
	GetConditionJobByJobID(jobID string) (*types.ConditionJobDataDTO, error)

	UpdateTimeJobInterval(jobID string, timeInterval int64) error
	UpdateTimeJobStatus(jobID string, isActive bool) error
	UpdateEventJobStatus(jobID string, isActive bool) error
	UpdateConditionJobStatus(jobID string, isActive bool) error
}

type specificJobRepository struct {
	db *database.Connection
}

func NewSpecificJobRepository(db *database.Connection) SpecificJobRepository {
	return &specificJobRepository{
		db: db,
	}
}

func (r *specificJobRepository) CreateTimeJob(timeJob *types.TimeJobDataEntity) error {
	err := r.db.Session().Query(CreateTimeJobDataQuery,
		timeJob.JobID, timeJob.TaskDefinitionID, timeJob.Network, timeJob.ScheduleType, timeJob.TimeInterval,
		timeJob.CronExpression, timeJob.SpecificSchedule, timeJob.Timezone, timeJob.NextExecutionTimestamp,
		timeJob.TargetChainID, strings.ToLower(timeJob.TargetContractAddress), timeJob.TargetFunction,
		timeJob.ABI, timeJob.ArgType, timeJob.Arguments, timeJob.ExecutionScriptURL, timeJob.ExecutionScriptLanguage, 
		timeJob.ExecutionScriptHash, timeJob.MaxExecutionTime, timeJob.ChallengePeriod,
		timeJob.IsActive, timeJob.LastExecutedAt, timeJob.ExpirationTime).Exec()
	if err != nil {
		return err
	}

	return nil
}

func (r *specificJobRepository) CreateEventJob(eventJob *types.EventJobDataEntity) error {
	err := r.db.Session().Query(CreateEventJobDataQuery,
		eventJob.JobID, eventJob.TaskDefinitionID, eventJob.Network, eventJob.Recurring,
		eventJob.TriggerChainID, strings.ToLower(eventJob.TriggerContractAddress), eventJob.TriggerEvent,
		eventJob.EventFilterParaName, eventJob.EventFilterValue,
		eventJob.TargetChainID, strings.ToLower(eventJob.TargetContractAddress), eventJob.TargetFunction,
		eventJob.ABI, eventJob.ArgType, eventJob.Arguments, eventJob.ExecutionScriptURL, eventJob.ExecutionScriptLanguage, 
		eventJob.ExecutionScriptHash, eventJob.MaxExecutionTime, eventJob.ChallengePeriod,
		eventJob.IsActive, eventJob.LastExecutedAt, eventJob.ExpirationTime).Exec()
	if err != nil {
		return err
	}

	return nil
}

func (r *specificJobRepository) CreateConditionJob(conditionJob *types.ConditionJobDataEntity) error {
	err := r.db.Session().Query(CreateConditionJobDataQuery,
		conditionJob.JobID, conditionJob.TaskDefinitionID, conditionJob.Network, conditionJob.Recurring,
		conditionJob.ConditionType, conditionJob.UpperLimit, conditionJob.LowerLimit,
		conditionJob.ValueSourceType, conditionJob.ValueSourceURL, conditionJob.SelectedKeyRoute,
		conditionJob.TargetChainID, strings.ToLower(conditionJob.TargetContractAddress), conditionJob.TargetFunction,
		conditionJob.ABI, conditionJob.ArgType, conditionJob.Arguments,
		conditionJob.ExecutionScriptURL, conditionJob.ExecutionScriptLanguage, conditionJob.ExecutionScriptHash,
		conditionJob.MaxExecutionTime, conditionJob.ChallengePeriod,
		conditionJob.IsActive, conditionJob.LastExecutedAt, conditionJob.ExpirationTime).Exec()

	if err != nil {
		return err
	}

	return nil
}

func (r *specificJobRepository) GetTimeJobByJobID(jobID string) (*types.TimeJobDataDTO, error) {
	var timeJob types.TimeJobDataEntity
	err := r.db.Session().Query(GetTimeJobDataByJobIDQuery, jobID).Scan(
		timeJob.JobID, timeJob.TaskDefinitionID, timeJob.Network, timeJob.ScheduleType, timeJob.TimeInterval,
		timeJob.CronExpression, timeJob.SpecificSchedule, timeJob.Timezone, timeJob.NextExecutionTimestamp,
		timeJob.TargetChainID, timeJob.TargetContractAddress, timeJob.TargetFunction,
		timeJob.ABI, timeJob.ArgType, timeJob.Arguments, timeJob.ExecutionScriptURL, timeJob.ExecutionScriptLanguage, 
		timeJob.ExecutionScriptHash, timeJob.MaxExecutionTime, timeJob.ChallengePeriod,
		timeJob.IsActive, timeJob.LastExecutedAt, timeJob.ExpirationTime)
	if err != nil {
		return nil, fmt.Errorf("failed to get time job by job ID: %v", err)
	}

	dto := types.TimeJobDataEntityToDTO(&timeJob)
	return dto, nil
}

func (r *specificJobRepository) GetEventJobByJobID(jobID string) (*types.EventJobDataDTO, error) {
	var eventJob types.EventJobDataEntity
	err := r.db.Session().Query(GetEventJobDataByJobIDQuery, jobID).Scan(
		eventJob.JobID, eventJob.TaskDefinitionID, eventJob.Network, eventJob.Recurring,
		eventJob.TriggerChainID, eventJob.TriggerContractAddress, eventJob.TriggerEvent,
		eventJob.EventFilterParaName, eventJob.EventFilterValue,
		eventJob.TargetChainID, eventJob.TargetContractAddress, eventJob.TargetFunction,
		eventJob.ABI, eventJob.ArgType, eventJob.Arguments, eventJob.ExecutionScriptURL, eventJob.ExecutionScriptLanguage, 
		eventJob.ExecutionScriptHash, eventJob.MaxExecutionTime, eventJob.ChallengePeriod,
		eventJob.IsActive, eventJob.LastExecutedAt, eventJob.ExpirationTime)
	if err != nil {
		return nil, errors.New("failed to get event job by job ID")
	}

	dto := types.EventJobDataEntityToDTO(&eventJob)
	return dto, nil
}

func (r *specificJobRepository) GetConditionJobByJobID(jobID string) (*types.ConditionJobDataDTO, error) {
	var conditionJob types.ConditionJobDataEntity
	err := r.db.Session().Query(GetConditionJobDataByJobIDQuery, jobID).Scan(
		conditionJob.JobID, conditionJob.TaskDefinitionID, conditionJob.Network, conditionJob.Recurring,
		conditionJob.ConditionType, conditionJob.UpperLimit, conditionJob.LowerLimit,
		conditionJob.ValueSourceType, conditionJob.ValueSourceURL, conditionJob.SelectedKeyRoute,
		conditionJob.TargetChainID, conditionJob.TargetContractAddress, conditionJob.TargetFunction,
		conditionJob.ABI, conditionJob.ArgType, conditionJob.Arguments,
		conditionJob.ExecutionScriptURL, conditionJob.ExecutionScriptLanguage, conditionJob.ExecutionScriptHash,
		conditionJob.MaxExecutionTime, conditionJob.ChallengePeriod,
		conditionJob.IsActive, conditionJob.LastExecutedAt, conditionJob.ExpirationTime)
	if err != nil {
		return nil, errors.New("failed to get condition job by job ID")
	}

	dto := types.ConditionJobDataEntityToDTO(&conditionJob)
	return dto, nil
}

func (r *specificJobRepository) UpdateTimeJobInterval(jobID string, timeInterval int64) error {
	err := r.db.Session().Query(UpdateTimeJobIntervalQuery, timeInterval, jobID).Exec()
	if err != nil {
		return errors.New("failed to update time_interval in time_job_data")
	}
	return nil
}

func (r *specificJobRepository) UpdateTimeJobStatus(jobID string, isActive bool) error {
	err := r.db.Session().Query(UpdateTimeJobStatusQuery, isActive, jobID).Exec()
	if err != nil {
		return errors.New("failed to update time job status")
	}

	return nil
}

func (r *specificJobRepository) UpdateEventJobStatus(jobID string, isActive bool) error {
	err := r.db.Session().Query(UpdateEventJobStatusQuery, isActive, jobID).Exec()
	if err != nil {
		return errors.New("failed to update event job status")
	}

	return nil
}

func (r *specificJobRepository) UpdateConditionJobStatus(jobID string, isActive bool) error {
	err := r.db.Session().Query(UpdateConditionJobStatusQuery, isActive, jobID).Exec()
	if err != nil {
		return errors.New("failed to update condition job status")
	}

	return nil
}
