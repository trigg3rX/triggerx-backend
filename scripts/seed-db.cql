USE triggerx;

-- Insert sample user data
INSERT INTO user_data (user_id, user_address, job_ids, stake_amount, created_at, last_updated_at)
VALUES (
    1, 
    '0x1234567890123456789012345678901234567890',
    {1, 2},
    100000,
    toTimestamp(now()),
    toTimestamp(now())
);

-- Insert sample job data
INSERT INTO job_data (
    job_id, jobType, user_id, user_address, chain_id, 
    time_frame, time_interval, contract_address, target_function,
    arg_type, arguments, status, job_cost_prediction,
    script_function, script_ipfs_url, time_check, created_at, last_executed_at
)
VALUES (
    1,
    1,
    1,
    '0x1234567890123456789012345678901234567890',
    1,
    3600,
    300,
    '0x2234567890123456789012345678901234567890',
    'checkPrice',
    1,
    ['1000', '2000'],
    true,
    100,
    'priceAlert',
    'ipfs://QmSampleHash',
    toTimestamp(now()),
    toTimestamp(now()),
    toTimestamp(now())
);

-- Insert sample keeper data
INSERT INTO keeper_data (
    keeper_id, withdrawal_address, stakes, strategies,
    verified, current_quorum_no, registered_tx, status,
    bls_signing_keys, connection_address
)
VALUES (
    1,
    '0x3234567890123456789012345678901234567890',
    [1000.5, 2000.75],
    ['strategy1', 'strategy2'],
    true,
    1,
    '0xtxhash123',
    true,
    ['key1', 'key2'],
    '127.0.0.1:8080'
);

-- Insert sample quorum data
INSERT INTO quorum_data (
    quorum_id, quorum_no, quorum_creation_block,
    quorum_tx_hash, keepers, quorum_stake_total,
    quorum_threshold, task_ids
)
VALUES (
    1,
    1,
    12345,
    '0xquorumtxhash123',
    ['0x3234567890123456789012345678901234567890'],
    10000,
    0.66,
    {1, 2}
);

-- Insert sample task data
INSERT INTO task_data (
    task_id, job_id, task_no, quorum_id,
    quorum_number, quorum_threshold, task_created_block,
    task_created_tx_hash, task_responded_block,
    task_responded_tx_hash, task_hash,
    task_response_hash, quorum_keeper_hash
)
VALUES (
    1,
    1,
    1,
    1,
    1,
    0.66,
    12345,
    '0xtaskcreated123',
    12346,
    '0xtaskresponded123',
    '0xtaskhash123',
    '0xresponsehash123',
    '0xkeeperhash123'
);

-- Insert sample task history
INSERT INTO task_history (
    task_id, quorum_id, keepers,
    responses, consensus_method,
    validation_status, tx_hash
)
VALUES (
    1,
    1,
    ['0x3234567890123456789012345678901234567890'],
    ['response1', 'response2'],
    'majority',
    true,
    '0xhistorytxhash123'
);
