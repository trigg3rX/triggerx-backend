package types

import (
	"math/big"
	"time"
)

type UserData struct {
	UserID      int64   `json:"user_id"`
	UserAddress string  `json:"user_address"`
	JobIDs      []int64 `json:"job_ids"`
	// Current staked ether balance on TriggerGasRegistry
	EtherBalance *big.Int `json:"ether_balance"`
	// Current TG balance in User Account
	TokenBalance *big.Int `json:"token_balance"`
	// User points for creating jobs, for Leaderboard
	UserPoints    float64   `json:"user_points"`
	TotalJobs     int64     `json:"total_jobs"`
	TotalTasks    int64     `json:"total_tasks"`
	CreatedAt     time.Time `json:"created_at"`
	LastUpdatedAt time.Time `json:"last_updated_at"`
}

type JobData struct {
	JobID            int64  `json:"job_id"`
	JobTitle         string `json:"job_title"`
	TaskDefinitionID int    `json:"task_definition_id"`
	UserID           int64  `json:"user_id"`
	LinkJobID        int64  `json:"link_job_id"`
	ChainStatus      int    `json:"chain_status"`
	Custom           bool   `json:"custom"`
	TimeFrame        int64  `json:"time_frame"`
	Recurring        bool   `json:"recurring"`
	Status           string `json:"status"`
	// Intial cost prediction for the job
	JobCostPrediction float64 `json:"job_cost_prediction"`
	// Actual cost of the job, updated after each task execution
	JobCostActual float64 `json:"job_cost_actual"`
	// Task IDs for the job
	TaskIDs        []int64   `json:"task_ids"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	LastExecutedAt time.Time `json:"last_executed_at"`
	Timezone       string    `json:"timezone"`
	IsImua         bool      `json:"is_imua"`
}

type TimeJobData struct {
	JobID            int64 `json:"job_id"`
	TaskDefinitionID int   `json:"task_definition_id"`
	// Time based job specific fields
	ScheduleType           string    `json:"schedule_type"`
	TimeInterval           int64     `json:"time_interval"`
	CronExpression         string    `json:"cron_expression"`
	SpecificSchedule       string    `json:"specific_schedule"`
	NextExecutionTimestamp time.Time `json:"next_execution_timestamp"`
	// Target fields (common for all job types)
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
	// ExpirationTime = CreatedAt + TimeFrame, easier than using TimeFrame in schedulers
	ExpirationTime time.Time `json:"expiration_time"`
}

type EventJobData struct {
	JobID            int64 `json:"job_id"`
	TaskDefinitionID int   `json:"task_definition_id"`
	Recurring        bool  `json:"recurring"`
	// Event based job specific fields
	TriggerChainID         string `json:"trigger_chain_id"`
	TriggerContractAddress string `json:"trigger_contract_address"`
	TriggerEvent           string `json:"trigger_event"`
	// Target fields (common for all job types)
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

type ConditionJobData struct {
	JobID            int64 `json:"job_id"`
	TaskDefinitionID int   `json:"task_definition_id"`
	Recurring        bool  `json:"recurring"`
	// Condition based job specific fields
	ConditionType   string  `json:"condition_type"`
	UpperLimit      float64 `json:"upper_limit"`
	LowerLimit      float64 `json:"lower_limit"`
	ValueSourceType string  `json:"value_source_type"`
	ValueSourceUrl  string  `json:"value_source_url"`
	// Target fields (common for all job types)
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

type TaskData struct {
	TaskID               int64     `json:"task_id"`
	TaskNumber           int       `json:"task_number"`
	JobID                int64     `json:"job_id"`
	TaskDefinitionID     int       `json:"task_definition_id"`
	CreatedAt            time.Time `json:"created_at"`
	TaskOpxPredictedCost float64   `json:"task_opx_predicted_cost"`
	TaskOpxCost          float64   `json:"task_opx_cost"`
	ExecutionTimestamp   time.Time `json:"execution_timestamp"`
	ExecutionTxHash      string    `json:"execution_tx_hash"`
	TaskPerformerID      int64     `json:"task_performer_id"`
	TaskAttesterIDs      []int64   `json:"task_attester_ids"`
	ProofOfTask          string    `json:"proof_of_task"`
	TriggerData          []byte    `json:"trigger_data"`
	TpSignature          []byte    `json:"tp_signature"`
	TaSignature          []byte    `json:"ta_signature"`
	TaskSubmissionTxHash string    `json:"task_submission_tx_hash"`
	IsSuccessful         bool      `json:"is_successful"`
	IsImua               bool      `json:"is_imua"`
}

type KeeperData struct {
	KeeperID          int64     `json:"keeper_id"`
	KeeperName        string    `json:"keeper_name"`
	KeeperAddress     string    `json:"keeper_address"`
	ConsensusAddress  string    `json:"consensus_address"`
	RegisteredTx      string    `json:"registered_tx"`
	OperatorID        int64     `json:"operator_id"`
	RewardsAddress    string    `json:"rewards_address"`
	RewardsBooster    float32   `json:"rewards_booster"`
	VotingPower       int64     `json:"voting_power"`
	KeeperPoints      float64   `json:"keeper_points"`
	ConnectionAddress string    `json:"connection_address"`
	PeerID            string    `json:"peer_id"`
	Whitelisted       bool      `json:"whitelisted"`
	Registered        bool      `json:"registered"`
	Online            bool      `json:"online"`
	Version           string    `json:"version"`
	NoExecutedTasks   int       `json:"no_executed_tasks"`
	NoAttestedTasks   int       `json:"no_attested_tasks"`
	ChatID            int64     `json:"chat_id"`
	EmailID           string    `json:"email_id"`
	LastCheckedIn     time.Time `json:"last_checked_in"`
}

type ApiKey struct {
	Key          string    `json:"key"`
	Owner        string    `json:"owner"`
	IsActive     bool      `json:"is_active"`
	IsKeeper     bool      `json:"is_keeper"`
	RateLimit    int       `json:"rate_limit"`
	SuccessCount int64     `json:"success_count"`
	FailedCount  int64     `json:"failed_count"`
	LastUsed     time.Time `json:"last_used"`
	CreatedAt    time.Time `json:"created_at"`
}
