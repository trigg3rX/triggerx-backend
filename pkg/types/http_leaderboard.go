package types

type UserLeaderboardEntry struct {
	UserAddress string  `json:"user_address"`
	TotalJobs   int64   `json:"total_jobs"`
	TotalTasks  int64   `json:"total_tasks"`
	UserPoints  string  `json:"user_points"`
}

type KeeperLeaderboardEntry struct {
	KeeperID        int64   `json:"keeper_id"`
	KeeperAddress   string  `json:"keeper_address"`
	KeeperName      string  `json:"keeper_name"`
	NoExecutedTasks int64   `json:"no_executed_tasks"`
	NoAttestedTasks int64   `json:"no_attested_tasks"`
	KeeperPoints    string  `json:"keeper_points"`
	OnImua          bool    `json:"on_imua"`
}
