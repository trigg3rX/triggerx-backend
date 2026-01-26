package types

import (
	"encoding/json"
	"fmt"
	"time"
)

// ScheduleConditionJobRequest represents data to pass to condition scheduler
// Owned by: scheduler RPC Server (used by dbserver to schedule/unschedule event/condition jobs)
type ScheduleConditionJobRequest struct {
	JobID               string              `json:"job_id"`
	TaskDefinitionID    int                 `json:"task_definition_id"`
}

// ScheduleConditionJobResponse represents the response from condition scheduler
// Owned by: scheduler RPC Server (response from condition scheduler to dbserver)
type ScheduleConditionJobResponse struct {
	Success bool   `json:"success"`
	JobID   string `json:"job_id"`
	Message string `json:"message"`
}

// CreateTaskDataRequest represents request to create task data
// Owned by: schedulers (used internally for task creation)
type CreateTaskDataRequest struct {
	JobID            string `json:"job_id" validate:"required"`
	TaskDefinitionID int    `json:"task_definition_id" validate:"required"`
	Network          string `json:"network" validate:"required"`
}

// EventWorkerData represents monitoring data for event workers
// Owned by: condition scheduler (used for event monitoring, sent to event monitor service)
type EventWorkerData struct {
	JobID                  string    `json:"job_id"`
	ExpirationTime         time.Time `json:"expiration_time"`
	Recurring              bool      `json:"recurring"`
	TriggerChainID         string    `json:"trigger_chain_id"`
	TriggerContractAddress string    `json:"trigger_contract_address"`
	TriggerEvent           string    `json:"trigger_event"`
	EventFilterParaName    string    `json:"event_filter_para_name"`
	EventFilterValue       string    `json:"event_filter_value"`
}

// ConditionWorkerData represents monitoring data for condition workers
// Owned by: condition scheduler (used internally for condition monitoring)
type ConditionWorkerData struct {
	JobID            string    `json:"job_id"`
	ExpirationTime   time.Time `json:"expiration_time"`
	Recurring        bool      `json:"recurring"`
	ConditionType    string    `json:"condition_type"`
	SelectedKeyRoute string    `json:"selected_key_route"`
	UpperLimit       float64   `json:"upper_limit"`
	LowerLimit       float64   `json:"lower_limit"`
	ValueSourceType  string    `json:"value_source_type"`
	ValueSourceUrl   string    `json:"value_source_url"`
}

// ScheduleConditionJobData represents complete job data for condition scheduler
// Owned by: condition scheduler (internal use for storing job data)
type ScheduleConditionJobData struct {
	JobID               string                `json:"job_id"`
	TaskDefinitionID    int                   `json:"task_definition_id"`
	Network              Network               `json:"network"`
	EventWorkerData      *EventWorkerData      `json:"event_worker_data,omitempty"`
	ConditionWorkerData  *ConditionWorkerData  `json:"condition_worker_data,omitempty"`
	TaskTargetData       TaskTargetData        `json:"task_target_data"`
}


// TaskTargetData represents target data for all task types
// Shared across: schedulers, keepers
// Merged into SendTaskDataToKeeper, which in turn is merged into TaskIPFSData
type TaskTargetData struct {
	JobID                     string   `json:"job_id"`
	TaskID                    int64    `json:"task_id"`
	TaskDefinitionID          int      `json:"task_definition_id"`
	TargetChainID             string   `json:"target_chain_id"`
	TargetContractAddress     string   `json:"target_contract_address"`
	TargetFunction            string   `json:"target_function"`
	ABI                       string   `json:"abi"`
	ArgType                   int      `json:"arg_type"`
	Arguments                 []string `json:"arguments"`
	ExecutionScriptURL      string            `json:"execution_script_url,omitempty"`      // IPFS URL of the agent script
	ExecutionScriptLanguage string            `json:"execution_script_language,omitempty"` // Script language: 'ts', 'go', 'python', 'javascript'
	ExecutionScriptHash     string            `json:"execution_script_hash,omitempty"`     // keccak256(scriptCode) for verification
	MaxExecutionTime        int               `json:"max_execution_time,omitempty"`        // Script timeout in seconds (default 60)
	ChallengePeriod         int64             `json:"challenge_period,omitempty"`          // Challenge period in seconds (default 21600 = 6 hours)
	ScriptStorage           []ScriptStorageDTO `json:"script_storage,omitempty"`            // Storage passed from scheduler (for agent jobs)
}

// TaskTriggerData represents trigger data from schedulers to keepers for validation
// Shared across: schedulers, keepers
// Merged into SendTaskDataToKeeper, which in turn is merged into TaskIPFSData
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

// SendTaskDataToKeeper represents data to send to keeper for task execution
// Shared across: schedulers, taskdispatcher, keeper, taskmonitor
type SendTaskDataToKeeper struct {
	TaskID           []int64           `json:"task_id"`
	PerformerData    PerformerData     `json:"performer_data"`
	TargetData       []TaskTargetData  `json:"target_data"`
	TriggerData      []TaskTriggerData `json:"trigger_data"`
	SchedulerID      string            `json:"scheduler_id"`
	ManagerSignature string            `json:"manager_signature"`
	Network          Network           `json:"network"` // Network: mainnet, sepolia, or imua
}

// UnmarshalJSON custom unmarshaler to handle scheduler_id as both string and number (for backward compatibility)
func (s *SendTaskDataToKeeper) UnmarshalJSON(data []byte) error {
	// Use an alias type to avoid infinite recursion
	type Alias SendTaskDataToKeeper
	aux := &struct {
		SchedulerID interface{} `json:"scheduler_id"` // Accept any type
		*Alias
	}{
		Alias: (*Alias)(s),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Convert scheduler_id to string regardless of input type
	switch v := aux.SchedulerID.(type) {
	case string:
		s.SchedulerID = v
	case float64:
		// JSON numbers are unmarshaled as float64
		s.SchedulerID = fmt.Sprintf("%.0f", v)
	case int:
		s.SchedulerID = fmt.Sprintf("%d", v)
	case int64:
		s.SchedulerID = fmt.Sprintf("%d", v)
	case nil:
		s.SchedulerID = ""
	default:
		s.SchedulerID = fmt.Sprintf("%v", v)
	}

	return nil
}
