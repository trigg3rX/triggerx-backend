package repository

import (
	"errors"
	"math/big"
	"time"

	"github.com/trigg3rX/triggerx-backend/internal/dbserver/repository/queries"
	"github.com/trigg3rX/triggerx-backend/pkg/database"
	commonTypes "github.com/trigg3rX/triggerx-backend/pkg/types"
)

type EventJobRepository interface {
	CreateEventJob(eventJob *commonTypes.EventJobData) error
	GetEventJobByJobID(jobID *big.Int) (commonTypes.EventJobData, error)
	UpdateEventJobStatus(jobID *big.Int, isActive bool) error
}

type eventJobRepository struct {
	db *database.Connection
}

func NewEventJobRepository(db *database.Connection) EventJobRepository {
	return &eventJobRepository{
		db: db,
	}
}

func (r *eventJobRepository) CreateEventJob(eventJob *commonTypes.EventJobData) error {
	err := r.db.Session().Query(queries.CreateEventJobDataQuery,
		eventJob.JobID.ToBigInt(), eventJob.TaskDefinitionID, eventJob.ExpirationTime, eventJob.Recurring,
		eventJob.TriggerChainID, eventJob.TriggerContractAddress, eventJob.TriggerEvent,
		eventJob.EventFilterParaName, eventJob.EventFilterValue,
		eventJob.TargetChainID, eventJob.TargetContractAddress, eventJob.TargetFunction,
		eventJob.ABI, eventJob.ArgType, eventJob.Arguments, eventJob.DynamicArgumentsScriptUrl,
		eventJob.IsCompleted, eventJob.IsActive, time.Now(), time.Now()).Exec()

	if err != nil {
		return err
	}

	return nil
}

func (r *eventJobRepository) GetEventJobByJobID(jobID *big.Int) (commonTypes.EventJobData, error) {
	var eventJob commonTypes.EventJobData
	var temp *big.Int
	eventJob.JobID = commonTypes.NewBigInt(jobID)
	err := r.db.Session().Query(queries.GetEventJobDataByJobIDQuery, jobID).Scan(
		&temp, &eventJob.ExpirationTime, &eventJob.Recurring, &eventJob.TriggerChainID,
		&eventJob.TriggerContractAddress, &eventJob.TriggerEvent, &eventJob.EventFilterParaName, &eventJob.EventFilterValue,
		&eventJob.TargetChainID, &eventJob.TargetContractAddress, &eventJob.TargetFunction, &eventJob.ABI, &eventJob.ArgType,
		&eventJob.Arguments, &eventJob.DynamicArgumentsScriptUrl, &eventJob.IsCompleted, &eventJob.IsActive)
	if err != nil {
		return commonTypes.EventJobData{}, errors.New("failed to get event job by job ID")
	}

	return eventJob, nil
}

func (r *eventJobRepository) UpdateEventJobStatus(jobID *big.Int, isActive bool) error {
	err := r.db.Session().Query(queries.UpdateEventJobStatusQuery, isActive, jobID).Exec()
	if err != nil {
		return errors.New("failed to update event job status")
	}

	return nil
}
