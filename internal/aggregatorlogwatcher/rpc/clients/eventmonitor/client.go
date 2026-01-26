package eventmonitor

import (
	"context"
	"fmt"
	"time"

	"github.com/trigg3rX/triggerx-backend/pkg/observability"
	rpcclient "github.com/trigg3rX/triggerx-backend/pkg/rpc/client"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// Client handles communication with the Event Monitor Service via gRPC
type Client struct {
	client *rpcclient.Client
	logger observability.Logger
}

// NewClient creates a new Event Monitor Service gRPC client
func NewClient(serviceURL string, logger observability.Logger, tracer observability.Tracer) (*Client, error) {
	client := rpcclient.NewClient(rpcclient.Config{
		ServiceName: serviceURL,
		Timeout:     60 * time.Second, // Longer timeout for transaction processing
		MaxRetries:  3,
		RetryDelay:  time.Second,
		PoolSize:    10,
		PoolTimeout: 5 * time.Second,
	}, logger, tracer)

	return &Client{
		client: client,
		logger: logger,
	}, nil
}

// ProcessTransaction processes a transaction and extracts TaskSubmitted/TaskRejected events
func (c *Client) ProcessTransaction(ctx context.Context, req *types.ProcessTransactionRequest) (*types.ProcessTransactionResponse, error) {
	var response types.ProcessTransactionResponse
	if err := c.client.Call(ctx, "processTransaction", req, &response); err != nil {
		return nil, fmt.Errorf("failed to process transaction: %w", err)
	}

	if !response.Success {
		return &response, fmt.Errorf("transaction processing failed: %s", response.Message)
	}

	c.logger.Info(ctx, "Transaction processed successfully",
		observability.String("tx_hash", req.TxHash),
		observability.String("chain_id", req.ChainID),
		observability.String("event_name", response.EventName))

	return &response, nil
}

// Close closes the RPC client
func (c *Client) Close(ctx context.Context) error {
	return c.client.Close(ctx)
}
