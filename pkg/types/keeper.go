package types

import "time"


// Passed By Manager, Received By Keeper, before Executing Action
type TriggerData struct {
	TaskID               int64                  `json:"task_id"`
	Timestamp            time.Time              `json:"timestamp"`
	
	LastExecuted         time.Time              `json:"last_executed"`
	TimeInterval         int64                  `json:"time_interval"`
	
	TriggerTxHash        string                 `json:"trigger_tx_hash"`
	
	ConditionParams      map[string]interface{} `json:"condition_params"`
}

// Created By Keeper, details about the Action
type ActionData struct {
	TaskID              int64                   `json:"task_id"`
	Timestamp           time.Time               `json:"timestamp"`
	ActionTxHash        string                  `json:"action_tx_hash"`
	GasUsed             string                  `json:"gas_used"`
	Status              bool                  `json:"status"`
	IPFSDataCID         string                  `json:"ipfs_data_cid"`
}

// Created By Keeper, details passed to Aggregator, submitted on chain upon successful consensus
type PerformerData struct {
	ProofOfTask       	string                  `json:"proof_of_task"`
	TaskDefinitionID  	string                   `json:"task_definition_id"`
	PerformerAddress  	string                  `json:"performer_address"`
	PerformerSignature 	string                  `json:"performer_signature"`
	Data              	string                  `json:"data"`
}


type ProofData struct {
	TaskID              int64                   `json:"task_id"`
	Timestamp           time.Time               `json:"timestamp"`

	ProofOfTask         string                  `json:"proof_of_task"`
	ActionDataCID       string                  `json:"action_data_cid"`

	CertificateHash     string                  `json:"certificateHash"`
	ResponseHash        string                  `json:"responseHash"`
}

type IPFSData struct {
	JobData 	HandleCreateJobData 		`json:"job_data"`
	
	TriggerData TriggerData `json:"trigger_data"`
	
	ActionData 	ActionData 	`json:"action_data"`

	ProofData 	ProofData 	`json:"proof_data"`
}

// KeeperHealth represents the health status of a keeper
type KeeperHealth struct {
	KeeperAddress string    `json:"keeper_address"`
	Version           string    `json:"version"`
	Timestamp         time.Time `json:"timestamp"`
	Signature         string    `json:"signature"`
	PeerID            string    `json:"peer_id"`
}

type UpdateKeeperHealth struct {
	KeeperAddress string    `json:"keeper_address"`
	Active        bool      `json:"active"`
	Timestamp     time.Time `json:"timestamp"`
	Version       string    `json:"version"`
	PeerID        string    `json:"peer_id"`
}