package queries

// Read Task Queries
const (
	GetTaskDataByIDQuery = `
        SELECT task_id, task_number, job_id, task_definition_id, created_at,
               task_opx_cost, execution_timestamp, execution_tx_hash, task_performer_id, 
			   proof_of_task, converted_arguments, task_attester_ids,
			   tp_signature, ta_signature, task_submission_tx_hash,
			   is_successful, task_status, task_error, is_imua
        FROM triggerx.task_data
        WHERE task_id = ?`

	GetTasksByJobIDQuery = `
		SELECT task_id, task_number, task_opx_cost, execution_timestamp, execution_tx_hash, task_performer_id, 
			   task_attester_ids, is_accepted, task_status, task_error, converted_arguments
		FROM triggerx.task_data
		WHERE job_id = ? ALLOW FILTERING`

	GetTaskFeeQuery = `
		SELECT task_opx_cost
		FROM triggerx.task_data
		WHERE task_id = ?`

	GetCreatedChainIDByJobIDQuery = `
        SELECT created_chain_id
        FROM triggerx.job_data
        WHERE job_id = ?`

	GetRecentTasksQuery = `
		SELECT task_id, task_number, job_id, task_definition_id, created_at,
		       task_opx_cost, execution_timestamp, execution_tx_hash, task_performer_id,
		       task_attester_ids, task_status, task_error, is_imua
		FROM triggerx.task_data
		LIMIT ?`
)
