package types

type KeeperRegistrationData struct {
	OperatorAddress string
	RewardsReceiver string
	TxHash          string
	OperatorID      int64
	VotingPower     int64
	Strategies      []string
}

type TaskSubmissionData struct {
	TaskID int
	TaskNumber int64
	IsAccepted bool
	TaskSubmissionTxHash string
	KeeperIds []string
	AttesterSignatures []int64
	PerformerSignature []int64
}

type DailyRewardsPoints struct {
	KeeperID       int64   `json:"keeper_id"`
	RewardsBooster float32 `json:"rewards_booster"`
	KeeperPoints   float64 `json:"keeper_points"`
}