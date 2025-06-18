package repository

import (
	"errors"
	"time"

	"github.com/gocql/gocql"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/repository/queries"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/types"
	"github.com/trigg3rX/triggerx-backend/pkg/database"
)

type JobRepository interface {
	CreateNewJob(job *types.JobData) (int64, error)
	UpdateJobFromUserInDB(job *types.UpdateJobDataFromUserRequest) error
	UpdateJobLastExecutedAt(jobID int64, taskID int64, jobCostActual float64, lastExecutedAt time.Time) error
	UpdateJobStatus(jobID int64, status string) error
	GetJobByID(jobID int64) (*types.JobData, error)
	GetTaskDefinitionIDByJobID(jobID int64) (int, error)
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

func (r *jobRepository) CreateNewJob(job *types.JobData) (int64, error) {
	var lastJobID int64
	err := r.db.Session().Query(queries.GetMaxJobIDQuery).Scan(&lastJobID)
	if err == gocql.ErrNotFound {
		return -1, nil
	}

	err = r.db.Session().Query(queries.CreateJobDataQuery,
		lastJobID+1, job.JobTitle, job.TaskDefinitionID, job.UserID, job.LinkJobID, job.ChainStatus,
		job.Custom, job.TimeFrame, job.Recurring, job.Status, job.JobCostPrediction, time.Now(), time.Now(), job.Timezone).Exec()

	if err != nil {
		return -1, err
	}

	return lastJobID + 1, nil
}

func (r *jobRepository) UpdateJobFromUserInDB(job *types.UpdateJobDataFromUserRequest) error {
	err := r.db.Session().Query(queries.UpdateJobDataFromUserQuery,
		job.JobTitle, job.TimeFrame, job.Recurring, job.Status, job.JobCostPrediction, time.Now()).Exec()
	if err != nil {
		return errors.New("failed to update job from user")
	}
	return nil
}

func (r *jobRepository) UpdateJobLastExecutedAt(jobID int64, taskID int64, jobCostActual float64, lastExecutedAt time.Time) error {
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

func (r *jobRepository) UpdateJobStatus(jobID int64, status string) error {
	err := r.db.Session().Query(queries.UpdateJobDataStatusQuery,
		status, time.Now()).Exec()
	if err != nil {
		return errors.New("failed to update job status")
	}
	return nil
}

func (r *jobRepository) GetJobByID(jobID int64) (*types.JobData, error) {
	var jobData types.JobData
	err := r.db.Session().Query(queries.GetJobDataByJobIDQuery, jobID).Scan(
		&jobData.JobID, &jobData.JobTitle, &jobData.TaskDefinitionID, &jobData.UserID,
		&jobData.LinkJobID, &jobData.ChainStatus, &jobData.Custom, &jobData.TimeFrame,
		&jobData.Recurring, &jobData.Status, &jobData.JobCostPrediction, &jobData.JobCostActual,
		&jobData.TaskIDs, &jobData.CreatedAt, &jobData.UpdatedAt, &jobData.LastExecutedAt,
		&jobData.Timezone)

	if err != nil {
		return nil, err
	}
	return &jobData, nil
}

func (r *jobRepository) GetTaskDefinitionIDByJobID(jobID int64) (int, error) {
	var taskDefinitionID int
	err := r.db.Session().Query(queries.GetTaskDefinitionIDByJobIDQuery, jobID).Scan(&taskDefinitionID)
	if err != nil {
		return 0, errors.New("failed to get task definition id by job id")
	}
	return taskDefinitionID, nil
}
