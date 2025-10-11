package types

import "time"

// UpdateUserEmailRequest is the request to update the email for the given address
type UpdateUserEmailRequest struct {
	UserAddress string `json:"user_address" validate:"required,eth_addr"`
	Email       string `json:"email" validate:"omitempty,email"`
}

// CreateJobDataRequest is the request to create a new job
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

// CreateJobResponse is the response to create a new job
type CreateJobResponse struct {
	JobIDs            []string `json:"job_ids"`
	TaskDefinitionIDs []int    `json:"task_definition_ids"`
	TimeFrames        []int64  `json:"time_frames"`
}

// UpdateJobDataFromUserRequest is the request to update a job data from user
type UpdateJobDataFromUserRequest struct {
	JobID             string  `json:"job_id"`
	JobTitle          string  `json:"job_title"`
	Recurring         bool    `json:"recurring"`
	TimeFrame         int64   `json:"time_frame"`
	JobCostPrediction string  `json:"job_cost_prediction"`
	TimeInterval      int64   `json:"time_interval"`
}

// GetTasksByJobIDResponse is the response to get the tasks for a job from dbserver
type GetTasksByJobIDResponse struct {
	TaskNumber           int64     `json:"task_number"`
	TaskOpXPredictedCost string    `json:"task_opx_predicted_cost"`
	TaskOpXActualCost    string    `json:"task_opx_actual_cost"`
	ExecutionTimestamp   time.Time `json:"execution_timestamp"`
	ExecutionTxHash      string    `json:"execution_tx_hash"`
	TaskPerformerID      int64     `json:"task_performer_id"`
	TaskAttesterIDs      []int64   `json:"task_attester_ids"`
	ConvertedArguments   []string  `json:"converted_arguments"`
	IsSuccessful         bool      `json:"is_successful"`
	IsAccepted           bool      `json:"is_accepted"`
	TxURL                string    `json:"tx_url"`
}

// Create New Keeper from Google Form (google script)
type CreateKeeperDataFromGoogleFormRequest struct {
	KeeperAddress  string `json:"keeper_address" validate:"required,eth_addr"`
	RewardsAddress string `json:"rewards_address"`
	KeeperName     string `json:"keeper_name" validate:"required,min=3,max=50"`
	EmailID        string `json:"email_id" validate:"required,email"`
	OnImua         bool   `json:"on_imua"`
}

// CreateApiKeyRequest is the request to create a new API key
type CreateApiKeyRequest struct {
	Owner     string `json:"owner" validate:"required,eth_addr"`
	RateLimit int    `json:"rate_limit" validate:"required,min=1,max=1000"`
}

// CreateApiKeyResponse is the response to create a new API key
type CreateApiKeyResponse struct {
	Key       string    `json:"key"`
	Owner     string    `json:"owner"`
	IsActive  bool      `json:"is_active"`
	LastUsed  time.Time `json:"last_used"`
	CreatedAt time.Time `json:"created_at"`
}

// DeleteApiKeyRequest is the request to delete an API key
type DeleteApiKeyRequest struct {
	Key string `json:"key" validate:"required"`
	Owner string `json:"owner" validate:"required,eth_addr"`
}

// GetApiKeyDataResponse is the response to get the data of an API key
type GetApiKeyDataResponse struct {
	Key       string    `json:"key"`
	Owner     string    `json:"owner"`
	IsActive  bool      `json:"is_active"`
	SuccessCount int64   `json:"success_count"`
	FailedCount  int64   `json:"failed_count"`
	LastUsed  time.Time `json:"last_used"`
	CreatedAt time.Time `json:"created_at"`
}

// KeeperLeaderboardEntry is the entry for the keeper leaderboard
type KeeperLeaderboardEntry struct {
	KeeperAddress   string  `json:"keeper_address"`
	OperatorID      int64   `json:"operator_id"`
	KeeperName      string  `json:"keeper_name"`
	NoExecutedTasks int64   `json:"no_executed_tasks"`
	NoAttestedTasks int64   `json:"no_attested_tasks"`
	KeeperPoints    string  `json:"keeper_points"`
	OnImua          bool    `json:"on_imua"`
}

// UserLeaderboardEntry is the entry for the user leaderboard
type UserLeaderboardEntry struct {
	UserAddress string  `json:"user_address"`
	TotalJobs   int64   `json:"total_jobs"`
	TotalTasks  int64   `json:"total_tasks"`
	UserPoints  string  `json:"user_points"`
}

// DEPRECATED: To be removed
type CreateTaskDataRequest struct {
	JobID            string   `json:"job_id" validate:"required"`
	TaskDefinitionID int      `json:"task_definition_id" validate:"required"`
	IsImua           bool     `json:"is_imua"`
}
type UpdateTaskExecutionDataRequest struct {
	TaskID             int64     `json:"task_id" validate:"required"`
	TaskPerformerID    int64     `json:"task_performer_id" validate:"required"`
	ExecutionTimestamp time.Time `json:"execution_timestamp" validate:"required"`
	ExecutionTxHash    string    `json:"execution_tx_hash" validate:"required"`
	ProofOfTask        string    `json:"proof_of_task" validate:"required"`
	TaskOpXCost        float64   `json:"task_opx_cost" validate:"required"`
}