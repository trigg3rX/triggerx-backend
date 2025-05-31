package queries

const (
	SelectUserTaskCountByAddressQuery = `
			SELECT COUNT(*) FROM triggerx.task_data 
			WHERE user_address = ? AND execution_timestamp IS NOT NULL ALLOW FILTERING`

	UpdateTaskFeeQuery = `
			UPDATE triggerx.task_data
			SET task_fee = ?
			WHERE task_id = ?`

	SelectTaskDataByIDQuery = `
        SELECT task_id, job_id, task_definition_id, created_at,
               task_fee, execution_timestamp, execution_tx_hash, task_performer_id, 
			   proof_of_task, action_data_cid, task_attester_ids,
			   is_approved, tp_signature, ta_signature, task_submission_tx_hash,
			   is_successful
        FROM triggerx.task_data
        WHERE task_id = ?`

	SelectMaxTaskIDQuery = `
			SELECT MAX(task_id) FROM triggerx.task_data`

	InsertTaskDataQuery = `
        INSERT INTO triggerx.task_data (
            task_id, job_id, task_definition_id, created_at,
            task_performer_id, is_approved
        ) VALUES (?, ?, ?, ?, ?, ?)`
)