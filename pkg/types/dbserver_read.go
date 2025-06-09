package types

type JobResponse struct {
	JobData          JobData           `json:"job_data"`
	TimeJobData      *TimeJobData      `json:"time_job_data,omitempty"`
	EventJobData     *EventJobData     `json:"event_job_data,omitempty"`
	ConditionJobData *ConditionJobData `json:"condition_job_data,omitempty"`
}

type GetPerformerData struct {
	KeeperID      int64  `json:"keeper_id"`
	KeeperAddress string `json:"keeper_address"`
}

type DailyRewardsPoints struct {
	KeeperID       int64   `json:"keeper_id"`
	RewardsBooster float32 `json:"rewards_booster"`
	KeeperPoints   float64 `json:"keeper_points"`
}

type KeeperLeaderboardEntry struct {
	KeeperID      int64   `json:"keeper_id"`
	KeeperAddress string  `json:"keeper_address"`
	KeeperName    string  `json:"keeper_name"`
	TasksExecuted int64   `json:"tasks_executed"`
	KeeperPoints  float64 `json:"keeper_points"`
}

type UserLeaderboardEntry struct {
	UserID         int64   `json:"user_id"`
	UserAddress    string  `json:"user_address"`
	TotalJobs      int64   `json:"total_jobs"`
	TasksCompleted int64   `json:"tasks_completed"`
	UserPoints     float64 `json:"user_points"`
}
