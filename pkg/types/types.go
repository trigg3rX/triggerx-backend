package types

// ActionData represents the result of executing a task
type ActionData struct {
	TaskID        uint64      `json:"taskId"`
	ResourceStats interface{} `json:"resourceStats,omitempty"`
	// ... other existing fields ...
}
