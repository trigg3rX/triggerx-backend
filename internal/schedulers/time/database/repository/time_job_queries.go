package repository

// Read Queries
const (
	getTimeJobsByNextExecutionTimestampQuery = `
		SELECT job_id, task_definition_id, network, schedule_type, time_interval, 
			cron_expression, specific_schedule, timezone, next_execution_timestamp,
			target_chain_id, target_contract_address, target_function, abi, 
			arg_type, arguments, execution_script_url, execution_script_language, 
			execution_script_hash, max_execution_time, challenge_period,
			is_active, last_executed_at, expiration_time
		FROM triggerx.time_job_data
		WHERE next_execution_timestamp <= ? 
			AND is_active = true
		ALLOW FILTERING`
	
	getMaxTaskIDQuery = `SELECT MAX(task_id) FROM triggerx.task_data`

	getTaskIDsByJobIDQuery = `
		SELECT task_ids FROM triggerx.job_data 
		WHERE job_id = ?`

	getStorageByJobIDQuery = `
		SELECT storage_key, storage_value, updated_at
		FROM triggerx.script_storage
		WHERE job_id = ?`

	getJobCostPredictionQuery = `
		SELECT job_cost_prediction
		FROM triggerx.job_data
		WHERE job_id = ?`
)

// Write Queries
const (
	updateJobDataToCompletedQuery = `
		UPDATE triggerx.job_data
		SET status = 'completed'
		WHERE job_id = ?`

	updateTimeJobStatusQuery = `
		UPDATE triggerx.time_job_data
		SET is_active = ?
		WHERE job_id = ?`

	updateTimeJobNextExecutionTimestampQuery = `
        UPDATE triggerx.time_job_data
        SET next_execution_timestamp = ?
        WHERE job_id = ?`

	createTaskDataQuery = `
        INSERT INTO triggerx.task_data (
            task_id, job_id, task_definition_id, network, task_status, created_at, task_opx_predicted_cost
        ) VALUES (?, ?, ?, ?, ?, ?, ?)`

	addTaskIDToJobQuery = `
		UPDATE triggerx.job_data
		SET task_ids = ?
		WHERE job_id = ?`
)
