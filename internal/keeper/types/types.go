package types

import (
	"time"
)

type KeeperHealthCheckIn struct {
	KeeperAddress    string    `json:"keeper_address" validate:"required,eth_addr"`
	ConsensusAddress string    `json:"consensus_address" validate:"required,eth_addr"`
	Version          string    `json:"version" validate:"required"`
	Timestamp        time.Time `json:"timestamp" validate:"required"`
	Signature        string    `json:"signature" validate:"required"`
	PeerID           string    `json:"peer_id" validate:"required"`
}

