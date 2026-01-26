package repository

// Read Queries
const (
	getActiveEventJobsQuery = `
        SELECT job_id, expiration_time
        FROM triggerx.event_job_data
        WHERE is_active = true
        ALLOW FILTERING`

	getEventJobByJobIDQuery = `
        SELECT job_id, task_definition_id, network, recurring, trigger_chain_id, trigger_contract_address, trigger_event,
            event_filter_para_name, event_filter_value, target_chain_id, target_contract_address, target_function,
            abi, arg_type, arguments, execution_script_url, execution_script_language, execution_script_hash,
            max_execution_time, challenge_period, is_active, last_executed_at, expiration_time
        FROM triggerx.event_job_data
        WHERE job_id = ?`

	getActiveConditionJobsQuery = `
        SELECT job_id, expiration_time
        FROM triggerx.condition_job_data
        WHERE is_active = true
        ALLOW FILTERING`

	getConditionJobByJobIDQuery = `
        SELECT job_id, task_definition_id, network, recurring, condition_type, upper_limit, lower_limit,
            value_source_type, value_source_url, selected_key_route, target_chain_id, target_contract_address,
            target_function, abi, arg_type, arguments, execution_script_url, execution_script_language,
            execution_script_hash, max_execution_time, challenge_period,
            is_active, last_executed_at, expiration_time
        FROM triggerx.condition_job_data
        WHERE job_id = ?`
)

// Write Queries
const (
	updateConditionJobStatusQuery = `
        UPDATE triggerx.condition_job_data
        SET is_active = ?
        WHERE job_id = ?`

	updateEventJobStatusQuery = `
        UPDATE triggerx.event_job_data
        SET is_active = ?
        WHERE job_id = ?`
)
