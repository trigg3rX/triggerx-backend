package types

import pkgTypes "github.com/trigg3rX/triggerx-backend/pkg/types"

// ReportTaskStatusRequest represents a request to report task execution status from a keeper
// This is called after the aggregator submission attempt (regardless of success or failure)
type ReportTaskStatusRequest struct {
	TaskID              int64  `json:"task_id" validate:"required"`
	KeeperAddress       string `json:"keeper_address" validate:"required"`
	ExecutionSuccessful bool   `json:"execution_successful"`          // Whether the task execution itself succeeded
	AggregatorSubmitted bool   `json:"aggregator_submitted"`          // Whether the aggregator submission succeeded
	Error               string `json:"error,omitempty"`               // Error message if any step failed
	ExecutionTxHash     string `json:"execution_tx_hash,omitempty"`   // Transaction hash from on-chain execution
	ProofCID            string `json:"proof_cid,omitempty"`           // IPFS CID of the proof data
	Signature           string `json:"signature" validate:"required"` // Keeper's signature for authentication
}

// ReportTaskStatusResponse represents the response to a task status report
type ReportTaskStatusResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

// ReportConsensusEventRequest represents a request to report a consensus event (TaskSubmitted or TaskRejected)
// This is called by eventmonitor when it detects on-chain consensus events
// EventMonitor fetches IPFS data (which contains trace context) and sends it along with minimal event data
type ReportConsensusEventRequest struct {
	TxHash     string             `json:"tx_hash" validate:"required"`   // Task submission transaction hash
	IsAccepted bool               `json:"is_accepted"`                   // true for TaskSubmitted, false for TaskRejected
	IPFSData   *pkgTypes.IPFSData `json:"ipfs_data" validate:"required"` // Full IPFS data including trace context
}

// ReportConsensusEventResponse represents the response to a consensus event report
type ReportConsensusEventResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

// --- DEPRECATED: ---
// These types were used for the old report-task-error endpoint.
// Since executor and validator are controlled by us, backward compatibility is unnecessary.
//
// type ReportTaskErrorRequest struct {
// 	TaskID        int64  `json:"task_id" validate:"required"`
// 	KeeperAddress string `json:"keeper_address" validate:"required"`
// 	Error         string `json:"error" validate:"required"`
// 	Signature     string `json:"signature" validate:"required"`
// }
//
// type ReportTaskErrorResponse struct {
// 	Success bool   `json:"success"`
// 	Message string `json:"message,omitempty"`
// }
