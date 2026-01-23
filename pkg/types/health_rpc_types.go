package types

// GetPerformerRequest represents the request for getting a performer
type GetPerformerRequest struct {
	Network    KeeperNetwork `json:"network"`
}

// GetPerformerResponse represents the response for getting a performer
type GetPerformerResponse struct {
	Performer PerformerData `json:"performer"`
	Success   bool                `json:"success"`
	Error     string              `json:"error,omitempty"`
}
