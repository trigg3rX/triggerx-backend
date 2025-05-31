package queries

const (
	SelectMaxUserIDQuery = `SELECT MAX(user_id) FROM triggerx.user_data`

	SelectUserByAddressQuery = `
		SELECT user_id, account_balance, token_balance, job_ids
		FROM triggerx.user_data 
		WHERE user_address = ? ALLOW FILTERING`

	SelectUserDataByIDQuery = `
			SELECT user_id, user_address, job_ids, account_balance
			FROM triggerx.user_data 
			WHERE user_id = ? ALLOW FILTERING`

	InsertUserDataQuery = `
			INSERT INTO triggerx.user_data (
				user_id, user_address, created_at, 
				job_ids, account_balance, token_balance, last_updated_at, user_points
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	UpdateUserDataQuery = `
			UPDATE triggerx.user_data 
			SET account_balance = ?, token_balance = ?, last_updated_at = ?
			WHERE user_id = ?`

	SelectUserPointsQuery = `
			SELECT user_points FROM triggerx.user_data 
			WHERE user_id = ?`

	UpdateUserPointsQuery = `
			UPDATE triggerx.user_data 
			SET user_points = ?, last_updated_at = ?
			WHERE user_id = ?`

	UpdateUserJobIDsQuery = `
			UPDATE triggerx.user_data 
			SET job_ids = ?, last_updated_at = ?
			WHERE user_id = ?`

	SelectUserJobIDsByAddressQuery = `
			SELECT user_id, job_ids
			FROM triggerx.user_data 
			WHERE user_address = ? ALLOW FILTERING`

	SelectUserLeaderboardQuery = `
			SELECT user_id, user_address, user_points 
			FROM triggerx.user_data`

	SelectUserJobCountQuery = `
			SELECT COUNT(*) FROM triggerx.job_data WHERE user_id = ? ALLOW FILTERING`

	SelectUserJobIDsByIDQuery = `
			SELECT user_id, job_ids
			FROM triggerx.user_data 
			WHERE user_id = ?`

	SelectUserPointsByAddressQuery = `
			SELECT account_balance
			FROM triggerx.user_data 
			WHERE user_address = ? ALLOW FILTERING`
)
