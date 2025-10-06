package types

import (
	"math/big"
	"time"
)

type CreateJobData struct {
	// Common fields for all job types
	JobID        string   `json:"job_id" validate:"required"`
	UserAddress  string   `json:"user_address" validate:"required,ethereum_address"`
	EtherBalance *big.Int `json:"ether_balance" validate:"required"`
	TokenBalance *big.Int `json:"token_balance" validate:"required"`

	JobTitle          string  `json:"job_title" validate:"required,min=3,max=100"`
	TaskDefinitionID  int     `json:"task_definition_id" validate:"required,min=1,max=6"`
	Custom            bool    `json:"custom"`
	TimeFrame         int64   `json:"time_frame" validate:"required,min=1"`
	Recurring         bool    `json:"recurring"`
	JobCostPrediction float64 `json:"job_cost_prediction" validate:"required,min=0"`
	Timezone          string  `json:"timezone" validate:"required"`
	CreatedChainID    string  `json:"created_chain_id" validate:"required,chain_id"`

	// Time job specific fields
	ScheduleType     string `json:"schedule_type,omitempty" validate:"omitempty,oneof=cron specific interval"`
	TimeInterval     int64  `json:"time_interval,omitempty" validate:"omitempty,min=1"`
	CronExpression   string `json:"cron_expression,omitempty" validate:"omitempty,cron"`
	SpecificSchedule string `json:"specific_schedule,omitempty" validate:"omitempty"`

	// Event job specific fields
	TriggerChainID         string `json:"trigger_chain_id,omitempty" validate:"omitempty,chain_id"`
	TriggerContractAddress string `json:"trigger_contract_address,omitempty" validate:"omitempty,ethereum_address"`
	TriggerEvent           string `json:"trigger_event,omitempty" validate:"omitempty"`
	EventFilterParaName    string `json:"event_filter_para_name,omitempty" validate:"omitempty"`
	EventFilterValue       string `json:"event_filter_value,omitempty" validate:"omitempty"`

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

	IsImua bool `json:"is_imua"`
}

type CreateJobResponse struct {
	UserID            int64      `json:"user_id"`
	AccountBalance    *big.Int   `json:"account_balance"`
	TokenBalance      *big.Int   `json:"token_balance"`
	JobIDs            []*big.Int `json:"job_ids"`
	TaskDefinitionIDs []int      `json:"task_definition_ids"`
	TimeFrames        []int64    `json:"time_frames"`
}

type UpdateJobDataFromUserRequest struct {
	JobID             string  `json:"job_id"`
	JobTitle          string  `json:"job_title"`
	Recurring         bool    `json:"recurring"`
	Status            string  `json:"status"`
	TimeFrame         int64   `json:"time_frame"`
	JobCostPrediction float64 `json:"job_cost_prediction"`
	Timezone          string  `json:"timezone"`
	TimeInterval      int64   `json:"time_interval"`
}

type UpdateJobLastExecutedAtRequest struct {
	JobID          *big.Int  `json:"job_id"`
	TaskIDs        int64     `json:"task_ids"`
	JobCostActual  float64   `json:"job_cost_actual"`
	LastExecutedAt time.Time `json:"last_executed_at"`
}

// TaskFeeResponse represents the response structure for task fees by job
// TaskOpxCost corresponds to the task_opx_cost field in the database
type TaskFeeResponse struct {
	TaskID      int64   `json:"task_id"`
	TaskOpxCost float64 `json:"task_opx_cost"`
}

type JobDataAPI struct {
	JobID             *big.Int  `json:"job_id"`
	JobTitle          string    `json:"job_title"`
	TaskDefinitionID  int       `json:"task_definition_id"`
	UserID            int64     `json:"user_id"`
	LinkJobID         *big.Int  `json:"link_job_id"`
	ChainStatus       int       `json:"chain_status"`
	Custom            bool      `json:"custom"`
	TimeFrame         int64     `json:"time_frame"`
	Recurring         bool      `json:"recurring"`
	Status            string    `json:"status"`
	JobCostPrediction float64   `json:"job_cost_prediction"`
	JobCostActual     float64   `json:"job_cost_actual"`
	TaskIDs           []int64   `json:"task_ids"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
	LastExecutedAt    time.Time `json:"last_executed_at"`
	Timezone          string    `json:"timezone"`
	IsImua            bool      `json:"is_imua"`
	CreatedChainID    string    `json:"created_chain_id"`
	JobType           string    `json:"job_type"`
}

type TimeJobDataAPI struct {
	JobID                     *big.Int  `json:"job_id"`
	TaskDefinitionID          int       `json:"task_definition_id"`
	Timezone                  string    `json:"timezone"`
	ScheduleType              string    `json:"schedule_type"`
	TimeInterval              int64     `json:"time_interval"`
	CronExpression            string    `json:"cron_expression"`
	SpecificSchedule          string    `json:"specific_schedule"`
	NextExecutionTimestamp    time.Time `json:"next_execution_timestamp"`
	TargetChainID             string    `json:"target_chain_id"`
	TargetContractAddress     string    `json:"target_contract_address"`
	TargetFunction            string    `json:"target_function"`
	ABI                       string    `json:"abi"`
	ArgType                   int       `json:"arg_type"`
	Arguments                 []string  `json:"arguments"`
	DynamicArgumentsScriptUrl string    `json:"dynamic_arguments_script_url"`
	IsCompleted               bool      `json:"is_completed"`
	IsActive                  bool      `json:"is_active"`
	CreatedAt                 time.Time `json:"created_at"`
	UpdatedAt                 time.Time `json:"updated_at"`
	LastExecutedAt            time.Time `json:"last_executed_at"`
	ExpirationTime            time.Time `json:"expiration_time"`
}

type EventJobDataAPI struct {
	JobID                     *big.Int  `json:"job_id"`
	TaskDefinitionID          int       `json:"task_definition_id"`
	Recurring                 bool      `json:"recurring"`
	TriggerChainID            string    `json:"trigger_chain_id"`
	TriggerContractAddress    string    `json:"trigger_contract_address"`
	TriggerEvent              string    `json:"trigger_event"`
	EventFilterParaName    string `json:"event_filter_para_name"`
	EventFilterValue       string `json:"event_filter_value"`
	TargetChainID             string    `json:"target_chain_id"`
	TargetContractAddress     string    `json:"target_contract_address"`
	TargetFunction            string    `json:"target_function"`
	ABI                       string    `json:"abi"`
	ArgType                   int       `json:"arg_type"`
	Arguments                 []string  `json:"arguments"`
	DynamicArgumentsScriptUrl string    `json:"dynamic_arguments_script_url"`
	IsCompleted               bool      `json:"is_completed"`
	IsActive                  bool      `json:"is_active"`
	CreatedAt                 time.Time `json:"created_at"`
	UpdatedAt                 time.Time `json:"updated_at"`
	LastExecutedAt            time.Time `json:"last_executed_at"`
	ExpirationTime            time.Time `json:"expiration_time"`
}

type ConditionJobDataAPI struct {
	JobID                     *big.Int  `json:"job_id"`
	TaskDefinitionID          int       `json:"task_definition_id"`
	Recurring                 bool      `json:"recurring"`
	ConditionType             string    `json:"condition_type"`
	UpperLimit                float64   `json:"upper_limit"`
	LowerLimit                float64   `json:"lower_limit"`
	ValueSourceType           string    `json:"value_source_type"`
	ValueSourceUrl            string    `json:"value_source_url"`
	SelectedKeyRoute          string    `json:"selected_key_route"`
	TargetChainID             string    `json:"target_chain_id"`
	TargetContractAddress     string    `json:"target_contract_address"`
	TargetFunction            string    `json:"target_function"`
	ABI                       string    `json:"abi"`
	ArgType                   int       `json:"arg_type"`
	Arguments                 []string  `json:"arguments"`
	DynamicArgumentsScriptUrl string    `json:"dynamic_arguments_script_url"`
	IsCompleted               bool      `json:"is_completed"`
	IsActive                  bool      `json:"is_active"`
	CreatedAt                 time.Time `json:"created_at"`
	UpdatedAt                 time.Time `json:"updated_at"`
	LastExecutedAt            time.Time `json:"last_executed_at"`
	ExpirationTime            time.Time `json:"expiration_time"`
}

type JobResponse struct {
	JobData          JobDataAPI           `json:"job_data"`
	TimeJobData      *TimeJobDataAPI      `json:"time_job_data,omitempty"`
	EventJobData     *EventJobDataAPI     `json:"event_job_data,omitempty"`
	ConditionJobData *ConditionJobDataAPI `json:"condition_job_data,omitempty"`
}

type JobResponseAPI struct {
	JobData          JobDataAPI           `json:"job_data"`
	TimeJobData      *TimeJobDataAPI      `json:"time_job_data,omitempty"`
	EventJobData     *EventJobDataAPI     `json:"event_job_data,omitempty"`
	ConditionJobData *ConditionJobDataAPI `json:"condition_job_data,omitempty"`
}

func ConvertJobResponseToAPI(j JobResponse) JobResponseAPI {
	costPrediction := j.JobData.JobCostPrediction
	costActual := j.JobData.JobCostActual

	return JobResponseAPI{
		JobData: JobDataAPI{
			JobID:             j.JobData.JobID,
			JobTitle:          j.JobData.JobTitle,
			TaskDefinitionID:  j.JobData.TaskDefinitionID,
			UserID:            j.JobData.UserID,
			LinkJobID:         j.JobData.LinkJobID,
			ChainStatus:       j.JobData.ChainStatus,
			JobType:           j.JobData.JobType,
			TimeFrame:         j.JobData.TimeFrame,
			Recurring:         j.JobData.Recurring,
			Status:            j.JobData.Status,
			JobCostPrediction: costPrediction,
			JobCostActual:     costActual,
			TaskIDs:           j.JobData.TaskIDs,
			CreatedAt:         j.JobData.CreatedAt,
			UpdatedAt:         j.JobData.UpdatedAt,
			LastExecutedAt:    j.JobData.LastExecutedAt,
			Timezone:          j.JobData.Timezone,
			IsImua:            j.JobData.IsImua,
			CreatedChainID:    j.JobData.CreatedChainID,
		},
		TimeJobData:      j.TimeJobData,
		EventJobData:     j.EventJobData,
		ConditionJobData: j.ConditionJobData,
	}
}
