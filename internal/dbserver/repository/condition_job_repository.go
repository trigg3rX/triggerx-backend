package repository

import (
	"errors"
	"math/big"
	"time"

	"github.com/trigg3rX/triggerx-backend/internal/dbserver/repository/queries"
	"github.com/trigg3rX/triggerx-backend/pkg/database"
	commonTypes "github.com/trigg3rX/triggerx-backend/pkg/types"
)

type ConditionJobRepository interface {
	CreateConditionJob(conditionJob *commonTypes.ConditionJobData) error
	GetConditionJobByJobID(jobID *big.Int) (commonTypes.ConditionJobData, error)
	CompleteConditionJob(jobID *big.Int) error
	UpdateConditionJobStatus(jobID *big.Int, isActive bool) error
	GetActiveConditionJobs() ([]commonTypes.ConditionJobData, error)
}

type conditionJobRepository struct {
	db *database.Connection
}

func NewConditionJobRepository(db *database.Connection) ConditionJobRepository {
	return &conditionJobRepository{
		db: db,
	}
}

func (r *conditionJobRepository) CreateConditionJob(conditionJob *commonTypes.ConditionJobData) error {
	err := r.db.Session().Query(queries.CreateConditionJobDataQuery,
		conditionJob.JobID.ToBigInt(), conditionJob.TaskDefinitionID, conditionJob.ExpirationTime, conditionJob.Recurring,
		conditionJob.ConditionType, conditionJob.UpperLimit, conditionJob.LowerLimit,
		conditionJob.ValueSourceType, conditionJob.ValueSourceUrl, conditionJob.TargetChainID,
		conditionJob.TargetContractAddress, conditionJob.TargetFunction,
		conditionJob.ABI, conditionJob.ArgType, conditionJob.Arguments,
		conditionJob.DynamicArgumentsScriptUrl, conditionJob.IsCompleted, conditionJob.IsActive,
		conditionJob.SelectedKeyRoute, time.Now(), time.Now()).Exec()

	if err != nil {
		return err
	}

	return nil
}

func (r *conditionJobRepository) GetConditionJobByJobID(jobID *big.Int) (commonTypes.ConditionJobData, error) {
	var conditionJob commonTypes.ConditionJobData
	var temp *big.Int
	conditionJob.JobID = commonTypes.NewBigInt(jobID)
	err := r.db.Session().Query(queries.GetConditionJobDataByJobIDQuery, jobID).Scan(
		&temp, &conditionJob.ExpirationTime, &conditionJob.Recurring, &conditionJob.ConditionType,
		&conditionJob.UpperLimit, &conditionJob.LowerLimit, &conditionJob.ValueSourceType,
		&conditionJob.ValueSourceUrl, &conditionJob.TargetChainID, &conditionJob.TargetContractAddress,
		&conditionJob.TargetFunction, &conditionJob.ABI, &conditionJob.ArgType, &conditionJob.Arguments,
		&conditionJob.DynamicArgumentsScriptUrl, &conditionJob.IsCompleted, &conditionJob.IsActive,
		&conditionJob.SelectedKeyRoute,
	)
	if err != nil {
		return commonTypes.ConditionJobData{}, errors.New("failed to get condition job by job ID")
	}

	return conditionJob, nil
}

func (r *conditionJobRepository) CompleteConditionJob(jobID *big.Int) error {
	err := r.db.Session().Query(queries.CompleteConditionJobStatusQuery, jobID).Exec()
	if err != nil {
		return errors.New("failed to complete condition job")
	}

	err = r.db.Session().Query(queries.UpdateJobDataToCompletedQuery, jobID).Exec()
	if err != nil {
		return errors.New("failed to update job_data status to completed")
	}

	return nil
}

func (r *conditionJobRepository) UpdateConditionJobStatus(jobID *big.Int, isActive bool) error {
	err := r.db.Session().Query(queries.UpdateConditionJobStatusQuery, isActive, jobID).Exec()
	if err != nil {
		return errors.New("failed to update condition job status")
	}

	return nil
}

func (r *conditionJobRepository) GetActiveConditionJobs() ([]commonTypes.ConditionJobData, error) {
	var conditionJobs []commonTypes.ConditionJobData
	iter := r.db.Session().Query(queries.GetActiveConditionJobsQuery).Iter()
	var conditionJob commonTypes.ConditionJobData
	var jobIDBigInt *big.Int
	for iter.Scan(
		&jobIDBigInt, &conditionJob.ExpirationTime, &conditionJob.Recurring,
		&conditionJob.ConditionType, &conditionJob.UpperLimit, &conditionJob.LowerLimit,
		&conditionJob.ValueSourceType, &conditionJob.ValueSourceUrl, &conditionJob.TargetChainID,
		&conditionJob.TargetContractAddress, &conditionJob.TargetFunction, &conditionJob.ABI,
		&conditionJob.ArgType, &conditionJob.Arguments, &conditionJob.DynamicArgumentsScriptUrl,
		&conditionJob.IsCompleted, &conditionJob.IsActive, &conditionJob.SelectedKeyRoute) {
		conditionJob.JobID = commonTypes.NewBigInt(jobIDBigInt)
		conditionJobs = append(conditionJobs, conditionJob)
	}
	if err := iter.Close(); err != nil {
		return nil, errors.New("failed to fetch active condition jobs")
	}
	return conditionJobs, nil
}
