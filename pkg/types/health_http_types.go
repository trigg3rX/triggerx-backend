package types

import "time"

// KeeperInfo represents the public information about a keeper
// Owned by: Health HTTP API
type KeeperInfo struct {
	KeeperName       string    `json:"keeper_name"`
	KeeperAddress    string    `json:"keeper_address"`
	ConsensusAddress string    `json:"consensus_address"`
	OperatorID       int     `json:"operator_id"`
	Version          string    `json:"version"`
	PeerID           string    `json:"peer_id"`
	IsActive         bool      `json:"is_active"`
	LastCheckedIn    time.Time `json:"last_checked_in"`
	Network          Network   `json:"network"`
}

// KeeperHealthCheckIn represents the health check-in data from a keeper
// Owned by: Health HTTP API
type KeeperHealthCheckInRequest struct {
	KeeperAddress    string    `json:"keeper_address" validate:"required,eth_addr"`
	ConsensusPubKey  string    `json:"consensus_pub_key" validate:"required"`
	ConsensusAddress string    `json:"consensus_address" validate:"required,eth_addr"`
	Version          string    `json:"version" validate:"required"`
	Signature        string    `json:"signature" validate:"required"`
	PeerID           string    `json:"peer_id" validate:"required"`
	ConnectionAddress string    `json:"connection_address" validate:"required"`
	OperatorID       int       `json:"operator_id" validate:"required"`
	Network          Network   `json:"network" validate:"required,oneof=mainnet imua sepolia"`
}

// KeeperHealthCheckInResponse represents the response from the health check-in endpoint
// Owned by: Health HTTP API
type KeeperHealthCheckInResponse struct {
	Status bool   `json:"status"`
	Data   string `json:"data"`
}
