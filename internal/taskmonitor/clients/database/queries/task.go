package queries

const (
	// Getters
	GetKeeperAddressByConsensusAddress = `
        SELECT keeper_address 
        FROM triggerx.keeper_data 
        WHERE consensus_address = ? 
        ALLOW FILTERING`
	GetConsensusAddressByKeeperAddress = `
        SELECT consensus_address 
        FROM triggerx.keeper_data 
        WHERE keeper_address = ? 
        ALLOW FILTERING`
	GetKeeperAddressByOperatorID = `
        SELECT keeper_address 
        FROM triggerx.keeper_data 
        WHERE operator_id = ? 
        ALLOW FILTERING`
	GetTaskCostAndJobId = `
        SELECT task_opx_predicted_cost, job_id 
        FROM triggerx.task_data 
        WHERE task_id = ?`
	GetUserAddressByJobId = `
        SELECT user_address 
        FROM triggerx.job_data 
        WHERE job_id = ?`
	GetAttesterPointsAndNoOfTasks = `
        SELECT keeper_address,
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
        WHERE keeper_address = ?`
	GetUserPoints = `
        SELECT user_points, total_tasks 
        FROM triggerx.user_data 
        WHERE user_address = ?`
	GetJobCostActual = `
        SELECT job_cost_actual 
        FROM triggerx.job_data 
        WHERE job_id = ?`
	// New getters for notification/email lookup
	GetUserEmailByUserAddress = `
        SELECT email_id 
        FROM triggerx.user_data 
        WHERE user_address = ?`
	GetTaskStatusByID = `
        SELECT task_status
        FROM triggerx.task_data
        WHERE task_id = ?`

	// Setters
	UpdateTaskSubmissionData = `
        UPDATE triggerx.task_data 
        SET task_number = ?, 
            is_accepted = ?, 
            is_successful = true,
            task_status = 'completed',
            submission_tx_hash = ?, 
            task_performer_address = ?, 
            task_attester_address = ?, 
            execution_tx_hash = ?,
            executed_at = ?,
            task_opx_actual_cost = ?,
            proof_of_task = ?,
            converted_arguments = ?
        WHERE task_id = ?`
	UpdateTaskFailed = `
        UPDATE triggerx.task_data 
        SET is_successful = false,
            task_status = 'failed'
        WHERE task_id = ?`
	UpdateTaskError = `
        UPDATE triggerx.task_data 
        SET is_successful = false,
            task_status = 'failed',
            task_error = ?
        WHERE task_id = ?`
	// UpdateTaskAggregatorFailed - Task failed (execution or aggregator submission failed)
	// Parameters: execution_successful, aggregator_submitted, execution_tx_hash, task_error, proof_of_task, task_id
	UpdateTaskAggregatorFailed = `
        UPDATE triggerx.task_data 
        SET task_status = 'failed',
            is_successful = false,
            task_error = ?,
            execution_tx_hash = ?,
            proof_of_task = ?
        WHERE task_id = ?`
	// UpdateTaskAggregatorSubmitted - Task succeeded (both execution and aggregator submission)
	// The task is now pending on-chain confirmation
	// Parameters: execution_tx_hash, proof_of_task, task_id
	UpdateTaskAggregatorSubmitted = `
        UPDATE triggerx.task_data 
        SET task_status = 'pending_confirmation',
            is_successful = true,
            execution_tx_hash = ?,
            proof_of_task = ?
        WHERE task_id = ?`
	UpdateAttesterPointsAndNoOfTasks = `
        UPDATE triggerx.keeper_data 
        SET keeper_points = ?,
            no_attested_tasks = ?
        WHERE keeper_address = ?`
	UpdatePerformerPointsAndNoOfTasks = `
        UPDATE triggerx.keeper_data 
        SET keeper_points = ?,
            no_executed_tasks = ?
        WHERE keeper_address = ?`
	UpdateUserPoints = `
        UPDATE triggerx.user_data 
        SET user_points = ?, total_tasks = ?, last_updated_at = ?
        WHERE user_address = ?`
	UpdateJobCostActual = `
        UPDATE triggerx.job_data
        SET job_cost_actual = ?
        WHERE job_id = ?`

	// Custom script storage queries (TaskDefinitionID = 7)
	UpsertScriptStorageQuery = `
        UPDATE triggerx.script_storage
        SET storage_value = ?,
            updated_at = ?
        WHERE job_id = ? AND storage_key = ?`

	GetJobIDByTaskIDQuery = `
        SELECT job_id
        FROM triggerx.task_data
        WHERE task_id = ?`

	GetTaskDefinitionIDQuery = `
        SELECT task_definition_id
        FROM triggerx.task_data
        WHERE task_id = ?`
)
