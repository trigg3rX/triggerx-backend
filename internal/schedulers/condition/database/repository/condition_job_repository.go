package repository

import (
	"errors"
	"time"

	"github.com/trigg3rX/triggerx-backend/pkg/database"
)

// ConditionJobRepository defines the interface for condition-based job operations
type ConditionJobRepository interface {
	GetActiveConditionJobs() ([]ActiveConditionJob, error)
	UpdateConditionJobStatus(jobID string, isActive bool) error
}

// ActiveConditionJob represents a condition job with minimal fields needed for expiration checking
type ActiveConditionJob struct {
	JobID          string
	ExpirationTime time.Time
}

type conditionJobRepository struct {
	db *database.Connection
}

// NewConditionJobRepository creates a new condition job repository instance
func NewConditionJobRepository(db *database.Connection) ConditionJobRepository {
	return &conditionJobRepository{
		db: db,
	}
}

// GetActiveConditionJobs retrieves all active condition jobs
func (r *conditionJobRepository) GetActiveConditionJobs() ([]ActiveConditionJob, error) {
	iter := r.db.Session().Query(getActiveConditionJobsQuery).Iter()

	var conditionJobs []ActiveConditionJob
	var job ActiveConditionJob

	for iter.Scan(&job.JobID, &job.ExpirationTime) {
		conditionJobs = append(conditionJobs, job)
	}

	if err := iter.Close(); err != nil {
		return nil, err
	}

	return conditionJobs, nil
}

// UpdateConditionJobStatus updates the active status of a condition job
func (r *conditionJobRepository) UpdateConditionJobStatus(jobID string, isActive bool) error {
	err := r.db.Session().Query(updateConditionJobStatusQuery, isActive, jobID).Exec()
	if err != nil {
		return errors.New("failed to update condition job status")
	}

	return nil
}

// Query constants
const (
	getActiveConditionJobsQuery = `
        SELECT job_id, expiration_time
        FROM triggerx.condition_job_data
        WHERE is_active = true
        ALLOW FILTERING`

	updateConditionJobStatusQuery = `
        UPDATE triggerx.condition_job_data
        SET is_active = ?
        WHERE job_id = ?`
)
