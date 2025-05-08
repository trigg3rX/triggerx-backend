from cassandra.cluster import Cluster
from cassandra.query import SimpleStatement

def transfer_data_step2():
    # Connect to the cluster
    cluster = Cluster(['localhost'])
    session = cluster.connect('triggerx')
    
    # Transfer user_data_new to user_data
    print("Transferring user_data_new to user_data...")
    rows = session.execute("SELECT * FROM user_data_new")
    for row in rows:
        session.execute(
            """
            INSERT INTO user_data (
                partition_key, user_id, user_address, created_at, job_ids, 
                account_balance, token_balance, last_updated_at, user_points
            ) VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s)
            """,
            (row.partition_key, row.user_id, row.user_address, row.created_at, row.job_ids,
             row.account_balance, row.token_balance, row.last_updated_at, row.user_points)
        )
    
    # Transfer job_data_new to job_data
    print("Transferring job_data_new to job_data...")
    rows = session.execute("SELECT * FROM job_data_new")
    for row in rows:
        session.execute(
            """
            INSERT INTO job_data (
                partition_key, job_id, task_definition_id, user_id, priority, security,
                link_job_id, chain_status, custom, time_frame, recurring,
                time_interval, trigger_chain_id, trigger_contract_address,
                trigger_event, script_ipfs_url, script_trigger_function,
                target_chain_id, target_contract_address, target_function,
                abi, arg_type, arguments, script_target_function, status,
                job_cost_prediction, created_at, last_executed_at, task_ids
            ) VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s,
                     %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
            """,
            (row.partition_key, row.job_id, row.task_definition_id, row.user_id, row.priority,
             row.security, row.link_job_id, row.chain_status, row.custom,
             row.time_frame, row.recurring, row.time_interval,
             row.trigger_chain_id, row.trigger_contract_address,
             row.trigger_event, row.script_ipfs_url, row.script_trigger_function,
             row.target_chain_id, row.target_contract_address,
             row.target_function, row.abi, row.arg_type, row.arguments,
             row.script_target_function, row.status, row.job_cost_prediction,
             row.created_at, row.last_executed_at, row.task_ids)
        )
    
    # Transfer task_data_new to task_data
    print("Transferring task_data_new to task_data...")
    rows = session.execute("SELECT * FROM task_data_new")
    for row in rows:
        session.execute(
            """
            INSERT INTO task_data (
                partition_key, task_id, task_number, job_id, task_definition_id,
                created_at, task_fee, execution_timestamp,
                execution_tx_hash, task_performer_id, proof_of_task,
                action_data_cid, task_attester_ids, is_approved,
                tp_signature, ta_signature, task_submission_tx_hash,
                is_successful
            ) VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
            """,
            (row.partition_key, row.task_id, row.task_number, row.job_id, row.task_definition_id,
             row.created_at, row.task_fee, row.execution_timestamp,
             row.execution_tx_hash, row.task_performer_id, row.proof_of_task,
             row.action_data_cid, row.task_attester_ids, row.is_approved,
             row.tp_signature, row.ta_signature, row.task_submission_tx_hash,
             row.is_successful)
        )
    
    # Transfer keeper_data_new to keeper_data
    print("Transferring keeper_data_new to keeper_data...")
    rows = session.execute("SELECT * FROM keeper_data_new")
    for row in rows:
        session.execute(
            """
            INSERT INTO keeper_data (
                partition_key, keeper_id, keeper_name, keeper_address, registered_tx,
                operator_id, rewards_address, rewards_booster,
                voting_power, keeper_points, connection_address,
                peer_id, strategies, verified, status, online,
                version, no_exctask, chat_id, email_id
            ) VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
            """,
            (row.partition_key, row.keeper_id, row.keeper_name, row.keeper_address,
             row.registered_tx, row.operator_id, row.rewards_address,
             row.rewards_booster, row.voting_power, row.keeper_points,
             row.connection_address, row.peer_id, row.strategies,
             row.verified, row.status, row.online, row.version,
             row.no_exctask, row.chat_id, row.email_id)
        )
    
    print("Data transfer completed successfully!")
    cluster.shutdown()

if __name__ == "__main__":
    transfer_data_step2() 