USE triggerx;

// Dummy API keys for two users
INSERT INTO apikeys (key, owner, is_active, rate_limit, success_count, failed_count, last_used, created_at)
VALUES ('TGRX-ADMIN-1234', 'admin', true, 120, 42, 3, '2026-01-20T12:00:00Z', '2026-01-10T09:00:00Z');

// User profiles referencing jobs below
INSERT INTO user_data (user_address, email_id, job_ids, user_points, total_jobs, total_tasks, created_at, last_updated_at)
VALUES (
    '0x1f213eb0C49Ff9810F06b44A3ed6389DC6A2c28d',
    'alice@example.com',
    {'1001', '1002'},
    '2000000000000000000',
    2,
    3,
    '2026-01-10T09:00:00Z',
    '2026-01-20T12:05:00Z'
);

INSERT INTO user_data (user_address, email_id, job_ids, user_points, total_jobs, total_tasks, created_at, last_updated_at)
VALUES (
    '0x73Eb3a01a72602C9a8cE2DaCfe8b51B7DC6ab882',
    'bob@example.com',
    {'2001'},
    '750000000000000000',
    1,
    1,
    '2026-01-12T11:30:00Z',
    '2026-01-18T16:45:00Z'
);

// Safe addresses owned by users, referenced by jobs
INSERT INTO safe_addresses (user_address, safe_address, safe_name, created_at)
VALUES ('0x1f213eb0C49Ff9810F06b44A3ed6389DC6A2c28d', '0x710eBf5F12FBCFBA31b545235D905B7942286f33', 'Alice Main Safe', '2026-01-09T08:00:00Z');

INSERT INTO safe_addresses (user_address, safe_address, safe_name, created_at)
VALUES ('0x1f213eb0C49Ff9810F06b44A3ed6389DC6A2c28d', '0xe093BfA19f4A1817054bAB9eBe41Bf4aA7aE5145', 'Alice Shared Safe', '2026-01-11T14:15:00Z');

INSERT INTO safe_addresses (user_address, safe_address, safe_name, created_at)
VALUES ('0x73Eb3a01a72602C9a8cE2DaCfe8b51B7DC6ab882', '0x427a41327c4b3f1455a55fBe7F1113707e4ec88E', 'Bob Trading Safe', '2026-01-13T10:10:00Z');

// Jobs with task_ids referencing task_data entries below
// Job 1001 - Time-based static job (TDI 1) for Alice
INSERT INTO job_data (job_id, job_title, task_definition_id, created_chain_id, user_address, link_job_id, chain_status,
                      safe_address, timezone, job_type, time_frame, recurring, status,
                      job_cost_prediction, job_cost_actual, task_ids, created_at, updated_at)
VALUES (
    '1001',
    'Daily ETH Balance Check',
    1,
    '1',
    '0x1f213eb0C49Ff9810F06b44A3ed6389DC6A2c28d',
    '',
    0,
    '0x710eBf5F12FBCFBA31b545235D905B7942286f33',
    'America/New_York',
    'frontend',
    86400,
    true,
    'running',
    '500000000000000',
    '750000000000000',
    {5001, 5002},
    '2026-01-10T09:30:00Z',
    '2026-01-20T12:00:00Z'
);

// Job 1002 - Event-based agent job (TDI 8) for Alice
INSERT INTO job_data (job_id, job_title, task_definition_id, created_chain_id, user_address, link_job_id, chain_status,
                      safe_address, timezone, job_type, time_frame, recurring, status,
                      job_cost_prediction, job_cost_actual, task_ids, created_at, updated_at)
VALUES (
    '1002',
    'USDC Whale Transfer Watcher',
    8,
    '1',
    '0x1f213eb0C49Ff9810F06b44A3ed6389DC6A2c28d',
    '',
    0,
    '0xe093BfA19f4A1817054bAB9eBe41Bf4aA7aE5145',
    'America/Los_Angeles',
    'sdk',
    604800,
    true,
    'completed',
    '1000000000000000',
    '1000000000000000',
    {5003},
    '2026-01-11T10:00:00Z',
    '2026-01-19T18:00:00Z'
);

// Job 2001 - Condition-based agent job (TDI 9) for Bob
INSERT INTO job_data (job_id, job_title, task_definition_id, created_chain_id, user_address, link_job_id, chain_status,
                      safe_address, timezone, job_type, time_frame, recurring, status,
                      job_cost_prediction, job_cost_actual, task_ids, created_at, updated_at)
VALUES (
    '2001',
    'ETH Price Take-Profit',
    9,
    '1',
    '0x73Eb3a01a72602C9a8cE2DaCfe8b51B7DC6ab882',
    '',
    0,
    '0x427a41327c4b3f1455a55fBe7F1113707e4ec88E',
    'Europe/Berlin',
    'frontend',
    2592000,
    false,
    'running',
    '1500000000000000',
    '1500000000000000',
    {6001},
    '2026-01-13T10:30:00Z',
    '2026-01-18T16:40:00Z'
);

// Tasks linked to jobs above
INSERT INTO task_data (task_id, task_number, task_status, task_error, job_id, task_definition_id,
                       created_at, executed_at, submitted_at,
                       task_opx_predicted_cost, task_opx_actual_cost,
                       execution_tx_hash, submission_tx_hash,
                       converted_arguments, task_performer_address, task_attester_address,
                       proof_of_task, is_successful, is_accepted, network)
VALUES (
    5001,
    1,
    'completed',
    '',
    '1001',
    1,
    '2026-01-10T10:00:00Z',
    '2026-01-10T10:05:00Z',
    '2026-01-10T10:06:00Z',
    '250000000000000',
    '260000000000000',
    '0xcbd95efb05fbc0d6f395bbf5c377943ff33bb1ef3a5a467da9b7f2abbb5e4694',
    '0xa8966b504b33eb5ef0c02e362598b8075e288846fb65fb50463266173ac23408',
    ['0x710eBf5F12FBCFBA31b545235D905B7942286f33', '1000000000000000000'],
    ['0x0a067a261c5f5e8c4c0b9137430b4fe1255eb62e'],
    ['0x011fcbae5f306cd793456ab7d4c0cc86756c693d'],
    'QmProofTask5001',
    true,
    true,
    'sepolia'
);

INSERT INTO task_data (task_id, task_number, task_status, task_error, job_id, task_definition_id,
                       created_at, executed_at, submitted_at,
                       task_opx_predicted_cost, task_opx_actual_cost,
                       execution_tx_hash, submission_tx_hash,
                       converted_arguments, task_performer_address, task_attester_address,
                       proof_of_task, is_successful, is_accepted, network)
VALUES (
    5002,
    2,
    'failed',
    'Balance RPC timeout',
    '1001',
    1,
    '2026-01-11T10:00:00Z',
    '2026-01-11T10:05:30Z',
    '2026-01-11T10:06:30Z',
    '250000000000000',
    '250000000000000',
    '',
    '',
    ['0x710eBf5F12FBCFBA31b545235D905B7942286f33', '1000000000000000000'],
    ['0x0a067a261c5f5e8c4c0b9137430b4fe1255eb62e'],
    ['0x011fcbae5f306cd793456ab7d4c0cc86756c693d'],
    'QmProofTask5002',
    false,
    false,
    'sepolia'
);

INSERT INTO task_data (task_id, task_number, task_status, task_error, job_id, task_definition_id,
                       created_at, executed_at, submitted_at,
                       task_opx_predicted_cost, task_opx_actual_cost,
                       execution_tx_hash, submission_tx_hash,
                       converted_arguments, task_performer_address, task_attester_address,
                       proof_of_task, is_successful, is_accepted, network)
VALUES (
    5003,
    3,
    'completed',
    '',
    '1002',
    8,
    '2026-01-19T17:00:00Z',
    '2026-01-19T17:02:00Z',
    '2026-01-19T17:03:00Z',
    '1000000000000000',
    '1000000000000000',
    '0x08c5dac5823648b77662d808c59c496543160311fc88241cb050bb0c23f5d94b',
    '0x4cdb83877eae84b669c805ed30d2165782b5e7155fee1e915b188961ddcdc5b0',
    ['0xWhaleAddress', '1000000000000000000000'],
    ['0x011fcbae5f306cd793456ab7d4c0cc86756c693d'],
    ['0x0a067a261c5f5e8c4c0b9137430b4fe1255eb62e'],
    'QmProofTask5003',
    true,
    true,
    'mainnet'
);

INSERT INTO task_data (task_id, task_number, task_status, task_error, job_id, task_definition_id,
                       created_at, executed_at, submitted_at,
                       task_opx_predicted_cost, task_opx_actual_cost,
                       execution_tx_hash, submission_tx_hash,
                       converted_arguments, task_performer_address, task_attester_address,
                       proof_of_task, is_successful, is_accepted, network)
VALUES (
    6001,
    1,
    'completed',
    '',
    '2001',
    9,
    '2026-01-18T16:00:00Z',
    '2026-01-18T16:05:00Z',
    '2026-01-18T16:06:00Z',
    '1500000000000000',
    '1500000000000000',
    '0x56c6cd882a02c995a3c1c4fd65d11f4bfb60443bcc4b78475565c9f590906813',
    '0xaf4720954e8a5b0f16a5ea6b0bad59d91e987c02efa96179fd5f5811d1f8ec14',
    ['ETHUSD', '2500.50'],
    ['0x0a067a261c5f5e8c4c0b9137430b4fe1255eb62e'],
    ['0x011fcbae5f306cd793456ab7d4c0cc86756c693d'],
    'QmProofTask6001',
    true,
    true,
    'mainnet'
);

// Keeper registry entries referenced from task_data and agent_script_executions
INSERT INTO keeper_data (keeper_address, keeper_name, rewards_address, consensus_address, operator_id,
                         voting_power, registered, registered_at, online, version, network,
                         connection_address, peer_id, uptime, last_checked_in, rewards_booster,
                         no_executed_tasks, no_attested_tasks, keeper_points, chat_id, email_id)
VALUES (
    '0x0a067a261c5f5e8c4c0b9137430b4fe1255eb62e',
    'Keeper One',
    '0x0a067a261c5f5e8c4c0b9137430b4fe1255eb62eRewards',
    '0x0a067a261c5f5e8c4c0b9137430b4fe1255eb62eConsensus',
    1,
    '100000000000000000000',
    true,
    ['2026-01-05T09:00:00Z'],
    true,
    '1.0.0',
    'sepolia',
    'keeper-one.triggerx.local:9000',
    'peer-id-keeper-one',
    86400,
    '2026-01-20T12:00:00Z',
    1.0,
    3,
    3,
    '500000000000000000',
    1000000001,
    'keeper.one@example.com'
);

INSERT INTO keeper_data (keeper_address, keeper_name, rewards_address, consensus_address, operator_id,
                         voting_power, registered, registered_at, online, version, network,
                         connection_address, peer_id, uptime, last_checked_in, rewards_booster,
                         no_executed_tasks, no_attested_tasks, keeper_points, chat_id, email_id)
VALUES (
    '0x011fcbae5f306cd793456ab7d4c0cc86756c693d',
    'Keeper Two',
    '0x011fcbae5f306cd793456ab7d4c0cc86756c693dRewards',
    '0x011fcbae5f306cd793456ab7d4c0cc86756c693dConsensus',
    2,
    '80000000000000000000',
    true,
    ['2026-01-06T10:30:00Z'],
    true,
    '1.0.0',
    'mainnet',
    'keeper-two.triggerx.local:9000',
    'peer-id-keeper-two',
    43200,
    '2026-01-20T11:45:00Z',
    1.0,
    2,
    2,
    '400000000000000000',
    1000000002,
    'keeper.two@example.com'
);

// Agent script executions associated with agent jobs (TDI 8, 9) and tasks 5003, 6001
INSERT INTO agent_script_executions (
    execution_id, job_id, task_id, task_definition_id,
    scheduled_time, actual_time, performer_address,
    input_timestamp, input_storage, input_hash,
    trigger_data,
    should_execute, target_contract, calldata, output_hash,
    execution_metadata,
    script_hash, signature,
    tx_hash, execution_status, execution_error,
    verification_status, challenge_deadline, is_challenged, challenge_count,
    created_at
)
VALUES (
    'exec-uuid-1002-1',
    '1002',
    5003,
    8,
    '2026-01-19T17:00:00Z',
    '2026-01-19T17:02:00Z',
    '0x011fcbae5f306cd793456ab7d4c0cc86756c693d',
    1737306000,
    '{"lastSeenBlock": 19500000}',
    '0xinputhash5003',
    '{"eventData":{"from":"0xWhale","to":"0xe093BfA19f4A1817054bAB9eBe41Bf4aA7aE5145","value":"1000000000000000000000"},"blockNumber":19500001}',
    true,
    '0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48',
    '0xcalldata_usdc_approve',
    '0xoutputhash5003',
    '{"apiCalls":[{"url":"https://api.etherscan.io","blockNumber":19500001}]}',
    '0xscripthash_job1002',
    '0xsignature_keeper_two_5003',
    '0x08c5dac5823648b77662d808c59c496543160311fc88241cb050bb0c23f5d94b',
    'success',
    '',
    'verified',
    '2026-01-19T23:02:00Z',
    false,
    0,
    '2026-01-19T17:02:05Z'
);

INSERT INTO agent_script_executions (
    execution_id, job_id, task_id, task_definition_id,
    scheduled_time, actual_time, performer_address,
    input_timestamp, input_storage, input_hash,
    trigger_data,
    should_execute, target_contract, calldata, output_hash,
    execution_metadata,
    script_hash, signature,
    tx_hash, execution_status, execution_error,
    verification_status, challenge_deadline, is_challenged, challenge_count,
    created_at
)
VALUES (
    'exec-uuid-2001-1',
    '2001',
    6001,
    9,
    '2026-01-18T16:00:00Z',
    '2026-01-18T16:05:00Z',
    '0x0a067a261c5f5e8c4c0b9137430b4fe1255eb62e',
    1737216000,
    '{"lastPrice":"2400.00"}',
    '0xinputhash6001',
    '{"conditionValue":2500.50,"conditionType":"price","timestamp":"2026-01-18T16:00:00Z"}',
    true,
    '0x427a41327c4b3f1455a55fBe7F1113707e4ec88E',
    '0xcalldata_take_profit',
    '0xoutputhash6001',
    '{"oracleCalls":[{"oracle":"chainlink-ethusd","blockNumber":19550000,"value":2500.50}]}',
    '0xscripthash_job2001',
    '0xsignature_keeper_one_6001',
    '0x56c6cd882a02c995a3c1c4fd65d11f4bfb60443bcc4b78475565c9f590906813',
    'success',
    '',
    'pending',
    '2026-01-18T22:05:00Z',
    true,
    1,
    '2026-01-18T16:05:05Z'
);

// Persistent script storage for agent jobs 1002 and 2001
INSERT INTO script_storage (job_id, storage_key, storage_value, updated_at)
VALUES (
    '1002',
    'lastSeenBlock',
    '19500001',
    '2026-01-19T17:02:05Z'
);

INSERT INTO script_storage (job_id, storage_key, storage_value, updated_at)
VALUES (
    '1002',
    'alertCounter',
    '3',
    '2026-01-19T17:02:05Z'
);

INSERT INTO script_storage (job_id, storage_key, storage_value, updated_at)
VALUES (
    '2001',
    'lastPrice',
    '2500.50',
    '2026-01-18T16:05:05Z'
);

INSERT INTO script_storage (job_id, storage_key, storage_value, updated_at)
VALUES (
    '2001',
    'takeProfitTriggered',
    'true',
    '2026-01-18T16:05:05Z'
);

// Challenge raised against execution exec-uuid-2001-1
INSERT INTO execution_challenges (
    challenge_id, execution_id, challenger_address, challenge_reason,
    challenger_output_hash, challenger_should_execute, challenger_target_contract,
    challenger_calldata, challenger_signature,
    bond_amount,
    resolution_status, resolution_time, validator_count, approve_count, reject_count,
    created_at
)
VALUES (
    'challenge-uuid-2001-1',
    'exec-uuid-2001-1',
    '0x52E8Ef5d5ab18FE10f4b62ae11DeD787611503E0',
    'wrong_output',
    '0xchallenger_outputhash6001',
    false,
    '0x427a41327c4b3f1455a55fBe7F1113707e4ec88E',
    '0xcalldata_take_profit_wrong',
    '0xsignature_challenger_6001',
    '100000000000000000',
    'pending',
    null,
    0,
    0,
    0,
    '2026-01-18T17:00:00Z'
);

