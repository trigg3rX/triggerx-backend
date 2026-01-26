package repository

import (
	"errors"
	"strings"
	"time"

	"github.com/trigg3rX/triggerx-backend/pkg/database"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

type JobRepository interface {
	CreateNewJob(job *types.JobDataEntity) error
	UpdateJobFromUser(jobID string, job *types.UpdateJobDataFromUserRequest) error
	UpdateJobStatus(jobID string, status types.JobStatus) error
	GetJobByID(jobID string) (*types.JobDataDTO, error)
	GetTaskDefinitionIDByJobID(jobID string) (int, error)
	GetJobsByUserAddressAndChainID(userAddress string, createdChainID string) ([]types.JobDataDTO, error)
	GetJobsBySafeAddress(safeAddress string) ([]types.JobDataDTO, error)
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

func (r *jobRepository) CreateNewJob(job *types.JobDataEntity) error {
	err := r.db.Session().Query(CreateJobDataQuery,
		job.JobID, job.JobTitle, job.TaskDefinitionID, job.CreatedChainID, strings.ToLower(job.UserAddress),
		job.LinkJobID, job.ChainStatus, strings.ToLower(job.SafeAddress), job.Timezone,
		job.JobType, job.TimeFrame, job.Recurring, job.Status, job.JobCostPrediction,
		job.CreatedAt, job.UpdatedAt).Exec()

	return err
}

func (r *jobRepository) UpdateJobFromUser(jobID string, job *types.UpdateJobDataFromUserRequest) error {
	err := r.db.Session().Query(UpdateJobDataFromUserQuery,
		job.JobTitle, job.TimeFrame, job.Recurring, job.Status, job.JobCostPrediction, time.Now(), jobID).Exec()
	if err != nil {
		return errors.New("failed to update job from user")
	}
	return nil
}

func (r *jobRepository) UpdateJobStatus(jobID string, status types.JobStatus) error {
	err := r.db.Session().Query(UpdateJobDataStatusQuery,
		string(status), time.Now(), jobID).Exec()
	if err != nil {
		return errors.New("failed to update job status")
	}
	return nil
}

func (r *jobRepository) GetJobByID(jobID string) (*types.JobDataDTO, error) {
	var entity types.JobDataEntity
	err := r.db.Session().Query(GetJobDataByJobIDQuery, jobID).Scan(
		&entity.JobID, &entity.JobTitle, &entity.TaskDefinitionID, &entity.CreatedChainID,
		&entity.UserAddress, &entity.LinkJobID, &entity.ChainStatus, &entity.SafeAddress,
		&entity.Timezone, &entity.JobType, &entity.TimeFrame,
		&entity.Recurring, &entity.Status, &entity.JobCostPrediction, &entity.JobCostActual,
		&entity.TaskIDs, &entity.CreatedAt, &entity.UpdatedAt, &entity.LastExecutedAt)

	if err != nil {
		return nil, err
	}

	dto := types.JobDataEntityToDTO(&entity)
	return dto, nil
}

func (r *jobRepository) GetTaskDefinitionIDByJobID(jobID string) (int, error) {
	var taskDefinitionID int
	err := r.db.Session().Query(GetTaskDefinitionIDByJobIDQuery, jobID).Scan(&taskDefinitionID)
	if err != nil {
		return 0, errors.New("failed to get task definition id by job id")
	}
	return taskDefinitionID, nil
}

func (r *jobRepository) GetJobsByUserAddressAndChainID(userAddress string, createdChainID string) ([]types.JobDataDTO, error) {
	session := r.db.Session()
	iter := session.Query(GetJobsByUserAddressAndChainIDQuery, strings.ToLower(userAddress), createdChainID).Iter()

	var jobs []types.JobDataDTO
	for {
		var entity types.JobDataEntity
		if !iter.Scan(
			&entity.JobID, &entity.JobTitle, &entity.TaskDefinitionID, &entity.CreatedChainID,
			&entity.UserAddress, &entity.LinkJobID, &entity.ChainStatus, &entity.SafeAddress,
			&entity.Timezone, &entity.JobType, &entity.TimeFrame,
			&entity.Recurring, &entity.Status, &entity.JobCostPrediction, &entity.JobCostActual,
			&entity.TaskIDs, &entity.CreatedAt, &entity.UpdatedAt, &entity.LastExecutedAt,
		) {
			break
		}
		dto := types.JobDataEntityToDTO(&entity)
		jobs = append(jobs, *dto)
	}

	if err := iter.Close(); err != nil {
		return nil, err
	}
	return jobs, nil
}

func (r *jobRepository) GetJobsBySafeAddress(safeAddress string) ([]types.JobDataDTO, error) {
	session := r.db.Session()
	iter := session.Query(GetJobsBySafeAddressQuery, strings.ToLower(safeAddress)).Iter()

	var jobs []types.JobDataDTO
	for {
		var entity types.JobDataEntity
		if !iter.Scan(
			&entity.JobID, &entity.JobTitle, &entity.TaskDefinitionID, &entity.CreatedChainID,
			&entity.UserAddress, &entity.LinkJobID, &entity.ChainStatus, &entity.SafeAddress,
			&entity.Timezone, &entity.JobType, &entity.TimeFrame,
			&entity.Recurring, &entity.Status, &entity.JobCostPrediction, &entity.JobCostActual,
			&entity.TaskIDs, &entity.CreatedAt, &entity.UpdatedAt, &entity.LastExecutedAt,
		) {
			break
		}
		dto := types.JobDataEntityToDTO(&entity)
		jobs = append(jobs, *dto)
	}

	if err := iter.Close(); err != nil {
		return nil, err
	}
	return jobs, nil
}
