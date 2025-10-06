package types

import (
	"math/big"
	"time"
)

// UserDataEntity represents the user_data table structure
type UserDataEntity struct {
	UserID        int64     `cql:"user_id"`
	UserAddress   string    `cql:"user_address"`
	EmailID       string    `cql:"email_id"`
	JobIDs        []big.Int `cql:"job_ids"`
	OpxConsumed   big.Int   `cql:"opx_consumed"`
	TotalJobs     int64     `cql:"total_jobs"`
	TotalTasks    int64     `cql:"total_tasks"`
	CreatedAt     time.Time `cql:"created_at"`
	LastUpdatedAt time.Time `cql:"last_updated_at"`
}

// JobDataEntity represents the job_data table structure
type JobDataEntity struct {
	JobID             big.Int   `cql:"job_id"`
	JobTitle          string    `cql:"job_title"`
	TaskDefinitionID  int       `cql:"task_definition_id"`
	CreatedChainID    string    `cql:"created_chain_id"`
	UserID            int64     `cql:"user_id"`
	LinkJobID         big.Int   `cql:"link_job_id"`
	ChainStatus       int       `cql:"chain_status"`
	Timezone          string    `cql:"timezone"`
	IsImua            bool      `cql:"is_imua"`
	JobType           string    `cql:"job_type"`
	TimeFrame         int64     `cql:"time_frame"`
	Recurring         bool      `cql:"recurring"`
	Status            string    `cql:"status"`
	JobCostPrediction big.Int   `cql:"job_cost_prediction"`
	JobCostActual     big.Int   `cql:"job_cost_actual"`
	TaskIDs           []int64   `cql:"task_ids"`
	CreatedAt         time.Time `cql:"created_at"`
	UpdatedAt         time.Time `cql:"updated_at"`
	LastExecutedAt    time.Time `cql:"last_executed_at"`
}

// TimeJobDataEntity represents the time_job_data table structure
type TimeJobDataEntity struct {
	JobID                     big.Int   `cql:"job_id"`
	TaskDefinitionID          int       `cql:"task_definition_id"`
	ScheduleType              string    `cql:"schedule_type"`
	TimeInterval              int64     `cql:"time_interval"`
	CronExpression            string    `cql:"cron_expression"`
	SpecificSchedule          string    `cql:"specific_schedule"`
	NextExecutionTimestamp    time.Time `cql:"next_execution_timestamp"`
	TargetChainID             string    `cql:"target_chain_id"`
	TargetContractAddress     string    `cql:"target_contract_address"`
	TargetFunction            string    `cql:"target_function"`
	ABI                       string    `cql:"abi"`
	ArgType                   int       `cql:"arg_type"`
	Arguments                 []string  `cql:"arguments"`
	DynamicArgumentsScriptURL string    `cql:"dynamic_arguments_script_url"`
	IsCompleted               bool      `cql:"is_completed"`
	LastExecutedAt            time.Time `cql:"last_executed_at"`
	ExpirationTime            time.Time `cql:"expiration_time"`
}

// EventJobDataEntity represents the event_job_data table structure
type EventJobDataEntity struct {
	JobID                      big.Int   `cql:"job_id"`
	TaskDefinitionID           int       `cql:"task_definition_id"`
	Recurring                  bool      `cql:"recurring"`
	TriggerChainID             string    `cql:"trigger_chain_id"`
	TriggerContractAddress     string    `cql:"trigger_contract_address"`
	TriggerEvent               string    `cql:"trigger_event"`
	TriggerEventFilterParaName string    `cql:"trigger_event_filter_para_name"`
	TriggerEventFilterValue    string    `cql:"trigger_event_filter_value"`
	TargetChainID              string    `cql:"target_chain_id"`
	TargetContractAddress      string    `cql:"target_contract_address"`
	TargetFunction             string    `cql:"target_function"`
	ABI                        string    `cql:"abi"`
	ArgType                    int       `cql:"arg_type"`
	Arguments                  []string  `cql:"arguments"`
	DynamicArgumentsScriptURL  string    `cql:"dynamic_arguments_script_url"`
	IsCompleted                bool      `cql:"is_completed"`
	LastExecutedAt             time.Time `cql:"last_executed_at"`
	ExpirationTime             time.Time `cql:"expiration_time"`
}

// ConditionJobDataEntity represents the condition_job_data table structure
type ConditionJobDataEntity struct {
	JobID                     big.Int   `cql:"job_id"`
	TaskDefinitionID          int       `cql:"task_definition_id"`
	Recurring                 bool      `cql:"recurring"`
	ConditionType             string    `cql:"condition_type"`
	UpperLimit                float64   `cql:"upper_limit"`
	LowerLimit                float64   `cql:"lower_limit"`
	ValueSourceType           string    `cql:"value_source_type"`
	ValueSourceURL            string    `cql:"value_source_url"`
	SelectedKeyRoute          string    `cql:"selected_key_route"`
	TargetChainID             string    `cql:"target_chain_id"`
	TargetContractAddress     string    `cql:"target_contract_address"`
	TargetFunction            string    `cql:"target_function"`
	ABI                       string    `cql:"abi"`
	ArgType                   int       `cql:"arg_type"`
	Arguments                 []string  `cql:"arguments"`
	DynamicArgumentsScriptURL string    `cql:"dynamic_arguments_script_url"`
	IsCompleted               bool      `cql:"is_completed"`
	LastExecutedAt            time.Time `cql:"last_executed_at"`
	ExpirationTime            time.Time `cql:"expiration_time"`
}

// TaskDataEntity represents the task_data table structure
type TaskDataEntity struct {
	TaskID               int64     `cql:"task_id"`
	TaskNumber           int64     `cql:"task_number"`
	JobID                big.Int   `cql:"job_id"`
	TaskDefinitionID     int       `cql:"task_definition_id"`
	CreatedAt            time.Time `cql:"created_at"`
	TaskOpxPredictedCost big.Int   `cql:"task_opx_predicted_cost"`
	TaskOpxActualCost    big.Int   `cql:"task_opx_actual_cost"`
	ExecutionTimestamp   time.Time `cql:"execution_timestamp"`
	ExecutionTxHash      string    `cql:"execution_tx_hash"`
	TaskPerformerID      int64     `cql:"task_performer_id"`
	TaskAttesterIDs      []int64   `cql:"task_attester_ids"`
	ConvertedArguments   string    `cql:"converted_arguments"`
	ProofOfTask          string    `cql:"proof_of_task"`
	SubmissionTxHash     string    `cql:"submission_tx_hash"`
	IsSuccessful         bool      `cql:"is_successful"`
	IsAccepted           bool      `cql:"is_accepted"`
	IsImua               bool      `cql:"is_imua"`
}

// KeeperDataEntity represents the keeper_data table structure
type KeeperDataEntity struct {
	KeeperID         int64     `cql:"keeper_id"`
	KeeperName       string    `cql:"keeper_name"`
	KeeperAddress    string    `cql:"keeper_address"`
	RewardsAddress   string    `cql:"rewards_address"`
	ConsensusAddress string    `cql:"consensus_address"`
	RegisteredTx     string    `cql:"registered_tx"`
	OperatorID       int64     `cql:"operator_id"`
	VotingPower      big.Int   `cql:"voting_power"`
	Whitelisted      bool      `cql:"whitelisted"`
	Registered       bool      `cql:"registered"`
	Online           bool      `cql:"online"`
	Version          string    `cql:"version"`
	OnImua           bool      `cql:"on_imua"`
	PublicIP         string    `cql:"public_ip"`
	PeerID           string    `cql:"peer_id"`
	ChatID           int64     `cql:"chat_id"`
	EmailID          string    `cql:"email_id"`
	RewardsBooster   big.Int   `cql:"rewards_booster"`
	NoExecutedTasks  int64     `cql:"no_executed_tasks"`
	NoAttestedTasks  int64     `cql:"no_attested_tasks"`
	Uptime           int64     `cql:"uptime"`
	KeeperPoints     big.Int   `cql:"keeper_points"`
	LastCheckedIn    time.Time `cql:"last_checked_in"`
}

// ApiKeyDataEntity represents the apikeys table structure
type ApiKeyDataEntity struct {
	Key          string    `cql:"key"`
	Owner        string    `cql:"owner"`
	IsActive     bool      `cql:"is_active"`
	RateLimit    int       `cql:"rate_limit"`
	SuccessCount int64     `cql:"success_count"`
	FailedCount  int64     `cql:"failed_count"`
	LastUsed     time.Time `cql:"last_used"`
	CreatedAt    time.Time `cql:"created_at"`
}
