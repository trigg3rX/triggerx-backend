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
DROP TABLE IF EXISTS safe_addresses;

-- Create User_data table (without counters)
CREATE TABLE IF NOT EXISTS user_data (
    user_id bigint,
    user_address text,
    job_ids set<varint>,
    ether_balance varint,
    token_balance varint,
    user_points double,
    total_jobs bigint,
    total_tasks bigint,
    created_at timestamp,
    last_updated_at timestamp,
    email_id text,
    PRIMARY KEY (user_id)
);

-- Create Job_data table    
CREATE TABLE IF NOT EXISTS job_data (
    job_id varint,
    job_title text,
    task_definition_id int,
    user_id bigint,
    link_job_id varint,
    chain_status int,
    custom boolean,
    time_frame bigint,
    recurring boolean,
    status text,  -- 'pending', 'in-queue', 'running'
    job_cost_prediction double,
    job_cost_actual double,
    task_ids set<bigint>,
    created_at timestamp,
    updated_at timestamp,
    last_executed_at timestamp,
    timezone text,
    is_imua boolean,
    created_chain_id text,
    safe_address text,
    PRIMARY KEY (job_id)
);

-- Create Time_Job_data table
CREATE TABLE IF NOT EXISTS time_job_data (
    job_id varint,
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
    is_completed boolean,
    is_active boolean,
    created_at timestamp,
    updated_at timestamp,
    last_executed_at timestamp,
    expiration_time timestamp,
    PRIMARY KEY (job_id)
);

-- Create Event_Job_data table
CREATE TABLE IF NOT EXISTS event_job_data (
    job_id varint,
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
    is_completed boolean,
    is_active boolean,
    created_at timestamp,
    updated_at timestamp,
    last_executed_at timestamp,
    expiration_time timestamp,
    timezone text,
    PRIMARY KEY (job_id)
);

-- Create Condition_Job_data table
CREATE TABLE IF NOT EXISTS condition_job_data (
    job_id varint,
    task_definition_id int,
    recurring boolean,
    condition_type text,
    upper_limit double,
    lower_limit double,
    value_source_type text,
    value_source_url text,
    target_chain_id text,
    target_contract_address text,
    target_function text,
    abi text,
    arg_type int,
    arguments list<text>,
    dynamic_arguments_script_url text,
    is_completed boolean,
    is_active boolean,
    created_at timestamp,
    updated_at timestamp,
    last_executed_at timestamp,
    expiration_time timestamp,
    timezone text,
    selected_key_route text, 
    PRIMARY KEY (job_id)
);

-- Create Task_data table
CREATE TABLE IF NOT EXISTS task_data (
    task_id bigint,
    task_number bigint,
    task_status text,
    task_error text,
    job_id varint,
    task_definition_id int,
    created_at timestamp,
    task_opx_predicted_cost double,
    task_opx_cost double,
    execution_timestamp timestamp,
    execution_tx_hash text,
    task_performer_id bigint,
    task_attester_ids list<bigint>,
    proof_of_task text,
    tp_signature blob,
    ta_signature blob,
    task_submission_tx_hash text,
    is_successful boolean,
    is_accepted boolean,
    is_imua boolean,
    PRIMARY KEY (task_id)
);

-- Create Keeper_data table
CREATE TABLE IF NOT EXISTS keeper_data (
    keeper_id bigint,
    keeper_address text,
    keeper_name text,
    consensus_address text,
    registered_tx text,
    operator_id bigint,
    rewards_address text,
    rewards_booster double,
    voting_power bigint,
    keeper_points double,
    connection_address text,
    peer_id text,
    strategies list<text>,
    whitelisted boolean,
    registered boolean,
    online boolean,
    version text,
    no_executed_tasks bigint,
    no_attested_tasks bigint,
    chat_id bigint,
    email_id text,
    last_checked_in timestamp,
    on_imua boolean,
    uptime bigint,
    PRIMARY KEY (keeper_id)
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

-- Create Safe_addresses table
CREATE TABLE IF NOT EXISTS safe_addresses (
    user_address text,
    safe_address text,
    safe_name text,
    created_at timestamp,
    PRIMARY KEY (user_address, safe_address)
);

-- Drop existing indexes if they exist
DROP INDEX IF EXISTS triggerx.job_data_status_idx;
DROP INDEX IF EXISTS triggerx.job_data_created_at_idx;
DROP INDEX IF EXISTS triggerx.job_data_updated_at_idx;
DROP INDEX IF EXISTS triggerx.job_data_last_executed_at_idx;
DROP INDEX IF EXISTS triggerx.job_data_timezone_idx;

DROP INDEX IF EXISTS triggerx.time_job_data_created_at_idx;
DROP INDEX IF EXISTS triggerx.time_job_data_updated_at_idx;
DROP INDEX IF EXISTS triggerx.time_job_data_last_executed_at_idx;
DROP INDEX IF EXISTS triggerx.time_job_data_next_execution_timestamp_idx;
DROP INDEX IF EXISTS triggerx.time_job_data_timezone_idx;

DROP INDEX IF EXISTS triggerx.event_job_data_created_at_idx;
DROP INDEX IF EXISTS triggerx.event_job_data_updated_at_idx;
DROP INDEX IF EXISTS triggerx.event_job_data_last_executed_at_idx;
DROP INDEX IF EXISTS triggerx.event_job_data_timezone_idx;

DROP INDEX IF EXISTS triggerx.condition_job_data_created_at_idx;
DROP INDEX IF EXISTS triggerx.condition_job_data_updated_at_idx;
DROP INDEX IF EXISTS triggerx.condition_job_data_last_executed_at_idx;
DROP INDEX IF EXISTS triggerx.condition_job_data_timezone_idx;

-- Create indexes for job_data table
CREATE INDEX IF NOT EXISTS job_data_status_idx ON triggerx.job_data (status);
CREATE INDEX IF NOT EXISTS job_data_created_at_idx ON triggerx.job_data (created_at);
CREATE INDEX IF NOT EXISTS job_data_updated_at_idx ON triggerx.job_data (updated_at);
CREATE INDEX IF NOT EXISTS job_data_last_executed_at_idx ON triggerx.job_data (last_executed_at);
CREATE INDEX IF NOT EXISTS job_data_timezone_idx ON triggerx.job_data (timezone);

-- Create indexes for time_job_data table
CREATE INDEX IF NOT EXISTS time_job_data_created_at_idx ON triggerx.time_job_data (created_at);
CREATE INDEX IF NOT EXISTS time_job_data_updated_at_idx ON triggerx.time_job_data (updated_at);
CREATE INDEX IF NOT EXISTS time_job_data_last_executed_at_idx ON triggerx.time_job_data (last_executed_at);
CREATE INDEX IF NOT EXISTS time_job_data_timezone_idx ON triggerx.time_job_data (timezone);
CREATE INDEX IF NOT EXISTS time_job_data_next_execution_timestamp_idx ON triggerx.time_job_data (next_execution_timestamp);

-- Create indexes for event_job_data table
CREATE INDEX IF NOT EXISTS event_job_data_created_at_idx ON triggerx.event_job_data (created_at);
CREATE INDEX IF NOT EXISTS event_job_data_updated_at_idx ON triggerx.event_job_data (updated_at);
CREATE INDEX IF NOT EXISTS event_job_data_last_executed_at_idx ON triggerx.event_job_data (last_executed_at);
CREATE INDEX IF NOT EXISTS event_job_data_timezone_idx ON triggerx.event_job_data (timezone);

-- Create indexes for condition_job_data table
CREATE INDEX IF NOT EXISTS condition_job_data_created_at_idx ON triggerx.condition_job_data (created_at);
CREATE INDEX IF NOT EXISTS condition_job_data_updated_at_idx ON triggerx.condition_job_data (updated_at);
CREATE INDEX IF NOT EXISTS condition_job_data_last_executed_at_idx ON triggerx.condition_job_data (last_executed_at);
CREATE INDEX IF NOT EXISTS condition_job_data_timezone_idx ON triggerx.condition_job_data (timezone);

-- Custom Script Jobs Schema
-- TaskDefinitionID = 7

-- Custom jobs table (extends existing job pattern)
CREATE TABLE IF NOT EXISTS triggerx.custom_jobs (
    job_id varint PRIMARY KEY,
    task_definition_id int,
    recurring boolean,
    custom_script_url text,
    time_interval bigint,              -- Execution interval in seconds
    is_completed boolean,
    is_active boolean,
    target_chain_id text,
    created_at timestamp,
    updated_at timestamp,
    last_executed_at timestamp,
    expiration_time timestamp,
    -- Script metadata
    script_language text,              -- 'ts', 'go', 'python', 'javascript'
    script_hash text,                  -- keccak256(scriptCode) for verification
    next_execution_time timestamp,     -- Next scheduled execution
    max_execution_time int,            -- Script timeout in seconds (default 60)
    challenge_period bigint            -- Challenge period in seconds (default 21600 = 6 hours)
);

-- Indexes for scheduler queries
CREATE INDEX IF NOT EXISTS custom_jobs_active_idx
ON triggerx.custom_jobs (is_active);

-- Script executions tracking table
CREATE TABLE IF NOT EXISTS triggerx.custom_script_executions (
    execution_id text PRIMARY KEY,     -- Unique execution identifier (UUID)
    job_id varint,
    task_id bigint,                    -- Link to task_data table
    scheduled_time timestamp,          -- When it was supposed to execute
    actual_time timestamp,             -- When it actually executed
    performer_address text,

    -- Input data (for deterministic re-execution)
    input_timestamp bigint,
    input_storage text,                -- JSON snapshot of storage at execution time
    input_hash text,                   -- Hash of inputs for verification

    -- Output data
    should_execute boolean,
    target_contract text,              -- Contract address to call
    calldata text,                     -- Encoded function call
    output_hash text,                  -- Hash of outputs for verification

    -- Metadata (CRITICAL: API calls, contract calls, block numbers)
    execution_metadata text,           -- JSON containing:
                                       -- - API calls made: [{url, blockNumber, response}, ...]
                                       -- - Contract calls: [{contract, blockNumber, function, response}, ...]
                                       -- - Oracle calls: [{oracle, blockNumber, data}, ...]

    -- Proof
    script_hash text,
    signature text,                    -- Performer's signature

    -- Execution result
    tx_hash text,                      -- Transaction hash if executed
    execution_status text,             -- 'success', 'failed', 'no_execution'
    execution_error text,              -- Error message if failed

    -- Verification status
    verification_status text,          -- 'pending', 'verified', 'challenged', 'slashed'
    challenge_deadline timestamp,      -- Deadline for challenges
    is_challenged boolean,
    challenge_count int,

    created_at timestamp
);

-- Indexes for execution queries
CREATE INDEX IF NOT EXISTS executions_job_id_idx
ON triggerx.custom_script_executions (job_id);

CREATE INDEX IF NOT EXISTS executions_verification_status_idx
ON triggerx.custom_script_executions (verification_status);

CREATE INDEX IF NOT EXISTS executions_performer_idx
ON triggerx.custom_script_executions (performer_address);

CREATE INDEX IF NOT EXISTS executions_task_id_idx
ON triggerx.custom_script_executions (task_id);

-- Script storage table (persistent key-value storage per job)
CREATE TABLE IF NOT EXISTS triggerx.script_storage (
    job_id varint,
    storage_key text,
    storage_value text,
    updated_at timestamp,
    PRIMARY KEY (job_id, storage_key)
);

-- Execution challenges table
CREATE TABLE IF NOT EXISTS triggerx.execution_challenges (
    challenge_id text PRIMARY KEY,
    execution_id text,
    challenger_address text,
    challenge_reason text,             -- 'wrong_output', 'missing_execution', 'invalid_calldata'

    -- Challenger's claimed output
    challenger_output_hash text,
    challenger_should_execute boolean,
    challenger_target_contract text,
    challenger_calldata text,
    challenger_signature text,

    -- Resolution
    resolution_status text,            -- 'pending', 'approved', 'rejected'
    resolution_time timestamp,
    validator_count int,
    approve_count int,
    reject_count int,

    created_at timestamp
);

CREATE INDEX IF NOT EXISTS challenges_execution_idx
ON triggerx.execution_challenges (execution_id);

CREATE INDEX IF NOT EXISTS challenges_status_idx
ON triggerx.execution_challenges (resolution_status);

