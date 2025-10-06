package types

import (
	"math/big"
	"time"
)

type CreateTaskDataRequest struct {
	JobID            *big.Int `json:"job_id" validate:"required"`
	TaskDefinitionID int      `json:"task_definition_id" validate:"required"`
	IsImua           bool     `json:"is_imua"`
}

type UpdateTaskExecutionDataRequest struct {
	TaskID             int64     `json:"task_id" validate:"required"`
	TaskPerformerID    int64     `json:"task_performer_id" validate:"required"`
	ExecutionTimestamp time.Time `json:"execution_timestamp" validate:"required"`
	ExecutionTxHash    string    `json:"execution_tx_hash" validate:"required"`
	ProofOfTask        string    `json:"proof_of_task" validate:"required"`
	TaskOpXCost        float64   `json:"task_opx_cost" validate:"required"`
}

type UpdateTaskAttestationDataRequest struct {
	TaskID               int64   `json:"task_id" validate:"required"`
	TaskNumber           int64   `json:"task_number" validate:"required"`
	TaskAttesterIDs      []int64 `json:"task_attester_ids" validate:"required"`
	TpSignature          []byte  `json:"tp_signature" validate:"required"`
	TaSignature          []byte  `json:"ta_signature" validate:"required"`
	TaskSubmissionTxHash string  `json:"task_submission_tx_hash" validate:"required"`
	IsSuccessful         bool    `json:"is_successful" validate:"required"`
}

type TasksByJobIDResponse struct {
	TaskID             int64     `json:"task_id"`
	TaskNumber         int64     `json:"task_number"`
	TaskOpXCost        float64   `json:"task_opx_cost"`
	ExecutionTimestamp time.Time `json:"execution_timestamp"`
	ExecutionTxHash    string    `json:"execution_tx_hash"`
	TaskPerformerID    int64     `json:"task_performer_id"`
	TaskAttesterIDs    []int64   `json:"task_attester_ids"`
	TaskStatus         string    `json:"task_status"`
	IsAccepted         bool      `json:"is_accepted"`
	TxURL              string    `json:"tx_url"`
	ConvertedArguments []string  `json:"converted_arguments"`
}

type GetTasksByJobID struct {
	TaskID             int64     `json:"task_id"`
	TaskNumber         int64     `json:"task_number"`
	TaskOpXCost        float64   `json:"task_opx_cost"`
	ExecutionTimestamp time.Time `json:"execution_timestamp"`
	ExecutionTxHash    string    `json:"execution_tx_hash"`
	TaskPerformerID    int64     `json:"task_performer_id"`
	TaskAttesterIDs    []int64   `json:"task_attester_ids"`
	IsAccepted         bool      `json:"is_accepted"`
	TxURL              string    `json:"tx_url"`
	TaskStatus         string    `json:"task_status"`
	ConvertedArguments []string  `json:"converted_arguments"`
}
