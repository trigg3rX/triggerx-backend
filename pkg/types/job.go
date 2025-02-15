package types

import "time"

type Job struct {
	JobID               string
	ArgType             string
	Arguments           map[string]interface{}
	ChainID             string
	ContractAddress     string
	JobCostPrediction   float64
	Stake               float64
	Status              string
	TargetFunction      string
	TargetEvent         string
	TimeFrame           int64
	TimeInterval        int64
	UserID              string
	CreatedAt           time.Time
	MaxRetries          int
	CurrentRetries      int
	LastExecuted        time.Time
	NextExecutionTime   time.Time
	Error               string
	DisputePeriodBlocks string
	Priority            int
	Security            int
	TaskIDs             []int64
	LinkID              int64
}

type JobMessage struct {
	Job       *Job   `json:"job"`
	Timestamp string `json:"timestamp"`
}
