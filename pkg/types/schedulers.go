package types

import (
	"math/big"
	"time"
)

// Target Data for all task types
// DEVNOTE: I separated this from all schedule data types to accomodate multiple target calls on same trigger in future
type TaskTargetData struct {
	JobID                     *big.Int `json:"job_id"`
	TaskID                    int64    `json:"task_id"`
	TaskDefinitionID          int      `json:"task_definition_id"`
	TargetChainID             string   `json:"target_chain_id"`
	TargetContractAddress     string   `json:"target_contract_address"`
	TargetFunction            string   `json:"target_function"`
	ABI                       string   `json:"abi"`
	ArgType                   int      `json:"arg_type"`
	Arguments                 []string `json:"arguments"`
	DynamicArgumentsScriptUrl string   `json:"dynamic_arguments_script_url"`
	IsImua                    bool     `json:"is_imua"`
}

// Monitoring Data for even and condition workers
type EventWorkerData struct {
	JobID                  *big.Int  `json:"job_id"`
	ExpirationTime         time.Time `json:"expiration_time"`
	Recurring              bool      `json:"recurring"`
	TriggerChainID         string    `json:"trigger_chain_id"`
	TriggerContractAddress string    `json:"trigger_contract_address"`
	TriggerEvent           string    `json:"trigger_event"`
}
type ConditionWorkerData struct {
	JobID           *big.Int  `json:"job_id"`
	ExpirationTime  time.Time `json:"expiration_time"`
	Recurring       bool      `json:"recurring"`
	ConditionType   string    `json:"condition_type"`
	UpperLimit      float64   `json:"upper_limit"`
	LowerLimit      float64   `json:"lower_limit"`
	ValueSourceType string    `json:"value_source_type"`
	ValueSourceUrl  string    `json:"value_source_url"`
}

// Data to pass to time scheduler
type ScheduleTimeTaskData struct {
	TaskID                 int64          `json:"task_id"`
	TaskDefinitionID       int            `json:"task_definition_id"`
	LastExecutedAt         time.Time      `json:"last_executed_at"`
	ExpirationTime         time.Time      `json:"expiration_time"`
	NextExecutionTimestamp time.Time      `json:"next_execution_timestamp"`
	ScheduleType           string         `json:"schedule_type"`
	TimeInterval           int64          `json:"time_interval"`
	CronExpression         string         `json:"cron_expression"`
	SpecificSchedule       string         `json:"specific_schedule"`
	TaskTargetData         TaskTargetData `json:"task_target_data"`
	IsImua                 bool           `json:"is_imua"`
}

// Data to pass to condition scheduler
type ScheduleConditionJobData struct {
	JobID               *big.Int            `json:"job_id"`
	TaskDefinitionID    int                 `json:"task_definition_id"`
	LastExecutedAt      time.Time           `json:"last_executed_at"`
	TaskTargetData      TaskTargetData      `json:"task_target_data"`
	EventWorkerData     EventWorkerData     `json:"event_worker_data"`
	ConditionWorkerData ConditionWorkerData `json:"condition_worker_data"`
	IsImua              bool                `json:"is_imua"`
}

// Trigger data from schedulers to keepers for validation
type TaskTriggerData struct {
	TaskID           int64     `json:"task_id"`
	TaskDefinitionID int       `json:"task_definition_id"`
	Recurring        bool      `json:"recurring"`
	ExpirationTime   time.Time `json:"expiration_time"`
	// For event and Condition, it would be when the trigger happened, for time it would when the Job.LastExecutedAt
	CurrentTriggerTimestamp time.Time `json:"trigger_timestamp"`

	NextTriggerTimestamp time.Time `json:"next_trigger_timestamp"`
	TimeScheduleType     string    `json:"time_schedule_type"`
	TimeCronExpression   string    `json:"time_cron_expression"`
	TimeSpecificSchedule string    `json:"time_specific_schedule"`
	TimeInterval         int64     `json:"time_interval"`

	EventChainId                string `json:"event_chain_id"`
	EventTxHash                 string `json:"event_tx_hash"`
	EventTriggerContractAddress string `json:"event_trigger_contract_address"`
	EventTriggerName            string `json:"event_trigger_name"`

	ConditionType           string `json:"condition_type"`
	ConditionSourceType     string `json:"condition_source_type"`
	ConditionSourceUrl      string `json:"condition_source_url"`
	ConditionUpperLimit     int    `json:"condition_upper_limit"`
	ConditionLowerLimit     int    `json:"condition_lower_limit"`
	ConditionSatisfiedValue int    `json:"condition_satisfied_value"`
}

type SendTaskDataToKeeper struct {
	TaskID             []int64                   `json:"task_id"`
	PerformerData      PerformerData           `json:"performer_data"`
	TargetData         []TaskTargetData        `json:"target_data"`
	TriggerData        []TaskTriggerData       `json:"trigger_data"`
	SchedulerID        int                     `json:"scheduler_id"`
	ManagerSignature   string                  `json:"manager_signature"`
}

// SchedulerTaskRequest represents the request format for TaskManager
type SchedulerTaskRequest struct {
	SendTaskDataToKeeper SendTaskDataToKeeper `json:"send_task_data_to_keeper"`
	Source               string             `json:"source"`
}

// TaskManagerAPIResponse represents the response from TaskManager
type TaskManagerAPIResponse struct {
	Success   bool                `json:"success"`
	TaskID    []int64             `json:"task_id"`
	Message   string              `json:"message"`
	Timestamp string              `json:"timestamp"`
	Error     string              `json:"error,omitempty"`
	Details   string              `json:"details,omitempty"`
}

type BroadcastDataForPerformer struct {
	TaskID           int64  `json:"task_id"`
	TaskDefinitionID int    `json:"task_definition_id"`
	PerformerAddress string `json:"performer_address"`
	Data             []byte `json:"data"`
}
