package queries

// Create Queries
const (
	CreateTimeJobDataQuery = `
			INSERT INTO triggerx.time_job_data (
				job_id, time_frame, recurring, 
				next_execution_timestamp, schedule_type,
				time_interval, cron_expression, specific_schedule, 
				target_chain_id, target_contract_address, target_function, 
				abi, arg_type, arguments, dynamic_arguments_script_url,
				is_completed, is_active
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	// 17 values to be inserted, so 17 ?s
	
	CreateEventJobDataQuery = `
			INSERT INTO triggerx.event_job_data (
				job_id, time_frame, recurring,
				trigger_chain_id, trigger_contract_address, trigger_event,
				target_chain_id, target_contract_address, target_function,
				abi, arg_type, arguments, dynamic_arguments_script_url,
				is_completed, is_active
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	// 15 values to be inserted, so 15 ?s

	CreateConditionJobDataQuery = `
			INSERT INTO triggerx.condition_job_data (
				job_id, time_frame, recurring,
				condition_type, upper_limit, lower_limit,
				value_source_type, value_source_url,
				target_chain_id, target_contract_address, target_function,
				abi, arg_type, arguments, dynamic_arguments_script_url,
				is_completed, is_active
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	// 17 values to be inserted, so 17 ?s
)

// Write Queries
const (
	CompleteTimeJobStatusQuery = `
			UPDATE triggerx.time_job_data
			SET is_completed = true
			WHERE job_id = ?`

	UpdateTimeJobStatusQuery = `
			UPDATE triggerx.time_job_data
			SET is_active = ?
			WHERE job_id = ?`

	CompleteEventJobStatusQuery = `
			UPDATE triggerx.event_job_data
			SET is_completed = true
			WHERE job_id = ?`

	UpdateEventJobStatusQuery = `
			UPDATE triggerx.event_job_data
			SET is_active = ?
			WHERE job_id = ?`

	CompleteConditionJobStatusQuery = `
			UPDATE triggerx.condition_job_data
			SET is_completed = true
			WHERE job_id = ?`

	UpdateConditionJobStatusQuery = `
			UPDATE triggerx.condition_job_data
			SET is_active = ?
			WHERE job_id = ?`
)

// Read Queries
const (
	GetTimeJobDataByJobIDQuery = `
			SELECT job_id, time_frame, recurring, 
				next_execution_timestamp, schedule_type,
				time_interval, cron_expression, specific_schedule, 
				target_chain_id, target_contract_address, target_function, 
				abi, arg_type, arguments, dynamic_arguments_script_url,
				is_completed, is_active
			FROM triggerx.time_job_data
			WHERE job_id = ?`

	GetEventJobDataByJobIDQuery = `
			SELECT job_id, time_frame, recurring,
				trigger_chain_id, trigger_contract_address, trigger_event,
				target_chain_id, target_contract_address, target_function,
				abi, arg_type, arguments, dynamic_arguments_script_url,
				is_completed, is_active
			FROM triggerx.event_job_data
			WHERE job_id = ?`

	GetConditionJobDataByJobIDQuery = `
			SELECT job_id, time_frame, recurring,
				condition_type, upper_limit, lower_limit,
				value_source_type, value_source_url,
				target_chain_id, target_contract_address, target_function,
				abi, arg_type, arguments, dynamic_arguments_script_url,
				is_completed, is_active
			FROM triggerx.condition_job_data
			WHERE job_id = ?`

	GetTimeJobsByNextExecutionTimestampQuery = `
			SELECT job_id, time_frame, recurring, next_execution_timestamp,
				time_interval, cron_expression, specific_schedule, 
				target_chain_id, target_contract_address, target_function, 
				abi, arg_type, arguments, dynamic_arguments_script_url,
				is_completed, is_active
			FROM triggerx.time_job_data
			WHERE next_execution_timestamp <= ? ALLOW FILTERING`
)