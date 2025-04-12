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
    PRIMARY KEY (user_id)
);

-- Create Job_data table    
CREATE TABLE IF NOT EXISTS job_data (
    job_id bigint,
    task_definition_id int,
    user_id bigint,
    priority int,
    security int,
    link_job_id bigint,
    chain_status int,
     
    time_frame bigint,
    recurring boolean,

    time_interval bigint,
    trigger_chain_id text,
    trigger_contract_address text,
    trigger_event text,
    script_ipfs_url text,
    script_trigger_function text,
    
    target_chain_id text,
    target_contract_address text,
    target_function text,
    arg_type int,
    arguments list<text>,
    script_target_function text,
    
    status boolean,
    job_cost_prediction double,
    created_at timestamp,
    last_executed_at timestamp,
    task_ids set<bigint>,
    PRIMARY KEY (job_id)
);

-- Create Task_data table
CREATE TABLE IF NOT EXISTS task_data (
    task_id bigint,
    task_number int,
    job_id bigint,
    task_definition_id int,
    created_at timestamp,
    task_fee bigint,
    
    execution_timestamp timestamp,
    execution_tx_hash text,
    task_performer_id bigint,
    
    proof_of_task text,
    action_data_cid text,
    task_attester_ids list<bigint>,
    
    is_approved boolean,
    tp_signature blob,
    ta_signature blob,
    task_submission_tx_hash text,
    
    is_successful boolean,
    PRIMARY KEY (task_id, task_number)
);

-- Create Keeper_data table
CREATE TABLE IF NOT EXISTS keeper_data (
    keeper_id bigint,
    keeper_name text,
    keeper_address text,
    registered_tx text,
    
    rewards_address text,
    rewards_booster int,
    voting_power bigint,
    keeper_points bigint,
    connection_address text,
    strategies list<text>,
    verified boolean,
    status boolean,
    online boolean,
    version text,
    no_exctask int,
    chat_id bigint,
    email_id text,
    PRIMARY KEY (keeper_id)
);

-- Create ApiKey table
CREATE TABLE IF NOT EXISTS apikeys (
    key text PRIMARY KEY,
    owner text,
    isActive boolean,
    rateLimit int,
    lastUsed timestamp,
    createdAt timestamp
);

