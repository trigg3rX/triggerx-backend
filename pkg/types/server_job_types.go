package types

import (
	"time"
)

// CreateJobData represents the request to create a new job
type CreateJobData struct {
	// Common fields for all job types
	JobID       string `json:"job_id" validate:"required"`
	UserAddress string `json:"user_address" validate:"required,ethereum_address"`
	EmailID     string `json:"email_id" validate:"required"`

	JobTitle          string  `json:"job_title" validate:"required,min=3,max=100"`
	TaskDefinitionID  int     `json:"task_definition_id" validate:"required,min=1,max=9"`
	Language          string  `json:"language,omitempty"`
	TimeFrame         int64   `json:"time_frame" validate:"required,min=1"`
	Recurring         bool    `json:"recurring"`
	JobCostPrediction float64 `json:"job_cost_prediction" validate:"required,min=0"`
	Timezone          string  `json:"timezone" validate:"required"`
	CreatedChainID    string  `json:"created_chain_id" validate:"required"`

	IsSafe      bool   `json:"is_safe"`
	SafeAddress string `json:"safe_address,omitempty" validate:"omitempty,ethereum_address"`
	SafeName    string `json:"safe_name,omitempty"`

	// Time job specific fields
	ScheduleType     string `json:"schedule_type,omitempty" validate:"omitempty,oneof=cron specific interval"`
	TimeInterval     int64  `json:"time_interval,omitempty" validate:"omitempty,min=1"`
	CronExpression   string `json:"cron_expression,omitempty" validate:"omitempty,cron"`
	SpecificSchedule string `json:"specific_schedule,omitempty" validate:"omitempty"`

	// Event job specific fields
	TriggerChainID         string `json:"trigger_chain_id,omitempty" validate:"omitempty"`
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

	// Target fields (common for all job types, nullable for agent jobs TDI 7,8,9)
	TargetChainID             string   `json:"target_chain_id,omitempty" validate:"omitempty"`
	TargetContractAddress     string   `json:"target_contract_address,omitempty" validate:"omitempty,ethereum_address"`
	TargetFunction            string   `json:"target_function,omitempty" validate:"omitempty"`
	ABI                       string   `json:"abi,omitempty" validate:"omitempty"`
	ArgType                   int      `json:"arg_type" validate:"required"`
	Arguments                 []string `json:"arguments" validate:"omitempty"`
	DynamicArgumentsScriptUrl string   `json:"dynamic_arguments_script_url,omitempty" validate:"omitempty,ipfs_url"`

	// Agent job fields (TDI 7, 8, 9)
	AgentScriptURL      string `json:"agent_script_url,omitempty" validate:"omitempty,ipfs_url"`
	AgentScriptLanguage string `json:"agent_script_language,omitempty" validate:"omitempty,oneof=ts go python javascript"`
	AgentTargetChainID  int    `json:"agent_target_chain_id,omitempty"`
	MaxExecutionTime    int    `json:"max_execution_time,omitempty"`
	ChallengePeriod     int64  `json:"challenge_period,omitempty"`

	IsImua bool `json:"is_imua"`
}

// CreateJobResponse represents the response after creating a job
type CreateJobResponse struct {
	UserPoints        string  `json:"user_points"`
	JobIDs            []string `json:"job_ids"`
	TaskDefinitionIDs []int    `json:"task_definition_ids"`
	TimeFrames        []int64  `json:"time_frames"`
}

// UpdateJobDataFromUserRequest represents the request to update job data from user
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

// UpdateJobLastExecutedAtRequest represents the request to update job last executed time
type UpdateJobLastExecutedAtRequest struct {
	JobID          string    `json:"job_id"`
	TaskIDs        int64     `json:"task_ids"`
	JobCostActual  float64   `json:"job_cost_actual"`
	LastExecutedAt time.Time `json:"last_executed_at"`
}

// JobDataAPI represents job data in API responses
type JobDataAPI struct {
	JobID             string    `json:"job_id"`
	JobTitle          string    `json:"job_title"`
	TaskDefinitionID  int       `json:"task_definition_id"`
	LinkJobID         string    `json:"link_job_id"`
	ChainStatus       int       `json:"chain_status"`
	TimeFrame         int64     `json:"time_frame"`
	Recurring         bool      `json:"recurring"`
	Status            string    `json:"status"`
	JobCostPrediction string    `json:"job_cost_prediction"` // Wei-based, string format
	JobCostActual     string    `json:"job_cost_actual"`     // Wei-based, string format
	TaskIDs           []int64   `json:"task_ids"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
	LastExecutedAt    time.Time `json:"last_executed_at"`
	Timezone          string    `json:"timezone"`
	IsImua            bool      `json:"is_imua"`
	CreatedChainID    string    `json:"created_chain_id"`
	SafeAddress       string    `json:"safe_address"`
	UserAddress       string    `json:"user_address"`
}

// TimeJobDataAPI represents time job data in API responses
type TimeJobDataAPI struct {
	JobID                     string    `json:"job_id"`
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
	AgentScriptURL            string    `json:"agent_script_url"`
	AgentScriptLanguage       string    `json:"agent_script_language"`
	AgentScriptHash           string    `json:"agent_script_hash"`
	AgentTargetChainID        int       `json:"agent_target_chain_id"`
	MaxExecutionTime          int       `json:"max_execution_time"`
	ChallengePeriod           int64     `json:"challenge_period"`
	IsActive                  bool      `json:"is_active"`        // Job is active (true on creation, false when expiration time is reached)
	LastExecutedAt            time.Time `json:"last_executed_at"` // Last execution time
	ExpirationTime            time.Time `json:"expiration_time"`  // Job expiration
}

// EventJobDataAPI represents event job data in API responses
type EventJobDataAPI struct {
	JobID                     string    `json:"job_id"`
	TaskDefinitionID          int       `json:"task_definition_id"`
	Recurring                 bool      `json:"recurring"`
	TriggerChainID            string    `json:"trigger_chain_id"`
	TriggerContractAddress    string    `json:"trigger_contract_address"`
	TriggerEvent              string    `json:"trigger_event"`
	EventFilterParaName       string    `json:"event_filter_para_name"`
	EventFilterValue          string    `json:"event_filter_value"`
	TargetChainID             string    `json:"target_chain_id"`
	TargetContractAddress     string    `json:"target_contract_address"`
	TargetFunction            string    `json:"target_function"`
	ABI                       string    `json:"abi"`
	ArgType                   int       `json:"arg_type"`
	Arguments                 []string  `json:"arguments"`
	DynamicArgumentsScriptUrl string    `json:"dynamic_arguments_script_url"`
	AgentScriptURL            string    `json:"agent_script_url"`
	AgentScriptLanguage       string    `json:"agent_script_language"`
	AgentScriptHash           string    `json:"agent_script_hash"`
	AgentTargetChainID        int       `json:"agent_target_chain_id"`
	MaxExecutionTime          int       `json:"max_execution_time"`
	ChallengePeriod           int64     `json:"challenge_period"`
	IsActive                  bool      `json:"is_active"`        // Job is active (true on creation, false when expiration time is reached)
	LastExecutedAt            time.Time `json:"last_executed_at"` // Last execution time
	ExpirationTime            time.Time `json:"expiration_time"`  // Job expiration
}

// ConditionJobDataAPI represents condition job data in API responses
type ConditionJobDataAPI struct {
	JobID                     string    `json:"job_id"`
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
	AgentScriptURL            string    `json:"agent_script_url"`
	AgentScriptLanguage       string    `json:"agent_script_language"`
	AgentScriptHash           string    `json:"agent_script_hash"`
	AgentTargetChainID        int       `json:"agent_target_chain_id"`
	MaxExecutionTime          int       `json:"max_execution_time"`
	ChallengePeriod           int64     `json:"challenge_period"`
	IsActive                  bool      `json:"is_active"`        // Job is active (true on creation, false when expiration time is reached)
	LastExecutedAt            time.Time `json:"last_executed_at"` // Last execution time
	ExpirationTime            time.Time `json:"expiration_time"`  // Job expiration
}

// JobResponse represents a complete job response with all job type data
type JobResponse struct {
	JobData          JobDataDTO           `json:"job_data"`
	TimeJobData      *TimeJobDataDTO      `json:"time_job_data,omitempty"`
	EventJobData     *EventJobDataDTO     `json:"event_job_data,omitempty"`
	ConditionJobData *ConditionJobDataDTO `json:"condition_job_data,omitempty"`
}

// JobResponseAPI represents a complete job response in API format
type JobResponseAPI struct {
	JobData          JobDataAPI           `json:"job_data"`
	TimeJobData      *TimeJobDataDTO      `json:"time_job_data,omitempty"`
	EventJobData     *EventJobDataDTO     `json:"event_job_data,omitempty"`
	ConditionJobData *ConditionJobDataDTO `json:"condition_job_data,omitempty"`
}

// ConvertJobResponseToAPI converts JobResponse to JobResponseAPI
func ConvertJobResponseToAPI(j JobResponse) JobResponseAPI {
	return JobResponseAPI{
		JobData: JobDataAPI{
			JobID:             j.JobData.JobID,
			JobTitle:          j.JobData.JobTitle,
			TaskDefinitionID:  j.JobData.TaskDefinitionID,
			LinkJobID:         j.JobData.LinkJobID,
			ChainStatus:       j.JobData.ChainStatus,
			TimeFrame:         j.JobData.TimeFrame,
			Recurring:         j.JobData.Recurring,
			Status:            j.JobData.Status,
			JobCostPrediction: j.JobData.JobCostPrediction,
			JobCostActual:     j.JobData.JobCostActual,
			TaskIDs:           j.JobData.TaskIDs,
			CreatedAt:         j.JobData.CreatedAt,
			UpdatedAt:         j.JobData.UpdatedAt,
			LastExecutedAt:    j.JobData.LastExecutedAt,
			Timezone:          j.JobData.Timezone,
			IsImua:            j.JobData.IsImua,
			CreatedChainID:    j.JobData.CreatedChainID,
			SafeAddress:       j.JobData.SafeAddress,
			UserAddress:       j.JobData.UserAddress,
		},
		TimeJobData:      j.TimeJobData,
		EventJobData:     j.EventJobData,
		ConditionJobData: j.ConditionJobData,
	}
}
