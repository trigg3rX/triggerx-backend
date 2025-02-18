package types

import "time"

type Job struct {
	JobID                 int64                 `json:"job_id"`
	TaskDefinitionID      int                   `json:"task_definition_id"`
	Priority              int                   `json:"priority"`
	Security              int                   `json:"security"`
	TimeFrame             int64                 `json:"time_frame"`
	Recurring             bool                  `json:"recurring"`
	LinkJobID             int64                 `json:"link_job_id"`

	TimeInterval          int64                 `json:"time_interval"`
	TriggerChainID        int                   `json:"trigger_chain_id"`
	TriggerContractAddress string               `json:"trigger_contract_address"`
	TriggerEvent          string                `json:"trigger_event"`
	ScriptIPFSUrl         string                `json:"script_ipfs_url"`
	ScriptTriggerFunction string                `json:"script_trigger_function"`

	TargetChainID         int                   `json:"target_chain_id"`
	TargetContractAddress string                `json:"target_contract_address"`
	TargetFunction        string                `json:"target_function"`
	ArgType               int                   `json:"arg_type"`
	Arguments             map[string]interface{} `json:"arguments"`
	ScriptTargetFunction  string                `json:"script_target_function"`

	CreatedAt             time.Time             `json:"created_at"`
	LastExecuted          time.Time             `json:"last_executed"`

	Error                 string                `json:"error"`
	Payload              map[string]interface{} `json:"payload"`
}

type TriggerData struct {
	TaskID               int64                  `json:"task_id"`
	Timestamp            time.Time              `json:"timestamp"`
	
	LastExecuted         time.Time              `json:"last_executed"`
	TimeInterval         int64                  `json:"time_interval"`
	
	TriggerTxHash        string                 `json:"trigger_tx_hash"`
	
	ConditionParams      map[string]interface{} `json:"condition_params"`
}

type ActionData struct {
	TaskID              int64                   `json:"task_id"`
	Timestamp           time.Time               `json:"timestamp"`

	Performer           string                  `json:"performer"`
	PerfomerSignature   string                  `json:"perfomer_signature"`

	ActionTxHash        string                  `json:"action_tx_hash"`
	GasUsed             string                  `json:"gas_used"`
	Status              bool                  `json:"status"`
}

type ProofData struct {
	TaskID              int64                   `json:"task_id"`
	Timestamp           time.Time               `json:"timestamp"`

	ProofOfTask         string                  `json:"proof_of_task"`
	ActionDataCID       string                  `json:"action_data_cid"`

	CertificateHash     string                  `json:"certificateHash"`
	ResponseHash        string                  `json:"responseHash"`
}