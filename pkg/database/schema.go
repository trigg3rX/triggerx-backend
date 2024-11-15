package database

import (
	"github.com/gocql/gocql"
	"log"
)

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
	dropTables := []string{"tasks"}
	for _, table := range dropTables {
		if err := session.Query(`DROP TABLE IF EXISTS triggerx.` + table).Exec(); err != nil {
			return err
		}
	}

	// Create User_data table
	if err := session.Query(`
		CREATE TABLE IF NOT EXISTS triggerx.user_data (
			user_id bigint PRIMARY KEY,
			user_address text CHECK (user_address MATCHES '^0x[0-9a-fA-F]{40}$'),
			job_ids set<bigint>
		)`).Exec(); err != nil {
		return err
	}

	// Create Job_data table
	if err := session.Query(`
		CREATE TABLE IF NOT EXISTS triggerx.job_data (
				job_id bigint PRIMARY KEY,
				user_id bigint,
				chain_id int,
				stake bigint,
				time_frame timestamp,
				time_interval int,
				contract_address text CHECK (contract_address MATCHES '^0x[0-9a-fA-F]{40}$'),
				target_function text,
				arg_type int,
				arguments list<text>,
				status boolean,
				job_cost_prediction decimal
		)`).Exec(); err != nil {
		return err
	}

	// Create Task_data table
	if err := session.Query(`
		CREATE TABLE IF NOT EXISTS triggerx.task_data (
			task_ref_no text PRIMARY KEY,
			job_id bigint,
			task_id bigint,
			quorum_id int,
			quorum_no int,
			quorum_threshold decimal,
			task_hash text,
			task_response_hash text
		)`).Exec(); err != nil {
		return err
	}

	// Create Keeper_data table
	if err := session.Query(`
		CREATE TABLE IF NOT EXISTS triggerx.keeper_data (
			keeper_id bigint PRIMARY KEY,
			withdrawal_address text CHECK (withdrawal_address MATCHES '^0x[0-9a-fA-F]{40}$'),
			verified boolean,
			stake_amount bigint,
			strategies int,
			quorum_id int,
			status boolean
		)`).Exec(); err != nil {
		return err
	}

	log.Println("Database schema initialized successfully")
	return nil
} 