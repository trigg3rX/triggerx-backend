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
	if err := session.Query(`DROP TABLE IF EXISTS triggerx.user_data`).Exec(); err != nil {
		return err
	}
	if err := session.Query(`DROP TABLE IF EXISTS triggerx.job_data`).Exec(); err != nil {
		return err
	}
	if err := session.Query(`DROP TABLE IF EXISTS triggerx.task_data`).Exec(); err != nil {
		return err
	}
	if err := session.Query(`DROP TABLE IF EXISTS triggerx.quorum_data`).Exec(); err != nil {
		return err
	}
	if err := session.Query(`DROP TABLE IF EXISTS triggerx.keeper_data`).Exec(); err != nil {
		return err
	}
	if err := session.Query(`DROP TABLE IF EXISTS triggerx.task_history`).Exec(); err != nil {
		return err
	}

	// Create User_data table
	if err := session.Query(`
		CREATE TABLE IF NOT EXISTS triggerx.user_data (
			userID bigint PRIMARY KEY,
			userAddress text,
			jobIDs set<bigint>,
			stakeAmount varint,
			accountBalance varint,
			createdAt timestamp,
			lastUpdatedAt timestamp
		)`).Exec(); err != nil {
		return err
	}

	// Create Job_data table
	if err := session.Query(`
		CREATE TABLE IF NOT EXISTS triggerx.job_data (
			jobID bigint PRIMARY KEY,
			jobType int,
			userID bigint,
			chainID int,
			timeFrame bigint,
			timeInterval int,
			triggerxContractAddress text,
			triggerEvent text,
			targetContractAddress text,
			targetFunction text,
			argType int,
			arguments list<text>,
			recurring boolean,
			scriptFunction text,
			scriptIPFSUrl text,
			status boolean,
			jobCostPrediction int,
			createdAt timestamp,
			lastExecutedAt timestamp,
			priority int,
			security int,
			taskIDs set<bigint>,
			linkJobID bigint
		)`).Exec(); err != nil {
		return err
	}

	// Create Task_data table
	if err := session.Query(`
		CREATE TABLE IF NOT EXISTS triggerx.task_data (
			taskID bigint PRIMARY KEY,
			jobID bigint,
			taskDefinitionID bigint,
			taskRespondedTxHash text,
			taskResponseHash text,
			taskFee string,
			proofOfTask text,
			data blob,
			taskPerformer text,
			isApproved boolean,
			tpSignature blob,
			taSignature list<varint>,
			operatorIds list<varint>,
			executedAt timestamp
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
			keeperID bigint PRIMARY KEY,
			keeperAddress text,
			rewardsAddress text,
			stakes list<decimal>,
			strategies list<text>,
			verified boolean,
			registeredTx text,
			status boolean,
			blsSigningKeys list<text>,
			connectionAddress text,
			keeperType int
		)`).Exec(); err != nil {
		return err
	}

	// Create Task_history table
	if err := session.Query(`
		CREATE TABLE IF NOT EXISTS triggerx.task_history (
			taskID bigint PRIMARY KEY,
			performer text,
			attesters list<text>,
			proofOfTask text,
			isSuccessful boolean,
			txHash text
		)`).Exec(); err != nil {
		return err
	}

	logger.Info("Database schema initialized successfully")
	return nil
}
