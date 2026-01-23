package types

import "time"

// MonitoringRequest represents a request to monitor a contract/event
// Used by eventmonitor service for external API requests
type MonitoringRequest struct {
	RequestID    string    `json:"request_id" binding:"required"`       // Unique request identifier
	ChainID      string    `json:"chain_id" binding:"required"`         // Chain identifier
	ContractAddr string    `json:"contract_address" binding:"required"` // Contract address to monitor
	EventSig     string    `json:"event_signature" binding:"required"`  // Event signature
	WebhookURL   string    `json:"webhook_url" binding:"required"`      // Webhook URL for notifications
	ExpiresAt    time.Time `json:"expires_at" binding:"required"`       // Expiration timestamp
	FilterParam  string    `json:"filter_param,omitempty"`              // Optional filter parameter name
	FilterValue  string    `json:"filter_value,omitempty"`              // Optional filter parameter value
}

// EventNotification represents an event notification sent to subscribers
// Used by eventmonitor service to notify subscribers of detected events
type EventNotification struct {
	RequestID    string    `json:"request_id"`       // Request identifier
	ChainID      string    `json:"chain_id"`         // Chain identifier
	ContractAddr string    `json:"contract_address"` // Contract address
	EventSig     string    `json:"event_signature"`  // Event signature
	BlockNumber  uint64    `json:"block_number"`     // Block number where event occurred
	TxHash       string    `json:"tx_hash"`          // Transaction hash
	LogIndex     uint      `json:"log_index"`        // Log index
	Topics       []string  `json:"topics"`           // Event topics
	Data         string    `json:"data"`             // Event data
	Timestamp    time.Time `json:"timestamp"`        // Event timestamp
}

// RegisterResponse represents the response for a register request
// Used by eventmonitor service for registration responses
type RegisterResponse struct {
	Success   bool   `json:"success"`    // Whether registration was successful
	RequestID string `json:"request_id"` // Request identifier
	Status    string `json:"status"`     // Registration status
	Message   string `json:"message"`    // Optional message
}

// UnregisterRequest represents a request to unregister monitoring
// Used by eventmonitor service for unregistration requests
type UnregisterRequest struct {
	RequestID string `json:"request_id" binding:"required"` // Request identifier to unregister
}

// UnregisterResponse represents the response for an unregister request
// Used by eventmonitor service for unregistration responses
type UnregisterResponse struct {
	Success   bool   `json:"success"`    // Whether unregistration was successful
	RequestID string `json:"request_id"` // Request identifier
	Status    string `json:"status"`     // Unregistration status
	Message   string `json:"message"`    // Optional message
}

// StatusResponse represents the status of a monitoring request
// Used by eventmonitor service for status queries
type StatusResponse struct {
	RequestID          string    `json:"request_id"`           // Request identifier
	Status             string    `json:"status"`               // Current status
	ChainID            string    `json:"chain_id"`             // Chain identifier
	ContractAddress    string    `json:"contract_address"`     // Contract address
	LastBlockProcessed uint64    `json:"last_block_processed"` // Last processed block number
	EventsFound        int       `json:"events_found"`         // Number of events found
	ExpiresAt          time.Time `json:"expires_at"`           // Expiration timestamp
}

// HealthResponse represents the health check response for eventmonitor
// Used by eventmonitor service for health checks
type HealthResponse struct {
	Status          string   `json:"status"`           // Service status
	Version         string   `json:"version"`          // Service version
	ActiveMonitors  int      `json:"active_monitors"`  // Number of active monitors
	ChainsSupported []string `json:"chains_supported"` // List of supported chains
}
