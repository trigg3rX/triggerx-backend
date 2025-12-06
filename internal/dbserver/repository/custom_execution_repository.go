package repository

import (
	"math/big"

	"github.com/trigg3rX/triggerx-backend/internal/dbserver/repository/queries"
	"github.com/trigg3rX/triggerx-backend/pkg/database"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// CustomExecutionRepository handles custom script execution tracking
type CustomExecutionRepository interface {
	CreateExecution(exec *types.CustomScriptExecution) error
	GetExecutionByID(executionID string) (*types.CustomScriptExecution, error)
	GetExecutionsByJobID(jobID *big.Int) ([]types.CustomScriptExecution, error)
	GetExecutionsByTaskID(taskID int64) ([]types.CustomScriptExecution, error)
	UpdateExecutionTxHash(executionID string, txHash string, status string) error
	UpdateVerificationStatus(executionID string, status string) error
	UpdateChallengeStatus(executionID string, isChallenged bool, count int) error
}

type customExecutionRepository struct {
	db *database.Connection
}

// NewCustomExecutionRepository creates a new execution repository
func NewCustomExecutionRepository(db *database.Connection) CustomExecutionRepository {
	return &customExecutionRepository{
		db: db,
	}
}

func (r *customExecutionRepository) CreateExecution(exec *types.CustomScriptExecution) error {
	return r.db.Session().Query(queries.CreateExecutionRecordQuery,
		exec.ExecutionID,
		exec.JobID.ToBigInt(),
		exec.TaskID,
		exec.ScheduledTime,
		exec.ActualTime,
		exec.PerformerAddress,
		exec.InputTimestamp,
		exec.InputStorage,
		exec.InputHash,
		exec.ShouldExecute,
		exec.TargetContract,
		exec.Calldata,
		exec.OutputHash,
		exec.ExecutionMetadata,
		exec.ScriptHash,
		exec.Signature,
		exec.TxHash,
		exec.ExecutionStatus,
		exec.ExecutionError,
		exec.VerificationStatus,
		exec.ChallengeDeadline,
		exec.IsChallenged,
		exec.ChallengeCount,
		exec.CreatedAt,
	).Exec()
}

func (r *customExecutionRepository) GetExecutionByID(executionID string) (*types.CustomScriptExecution, error) {
	var exec types.CustomScriptExecution

	err := r.db.Session().Query(queries.GetExecutionByIDQuery, executionID).Scan(
		&exec.ExecutionID,
		&exec.JobID,
		&exec.TaskID,
		&exec.ScheduledTime,
		&exec.ActualTime,
		&exec.PerformerAddress,
		&exec.InputTimestamp,
		&exec.InputStorage,
		&exec.InputHash,
		&exec.ShouldExecute,
		&exec.TargetContract,
		&exec.Calldata,
		&exec.OutputHash,
		&exec.ExecutionMetadata,
		&exec.ScriptHash,
		&exec.Signature,
		&exec.TxHash,
		&exec.ExecutionStatus,
		&exec.ExecutionError,
		&exec.VerificationStatus,
		&exec.ChallengeDeadline,
		&exec.IsChallenged,
		&exec.ChallengeCount,
		&exec.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &exec, nil
}

func (r *customExecutionRepository) GetExecutionsByJobID(jobID *big.Int) ([]types.CustomScriptExecution, error) {
	iter := r.db.Session().Query(queries.GetExecutionsByJobIDQuery, jobID).Iter()

	var executions []types.CustomScriptExecution
	var exec types.CustomScriptExecution

	for iter.Scan(
		&exec.ExecutionID,
		&exec.JobID,
		&exec.TaskID,
		&exec.ScheduledTime,
		&exec.ActualTime,
		&exec.PerformerAddress,
		&exec.InputTimestamp,
		&exec.InputStorage,
		&exec.InputHash,
		&exec.ShouldExecute,
		&exec.TargetContract,
		&exec.Calldata,
		&exec.OutputHash,
		&exec.ExecutionMetadata,
		&exec.ScriptHash,
		&exec.Signature,
		&exec.TxHash,
		&exec.ExecutionStatus,
		&exec.ExecutionError,
		&exec.VerificationStatus,
		&exec.ChallengeDeadline,
		&exec.IsChallenged,
		&exec.ChallengeCount,
		&exec.CreatedAt,
	) {
		executions = append(executions, exec)
	}

	if err := iter.Close(); err != nil {
		return nil, err
	}

	return executions, nil
}

func (r *customExecutionRepository) GetExecutionsByTaskID(taskID int64) ([]types.CustomScriptExecution, error) {
	iter := r.db.Session().Query(queries.GetExecutionsByTaskIDQuery, taskID).Iter()

	var executions []types.CustomScriptExecution
	var exec types.CustomScriptExecution

	for iter.Scan(
		&exec.ExecutionID,
		&exec.JobID,
		&exec.TaskID,
		&exec.ScheduledTime,
		&exec.ActualTime,
		&exec.PerformerAddress,
		&exec.InputTimestamp,
		&exec.InputStorage,
		&exec.InputHash,
		&exec.ShouldExecute,
		&exec.TargetContract,
		&exec.Calldata,
		&exec.OutputHash,
		&exec.ExecutionMetadata,
		&exec.ScriptHash,
		&exec.Signature,
		&exec.TxHash,
		&exec.ExecutionStatus,
		&exec.ExecutionError,
		&exec.VerificationStatus,
		&exec.ChallengeDeadline,
		&exec.IsChallenged,
		&exec.ChallengeCount,
		&exec.CreatedAt,
	) {
		executions = append(executions, exec)
	}

	if err := iter.Close(); err != nil {
		return nil, err
	}

	return executions, nil
}

func (r *customExecutionRepository) UpdateExecutionTxHash(executionID string, txHash string, status string) error {
	return r.db.Session().Query(queries.UpdateExecutionTxHashQuery,
		txHash,
		status,
		executionID,
	).Exec()
}

func (r *customExecutionRepository) UpdateVerificationStatus(executionID string, status string) error {
	return r.db.Session().Query(queries.UpdateExecutionVerificationStatusQuery,
		status,
		executionID,
	).Exec()
}

func (r *customExecutionRepository) UpdateChallengeStatus(executionID string, isChallenged bool, count int) error {
	return r.db.Session().Query(queries.UpdateExecutionChallengeStatusQuery,
		isChallenged,
		count,
		executionID,
	).Exec()
}
