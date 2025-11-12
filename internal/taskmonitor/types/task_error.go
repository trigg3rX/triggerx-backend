package types

// ReportTaskErrorRequest represents a request to report a task error from a keeper
type ReportTaskErrorRequest struct {
	TaskID        int64  `json:"task_id" validate:"required"`
	KeeperAddress string `json:"keeper_address" validate:"required"`
	Error         string `json:"error" validate:"required"`
	Signature     string `json:"signature" validate:"required"`
}

// ReportTaskErrorResponse represents the response to a task error report
type ReportTaskErrorResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}
