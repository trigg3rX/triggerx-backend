package repository

import (
	"errors"

	"github.com/trigg3rX/triggerx-backend/pkg/database"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

type EventJobRepository interface {
	CreateEventJob(eventJob *types.EventJobDataEntity) error
	GetEventJobByJobID(jobID string) (*types.EventJobDataDTO, error)
	UpdateEventJobStatus(jobID string, isActive bool) error
}

type eventJobRepository struct {
	db *database.Connection
}

func NewEventJobRepository(db *database.Connection) EventJobRepository {
	return &eventJobRepository{
		db: db,
	}
}

func (r *eventJobRepository) CreateEventJob(eventJob *types.EventJobDataEntity) error {
	err := r.db.Session().Query(CreateEventJobDataQuery,
		eventJob.JobID, eventJob.TaskDefinitionID, eventJob.Recurring,
		eventJob.TriggerChainID, eventJob.TriggerContractAddress, eventJob.TriggerEvent,
		eventJob.EventFilterParaName, eventJob.EventFilterValue,
		eventJob.TargetChainID, eventJob.TargetContractAddress, eventJob.TargetFunction,
		eventJob.ABI, eventJob.ArgType, eventJob.Arguments, eventJob.DynamicArgumentsScriptURL,
		eventJob.AgentScriptURL, eventJob.AgentScriptLanguage, eventJob.AgentScriptHash,
		eventJob.AgentTargetChainID, eventJob.MaxExecutionTime, eventJob.ChallengePeriod,
		eventJob.IsActive, eventJob.LastExecutedAt, eventJob.ExpirationTime).Exec()

	if err != nil {
		return err
	}

	return nil
}

func (r *eventJobRepository) GetEventJobByJobID(jobID string) (*types.EventJobDataDTO, error) {
	var entity types.EventJobDataEntity
	err := r.db.Session().Query(GetEventJobDataByJobIDQuery, jobID).Scan(
		&entity.JobID, &entity.TaskDefinitionID, &entity.Recurring,
		&entity.TriggerChainID, &entity.TriggerContractAddress, &entity.TriggerEvent,
		&entity.EventFilterParaName, &entity.EventFilterValue,
		&entity.TargetChainID, &entity.TargetContractAddress, &entity.TargetFunction,
		&entity.ABI, &entity.ArgType, &entity.Arguments, &entity.DynamicArgumentsScriptURL,
		&entity.AgentScriptURL, &entity.AgentScriptLanguage, &entity.AgentScriptHash,
		&entity.AgentTargetChainID, &entity.MaxExecutionTime, &entity.ChallengePeriod,
		&entity.IsActive, &entity.LastExecutedAt, &entity.ExpirationTime)
	if err != nil {
		return nil, errors.New("failed to get event job by job ID")
	}

	dto := types.EventJobDataEntityToDTO(&entity)
	return dto, nil
}

func (r *eventJobRepository) UpdateEventJobStatus(jobID string, isActive bool) error {
	err := r.db.Session().Query(UpdateEventJobStatusQuery, isActive, jobID).Exec()
	if err != nil {
		return errors.New("failed to update event job status")
	}

	return nil
}
