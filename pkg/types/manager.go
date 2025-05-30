package types

import "time"

type HandleCreateJobData struct {
	JobID                  int64    `json:"job_id"`
	TaskDefinitionID       int      `json:"task_definition_id"`
	UserID                 int64    `json:"user_id"`
	LinkJobID              int64    `json:"link_job_id"`
	ChainStatus            int      `json:"chain_status"`
	JobTitle               string   `json:"job_title"`
	Custom                 bool     `json:"custom"`
	TimeFrame              int64    `json:"time_frame"`
	Recurring              bool     `json:"recurring"`
	Status                 string     `json:"status"`
	JobCostPrediction      float64  `json:"job_cost_prediction"`
	TimeInterval           int64    `json:"time_interval"`
	TriggerChainID         string   `json:"trigger_chain_id"`
	TriggerContractAddress string   `json:"trigger_contract_address"`
	TriggerEvent           string   `json:"trigger_event"`
	ScriptIPFSUrl          string   `json:"script_ipfs_url"`
	TargetChainID          string   `json:"target_chain_id"`
	TargetContractAddress  string   `json:"target_contract_address"`
	TargetFunction         string   `json:"target_function"`
	ABI                    string   `json:"abi"`
	ArgType                int      `json:"arg_type"`
	Arguments              []string `json:"arguments"`
	// Condition job specific fields
	ConditionType         string    `json:"condition_type"`
	UpperLimit            float64   `json:"upper_limit"`
	LowerLimit            float64   `json:"lower_limit"`
	ValueSourceType       string    `json:"value_source_type"`
	ValueSourceUrl        string    `json:"value_source_url"`
	CreatedAt             time.Time `json:"created_at"`
	LastExecutedAt        time.Time `json:"last_executed_at"`
	ScriptTriggerFunction string    `json:"script_trigger_function"`
}

type HandleUpdateJobData struct {
	JobID            int64 `json:"job_id"`
	TaskDefinitionID int   `json:"task_definition_id"`
	Recurring        bool  `json:"recurring"`
	TimeFrame        int64 `json:"time_frame"`
}

type HandlePauseJobData struct {
	JobID            int64 `json:"job_id"`
	TaskDefinitionID int   `json:"task_definition_id"`
}

type HandleResumeJobData struct {
	JobID            int64  `json:"job_id"`
	TaskDefinitionID int    `json:"task_definition_id"`
	LinkJobID        int64  `json:"link_job_id"`
	ChainStatus      string `json:"chain_status"`
}
