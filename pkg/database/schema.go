package database

import (
	"github.com/gocql/gocql"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

var logger = logging.GetLogger(logging.Development, logging.DatabaseProcess)

func InitSchema(session *gocql.Session) error {
	// Create keyspace
	if err := session.Query(`
			CREATE KEYSPACE IF NOT EXISTS triggerx
			WITH replication = {
				'class': 'SimpleStrategy',
				'replication_factor': 1
			}`).Exec(); err != nil {
		return err
	}

	// Drop existing tables if any

	// Create User_data table
	if err := session.Query(`
		CREATE TABLE IF NOT EXISTS triggerx.user_data (
			user_id bigint PRIMARY KEY,
			user_address text,
			job_ids set<bigint>,
			stake_amount varint,
			account_balance varint,
			created_at timestamp,
			last_updated_at timestamp
		)`).Exec(); err != nil {
		return err
	}

	// Create Job_data table
	if err := session.Query(`
		CREATE TABLE IF NOT EXISTS triggerx.job_data (
			job_id bigint PRIMARY KEY,
			jobType int,
			user_id bigint,
			user_address text,
			chain_id int,
			time_frame bigint,
			time_interval int,
			contract_address text,
			contract_address text,
			target_function text,
			arg_type int,
			arguments list<text>,
			status boolean,
			job_cost_prediction int,
			script_function text,
			script_ipfs_url text,
			created_at timestamp,
			last_executed_at timestamp
		)`).Exec(); err != nil {
		return err
	}

	// Create Task_data table
	if err := session.Query(`
		CREATE TABLE IF NOT EXISTS triggerx.task_data (
			task_id bigint PRIMARY KEY,
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
			task_fee decimal
		)`).Exec(); err != nil {
		return err
	}

	// Create Quorum_data table
	if err := session.Query(`
		CREATE TABLE IF NOT EXISTS triggerx.quorum_data (
			quorum_id bigint PRIMARY KEY,
			quorum_no int,
			quorum_creation_block bigint,
			quorum_termination_block bigint,
			quorum_tx_hash text,
			keepers list<text>,
			quorum_stake_total bigint,
			task_ids set<bigint>,
			quorum_status boolean
		)`).Exec(); err != nil {
		return err
	}

	// Create Keeper_data table
	if err := session.Query(`
		CREATE TABLE IF NOT EXISTS triggerx.keeper_data (
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
		)`).Exec(); err != nil {
		return err
	}

	// Create Task_history table
	if err := session.Query(`
		CREATE TABLE IF NOT EXISTS triggerx.task_history (
			task_id bigint PRIMARY KEY,
			quorum_id bigint,
			keepers list<text>,
			responses list<text>,
			consensus_method text,
			validation_status boolean,
			tx_hash text
		)`).Exec(); err != nil {
		return err
	}

	logger.Info("Database schema initialized successfully")
	return nil
}
