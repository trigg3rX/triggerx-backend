package types

import (
	"math/big"
	"time"
)

type UpdateJobRequest struct {
	JobID          *big.Int     `json:"job_id"`
	Recurring      bool      `json:"recurring"`
	TimeFrame      int64     `json:"time_frame"`
	UpdatedAt      time.Time `json:"updated_at"`
	LastExecutedAt time.Time `json:"last_executed_at"`
}

type UpdateTaskExecutionDataRequest struct {
	TaskID             int64     `json:"task_id" validate:"required"`
	TaskPerformerID    int64     `json:"task_performer_id" validate:"required"`
	ExecutionTimestamp time.Time `json:"execution_timestamp" validate:"required"`
	ExecutionTxHash    string    `json:"execution_tx_hash" validate:"required"`
	ProofOfTask        string    `json:"proof_of_task" validate:"required"`
	TaskOpXCost        float64   `json:"task_opx_cost" validate:"required"`
}