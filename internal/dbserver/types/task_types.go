package types

import "time"

// Task status constants
type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"
	TaskStatusSubmitted TaskStatus = "submitted"
	TaskStatusRejected  TaskStatus = "rejected"
	TaskStatusCompleted TaskStatus = "completed"
	TaskStatusFailed    TaskStatus = "failed"
)

type TaskData struct {
	TaskID               int64     `json:"task_id"`
	TaskNumber           int64     `json:"task_number"`
	JobID                int64     `json:"job_id"`
	TaskDefinitionID     int       `json:"task_definition_id"`
	CreatedAt            time.Time `json:"created_at"`
	TaskOpXCost          float64   `json:"task_opx_cost"`
	ExecutionTimestamp   time.Time `json:"execution_timestamp"`
	ExecutionTxHash      string    `json:"execution_tx_hash"`
	TaskPerformerID      int64     `json:"task_performer_id"`
	ProofOfTask          string    `json:"proof_of_task"`
	TaskAttesterIDs      []int64   `json:"task_attester_ids"`
	TpSignature          []byte    `json:"tp_signature"`
	TaSignature          []byte    `json:"ta_signature"`
	TaskSubmissionTxHash string    `json:"task_submission_tx_hash"`
	IsSuccessful         bool      `json:"is_successful"`
	TaskStatus           string    `json:"task_status"`
}

type CreateTaskDataRequest struct {
	JobID            int64 `json:"job_id" validate:"required"`
	TaskDefinitionID int   `json:"task_definition_id" validate:"required"`
}

type UpdateTaskExecutionDataRequest struct {
	TaskID             int64     `json:"task_id" validate:"required"`
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
	IsSuccessful       bool      `json:"is_successful"`
	TaskStatus         string    `json:"task_status"`
}
