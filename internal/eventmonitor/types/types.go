package types

import (
	"context"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

// MonitoringRequest represents a request to monitor a contract/event
type MonitoringRequest struct {
	RequestID    string    `json:"request_id" binding:"required"`
	ChainID      string    `json:"chain_id" binding:"required"`
	ContractAddr string    `json:"contract_address" binding:"required"`
	EventSig     string    `json:"event_signature" binding:"required"`
	WebhookURL   string    `json:"webhook_url" binding:"required"`
	ExpiresAt    time.Time `json:"expires_at" binding:"required"`
	FilterParam  string    `json:"filter_param,omitempty"`
	FilterValue  string    `json:"filter_value,omitempty"`
}

// EventNotification represents an event notification sent to subscribers
type EventNotification struct {
	RequestID    string    `json:"request_id"`
	ChainID      string    `json:"chain_id"`
	ContractAddr string    `json:"contract_address"`
	EventSig     string    `json:"event_signature"`
	BlockNumber  uint64    `json:"block_number"`
	TxHash       string    `json:"tx_hash"`
	LogIndex     uint      `json:"log_index"`
	Topics       []string  `json:"topics"`
	Data         string    `json:"data"`
	Timestamp    time.Time `json:"timestamp"`
}

// RegistryEntry represents a registry entry for a contract/event combination
type RegistryEntry struct {
	Key          string
	ChainID      string
	ContractAddr common.Address
	EventSig     common.Hash
	Subscribers  map[string]*Subscriber
	LastBlock    uint64
	WorkerCtx    context.Context
	WorkerCancel context.CancelFunc
	Mu           sync.RWMutex
}

// Subscriber represents a subscriber to a contract/event
type Subscriber struct {
	RequestID   string
	WebhookURL  string
	ExpiresAt   time.Time
	FilterParam string
	FilterValue string
}

// RegisterResponse represents the response for a register request
type RegisterResponse struct {
	Success   bool   `json:"success"`
	RequestID string `json:"request_id"`
	Status    string `json:"status"`
	Message   string `json:"message"`
}

// UnregisterRequest represents a request to unregister monitoring
type UnregisterRequest struct {
	RequestID string `json:"request_id" binding:"required"`
}

// UnregisterResponse represents the response for an unregister request
type UnregisterResponse struct {
	Success   bool   `json:"success"`
	RequestID string `json:"request_id"`
	Status    string `json:"status"`
	Message   string `json:"message"`
}

// StatusResponse represents the status of a monitoring request
type StatusResponse struct {
	RequestID          string    `json:"request_id"`
	Status             string    `json:"status"`
	ChainID            string    `json:"chain_id"`
	ContractAddress    string    `json:"contract_address"`
	LastBlockProcessed uint64    `json:"last_block_processed"`
	EventsFound        int       `json:"events_found"`
	ExpiresAt          time.Time `json:"expires_at"`
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status          string   `json:"status"`
	Version         string   `json:"version"`
	ActiveMonitors  int      `json:"active_monitors"`
	ChainsSupported []string `json:"chains_supported"`
}
