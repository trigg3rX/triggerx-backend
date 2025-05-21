package types

import (
	"time"
)

// KeeperHealthCheckIn represents the health check-in data from a keeper
type KeeperHealthCheckIn struct {
	KeeperAddress    string    `json:"keeper_address" validate:"required,eth_addr"`
	ConsensusAddress string    `json:"consensus_address" validate:"required,eth_addr"`
	Version          string    `json:"version" validate:"required"`
	Timestamp        time.Time `json:"timestamp" validate:"required"`
	Signature        string    `json:"signature" validate:"required"`
	PeerID           string    `json:"peer_id" validate:"required"`
}

// KeeperInfo represents the public information about a keeper
type KeeperInfo struct {
	KeeperName       string    `json:"keeper_name"`
	KeeperAddress    string    `json:"keeper_address"`
	ConsensusAddress string    `json:"consensus_address"`
	OperatorID       string    `json:"operator_id"`
	Version          string    `json:"version"`
	PeerID           string    `json:"peer_id"`
	IsActive         bool      `json:"is_active"`
	LastCheckedIn   time.Time `json:"last_checked_in"`
}