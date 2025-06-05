package types

import "time"

type ScheduleTimeJobData struct {
	JobID                         int64     `json:"job_id"`
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

type ScheduleEventJobData struct {
	
}

type ScheduleConditionJobData struct {
	
}