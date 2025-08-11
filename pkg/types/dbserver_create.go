package types

import (
	"math/big"
	"time"
)

type JobStatus string

const (
	JobStatusPending JobStatus = "pending"
	JobStatusInQueue JobStatus = "in-queue"
	JobStatusRunning JobStatus = "running"
)

// Create New Job
type CreateJobRequest struct {
	// Common fields for all job types
	JobID             *big.Int  `json:"job_id" validate:"required"`
	UserAddress       string    `json:"user_address" validate:"required,ethereum_address"`
	DepositedEther    *big.Int  `json:"deposited_ether" validate:"required"`
	DepositedToken    *big.Int  `json:"deposited_token" validate:"required"`
	TaskDefinitionID  int       `json:"task_definition_id" validate:"required,min=1,max=6"`
	Custom            bool      `json:"custom"`
	JobTitle          string    `json:"job_title" validate:"required,min=3,max=100"`
	TimeFrame         int64     `json:"time_frame" validate:"required,min=1"`
	Recurring         bool      `json:"recurring"`
	JobCostPrediction float64   `json:"job_cost_prediction" validate:"required,min=0"`
	CreatedAt         time.Time `json:"created_at"`
	Timezone          string    `json:"timezone" validate:"required,timezone"`
	CreatedChainID    string    `json:"created_chain_id" validate:"required,chain_id"`
	// Time job specific fields
	ScheduleType     string `json:"schedule_type" validate:"required,oneof=cron interval specific"`
	TimeInterval     int64  `json:"time_interval,omitempty" validate:"omitempty,min=30"`
	CronExpression   string `json:"cron_expression,omitempty"`
	SpecificSchedule string `json:"specific_schedule,omitempty"`
	// Event job specific fields
	TriggerChainID         string `json:"trigger_chain_id,omitempty" validate:"omitempty,chain_id"`
	TriggerContractAddress string `json:"trigger_contract_address,omitempty" validate:"omitempty,ethereum_address"`
	TriggerEvent           string `json:"trigger_event,omitempty" validate:"omitempty"`
	// Condition job specific fields
	ConditionType   string  `json:"condition_type,omitempty" validate:"omitempty,oneof=price volume"`
	UpperLimit      float64 `json:"upper_limit,omitempty" validate:"omitempty,gt=0"`
	LowerLimit      float64 `json:"lower_limit,omitempty" validate:"omitempty,gt=0"`
	ValueSourceType string  `json:"value_source_type,omitempty" validate:"omitempty,oneof=api websocket"`
	ValueSourceUrl  string  `json:"value_source_url,omitempty" validate:"omitempty,url"`
	// Target fields (common for all job types)
	TargetChainID             string   `json:"target_chain_id" validate:"required,chain_id"`
	TargetContractAddress     string   `json:"target_contract_address" validate:"required,ethereum_address"`
	TargetFunction            string   `json:"target_function" validate:"required"`
	ABI                       string   `json:"abi" validate:"required"`
	ArgType                   int      `json:"arg_type" validate:"required"`
	Arguments                 []string `json:"arguments" validate:"required"`
	DynamicArgumentsScriptUrl string   `json:"dynamic_arguments_script_url,omitempty" validate:"omitempty,url"`
	IsImua                    bool     `json:"is_imua"`
}

type CreateJobResponse struct {
	UserID            int64      `json:"user_id"`
	AccountBalance    *big.Int   `json:"account_balance"`
	TokenBalance      *big.Int   `json:"token_balance"`
	JobIDs            []*big.Int `json:"job_ids"`
	TaskDefinitionIDs []int      `json:"task_definition_ids"`
	TimeFrames        []int64    `json:"time_frames"`
}

// Create New Task
type CreateTaskRequest struct {
	JobID            *big.Int `json:"job_id" validate:"required"`
	TaskDefinitionID int      `json:"task_definition_id" validate:"required"`
}

type CreateTaskResponse struct {
	TaskID           int64    `json:"task_id"`
	JobID            *big.Int `json:"job_id"`
	TaskDefinitionID int      `json:"task_definition_id"`
}

// Create New Keeper from Contracts (registrar)
type CreateKeeperData struct {
	KeeperAddress  string `json:"keeper_address"`
	RegisteredTx   string `json:"registered_tx"`
	RewardsAddress string `json:"rewards_address"`
	ChatID         int64  `json:"chat_id"`
}

// Create New Keeper from Google Form (google script)
type GoogleFormCreateKeeperData struct {
	KeeperAddress  string `json:"keeper_address" validate:"required,ethereum_address"`
	RewardsAddress string `json:"rewards_address" validate:"required,ethereum_address"`
	KeeperName     string `json:"keeper_name" validate:"required,min=3,max=50"`
	EmailID        string `json:"email_id" validate:"required,email"`
	OnImua         bool   `json:"on_imua"`
}

// Create New API Key (SDK)
type CreateApiKeyRequest struct {
	Owner     string `json:"owner" validate:"required,min=3,max=50"`
	RateLimit int    `json:"rateLimit" validate:"required,min=1,max=1000"`
}
