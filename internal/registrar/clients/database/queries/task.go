package queries

const (
	// Getters
	GetTaskIdsByJobId = `
		SELECT task_ids 
		FROM triggerx.job_data 
		WHERE job_id = ?`
	GetTaskCostAndJobId = `
		SELECT task_opx_predicted_cost, job_id 
		FROM triggerx.task_data 
		WHERE task_id = ?`
	GetUserIdByJobId = `
		SELECT user_id 
		FROM triggerx.job_data 
		WHERE job_id = ?`
	GetAttesterPointsAndNoOfTasks = `
		SELECT keeper_id,
			keeper_points, 
			rewards_booster,
			no_attested_tasks
		FROM triggerx.keeper_data 
		WHERE operator_id = ?
		ALLOW FILTERING`
	GetPerformerPointsAndNoOfTasks = `
		SELECT keeper_points, 
			rewards_booster,
			no_executed_tasks
		FROM triggerx.keeper_data 
		WHERE keeper_id = ?`
	GetUserPoints = `
		SELECT user_points 
		FROM triggerx.user_data 
		WHERE user_id = ?`

	// Setters
	UpdateTaskSubmissionData = `
		UPDATE triggerx.task_data 
		SET task_number = ?, 
			is_accepted = ?, 
			task_submission_tx_hash = ?, 
			task_performer_id = ?, 
			task_attester_ids = ?, 
			execution_tx_hash = ?,
			execution_timestamp = ?,
			task_opx_cost = ?,
			proof_of_task = ?
		WHERE task_id = ?`
	UpdateAttesterPointsAndNoOfTasks = `
		UPDATE triggerx.keeper_data 
		SET keeper_points = ?,
			no_attested_tasks = ?
		WHERE keeper_id = ?`
	UpdatePerformerPointsAndNoOfTasks = `
		UPDATE triggerx.keeper_data 
		SET keeper_points = ?,
			no_executed_tasks = ?
		WHERE keeper_id = ?`
	UpdateUserPoints = `
		UPDATE triggerx.user_data 
		SET user_points = ?, last_updated_at = ?
		WHERE user_id = ?`

)