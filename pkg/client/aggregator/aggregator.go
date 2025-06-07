package aggregator

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/retry"
)

// AggregatorClient handles communication with the aggregator service
type AggregatorClient struct {
	logger     logging.Logger
	config     AggregatorClientConfig
	privateKey *ecdsa.PrivateKey
	publicKey  *ecdsa.PublicKey
	retry      *retry.HTTPClient
}

// NewAggregatorClient creates a new instance of AggregatorClient
func NewAggregatorClient(logger logging.Logger, cfg AggregatorClientConfig) (*AggregatorClient, error) {
	if logger == nil {
		return nil, fmt.Errorf("logger cannot be nil")
	}
	if cfg.AggregatorRPCUrl == "" {
		return nil, fmt.Errorf("RPC address cannot be empty")
	}
	if cfg.SenderPrivateKey == "" {
		return nil, fmt.Errorf("private key cannot be empty")
	}
	if cfg.RequestTimeout <= 0 {
		cfg.RequestTimeout = 10 * time.Second
	}

	privateKey, err := crypto.HexToECDSA(cfg.SenderPrivateKey)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to convert private key: %v", ErrInvalidKey, err)
	}

	publicKey, ok := privateKey.Public().(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("%w: invalid public key type", ErrInvalidKey)
	}

	// Create retry client with configuration
	retryConfig := &retry.HTTPRetryConfig{
		Config: retry.Config{
			MaxRetries:      cfg.RetryAttempts,
			InitialDelay:    cfg.RetryDelay,
			BackoffFactor:   2.0,
			JitterFactor:    0.1,
			LogRetryAttempt: true,
		},
		Timeout:         cfg.RequestTimeout,
		IdleConnTimeout: 30 * time.Second,
	}

	return &AggregatorClient{
		logger:     logger,
		config:     cfg,
		privateKey: privateKey,
		publicKey:  publicKey,
		retry:      retry.NewHTTPClient(retryConfig, logger),
	}, nil
}

// executeWithRetry executes an RPC call with retry logic using the retry package
func (c *AggregatorClient) executeWithRetry(ctx context.Context, method string, result interface{}, params interface{}) error {
	// Create a new request for each attempt to ensure fresh state
	operation := func() (interface{}, error) {
		rpcClient, err := rpc.Dial(c.config.AggregatorRPCUrl)
		if err != nil {
			return nil, fmt.Errorf("failed to dial aggregator RPC: %w", err)
		}
		defer rpcClient.Close()

		// Create a context with timeout for this attempt
		ctxWithTimeout, cancel := context.WithTimeout(ctx, c.config.RequestTimeout)
		defer cancel()

		err = rpcClient.CallContext(ctxWithTimeout, result, method, params)
		if err != nil {
			return nil, fmt.Errorf("RPC call failed: %w", err)
		}

		return result, nil
	}

	_, err := retry.Retry(ctx, operation, &retry.Config{
		MaxRetries:      c.config.RetryAttempts,
		InitialDelay:    c.config.RetryDelay,
		BackoffFactor:   2.0,
		JitterFactor:    0.1,
		LogRetryAttempt: true,
	}, c.logger)

	return err
}