package types

import (
	"math/big"
	"time"
)

type UserData struct {
	UserID         int64     `json:"user_id"`
	UserAddress    string    `json:"user_address"`
	JobIDs         []int64   `json:"job_ids"`
	StakeAmount    *big.Int  `json:"stake_amount"`
	AccountBalance *big.Int  `json:"account_balance"`
	CreatedAt      time.Time `json:"created_at"`
	LastUpdatedAt  time.Time `json:"last_updated_at"`
}

type JobData struct {
	JobID                  int64     `json:"job_id"`
	JobType               int       `json:"job_type"`
	UserID                int64     `json:"user_id"`
	ChainID               int       `json:"chain_id"`
	TimeFrame             int64     `json:"time_frame"`
	TimeInterval          int       `json:"time_interval"`
	TriggerContractAddress string    `json:"trigger_contract_address"`
	TriggerEvent          string    `json:"trigger_event"`
	TargetContractAddress string    `json:"target_contract_address"`
	TargetFunction        string    `json:"target_function"`
	ArgType               int       `json:"arg_type"`
	Arguments             []string  `json:"arguments"`
	Recurring             bool      `json:"recurring"`
	ScriptFunction        string    `json:"script_function"`
	ScriptIPFSUrl         string    `json:"script_ipfs_url"`
	Status                bool      `json:"status"`
	JobCostPrediction     int       `json:"job_cost_prediction"`
	CreatedAt             time.Time `json:"created_at"`
	LastExecutedAt        time.Time `json:"last_executed_at"`
	Priority              int       `json:"priority"`
	Security              int       `json:"security"`
	TaskIDs               []int64   `json:"task_ids"`
	LinkJobID             int64     `json:"link_job_id"`
}

type TaskData struct {
	TaskID                     int64     `json:"task_id"`
	JobID                      int64     `json:"job_id"`
	TaskDefinitionID           int64     `json:"task_definition_id"`
	TaskCreatedTxHash         string    `json:"task_created_tx_hash"`
	TaskRespondedTxHash       string    `json:"task_responded_tx_hash"`
	TaskHash                  string    `json:"task_hash"`
	TaskResponseHash          string    `json:"task_response_hash"`
	TaskFee                   float64   `json:"task_fee"`
	JobType                   string    `json:"job_type"`
	BlockExpiry               *big.Int  `json:"block_expiry"`
	BaseRewardFeeForAttesters *big.Int  `json:"base_reward_fee_for_attesters"`
	BaseRewardFeeForPerformer *big.Int  `json:"base_reward_fee_for_performer"`
	BaseRewardFeeForAggregator *big.Int  `json:"base_reward_fee_for_aggregator"`
	DisputePeriodBlocks       *big.Int  `json:"dispute_period_blocks"`
	MinimumVotingPower        *big.Int  `json:"minimum_voting_power"`
	RestrictedOperatorIndexes []big.Int `json:"restricted_operator_indexes"`
	ProofOfTask               string    `json:"proof_of_task"`
	Data                      []byte    `json:"data"`
	TaskPerformer             string    `json:"task_performer"`
	IsApproved                bool      `json:"is_approved"`
	TpSignature               []byte    `json:"tp_signature"`
	TaSignature               []big.Int `json:"ta_signature"`
	OperatorIds               []big.Int `json:"operator_ids"`
}

type QuorumData struct {
	QuorumID                int64    `json:"quorum_id"`
	QuorumNo                int      `json:"quorum_no"`
	QuorumCreationBlock     int64    `json:"quorum_creation_block"`
	QuorumTerminationBlock  int64    `json:"quorum_termination_block"`
	QuorumTxHash            string   `json:"quorum_tx_hash"`
	Keepers                 []string `json:"keepers"`
	QuorumStakeTotal        int64    `json:"quorum_stake_total"`
	TaskIDs                 []int64  `json:"task_ids"`
	QuorumStatus            bool     `json:"quorum_status"`
}

type KeeperData struct {
	KeeperID          int64     `json:"keeper_id"`
	KeeperAddress     string    `json:"keeper_address"`
	RewardsAddress    string    `json:"rewards_address"`
	Stakes            []float64 `json:"stakes"`
	Strategies        []string  `json:"strategies"`
	Verified          bool      `json:"verified"`
	RegisteredTx      string    `json:"registered_tx"`
	Status            bool      `json:"status"`
	BlsSigningKeys    []string  `json:"bls_signing_keys"`
	ConnectionAddress string    `json:"connection_address"`
}

type TaskHistory struct {
	TaskID       int64    `json:"task_id"`
	Performer    string   `json:"performer"`
	Attesters    []string `json:"attesters"`
	ProofOfTask  string   `json:"proof_of_task"`
	IsSuccessful bool     `json:"is_successful"`
	TxHash       string   `json:"tx_hash"`
}