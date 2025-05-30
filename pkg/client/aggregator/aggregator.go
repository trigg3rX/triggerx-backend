package aggregator

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rpc"
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
	AggregatorRPCAddress string
	SenderPrivateKey     string
	SenderAddress        string
	RetryAttempts        int
	RetryDelay           time.Duration
	RequestTimeout       time.Duration
}

// taskParams represents the parameters for sending a task
type taskParams struct {
	proofOfTask      string
	data             string
	taskDefinitionID int
	performerAddress string
	signature        string
}

// AggregatorClient handles communication with the aggregator service
type AggregatorClient struct {
	logger     logging.Logger
	config     AggregatorClientConfig
	privateKey *ecdsa.PrivateKey
	publicKey  *ecdsa.PublicKey
}

// NewAggregatorClient creates a new instance of AggregatorClient
func NewAggregatorClient(logger logging.Logger, cfg AggregatorClientConfig) (*AggregatorClient, error) {
	if logger == nil {
		return nil, fmt.Errorf("logger cannot be nil")
	}
	if cfg.AggregatorRPCAddress == "" {
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

	return &AggregatorClient{
		logger:     logger,
		config:     cfg,
		privateKey: privateKey,
		publicKey:  publicKey,
	}, nil
}

// Close cleans up any resources used by the client
func (c *AggregatorClient) Close() error {
	// Currently no resources to clean up
	return nil
}

// getPerformerAddress returns the performer's Ethereum address
func (c *AggregatorClient) getPerformerAddress() string {
	return crypto.PubkeyToAddress(*c.publicKey).Hex()
}

// signMessage signs the given data with the client's private key
func (c *AggregatorClient) signMessage(data []byte) (string, error) {
	messageHash := crypto.Keccak256Hash(data)

	sig, err := crypto.Sign(messageHash.Bytes(), c.privateKey)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrSigningFailed, err)
	}

	sig[64] += 27
	return hexutil.Encode(sig), nil
}

// executeWithRetry executes an RPC call with retry logic
func (c *AggregatorClient) executeWithRetry(ctx context.Context, method string, result interface{}, params struct {
	ProofOfTask      string `json:"proofOfTask"`
	Data             string `json:"data"`
	TaskDefinitionID int    `json:"taskDefinitionId"`
	PerformerAddress string `json:"performerAddress"`
	Signature        string `json:"signature"`
}) error {
	var lastErr error

	for attempt := 0; attempt < c.config.RetryAttempts; attempt++ {
		rpcClient, err := rpc.Dial(c.config.AggregatorRPCAddress)
		if err != nil {
			return fmt.Errorf("failed to dial aggregator RPC: %w", err)
		}
		defer rpcClient.Close()

		// Create a context with timeout for this attempt
		ctxWithTimeout, cancel := context.WithTimeout(ctx, c.config.RequestTimeout)
		defer cancel()

		err = rpcClient.CallContext(ctxWithTimeout, result, method,
			params.ProofOfTask,
			params.Data,
			params.TaskDefinitionID,
			params.PerformerAddress,
			params.Signature)

		if err == nil {
			return nil
		}

		lastErr = err
		c.logger.Warn("RPC request failed, retrying",
			"attempt", attempt+1,
			"maxAttempts", c.config.RetryAttempts,
			"error", err)

		// Check if context is cancelled before sleeping
		if ctx.Err() != nil {
			return ctx.Err()
		}

		if attempt < c.config.RetryAttempts-1 {
			time.Sleep(c.config.RetryDelay)
		}
	}

	return fmt.Errorf("all retry attempts failed: %w", lastErr)
}
