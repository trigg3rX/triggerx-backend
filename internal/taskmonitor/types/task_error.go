package types

// ReportTaskStatusRequest represents a request to report task execution status from a keeper
// This is called after the aggregator submission attempt (regardless of success or failure)
type ReportTaskStatusRequest struct {
	TaskID        int64  `json:"task_id" validate:"required"`
	KeeperAddress string `json:"keeper_address" validate:"required"`
	Success       bool   `json:"success"`                       // Whether the task was successfully submitted to aggregator
	ProofCID      string `json:"proof_cid,omitempty"`           // IPFS CID containing all execution data
	Error         string `json:"error,omitempty"`               // Error message if any step failed
	Signature     string `json:"signature" validate:"required"` // Keeper's signature for authentication
}

// ReportTaskStatusResponse represents the response to a task status report
type ReportTaskStatusResponse struct {
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
