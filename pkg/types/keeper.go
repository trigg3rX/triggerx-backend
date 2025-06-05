package types

import "time"

// Data from performer's action execution
type PerformerActionData struct {
	TaskID       int64     `json:"task_id"`
	ActionTxHash string    `json:"action_tx_hash"`
	GasUsed      string    `json:"gas_used"`
	Status       bool      `json:"status"`

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
	ExecutionTimestamp     time.Duration `json:"execution_timestamp"`
}

// Data from keeper's proof generation for execution done above
type ProofData struct {
	TaskID    int64     `json:"task_id"`
	TxTimestamp time.Time `json:"tx_timestamp"`

	ProofOfTask   string `json:"proof_of_task"`
	CertificateHash string `json:"certificate_hash"`
	ResponseHash    string `json:"response_hash"`
}

// Data to Upload to IPFS
type IPFSData struct {
	ScheduleTimeJobData ScheduleTimeJobData `json:"schedule_time_job_data"`

	SendTaskTargetData SendTaskTargetData `json:"send_task_target_data"`

	SendTriggerData SendTriggerData `json:"send_trigger_data"`

	PerformerActionData PerformerActionData `json:"performer_action_data"`

	SendProofData ProofData `json:"send_proof_data"`
}

// Data to Broadcast to Attesters from performer
type PerformerBroadcastData struct {
	ProofOfTask        string `json:"proof_of_task"`
	TaskDefinitionID   string `json:"task_definition_id"`
	PerformerAddress   string `json:"performer_address"`
	PerformerSignature string `json:"performer_signature"`
	IPFSDataCID        string `json:"ipfs_data_cid"`
}