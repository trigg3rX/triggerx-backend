package aggregator

import (
	"fmt"
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
}
