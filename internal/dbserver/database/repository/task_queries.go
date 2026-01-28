package repository

// Read Task Queries
const (
	GetTaskDataByIDQuery = `
        SELECT task_id, task_number, task_status, task_error, job_id, task_definition_id, created_at,
            executed_at, submitted_at, task_opx_predicted_cost, task_opx_actual_cost,
            execution_tx_hash, submission_tx_hash, converted_arguments, task_performer_address,
            task_attester_address, proof_of_task, is_successful, is_accepted, network
        FROM triggerx.task_data
        WHERE task_id = ?`

	GetTasksByJobIDQuery = `
		SELECT task_id, task_number, task_status, task_error, created_at, executed_at, submitted_at, 
			task_opx_predicted_cost, task_opx_actual_cost, execution_tx_hash, converted_arguments, 
			task_performer_address, task_attester_address, is_successful, is_accepted
		FROM triggerx.task_data
		WHERE job_id = ? ALLOW FILTERING`

	GetCreatedChainIDByJobIDQuery = `
        SELECT created_chain_id
        FROM triggerx.job_data
        WHERE job_id = ?`

	GetRecentTasksQuery = `
		SELECT task_id, task_number, job_id, task_definition_id, created_at,
		       task_opx_actual_cost, executed_at, execution_tx_hash, task_performer_address,
		       task_attester_address, task_status, task_error, network
		FROM triggerx.task_data
		LIMIT ?`

	// Statistics Queries for Global Counts
	GetTotalTasksCountQuery = `SELECT COUNT(*) FROM triggerx.task_data`

	GetTotalUsersCountQuery = `SELECT COUNT(*) FROM triggerx.user_data`

	GetTotalKeepersCountQuery = `SELECT COUNT(*) FROM triggerx.keeper_data WHERE whitelisted = true AND registered = true ALLOW FILTERING`

	GetTotalJobsCountQuery = `SELECT COUNT(*) FROM triggerx.job_data`
)
