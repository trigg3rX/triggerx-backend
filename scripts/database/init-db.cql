-- Create keyspace
CREATE KEYSPACE IF NOT EXISTS triggerx
WITH replication = {
    'class': 'SimpleStrategy',
    'replication_factor': 1
};

-- Switch to keyspace
USE triggerx;

-- Drop existing tables
DROP TABLE IF EXISTS user_data;
DROP TABLE IF EXISTS job_data;
DROP TABLE IF EXISTS time_job_data;
DROP TABLE IF EXISTS event_job_data;
DROP TABLE IF EXISTS condition_job_data;
DROP TABLE IF EXISTS task_data;
DROP TABLE IF EXISTS keeper_data;
DROP TABLE IF EXISTS apikeys;

-- Create User_data table (without counters)
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

-- Create Job_data table    
CREATE TABLE IF NOT EXISTS job_data (
    job_id text,
    job_title text,
    task_definition_id int,
    created_chain_id text,
    user_address text,
    link_job_id text,
    chain_status int,
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

-- Create Time_Job_data table
CREATE TABLE IF NOT EXISTS time_job_data (
    job_id text,
    task_definition_id int,
    schedule_type text,
    time_interval bigint,
    cron_expression text,
    specific_schedule text,
    next_execution_timestamp timestamp,
    target_chain_id text,
    target_contract_address text,
    target_function text,
    abi text,
    arg_type int,
    arguments list<text>,
    dynamic_arguments_script_url text,
    is_completed boolean,
    last_executed_at timestamp,
    expiration_time timestamp,
    PRIMARY KEY (job_id)
);

-- Create Event_Job_data table
CREATE TABLE IF NOT EXISTS event_job_data (
    job_id text,
    task_definition_id int,
    recurring boolean,
    trigger_chain_id text,
    trigger_contract_address text,
    trigger_event text,
    trigger_event_filter_para_name text,
    trigger_event_filter_value text,
    target_chain_id text,
    target_contract_address text,
    target_function text,
    abi text,
    arg_type int,
    arguments list<text>,
    dynamic_arguments_script_url text,
    is_completed boolean,
    last_executed_at timestamp,
    expiration_time timestamp,
    PRIMARY KEY (job_id)
);

-- Create Condition_Job_data table
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
    is_completed boolean,
    last_executed_at timestamp,
    expiration_time timestamp,
    PRIMARY KEY (job_id)
);

-- Create Task_data table
CREATE TABLE IF NOT EXISTS task_data (
    task_id bigint,
    task_number bigint,
    job_id text,
    task_definition_id int,
    created_at timestamp,
    task_opx_predicted_cost text,
    task_opx_actual_cost text,
    execution_timestamp timestamp,
    execution_tx_hash text,
    task_performer_id bigint,
    task_attester_ids list<bigint>,
    proof_of_task text,
    submission_tx_hash text,
    is_successful boolean,
    is_accepted boolean,
    is_imua boolean,
    PRIMARY KEY (task_id)
);

-- Create Keeper_data table
CREATE TABLE IF NOT EXISTS keeper_data (
    keeper_name text,
    keeper_address text,
    rewards_address text,
    consensus_address text,
    registered_tx text,
    operator_id bigint,
    voting_power text,
    whitelisted boolean,
    registered boolean,
    online boolean,
    version text,
    on_imua boolean,
    public_ip text,
    chat_id bigint,
    email_id text,
    rewards_booster text,
    no_executed_tasks bigint,
    no_attested_tasks bigint,
    uptime bigint,
    keeper_points text,
    last_checked_in timestamp,
    PRIMARY KEY (keeper_address)
);

-- Create ApiKey table
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

-- Drop existing indexes if they exist
DROP INDEX IF EXISTS triggerx.job_data_status_idx;
DROP INDEX IF EXISTS triggerx.job_data_created_at_idx;
DROP INDEX IF EXISTS triggerx.job_data_updated_at_idx;
DROP INDEX IF EXISTS triggerx.job_data_last_executed_at_idx;
DROP INDEX IF EXISTS triggerx.job_data_timezone_idx;

DROP INDEX IF EXISTS triggerx.time_job_data_last_executed_at_idx;
DROP INDEX IF EXISTS triggerx.time_job_data_next_execution_timestamp_idx;
DROP INDEX IF EXISTS triggerx.time_job_data_timezone_idx;

DROP INDEX IF EXISTS triggerx.event_job_data_last_executed_at_idx;
DROP INDEX IF EXISTS triggerx.event_job_data_timezone_idx;

DROP INDEX IF EXISTS triggerx.condition_job_data_last_executed_at_idx;
DROP INDEX IF EXISTS triggerx.condition_job_data_timezone_idx;

-- Create indexes for job_data table
CREATE INDEX IF NOT EXISTS job_data_status_idx ON triggerx.job_data (status);
CREATE INDEX IF NOT EXISTS job_data_created_at_idx ON triggerx.job_data (created_at);
CREATE INDEX IF NOT EXISTS job_data_updated_at_idx ON triggerx.job_data (updated_at);
CREATE INDEX IF NOT EXISTS job_data_last_executed_at_idx ON triggerx.job_data (last_executed_at);
CREATE INDEX IF NOT EXISTS job_data_timezone_idx ON triggerx.job_data (timezone);

-- Create indexes for time_job_data table
CREATE INDEX IF NOT EXISTS time_job_data_last_executed_at_idx ON triggerx.time_job_data (last_executed_at);
CREATE INDEX IF NOT EXISTS time_job_data_next_execution_timestamp_idx ON triggerx.time_job_data (next_execution_timestamp);

-- Create indexes for event_job_data table
CREATE INDEX IF NOT EXISTS event_job_data_last_executed_at_idx ON triggerx.event_job_data (last_executed_at);

-- Create indexes for condition_job_data table
CREATE INDEX IF NOT EXISTS condition_job_data_last_executed_at_idx ON triggerx.condition_job_data (last_executed_at);
