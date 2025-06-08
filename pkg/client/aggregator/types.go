package aggregator

import (
	"fmt"
	"time"

	"github.com/trigg3rX/triggerx-backend/internal/redis"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
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

// AggregatorClient handles communication with the aggregator service
type AggregatorClient struct {
	logger            logging.Logger
	config            AggregatorClientConfig
	privateKey        *ecdsa.PrivateKey
	publicKey         *ecdsa.PublicKey
	TaskStreamManager *redis.TaskStreamManager
}
