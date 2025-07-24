package types

import (
	"time"
)

// KeeperInfo represents the public information about a keeper
type KeeperInfo struct {
	KeeperName       string    `json:"keeper_name"`
	KeeperAddress    string    `json:"keeper_address"`
	ConsensusAddress string    `json:"consensus_address"`
	OperatorID       string    `json:"operator_id"`
	Version          string    `json:"version"`
	PeerID           string    `json:"peer_id"`
	IsActive         bool      `json:"is_active"`
	LastCheckedIn    time.Time `json:"last_checked_in"`
	IsImua           bool      `json:"is_imua"`
}
