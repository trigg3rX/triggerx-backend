CREATE KEYSPACE IF NOT EXISTS triggerx
WITH replication = {
    'class': 'SimpleStrategy',
    'replication_factor': 1
};

USE triggerx;

DROP TABLE IF EXISTS apikeys;
DROP TABLE IF EXISTS user_data;
DROP TABLE IF EXISTS safe_addresses;
DROP TABLE IF EXISTS job_data;
DROP TABLE IF EXISTS time_job_data;
DROP TABLE IF EXISTS event_job_data;
DROP TABLE IF EXISTS condition_job_data;
DROP TABLE IF EXISTS agent_script_executions;
DROP TABLE IF EXISTS script_storage;
DROP TABLE IF EXISTS execution_challenges;
DROP TABLE IF EXISTS task_data;
DROP TABLE IF EXISTS keeper_data;

CREATE TABLE IF NOT EXISTS apikeys (
    key text,
    owner text,
    is_active boolean,
    rate_limit int,
    success_count bigint,
    failed_count bigint,
    last_used timestamp,
    created_at timestamp,
    PRIMARY KEY (key)
);

CREATE TABLE IF NOT EXISTS user_data (
    user_address text,
    email_id text,
    job_ids set<text>,
    user_points text,
    total_jobs bigint,
    total_tasks bigint,
    created_at timestamp,
    last_updated_at timestamp,
    PRIMARY KEY (user_address)
);

CREATE TABLE IF NOT EXISTS safe_addresses (
    user_address text,
    safe_address text,
    safe_name text,
    created_at timestamp,
    PRIMARY KEY (user_address, safe_address)
);

CREATE TABLE IF NOT EXISTS job_data (
    job_id text,
    job_title text,
    task_definition_id int,
    created_chain_id text,
    user_address text,
    link_job_id text,
    chain_status int,
    safe_address text,
    timezone text,
    is_imua boolean,
    job_type text,
    time_frame bigint,
    recurring boolean,
    status text,
    job_cost_prediction text,
    job_cost_actual text,
    task_ids set<bigint>,
    created_at timestamp,
    updated_at timestamp,
    last_executed_at timestamp,
    PRIMARY KEY (job_id)
);

CREATE TABLE IF NOT EXISTS time_job_data (
    job_id text,
    task_definition_id int,
    schedule_type text,
    time_interval bigint,
    cron_expression text,
    specific_schedule text,
    timezone text,
    next_execution_timestamp timestamp,
    target_chain_id text,
    target_contract_address text,
    target_function text,
    abi text,
    arg_type int,
    arguments list<text>,
    dynamic_arguments_script_url text,
    agent_script_url text,
    agent_script_language text,
    agent_script_hash text,
    agent_target_chain_id int,
    max_execution_time int,
    challenge_period bigint,
    is_active boolean,
    last_executed_at timestamp,
    expiration_time timestamp,
    PRIMARY KEY (job_id)
);

CREATE TABLE IF NOT EXISTS event_job_data (
    job_id text,
    task_definition_id int,
    recurring boolean,
    trigger_chain_id text,
    trigger_contract_address text,
    trigger_event text,
    event_filter_para_name text,
    event_filter_value text,
    target_chain_id text,
    target_contract_address text,
    target_function text,
    abi text,
    arg_type int,
    arguments list<text>,
    dynamic_arguments_script_url text,
    agent_script_url text,
    agent_script_language text,
    agent_script_hash text,
    agent_target_chain_id int,
    max_execution_time int,
    challenge_period bigint,
    is_active boolean,
    last_executed_at timestamp,
    expiration_time timestamp,
    PRIMARY KEY (job_id)
);

CREATE TABLE IF NOT EXISTS condition_job_data (
    job_id text,
    task_definition_id int,
    recurring boolean,
    condition_type text,
    upper_limit double,
    lower_limit double,
    value_source_type text,
    value_source_url text,
    selected_key_route text,
    target_chain_id text,
    target_contract_address text,
    target_function text,
    abi text,
    arg_type int,
    arguments list<text>,
    dynamic_arguments_script_url text,
    agent_script_url text,
    agent_script_language text,
    agent_script_hash text,
    agent_target_chain_id int,
    max_execution_time int,
    challenge_period bigint,
    is_active boolean,
    last_executed_at timestamp,
    expiration_time timestamp,
    PRIMARY KEY (job_id)
);

CREATE TABLE IF NOT EXISTS triggerx.agent_script_executions (
    execution_id text PRIMARY KEY,
    job_id text,
    task_id bigint,
    task_definition_id int,
    scheduled_time timestamp,
    actual_time timestamp,
    performer_address text,
    input_timestamp bigint,
    input_storage text,
    input_hash text,
    trigger_data text,
    should_execute boolean,
    target_contract text,
    calldata text,
    output_hash text,
    execution_metadata text,
    script_hash text,
    signature text,
    tx_hash text,
    execution_status text,
    execution_error text,
    verification_status text,
    challenge_deadline timestamp,
    is_challenged boolean,
    challenge_count int,
    created_at timestamp
);

CREATE TABLE IF NOT EXISTS triggerx.script_storage (
    job_id text,
    storage_key text,
    storage_value text,
    updated_at timestamp,
    PRIMARY KEY (job_id, storage_key)
);

CREATE TABLE IF NOT EXISTS triggerx.execution_challenges (
    challenge_id text PRIMARY KEY,
    execution_id text,
    challenger_address text,
    challenge_reason text,
    challenger_output_hash text,
    challenger_should_execute boolean,
    challenger_target_contract text,
    challenger_calldata text,
    challenger_signature text,
    bond_amount text,
    resolution_status text,
    resolution_time timestamp,
    validator_count int,
    approve_count int,
    reject_count int,
    created_at timestamp
);

CREATE TABLE IF NOT EXISTS task_data (
    task_id bigint,
    task_number bigint,
    task_status text,
    task_error text,
    job_id text,
    task_definition_id int,
    created_at timestamp,
    executed_at timestamp,
    submitted_at timestamp,
    task_opx_predicted_cost text,
    task_opx_actual_cost text,
    execution_tx_hash text,
    submission_tx_hash text,
    converted_arguments list<text>,
    task_performer_address list<text>,
    task_attester_address list<text>,
    proof_of_task text,
    is_successful boolean,
    is_accepted boolean,
    is_imua boolean,
    PRIMARY KEY (task_id)
);

CREATE TABLE IF NOT EXISTS keeper_data (
    keeper_address text,
    keeper_name text,
    rewards_address text,
    consensus_address text,
    operator_id int,
    voting_power text,
    registered boolean,
    registered_at list<timestamp>,
    online boolean,
    version text,
    network text,
    connection_address text,
    peer_id text,
    uptime bigint,
    last_checked_in timestamp,
    rewards_booster double,
    no_executed_tasks bigint,
    no_attested_tasks bigint,
    keeper_points text,
    chat_id bigint,
    email_id text,
    PRIMARY KEY (keeper_address)
);

CREATE INDEX IF NOT EXISTS job_data_status_idx ON triggerx.job_data (status);
CREATE INDEX IF NOT EXISTS job_data_created_at_idx ON triggerx.job_data (created_at);
CREATE INDEX IF NOT EXISTS job_data_updated_at_idx ON triggerx.job_data (updated_at);
CREATE INDEX IF NOT EXISTS job_data_last_executed_at_idx ON triggerx.job_data (last_executed_at);
CREATE INDEX IF NOT EXISTS job_data_safe_address_idx ON triggerx.job_data (safe_address);

CREATE INDEX IF NOT EXISTS time_job_data_last_executed_at_idx ON triggerx.time_job_data (last_executed_at);
CREATE INDEX IF NOT EXISTS time_job_data_next_execution_timestamp_idx ON triggerx.time_job_data (next_execution_timestamp);

CREATE INDEX IF NOT EXISTS event_job_data_last_executed_at_idx ON triggerx.event_job_data (last_executed_at);

CREATE INDEX IF NOT EXISTS condition_job_data_last_executed_at_idx ON triggerx.condition_job_data (last_executed_at);

CREATE INDEX IF NOT EXISTS task_data_job_id_idx ON triggerx.task_data (job_id);

CREATE INDEX IF NOT EXISTS agent_exec_job_id_idx ON triggerx.agent_script_executions (job_id);
CREATE INDEX IF NOT EXISTS agent_exec_task_id_idx ON triggerx.agent_script_executions (task_id);
CREATE INDEX IF NOT EXISTS agent_exec_verification_status_idx ON triggerx.agent_script_executions (verification_status);
CREATE INDEX IF NOT EXISTS agent_exec_performer_idx ON triggerx.agent_script_executions (performer_address);
CREATE INDEX IF NOT EXISTS agent_exec_tdi_idx ON triggerx.agent_script_executions (task_definition_id);

CREATE INDEX IF NOT EXISTS challenges_execution_idx ON triggerx.execution_challenges (execution_id);
CREATE INDEX IF NOT EXISTS challenges_status_idx ON triggerx.execution_challenges (resolution_status);
CREATE INDEX IF NOT EXISTS challenges_challenger_idx ON triggerx.execution_challenges (challenger_address);
