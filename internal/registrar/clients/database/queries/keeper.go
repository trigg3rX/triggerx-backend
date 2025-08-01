package queries

const (
	// Getters
	GetMaxKeeperID = `
		SELECT MAX(keeper_id) 
		FROM triggerx.keeper_data`
	GetKeeperIDByAddress = `
		SELECT keeper_id 
		FROM triggerx.keeper_data 
		WHERE keeper_address = ? 
		ALLOW FILTERING`
	GetDailyRewardsPoints = `
		SELECT keeper_id, rewards_booster, keeper_points 
		FROM triggerx.keeper_data
		WHERE registered = true AND whitelisted = true 
		ALLOW FILTERING`

	// Setters
	CreateKeeper = `
		INSERT INTO triggerx.keeper_data (
			keeper_id, 
			keeper_address, 
			rewards_address, 
			registered_tx, 
			operator_id, 
			voting_power, 
			strategies, 
			registered, 
			rewards_booster
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
	UpdateKeeper = `
		UPDATE triggerx.keeper_data 
		SET rewards_address = ?, 
			registered_tx = ?, 
			operator_id = ?, 
			voting_power = ?, 
			strategies = ?, 
			registered = ?, 
			rewards_booster = ?
		WHERE keeper_id = ?`
	UpdateKeeperRegistrationStatus = `
		UPDATE triggerx.keeper_data 
		SET registered = ?
		WHERE keeper_id = ?`
	UpdateKeeperPoints = `
		UPDATE triggerx.keeper_data 
		SET keeper_points = ? 
		WHERE keeper_id = ?`
)
