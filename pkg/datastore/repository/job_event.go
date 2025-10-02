package repository

import (
	"errors"
	"math/big"
	"time"

	"github.com/trigg3rX/triggerx-backend/pkg/datastore/connection"
	"github.com/trigg3rX/triggerx-backend/pkg/datastore/repository/queries"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

type eventJobRepository struct {
	db connection.ConnectionManager
}

func NewEventJobRepository(db connection.ConnectionManager) EventJobRepository {
	return &eventJobRepository{
		db: db,
	}
}

func (r *eventJobRepository) CreateEventJob(eventJob *types.EventJobData) error {
	err := r.db.GetSession().Query(queries.CreateEventJobDataQuery,
		eventJob.JobID, eventJob.TaskDefinitionID, eventJob.Recurring,
		eventJob.TriggerChainID, eventJob.TriggerContractAddress, eventJob.TriggerEvent,
		eventJob.TargetChainID, eventJob.TargetContractAddress, eventJob.TargetFunction,
		eventJob.ABI, eventJob.ArgType, eventJob.Arguments, eventJob.DynamicArgumentsScriptUrl,
		eventJob.IsCompleted, eventJob.ExpirationTime, time.Now()).Exec()

	if err != nil {
		return err
	}

	return nil
}

func (r *eventJobRepository) GetEventJobByJobID(jobID *big.Int) (types.EventJobData, error) {
	var eventJob types.EventJobData
	var temp *big.Int
	eventJob.JobID = jobID
	err := r.db.GetSession().Query(queries.GetEventJobDataByJobIDQuery, jobID).Scan(
		&temp, &eventJob.ExpirationTime, &eventJob.Recurring, &eventJob.TriggerChainID,
		&eventJob.TriggerContractAddress, &eventJob.TriggerEvent, &eventJob.TargetChainID,
		&eventJob.TargetContractAddress, &eventJob.TargetFunction, &eventJob.ABI, &eventJob.ArgType,
		&eventJob.Arguments, &eventJob.DynamicArgumentsScriptUrl, &eventJob.IsCompleted)
	if err != nil {
		return types.EventJobData{}, errors.New("failed to get event job by job ID")
	}

	return eventJob, nil
}

func (r *eventJobRepository) CompleteEventJob(jobID *big.Int) error {
	err := r.db.GetSession().Query(queries.CompleteEventJobStatusQuery, jobID).Exec()
	if err != nil {
		return errors.New("failed to complete event job")
	}

	err = r.db.GetSession().Query(queries.UpdateJobDataToCompletedQuery, jobID).Exec()
	if err != nil {
		return errors.New("failed to update job_data status to completed")
	}

	return nil
}

func (r *eventJobRepository) UpdateEventJobStatus(jobID *big.Int, isActive bool) error {
	err := r.db.GetSession().Query(queries.UpdateEventJobStatusQuery, isActive, jobID).Exec()
	if err != nil {
		return errors.New("failed to update event job status")
	}

	return nil
}

func (r *eventJobRepository) GetActiveEventJobs() ([]types.EventJobData, error) {
	var eventJobs []types.EventJobData
	iter := r.db.GetSession().Query(queries.GetActiveEventJobsQuery).Iter()
	var eventJob types.EventJobData
	for iter.Scan(
		&eventJob.JobID, &eventJob.ExpirationTime, &eventJob.Recurring,
		&eventJob.TriggerChainID, &eventJob.TriggerContractAddress, &eventJob.TriggerEvent,
		&eventJob.TargetChainID, &eventJob.TargetContractAddress, &eventJob.TargetFunction,
		&eventJob.ABI, &eventJob.ArgType, &eventJob.Arguments, &eventJob.DynamicArgumentsScriptUrl,
		&eventJob.IsCompleted) {
		eventJobs = append(eventJobs, eventJob)
	}
	if err := iter.Close(); err != nil {
		return nil, errors.New("failed to fetch active event jobs")
	}
	return eventJobs, nil
}
