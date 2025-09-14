package types

import (
	"math/big"
	"time"
)

// KeeperHealthCheckIn represents the health check-in data from a keeper
type KeeperHealthCheckIn struct {
	KeeperAddress    string    `json:"keeper_address" validate:"required,eth_addr"`
	ConsensusPubKey  string    `json:"consensus_pub_key" validate:"required"`
	ConsensusAddress string    `json:"consensus_address" validate:"required,eth_addr"`
	Version          string    `json:"version" validate:"required"`
	Timestamp        time.Time `json:"timestamp" validate:"required"`
	Signature        string    `json:"signature" validate:"required"`
	PeerID           string    `json:"peer_id" validate:"required"`
	IsImua           bool      `json:"is_imua" validate:"required"`
}

// KeeperHealthCheckInResponse represents the response from the health check-in endpoint
type KeeperHealthCheckInResponse struct {
	Status bool   `json:"status"`
	Data   string `json:"data"`
}

// Data from performer's action execution
type PerformerActionData struct {
	TaskID       int64  `json:"task_id"`
	ActionTxHash string `json:"action_tx_hash"`
	GasUsed      string `json:"gas_used"`
	Status       bool   `json:"status"`

	MemoryUsage   uint64  `json:"memory_usage"`
	CPUPercentage float64 `json:"cpu_percentage"`
	NetworkRx     uint64  `json:"network_rx"`
	NetworkTx     uint64  `json:"network_tx"`
	BlockRead     uint64  `json:"block_read"`
	BlockWrite    uint64  `json:"block_write"`
	BandwidthRate float64 `json:"bandwidth_rate"`

	TotalFee           *big.Int      `json:"total_fee"`
	StaticComplexity   float64       `json:"static_complexity"`
	DynamicComplexity  float64       `json:"dynamic_complexity"`
	ComplexityIndex    float64       `json:"complexity_index"`
	ExecutionTimestamp time.Time     `json:"execution_timestamp"`
	ConvertedArguments []interface{} `json:"converted_arguments"`
}

// Data from keeper's proof generation for execution done above
type ProofData struct {
	TaskID               int64     `json:"task_id"`
	ProofOfTask          string    `json:"proof_of_task"`
	CertificateHash      string    `json:"certificate_hash"`
	CertificateTimestamp time.Time `json:"certificate_timestamp"`
}

type PerformerSignatureData struct {
	TaskID                  int64  `json:"task_id"`
	PerformerSigningAddress string `json:"performer_signing_address"`
	PerformerSignature      string `json:"performer_signature"`
}

// Data to Upload to IPFS
type IPFSData struct {
	TaskData           *SendTaskDataToKeeper   `json:"task_data"`
	ActionData         *PerformerActionData    `json:"action_data"`
	ProofData          *ProofData              `json:"proof_data"`
	PerformerSignature *PerformerSignatureData `json:"performer_signature_data"`
}

// Data to Broadcast to Attesters from performer
type BroadcastDataForValidators struct {
	ProofOfTask        string `json:"proof_of_task"`
	Data               []byte `json:"data"`
	TaskDefinitionID   int    `json:"task_definition_id"`
	PerformerAddress   string `json:"performer_address"`
	PerformerSignature string `json:"performer_signature"`
	SignatureType      string `json:"signature_type"`
	TargetChainID      int    `json:"target_chain_id"`
}
