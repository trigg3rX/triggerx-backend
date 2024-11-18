-- Create keyspace
CREATE KEYSPACE IF NOT EXISTS triggerx
WITH replication = {
    'class': 'SimpleStrategy',
    'replication_factor': 1
};

-- Switch to keyspace
USE triggerx;

-- Drop existing tables
DROP TABLE IF EXISTS job_data;

-- Create User_data table
CREATE TABLE IF NOT EXISTS user_data (
    user_id bigint PRIMARY KEY,
    user_address text ,
    job_ids set<bigint>
);

-- Create Job_data table
CREATE TABLE IF NOT EXISTS job_data (
    job_id bigint PRIMARY KEY,
    jobType int,
    user_id bigint,
    chain_id int,
    time_frame bigint,
    time_interval int,
    contract_address text,
    target_function text,
    arg_type int,
    arguments list<text>,
    status boolean,
    job_cost_prediction double
);

-- Create Task_data table
CREATE TABLE IF NOT EXISTS task_data (
    task_id bigint,
    job_id bigint,
    task_no int,
    quorum_id bigint,
    quorum_number int,
    quorum_threshold decimal,
    task_created_block bigint,
    task_created_tx_hash text,
    task_responded_block bigint,
    task_responded_tx_hash text,
    task_hash text,
    task_response_hash text,
    quorum_keeper_hash text,
    PRIMARY KEY (task_id)
);

-- Create Quorum_data table
CREATE TABLE IF NOT EXISTS quorum_data (
    quorum_id bigint PRIMARY KEY,
    quorum_no int,
    quorum_creation_block bigint,
    quorum_tx_hash text,
    keepers list<text>,
    quorum_stake_total bigint,
    quorum_threshold decimal,
    task_ids set<bigint>
);

-- Create Keeper_data table
CREATE TABLE IF NOT EXISTS keeper_data (
    keeper_id bigint PRIMARY KEY,
    withdrawal_address text CHECK (withdrawal_address MATCHES '^0x[0-9a-fA-F]{40}$'),
    stakes bigint,
    strategies int,
    verified boolean,
    status boolean,
    current_quorum_no int,
    registered_block_no bigint,
    register_tx_hash text,
    connection_address text,
    keystore_data text
);

-- Create Task_history table
CREATE TABLE IF NOT EXISTS task_history (
    task_id bigint PRIMARY KEY,
    quorum_id bigint,
    keepers list<text>,
    responses list<text>,
    consensus_method text,
    validation_status boolean,
    tx_hash text
);