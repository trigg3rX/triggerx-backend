package types

import "time"

// CreateApiKeyRequest represents the request to create a new API key
// Owned by: dbserver HTTP API
type CreateApiKeyRequest struct {
	Owner     string `json:"owner" validate:"required,min=3,max=50"`
	RateLimit int    `json:"rate_limit" validate:"required,min=1,max=1000"`
}

// UpdateApiKeyRequest represents the request to update an API key
// Owned by: dbserver HTTP API
type UpdateApiKeyRequest struct {
	Key       string `json:"key"`
	IsActive  *bool  `json:"isActive,omitempty"`
	RateLimit *int   `json:"rateLimit,omitempty"`
}

// CreateUserDataRequest represents the request to create a new user
// Owned by: dbserver HTTP API
type CreateUserDataRequest struct {
	UserAddress string `json:"user_address"`
	EmailID     string `json:"email_id"`
	UserPoints  string `json:"user_points"`
}

// UserLeaderboardEntry represents a user leaderboard entry
// Owned by: dbserver HTTP API
type UserLeaderboardEntry struct {
	UserAddress string `json:"user_address"`
	TotalJobs   int64  `json:"total_jobs"`
	TotalTasks  int64  `json:"total_tasks"`
	UserPoints  string `json:"user_points"`
}

// KeeperLeaderboardEntry represents a keeper leaderboard entry
// Owned by: dbserver HTTP API
type KeeperLeaderboardEntry struct {
	KeeperAddress   string `json:"keeper_address"`
	KeeperName      string `json:"keeper_name"`
	NoExecutedTasks int64  `json:"no_executed_tasks"`
	NoAttestedTasks int64  `json:"no_attested_tasks"`
	KeeperPoints    string `json:"keeper_points"`
}


// CreateJobData represents the request to create a new job
// Owned by: dbserver HTTP API
type CreateJobData struct {
	// Common fields for all job types
	JobID       string `json:"job_id" validate:"required"`
	UserAddress string `json:"user_address" validate:"required,ethereum_address"`
	EmailID     string `json:"email_id"` // validation: required for TDI 3,4,5,6,8,9 when recurring is set to true (handled in struct-level validation)

	JobTitle          string  `json:"job_title" validate:"required,min=3,max=100"`
	JobType           string  `json:"job_type" validate:"required,oneof=frontend sdk template"`
	TaskDefinitionID  int     `json:"task_definition_id" validate:"required,min=1,max=9"`
	TimeFrame         int64   `json:"time_frame" validate:"required,min=31,max=604800"`
	Recurring         bool    `json:"recurring"` // defaults to false if not provided
	JobCostPrediction string  `json:"job_cost_prediction" validate:"required,numeric"`
	Timezone          string  `json:"timezone" validate:"required,timezone"`
	CreatedChainID    string  `json:"created_chain_id" validate:"required,oneof=42161 421614 11155111 11155420 84532 imua"`

	IsSafe      bool   `json:"is_safe"` // defaults to false if not provided
	SafeAddress string `json:"safe_address,omitempty" validate:"omitempty,ethereum_address"` // validation: required when is_safe is true (handled in struct-level validation)
	SafeName    string `json:"safe_name,omitempty" validate:"omitempty,min=3,max=50"` // validation: required when is_safe is true (handled in struct-level validation)

	// Time job specific fields
	ScheduleType     string `json:"schedule_type,omitempty" validate:"omitempty,oneof=cron specific interval"`
	TimeInterval     int64  `json:"time_interval,omitempty" validate:"omitempty,min=30"`
	CronExpression   string `json:"cron_expression,omitempty" validate:"omitempty,cron"`
	SpecificSchedule string `json:"specific_schedule,omitempty" validate:"omitempty"`

	// Event job specific fields
	TriggerChainID         string `json:"trigger_chain_id,omitempty" validate:"omitempty,oneof=1 10 8453 42161 11155111 11155420 84532 421614 imua"`
	TriggerContractAddress string `json:"trigger_contract_address,omitempty" validate:"omitempty,ethereum_address"`
	TriggerEvent           string `json:"trigger_event,omitempty" validate:"omitempty"`
	EventFilterParaName    string `json:"event_filter_para_name,omitempty" validate:"omitempty"`
	EventFilterValue       string `json:"event_filter_value,omitempty" validate:"omitempty"`

	// Condition job specific fields
	ConditionType    string  `json:"condition_type,omitempty" validate:"omitempty,oneof=greater_than less_than between equals not_equals greater_equal less_equal"`
	UpperLimit       float64 `json:"upper_limit,omitempty" validate:"omitempty"`
	LowerLimit       float64 `json:"lower_limit,omitempty" validate:"omitempty"`
	ValueSourceType  string  `json:"value_source_type,omitempty" validate:"omitempty,oneof=api oracle websocket"`
	ValueSourceUrl   string  `json:"value_source_url,omitempty" validate:"omitempty,url"`
	SelectedKeyRoute string  `json:"selected_key_route,omitempty" validate:"omitempty"`

	// Target fields (common for all job types, nullable for agent jobs TDI 7,8,9)
	TargetChainID             string   `json:"target_chain_id,omitempty" validate:"omitempty,oneof=42161 421614 11155111 11155420 84532 imua"`
	TargetContractAddress     string   `json:"target_contract_address,omitempty" validate:"omitempty,ethereum_address"`
	TargetFunction            string   `json:"target_function,omitempty" validate:"omitempty"`
	ABI                       string   `json:"abi,omitempty" validate:"omitempty"`
	ArgType                   int      `json:"arg_type" validate:"required,oneof=0 1 2"`
	Arguments                 []string `json:"arguments" validate:"omitempty"`
	ExecutionScriptURL      string `json:"execution_script_url,omitempty" validate:"omitempty,ipfs_url"`
	ExecutionScriptLanguage string `json:"execution_script_language,omitempty" validate:"omitempty,oneof=ts go"`
	ExecutionScriptHash     string `json:"execution_script_hash,omitempty" validate:"omitempty"`
	MaxExecutionTime        int    `json:"max_execution_time,omitempty" validate:"omitempty,max=59"`
	ChallengePeriod         int64  `json:"challenge_period,omitempty" validate:"omitempty,max=21600"`
}

// CreateJobResponse represents the response after creating a job
// Owned by: dbserver HTTP API
type CreateJobResponse struct {
	UserPoints        string   `json:"user_points"`
	JobIDs            []string `json:"job_ids"`
	TaskDefinitionIDs []int    `json:"task_definition_ids"`
	TimeFrames        []int64  `json:"time_frames"`
}

// UpdateJobDataFromUserRequest represents the request to update job data from user
// Owned by: dbserver HTTP API
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

// JobResponseAPI represents a complete job response in API format
// Owned by: dbserver HTTP API
type GetJobDataResponse struct {
	JobData          JobDataDTO           `json:"job_data"`
	TimeJobData      *TimeJobDataDTO      `json:"time_job_data,omitempty"`
	EventJobData     *EventJobDataDTO     `json:"event_job_data,omitempty"`
	ConditionJobData *ConditionJobDataDTO `json:"condition_job_data,omitempty"`
}

// TasksByJobIDResponse represents task data grouped by job ID
// Owned by: dbserver HTTP API
type TasksByJobIDResponse struct {
	TaskID                int64     `json:"task_id"`
	TaskNumber            int64     `json:"task_number"`
	TaskStatus            string    `json:"task_status"`
	TaskError             string    `json:"task_error"`
	CreatedAt             time.Time `json:"created_at"`
	ExecutedAt            time.Time `json:"executed_at"`
	SubmittedAt           time.Time `json:"submitted_at"`
	TaskOpxPredictedCost  string    `json:"task_opx_predicted_cost"`
	TaskOpxActualCost     string    `json:"task_opx_actual_cost"`
	ExecutionTxHash       string    `json:"execution_tx_hash"`
	ConvertedArguments    []string  `json:"converted_arguments"`	
	TaskPerformerAddress  []string  `json:"task_performer_address"`
	TaskAttesterAddress   []string  `json:"task_attester_address"`
	IsSuccessful          bool      `json:"is_successful"`
	IsAccepted            bool      `json:"is_accepted"`
	TxURL                 string    `json:"tx_url"`
}

// TasksByJobGroupResponse groups task data by job identifier
// Owned by: dbserver HTTP API
type TasksByJobGroupResponse struct {
	JobID string                 `json:"job_id"`
	Tasks []TasksByJobIDResponse `json:"tasks"`
}

// RecentTaskResponse represents a task in the recent tasks list for the landing page
// Owned by: dbserver HTTP API
type RecentTaskResponse struct {
	TaskID                int64     `json:"task_id"`
	TaskNumber            int64     `json:"task_number"`
	JobID                 string    `json:"job_id"`
	TaskDefinitionID      int       `json:"task_definition_id"`
	CreatedAt             time.Time `json:"created_at"`
	TaskOpXCost           float64   `json:"task_opx_cost"`
	ExecutionTimestamp    time.Time `json:"execution_timestamp"`
	ExecutionTxHash       string    `json:"execution_tx_hash"`
	TxURL                 string    `json:"tx_url"`
	TaskPerformerAddress  string    `json:"task_performer_address"`
	TaskAttesterAddresses []string  `json:"task_attester_addresses"`
	TaskStatus            string    `json:"task_status"`
	TaskError             string    `json:"task_error"`
	Network               string    `json:"network"`
}
