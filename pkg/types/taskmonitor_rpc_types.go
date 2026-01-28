package types

import "time"

// ReportTaskExecutionStatusRequest represents a request to report task execution status from a keeper
// This is called after the aggregator submission attempt (regardless of success or failure)
// Owned by: taskmonitor RPC Server (keeper reports execution status to taskmonitor)
type ReportTaskExecutionStatusRequest struct {
	TaskID              int64         `json:"task_id" validate:"required"`        // Task identifier
	KeeperAddress       string        `json:"keeper_address" validate:"required"` // Keeper address
	Signature           string        `json:"signature" validate:"required"`      // Keeper's signature for authentication
	IPFSDataCID         string        `json:"ipfs_data_cid,omitempty"`            // IPFS CID of the proof data
	ExecutionSuccessful bool          `json:"execution_successful"`               // Whether the task execution itself succeeded
	AggregatorSubmitted bool          `json:"aggregator_submitted"`               // Whether the aggregator submission succeeded
	Error               string        `json:"error,omitempty"`                    // Error message if any step failed
	ExecutionTxHash     string        `json:"execution_tx_hash,omitempty"`        // Transaction hash from on-chain execution
	ExecutedAt          time.Time     `json:"executed_at,omitempty"`              // Time when the task execution was performed
	TaskOpxActualCost   string        `json:"task_opx_actual_cost,omitempty"`     // Actual cost in Wei (from TotalFee)
	ConvertedArguments  []interface{} `json:"converted_arguments,omitempty"`      // Arguments used in execution
}

// ReportTaskExecutionStatusResponse represents the response to a task execution status report
// Owned by: taskmonitor RPC Server
type ReportTaskExecutionStatusResponse struct {
	Success bool   `json:"success"`           // Whether the report was successful
	Message string `json:"message,omitempty"` // Optional message
}

// ReportTaskConsensusStatusRequest represents a request to report a consensus event (TaskSubmitted or TaskRejected)
// This is called by eventmonitor when it detects on-chain consensus events
// EventMonitor fetches IPFS data (which contains trace context) and sends it along with minimal event data
// Owned by: taskmonitor RPC Server (eventmonitor reports consensus events to taskmonitor)
type ReportTaskConsensusStatusRequest struct {
	TaskID               int64   `json:"task_id" validate:"required"`                 // Task identifier
	Network              string  `json:"network" validate:"required"`                 // Network name
	TaskNumber           int64   `json:"task_number" validate:"required"`             // Task number
	TaskOpxActualCost    string  `json:"task_opx_actual_cost" validate:"required"`    // Task actual cost in Wei
	TaskSubmissionTxHash string  `json:"task_submission_tx_hash" validate:"required"` // Task submission transaction hash
	IsAccepted           bool    `json:"is_accepted"`                                 // true for TaskSubmitted, false for TaskRejected
	AttesterIds          []int64 `json:"attester_ids"`                                // Attester IDs
	IPFSDataCID          string  `json:"ipfs_data_cid" validate:"required"`          // IPFS CID of the proof data
}

// ReportTaskConsensusStatusResponse represents the response to a consensus event report
// Owned by: taskmonitor RPC Server
type ReportTaskConsensusStatusResponse struct {
	Success bool   `json:"success"`           // Whether the report was successful
	Message string `json:"message,omitempty"` // Optional message
}

// TaskSubmissionData represents aggregated data for updating task_data table after consensus
// This is an internal type used by taskmonitor to pass data to the database layer
// It combines data from IPFSData and the consensus event for the final DB update
// Owned by: taskmonitor (internal use only)
type TaskSubmissionData struct {
	// Task identification
	TaskID           int64 `json:"task_id"`            // Task identifier (from IPFSData.ActionData)
	TaskDefinitionID int   `json:"task_definition_id"` // Task definition ID (1-9, from IPFSData.TaskData)

	// Consensus data (from on-chain event)
	TaskNumber           int64   `json:"task_number"`             // Task sequence number on the contract
	IsAccepted           bool    `json:"is_accepted"`             // true for TaskSubmitted, false for TaskRejected
	TaskSubmissionTxHash string  `json:"task_submission_tx_hash"` // Task submission transaction hash
	AttesterIds          []int64 `json:"attester_ids"`            // Attester operator IDs (converted to addresses in repository)

	// Proof data (from IPFSData.ProofData)
	ProofOfTask string `json:"proof_of_task"` // Proof of task (IPFS hash of proof data)

	// Performer data (from IPFSData.PerformerSignature)
	PerformerAddress string `json:"performer_address"` // Performer's consensus address (converted to keeper address in repository)
}
