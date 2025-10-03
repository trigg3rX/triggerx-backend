# Types

## Database Schema

### User Data

- `user_id`: Unique auto-incremented identifier for the user (internal use only, not exposed to SDK/UI).
- `user_address`: Wallet address of the user (used for user identification in UI).
- `email_id`: Email address of the user (used for notifications).
- `job_ids`: List of job IDs (big integers) associated with the user.
- `opx_consumed`: Total OPX tokens consumed by the user (big integer).
- `total_jobs`: Total number of jobs created by the user.
- `total_tasks`: Total number of tasks executed for the user's jobs.
- `created_at`: Timestamp when the user was created.
- `last_updated_at`: Timestamp of the last update to the user record.

### Job Data

- `job_id`: Unique identifier for the job (big integer).
- `job_title`: Title or name of the job.
- `task_definition_id`: Identifier for the task definition (integer, see Task Data for mapping).
- `created_chain_id`: Chain ID where the job was created.
- `user_id`: ID of the user who created the job.
- `link_job_id`: ID of a linked job (big integer). If not linked, may be zero or a sentinel value.
- `chain_status`: Status of the job in a chain of jobs (0 = None, 1 = Chain Head, 2 = Chain Block).
- `timezone`: Timezone of the user who created the job.
- `is_imua`: Boolean indicating if the job is IMUA (special job type).
- `job_type`: Type of the job (e.g., "sdk", "frontend", "contract", "template").
- `time_frame`: Duration (in seconds or other unit) for which the job is valid.
- `recurring`: Boolean indicating if the job is recurring.
- `status`: Status of the job (e.g., "created", "running", "completed", "failed", "expired").
- `job_cost_prediction`: Predicted cost of the job (big integer, for estimation only).
- `job_cost_actual`: Actual cost incurred for the job (big integer).
- `task_ids`: List of task IDs (int64) executed for the job.
- `created_at`: Timestamp when the job was created.
- `updated_at`: Timestamp of the last update to the job.
- `last_executed_at`: Timestamp of the last execution of the job.

### Time Job Data

- `job_id`: Unique identifier for the job (big integer).
- `task_definition_id`: Task definition ID (integer).
- `schedule_type`: Type of schedule (e.g., "interval", "cron", "specific").
- `time_interval`: Time interval for execution (int64, in seconds or as defined).
- `cron_expression`: Cron expression for scheduling (string).
- `specific_schedule`: Specific schedule details (string, e.g., ISO8601).
- `next_execution_timestamp`: Timestamp for the next scheduled execution.
- `target_chain_id`: Chain ID where the target contract resides.
- `target_contract_address`: Address of the target contract.
- `target_function`: Name of the function to be called on the target contract.
- `abi`: ABI of the target function (string).
- `arg_type`: Type of arguments (integer, e.g., 0 = None, 1 = Static, 2 = Dynamic).
- `arguments`: List of arguments (array of strings) for the function.
- `dynamic_arguments_script_url`: URL to the script for dynamic arguments (string).
- `is_completed`: Boolean indicating if the job is completed.
- `last_executed_at`: Timestamp of the last execution.
- `expiration_time`: Timestamp when the job expires.

### Event Job Data

- `job_id`: Unique identifier for the job (big integer).
- `task_definition_id`: Task definition ID (integer).
- `recurring`: Boolean indicating if the job is recurring.
- `trigger_chain_id`: Chain ID where the trigger event is monitored.
- `trigger_contract_address`: Address of the contract emitting the trigger event.
- `trigger_event`: Name or signature of the trigger event.
- `trigger_event_filter_para_name`: Name of the parameter to filter the event (string).
- `trigger_event_filter_value`: Value to filter the event parameter (string).
- `target_chain_id`: Chain ID where the target contract resides.
- `target_contract_address`: Address of the target contract.
- `target_function`: Name of the function to be called on the target contract.
- `abi`: ABI of the target function (string).
- `arg_type`: Type of arguments (integer, e.g., 0 = None, 1 = Static, 2 = Dynamic).
- `arguments`: List of arguments (array of strings) for the function.
- `dynamic_arguments_script_url`: URL to the script for dynamic arguments (string).
- `is_completed`: Boolean indicating if the job is completed.
- `last_executed_at`: Timestamp of the last execution.
- `expiration_time`: Timestamp when the job expires.

### Condition Job Data

- `job_id`: Unique identifier for the job (big integer).
- `task_definition_id`: Task definition ID (integer).
- `recurring`: Boolean indicating if the job is recurring.
- `condition_type`: Type of condition (string, e.g., "greater_than", "less_than", "equal_to", "not_equal_to").
- `upper_limit`: Upper limit for the condition (float64).
- `lower_limit`: Lower limit for the condition (float64).
- `value_source_type`: Type of value source (string, e.g., "api", "oracle").
- `value_source_url`: URL to fetch the value for the condition (string).
- `selected_key_route`: Key path or selector for extracting value from the source (string).
- `target_chain_id`: Chain ID where the target contract resides.
- `target_contract_address`: Address of the target contract.
- `target_function`: Name of the function to be called on the target contract.
- `abi`: ABI of the target function (string).
- `arg_type`: Type of arguments (integer, e.g., 0 = None, 1 = Static, 2 = Dynamic).
- `arguments`: List of arguments (array of strings) for the function.
- `dynamic_arguments_script_url`: URL to the script for dynamic arguments (string).
- `is_completed`: Boolean indicating if the job is completed.
- `last_executed_at`: Timestamp of the last execution.
- `expiration_time`: Timestamp when the job expires.

### Task Data

- `task_id`: Unique auto-incremented identifier for the task (int64).
- `task_number`: Task number as tracked on the Attestation Center contract (int64).
- `job_id`: ID of the job this task belongs to (big integer).
- `task_definition_id`: Task definition ID (integer).
- `created_at`: Timestamp when the task was created.
- `task_opx_predicted_cost`: Predicted OPX cost for the task (big integer).
- `task_opx_actual_cost`: Actual OPX cost incurred for the task (big integer).
- `execution_timestamp`: Timestamp when the task was executed.
- `execution_tx_hash`: Transaction hash of the execution.
- `task_performer_id`: ID of the keeper who performed the task (int64).
- `task_attester_ids`: List of attester IDs (array of int64) who attested the task.
- `proof_of_task`: Proof of task (e.g., TLS proof, string).
- `submission_tx_hash`: Transaction hash of the task submission.
- `is_successful`: Boolean indicating if the task was successfully executed.
- `is_accepted`: Boolean indicating if the task was accepted.
- `is_imua`: Boolean indicating if the task is IMUA (special type).

### Keeper Data

- `keeper_id`: Unique auto-incremented identifier for the keeper (int64).
- `keeper_name`: Name of the keeper (string).
- `keeper_address`: Wallet address of the keeper (string).
- `rewards_address`: Address to receive rewards (string).
- `consensus_address`: Address used for consensus (string).
- `registered_tx`: Transaction hash of the keeper's registration (string).
- `operator_id`: Operator index on the Attestation Center contract (int64).
- `voting_power`: Voting power of the keeper (big integer).
- `whitelisted`: Boolean indicating if the keeper is whitelisted.
- `registered`: Boolean indicating if the keeper is registered.
- `online`: Boolean indicating if the keeper is currently online.
- `version`: Version of the keeper's execution binary (string).
- `on_imua`: Boolean indicating if the keeper is on IMUA.
- `public_ip`: Public IP address of the keeper (string).
- `peer_id`: Peer ID of the keeper on the P2P network (string).
- `chat_id`: Telegram chat ID for notifications (int64).
- `email_id`: Email address for notifications (string).
- `rewards_booster`: Multiplier for rewards (big integer).
- `no_executed_tasks`: Number of tasks performed by the keeper (int64).
- `no_attested_tasks`: Number of tasks attested by the keeper (int64).
- `uptime`: Uptime of the keeper (int64, e.g., seconds online).
- `keeper_points`: Points accumulated by the keeper (big integer).
- `last_checked_in`: Timestamp of the last check-in.

### ApiKey Data

- `key`: API key string.
- `owner`: Owner of the API key (string).
- `is_active`: Boolean indicating if the API key is active.
- `rate_limit`: Rate limit for the API key (integer).
- `success_count`: Number of successful API calls (int64).
- `failed_count`: Number of failed API calls (int64).
- `last_used`: Timestamp of the last usage.
- `created_at`: Timestamp when the API key was created.