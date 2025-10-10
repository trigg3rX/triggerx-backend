package types

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


///////////////////////////////////////////////////

type GetPerformerData struct {
	KeeperID      int64  `json:"keeper_id"`
	KeeperAddress string `json:"keeper_address"`
}
