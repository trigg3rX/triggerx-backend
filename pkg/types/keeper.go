package types

import "time"

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

type IPFSData struct {
	JobData 	HandleCreateJobData 		`json:"job_data"`
	
	TriggerData TriggerData `json:"trigger_data"`
	
	ActionData 	ActionData 	`json:"action_data"`

	ProofData 	ProofData 	`json:"proof_data"`
}