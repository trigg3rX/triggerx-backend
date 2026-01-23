package types

// CreateKeeperData represents the request to create a new keeper
type CreateKeeperData struct {
	KeeperName    string `json:"keeper_name"`
	KeeperAddress string `json:"keeper_address"`
	EmailID       string `json:"email_id"`
}

// GoogleFormCreateKeeperData represents keeper creation data from Google Form
type GoogleFormCreateKeeperData struct {
	KeeperAddress  string `json:"keeper_address" validate:"required,ethereum_address"`
	RewardsAddress string `json:"rewards_address" validate:"required,ethereum_address"`
	KeeperName     string `json:"keeper_name" validate:"required,min=3,max=50"`
	EmailID        string `json:"email_id" validate:"required,email"`
	OnImua         bool   `json:"on_imua"`
}

// UpdateKeeperChatIDRequest represents the request to update keeper chat ID
type UpdateKeeperChatIDRequest struct {
	KeeperAddress string `json:"keeper_address"`
	ChatID        int64  `json:"chat_id"`
}

// KeeperCommunicationInfo represents keeper communication information
type KeeperCommunicationInfo struct {
	ChatID     int64  `json:"chat_id"`
	KeeperName string `json:"keeper_name"`
	EmailID    string `json:"email_id"`
}

// KeeperLeaderboardEntry represents a keeper leaderboard entry
type KeeperLeaderboardEntry struct {
	KeeperAddress   string  `json:"keeper_address"`
	KeeperName      string  `json:"keeper_name"`
	NoExecutedTasks int64   `json:"no_executed_tasks"`
	NoAttestedTasks int64   `json:"no_attested_tasks"`
	KeeperPoints    float64 `json:"keeper_points"`
	OnImua          bool    `json:"on_imua"`
}

// GetPerformerData represents performer data
type GetPerformerData struct {
	KeeperAddress string `json:"keeper_address"`
}
