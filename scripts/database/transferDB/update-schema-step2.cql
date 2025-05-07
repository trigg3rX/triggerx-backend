-- Switch to keyspace
USE triggerx;

-- Drop old tables first
DROP TABLE IF EXISTS user_data;
DROP TABLE IF EXISTS job_data;
DROP TABLE IF EXISTS task_data;
DROP TABLE IF EXISTS keeper_data;

-- Create new tables with original names
CREATE TABLE IF NOT EXISTS user_data (
    partition_key text,
    user_id bigint,
    user_address text,
    created_at timestamp,
    job_ids set<bigint>,
    account_balance varint,
    token_balance varint,
    last_updated_at timestamp,
    user_points double,
    PRIMARY KEY ((partition_key), user_id, user_points)
) WITH CLUSTERING ORDER BY (user_id ASC, user_points DESC);

CREATE TABLE IF NOT EXISTS job_data (
    partition_key text,
    job_id bigint,
    task_definition_id int,
    user_id bigint,
    priority int,
    security int,
    link_job_id bigint,
    chain_status int,
    custom boolean,
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
    abi text,
    arg_type int,
    arguments list<text>,
    script_target_function text,
    status boolean,
    job_cost_prediction double,
    created_at timestamp,
    last_executed_at timestamp,
    task_ids set<bigint>,
    PRIMARY KEY ((partition_key), created_at, job_id)
) WITH CLUSTERING ORDER BY (created_at DESC, job_id ASC);

CREATE TABLE IF NOT EXISTS task_data (
    partition_key text,
    task_id bigint,
    task_number int,
    job_id bigint,
    task_definition_id int,
    created_at timestamp,
    task_fee double,
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
    PRIMARY KEY ((partition_key), created_at, task_id)
) WITH CLUSTERING ORDER BY (created_at DESC, task_id ASC);

CREATE TABLE IF NOT EXISTS keeper_data (
    partition_key text,
    keeper_id bigint,
    keeper_name text,
    keeper_address text,
    registered_tx text,
    operator_id text,
    rewards_address text,
    rewards_booster float,
    voting_power text,
    keeper_points double,
    connection_address text,
    peer_id text,
    strategies list<text>,
    verified boolean,
    status boolean,
    online boolean,
    version text,
    no_exctask int,
    chat_id bigint,
    email_id text,
    PRIMARY KEY ((partition_key), keeper_points, keeper_id)
) WITH CLUSTERING ORDER BY (keeper_points DESC, keeper_id ASC);