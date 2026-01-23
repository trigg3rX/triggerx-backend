package types

// PerformerData represents the data for next performer to be selected
type PerformerData struct {
	OperatorID    int64  `json:"operator_id"`
	KeeperAddress string `json:"keeper_address"`
	Network       KeeperNetwork `json:"network"`
}
