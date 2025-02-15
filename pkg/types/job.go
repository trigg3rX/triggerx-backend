package types

import "time"

type Job struct {
	JobID               int64
	JobType             int
	ArgType             int
	Arguments           map[string]interface{}
	ChainID             int
	TriggerContractAddress     string
	TriggerEvent         string
	TargetContractAddress  string
	TargetFunction       string
	JobCostPrediction   float64
	Recurring            bool
	Status              bool
	TimeFrame           int64
	TimeInterval        int64
	UserID              int64
	CreatedAt           time.Time
	LastExecuted        time.Time
	NextExecutionTime   time.Time
	Error               string
	Priority            int
	Security            int
	TaskIDs             []int64
	LinkID              int64
	ScriptFunction      string
	ScriptIPFSUrl       string
}

type JobMessage struct {
	Job       *Job   `json:"job"`
	Timestamp string `json:"timestamp"`
}
