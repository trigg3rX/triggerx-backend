package types

// GetPerformerRequest represents the request for getting a performer
// Owned by: Health gRPC Server
type GetPerformerRequest struct {
	Network    Network `json:"network"`
}

// PerformerData represents the data for next performer to be selected
// Owned by: Health gRPC Server
type PerformerData struct {
	OperatorID    int   `json:"operator_id"`
	KeeperAddress string  `json:"keeper_address"`
	Network       Network `json:"network"`
}

// GetPerformerResponse represents the response for getting a performer
// Owned by: Health gRPC Server
type GetPerformerResponse struct {
	Performer PerformerData `json:"performer"`
	Success   bool                `json:"success"`
	Error     string              `json:"error,omitempty"`
}
