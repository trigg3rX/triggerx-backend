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
DROP TABLE IF EXISTS quorum_data;
DROP TABLE IF EXISTS keeper_data;
DROP TABLE IF EXISTS task_history;

-- Create User_data table
CREATE TABLE IF NOT EXISTS user_data (
    user_id bigint PRIMARY KEY,
    user_address text,
    job_ids set<bigint>,
    stake_amount varint,
    account_balance varint,
    created_at timestamp,
    last_updated_at timestamp
);

-- Create Job_data table
CREATE TABLE IF NOT EXISTS job_data (
    job_id bigint PRIMARY KEY,
    jobType int,
    user_id bigint,
    user_address text,
    chain_id int,
    time_frame bigint,
    time_interval int,
    contract_address text,
    target_function text,
    target_event text,
    arg_type int,
    arguments list<text>,
    status boolean,
    job_cost_prediction int,
    script_function text,
    script_ipfs_url text,
    created_at timestamp,
    last_executed_at timestamp,
    dispute_period_blocks varint,
    priority int,
    security int,
    task_ids set<bigint>
);

-- Create Task_data table
CREATE TABLE IF NOT EXISTS task_data (
    task_id bigint PRIMARY KEY,
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
    task_fee decimal,
    job_type text,
    blockExpiry varint,
    baseRewardFeeForAttesters varint,
    baseRewardFeeForPerformer varint,
    baseRewardFeeForAggregator varint,
    disputePeriodBlocks varint,
    minimumVotingPower varint,
    _restrictedOperatorIndexes list<varint>,
    proofOfTask text,
    data blob,
    taskPerformer text,
    _isApproved boolean,
    _tpSignature blob,
    _taSignature list<varint>,
    _operatorIds list<varint>
);

-- Create Quorum_data table
CREATE TABLE IF NOT EXISTS quorum_data (
    quorum_id bigint PRIMARY KEY,
    quorum_no int,
    quorum_creation_block bigint,
    quorum_termination_block bigint,
    quorum_tx_hash text,
    keepers list<text>,
    quorum_stake_total bigint,
    task_ids set<bigint>,
    quorum_status boolean
);

-- Create Keeper_data table
CREATE TABLE IF NOT EXISTS keeper_data (
    keeper_id bigint PRIMARY KEY,
    withdrawal_address text,
    stakes list<decimal>,
    strategies list<text>,
    verified boolean,
    current_quorum_no int,
    registered_tx text,
    status boolean,
    bls_signing_keys list<text>,
    connection_address text
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
