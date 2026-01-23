package types

// CreateUserDataRequest represents the request to create a new user
type CreateUserDataRequest struct {
	UserAddress string  `json:"user_address"`
	EmailID     string  `json:"email_id"`
	UserPoints  string  `json:"user_points"`
}

// UserLeaderboardEntry represents a user leaderboard entry
type UserLeaderboardEntry struct {
	UserAddress string  `json:"user_address"`
	TotalJobs   int64   `json:"total_jobs"`
	TotalTasks  int64   `json:"total_tasks"`
	UserPoints  string  `json:"user_points"`
}
