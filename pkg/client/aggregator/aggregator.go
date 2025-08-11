package aggregator

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"net"
	"syscall"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rpc"
	httppkg "github.com/trigg3rX/triggerx-backend/pkg/http"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/retry"
)

// AggregatorClient handles communication with the aggregator service
type AggregatorClient struct {
	logger     logging.Logger
	config     AggregatorClientConfig
	privateKey *ecdsa.PrivateKey
	publicKey  *ecdsa.PublicKey
	httpClient *httppkg.HTTPClient
	rpcClient  *rpc.Client
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
		return nil, fmt.Errorf("sender private key cannot be empty")
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
	retryConfig := httppkg.DefaultHTTPRetryConfig()

	httpClient, err := httppkg.NewHTTPClient(retryConfig, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP client: %w", err)
	}

	// Create RPC client using the HTTP client's underlying http.Client for persistent connections
	rpcClient, err := rpc.DialHTTP(cfg.AggregatorRPCUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to create RPC client: %w", err)
	}

	return &AggregatorClient{
		logger:     logger,
		config:     cfg,
		privateKey: privateKey,
		publicKey:  publicKey,
		httpClient: httpClient,
		rpcClient:  rpcClient,
	}, nil
}

// executeWithRetry executes an RPC call with retry logic using the retry package
func (c *AggregatorClient) executeWithRetry(ctx context.Context, method string, result interface{}, params CallParams) error {
	// Use the persistent RPC client for all attempts
	operation := func() (interface{}, error) {
		var err error
		switch method {
		case "sendTask":
			err = c.rpcClient.Call(
				result,
				method,
				params.ProofOfTask,
				params.Data,
				params.TaskDefinitionID,
				params.PerformerAddress,
				params.Signature,
				params.SignatureType,
				params.TargetChainID,
			)
		case "sendCustomMessage":
			err = c.rpcClient.Call(
				result,
				method,
				params.Data,
				params.TaskDefinitionID,
			)
		}
		if err != nil {
			// Classify network errors as dial failures for clearer diagnostics
			if isDialError(err) {
				return nil, fmt.Errorf("failed to dial aggregator RPC: %w", err)
			}
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

func (c *AggregatorClient) Close() {
	if c.rpcClient != nil {
		c.rpcClient.Close()
	}
	c.httpClient.Close()
}

// isDialError inspects an error to determine if it represents a network dialing error
// such as connection refused, host unreachable, or DNS resolution failures.
func isDialError(err error) bool {
	// Unwrap common network errors
	var opErr *net.OpError
	if errors.As(err, &opErr) {
		// Check for syscall-level connection refused / network unreachable
		if opErr.Err == syscall.ECONNREFUSED || opErr.Err == syscall.ENETUNREACH || opErr.Err == syscall.EHOSTUNREACH {
			return true
		}
		// Sometimes nested errors wrap syscall.Errno
		var errno syscall.Errno
		if errors.As(opErr.Err, &errno) {
			switch errno {
			case syscall.ECONNREFUSED, syscall.ENETUNREACH, syscall.EHOSTUNREACH, syscall.ETIMEDOUT:
				return true
			}
		}
		// DNS resolution failures
		var dnserr *net.DNSError
		if errors.As(opErr.Err, &dnserr) {
			return true
		}
	}
	// Direct DNS errors
	var dnserr *net.DNSError
	return errors.As(err, &dnserr)
}
