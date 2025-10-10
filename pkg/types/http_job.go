package types

import (
	"time"
)

type CreateJobDataRequest struct {
	// Common fields for all job types
	JobID             string `json:"job_id" validate:"required"`
	JobTitle          string `json:"job_title" validate:"required,min=3,max=100"`
	TaskDefinitionID  int    `json:"task_definition_id" validate:"required,min=1,max=6"`
	CreatedChainID    string `json:"created_chain_id" validate:"required,chain_id"`
	UserAddress       string `json:"user_address" validate:"required,ethereum_address"`
	Timezone          string `json:"timezone" validate:"required"`
	IsImua            bool   `json:"is_imua"`
	JobType           string `json:"job_type" validate:"required,oneof=sdk frontend contract template"`
	TimeFrame         int64  `json:"time_frame" validate:"required,min=1,max=2592000"`
	Recurring         bool   `json:"recurring"`
	JobCostPrediction string `json:"job_cost_prediction" validate:"required,min=0"`

	// Time job specific fields
	ScheduleType     string `json:"schedule_type,omitempty" validate:"omitempty,oneof=cron specific interval"`
	TimeInterval     int64  `json:"time_interval,omitempty" validate:"omitempty,min=1"`
	CronExpression   string `json:"cron_expression,omitempty" validate:"omitempty,cron"`
	SpecificSchedule string `json:"specific_schedule,omitempty" validate:"omitempty"`

	// Event job specific fields
	TriggerChainID             string `json:"trigger_chain_id,omitempty" validate:"omitempty,chain_id"`
	TriggerContractAddress     string `json:"trigger_contract_address,omitempty" validate:"omitempty,ethereum_address"`
	TriggerEvent               string `json:"trigger_event,omitempty" validate:"omitempty"`
	TriggerEventFilterParaName string `json:"trigger_event_filter_para_name,omitempty" validate:"omitempty"`
	TriggerEventFilterValue    string `json:"trigger_event_filter_value,omitempty" validate:"omitempty"`

	// Condition job specific fields
	ConditionType    string  `json:"condition_type,omitempty" validate:"omitempty"`
	UpperLimit       float64 `json:"upper_limit,omitempty" validate:"omitempty,gt=0"`
	LowerLimit       float64 `json:"lower_limit,omitempty" validate:"omitempty,gt=0"`
	ValueSourceType  string  `json:"value_source_type,omitempty" validate:"omitempty"`
	ValueSourceUrl   string  `json:"value_source_url,omitempty" validate:"omitempty"`
	SelectedKeyRoute string  `json:"selected_key_route,omitempty" validate:"omitempty"`

	// Target fields (common for all job types)
	TargetChainID             string   `json:"target_chain_id" validate:"required,chain_id"`
	TargetContractAddress     string   `json:"target_contract_address" validate:"required,ethereum_address"`
	TargetFunction            string   `json:"target_function" validate:"required"`
	ABI                       string   `json:"abi" validate:"required"`
	ArgType                   int      `json:"arg_type" validate:"required"`
	Arguments                 []string `json:"arguments" validate:"omitempty"`
	DynamicArgumentsScriptUrl string   `json:"dynamic_arguments_script_url,omitempty" validate:"omitempty,ipfs_url"`
}

type CreateJobResponse struct {
	JobIDs            []string `json:"job_ids"`
	TaskDefinitionIDs []int    `json:"task_definition_ids"`
	TimeFrames        []int64  `json:"time_frames"`
}

type UpdateJobDataFromUserRequest struct {
	JobID             string  `json:"job_id"`
	JobTitle          string  `json:"job_title"`
	Recurring         bool    `json:"recurring"`
	Status            string  `json:"status"`
	TimeFrame         int64   `json:"time_frame"`
	JobCostPrediction string  `json:"job_cost_prediction"`
	Timezone          string  `json:"timezone"`
	TimeInterval      int64   `json:"time_interval"`
}

type UpdateJobLastExecutedAtRequest struct {
	JobID          string    `json:"job_id"`
	TaskIDs        int64     `json:"task_ids"`
	JobCostActual  string    `json:"job_cost_actual"`
	LastExecutedAt time.Time `json:"last_executed_at"`
}

// TaskFeeResponse represents the response structure for task fees by job ID
type GetTaskFeesByJobIDResponse struct {
	TaskID                int64   `json:"task_id"`
	TaskOpxPredictedCost  string  `json:"task_opx_predicted_cost"`
	TaskOpxActualCost     string  `json:"task_opx_actual_cost"`
}
