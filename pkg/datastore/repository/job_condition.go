package repository

import (
	"errors"
	"math/big"
	"time"

	"github.com/trigg3rX/triggerx-backend/pkg/datastore/connection"
	"github.com/trigg3rX/triggerx-backend/pkg/datastore/repository/queries"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

type conditionJobRepository struct {
	db connection.ConnectionManager
}

func NewConditionJobRepository(db connection.ConnectionManager) ConditionJobRepository {
	return &conditionJobRepository{
		db: db,
	}
}

func (r *conditionJobRepository) CreateConditionJob(conditionJob *types.ConditionJobData) error {
	err := r.db.GetSession().Query(queries.CreateConditionJobDataQuery,
		conditionJob.JobID, conditionJob.TaskDefinitionID, conditionJob.ExpirationTime, conditionJob.Recurring,
		conditionJob.ConditionType, conditionJob.UpperLimit, conditionJob.LowerLimit,
		conditionJob.ValueSourceType, conditionJob.ValueSourceUrl, conditionJob.TargetChainID,
		conditionJob.TargetContractAddress, conditionJob.TargetFunction,
		conditionJob.ABI, conditionJob.ArgType, conditionJob.Arguments,
		conditionJob.DynamicArgumentsScriptUrl, conditionJob.IsCompleted, conditionJob.ExpirationTime,
		conditionJob.SelectedKeyRoute, time.Now()).Exec()

	if err != nil {
		return err
	}

	return nil
}

func (r *conditionJobRepository) GetConditionJobByJobID(jobID *big.Int) (types.ConditionJobData, error) {
	var conditionJob types.ConditionJobData
	var temp *big.Int
	conditionJob.JobID = jobID
	err := r.db.GetSession().Query(queries.GetConditionJobDataByJobIDQuery, jobID).Scan(
		&temp, &conditionJob.ExpirationTime, &conditionJob.Recurring, &conditionJob.ConditionType,
		&conditionJob.UpperLimit, &conditionJob.LowerLimit, &conditionJob.ValueSourceType,
		&conditionJob.ValueSourceUrl, &conditionJob.TargetChainID, &conditionJob.TargetContractAddress,
		&conditionJob.TargetFunction, &conditionJob.ABI, &conditionJob.ArgType, &conditionJob.Arguments,
		&conditionJob.DynamicArgumentsScriptUrl, &conditionJob.IsCompleted,
		&conditionJob.SelectedKeyRoute,
	)
	if err != nil {
		return types.ConditionJobData{}, errors.New("failed to get condition job by job ID")
	}

	return conditionJob, nil
}

func (r *conditionJobRepository) CompleteConditionJob(jobID *big.Int) error {
	err := r.db.GetSession().Query(queries.CompleteConditionJobStatusQuery, jobID).Exec()
	if err != nil {
		return errors.New("failed to complete condition job")
	}

	err = r.db.GetSession().Query(queries.UpdateJobDataToCompletedQuery, jobID).Exec()
	if err != nil {
		return errors.New("failed to update job_data status to completed")
	}

	return nil
}

func (r *conditionJobRepository) UpdateConditionJobStatus(jobID *big.Int, isActive bool) error {
	err := r.db.GetSession().Query(queries.UpdateConditionJobStatusQuery, isActive, jobID).Exec()
	if err != nil {
		return errors.New("failed to update condition job status")
	}

	return nil
}

func (r *conditionJobRepository) GetActiveConditionJobs() ([]types.ConditionJobData, error) {
	var conditionJobs []types.ConditionJobData
	iter := r.db.GetSession().Query(queries.GetActiveConditionJobsQuery).Iter()
	var conditionJob types.ConditionJobData
	for iter.Scan(
		&conditionJob.JobID, &conditionJob.ExpirationTime, &conditionJob.Recurring,
		&conditionJob.ConditionType, &conditionJob.UpperLimit, &conditionJob.LowerLimit,
		&conditionJob.ValueSourceType, &conditionJob.ValueSourceUrl, &conditionJob.TargetChainID,
		&conditionJob.TargetContractAddress, &conditionJob.TargetFunction, &conditionJob.ABI,
		&conditionJob.ArgType, &conditionJob.Arguments, &conditionJob.DynamicArgumentsScriptUrl,
		&conditionJob.IsCompleted, &conditionJob.SelectedKeyRoute) {
		conditionJobs = append(conditionJobs, conditionJob)
	}
	if err := iter.Close(); err != nil {
		return nil, errors.New("failed to fetch active condition jobs")
	}
	return conditionJobs, nil
}
