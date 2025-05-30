package types

import "time"

type TriggerData struct {
	TaskID    int64     `json:"task_id"`
	Timestamp time.Time `json:"timestamp"`

	LastExecuted time.Time `json:"last_executed"`
	TimeInterval int64     `json:"time_interval"`

	TriggerTxHash string `json:"trigger_tx_hash"`

	ConditionParams map[string]interface{} `json:"condition_params"`
}

type ActionData struct {
	TaskID       int64     `json:"task_id"`
	Timestamp    time.Time `json:"timestamp"`
	ActionTxHash string    `json:"action_tx_hash"`
	GasUsed      string    `json:"gas_used"`
	Status       bool      `json:"status"`
	IPFSDataCID  string    `json:"ipfs_data_cid"`

	MemoryUsage   uint64  `json:"memory_usage"`
	CPUPercentage float64 `json:"cpu_percentage"`
	NetworkRx     uint64  `json:"network_rx"`
	NetworkTx     uint64  `json:"network_tx"`
	BlockRead     uint64  `json:"block_read"`
	BlockWrite    uint64  `json:"block_write"`
	BandwidthRate float64 `json:"bandwidth_rate"`

	TotalFee          float64       `json:"total_fee"`
	StaticComplexity  float64       `json:"static_complexity"`
	DynamicComplexity float64       `json:"dynamic_complexity"`
	ComplexityIndex   float64       `json:"complexity_index"`
	ExecutionTime     time.Duration `json:"execution_time"`
}

type PerformerData struct {
	ProofOfTask        string `json:"proof_of_task"`
	TaskDefinitionID   string `json:"task_definition_id"`
	PerformerAddress   string `json:"performer_address"`
	PerformerSignature string `json:"performer_signature"`
	Data               string `json:"data"`
}

type ProofData struct {
	TaskID    int64     `json:"task_id"`
	Timestamp time.Time `json:"timestamp"`

	ProofOfTask   string `json:"proof_of_task"`
	ActionDataCID string `json:"action_data_cid"`

	CertificateHash string `json:"certificateHash"`
	ResponseHash    string `json:"responseHash"`
}

type IPFSData struct {
	JobData HandleCreateJobData `json:"job_data"`

	TriggerData TriggerData `json:"trigger_data"`

	ActionData ActionData `json:"action_data"`

	ProofData ProofData `json:"proof_data"`
}

type ProofResponse struct {
	ProofHash string `json:"proofHash"`
	CID       string `json:"cid"`
}

type TaskValidationRequest struct {
	ProofOfTask      string `json:"proofOfTask"`
	Data             string `json:"data"`
	TaskDefinitionID uint16 `json:"taskDefinitionId"`
	Performer        string `json:"performer"`
}

type ValidationResult struct {
	IsValid bool   `json:"isValid"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

type ValidationResponse struct {
	Data    bool   `json:"data"`
	Error   bool   `json:"error"`
	Message string `json:"message,omitempty"`
}
