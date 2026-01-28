package repository

// Read Queries
const (
        GetJobCostActual = `
        SELECT job_cost_actual 
        FROM triggerx.job_data 
        WHERE job_id = ?`
	GetUserPoints = `
        SELECT user_points 
        FROM triggerx.user_data 
        WHERE user_address = ?`
        GetKeeperAddressByOperatorIDAndNetwork = `
        SELECT keeper_address 
        FROM triggerx.keeper_data 
        WHERE operator_id = ? AND network = ?
        ALLOW FILTERING`
	GetKeeperPointsAndNoOfTasks = `
        SELECT keeper_points, no_executed_tasks, no_attested_tasks
        FROM triggerx.keeper_data 
        WHERE keeper_address = ?`
        GetJobIDByTaskID = `
        SELECT job_id 
        FROM triggerx.task_data 
        WHERE task_id = ?`
	GetUserAddressByJobId = `
        SELECT user_address 
        FROM triggerx.job_data 
        WHERE job_id = ?`
	GetUserEmailByUserAddress = `
        SELECT email_id 
        FROM triggerx.user_data 
        WHERE user_address = ?`
)

// Write Queries
const (
        UpdateTaskExecutionData = `
        UPDATE triggerx.task_data 
        SET task_performer_address = ?,
            proof_of_task = ?,
            is_successful = ?,
            task_error = ?,
            task_status = ?,
            execution_tx_hash = ?,
            executed_at = ?,
            task_opx_actual_cost = ?,
            converted_arguments = ?
        WHERE task_id = ?`
        UpdateTimeJobLastExecutedAt = `
        UPDATE triggerx.time_job_data
        SET last_executed_at = ?
        WHERE job_id = ?`
	UpdateEventJobLastExecutedAt = `
        UPDATE triggerx.event_job_data
        SET last_executed_at = ?
        WHERE job_id = ?`
	UpdateConditionJobLastExecutedAt = `
        UPDATE triggerx.condition_job_data
        SET last_executed_at = ?
        WHERE job_id = ?`
        UpdateScriptStorageQuery = `
        UPDATE triggerx.script_storage
        SET storage_value = ?,
            updated_at = ?
        WHERE job_id = ? AND storage_key = ?`
        UpdateJobCostActual = `
        UPDATE triggerx.job_data
        SET job_cost_actual = ?
        WHERE job_id = ?`
        UpdateUserPoints = `
        UPDATE triggerx.user_data 
        SET user_points = ?, last_updated_at = ?
        WHERE user_address = ?`
        UpdateKeeperPointsAndNoOfTasks = `
        UPDATE triggerx.keeper_data 
        SET keeper_points = ?,
            no_executed_tasks = ?,
            no_attested_tasks = ?
        WHERE keeper_address = ?`
	UpdateTaskSubmissionData = `
        UPDATE triggerx.task_data 
        SET task_number = ?, 
            is_accepted = ?, 
            task_status = 'completed',
            submission_tx_hash = ?, 
            task_attester_address = ?, 
            submitted_at = ?
        WHERE task_id = ?`
        UpdateTaskAttestationTimeoutFailure = `
        UPDATE triggerx.task_data 
        SET is_accepted = false,
            task_status = 'failed',
            task_error = ?
        WHERE task_id = ?`
)
