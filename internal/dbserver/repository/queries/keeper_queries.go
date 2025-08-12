package queries

// Create Queries
const (
	GetMaxKeeperIDQuery = `SELECT MAX(keeper_id) FROM triggerx.keeper_data`

	CreateNewKeeperQuery = `
			INSERT INTO triggerx.keeper_data (
				keeper_id, keeper_name, keeper_address, rewards_booster,
				keeper_points, whitelisted, email_id
			) VALUES (?, ?, ?, ?, ?, ?, ?)`

	CreateNewKeeperFromGoogleFormQuery = `
			INSERT INTO triggerx.keeper_data (
				keeper_id, keeper_name, keeper_address,
				rewards_address, email_id, on_imua
			) VALUES (?, ?, ?, ?, ?, ?)`

	UpdateKeeperFromGoogleFormQuery = `
			UPDATE triggerx.keeper_data SET 
				keeper_name = ?, keeper_address = ?, rewards_address = ?, email_id = ?, on_imua = ?
			WHERE keeper_id = ?`
)

// Update Queries
const (
	UpdateKeeperDataFromL2Query = `
			UPDATE triggerx.keeper_data 
			SET registered = true, registered_tx = ?, rewards_address = ?, operator_id = ?, 
			voting_power = ?, strategies = ?,
			WHERE keeper_id = ?`

	UpdateKeeperTaskCountsQuery = `
			UPDATE triggerx.keeper_data 
			SET no_executed_tasks = ?, no_attested_tasks = ? 
			WHERE keeper_id = ?`

	UpdateKeeperPointsQuery = `
			UPDATE triggerx.keeper_data 
			SET keeper_points = ? 
			WHERE keeper_id = ?`

	UpdateKeeperChatIDQuery = `
			UPDATE triggerx.keeper_data 
			SET chat_id = ? 
			WHERE keeper_id = ?`

	UpdateKeeperFromHealthCheckInQuery = `
			UPDATE triggerx.keeper_data 
			SET consensus_address = ?, connection_address = ?, peer_id = ?, version = ?, 
			last_checked_in = ? , online = ? 
			WHERE keeper_id = ?`

	UpdateKeeperFromHealthCheckOutQuery = `
			UPDATE triggerx.keeper_data 
			SET online = false, last_checked_in = ? 
			WHERE keeper_id = ?`

	UpdateKeeperTaskCountQuery = `
			UPDATE triggerx.keeper_data 
			SET no_executed_tasks = ? 
			WHERE keeper_id = ?`
)

// Read Queries
const (
	GetKeeperDataByIDQuery = `
		SELECT keeper_id, keeper_name, keeper_address, consensus_address, registered_tx, operator_id,
			rewards_address, rewards_booster, voting_power, keeper_points, connection_address,
			peer_id, whitelisted, registered, online, version, no_executed_tasks,
			no_attested_tasks, chat_id, email_id, last_checked_in, on_imua, uptime
		FROM triggerx.keeper_data 
		WHERE keeper_id = ?`

	GetKeeperIDByAddressQuery = `
		SELECT keeper_id
		FROM triggerx.keeper_data 
		WHERE keeper_address = ? ALLOW FILTERING`

	GetKeeperIDByOperatorIDQuery = `
		SELECT keeper_id
		FROM triggerx.keeper_data 
		WHERE operator_id = ? ALLOW FILTERING`

	GetKeeperIDByNameQuery = `
		SELECT keeper_id
		FROM triggerx.keeper_data 
		WHERE keeper_name = ? ALLOW FILTERING`

	GetKeeperLeaderboardQuery = `
		SELECT keeper_id, keeper_address, keeper_name, no_executed_tasks, no_attested_tasks, keeper_points, on_imua
		FROM triggerx.keeper_data 
		WHERE registered = true AND whitelisted = true ALLOW FILTERING`

	GetKeeperLeaderboardByAddressQuery = `
		SELECT keeper_id, keeper_address, keeper_name, no_executed_tasks, no_attested_tasks, keeper_points 
		FROM triggerx.keeper_data 
		WHERE registered = true AND keeper_address = ? ALLOW FILTERING`

	GetKeeperLeaderboardByNameQuery = `
		SELECT keeper_id, keeper_address, keeper_name, no_executed_tasks, no_attested_tasks, keeper_points 
		FROM triggerx.keeper_data 
		WHERE registered = true AND keeper_name = ? ALLOW FILTERING`

	GetKeeperLeaderboardByOnImuaQuery = `
		SELECT keeper_id, keeper_address, keeper_name, no_executed_tasks, no_attested_tasks, keeper_points, on_imua
		FROM triggerx.keeper_data 
		WHERE registered = true AND whitelisted = true AND on_imua = ? ALLOW FILTERING`

	GetKeeperAsPerformersQuery = `
		SELECT keeper_id, keeper_address 
		FROM triggerx.keeper_data 
		WHERE whitelisted = true AND registered = true ALLOW FILTERING`

	GetKeeperTaskCountByIDQuery = `
		SELECT no_executed_tasks FROM triggerx.keeper_data WHERE keeper_id = ?`

	GetKeeperPointsByIDQuery = `
		SELECT keeper_points FROM triggerx.keeper_data WHERE keeper_id = ?`

	GetKeeperCommunicationInfoQuery = `
		SELECT chat_id, keeper_name, email_id 
		FROM triggerx.keeper_data 
		WHERE keeper_id = ?`
)
