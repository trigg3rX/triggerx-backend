package types

import "time"

type HandleCreateJobData struct {
	JobID                  int64     `json:"job_id"`
	TaskDefinitionID       int       `json:"task_definition_id"`
	UserID                 int64     `json:"user_id"`
	Priority               int       `json:"priority"`
	Security               int       `json:"security"`
	LinkJobID              int64     `json:"link_job_id"`
	ChainStatus            int       `json:"chain_status"`
	JobTitle               string    `json:"job_title"`
	TimeFrame              int64     `json:"time_frame"`
	Recurring              bool      `json:"recurring"`
	TimeInterval           int64     `json:"time_interval"`
	TriggerChainID         string    `json:"trigger_chain_id"`
	TriggerContractAddress string    `json:"trigger_contract_address"`
	TriggerEvent           string    `json:"trigger_event"`
	ScriptIPFSUrl          string    `json:"script_ipfs_url"`
	ScriptTriggerFunction  string    `json:"script_trigger_function"`
	TargetChainID          string    `json:"target_chain_id"`
	TargetContractAddress  string    `json:"target_contract_address"`
	ABI                    string    `json:"abi"`
	TargetFunction         string    `json:"target_function"`
	ArgType                int       `json:"arg_type"`
	Arguments              []string  `json:"arguments"`
	ScriptTargetFunction   string    `json:"script_target_function"`
	CreatedAt              time.Time `json:"created_at"`
	LastExecutedAt         time.Time `json:"last_executed_at"`
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
