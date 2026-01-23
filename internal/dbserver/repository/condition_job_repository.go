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
	UpdateConditionJobStatus(jobID *big.Int, isActive bool) error
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

func (r *conditionJobRepository) UpdateConditionJobStatus(jobID *big.Int, isActive bool) error {
	err := r.db.Session().Query(queries.UpdateConditionJobStatusQuery, isActive, jobID).Exec()
	if err != nil {
		return errors.New("failed to update condition job status")
	}

	return nil
}
