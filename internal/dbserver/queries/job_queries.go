package queries

const (
	SelectMaxJobIDQuery = `SELECT MAX(job_id) FROM triggerx.job_data`

	InsertJobDataQuery = `
			INSERT INTO triggerx.job_data (
				job_id, job_title, task_definition_id, user_id, link_job_id, chain_status,
				custom, time_frame, recurring, status, job_cost_prediction, task_ids,
				created_at, updated_at, last_executed_at, timezone
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	InsertTimeJobDataQuery = `
			INSERT INTO triggerx.time_job_data (
				job_id, time_frame, recurring, time_interval,
				target_chain_id, target_contract_address, target_function,
				abi, arg_type, arguments, dynamic_arguments_script_ipfs_url
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	InsertEventJobDataQuery = `
			INSERT INTO triggerx.event_job_data (
				job_id, time_frame, recurring,
				trigger_chain_id, trigger_contract_address, trigger_event,
				target_chain_id, target_contract_address, target_function,
				abi, arg_type, arguments, dynamic_arguments_script_ipfs_url
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	InsertConditionJobDataQuery = `
			INSERT INTO triggerx.condition_job_data (
				job_id, time_frame, recurring,
				condition_type, upper_limit, lower_limit,
				value_source_type, value_source_url,
				target_chain_id, target_contract_address, target_function,
				abi, arg_type, arguments, dynamic_arguments_script_ipfs_url
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	DeleteTimeJobDataQuery = `
			DELETE FROM triggerx.time_job_data 
			WHERE job_id = ?`

	DeleteEventJobDataQuery = `
			DELETE FROM triggerx.event_job_data 
			WHERE job_id = ?`

	DeleteConditionJobDataQuery = `
			DELETE FROM triggerx.condition_job_data 
			WHERE job_id = ?`

	DeleteJobDataQuery = `
			DELETE FROM triggerx.job_data 
			WHERE job_id = ?`

	SelectJobDataByJobIDQuery = `
			SELECT job_id, task_definition_id
			FROM triggerx.job_data 
			WHERE job_id = ?`

	SelectTaskDefinitionIDByJobIDQuery = `
			SELECT task_definition_id FROM triggerx.job_data 
			WHERE job_id = ?`

	UpdateJobDataLastUpdatedAtQuery = `
			UPDATE triggerx.job_data 
			SET updated_at = ?
			WHERE job_id = ?`

	UpdateTimeJobDataByUserQuery = `
			UPDATE triggerx.time_job_data 
			SET time_frame = ?, recurring = ?, updated_at = ?
			WHERE job_id = ?`

	UpdateEventJobDataByUserQuery = `
			UPDATE triggerx.event_job_data 
			SET time_frame = ?, recurring = ?, updated_at = ?
			WHERE job_id = ?`

	UpdateConditionJobDataByUserQuery = `
			UPDATE triggerx.condition_job_data 
			SET time_frame = ?, recurring = ?, updated_at = ?
			WHERE job_id = ?`

	UpdateJobStatusQuery = `
			UPDATE triggerx.job_data 
			SET status = ?, updated_at = ?
			WHERE job_id = ?`

	UpdateJobLastExecutedAtQuery = `
			UPDATE triggerx.job_data 
			SET last_executed_at = ?, updated_at = ?
			WHERE job_id = ?`

	UpdateTimeJobDataLastExecutedAtQuery = `
			UPDATE triggerx.time_job_data 
			SET last_executed_at = ?, updated_at = ?
			WHERE job_id = ?`

	UpdateEventJobDataLastExecutedAtQuery = `
			UPDATE triggerx.event_job_data 
			SET last_executed_at = ?, updated_at = ?
			WHERE job_id = ?`
			
	UpdateConditionJobDataLastExecutedAtQuery = `
			UPDATE triggerx.condition_job_data 
			SET last_executed_at = ?, updated_at = ?
			WHERE job_id = ?`

	SelectCompleteJobDataByJobIDQuery = `
			SELECT job_id, job_title, task_definition_id, user_id, link_job_id, chain_status,
				custom, time_frame, recurring, status, job_cost_prediction, task_ids
			FROM triggerx.job_data 
			WHERE job_id = ?`

	SelectCompleteTimeJobDataByJobIDQuery = `
			SELECT job_id, time_frame, recurring, time_interval,
				target_chain_id, target_contract_address, target_function,
				abi, arg_type, arguments, dynamic_arguments_script_ipfs_url
			FROM triggerx.time_job_data 
			WHERE job_id = ?`

	SelectCompleteEventJobDataByJobIDQuery = `
			SELECT job_id, time_frame, recurring,
				trigger_chain_id, trigger_contract_address, trigger_event,
				target_chain_id, target_contract_address, target_function,
				abi, arg_type, arguments, dynamic_arguments_script_ipfs_url
			FROM triggerx.event_job_data 
			WHERE job_id = ?`

	SelectCompleteConditionJobDataByJobIDQuery = `
			SELECT job_id, time_frame, recurring,
				condition_type, upper_limit, lower_limit,
				value_source_type, value_source_url,
				target_chain_id, target_contract_address, target_function,
				abi, arg_type, arguments, dynamic_arguments_script_ipfs_url
			FROM triggerx.condition_job_data 
			WHERE job_id = ?`

	SelectUserJobCountByAddressQuery = `
			SELECT COUNT(*) FROM triggerx.job_data WHERE user_address = ? ALLOW FILTERING`

	SelectTimeBasedJobsQuery = `
			SELECT job_id, time_frame, recurring, schedule_type, time_interval, 
				cron_expression, specific_schedule, timezone, next_execution_timestamp,
				target_chain_id, target_contract_address, target_function,
				abi, arg_type, arguments, dynamic_arguments_script_ipfs_url 
			FROM triggerx.time_job_data 
			WHERE next_execution_timestamp <= ? ALLOW FILTERING`
)
