package repository

import (
	"errors"
	"fmt"
	"time"

	"github.com/trigg3rX/triggerx-backend/pkg/database"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// EventJobRepository defines the interface for event-based job operations
type JobRepository interface {
	GetActiveEventJobs() ([]ActiveJob, error)
	UpdateEventJobStatus(jobID string, isActive bool) error
	GetEventJobByJobID(jobID string) (*types.EventJobDataEntity, error)

	GetActiveConditionJobs() ([]ActiveJob, error)
	UpdateConditionJobStatus(jobID string, isActive bool) error
	GetConditionJobByJobID(jobID string) (*types.ConditionJobDataEntity, error)
}

// ActiveJob represents an event job with minimal fields needed for expiration checking
type ActiveJob struct {
	JobID          string
	ExpirationTime time.Time
}

type jobRepository struct {
	db *database.Connection
}

// NewEventJobRepository creates a new event job repository instance
func NewJobRepository(db *database.Connection) JobRepository {
	return &jobRepository{
		db: db,
	}
}

// GetActiveEventJobs retrieves all active event jobs
func (r *jobRepository) GetActiveEventJobs() ([]ActiveJob, error) {
	iter := r.db.Session().Query(getActiveEventJobsQuery).Iter()

	var eventJobs []ActiveJob
	var job ActiveJob

	for iter.Scan(&job.JobID, &job.ExpirationTime) {
		eventJobs = append(eventJobs, job)
	}

	if err := iter.Close(); err != nil {
		return nil, err
	}

	return eventJobs, nil
}

// UpdateEventJobStatus updates the active status of an event job
func (r *jobRepository) UpdateEventJobStatus(jobID string, isActive bool) error {
	err := r.db.Session().Query(updateEventJobStatusQuery, isActive, jobID).Exec()
	if err != nil {
		return errors.New("failed to update event job status")
	}

	return nil
}

// GetEventJobByJobID retrieves an event job by job ID
func (r *jobRepository) GetEventJobByJobID(jobID string) (*types.EventJobDataEntity, error) {
	var entity types.EventJobDataEntity
	err := r.db.Session().Query(getEventJobByJobIDQuery, jobID).Scan(
		&entity.JobID, &entity.TaskDefinitionID, &entity.Network, &entity.Recurring,
		&entity.TriggerChainID, &entity.TriggerContractAddress, &entity.TriggerEvent,
		&entity.EventFilterParaName, &entity.EventFilterValue,
		&entity.TargetChainID, &entity.TargetContractAddress, &entity.TargetFunction,
		&entity.ABI, &entity.ArgType, &entity.Arguments, &entity.ExecutionScriptURL, &entity.ExecutionScriptLanguage, &entity.ExecutionScriptHash,
		&entity.MaxExecutionTime, &entity.ChallengePeriod,
		&entity.IsActive, &entity.LastExecutedAt, &entity.ExpirationTime)
	if err != nil {
		return nil, errors.New("failed to get event job by job ID")
	}

	return &entity, nil
}

// GetActiveConditionJobs retrieves all active condition jobs
func (r *jobRepository) GetActiveConditionJobs() ([]ActiveJob, error) {
	iter := r.db.Session().Query(getActiveConditionJobsQuery).Iter()

	var conditionJobs []ActiveJob
	var job ActiveJob

	for iter.Scan(&job.JobID, &job.ExpirationTime) {
		conditionJobs = append(conditionJobs, job)
	}

	if err := iter.Close(); err != nil {
		return nil, err
	}

	return conditionJobs, nil
}

// UpdateConditionJobStatus updates the active status of a condition job
func (r *jobRepository) UpdateConditionJobStatus(jobID string, isActive bool) error {
	err := r.db.Session().Query(updateConditionJobStatusQuery, isActive, jobID).Exec()
	if err != nil {
		return errors.New("failed to update condition job status")
	}

	return nil
}

// GetConditionJobByJobID retrieves a condition job by job ID
func (r *jobRepository) GetConditionJobByJobID(jobID string) (*types.ConditionJobDataEntity, error) {
	var entity types.ConditionJobDataEntity
	err := r.db.Session().Query(getConditionJobByJobIDQuery, jobID).Scan(
		&entity.JobID, &entity.TaskDefinitionID, &entity.Network, &entity.Recurring, &entity.ConditionType,
		&entity.UpperLimit, &entity.LowerLimit, &entity.ValueSourceType,
		&entity.ValueSourceURL, &entity.SelectedKeyRoute, &entity.TargetChainID,
		&entity.TargetContractAddress, &entity.TargetFunction, &entity.ABI, &entity.ArgType,
		&entity.Arguments, &entity.ExecutionScriptURL, &entity.ExecutionScriptLanguage, &entity.ExecutionScriptHash,
		&entity.MaxExecutionTime, &entity.ChallengePeriod,
		&entity.IsActive, &entity.LastExecutedAt, &entity.ExpirationTime)
	if err != nil {
		return nil, fmt.Errorf("failed to get condition job by job ID: %w", err)
	}

	return &entity, nil
}
