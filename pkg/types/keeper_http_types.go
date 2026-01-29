package types

import (
	"time"
)

// PerformerActionData represents data from performer's action execution
// Owned by: keeper (added to IPFSData struct)
type PerformerActionData struct {
	TaskID       int64  `json:"task_id"`
	ActionTxHash string `json:"action_tx_hash"`
	GasUsed      string `json:"gas_used"`
	Status       bool   `json:"status"`

	// TransactionSubmitted indicates whether an on-chain transaction was actually
	// broadcast for this action. For agent jobs (TDI 7/8/9), scripts can decide
	// not to execute on-chain (shouldExecute=false); in that case this will be false
	// and validators must not expect an on-chain transaction/receipt.
	TransactionSubmitted bool `json:"transaction_submitted"`

	MemoryUsage   uint64  `json:"memory_usage"`
	CPUPercentage float64 `json:"cpu_percentage"`
	NetworkRx     uint64  `json:"network_rx"`
	NetworkTx     uint64  `json:"network_tx"`
	BlockRead     uint64  `json:"block_read"`
	BlockWrite    uint64  `json:"block_write"`
	BandwidthRate float64 `json:"bandwidth_rate"`

	TotalFee           string        `json:"total_fee"` // Total fee in Wei (as string to handle large values)
	StaticComplexity   float64       `json:"static_complexity"`
	DynamicComplexity  float64       `json:"dynamic_complexity"`
	ComplexityIndex    float64       `json:"complexity_index"`
	ExecutionTimestamp time.Time     `json:"execution_timestamp"`
	ConvertedArguments []interface{} `json:"converted_arguments"`

	// Agentic script fields (TaskDefinitionID = 7/8/9)
	StorageUpdates       map[string]string `json:"storage_updates,omitempty"`        // Storage updates to save in DB
	ScriptTargetContract string            `json:"script_target_contract,omitempty"` // Target contract from script output
	ScriptCalldata       string            `json:"script_calldata,omitempty"`        // Calldata from script output
	ScriptMetadata       *ScriptMetadata   `json:"script_metadata,omitempty"`        // Script execution metadata
}

// ScriptMetadata contains metadata from custom script execution
// Owned by: keeper (added to IPFSData struct)
type ScriptMetadata struct {
	Timestamp   int64  `json:"timestamp"`
	Reason      string `json:"reason,omitempty"`
	GasEstimate uint64 `json:"gas_estimate,omitempty"`
}

// ProofData represents data from keeper's proof generation for execution
// Owned by: keeper (added to IPFSData struct)
type ProofData struct {
	TaskID               int64     `json:"task_id"`
	ProofOfTask          string    `json:"proof_of_task"`
	CertificateHash      string    `json:"certificate_hash"`
	CertificateTimestamp time.Time `json:"certificate_timestamp"`
}

// PerformerSignatureData represents performer signature data
// Owned by: keeper (added to IPFSData struct)
type PerformerSignatureData struct {
	TaskID                  int64  `json:"task_id"`
	PerformerSigningAddress string `json:"performer_signing_address"`
	PerformerSignature      string `json:"performer_signature"`
}

// IPFSData represents data to upload to IPFS
// Shared across: keeper, taskmonitor, eventmonitor
type IPFSData struct {
	TaskData           *SendTaskDataToKeeper   `json:"task_data"`
	ActionData         *PerformerActionData    `json:"action_data"`
	ProofData          *ProofData              `json:"proof_data"`
	PerformerSignature *PerformerSignatureData `json:"performer_signature_data"`
	// Trace context fields for end-to-end tracing
	TraceID string `json:"trace_id,omitempty"`
	SpanID  string `json:"span_id,omitempty"`
}

// BroadcastDataForValidators represents data to broadcast to attesters (validators) from performer via aggregator JSON-RPC API, SendTask method
// Owned by: keeper (used by keeper to broadcast data to attesters, via pkg/client/aggregator/task.go)
type BroadcastDataForValidators struct {
	ProofOfTask        string `json:"proof_of_task"`  // Hash of Serialized IPFSData struct
	Data               []byte `json:"data"`             // CID of IPFSData struct uploaded to IPFS
	TaskDefinitionID   int    `json:"task_definition_id"` // Task definition ID (1-9)
	PerformerAddress   string `json:"performer_address"`  // Performer address (Controller key, Consensus key would be used for BLS)
	PerformerSignature string `json:"performer_signature"` // Performer signature (ECDSA signature)
	SignatureType      string `json:"signature_type"`     // Signature type (currently only ECDSA is supported, BLS is WIP)
	TargetChainID      int    `json:"target_chain_id"`    // Base Mainnet or Sepolia based on Network type
}
