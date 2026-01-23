package queries

// Read Task Queries
const (
	GetTaskDataByIDQuery = `
        SELECT task_id, task_number, task_status, task_error, job_id, task_definition_id, created_at,
               executed_at, submitted_at, task_opx_predicted_cost, task_opx_actual_cost,
               execution_tx_hash, submission_tx_hash, converted_arguments, task_performer_address,
               task_attester_address, proof_of_task, is_successful, is_accepted, is_imua
        FROM triggerx.task_data
        WHERE task_id = ?`

	GetTasksByJobIDQuery = `
		SELECT task_id, task_number, task_status, task_error, task_definition_id, created_at,
               executed_at, submitted_at, task_opx_predicted_cost, task_opx_actual_cost,
               execution_tx_hash, submission_tx_hash, converted_arguments, task_performer_address,
               task_attester_address, is_successful, is_accepted
		FROM triggerx.task_data
		WHERE job_id = ? ALLOW FILTERING`

	GetTaskFeeQuery = `
		SELECT task_opx_actual_cost
		FROM triggerx.task_data
		WHERE task_id = ?`

	GetCreatedChainIDByJobIDQuery = `
        SELECT created_chain_id
        FROM triggerx.job_data
        WHERE job_id = ?`

	GetRecentTasksQuery = `
		SELECT task_id, task_number, job_id, task_definition_id, created_at,
		       task_opx_actual_cost, executed_at, execution_tx_hash, task_performer_address,
		       task_attester_address, task_status, task_error, is_imua
		FROM triggerx.task_data
		LIMIT ?`
)
