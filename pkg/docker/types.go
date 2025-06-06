package docker

import "time"

type ResourceStats struct {
	MemoryUsage       uint64  `json:"memory_usage"`
	CPUPercentage     float64 `json:"cpu_percentage"`
	NetworkRx         uint64  `json:"network_rx"`
	NetworkTx         uint64  `json:"network_tx"`
	BlockRead         uint64  `json:"block_read"`
	BlockWrite        uint64  `json:"block_write"`
	RxBytes           uint64  `json:"rx_bytes"`
	RxPackets         uint64  `json:"rx_packets"`
	RxErrors          uint64  `json:"rx_errors"`
	RxDropped         uint64  `json:"rx_dropped"`
	TxBytes           uint64  `json:"tx_bytes"`
	TxPackets         uint64  `json:"tx_packets"`
	TxErrors          uint64  `json:"tx_errors"`
	TxDropped         uint64  `json:"tx_dropped"`
	BandwidthRate     float64 `json:"bandwidth_rate"`
	TotalCost         float64 `json:"total_cost"`
	StaticComplexity  float64 `json:"static_complexity"`
	DynamicComplexity float64 `json:"dynamic_complexity"`
	ExecutionTime     time.Duration `json:"execution_time"`
}

type ExecutionResult struct {
	Stats    ResourceStats `json:"stats"`
	Output   string       `json:"output"`
	Success  bool         `json:"success"`
	Error    error        `json:"error,omitempty"`
	Warnings []string     `json:"warnings,omitempty"`
}
