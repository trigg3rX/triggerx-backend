package types

import (
	"math/big"
	"time"
)

type UserData struct {
	UserID        int64     `json:"user_id"`
	UserAddress   string    `json:"user_address"`
	JobIDs        []int64   `json:"job_ids"`
	EtherBalance  *big.Int  `json:"ether_balance"`
	TokenBalance  *big.Int  `json:"token_balance"`
	UserPoints    float64   `json:"user_points"`
	TotalJobs     int64     `json:"total_jobs"`
	TotalTasks    int64     `json:"total_tasks"`
	CreatedAt     time.Time `json:"created_at"`
	LastUpdatedAt time.Time `json:"last_updated_at"`
	Email         string    `json:"email_id"`
}

type CreateUserDataRequest struct {
	UserAddress  string   `json:"user_address"`
	EtherBalance *big.Int `json:"ether_balance"`
	TokenBalance *big.Int `json:"token_balance"`
	UserPoints   float64  `json:"user_points"`
}

type UpdateUserBalanceRequest struct {
	UserID       int64    `json:"user_id"`
	EtherBalance *big.Int `json:"ether_balance"`
	TokenBalance *big.Int `json:"token_balance"`
}

type UserLeaderboardEntry struct {
	UserID      int64   `json:"user_id"`
	UserAddress string  `json:"user_address"`
	TotalJobs   int64   `json:"total_jobs"`
	TotalTasks  int64   `json:"total_tasks"`
	UserPoints  float64 `json:"user_points"`
}
