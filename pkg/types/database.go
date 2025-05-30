package types

import (
	"math/big"
	"time"
)

type UserData struct {
	UserID         int64     `json:"user_id"`
	UserAddress    string    `json:"user_address"`
	CreatedAt      time.Time `json:"created_at"`
	JobIDs         []int64   `json:"job_ids"`
	AccountBalance *big.Int  `json:"account_balance"`
	TokenBalance   *big.Int  `json:"token_balance"`
	LastUpdatedAt  time.Time `json:"last_updated_at"`
	UserPoints     float64   `json:"user_points"`
}

type JobData struct {
	JobID             int64   `json:"job_id"`
	JobTitle          string  `json:"job_title"`
	TaskDefinitionID  int     `json:"task_definition_id"`
	UserID            int64   `json:"user_id"`
	LinkJobID         int64   `json:"link_job_id"`
	ChainStatus       int     `json:"chain_status"`
	Custom            bool    `json:"custom"`
	TimeFrame         int64   `json:"time_frame"`
	Recurring         bool    `json:"recurring"`
	Status            string  `json:"status"`
	JobCostPrediction float64 `json:"job_cost_prediction"`
	TaskIDs           []int64 `json:"task_ids"`
}

type TaskData struct {
	TaskID               int64     `json:"task_id"`
	TaskNumber           int       `json:"task_number"`
	JobID                int64     `json:"job_id"`
	TaskDefinitionID     int       `json:"task_definition_id"`
	CreatedAt            time.Time `json:"created_at"`
	TaskFee              float64   `json:"task_fee"`
	ExecutionTimestamp   time.Time `json:"execution_timestamp"`
	ExecutionTxHash      string    `json:"execution_tx_hash"`
	TaskPerformerID      int64     `json:"task_performer_id"`
	ProofOfTask          string    `json:"proof_of_task"`
	ActionDataCID        string    `json:"action_data_cid"`
	TaskAttesterIDs      []int64   `json:"task_attester_ids"`
	IsApproved           bool      `json:"is_approved"`
	TpSignature          []byte    `json:"tp_signature"`
	TaSignature          []byte    `json:"ta_signature"`
	TaskSubmissionTxHash string    `json:"task_submission_tx_hash"`
	IsSuccessful         bool      `json:"is_successful"`
}

type KeeperData struct {
	KeeperID          int64    `json:"keeper_id"`
	KeeperAddress     string   `json:"keeper_address"`
	KeeperName        string   `json:"keeper_name"`
	RegisteredTx      string   `json:"registered_tx"`
	RewardsBooster    float32  `json:"rewards_booster"`
	OperatorID        int64    `json:"operator_id"`
	RewardsAddress    string   `json:"rewards_address"`
	KeeperPoints      float64  `json:"keeper_points"`
	ConnectionAddress string   `json:"connection_address"`
	PeerID            string   `json:"peer_id"`
	Strategies        []string `json:"strategies"`
	VotingPower       int64    `json:"voting_power"`
	Verified          bool     `json:"verified"`
	Status            bool     `json:"status"`
	Online            bool     `json:"online"`
	Version           string   `json:"version"`
	NoExcTask         int      `json:"no_executed_tasks"`
	ChatID            int64    `json:"chat_id"`
	EmailID           string   `json:"email_id"`
}

type ApiKey struct {
	Key       string    `json:"key"`
	Owner     string    `json:"owner"`
	IsActive  bool      `json:"isActive"`
	RateLimit int       `json:"rateLimit"`
	LastUsed  time.Time `json:"lastUsed"`
	CreatedAt time.Time `json:"createdAt"`
}

type TimeJobData struct {
	JobID                         int64     `json:"job_id"`
	TimeFrame                     int64     `json:"time_frame"`
	Recurring                     bool      `json:"recurring"`
	ScheduleType                  string    `json:"schedule_type"`
	TimeInterval                  int64     `json:"time_interval"`
	CronExpression                string    `json:"cron_expression"`
	SpecificSchedule              string    `json:"specific_schedule"`
	Timezone                      string    `json:"timezone"`
	NextExecutionTimestamp        time.Time `json:"next_execution_timestamp"`
	TargetChainID                 string    `json:"target_chain_id"`
	TargetContractAddress         string    `json:"target_contract_address"`
	TargetFunction                string    `json:"target_function"`
	ABI                           string    `json:"abi"`
	ArgType                       int       `json:"arg_type"`
	Arguments                     []string  `json:"arguments"`
	DynamicArgumentsScriptIPFSUrl string    `json:"dynamic_arguments_script_ipfs_url"`
}

type EventJobData struct {
	JobID                         int64    `json:"job_id"`
	TimeFrame                     int64    `json:"time_frame"`
	Recurring                     bool     `json:"recurring"`
	TriggerChainID                string   `json:"trigger_chain_id"`
	TriggerContractAddress        string   `json:"trigger_contract_address"`
	TriggerEvent                  string   `json:"trigger_event"`
	TargetChainID                 string   `json:"target_chain_id"`
	TargetContractAddress         string   `json:"target_contract_address"`
	TargetFunction                string   `json:"target_function"`
	ABI                           string   `json:"abi"`
	ArgType                       int      `json:"arg_type"`
	Arguments                     []string `json:"arguments"`
	DynamicArgumentsScriptIPFSUrl string   `json:"dynamic_arguments_script_ipfs_url"`
}

type ConditionJobData struct {
	JobID                         int64    `json:"job_id"`
	TimeFrame                     int64    `json:"time_frame"`
	Recurring                     bool     `json:"recurring"`
	ConditionType                 string   `json:"condition_type"`
	UpperLimit                    float64  `json:"upper_limit"`
	LowerLimit                    float64  `json:"lower_limit"`
	ValueSourceType               string   `json:"value_source_type"`
	ValueSourceUrl                string   `json:"value_source_url"`
	TargetChainID                 string   `json:"target_chain_id"`
	TargetContractAddress         string   `json:"target_contract_address"`
	TargetFunction                string   `json:"target_function"`
	ABI                           string   `json:"abi"`
	ArgType                       int      `json:"arg_type"`
	Arguments                     []string `json:"arguments"`
	DynamicArgumentsScriptIPFSUrl string   `json:"dynamic_arguments_script_ipfs_url"`
}
