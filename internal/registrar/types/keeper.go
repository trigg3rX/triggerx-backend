package types

import "time"

type KeeperRegistrationData struct {
	OperatorAddress string
	RewardsReceiver string
	TxHash          string
	OperatorID      int64
	VotingPower     int64
	Strategies      []string
}

type TaskSubmissionData struct {
	TaskID int64
	TaskNumber int64
	TaskDefinitionID int
	IsAccepted bool
	TaskSubmissionTxHash string
	PerformerAddress string
	AttesterIds []int64
	ExecutionTxHash string
	ExecutionTimestamp time.Time
	TaskOpxCost float64
	ProofOfTask string
}

type DailyRewardsPoints struct {
	KeeperID       int64   `json:"keeper_id"`
	RewardsBooster float64 `json:"rewards_booster"`
	KeeperPoints   float64 `json:"keeper_points"`
}