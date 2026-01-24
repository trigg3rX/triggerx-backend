package types

import (
	"context"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

// RegistryEntry represents a registry entry for a contract/event combination
// This is an internal implementation type and should remain in this package
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
// This is an internal implementation type and should remain in this package
type Subscriber struct {
	RequestID   string
	WebhookURL  string
	ExpiresAt   time.Time
	FilterParam string
	FilterValue string
}
