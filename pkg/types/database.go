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
	AccountBalance float64   `json:"account_balance"`
	CreatedAt      time.Time `json:"created_at"`
	LastUpdatedAt  time.Time `json:"last_updated_at"`
}

type JobData struct {
	JobID             int64     `json:"job_id"`
	JobType           int       `json:"jobType"`
	UserID            int64     `json:"user_id"`
	UserAddress       string    `json:"user_address"`
	ChainID           int       `json:"chain_id"`
	TimeFrame         int64     `json:"time_frame"`
	TimeInterval      int       `json:"time_interval"`
	ContractAddress   string    `json:"contract_address"`
	TargetFunction    string    `json:"target_function"`
	ArgType           int       `json:"arg_type"`
	Arguments         []string  `json:"arguments"`
	Status            bool      `json:"status"`
	JobCostPrediction int       `json:"job_cost_prediction"`
	ScriptFunction    string    `json:"script_function"`
	ScriptIpfsUrl     string    `json:"script_ipfs_url"`
	TimeCheck         time.Time `json:"time_check"`
	CreatedAt         time.Time `json:"created_at"`
	LastExecutedAt    time.Time `json:"last_executed_at"`
	UserBalance       float64   `json:"user_balance"`
}

type TaskData struct {
	TaskID              int64   `json:"task_id"`
	JobID               int64   `json:"job_id"`
	TaskNo              int     `json:"task_no"`
	QuorumID            int64   `json:"quorum_id"`
	QuorumNumber        int     `json:"quorum_number"`
	QuorumThreshold     float64 `json:"quorum_threshold"`
	TaskCreatedBlock    int64   `json:"task_created_block"`
	TaskCreatedTxHash   string  `json:"task_created_tx_hash"`
	TaskRespondedBlock  int64   `json:"task_responded_block"`
	TaskRespondedTxHash string  `json:"task_responded_tx_hash"`
	TaskHash            string  `json:"task_hash"`
	TaskResponseHash    string  `json:"task_response_hash"`
	QuorumKeeperHash    string  `json:"quorum_keeper_hash"`
	TaskFee             float64 `json:"task_fee"`
}

type QuorumData struct {
	QuorumID               int64    `json:"quorum_id"`
	QuorumNo               int      `json:"quorum_no"`
	QuorumCreationBlock    int64    `json:"quorum_creation_block"`
	QuorumTerminationBlock int64    `json:"quorum_termination_block"`
	QuorumTxHash           string   `json:"quorum_tx_hash"`
	Keepers                []string `json:"keepers"`
	QuorumStakeTotal       int64    `json:"quorum_stake_total"`
	TaskIDs                []int64  `json:"task_ids"`
	QuorumStatus           bool     `json:"quorum_status"`
}

type QuorumDataResponse struct {
	QuorumID          int64    `json:"quorum_id"`
	QuorumNo          int      `json:"quorum_no"`
	QuorumStatus      bool     `json:"quorum_status"`
	QuorumStakeTotal  int64    `json:"quorum_stake_total"`
	QuorumStrength     int      `json:"quorum_strength"`
}

type KeeperData struct {
	KeeperID          int64     `json:"keeper_id"`
	WithdrawalAddress string    `json:"withdrawal_address"`
	Stakes            []float64 `json:"stakes"`
	Strategies        []string  `json:"strategies"`
	Verified          bool      `json:"verified"`
	CurrentQuorumNo   int       `json:"current_quorum_no"`
	RegisteredTx      string    `json:"registered_tx"`
	Status            bool      `json:"status"`
	BlsSigningKeys    []string  `json:"bls_signing_keys"`
	ConnectionAddress string    `json:"connection_address"`
}

type TaskHistory struct {
	TaskID           int64    `json:"task_id"`
	QuorumID         int64    `json:"quorum_id"`
	Keepers          []string `json:"keepers"`
	Responses        []string `json:"responses"`
	ConsensusMethod  string   `json:"consensus_method"`
	ValidationStatus bool     `json:"validation_status"`
	TxHash           string   `json:"tx_hash"`
}

type PeerInfo struct {
	ID        string `json:"id"`
	Addresses []string `json:"addresses"`
}
