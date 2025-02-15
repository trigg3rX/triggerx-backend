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
    userID bigint PRIMARY KEY,
    userAddress text,
    jobIDs set<bigint>,
    stakeAmount varint,
    accountBalance varint,
    createdAt timestamp,
    lastUpdatedAt timestamp
);

-- Create Job_data table
CREATE TABLE IF NOT EXISTS job_data (
    jobID bigint PRIMARY KEY,
    jobType int,
    userID bigint,
    chainID int,
    timeFrame bigint,
    timeInterval int,
    triggerContractAddress text,
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
);

-- Create Task_data table
CREATE TABLE IF NOT EXISTS task_data (
    taskID bigint PRIMARY KEY,
    jobID bigint,
    taskDefinitionID bigint,
    taskCreatedTxHash text,
    taskRespondedTxHash text,
    taskHash text,
    taskResponseHash text,
    taskFee decimal,
    jobType text,
    blockExpiry varint,
    baseRewardFeeForAttesters varint,
    baseRewardFeeForPerformer varint,
    baseRewardFeeForAggregator varint,
    disputePeriodBlocks varint,
    minimumVotingPower varint,
    restrictedOperatorIndexes list<varint>,
    proofOfTask text,
    data blob,
    taskPerformer text,
    isApproved boolean,
    tpSignature blob,
    taSignature list<varint>,
    operatorIds list<varint>
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
    keeperID bigint PRIMARY KEY,
    keeperAddress text,
    rewardsAddress text,
    stakes list<decimal>,
    strategies list<text>,
    verified boolean,
    registeredTx text,
    status boolean,
    blsSigningKeys list<text>,
    connectionAddress text
);

-- Create Task_history table
CREATE TABLE IF NOT EXISTS task_history (
    taskID bigint PRIMARY KEY,
    performer text,
    attesters list<text>,
    proofOfTask text,
    isSuccessful boolean,
    txHash text
);
