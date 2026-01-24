package repository

import (
	"errors"

	"github.com/trigg3rX/triggerx-backend/pkg/database"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

type ConditionJobRepository interface {
	CreateConditionJob(conditionJob *types.ConditionJobDataEntity) error
	GetConditionJobByJobID(jobID string) (*types.ConditionJobDataDTO, error)
	UpdateConditionJobStatus(jobID string, isActive bool) error
}

type conditionJobRepository struct {
	db *database.Connection
}

func NewConditionJobRepository(db *database.Connection) ConditionJobRepository {
	return &conditionJobRepository{
		db: db,
	}
}

func (r *conditionJobRepository) CreateConditionJob(conditionJob *types.ConditionJobDataEntity) error {
	err := r.db.Session().Query(CreateConditionJobDataQuery,
		conditionJob.JobID, conditionJob.TaskDefinitionID, conditionJob.Recurring,
		conditionJob.ConditionType, conditionJob.UpperLimit, conditionJob.LowerLimit,
		conditionJob.ValueSourceType, conditionJob.ValueSourceURL, conditionJob.SelectedKeyRoute,
		conditionJob.TargetChainID, conditionJob.TargetContractAddress, conditionJob.TargetFunction,
		conditionJob.ABI, conditionJob.ArgType, conditionJob.Arguments,
		conditionJob.DynamicArgumentsScriptURL, conditionJob.AgentScriptURL,
		conditionJob.AgentScriptLanguage, conditionJob.AgentScriptHash,
		conditionJob.AgentTargetChainID, conditionJob.MaxExecutionTime, conditionJob.ChallengePeriod,
		conditionJob.IsActive, conditionJob.LastExecutedAt, conditionJob.ExpirationTime).Exec()

	if err != nil {
		return err
	}

	return nil
}

func (r *conditionJobRepository) GetConditionJobByJobID(jobID string) (*types.ConditionJobDataDTO, error) {
	var entity types.ConditionJobDataEntity
	err := r.db.Session().Query(GetConditionJobDataByJobIDQuery, jobID).Scan(
		&entity.JobID, &entity.TaskDefinitionID, &entity.Recurring, &entity.ConditionType,
		&entity.UpperLimit, &entity.LowerLimit, &entity.ValueSourceType,
		&entity.ValueSourceURL, &entity.SelectedKeyRoute, &entity.TargetChainID,
		&entity.TargetContractAddress, &entity.TargetFunction, &entity.ABI, &entity.ArgType,
		&entity.Arguments, &entity.DynamicArgumentsScriptURL,
		&entity.AgentScriptURL, &entity.AgentScriptLanguage, &entity.AgentScriptHash,
		&entity.AgentTargetChainID, &entity.MaxExecutionTime, &entity.ChallengePeriod,
		&entity.IsActive, &entity.LastExecutedAt, &entity.ExpirationTime)
	if err != nil {
		return nil, errors.New("failed to get condition job by job ID")
	}

	dto := types.ConditionJobDataEntityToDTO(&entity)
	return dto, nil
}

func (r *conditionJobRepository) UpdateConditionJobStatus(jobID string, isActive bool) error {
	err := r.db.Session().Query(UpdateConditionJobStatusQuery, isActive, jobID).Exec()
	if err != nil {
		return errors.New("failed to update condition job status")
	}

	return nil
}
