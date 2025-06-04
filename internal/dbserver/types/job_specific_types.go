package types

import "time"

type TimeJobData struct {
	JobID                         int64     `json:"job_id"`
	ExpirationTime                time.Time `json:"expiration_time"`
	Recurring                     bool      `json:"recurring"`
	TimeInterval                  int64     `json:"time_interval"`
	ScheduleType                  string    `json:"schedule_type"`
	CronExpression                string    `json:"cron_expression"`
	SpecificSchedule              string    `json:"specific_schedule"`
	Timezone					  string	`json:"timezone"`
	NextExecutionTimestamp        time.Time `json:"next_execution_timestamp"`
	TargetChainID                 string    `json:"target_chain_id"`
	TargetContractAddress         string    `json:"target_contract_address"`
	TargetFunction                string    `json:"target_function"`
	ABI                           string    `json:"abi"`
	ArgType                       int       `json:"arg_type"`
	Arguments                     []string  `json:"arguments"`
	DynamicArgumentsScriptUrl string    `json:"dynamic_arguments_script_url"`
	IsCompleted                 bool      `json:"is_completed"`
	IsActive                    bool      `json:"is_active"`
}

type EventJobData struct {
	JobID                         int64    `json:"job_id"`
	ExpirationTime                time.Time `json:"expiration_time"`
	Recurring                     bool     `json:"recurring"`
	TriggerChainID                string   `json:"trigger_chain_id"`
	TriggerContractAddress        string   `json:"trigger_contract_address"`
	TriggerEvent                  string   `json:"trigger_event"`
	TargetChainID                 string   `json:"target_chain_id"`
	TargetContractAddress         string   `json:"target_contract_address"`
	TargetFunction                string   `json:"target_function"`
	ABI                           string   `json:"abi"`
	ArgType                       int      `json:"arg_type"`
	Arguments                     []string `json:"arguments"`
	DynamicArgumentsScriptUrl string   `json:"dynamic_arguments_script_url"`
	IsCompleted                 bool      `json:"is_completed"`
	IsActive                    bool      `json:"is_active"`
}

type ConditionJobData struct {
	JobID                         int64    `json:"job_id"`
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
	IsCompleted                 bool      `json:"is_completed"`
	IsActive 					bool      `json:"is_active"`
}