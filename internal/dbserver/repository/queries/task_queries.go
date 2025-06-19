package queries

// Create Queries
const (
	GetMaxTaskIDQuery = `SELECT MAX(task_id) FROM triggerx.task_data`

	// Schedulers will Create Task Data in DB before Passing the task to the Performer
	CreateTaskDataQuery = `
        INSERT INTO triggerx.task_data (
            task_id, job_id, task_definition_id, created_at, task_status
        ) VALUES (?, ?, ?, ?, ?)`
)

// Update Queries
const (
	AddTaskPerformerIDQuery = `
		UPDATE triggerx.task_data
		SET task_performer_id = ?
		WHERE task_id = ?`

	UpdateTaskExecutionDataQuery = `
		UPDATE triggerx.task_data
		SET execution_timestamp = ?, execution_tx_hash = ?, proof_of_task = ?, task_opx_cost = ?
		WHERE task_id = ?`

	UpdateTaskAttestationDataQuery = `
		UPDATE triggerx.task_data
		SET task_number = ?, task_attester_ids = ?, tp_signature = ?, ta_signature = ?, task_submission_tx_hash = ?, is_successful = ?
		WHERE task_id = ?`

	UpdateTaskFeeQuery = `
		UPDATE triggerx.task_data
		SET task_fee = ?
		WHERE task_id = ?`

	UpdateTaskNumberAndStatusQuery = `
		UPDATE triggerx.task_data
		SET task_number = ?, task_status = ?, task_submission_tx_hash = ?
		WHERE task_id = ?`
)

// Read Task Queries
const (
	GetTaskDataByIDQuery = `
        SELECT task_id, task_number, job_id, task_definition_id, created_at,
               task_opx_cost, execution_timestamp, execution_tx_hash, task_performer_id, 
			   proof_of_task, task_attester_ids,
			   tp_signature, ta_signature, task_submission_tx_hash,
			   is_successful, task_status
        FROM triggerx.task_data
        WHERE task_id = ?`

	GetTasksByJobIDQuery = `
		SELECT task_id, task_number, task_opx_cost, execution_timestamp, execution_tx_hash, task_performer_id, 
			   task_attester_ids, is_successful, task_status
		FROM triggerx.task_data
		WHERE job_id = ? ALLOW FILTERING`

	GetTaskFeeQuery = `
		SELECT task_opx_cost
		FROM triggerx.task_data
		WHERE task_id = ?`
)
