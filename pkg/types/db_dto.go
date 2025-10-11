package types

import "time"

// UserDataDTO represents the user data transfer object
type UserDataDTO struct {
	UserAddress   string    `json:"user_address"`
	EmailID       string    `json:"email_id"`
	JobIDs        []string  `json:"job_ids"`
	UserPoints    string    `json:"user_points"`
	TotalJobs     int64     `json:"total_jobs"`
	TotalTasks    int64     `json:"total_tasks"`
	CreatedAt     time.Time `json:"created_at"`
	LastUpdatedAt time.Time `json:"last_updated_at"`
}

// JobDataDTO represents the job data transfer object
type JobDataDTO struct {
	JobID             string    `json:"job_id"`
	JobTitle          string    `json:"job_title"`
	TaskDefinitionID  int       `json:"task_definition_id"`
	CreatedChainID    string    `json:"created_chain_id"`
	UserAddress       string    `json:"user_address"`
	LinkJobID         string    `json:"link_job_id"`
	ChainStatus       int       `json:"chain_status"`
	Timezone          string    `json:"timezone"`
	IsImua            bool      `json:"is_imua"`
	JobType           string    `json:"job_type"`
	TimeFrame         int64     `json:"time_frame"`
	Recurring         bool      `json:"recurring"`
	Status            string    `json:"status"`
	JobCostPrediction string    `json:"job_cost_prediction"`
	JobCostActual     string    `json:"job_cost_actual"`
	TaskIDs           []int64   `json:"task_ids"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
	LastExecutedAt    time.Time `json:"last_executed_at"`
}

// TimeJobDataDTO represents the time job data transfer object
type TimeJobDataDTO struct {
	JobID                     string    `json:"job_id"`
	TaskDefinitionID          int       `json:"task_definition_id"`
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
	DynamicArgumentsScriptURL string    `json:"dynamic_arguments_script_url"`
	IsCompleted               bool      `json:"is_completed"`
	LastExecutedAt            time.Time `json:"last_executed_at"`
	ExpirationTime            time.Time `json:"expiration_time"`
}

// EventJobDataDTO represents the event job data transfer object
type EventJobDataDTO struct {
	JobID                      string    `json:"job_id"`
	TaskDefinitionID           int       `json:"task_definition_id"`
	Recurring                  bool      `json:"recurring"`
	TriggerChainID             string    `json:"trigger_chain_id"`
	TriggerContractAddress     string    `json:"trigger_contract_address"`
	TriggerEvent               string    `json:"trigger_event"`
	TriggerEventFilterParaName string    `json:"trigger_event_filter_para_name"`
	TriggerEventFilterValue    string    `json:"trigger_event_filter_value"`
	TargetChainID              string    `json:"target_chain_id"`
	TargetContractAddress      string    `json:"target_contract_address"`
	TargetFunction             string    `json:"target_function"`
	ABI                        string    `json:"abi"`
	ArgType                    int       `json:"arg_type"`
	Arguments                  []string  `json:"arguments"`
	DynamicArgumentsScriptURL  string    `json:"dynamic_arguments_script_url"`
	IsCompleted                bool      `json:"is_completed"`
	LastExecutedAt             time.Time `json:"last_executed_at"`
	ExpirationTime             time.Time `json:"expiration_time"`
}

// ConditionJobDataDTO represents the condition job data transfer object
type ConditionJobDataDTO struct {
	JobID                     string    `json:"job_id"`
	TaskDefinitionID          int       `json:"task_definition_id"`
	Recurring                 bool      `json:"recurring"`
	ConditionType             string    `json:"condition_type"`
	UpperLimit                float64   `json:"upper_limit"`
	LowerLimit                float64   `json:"lower_limit"`
	ValueSourceType           string    `json:"value_source_type"`
	ValueSourceURL            string    `json:"value_source_url"`
	SelectedKeyRoute          string    `json:"selected_key_route"`
	TargetChainID             string    `json:"target_chain_id"`
	TargetContractAddress     string    `json:"target_contract_address"`
	TargetFunction            string    `json:"target_function"`
	ABI                       string    `json:"abi"`
	ArgType                   int       `json:"arg_type"`
	Arguments                 []string  `json:"arguments"`
	DynamicArgumentsScriptURL string    `json:"dynamic_arguments_script_url"`
	IsCompleted               bool      `json:"is_completed"`
	LastExecutedAt            time.Time `json:"last_executed_at"`
	ExpirationTime            time.Time `json:"expiration_time"`
}

// TaskDataDTO represents the task data transfer object
type TaskDataDTO struct {
	TaskID               int64     `json:"task_id"`
	TaskNumber           int64     `json:"task_number"`
	JobID                string    `json:"job_id"`
	TaskDefinitionID     int       `json:"task_definition_id"`
	CreatedAt            time.Time `json:"created_at"`
	TaskOpxPredictedCost string    `json:"task_opx_predicted_cost"`
	TaskOpxActualCost    string    `json:"task_opx_actual_cost"`
	ExecutionTimestamp   time.Time `json:"execution_timestamp"`
	ExecutionTxHash      string    `json:"execution_tx_hash"`
	TaskPerformerID      int64     `json:"task_performer_id"`
	TaskAttesterIDs      []int64   `json:"task_attester_ids"`
	ConvertedArguments   []interface{} `json:"converted_arguments"`
	ProofOfTask          string    `json:"proof_of_task"`
	SubmissionTxHash     string    `json:"submission_tx_hash"`
	IsSuccessful         bool      `json:"is_successful"`
	IsAccepted           bool      `json:"is_accepted"`
	IsImua               bool      `json:"is_imua"`
}

// KeeperDataDTO represents the keeper data transfer object
type KeeperDataDTO struct {
	KeeperName       string    `json:"keeper_name"`
	KeeperAddress    string    `json:"keeper_address"`
	RewardsAddress   string    `json:"rewards_address"`
	ConsensusAddress string    `json:"consensus_address"`
	RegisteredTx     string    `json:"registered_tx"`
	OperatorID       int64     `json:"operator_id"`
	VotingPower      string    `json:"voting_power"`
	Whitelisted      bool      `json:"whitelisted"`
	Registered       bool      `json:"registered"`
	Online           bool      `json:"online"`
	Version          string    `json:"version"`
	OnImua           bool      `json:"on_imua"`
	PublicIP         string    `json:"public_ip"`
	ChatID           int64     `json:"chat_id"`
	EmailID          string    `json:"email_id"`
	RewardsBooster   string    `json:"rewards_booster"`
	NoExecutedTasks  int64     `json:"no_executed_tasks"`
	NoAttestedTasks  int64     `json:"no_attested_tasks"`
	Uptime           int64     `json:"uptime"`
	KeeperPoints     string    `json:"keeper_points"`
	LastCheckedIn    time.Time `json:"last_checked_in"`
}

// ApiKeyDataDTO represents the API key data transfer object
type ApiKeyDataDTO struct {
	Key          string    `json:"key"`
	Owner        string    `json:"owner"`
	IsActive     bool      `json:"is_active"`
	RateLimit    int       `json:"rate_limit"`
	SuccessCount int64     `json:"success_count"`
	FailedCount  int64     `json:"failed_count"`
	LastUsed     time.Time `json:"last_used"`
	CreatedAt    time.Time `json:"created_at"`
}

// CompleteJobDataDTO represents the complete job data transfer object,
// where only the relevant job type data is populated (others are nil).
type CompleteJobDataDTO struct {
	JobDataDTO           JobDataDTO           `json:"job_data"`
	TimeJobDataDTO       *TimeJobDataDTO      `json:"time_job_data,omitempty"`
	EventJobDataDTO      *EventJobDataDTO     `json:"event_job_data,omitempty"`
	ConditionJobDataDTO  *ConditionJobDataDTO `json:"condition_job_data,omitempty"`
}