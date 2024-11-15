-- Create keyspace
CREATE KEYSPACE IF NOT EXISTS triggerx
WITH replication = {
    'class': 'SimpleStrategy',
    'replication_factor': 1
};

-- Switch to keyspace
USE triggerx;

-- Drop existing tables
DROP TABLE IF EXISTS tasks;

-- Create User_data table
CREATE TABLE IF NOT EXISTS user_data (
    user_id bigint PRIMARY KEY,
    user_address text ,
    job_ids set<bigint>
);

-- Create Job_data table
CREATE TABLE IF NOT EXISTS job_data (
    job_id bigint PRIMARY KEY,
    user_id bigint,
    chain_id int,
    stake bigint,
    time_frame timestamp,
    time_interval int,
    contract_address text,
    target_function text,
    arg_type int,
    arguments list<text>,
    status boolean,
    job_cost_prediction decimal
);

-- Create Task_data table
CREATE TABLE IF NOT EXISTS task_data (
    task_ref_no text PRIMARY KEY,
    job_id bigint,
    task_id bigint,
    quorum_id int,
    quorum_no int,
    quorum_threshold decimal,
    task_hash text,
    task_response_hash text
);

-- Create Keeper_data table
CREATE TABLE IF NOT EXISTS keeper_data (
    keeper_id bigint PRIMARY KEY,
    withdrawal_address text,
    verified boolean,
    stake_amount bigint,
    strategies int,
    quorum_id int,
    status boolean
);