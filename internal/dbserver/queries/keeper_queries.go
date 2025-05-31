package queries

const (
	SelectKeeperLeaderboardQuery = `
			SELECT keeper_id, keeper_address, keeper_name, no_executed_tasks, keeper_points 
			FROM triggerx.keeper_data 
			WHERE registered = true AND whitelisted = true ALLOW FILTERING`

	SelectTaskCountByJobIDQuery = `
			SELECT COUNT(*) FROM triggerx.task_data WHERE job_id = ? ALLOW FILTERING`

	SelectKeeperByAddressQuery = `
			SELECT keeper_id, keeper_address, keeper_name, no_executed_tasks, keeper_points 
			FROM triggerx.keeper_data 
			WHERE registered = true AND keeper_address = ? ALLOW FILTERING`

	SelectKeeperByNameQuery = `
			SELECT keeper_id, keeper_address, keeper_name, no_executed_tasks, keeper_points 
                FROM triggerx.keeper_data 
                WHERE registered = true AND keeper_name = ? ALLOW FILTERING`

	SelectKeeperPointsByAddressQuery = `
			SELECT keeper_points
			FROM triggerx.keeper_data 
			WHERE keeper_address = ? ALLOW FILTERING`

		SelectKeeperIDByAddressQuery = `
			SELECT keeper_id FROM triggerx.keeper_data 
			WHERE keeper_address = ? ALLOW FILTERING`

	UpdateKeeperDataQuery = `
			UPDATE triggerx.keeper_data SET 
			keeper_name = ?, keeper_address = ?, rewards_address = ?, email_id = ?
			WHERE keeper_id = ?`

	InsertKeeperDataQuery = `
			INSERT INTO triggerx.keeper_data (
				keeper_id, keeper_name, keeper_address, rewards_booster,
				rewards_address, keeper_points, verified, email_id
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	SelectMaxKeeperIDQuery = `
			SELECT MAX(keeper_id) FROM triggerx.keeper_data`

	SelectKeeperDataByIDQuery = `
			SELECT keeper_id, keeper_name, keeper_address, registered_tx, operator_id,
				rewards_address, rewards_booster, voting_power, keeper_points, connection_address,
				strategies, verified, status, online, version, no_executed_tasks, chat_id, email_id
			FROM triggerx.keeper_data 
			WHERE keeper_id = ?`

	SelectPerformersQuery = `
			SELECT keeper_id, keeper_address 
			FROM triggerx.keeper_data 
			WHERE keeper_id = 2
			ALLOW FILTERING`

	SelectAllKeepersQuery = `
		SELECT keeper_id, keeper_name, keeper_address, registered_tx, operator_id,
		       rewards_address, rewards_booster, voting_power, keeper_points, connection_address,
		       strategies, verified, status, online, version, no_executed_tasks, chat_id, email_id
		FROM triggerx.keeper_data`

	SelectKeeperTaskCountQuery = `
			SELECT no_executed_tasks FROM triggerx.keeper_data WHERE keeper_id = ?`

	UpdateKeeperTaskCountQuery = `
			UPDATE triggerx.keeper_data SET no_executed_tasks = ? WHERE keeper_id = ?`

	SelectTaskFeeQuery = `
			SELECT task_fee FROM triggerx.task_data WHERE task_id = ?`

	SelectKeeperPointsQuery = `
			SELECT keeper_points FROM triggerx.keeper_data WHERE keeper_id = ?`

	UpdateKeeperPointsQuery = `
			UPDATE triggerx.keeper_data SET keeper_points = ? WHERE keeper_id = ?`

	SelectKeeperPointsByIDQuery = `
			SELECT keeper_points FROM triggerx.keeper_data WHERE keeper_id = ?`

	SelectKeeperIDByNameQuery = `
			SELECT keeper_id FROM triggerx.keeper_data 
			WHERE keeper_name = ?  ALLOW FILTERING`

	UpdateKeeperChatIDQuery = `
			UPDATE triggerx.keeper_data 
			SET chat_id = ? 
			WHERE keeper_id = ?`

	SelectKeeperCommunicationInfoQuery = `
			SELECT chat_id, keeper_name, email_id 
			FROM triggerx.keeper_data 
			WHERE keeper_id = ? ALLOW FILTERING`

)