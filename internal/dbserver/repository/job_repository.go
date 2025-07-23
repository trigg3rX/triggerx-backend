package repository

import (
	"errors"
	"math/big"
	"time"

	"github.com/trigg3rX/triggerx-backend/internal/dbserver/repository/queries"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/types"
	"github.com/trigg3rX/triggerx-backend/pkg/database"
)

type JobRepository interface {
	CreateNewJob(job *types.JobData) (*big.Int, error)
	UpdateJobFromUserInDB(job *types.UpdateJobDataFromUserRequest) error
	UpdateJobLastExecutedAt(jobID *big.Int, taskID int64, jobCostActual float64, lastExecutedAt time.Time) error
	UpdateJobStatus(jobID *big.Int, status string) error
	GetJobByID(jobID *big.Int) (*types.JobData, error)
	GetTaskDefinitionIDByJobID(jobID *big.Int) (int, error)
	GetTaskFeesByJobID(jobID *big.Int) ([]types.TaskFeeResponse, error)
}

type jobRepository struct {
	db *database.Connection
}

// NewJobRepository creates a new job repository instance
func NewJobRepository(db *database.Connection) JobRepository {
	return &jobRepository{
		db: db,
	}
}

func (r *jobRepository) CreateNewJob(job *types.JobData) (*big.Int, error) {
	// var lastJobID int64
	// err := r.db.Session().Query(queries.GetMaxJobIDQuery).Scan(&lastJobID)
	// if err == gocql.ErrNotFound {
	// 	return -1, nil
	// }

	err := r.db.Session().Query(queries.CreateJobDataQuery,
		job.JobID, job.JobTitle, job.TaskDefinitionID, job.UserID, job.LinkJobID, job.ChainStatus,
		job.Custom, job.TimeFrame, job.Recurring, job.Status, job.JobCostPrediction, time.Now(), time.Now(), job.Timezone, job.IsImua, job.CreatedChainID).Exec()

	if err != nil {
		return nil, err
	}

	return job.JobID, nil
}

func (r *jobRepository) UpdateJobFromUserInDB(job *types.UpdateJobDataFromUserRequest) error {
	err := r.db.Session().Query(queries.UpdateJobDataFromUserQuery,
		job.JobTitle, job.TimeFrame, job.Recurring, job.Status, job.JobCostPrediction, time.Now(), job.JobID).Exec()
	if err != nil {
		return errors.New("failed to update job from user")
	}
	return nil
}

func (r *jobRepository) UpdateJobLastExecutedAt(jobID *big.Int, taskID int64, jobCostActual float64, lastExecutedAt time.Time) error {
	var existingTaskIDs []int64
	err := r.db.Session().Query(queries.GetTaskIDsByJobIDQuery, jobID).Scan(&existingTaskIDs)
	if err != nil {
		return errors.New("failed to get task ids by job id")
	}

	existingTaskIDs = append(existingTaskIDs, taskID)
	err = r.db.Session().Query(queries.UpdateJobDataLastExecutedAtQuery,
		existingTaskIDs, jobCostActual, lastExecutedAt).Exec()
	if err != nil {
		return errors.New("failed to update job last executed at")
	}
	return nil
}

func (r *jobRepository) UpdateJobStatus(jobID *big.Int, status string) error {
	err := r.db.Session().Query(queries.UpdateJobDataStatusQuery,
		status, time.Now(), jobID).Exec()
	if err != nil {
		return errors.New("failed to update job status")
	}
	return nil
}

func (r *jobRepository) GetJobByID(jobID *big.Int) (*types.JobData, error) {
	var jobData types.JobData
	err := r.db.Session().Query(queries.GetJobDataByJobIDQuery, jobID).Scan(
		&jobData.JobID, &jobData.JobTitle, &jobData.TaskDefinitionID, &jobData.UserID,
		&jobData.LinkJobID, &jobData.ChainStatus, &jobData.Custom, &jobData.TimeFrame,
		&jobData.Recurring, &jobData.Status, &jobData.JobCostPrediction, &jobData.JobCostActual,
		&jobData.TaskIDs, &jobData.CreatedAt, &jobData.UpdatedAt, &jobData.LastExecutedAt,
		&jobData.Timezone, &jobData.IsImua, &jobData.CreatedChainID)

	if err != nil {
		return nil, err
	}
	return &jobData, nil
}

func (r *jobRepository) GetTaskDefinitionIDByJobID(jobID *big.Int) (int, error) {
	var taskDefinitionID int
	err := r.db.Session().Query(queries.GetTaskDefinitionIDByJobIDQuery, jobID).Scan(&taskDefinitionID)
	if err != nil {
		return 0, errors.New("failed to get task definition id by job id")
	}
	return taskDefinitionID, nil
}

func (r *jobRepository) GetTaskFeesByJobID(jobID *big.Int) ([]types.TaskFeeResponse, error) {
	session := r.db.Session()
	iter := session.Query(queries.GetTaskFeesByJobIDQuery, jobID).Iter()

	var results []types.TaskFeeResponse
	var taskID int64
	var taskOpxCost float64
	for iter.Scan(&taskID, &taskOpxCost) {
		results = append(results, types.TaskFeeResponse{
			TaskID:      taskID,
			TaskOpxCost: taskOpxCost,
		})
	}
	if err := iter.Close(); err != nil {
		return nil, err
	}
	return results, nil
}
