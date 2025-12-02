package types

import (
	"context"
	"math/big"
	"time"
)

type DockerResourceStats struct {
	NoOfAttesters     int           `json:"no_of_attesters"`
	MemoryUsage       uint64        `json:"memory_usage"`
	CPUPercentage     float64       `json:"cpu_percentage"`
	NetworkRx         uint64        `json:"network_rx"`
	NetworkTx         uint64        `json:"network_tx"`
	BlockRead         uint64        `json:"block_read"`
	BlockWrite        uint64        `json:"block_write"`
	RxBytes           uint64        `json:"rx_bytes"`
	RxPackets         uint64        `json:"rx_packets"`
	RxErrors          uint64        `json:"rx_errors"`
	RxDropped         uint64        `json:"rx_dropped"`
	TxBytes           uint64        `json:"tx_bytes"`
	TxPackets         uint64        `json:"tx_packets"`
	TxErrors          uint64        `json:"tx_errors"`
	TxDropped         uint64        `json:"tx_dropped"`
	BandwidthRate     float64       `json:"bandwidth_rate"`
	TotalCost         *big.Int      `json:"total_cost"`
	CurrentTotalCost  *big.Int      `json:"current_total_cost"`
	StaticComplexity  float64       `json:"static_complexity"`
	DynamicComplexity float64       `json:"dynamic_complexity"`
	ExecutionTime     time.Duration `json:"execution_time"`
}

type ExecutionResult struct {
	Stats    DockerResourceStats `json:"stats"`
	Output   string              `json:"output"`
	Success  bool                `json:"success"`
	Error    error               `json:"error,omitempty"`
	Warnings []string            `json:"warnings,omitempty"`
}

type ExecutionState struct {
    CancelFunc  context.CancelFunc
    ExecID      string
    ContainerID string
}

type ExecutionContext struct {
    FileURL       string            `json:"file_url"`
    FileLanguage  string            `json:"file_language"`
    NoOfAttesters int               `json:"no_of_attesters"`
    TraceID       string            `json:"trace_id"`
    StartedAt     time.Time         `json:"started_at"`
    CompletedAt   time.Time         `json:"completed_at,omitempty"`
    Metadata      map[string]string `json:"metadata,omitempty"`
    State         ExecutionState    `json:"state"`
}

type ExecutionStage string

const (
	StageDownloading ExecutionStage = "downloading"
	StageValidating  ExecutionStage = "validating"
	StagePreparing   ExecutionStage = "preparing"
	StageExecuting   ExecutionStage = "executing"
	StageCompleted   ExecutionStage = "completed"
	StageFailed      ExecutionStage = "failed"
)

type ExecutionPipeline struct {
	Context      *ExecutionContext `json:"context"`
	CurrentStage ExecutionStage    `json:"current_stage"`
	Stages       []PipelineStage   `json:"stages"`
	Result       *ExecutionResult  `json:"result,omitempty"`
}

type PipelineStage struct {
	Name      ExecutionStage `json:"name"`
	StartedAt time.Time      `json:"started_at"`
	EndedAt   time.Time      `json:"ended_at,omitempty"`
	Duration  time.Duration  `json:"duration,omitempty"`
	Error     error          `json:"error,omitempty"`
}

type ValidationResult struct {
	IsValid    bool     `json:"is_valid"`
	Errors     []string `json:"errors,omitempty"`
	Warnings   []string `json:"warnings,omitempty"`
	Complexity float64  `json:"complexity"`
}
