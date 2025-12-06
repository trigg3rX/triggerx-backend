package repository

import (
	"math/big"
	"time"

	"github.com/trigg3rX/triggerx-backend/internal/dbserver/repository/queries"
	"github.com/trigg3rX/triggerx-backend/pkg/database"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// CustomJobRepository handles custom job CRUD operations (TaskDefinitionID = 7)
type CustomJobRepository interface {
	CreateCustomJob(job *types.CustomJobData) error
	GetCustomJobByID(jobID *big.Int) (*types.CustomJobData, error)
	GetCustomJobsDueForExecution(currentTime time.Time) ([]types.CustomJobData, error)
	GetActiveCustomJobs() ([]types.CustomJobData, error)
	UpdateNextExecutionTime(jobID *big.Int, nextTime time.Time, lastExecutedAt time.Time) error
	UpdateCustomJobStatus(jobID *big.Int, isActive, isCompleted bool) error
	UpdateCustomJob(jobID *big.Int, scriptURL string, timeInterval int64, recurring bool) error
	DeleteCustomJob(jobID *big.Int) error
}

type customJobRepository struct {
	db *database.Connection
}

// NewCustomJobRepository creates a new custom job repository
func NewCustomJobRepository(db *database.Connection) CustomJobRepository {
	return &customJobRepository{
		db: db,
	}
}

func (r *customJobRepository) CreateCustomJob(job *types.CustomJobData) error {
	return r.db.Session().Query(queries.CreateCustomJobQuery,
		job.JobID.ToBigInt(),
		job.TaskDefinitionID,
		job.Recurring,
		job.CustomScriptUrl,
		job.TimeInterval,
		job.TargetChainID,
		job.IsCompleted,
		job.IsActive,
		job.CreatedAt,
		job.UpdatedAt,
		job.LastExecutedAt,
		job.ExpirationTime,
		job.ScriptLanguage,
		job.ScriptHash,
		job.NextExecutionTime,
		job.MaxExecutionTime,
		job.ChallengePeriod,
	).Exec()
}

func (r *customJobRepository) GetCustomJobByID(jobID *big.Int) (*types.CustomJobData, error) {
	var job types.CustomJobData
	var rawJobID *big.Int

	err := r.db.Session().Query(queries.GetCustomJobByIDQuery, jobID).Scan(
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
	)

	if err != nil {
		return nil, err
	}

	job.JobID = types.FromBigInt(rawJobID)

	return &job, nil
}

func (r *customJobRepository) GetCustomJobsDueForExecution(currentTime time.Time) ([]types.CustomJobData, error) {
	iter := r.db.Session().Query(queries.GetCustomJobsDueForExecutionQuery, true).Iter()

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

func (r *customJobRepository) GetActiveCustomJobs() ([]types.CustomJobData, error) {
	iter := r.db.Session().Query(queries.GetActiveCustomJobsQuery, true).Iter()

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
		jobs = append(jobs, job)
	}

	if err := iter.Close(); err != nil {
		return nil, err
	}

	return jobs, nil
}

func (r *customJobRepository) UpdateNextExecutionTime(jobID *big.Int, nextTime time.Time, lastExecutedAt time.Time) error {
	return r.db.Session().Query(queries.UpdateCustomJobNextExecutionQuery,
		nextTime,
		lastExecutedAt,
		time.Now(),
		jobID,
	).Exec()
}

func (r *customJobRepository) UpdateCustomJobStatus(jobID *big.Int, isActive, isCompleted bool) error {
	return r.db.Session().Query(queries.UpdateCustomJobStatusQuery,
		isActive,
		isCompleted,
		time.Now(),
		jobID,
	).Exec()
}

func (r *customJobRepository) UpdateCustomJob(jobID *big.Int, scriptURL string, timeInterval int64, recurring bool) error {
	return r.db.Session().Query(queries.UpdateCustomJobQuery,
		scriptURL,
		timeInterval,
		recurring,
		time.Now(),
		jobID,
	).Exec()
}

func (r *customJobRepository) DeleteCustomJob(jobID *big.Int) error {
	return r.db.Session().Query(queries.DeleteCustomJobQuery, jobID).Exec()
}
