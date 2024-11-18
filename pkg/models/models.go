package models

import (
	// "time"
)

type UserData struct {
    UserID      int64   `json:"user_id"`
    UserAddress string  `json:"user_address"`
    JobIDs      []int64 `json:"job_ids"`
}

type JobData struct {
    JobID             int64    `json:"job_id"`
    JobType           int      `json:"jobType"`
    UserID            int64    `json:"user_id"`
    ChainID           int      `json:"chain_id"`
    TimeFrame         int64    `json:"time_frame"`
    TimeInterval      int      `json:"time_interval"`
    ContractAddress   string   `json:"contract_address"`
    TargetFunction    string   `json:"target_function"`
    ArgType           int      `json:"arg_type"`
    Arguments         []string `json:"arguments"`
    Status            bool     `json:"status"`
    JobCostPrediction float64  `json:"job_cost_prediction"`
}

type TaskData struct {
    TaskID              int64   `json:"task_id"`
    JobID               int64   `json:"job_id"`
    TaskNo              int     `json:"task_no"`
    QuorumID            int64   `json:"quorum_id"`
    QuorumNumber        int     `json:"quorum_number"`
    QuorumThreshold     float64 `json:"quorum_threshold"`
    TaskCreatedBlock    int64   `json:"task_created_block"`
    TaskCreatedTxHash   string  `json:"task_created_tx_hash"`
    TaskRespondedBlock  int64   `json:"task_responded_block"`
    TaskRespondedTxHash string  `json:"task_responded_tx_hash"`
    TaskHash            string  `json:"task_hash"`
    TaskResponseHash    string  `json:"task_response_hash"`
    QuorumKeeperHash    string  `json:"quorum_keeper_hash"`
} 