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
DROP TABLE IF EXISTS task_data;
DROP TABLE IF EXISTS keeper_data;

-- Create User_data table
CREATE TABLE IF NOT EXISTS user_data (
    user_id bigint,
    user_address text,
    created_at timestamp,
    job_ids set<bigint>,
    account_balance varint,
    token_balance varint,
    last_updated_at timestamp,
    user_points double,
    PRIMARY KEY (user_id)
);

-- Create Job_data table    
CREATE TABLE IF NOT EXISTS job_data (
    job_id bigint,
    job_title text,
    task_definition_id int,
    user_id bigint,
    link_job_id bigint,
    chain_status int,
    custom boolean,
    time_frame bigint,
    recurring boolean,
    status text,  -- 'pending', 'in-queue', 'running'
    job_cost_prediction double,
    task_ids set<bigint>,
    created_at timestamp,
    updated_at timestamp,
    last_executed_at timestamp,
    timezone text,
    PRIMARY KEY (job_id)
);

-- Create Time_Job_data table
CREATE TABLE IF NOT EXISTS time_job_data (
    job_id bigint,
    time_frame bigint,
    recurring boolean,
    time_interval bigint,
    target_chain_id text,
    target_contract_address text,
    target_function text,
    abi text,
    arg_type int,
    arguments list<text>,
    dynamic_arguments_script_ipfs_url text,
    created_at timestamp,
    updated_at timestamp,
    last_executed_at timestamp,
    timezone text,
    PRIMARY KEY (job_id)
);

-- Create Event_Job_data table
CREATE TABLE IF NOT EXISTS event_job_data (
    job_id bigint,
    time_frame bigint,
    recurring boolean,
    trigger_chain_id text,
    trigger_contract_address text,
    trigger_event text,
    target_chain_id text,
    target_contract_address text,
    target_function text,
    abi text,
    arg_type int,
    arguments list<text>,
    dynamic_arguments_script_ipfs_url text,
    created_at timestamp,
    updated_at timestamp,
    last_executed_at timestamp,
    timezone text,
    PRIMARY KEY (job_id)
);

-- Create Condition_Job_data table
CREATE TABLE IF NOT EXISTS condition_job_data (
    job_id bigint,
    time_frame bigint,
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
    dynamic_arguments_script_ipfs_url text,
    created_at timestamp,
    updated_at timestamp,
    last_executed_at timestamp,
    timezone text,
    PRIMARY KEY (job_id)
);

-- Create Task_data table
CREATE TABLE IF NOT EXISTS task_data (
    task_id bigint,
    task_number int,
    job_id bigint,
    task_definition_id int,
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
    PRIMARY KEY (task_id)
);

-- Create Keeper_data table
CREATE TABLE IF NOT EXISTS keeper_data (
    keeper_id bigint,
    keeper_name text,
    keeper_address text,
    consensus_address text,
    registered_tx text,
    operator_id text,
    rewards_address text,
    rewards_booster float,
    voting_power text,
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
    PRIMARY KEY (keeper_id)
);

-- Create ApiKey table
CREATE TABLE IF NOT EXISTS apikeys (
    key text,
    owner text,
    isActive boolean,
    rateLimit int,
    lastUsed timestamp,
    createdAt timestamp,
    PRIMARY KEY (key)
);

-- Create indexes for job_data table
CREATE INDEX ON triggerx.job_data (status);
CREATE INDEX ON triggerx.job_data (created_at);
CREATE INDEX ON triggerx.job_data (updated_at);
CREATE INDEX ON triggerx.job_data (last_executed_at);
CREATE INDEX ON triggerx.job_data (timezone);

-- Create indexes for time_job_data table
CREATE INDEX ON triggerx.time_job_data (created_at);
CREATE INDEX ON triggerx.time_job_data (updated_at);
CREATE INDEX ON triggerx.time_job_data (last_executed_at);
CREATE INDEX ON triggerx.time_job_data (timezone);

-- Create indexes for event_job_data table
CREATE INDEX ON triggerx.event_job_data (created_at);
CREATE INDEX ON triggerx.event_job_data (updated_at);
CREATE INDEX ON triggerx.event_job_data (last_executed_at);
CREATE INDEX ON triggerx.event_job_data (timezone);

-- Create indexes for condition_job_data table
CREATE INDEX ON triggerx.condition_job_data (created_at);
CREATE INDEX ON triggerx.condition_job_data (updated_at);
CREATE INDEX ON triggerx.condition_job_data (last_executed_at);
CREATE INDEX ON triggerx.condition_job_data (timezone);

