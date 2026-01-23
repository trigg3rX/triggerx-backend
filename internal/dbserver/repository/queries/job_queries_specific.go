package queries

// Create Queries
const (
	CreateTimeJobDataQuery = `
			INSERT INTO triggerx.time_job_data (
				job_id, task_definition_id, schedule_type, time_interval, cron_expression, specific_schedule,
				timezone, next_execution_timestamp, target_chain_id, target_contract_address, target_function,
				abi, arg_type, arguments, dynamic_arguments_script_url, agent_script_url, agent_script_language,
				agent_script_hash, agent_target_chain_id, max_execution_time, challenge_period, is_active,
				last_executed_at, expiration_time
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	// 24 values to be inserted, so 24 ?s

	CreateEventJobDataQuery = `
			INSERT INTO triggerx.event_job_data (
				job_id, task_definition_id, recurring, trigger_chain_id, trigger_contract_address, trigger_event,
				event_filter_para_name, event_filter_value, target_chain_id, target_contract_address, target_function,
				abi, arg_type, arguments, dynamic_arguments_script_url, agent_script_url, agent_script_language,
				agent_script_hash, agent_target_chain_id, max_execution_time, challenge_period, is_active,
				last_executed_at, expiration_time
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	// 24 values to be inserted, so 24 ?s

	CreateConditionJobDataQuery = `
			INSERT INTO triggerx.condition_job_data (
				job_id, task_definition_id, recurring, condition_type, upper_limit, lower_limit,
				value_source_type, value_source_url, selected_key_route, target_chain_id, target_contract_address,
				target_function, abi, arg_type, arguments, dynamic_arguments_script_url, agent_script_url,
				agent_script_language, agent_script_hash, agent_target_chain_id, max_execution_time, challenge_period,
				is_active, last_executed_at, expiration_time
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	// 25 values to be inserted, so 25 ?s
)

// Write Queries
const (
	UpdateTimeJobStatusQuery = `
			UPDATE triggerx.time_job_data
			SET is_active = ?
			WHERE job_id = ?`

	UpdateEventJobStatusQuery = `
			UPDATE triggerx.event_job_data
			SET is_active = ?
			WHERE job_id = ?`

	UpdateConditionJobStatusQuery = `
			UPDATE triggerx.condition_job_data
			SET is_active = ?
			WHERE job_id = ?`
)

// Read Queries
const (
	IsJobImuaQuery = `
			SELECT is_imua
			FROM triggerx.job_data
			WHERE job_id = ?`

	GetTimeJobDataByJobIDQuery = `
			SELECT job_id, task_definition_id, schedule_type, time_interval, cron_expression, specific_schedule,
				timezone, next_execution_timestamp, target_chain_id, target_contract_address, target_function,
				abi, arg_type, arguments, dynamic_arguments_script_url, agent_script_url, agent_script_language,
				agent_script_hash, agent_target_chain_id, max_execution_time, challenge_period, is_active,
				last_executed_at, expiration_time
			FROM triggerx.time_job_data
			WHERE job_id = ?`

	GetEventJobDataByJobIDQuery = `
			SELECT job_id, task_definition_id, recurring, trigger_chain_id, trigger_contract_address, trigger_event,
				event_filter_para_name, event_filter_value, target_chain_id, target_contract_address, target_function,
				abi, arg_type, arguments, dynamic_arguments_script_url, agent_script_url, agent_script_language,
				agent_script_hash, agent_target_chain_id, max_execution_time, challenge_period, is_active,
				last_executed_at, expiration_time
			FROM triggerx.event_job_data
			WHERE job_id = ?`

	GetConditionJobDataByJobIDQuery = `
			SELECT job_id, task_definition_id, recurring, condition_type, upper_limit, lower_limit,
				value_source_type, value_source_url, selected_key_route, target_chain_id, target_contract_address,
				target_function, abi, arg_type, arguments, dynamic_arguments_script_url, agent_script_url,
				agent_script_language, agent_script_hash, agent_target_chain_id, max_execution_time, challenge_period,
				is_active, last_executed_at, expiration_time
			FROM triggerx.condition_job_data
			WHERE job_id = ?`
)
