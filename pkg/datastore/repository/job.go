package repository

import (
	"errors"
	"math/big"
	"time"

	"github.com/trigg3rX/triggerx-backend/pkg/datastore/connection"
	"github.com/trigg3rX/triggerx-backend/pkg/datastore/repository/queries"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

type JobRepository interface {
	CreateNewJob(job *types.JobData) (*big.Int, error)
	UpdateJobFromUserInDB(jobID *big.Int, job *types.UpdateJobDataFromUserRequest) error
	UpdateJobLastExecutedAt(jobID *big.Int, taskID int64, jobCostActual float64, lastExecutedAt time.Time) error
	UpdateJobStatus(jobID *big.Int, status string) error
	GetJobByID(jobID *big.Int) (*types.JobData, error)
	GetTaskDefinitionIDByJobID(jobID *big.Int) (int, error)
	GetTaskFeesByJobID(jobID *big.Int) ([]types.TaskFeeResponse, error)
	GetJobsByUserIDAndChainID(userID int64, createdChainID string) ([]types.JobData, error)
}

type jobRepository struct {
	db connection.ConnectionManager
}

// NewJobRepository creates a new job repository instance
func NewJobRepository(db connection.ConnectionManager) JobRepository {
	return &jobRepository{
		db: db,
	}
}

func (r *jobRepository) CreateNewJob(job *types.JobData) (*big.Int, error) {
	err := r.db.GetSession().Query(queries.CreateJobDataQuery,
		job.JobID, job.JobTitle, job.TaskDefinitionID, job.CreatedChainID, 
		job.UserID, job.LinkJobID, job.ChainStatus, job.TimeFrame, job.IsImua,
		job.JobType, job.TimeFrame, job.Recurring, job.Status, job.JobCostPrediction, time.Now()).Exec()

	if err != nil {
		return nil, err
	}

	return job.JobID, nil
}

func (r *jobRepository) UpdateJobFromUserInDB(jobID *big.Int, job *types.UpdateJobDataFromUserRequest) error {
	err := r.db.GetSession().Query(queries.UpdateJobDataFromUserQuery,
		job.JobTitle, job.TimeFrame, job.Recurring, job.Status, job.JobCostPrediction, time.Now(), jobID).Exec()
	if err != nil {
		return errors.New("failed to update job from user")
	}
	return nil
}

func (r *jobRepository) UpdateJobLastExecutedAt(jobID *big.Int, taskID int64, jobCostActual float64, lastExecutedAt time.Time) error {
	var existingTaskIDs []int64
	err := r.db.GetSession().Query(queries.GetTaskIDsByJobIDQuery, jobID).Scan(&existingTaskIDs)
	if err != nil {
		return errors.New("failed to get task ids by job id")
	}

	existingTaskIDs = append(existingTaskIDs, taskID)
	err = r.db.GetSession().Query(queries.UpdateJobDataLastExecutedAtQuery,
		existingTaskIDs, jobCostActual, lastExecutedAt).Exec()
	if err != nil {
		return errors.New("failed to update job last executed at")
	}
	return nil
}

func (r *jobRepository) UpdateJobStatus(jobID *big.Int, status string) error {
	err := r.db.GetSession().Query(queries.UpdateJobDataStatusQuery,
		status, time.Now(), jobID).Exec()
	if err != nil {
		return errors.New("failed to update job status")
	}
	return nil
}

func (r *jobRepository) GetJobByID(jobID *big.Int) (*types.JobData, error) {
	var jobData types.JobData
	var jobIDBigInt, linkJobIDBigInt *big.Int
	err := r.db.GetSession().Query(queries.GetJobDataByJobIDQuery, jobID).Scan(
		&jobIDBigInt, &jobData.JobTitle, &jobData.TaskDefinitionID, &jobData.UserID,
		&linkJobIDBigInt, &jobData.ChainStatus, &jobData.TimeFrame,
		&jobData.Recurring, &jobData.Status, &jobData.JobCostPrediction, &jobData.JobCostActual,
		&jobData.TaskIDs, &jobData.CreatedAt, &jobData.UpdatedAt, &jobData.LastExecutedAt,
		&jobData.Timezone, &jobData.IsImua, &jobData.CreatedChainID)

	if err != nil {
		return nil, err
	}
	jobData.JobID = jobIDBigInt
	jobData.LinkJobID = linkJobIDBigInt
	return &jobData, nil
}

func (r *jobRepository) GetTaskDefinitionIDByJobID(jobID *big.Int) (int, error) {
	var taskDefinitionID int
	err := r.db.GetSession().Query(queries.GetTaskDefinitionIDByJobIDQuery, jobID).Scan(&taskDefinitionID)
	if err != nil {
		return 0, errors.New("failed to get task definition id by job id")
	}
	return taskDefinitionID, nil
}

func (r *jobRepository) GetTaskFeesByJobID(jobID *big.Int) ([]types.TaskFeeResponse, error) {
	session := r.db.GetSession()
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

func (r *jobRepository) GetJobsByUserIDAndChainID(userID int64, createdChainID string) ([]types.JobData, error) {
	session := r.db.GetSession()
	iter := session.Query(queries.GetJobsByUserIDAndChainIDQuery, userID, createdChainID).Iter()

	var jobs []types.JobData
	for {
		var (
			jobIDBigInt     *big.Int
			linkJobIDBigInt *big.Int
			job             types.JobData
		)
		if !iter.Scan(
			&jobIDBigInt, &job.JobTitle, &job.TaskDefinitionID, &job.UserID,
			&linkJobIDBigInt, &job.ChainStatus, &job.TimeFrame,
			&job.Recurring, &job.Status, &job.JobCostPrediction, &job.JobCostActual,
			&job.TaskIDs, &job.CreatedAt, &job.UpdatedAt, &job.LastExecutedAt,
			&job.Timezone, &job.IsImua, &job.CreatedChainID,
		) {
			break
		}
		job.JobID = jobIDBigInt
		job.LinkJobID = linkJobIDBigInt
		jobs = append(jobs, job)
	}

	if err := iter.Close(); err != nil {
		return nil, err
	}
	return jobs, nil
}
