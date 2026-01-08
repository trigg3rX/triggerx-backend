package repository

import (
	"errors"
	"math/big"
	"time"

	"github.com/trigg3rX/triggerx-backend/pkg/database"
)

// EventJobRepository defines the interface for event-based job operations
type EventJobRepository interface {
	GetActiveEventJobs() ([]ActiveEventJob, error)
	UpdateEventJobStatus(jobID *big.Int, isActive bool) error
}

// ActiveEventJob represents an event job with minimal fields needed for expiration checking
type ActiveEventJob struct {
	JobID          *big.Int
	ExpirationTime time.Time
}

type eventJobRepository struct {
	db *database.Connection
}

// NewEventJobRepository creates a new event job repository instance
func NewEventJobRepository(db *database.Connection) EventJobRepository {
	return &eventJobRepository{
		db: db,
	}
}

// GetActiveEventJobs retrieves all active event jobs
func (r *eventJobRepository) GetActiveEventJobs() ([]ActiveEventJob, error) {
	iter := r.db.Session().Query(getActiveEventJobsQuery).Iter()

	var eventJobs []ActiveEventJob
	var job ActiveEventJob
	var jobIDBigInt *big.Int

	for iter.Scan(&jobIDBigInt, &job.ExpirationTime) {
		job.JobID = jobIDBigInt
		eventJobs = append(eventJobs, job)
	}

	if err := iter.Close(); err != nil {
		return nil, err
	}

	return eventJobs, nil
}

// UpdateEventJobStatus updates the active status of an event job
func (r *eventJobRepository) UpdateEventJobStatus(jobID *big.Int, isActive bool) error {
	err := r.db.Session().Query(updateEventJobStatusQuery, isActive, jobID).Exec()
	if err != nil {
		return errors.New("failed to update event job status")
	}

	return nil
}

// Query constants
const (
	getActiveEventJobsQuery = `
        SELECT job_id, expiration_time
        FROM triggerx.event_job_data
        WHERE is_active = true
        ALLOW FILTERING`

	updateEventJobStatusQuery = `
        UPDATE triggerx.event_job_data
        SET is_active = ?
        WHERE job_id = ?`
)
