package aggregator

import (
	"context"
	"crypto/ecdsa"
	"fmt"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/retry"
)

// AggregatorClient handles communication with the aggregator service
type AggregatorClient struct {
	logger            logging.Logger
	config            AggregatorClientConfig
	privateKey        *ecdsa.PrivateKey
	publicKey         *ecdsa.PublicKey
	httpClient        *retry.HTTPClient
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

	privateKey, err := crypto.HexToECDSA(cfg.SenderPrivateKey)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to convert private key: %v", ErrInvalidKey, err)
	}

	publicKey, ok := privateKey.Public().(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("%w: invalid public key type", ErrInvalidKey)
	}

	// Create retry client with configuration
	retryConfig := retry.DefaultHTTPRetryConfig()

	httpClient, err := retry.NewHTTPClient(retryConfig, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP client: %w", err)
	}

	return &AggregatorClient{
		logger:            logger,
		config:            cfg,
		privateKey:        privateKey,
		publicKey:         publicKey,
		httpClient:        httpClient,
	}, nil
}

// executeWithRetry executes an RPC call with retry logic using the retry package
func (c *AggregatorClient) executeWithRetry(ctx context.Context, method string, result interface{}, params CallParams) error {
	// Create a new request for each attempt to ensure fresh state
	operation := func() (interface{}, error) {
		rpcClient, err := rpc.Dial(c.config.AggregatorRPCUrl)
		if err != nil {
			return nil, fmt.Errorf("failed to dial aggregator RPC: %w", err)
		}
		defer rpcClient.Close()

		switch method {
		case "sendTask":
			// If this fails, we need to use params individually instead of a single params object, like params.ProofOfTask, params.Data, ... and so on
			err = rpcClient.Call(result, method, params.ProofOfTask, params.Data, params.TaskDefinitionID, params.PerformerAddress, params.Signature, params.SignatureType, params.TargetChainID)
		case "sendCustomMessage":
			err = rpcClient.Call(result, method, params.Data, params.TaskDefinitionID)
		}
		if err != nil {
			return nil, fmt.Errorf("RPC call failed: %w", err)
		}

		return result, nil
	}

	_, err := retry.Retry(ctx, operation, &retry.RetryConfig{
		MaxRetries:      c.httpClient.HTTPConfig.RetryConfig.MaxRetries,
		InitialDelay:    c.httpClient.HTTPConfig.RetryConfig.InitialDelay,
		MaxDelay:        c.httpClient.HTTPConfig.RetryConfig.MaxDelay,
		BackoffFactor:   2.0,
		JitterFactor:    0.1,
		LogRetryAttempt: true,
	}, c.logger)

	return err
}
