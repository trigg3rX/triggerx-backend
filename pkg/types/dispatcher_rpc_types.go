package types

// SchedulerTaskRequest represents the request format for TaskManager
// Owned by: scheduler RPC API (schedulers send this to taskdispatcher)
type SchedulerTaskRequest struct {
	SendTaskDataToKeeper SendTaskDataToKeeper `json:"send_task_data_to_keeper"`
	Source               string               `json:"source"`
}

// TaskDispatcherRPCResponse represents the response from TaskDispatcher
// Owned by: taskdispatcher RPC server (response from taskdispatcher to schedulers)
type TaskDispatcherRPCResponse struct {
	Success   bool    `json:"success"`
	TaskID    []int64 `json:"task_id"`
	Message   string  `json:"message"`
	Timestamp string  `json:"timestamp"`
	Error     string  `json:"error,omitempty"`
	Details   string  `json:"details,omitempty"`
}

// BroadcastDataForPerformer represents data to broadcast to performer (via the aggregator JSON-RPC API, SendCustomMessage method)
// Owned by: taskdispatcher (used by taskdispatcher to broadcast data to performer, via pkg/client/aggregator/custom.go)
type BroadcastDataForPerformer struct {
	TaskID           int64  `json:"task_id"`  // Debug Utility
	TaskDefinitionID int    `json:"task_definition_id"`
	PerformerAddress string `json:"performer_address"`  // Debug Utility
	Data             []byte `json:"data"`    // Serialized SendTaskDataToKeeper struct
}
