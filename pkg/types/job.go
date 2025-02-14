package types

import "time"

type Job struct {
    JobID             string
    JobType           int
    UserID            string
    UserAddress       string
    ChainID           string
    TimeFrame         int64
    TimeInterval      int64
    ContractAddress   string
    TargetFunction    string
    ArgType           string
    Arguments         map[string]interface{}
    Status            string
    JobCostPrediction float64
    ScriptFunction    string
    ScriptIpfsUrl     string
    CreatedAt         time.Time
    LastExecuted      time.Time
}

type JobMessage struct {
    Job       *Job   `json:"job"`
    Timestamp string `json:"timestamp"`
}