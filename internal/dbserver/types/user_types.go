package types

import (
	commonTypes "github.com/trigg3rX/triggerx-backend/pkg/types"
)

type CreateUserDataRequest struct {
	UserAddress  string              `json:"user_address"`
	EtherBalance *commonTypes.BigInt `json:"ether_balance"`
	TokenBalance *commonTypes.BigInt `json:"token_balance"`
	UserPoints   float64             `json:"user_points"`
}

type UpdateUserBalanceRequest struct {
	UserID       int64               `json:"user_id"`
	EtherBalance *commonTypes.BigInt `json:"ether_balance"`
	TokenBalance *commonTypes.BigInt `json:"token_balance"`
}

type UserLeaderboardEntry struct {
	UserID      int64   `json:"user_id"`
	UserAddress string  `json:"user_address"`
	TotalJobs   int64   `json:"total_jobs"`
	TotalTasks  int64   `json:"total_tasks"`
	UserPoints  float64 `json:"user_points"`
}
