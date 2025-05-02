package types

import (
	"math/big"
	"time"
)

type UserData struct {
	// Fixed Values
	UserID      int64     `json:"user_id"`
	UserAddress string    `json:"user_address"`
	CreatedAt   time.Time `json:"created_at"`

	// Active Values
	JobIDs         []int64   `json:"job_ids"`
	AccountBalance *big.Int  `json:"account_balance"` // Balance in Wei (ETH)
	TokenBalance   *big.Int  `json:"token_balance"`   // Balance in Wei (ETH)
	LastUpdatedAt  time.Time `json:"last_updated_at"`
	UserPoints     float64   `json:"user_points"` // Points earned by user
}

type JobData struct {
	// Fixed Values
	JobID            int64 `json:"job_id"`
	TaskDefinitionID int   `json:"task_definition_id"`
	UserID           int64 `json:"user_id"`
	Priority         int   `json:"priority"` // Defines BaseFee for Keepers
	Security         int   `json:"security"` // Defines VotingPower for Aggregator
	LinkJobID        int64 `json:"link_job_id"`
	ChainStatus      int   `json:"chain_status"`
	// 0 = Chain Head, 1 = Chain Block
	Custom bool `json:"custom"` // Indicates if job is customized or fixed

	// Can be Updated By User
	TimeFrame int64 `json:"time_frame"`
	Recurring bool  `json:"recurring"` // Only True -> False is allowed

	// Trigger Values
	TimeInterval           int64  `json:"time_interval"`
	TriggerChainID         string `json:"trigger_chain_id"`
	TriggerContractAddress string `json:"trigger_contract_address"`
	TriggerEvent           string `json:"trigger_event"`
	ScriptIPFSUrl          string `json:"script_ipfs_url"`
	ScriptTriggerFunction  string `json:"script_trigger_function"`

	// Action Values
	TargetChainID         string   `json:"target_chain_id"`
	TargetContractAddress string   `json:"target_contract_address"`
	TargetFunction        string   `json:"target_function"`
	ArgType               int      `json:"arg_type"`
	Arguments             []string `json:"arguments"`
	ScriptTargetFunction  string   `json:"script_target_function"`
	ABI                   string   `json:"abi"`
	// Status Values
	Status            bool      `json:"status"`
	JobCostPrediction float64   `json:"job_cost_prediction"`
	CreatedAt         time.Time `json:"created_at"`
	LastExecutedAt    time.Time `json:"last_executed_at"`
	TaskIDs           []int64   `json:"task_ids"`
}

type TaskData struct {
	// Fixed Values
	TaskID           int64     `json:"task_id"`
	TaskNumber       int       `json:"task_number"`
	JobID            int64     `json:"job_id"`
	TaskDefinitionID int       `json:"task_definition_id"`
	CreatedAt        time.Time `json:"created_at"`
	TaskFee          float64   `json:"task_fee"`

	// Action Values
	ExecutionTimestamp time.Time `json:"execution_timestamp"`
	ExecutionTxHash    string    `json:"execution_tx_hash"`
	TaskPerformerID    int64     `json:"task_performer_id"`

	// ProofOfTask Values
	ProofOfTask     string  `json:"proof_of_task"`
	ActionDataCID   string  `json:"action_data_cid"`
	TaskAttesterIDs []int64 `json:"task_attester_ids"`

	// Contract Values
	IsApproved           bool   `json:"is_approved"`
	TpSignature          []byte `json:"tp_signature"`
	TaSignature          []byte `json:"ta_signature"`
	TaskSubmissionTxHash string `json:"task_submission_tx_hash"`

	// Status Values
	IsSuccessful bool `json:"is_successful"`
}

type KeeperData struct {
	// Fixed Values
	KeeperID       int64   `json:"keeper_id"`
	KeeperAddress  string  `json:"keeper_address"`
	KeeperName     string  `json:"keeper_name"`
	RegisteredTx   string  `json:"registered_tx"`
	RewardsBooster float32 `json:"rewards_booster"`

	// Active Values
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
	NoExcTask         int      `json:"no_exctask"`
	ChatID            int64    `json:"chat_id"`
	EmailID           string   `json:"email_id"`
}

// ApiKey represents an API key in the system
type ApiKey struct {
	Key       string    `json:"key"`
	Owner     string    `json:"owner"`
	IsActive  bool      `json:"isActive"`
	RateLimit int       `json:"rateLimit"`
	LastUsed  time.Time `json:"lastUsed"`
	CreatedAt time.Time `json:"createdAt"`
}
