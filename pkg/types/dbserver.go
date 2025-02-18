package types

import (
	"math/big"
)

type CreateJobData struct {
	UserAddress            string   `json:"userAddress"`
	StakeAmount            *big.Int `json:"stakeAmount"`
	TokenAmount            *big.Int `json:"tokenAmount"`

	TaskDefinitionID       int      `json:"taskDefinitionID"`
	Priority               int      `json:"priority"`
	Security               int      `json:"security"`
	LinkJobID              int64    `json:"linkJobID"`

	TimeFrame              int64    `json:"timeFrame"`
	Recurring              bool     `json:"recurring"`

	TimeInterval           int      `json:"timeInterval"`
	TriggerChainID         int      `json:"triggerChainID"`
	TriggerContractAddress string   `json:"triggerContractAddress"`
	TriggerEvent           string   `json:"triggerEvent"`
	ScriptIPFSUrl          string   `json:"scriptIPFSUrl"`
	ScriptTriggerFunction  string   `json:"scriptTriggerFunction"`
	
	TargetChainID          int      `json:"targetChainID"`
	TargetContractAddress  string   `json:"targetContractAddress"`
	TargetFunction         string   `json:"targetFunction"`
	ArgType                int      `json:"argType"`
	Arguments              []string `json:"arguments"`
	ScriptTargetFunction   string   `json:"scriptTargetFunction"`
	
	JobCostPrediction      int      `json:"jobCostPrediction"`
}

type CreateJobResponse struct {
	UserID 				int64 `json:"userID"`
	AccountBalance 		*big.Int `json:"accountBalance"`
	TokenBalance 		*big.Int `json:"tokenBalance"`

	JobID 				int64 `json:"jobID"`
	TaskDefinitionID 	int `json:"taskDefinitionID"`
	
	TimeFrame 			int64 `json:"timeFrame"`
}

type UpdateJobData struct {
	JobID 				int64 `json:"jobID"`
	Recurring 			bool `json:"recurring"`
	TimeFrame 			int64 `json:"timeFrame"`
}

type CreateTaskData struct {
	JobID            int64  `json:"jobID"`
	TaskDefinitionID int    `json:"taskDefinitionID"`
	TaskPerformerID  int64  `json:"taskPerformerID"`
}

type CreateTaskResponse struct {
	TaskID 				int64 `json:"taskID"`
	JobID 				int64 `json:"jobID"`
	TaskDefinitionID 	int `json:"taskDefinitionID"`
	TaskPerformerID 	int64 `json:"taskPerformerID"`
	IsApproved 			bool `json:"isApproved"`
}

type GetPerformerData struct {
	KeeperID          int64  `json:"keeper_id"`
	KeeperAddress     string `json:"keeper_address"`
	ConnectionAddress string `json:"connection_address"`
}