package repository

// Query constants
const (
	getMaxTaskIDQuery = `SELECT MAX(task_id) FROM triggerx.task_data`

	createTaskDataQuery = `
        INSERT INTO triggerx.task_data (
            task_id, job_id, task_definition_id, network, task_status, created_at, task_opx_predicted_cost
        ) VALUES (?, ?, ?, ?, ?, ?, ?)`

	getTaskIDsByJobIDQuery = `
		SELECT task_ids FROM triggerx.job_data 
		WHERE job_id = ?`

	getJobCostPredictionQuery = `
		SELECT job_cost_prediction
		FROM triggerx.job_data
		WHERE job_id = ?`

	getUserAddressByJobIDQuery = `
		SELECT user_address
		FROM triggerx.job_data
		WHERE job_id = ?`

	getUserTotalTasksQuery = `
		SELECT total_tasks
		FROM triggerx.user_data
		WHERE user_address = ?`

	addTaskIDToJobQuery = `
		UPDATE triggerx.job_data
		SET task_ids = ?
		WHERE job_id = ?`

	incrementUserTotalTasksQuery = `
		UPDATE triggerx.user_data
		SET total_tasks = ?, last_updated_at = ?
		WHERE user_address = ?`
)
