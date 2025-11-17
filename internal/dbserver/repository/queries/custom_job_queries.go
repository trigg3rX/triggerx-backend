package queries

// Custom Job Queries (TaskDefinitionID = 7)

// Create Queries
const (
	CreateCustomJobQuery = `
		INSERT INTO triggerx.custom_jobs (
			job_id, task_definition_id, recurring, custom_script_url, time_interval,
			target_chain_id, is_completed, is_active, created_at, updated_at, last_executed_at,
			expiration_time, script_language, script_hash, next_execution_time,
			max_execution_time, challenge_period
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
)

// Read Queries
const (
	GetCustomJobByIDQuery = `
		SELECT job_id, task_definition_id, recurring, custom_script_url, time_interval,
			target_chain_id, is_completed, is_active, created_at, updated_at, last_executed_at,
			expiration_time, script_language, script_hash, next_execution_time,
			max_execution_time, challenge_period
		FROM triggerx.custom_jobs
		WHERE job_id = ?`

	GetCustomJobsDueForExecutionQuery = `
		SELECT job_id, task_definition_id, recurring, custom_script_url, time_interval,
			target_chain_id, is_completed, is_active, created_at, updated_at, last_executed_at,
			expiration_time, script_language, script_hash, next_execution_time,
			max_execution_time, challenge_period
		FROM triggerx.custom_jobs
		WHERE is_active = ? ALLOW FILTERING`

	GetActiveCustomJobsQuery = `
		SELECT job_id, task_definition_id, recurring, custom_script_url, time_interval,
			target_chain_id, is_completed, is_active, created_at, updated_at, last_executed_at,
			expiration_time, script_language, script_hash, next_execution_time,
			max_execution_time, challenge_period
		FROM triggerx.custom_jobs
		WHERE is_active = ? ALLOW FILTERING`
)

// Update Queries
const (
	UpdateCustomJobNextExecutionQuery = `
		UPDATE triggerx.custom_jobs
		SET next_execution_time = ?, last_executed_at = ?, updated_at = ?
		WHERE job_id = ?`

	UpdateCustomJobStatusQuery = `
		UPDATE triggerx.custom_jobs
		SET is_active = ?, is_completed = ?, updated_at = ?
		WHERE job_id = ?`

	UpdateCustomJobQuery = `
		UPDATE triggerx.custom_jobs
		SET custom_script_url = ?, time_interval = ?, recurring = ?, updated_at = ?
		WHERE job_id = ?`
)

// Delete Queries
const (
	DeleteCustomJobQuery = `
		DELETE FROM triggerx.custom_jobs
		WHERE job_id = ?`
)

// Custom Script Execution Queries
const (
	CreateExecutionRecordQuery = `
		INSERT INTO triggerx.custom_script_executions (
			execution_id, job_id, task_id, scheduled_time, actual_time, performer_address,
			input_timestamp, input_storage, input_hash, should_execute, target_contract,
			calldata, output_hash, execution_metadata, script_hash, signature,
			tx_hash, execution_status, execution_error, verification_status,
			challenge_deadline, is_challenged, challenge_count, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	GetExecutionByIDQuery = `
		SELECT execution_id, job_id, task_id, scheduled_time, actual_time, performer_address,
			input_timestamp, input_storage, input_hash, should_execute, target_contract,
			calldata, output_hash, execution_metadata, script_hash, signature,
			tx_hash, execution_status, execution_error, verification_status,
			challenge_deadline, is_challenged, challenge_count, created_at
		FROM triggerx.custom_script_executions
		WHERE execution_id = ?`

	GetExecutionsByJobIDQuery = `
		SELECT execution_id, job_id, task_id, scheduled_time, actual_time, performer_address,
			input_timestamp, input_storage, input_hash, should_execute, target_contract,
			calldata, output_hash, execution_metadata, script_hash, signature,
			tx_hash, execution_status, execution_error, verification_status,
			challenge_deadline, is_challenged, challenge_count, created_at
		FROM triggerx.custom_script_executions
		WHERE job_id = ? ALLOW FILTERING`

	GetExecutionsByTaskIDQuery = `
		SELECT execution_id, job_id, task_id, scheduled_time, actual_time, performer_address,
			input_timestamp, input_storage, input_hash, should_execute, target_contract,
			calldata, output_hash, execution_metadata, script_hash, signature,
			tx_hash, execution_status, execution_error, verification_status,
			challenge_deadline, is_challenged, challenge_count, created_at
		FROM triggerx.custom_script_executions
		WHERE task_id = ? ALLOW FILTERING`

	UpdateExecutionTxHashQuery = `
		UPDATE triggerx.custom_script_executions
		SET tx_hash = ?, execution_status = ?
		WHERE execution_id = ?`

	UpdateExecutionVerificationStatusQuery = `
		UPDATE triggerx.custom_script_executions
		SET verification_status = ?
		WHERE execution_id = ?`

	UpdateExecutionChallengeStatusQuery = `
		UPDATE triggerx.custom_script_executions
		SET is_challenged = ?, challenge_count = ?
		WHERE execution_id = ?`
)

// Script Storage Queries
const (
	GetStorageByJobIDQuery = `
		SELECT storage_key, storage_value, updated_at
		FROM triggerx.script_storage
		WHERE job_id = ?`

	GetStorageValueQuery = `
		SELECT storage_value
		FROM triggerx.script_storage
		WHERE job_id = ? AND storage_key = ?`

	UpsertStorageQuery = `
		INSERT INTO triggerx.script_storage (job_id, storage_key, storage_value, updated_at)
		VALUES (?, ?, ?, ?)`

	DeleteStorageKeyQuery = `
		DELETE FROM triggerx.script_storage
		WHERE job_id = ? AND storage_key = ?`

	DeleteAllStorageForJobQuery = `
		DELETE FROM triggerx.script_storage
		WHERE job_id = ?`
)

// Challenge Queries
const (
	CreateChallengeQuery = `
		INSERT INTO triggerx.execution_challenges (
			challenge_id, execution_id, challenger_address, challenge_reason,
			challenger_output_hash, challenger_should_execute, challenger_target_contract,
			challenger_calldata, challenger_signature, resolution_status,
			validator_count, approve_count, reject_count, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	GetChallengesByExecutionIDQuery = `
		SELECT challenge_id, execution_id, challenger_address, challenge_reason,
			challenger_output_hash, challenger_should_execute, challenger_target_contract,
			challenger_calldata, challenger_signature, resolution_status,
			resolution_time, validator_count, approve_count, reject_count, created_at
		FROM triggerx.execution_challenges
		WHERE execution_id = ? ALLOW FILTERING`

	UpdateChallengeResolutionQuery = `
		UPDATE triggerx.execution_challenges
		SET resolution_status = ?, resolution_time = ?, validator_count = ?,
			approve_count = ?, reject_count = ?
		WHERE challenge_id = ?`
)
