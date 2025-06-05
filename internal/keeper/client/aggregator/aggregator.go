package aggregator

import (
	"context"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/rpc"
	"github.com/trigg3rX/triggerx-backend/pkg/encrypt"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

// Client represents an Aggregator RPC client
type Client struct {
	rpcClient *rpc.Client
	logger    logging.Logger
	config    Config
}

// Config holds the configuration for the Aggregator client
type Config struct {
	RPCAddress     string
	PrivateKey     string
	KeeperAddress  string
	RetryAttempts  int
	RetryDelay     time.Duration
	RequestTimeout time.Duration
}

// TaskResult represents the data to be sent to the aggregator
type TaskResult struct {
	ProofOfTask      string
	Data             string
	TaskDefinitionID int
	PerformerAddress string
}

// NewClient creates a new Aggregator client
func NewClient(logger logging.Logger, cfg Config) (*Client, error) {
	if cfg.RetryAttempts == 0 {
		cfg.RetryAttempts = 3
	}
	if cfg.RetryDelay == 0 {
		cfg.RetryDelay = 2 * time.Second
	}
	if cfg.RequestTimeout == 0 {
		cfg.RequestTimeout = 10 * time.Second
	}

	rpcClient, err := rpc.Dial(cfg.RPCAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to dial aggregator RPC: %w", err)
	}

	return &Client{
		rpcClient: rpcClient,
		logger:    logger,
		config:    cfg,
	}, nil
}

// Close closes the RPC client connection
func (c *Client) Close() {
	c.rpcClient.Close()
}

// SendTaskResult sends a task result to the aggregator
func (c *Client) SendTaskResult(ctx context.Context, result *TaskResult) error {
	// Sign the task data
	signature, err := c.signTaskData(result)
	if err != nil {
		return fmt.Errorf("failed to sign task data: %w", err)
	}

	// Prepare parameters
	params := struct {
		ProofOfTask      string `json:"proofOfTask"`
		Data             string `json:"data"`
		TaskDefinitionID int    `json:"taskDefinitionId"`
		PerformerAddress string `json:"performerAddress"`
		Signature        string `json:"signature"`
	}{
		ProofOfTask:      result.ProofOfTask,
		Data:             "0x" + hex.EncodeToString([]byte(result.Data)),
		TaskDefinitionID: result.TaskDefinitionID,
		PerformerAddress: result.PerformerAddress,
		Signature:        signature,
	}

	var response interface{}
	err = c.executeWithRetry(ctx, "sendTask", &response, params)
	if err != nil {
		return fmt.Errorf("failed to send task result: %w", err)
	}

	c.logger.Info("Successfully sent task result to aggregator",
		"taskDefinitionId", result.TaskDefinitionID,
		"proofOfTask", result.ProofOfTask)

	return nil
}

// signTaskData signs the task data using the keeper's private key
func (c *Client) signTaskData(result *TaskResult) (string, error) {
	// Pack data for signing
	message := fmt.Sprintf("%s:%s:%d:%s",
		result.ProofOfTask,
		result.Data,
		result.TaskDefinitionID,
		result.PerformerAddress)

	signature, err := encrypt.SignMessage(message, c.config.PrivateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign message: %w", err)
	}

	return signature, nil
}

// executeWithRetry executes an RPC call with retry logic
func (c *Client) executeWithRetry(ctx context.Context, method string, result interface{}, params struct {
	ProofOfTask      string `json:"proofOfTask"`
	Data             string `json:"data"`
	TaskDefinitionID int    `json:"taskDefinitionId"`
	PerformerAddress string `json:"performerAddress"`
	Signature        string `json:"signature"`
}) error {
	var lastErr error

	for attempt := 0; attempt < c.config.RetryAttempts; attempt++ {
		// Create a context with timeout for this attempt
		ctxWithTimeout, cancel := context.WithTimeout(ctx, c.config.RequestTimeout)
		defer cancel()

		err := c.rpcClient.CallContext(ctxWithTimeout, result, method,
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
