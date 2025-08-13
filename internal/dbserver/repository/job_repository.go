package repository

import (
	"errors"
	"math/big"
	"time"

	"github.com/trigg3rX/triggerx-backend/internal/dbserver/repository/queries"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/types"
	"github.com/trigg3rX/triggerx-backend/pkg/database"
	commonTypes "github.com/trigg3rX/triggerx-backend/pkg/types"
)

type JobRepository interface {
	CreateNewJob(job *commonTypes.JobData) (*big.Int, error)
	UpdateJobFromUserInDB(jobID *big.Int, job *types.UpdateJobDataFromUserRequest) error
	UpdateJobLastExecutedAt(jobID *big.Int, taskID int64, jobCostActual float64, lastExecutedAt time.Time) error
	UpdateJobStatus(jobID *big.Int, status string) error
	GetJobByID(jobID *big.Int) (*commonTypes.JobData, error)
	GetTaskDefinitionIDByJobID(jobID *big.Int) (int, error)
	GetTaskFeesByJobID(jobID *big.Int) ([]types.TaskFeeResponse, error)
	GetJobsByUserIDAndChainID(userID int64, createdChainID string) ([]commonTypes.JobData, error)
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

func (r *jobRepository) CreateNewJob(job *commonTypes.JobData) (*big.Int, error) {
	// var lastJobID int64
	// err := r.db.Session().Query(queries.GetMaxJobIDQuery).Scan(&lastJobID)
	// if err == gocql.ErrNotFound {
	// 	return -1, nil
	// }

	err := r.db.Session().Query(queries.CreateJobDataQuery,
		job.JobID.ToBigInt(), job.JobTitle, job.TaskDefinitionID, job.UserID, job.LinkJobID.ToBigInt(), job.ChainStatus,
		job.Custom, job.TimeFrame, job.Recurring, job.Status, job.JobCostPrediction, time.Now(), time.Now(), job.Timezone, job.IsImua, job.CreatedChainID).Exec()

	if err != nil {
		return nil, err
	}

	return job.JobID.Int, nil
}

func (r *jobRepository) UpdateJobFromUserInDB(jobID *big.Int, job *types.UpdateJobDataFromUserRequest) error {
	err := r.db.Session().Query(queries.UpdateJobDataFromUserQuery,
		job.JobTitle, job.TimeFrame, job.Recurring, job.Status, job.JobCostPrediction, time.Now(), jobID).Exec()
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

func (r *jobRepository) GetJobByID(jobID *big.Int) (*commonTypes.JobData, error) {
	var jobData commonTypes.JobData
	var jobIDBigInt, linkJobIDBigInt *big.Int
	err := r.db.Session().Query(queries.GetJobDataByJobIDQuery, jobID).Scan(
		&jobIDBigInt, &jobData.JobTitle, &jobData.TaskDefinitionID, &jobData.UserID,
		&linkJobIDBigInt, &jobData.ChainStatus, &jobData.Custom, &jobData.TimeFrame,
		&jobData.Recurring, &jobData.Status, &jobData.JobCostPrediction, &jobData.JobCostActual,
		&jobData.TaskIDs, &jobData.CreatedAt, &jobData.UpdatedAt, &jobData.LastExecutedAt,
		&jobData.Timezone, &jobData.IsImua, &jobData.CreatedChainID)

	if err != nil {
		return nil, err
	}
	jobData.JobID = commonTypes.NewBigInt(jobIDBigInt)
	jobData.LinkJobID = commonTypes.NewBigInt(linkJobIDBigInt)
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

func (r *jobRepository) GetJobsByUserIDAndChainID(userID int64, createdChainID string) ([]commonTypes.JobData, error) {
	session := r.db.Session()
	iter := session.Query(queries.GetJobsByUserIDAndChainIDQuery, userID, createdChainID).Iter()

	var jobs []commonTypes.JobData
	for {
		var (
			jobIDBigInt     *big.Int
			linkJobIDBigInt *big.Int
			job             commonTypes.JobData
		)
		if !iter.Scan(
			&jobIDBigInt, &job.JobTitle, &job.TaskDefinitionID, &job.UserID,
			&linkJobIDBigInt, &job.ChainStatus, &job.Custom, &job.TimeFrame,
			&job.Recurring, &job.Status, &job.JobCostPrediction, &job.JobCostActual,
			&job.TaskIDs, &job.CreatedAt, &job.UpdatedAt, &job.LastExecutedAt,
			&job.Timezone, &job.IsImua, &job.CreatedChainID,
		) {
			break
		}
		job.JobID = commonTypes.NewBigInt(jobIDBigInt)
		job.LinkJobID = commonTypes.NewBigInt(linkJobIDBigInt)
		jobs = append(jobs, job)
	}

	if err := iter.Close(); err != nil {
		return nil, err
	}
	return jobs, nil
}
