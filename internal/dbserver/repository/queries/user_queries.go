package queries

// Create Queries
const (
	CreateUserDataQuery = `
			INSERT INTO triggerx.user_data (
				user_address, email_id, job_ids, user_points,
				total_jobs, total_tasks, created_at, last_updated_at
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
)

// Write Queries
const (
	// Update User Job IDs on Job Creation
	UpdateUserJobIDsQuery = `
			UPDATE triggerx.user_data 
			SET job_ids = ?, total_jobs = ?, last_updated_at = ?
			WHERE user_address = ?`

	UpdateUserEmailByAddressQuery = `
		UPDATE triggerx.user_data
		SET email_id = ?, last_updated_at = ?
		WHERE user_address = ?`

	// Update User Points
	UpdateUserPointsQuery = `
		UPDATE triggerx.user_data 
		SET user_points = ?, last_updated_at = ?
		WHERE user_address = ?`
)

// Read Queries
const (
	// Get User Data by Address
	GetUserDataByAddressQuery = `
			SELECT user_address, email_id, job_ids, user_points,
				total_jobs, total_tasks, created_at, last_updated_at
			FROM triggerx.user_data 
			WHERE user_address = ?`

	// Get User Points by Address for Update after Task Execution
	GetUserPointsByAddressQuery = `
			SELECT user_points 
			FROM triggerx.user_data 
			WHERE user_address = ?`

	// Get User Job IDs by Address for Frontend Display
	GetUserJobIDsByAddressQuery = `
			SELECT job_ids
			FROM triggerx.user_data 
			WHERE user_address = ?`

	GetUserCountersByAddressQuery = `
			SELECT total_jobs, total_tasks
			FROM triggerx.user_data 
			WHERE user_address = ?`

	// Get User Leaderboard for Frontend Display
	GetUserLeaderboardQuery = `
			SELECT user_address, total_jobs, total_tasks, user_points 
			FROM triggerx.user_data`
)
