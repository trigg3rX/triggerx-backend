package queries

// Read Queries
const (
	GetKeeperLeaderboardQuery = `
		SELECT keeper_id, keeper_address, keeper_name, no_executed_tasks, no_attested_tasks, keeper_points, on_imua
		FROM triggerx.keeper_data 
		WHERE registered = true AND whitelisted = true ALLOW FILTERING`
)
