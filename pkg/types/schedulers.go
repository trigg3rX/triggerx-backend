package types

import "time"

// Data to pass to time scheduler
type ScheduleTimeJobData struct {
	JobID                         int64     `json:"job_id"`
	TaskDefinitionID              int       `json:"task_definition_id"`
	LastExecutedAt                time.Time `json:"last_executed_at"`
	ExpirationTime                time.Time `json:"expiration_time"`
	TimeInterval                  int64     `json:"time_interval"`
	ScheduleType                  string    `json:"schedule_type"`
	CronExpression                string    `json:"cron_expression"`
	SpecificSchedule              string    `json:"specific_schedule"`
	NextExecutionTimestamp        time.Time `json:"next_execution_timestamp"`
	TargetChainID                 string    `json:"target_chain_id"`
	TargetContractAddress         string    `json:"target_contract_address"`
	TargetFunction                string    `json:"target_function"`
	ABI                           string    `json:"abi"`
	ArgType                       int       `json:"arg_type"`
	Arguments                     []string  `json:"arguments"`
	DynamicArgumentsScriptUrl 	  string    `json:"dynamic_arguments_script_url"`
}

// Data to pass to event scheduler
type ScheduleEventJobData struct {
	JobID                         int64     `json:"job_id"`
	TaskDefinitionID              int       `json:"task_definition_id"`
	LastExecutedAt                time.Time `json:"last_executed_at"`
	ExpirationTime                time.Time `json:"expiration_time"`
	Recurring                     bool     `json:"recurring"`
	TriggerChainID                string   `json:"trigger_chain_id"`
	TriggerContractAddress        string   `json:"trigger_contract_address"`
	TriggerEvent                  string   `json:"trigger_event"`
	TargetChainID                 string    `json:"target_chain_id"`
	TargetContractAddress         string    `json:"target_contract_address"`
	TargetFunction                string    `json:"target_function"`
	ABI                           string    `json:"abi"`
	ArgType                       int       `json:"arg_type"`
	Arguments                     []string  `json:"arguments"`
	DynamicArgumentsScriptUrl 	  string    `json:"dynamic_arguments_script_url"`
}

// Data to pass to condition scheduler
type ScheduleConditionJobData struct {
	JobID                         int64     `json:"job_id"`
	TaskDefinitionID              int       `json:"task_definition_id"`
	LastExecutedAt                time.Time `json:"last_executed_at"`
	ExpirationTime                time.Time `json:"expiration_time"`
	Recurring                     bool     `json:"recurring"`
	ConditionType                 string   `json:"condition_type"`
	UpperLimit                    float64  `json:"upper_limit"`
	LowerLimit                    float64  `json:"lower_limit"`
	ValueSourceType               string   `json:"value_source_type"`
	ValueSourceUrl                string   `json:"value_source_url"`
	TargetChainID                 string   `json:"target_chain_id"`
	TargetContractAddress         string   `json:"target_contract_address"`
	TargetFunction                string   `json:"target_function"`
	ABI                           string   `json:"abi"`
	ArgType                       int      `json:"arg_type"`
	Arguments                     []string `json:"arguments"`
	DynamicArgumentsScriptUrl string   `json:"dynamic_arguments_script_url"`
}

// Data to pass to Performer to execution action
type SendTaskTargetData struct {
    JobID                         int64     `json:"job_id"`
    TaskDefinitionID              int       `json:"task_definition_id"`
	Recurring                     bool      `json:"recurring"`
	ExpirationTime                time.Time `json:"expiration_time"`
	TimeFrame                     int64     `json:"time_frame"`
	TargetChainID                 string    `json:"target_chain_id"`
    TargetContractAddress         string    `json:"target_contract_address"`
    TargetFunction                string    `json:"target_function"`
    ABI                           string    `json:"abi"`
    ArgType                       int       `json:"arg_type"`
    Arguments                     []string  `json:"arguments"`
    DynamicArgumentsScriptUrl     string    `json:"dynamic_arguments_script_url"`
}

// Trigger data from schedulers to keepers for validation
type SendTriggerData struct {
	TaskID                        int64     `json:"task_id"`
    TaskDefinitionID              int       `json:"task_definition_id"`
	Timestamp                     time.Time `json:"timestamp"`

    EventChainId                  string    `json:"event_chain_id"`
    EventTxHash                   string    `json:"event_tx_hash"`
    EventTriggerContractAddress   string    `json:"event_trigger_contract_address"`
    EventTriggerFunction          string    `json:"event_trigger_function"`

    ConditionType                 string    `json:"condition_type"`
    ConditionSourceType           string    `json:"condition_source_type"`
    ConditionSourceUrl            string    `json:"condition_source_url"`
    ConditionUpperLimit           int       `json:"condition_upper_limit"`
    ConditionLowerLimit           int       `json:"condition_lower_limit"`
    ConditionSatisfiedValue       int       `json:"condition_satisfied_value"`
}
