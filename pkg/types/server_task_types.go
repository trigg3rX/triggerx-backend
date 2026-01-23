package types

import "time"

// UpdateTaskExecutionDataRequest represents the request to update task execution data
type UpdateTaskExecutionDataRequest struct {
	TaskID               int64     `json:"task_id" validate:"required"`
	TaskPerformerAddress string    `json:"task_performer_address" validate:"required"`
	ExecutionTimestamp   time.Time `json:"execution_timestamp" validate:"required"`
	ExecutionTxHash      string    `json:"execution_tx_hash" validate:"required"`
	ProofOfTask          string    `json:"proof_of_task" validate:"required"`
	TaskOpXCost          float64   `json:"task_opx_cost" validate:"required"`
}

// UpdateTaskAttestationDataRequest represents the request to update task attestation data
type UpdateTaskAttestationDataRequest struct {
	TaskID                int64    `json:"task_id" validate:"required"`
	TaskNumber            int64    `json:"task_number" validate:"required"`
	TaskAttesterAddresses []string `json:"task_attester_addresses" validate:"required"`
	TpSignature           []byte   `json:"tp_signature" validate:"required"`
	TaSignature           []byte   `json:"ta_signature" validate:"required"`
	TaskSubmissionTxHash  string   `json:"task_submission_tx_hash" validate:"required"`
	IsSuccessful          bool     `json:"is_successful" validate:"required"`
}

// TasksByJobIDResponse represents task data grouped by job ID
type TasksByJobIDResponse struct {
	TaskID                int64     `json:"task_id"`
	TaskNumber            int64     `json:"task_number"`
	TaskOpXCost           float64   `json:"task_opx_cost"`
	ExecutionTimestamp    time.Time `json:"execution_timestamp"`
	ExecutionTxHash       string    `json:"execution_tx_hash"`
	TaskPerformerAddress  string    `json:"task_performer_address"`
	TaskAttesterAddresses []string  `json:"task_attester_addresses"`
	TaskStatus            string    `json:"task_status"`
	TaskError             string    `json:"task_error"`
	IsAccepted            bool      `json:"is_accepted"`
	TxURL                 string    `json:"tx_url"`
	ConvertedArguments    []string  `json:"converted_arguments"`
}

// GetTasksByJobID represents task data for a specific job
type GetTasksByJobID struct {
	TaskID                int64     `json:"task_id"`
	TaskNumber            int64     `json:"task_number"`
	TaskOpXCost           float64   `json:"task_opx_cost"`
	ExecutionTimestamp    time.Time `json:"execution_timestamp"`
	ExecutionTxHash       string    `json:"execution_tx_hash"`
	TaskPerformerAddress  []string   `json:"task_performer_address"`
	TaskAttesterAddress []string  `json:"task_attester_address"`
	IsAccepted            bool      `json:"is_accepted"`
	TxURL                 string    `json:"tx_url"`
	TaskStatus            string    `json:"task_status"`
	TaskError             string    `json:"task_error"`
	ConvertedArguments    []string  `json:"converted_arguments"`
}

// TasksByJobGroupResponse groups task data by job identifier
type TasksByJobGroupResponse struct {
	JobID string                 `json:"job_id"`
	Tasks []TasksByJobIDResponse `json:"tasks"`
}

// RecentTaskResponse represents a task in the recent tasks list for the landing page
type RecentTaskResponse struct {
	TaskID                int64     `json:"task_id"`
	TaskNumber            int64     `json:"task_number"`
	JobID                 string    `json:"job_id"`
	TaskDefinitionID      int       `json:"task_definition_id"`
	CreatedAt             time.Time `json:"created_at"`
	TaskOpXCost           float64   `json:"task_opx_cost"`
	ExecutionTimestamp    time.Time `json:"execution_timestamp"`
	ExecutionTxHash       string    `json:"execution_tx_hash"`
	TxURL                 string    `json:"tx_url"`
	TaskPerformerAddress  string    `json:"task_performer_address"`
	TaskAttesterAddresses []string  `json:"task_attester_addresses"`
	TaskStatus            string    `json:"task_status"`
	TaskError             string    `json:"task_error"`
	IsImua                bool      `json:"is_imua"`
}
