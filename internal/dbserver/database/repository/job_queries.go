package repository

// Create Queries
const (
	CreateJobDataQuery = `
		INSERT INTO triggerx.job_data (
			job_id, job_title, task_definition_id, created_chain_id, user_address, link_job_id, chain_status,
			safe_address, timezone, job_type, time_frame, recurring, status, job_cost_prediction,
			created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
)

// Write Queries
const (
	UpdateJobDataFromUserQuery = `
		UPDATE triggerx.job_data 
		SET job_title = ?, time_frame = ?, recurring = ?, status = ?,
			job_cost_prediction = ?, updated_at = ?
		WHERE job_id = ?`

	UpdateJobDataStatusQuery = `
		UPDATE triggerx.job_data 
		SET status = ?, updated_at = ?
		WHERE job_id = ?`
)

// Read Queries
const (
	GetJobDataByJobIDQuery = `
		SELECT job_id, job_title, task_definition_id, created_chain_id, user_address, link_job_id, chain_status,
			safe_address, timezone, job_type, time_frame, recurring, status, job_cost_prediction, job_cost_actual,
			task_ids, created_at, updated_at
		FROM triggerx.job_data 
		WHERE job_id = ?`

	GetTaskDefinitionIDByJobIDQuery = `
		SELECT task_definition_id FROM triggerx.job_data 
		WHERE job_id = ?`

	GetTaskIDsByJobIDQuery = `
		SELECT task_ids FROM triggerx.job_data 
		WHERE job_id = ?`

	GetJobsByUserAddressAndChainIDQuery = `
		SELECT job_id, job_title, task_definition_id, created_chain_id, user_address, link_job_id, chain_status,
			safe_address, timezone, job_type, time_frame, recurring, status, job_cost_prediction, job_cost_actual,
			task_ids, created_at, updated_at
		FROM triggerx.job_data
		WHERE user_address = ? AND created_chain_id = ? ALLOW FILTERING`

	GetJobsBySafeAddressQuery = `
		SELECT job_id, job_title, task_definition_id, created_chain_id, user_address, link_job_id, chain_status,
			safe_address, timezone, job_type, time_frame, recurring, status, job_cost_prediction, job_cost_actual,
			task_ids, created_at, updated_at
		FROM triggerx.job_data
		WHERE safe_address = ? ALLOW FILTERING`
)
