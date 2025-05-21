package types

import (
	"math/big"
)

type CreateJobData struct {
	UserAddress            string   `json:"user_address"`
	StakeAmount            *big.Int `json:"stake_amount"`
	TokenAmount            *big.Int `json:"token_amount"`
	TaskDefinitionID       int      `json:"task_definition_id"`
	Priority               int      `json:"priority"`
	Security               int      `json:"security"`
	Custom                 bool     `json:"custom"`
	TimeFrame              int64    `json:"time_frame"`
	Recurring              bool     `json:"recurring"`
	TimeInterval           int64    `json:"time_interval"`
	TriggerChainID         string   `json:"trigger_chain_id"`
	TriggerContractAddress string   `json:"trigger_contract_address"`
	TriggerEvent           string   `json:"trigger_event"`
	ScriptIPFSUrl          string   `json:"script_ipfs_url"`
	ScriptTriggerFunction  string   `json:"script_trigger_function"`
	TargetChainID          string   `json:"target_chain_id"`
	TargetContractAddress  string   `json:"target_contract_address"`
	TargetFunction         string   `json:"target_function"`
	ABI                    string   `json:"abi"`
	ArgType                int      `json:"arg_type"`
	Arguments              []string `json:"arguments"`
	ScriptTargetFunction   string   `json:"script_target_function"`
	JobCostPrediction      float64  `json:"job_cost_prediction"`
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
	JobID     int64 `json:"job_id"`
	Recurring bool  `json:"recurring"`
	TimeFrame int64 `json:"time_frame"`
}

type CreateTaskData struct {
	JobID            int64 `json:"job_id"`
	TaskDefinitionID int   `json:"task_definition_id"`
	TaskPerformerID  int64 `json:"task_performer_id"`
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
	KeeperAddress  string `json:"keeper_address"`
	RewardsAddress string `json:"rewards_address"`
	KeeperName     string `json:"keeper_name"`
	EmailID        string `json:"email_id"`
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
	Owner     string `json:"owner"`
	RateLimit int    `json:"rateLimit"`
}

type DailyRewardsPoints struct {
	KeeperID       int64   `json:"keeper_id"`
	RewardsBooster float32 `json:"rewards_booster"`
	KeeperPoints   float64 `json:"keeper_points"`
}
