package types

import (
	"math/big"
	"time"
)

const (
	// Time-based job task definitions
	TaskDefTimeBasedStart = 1
	TaskDefTimeBasedEnd   = 2

	// Event-based job task definitions
	TaskDefEventBasedStart = 3
	TaskDefEventBasedEnd   = 4

	// Condition-based job task definitions
	TaskDefConditionBasedStart = 5
	TaskDefConditionBasedEnd   = 6
)

type JobType string

const (
	JobTypeTime      JobType = "time"
	JobTypeEvent     JobType = "event"
	JobTypeCondition JobType = "condition"
)

type CreateJobData struct {
	// Common fields for all job types
	UserAddress       string    `json:"user_address" validate:"required,ethereum_address"`
	StakeAmount       *big.Int  `json:"stake_amount" validate:"required"`
	TokenAmount       *big.Int  `json:"token_amount" validate:"required"`
	TaskDefinitionID  int       `json:"task_definition_id" validate:"required,min=1,max=6"`
	Priority          int       `json:"priority" validate:"required,min=1,max=10"`
	Security          int       `json:"security" validate:"required,min=1,max=10"`
	Custom            bool      `json:"custom"`
	JobTitle          string    `json:"job_title" validate:"required,min=3,max=100"`
	TimeFrame         int64     `json:"time_frame" validate:"required,min=1"`
	Recurring         bool      `json:"recurring"`
	JobCostPrediction float64   `json:"job_cost_prediction" validate:"required,min=0"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
	LastExecutedAt    time.Time `json:"last_executed_at"`
	Timezone          string    `json:"timezone" validate:"required,timezone"`

	// Job type specific fields
	JobType JobType `json:"job_type" validate:"required,oneof=time event condition"`

	// Time job specific fields
	TimeInterval int64 `json:"time_interval,omitempty" validate:"omitempty,min=1"`

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
	TargetChainID         string   `json:"target_chain_id" validate:"required,chain_id"`
	TargetContractAddress string   `json:"target_contract_address" validate:"required,ethereum_address"`
	TargetFunction        string   `json:"target_function" validate:"required"`
	ABI                   string   `json:"abi" validate:"required"`
	ArgType               int      `json:"arg_type" validate:"required"`
	Arguments             []string `json:"arguments" validate:"required"`

	// Script fields (optional)
	ScriptIPFSUrl         string `json:"script_ipfs_url,omitempty" validate:"omitempty,ipfs_url"`
	ScriptTriggerFunction string `json:"script_trigger_function,omitempty" validate:"omitempty"`
	ScriptTargetFunction  string `json:"script_target_function,omitempty" validate:"omitempty"`
}

type CreateJobResponse struct {
	UserID            int64    `json:"user_id"`
	AccountBalance    *big.Int `json:"account_balance"`
	TokenBalance      *big.Int `json:"token_balance"`
	JobIDs            []int64  `json:"job_ids"`
	TaskDefinitionIDs []int    `json:"task_definition_ids"`
	TimeFrames        []int64  `json:"time_frames"`
}

type UpdateJobData struct {
	JobID          int64     `json:"job_id"`
	Recurring      bool      `json:"recurring"`
	TimeFrame      int64     `json:"time_frame"`
	UpdatedAt      time.Time `json:"updated_at"`
	LastExecutedAt time.Time `json:"last_executed_at"`
}

type CreateTaskData struct {
	JobID            int64 `json:"job_id" validate:"required"`
	TaskDefinitionID int   `json:"task_definition_id" validate:"required"`
	TaskPerformerID  int64 `json:"task_performer_id" validate:"required"`
}

type CreateTaskResponse struct {
	TaskID           int64 `json:"task_id"`
	JobID            int64 `json:"job_id"`
	TaskDefinitionID int   `json:"task_definition_id"`
	TaskPerformerID  int64 `json:"task_performer_id"`
	IsApproved       bool  `json:"is_approved"`
}

type GetPerformerData struct {
	KeeperID      int64  `json:"keeper_id"`
	KeeperAddress string `json:"keeper_address"`
}

type CreateKeeperData struct {
	KeeperAddress  string `json:"keeper_address"`
	RegisteredTx   string `json:"registered_tx"`
	RewardsAddress string `json:"rewards_address"`
	ChatID         int64  `json:"chat_id"`
}

type GoogleFormCreateKeeperData struct {
	KeeperAddress  string `json:"keeper_address" validate:"required,ethereum_address"`
	RewardsAddress string `json:"rewards_address" validate:"required,ethereum_address"`
	KeeperName     string `json:"keeper_name" validate:"required,min=3,max=50"`
	EmailID        string `json:"email_id" validate:"required,email"`
}

type UpdateKeeperConnectionDataResponse struct {
	KeeperID      int64  `json:"keeper_id"`
	KeeperAddress string `json:"keeper_address"`
	Verified      bool   `json:"verified"`
}

type UpdateKeeperStakeData struct {
	KeeperID      int64     `json:"keeper_id"`
	KeeperAddress string    `json:"keeper_address"`
	Stakes        []float64 `json:"stakes"`
	Strategies    []string  `json:"strategies"`
}

type KeeperLeaderboardEntry struct {
	KeeperID      int64   `json:"keeper_id"`
	KeeperAddress string  `json:"keeper_address"`
	KeeperName    string  `json:"keeper_name"`
	TasksExecuted int64   `json:"tasks_executed"`
	KeeperPoints  float64 `json:"keeper_points"`
}

type UserLeaderboardEntry struct {
	UserID         int64   `json:"user_id"`
	UserAddress    string  `json:"user_address"`
	TotalJobs      int64   `json:"total_jobs"`
	TasksCompleted int64   `json:"tasks_completed"`
	UserPoints     float64 `json:"user_points"`
}

type KeeperStatusUpdate struct {
	Status bool `json:"status"`
}

type CreateApiKeyRequest struct {
	Owner     string `json:"owner" validate:"required,min=3,max=50"`
	RateLimit int    `json:"rateLimit" validate:"required,min=1,max=1000"`
}

type DailyRewardsPoints struct {
	KeeperID       int64   `json:"keeper_id"`
	RewardsBooster float32 `json:"rewards_booster"`
	KeeperPoints   float64 `json:"keeper_points"`
}
