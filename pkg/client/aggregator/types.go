package aggregator

import (
	"fmt"
	"time"
)

// Common errors
var (
	ErrInvalidKey    = fmt.Errorf("invalid key")
	ErrSigningFailed = fmt.Errorf("signing operation failed")
	ErrRPCFailed     = fmt.Errorf("RPC operation failed")
	ErrMarshalFailed = fmt.Errorf("marshaling operation failed")
)

// AggregatorClientConfig holds the configuration for AggregatorClient
type AggregatorClientConfig struct {
	AggregatorRPCUrl string
	SenderPrivateKey string
	SenderAddress    string
	RetryAttempts    int
	RetryDelay       time.Duration
	RequestTimeout   time.Duration
}

type CallParams struct {
	ProofOfTask      string
	Data             string
	TaskDefinitionID int
	PerformerAddress string
	Signature        string
}