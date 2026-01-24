package repository

// Read Queries
const (
	GetKeeperLeaderboardQuery = `
		SELECT keeper_address, keeper_name, no_executed_tasks, no_attested_tasks, keeper_points
		FROM triggerx.keeper_data 
		WHERE registered = true ALLOW FILTERING`
)
