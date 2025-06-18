package queries

// Create Queries
const (
	GetMaxUserIDQuery = `SELECT MAX(user_id) FROM triggerx.user_data`

	CreateUserDataQuery = `
			INSERT INTO triggerx.user_data (
				user_id, user_address, 
				ether_balance, token_balance, user_points, 
				total_jobs, total_tasks, created_at
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
)

// Write Queries
const (
	// Update User Balance on Job Creation, Task Execution, Deposit Withdrawal
	UpdateUserBalanceQuery = `
			UPDATE triggerx.user_data 
			SET ether_balance = ?, token_balance = ?
			WHERE user_id = ?`

	// Update User Job IDs on Job Creation
	UpdateUserJobIDsQuery = `
			UPDATE triggerx.user_data 
			SET job_ids = ?, total_jobs = ?, last_updated_at = ?
			WHERE user_id = ?`

	// Update User Tasks and Points on Task Execution
	UpdateUserTasksAndPointsQuery = `
			UPDATE triggerx.user_data 
			SET total_tasks = ?, user_points = ?
			WHERE user_id = ?`
)

// Read Queries
const (
	GetUserIDByAddressQuery = `
			SELECT user_id
			FROM triggerx.user_data 
			WHERE user_address = ? ALLOW FILTERING`

	// Get User Data by ID
	GetUserDataByIDQuery = `
			SELECT user_id, user_address, 
				job_ids, total_jobs, total_tasks, 
				ether_balance, token_balance, user_points, 
				created_at, last_updated_at
			FROM triggerx.user_data 
			WHERE user_id = ?`

	// Get User Points by ID for Update after Task Execution
	GetUserPointsByIDQuery = `
			SELECT user_points 
			FROM triggerx.user_data 
			WHERE user_id = ?`

	// Get User Points by Address for Frontend Display
	GetUserPointsByAddressQuery = `
			SELECT user_points
			FROM triggerx.user_data 
			WHERE user_address = ? ALLOW FILTERING`

	// Get User Job IDs by Address for Frontend Display
	GetUserJobIDsByAddressQuery = `
			SELECT user_id, job_ids
			FROM triggerx.user_data 
			WHERE user_address = ? ALLOW FILTERING`

	GetUserCountersByIDQuery = `
			SELECT total_jobs, total_tasks
			FROM triggerx.user_data 
			WHERE user_id = ?`

	// Get User Leaderboard for Frontend Display
	GetUserLeaderboardQuery = `
			SELECT user_id, user_address, total_jobs, total_tasks, user_points 
			FROM triggerx.user_data`

	// Get User Leaderboard by Address for Frontend Serach functionality
	GetUserLeaderboardByAddressQuery = `
			SELECT user_id, user_address, total_jobs, total_tasks, user_points 
			FROM triggerx.user_data 
			WHERE user_address = ? ALLOW FILTERING`
)
