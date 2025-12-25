package repository

import (
	"math/big"
	"time"

	"github.com/trigg3rX/triggerx-backend/pkg/database"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// CustomJobRepository defines the interface for custom job operations.
type CustomJobRepository interface {
	// GetCustomJobsDueForExecution retrieves custom jobs that are due for execution.
	GetCustomJobsDueForExecution(currentTime time.Time) ([]types.CustomJobData, error)
}

type customJobRepository struct {
	db *database.Connection
}

// NewCustomJobRepository creates a new custom job repository instance.
func NewCustomJobRepository(db *database.Connection) CustomJobRepository {
	return &customJobRepository{
		db: db,
	}
}

// GetCustomJobsDueForExecution retrieves custom jobs that are due for execution.
func (r *customJobRepository) GetCustomJobsDueForExecution(currentTime time.Time) ([]types.CustomJobData, error) {
	iter := r.db.Session().Query(getCustomJobsDueForExecutionQuery, true).Iter()

	var jobs []types.CustomJobData

	for {
		var rawJobID *big.Int
		var job types.CustomJobData

		if !iter.Scan(
			&rawJobID,
			&job.TaskDefinitionID,
			&job.Recurring,
			&job.CustomScriptUrl,
			&job.TimeInterval,
			&job.TargetChainID,
			&job.IsCompleted,
			&job.IsActive,
			&job.CreatedAt,
			&job.UpdatedAt,
			&job.LastExecutedAt,
			&job.ExpirationTime,
			&job.ScriptLanguage,
			&job.ScriptHash,
			&job.NextExecutionTime,
			&job.MaxExecutionTime,
			&job.ChallengePeriod,
		) {
			break
		}

		job.JobID = types.FromBigInt(rawJobID)

		// Filter jobs that are due (next_execution_time <= current_time)
		if job.IsActive && !job.IsCompleted && job.NextExecutionTime.Before(currentTime) {
			jobs = append(jobs, job)
		}
	}

	if err := iter.Close(); err != nil {
		return nil, err
	}

	return jobs, nil
}

// Query constants
const (
	getCustomJobsDueForExecutionQuery = `
		SELECT job_id, task_definition_id, recurring, custom_script_url, time_interval,
			target_chain_id, is_completed, is_active, created_at, updated_at, last_executed_at,
			expiration_time, script_language, script_hash, next_execution_time,
			max_execution_time, challenge_period
		FROM triggerx.custom_jobs
		WHERE is_active = ? ALLOW FILTERING`
)
