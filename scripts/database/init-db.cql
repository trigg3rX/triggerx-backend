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
    user_id bigint PRIMARY KEY,
    user_address text,
    created_at timestamp,
    job_ids set<bigint>,
    account_balance varint,
    token_balance varint,
    last_updated_at timestamp
);

-- Create Job_data table    
CREATE TABLE IF NOT EXISTS job_data (
    job_id bigint PRIMARY KEY,
    task_definition_id int,
    user_id bigint,
    priority int,
    security int,
    link_job_id bigint,
    chain_status int,
     
    time_frame bigint,
    recurring boolean,

    time_interval bigint,
    trigger_chain_id int,
    trigger_contract_address text,
    trigger_event text,
    script_ipfs_url text,
    script_trigger_function text,
    
    target_chain_id int,
    target_contract_address text,
    target_function text,
    arg_type int,
    arguments list<text>,
    script_target_function text,
    
    status boolean,
    job_cost_prediction int,
    created_at timestamp,
    last_executed_at timestamp,
    task_ids set<bigint>
);

-- Create Task_data table
CREATE TABLE IF NOT EXISTS task_data (
    task_id bigint PRIMARY KEY,
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
    
    is_successful boolean
);

-- Create Keeper_data table
CREATE TABLE IF NOT EXISTS keeper_data (
    keeper_id bigint PRIMARY KEY,
    keeper_address text,
    keeper_type int,
    registered_tx text,
    
    rewards_address text,
    stakes list<decimal>,
    strategies list<text>,
    verified boolean,
    status boolean,
    consensus_keys list<text>,
    connection_address text
);
