package types

import "time"

type KeeperData struct {
	KeeperID          int64     `json:"keeper_id"`
	KeeperName        string    `json:"keeper_name"`
	KeeperAddress     string    `json:"keeper_address"`
	ConsensusAddress  string    `json:"consensus_address"`
	RegisteredTx      string    `json:"registered_tx"`
	OperatorID        int64     `json:"operator_id"`
	RewardsAddress    string    `json:"rewards_address"`
	RewardsBooster    float32   `json:"rewards_booster"`
	VotingPower       int64     `json:"voting_power"`
	KeeperPoints      float64   `json:"keeper_points"`
	ConnectionAddress string    `json:"connection_address"`
	PeerID            string    `json:"peer_id"`
	Strategies        []string  `json:"strategies"`
	Whitelisted       bool      `json:"whitelisted"`
	Registered        bool      `json:"registered"`
	Online            bool      `json:"online"`
	Version           string    `json:"version"`
	NoExecutedTasks   int64     `json:"no_executed_tasks"`
	NoAttestedTasks   int64     `json:"no_attested_tasks"`
	ChatID            int64     `json:"chat_id"`
	EmailID           string    `json:"email_id"`
	LastCheckedIn     time.Time `json:"last_checked_in"`
	OnImua            bool      `json:"on_imua"`
	Uptime            int64     `json:"uptime"`
}

type CreateKeeperData struct {
	KeeperName    string `json:"keeper_name"`
	KeeperAddress string `json:"keeper_address"`
	EmailID       string `json:"email_id"`
}

// Create New Keeper from Google Form (google script)
type GoogleFormCreateKeeperData struct {
	KeeperAddress  string `json:"keeper_address" validate:"required,ethereum_address"`
	RewardsAddress string `json:"rewards_address" validate:"required,ethereum_address"`
	KeeperName     string `json:"keeper_name" validate:"required,min=3,max=50"`
	EmailID        string `json:"email_id" validate:"required,email"`
	OnImua         bool   `json:"on_imua"`
}

type UpdateKeeperChatIDRequest struct {
	KeeperAddress string `json:"keeper_address"`
	ChatID        int64  `json:"chat_id"`
}

type KeeperCommunicationInfo struct {
	ChatID     int64  `json:"chat_id"`
	KeeperName string `json:"keeper_name"`
	EmailID    string `json:"email_id"`
}

type KeeperLeaderboardEntry struct {
	KeeperID        int64   `json:"keeper_id"`
	KeeperAddress   string  `json:"keeper_address"`
	KeeperName      string  `json:"keeper_name"`
	NoExecutedTasks int64   `json:"no_executed_tasks"`
	NoAttestedTasks int64   `json:"no_attested_tasks"`
	KeeperPoints    float64 `json:"keeper_points"`
}

///////////////////////////////////////////////////

type GetPerformerData struct {
	KeeperID      int64  `json:"keeper_id"`
	KeeperAddress string `json:"keeper_address"`
}
