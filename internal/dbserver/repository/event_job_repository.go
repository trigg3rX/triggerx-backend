package repository

import (
	"errors"

	"github.com/trigg3rX/triggerx-backend/internal/dbserver/repository/queries"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/types"
	"github.com/trigg3rX/triggerx-backend/pkg/database"
)

type EventJobRepository interface {
	CreateEventJob(eventJob *types.EventJobData) error
	GetEventJobByJobID(jobID int64) (types.EventJobData, error)
	CompleteEventJob(jobID int64) error
	UpdateEventJobStatus(jobID int64, isActive bool) error
}

type eventJobRepository struct {
	db *database.Connection
}

func NewEventJobRepository(db *database.Connection) EventJobRepository {
	return &eventJobRepository{
		db: db,
	}
}

func (r *eventJobRepository) CreateEventJob(eventJob *types.EventJobData) error {
	err := r.db.Session().Query(queries.CreateEventJobDataQuery,
		eventJob.JobID, eventJob.ExpirationTime, eventJob.Recurring, eventJob.TriggerChainID, eventJob.TriggerContractAddress, eventJob.TriggerEvent,
		eventJob.TargetChainID, eventJob.TargetContractAddress, eventJob.TargetFunction, eventJob.ABI, eventJob.ArgType, eventJob.Arguments,
		eventJob.DynamicArgumentsScriptUrl, false, true).Exec()

	if err != nil {
		return err
	}

	return nil
}

func (r *eventJobRepository) GetEventJobByJobID(jobID int64) (types.EventJobData, error) {
	var eventJob types.EventJobData
	err := r.db.Session().Query(queries.GetEventJobDataByJobIDQuery, jobID).Scan(
		&eventJob.JobID, &eventJob.TimeFrame, &eventJob.Recurring, &eventJob.TriggerChainID,
		&eventJob.TriggerContractAddress, &eventJob.TriggerEvent, &eventJob.TargetChainID,
		&eventJob.TargetContractAddress, &eventJob.TargetFunction, &eventJob.ABI, &eventJob.ArgType,
		&eventJob.Arguments, &eventJob.DynamicArgumentsScriptUrl, &eventJob.IsCompleted, &eventJob.IsActive,
	)
	if err != nil {
		return types.EventJobData{}, errors.New("failed to get event job by job ID")
	}

	return eventJob, nil
}

func (r *eventJobRepository) CompleteEventJob(jobID int64) error {
	err := r.db.Session().Query(queries.CompleteEventJobStatusQuery, jobID).Exec()
	if err != nil {
		return errors.New("failed to complete event job")
	}

	return nil
}

func (r *eventJobRepository) UpdateEventJobStatus(jobID int64, isActive bool) error {
	err := r.db.Session().Query(queries.UpdateEventJobStatusQuery, isActive, jobID).Exec()
	if err != nil {
		return errors.New("failed to update event job status")
	}

	return nil
}
