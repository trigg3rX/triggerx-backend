package types

import "time"

// HealthKeeperInfo represents the public information about a keeper in the health service
type HealthKeeperInfo struct {
	KeeperName       string    `json:"keeper_name"`
	KeeperAddress    string    `json:"keeper_address"`
	ConsensusAddress string    `json:"consensus_address"`
	OperatorID       string    `json:"operator_id"`
	Version          string    `json:"version"`
	IsActive         bool      `json:"is_active"`
	Uptime           int64     `json:"uptime"`
	LastCheckedIn    time.Time `json:"last_checked_in"`
	IsImua           bool      `json:"is_imua"`
}

// KeeperHealthCheckIn represents the health check-in data from a keeper
type HealthKeeperCheckInRequest struct {
	KeeperAddress    string    `json:"keeper_address" validate:"required,eth_addr"`
	ConsensusPubKey  string    `json:"consensus_pub_key" validate:"required"`
	ConsensusAddress string    `json:"consensus_address" validate:"required,eth_addr"`
	Version          string    `json:"version" validate:"required"`
	Timestamp        time.Time `json:"timestamp" validate:"required"`
	Signature        string    `json:"signature" validate:"required"`
	IsImua           bool      `json:"is_imua"`
}

// KeeperHealthCheckInResponse represents the response from the health check-in endpoint
type HealthKeeperCheckInResponse struct {
	Status bool   `json:"status"`
	Data   string `json:"data"`
}
