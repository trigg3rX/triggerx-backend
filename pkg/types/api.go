package types

import (
	"math/big"
)

type CreateJobData struct {
	UserAddress    		  string    `json:"userAddress"`
	StakeAmount    		  *big.Int  `json:"stakeAmount"`
	JobID                 int64     `json:"jobID"`
	JobType               int       `json:"jobType"`
	ChainID               int       `json:"chainID"`
	TimeFrame             int64     `json:"timeFrame"`
	TimeInterval          int       `json:"timeInterval"`
	TriggerContractAddress string    `json:"triggerContractAddress"`
	TriggerEvent          string    `json:"triggerEvent"`
	TargetContractAddress string    `json:"targetContractAddress"`
	TargetFunction        string    `json:"targetFunction"`
	ArgType               int       `json:"argType"`
	Arguments             []string  `json:"arguments"`
	Recurring             bool      `json:"recurring"`
	ScriptFunction        string    `json:"scriptFunction"`
	ScriptIPFSUrl         string    `json:"scriptIPFSUrl"`
	JobCostPrediction     int       `json:"jobCostPrediction"`
	Priority              int       `json:"priority"`
	Security              int       `json:"security"`
	LinkJobID             int64     `json:"linkJobID"`
}