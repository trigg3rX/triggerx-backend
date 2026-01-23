package eventmonitor

import (
	"context"
	"fmt"
	"time"

	"github.com/trigg3rX/triggerx-backend/pkg/observability"
	rpcclient "github.com/trigg3rX/triggerx-backend/pkg/rpc/client"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// GRPCClient handles communication with the Event Monitor Service via gRPC
type GRPCClient struct {
	client *rpcclient.Client
	logger observability.Logger
}

// NewGRPCClient creates a new Event Monitor Service gRPC client
func NewGRPCClient(serviceURL string, logger observability.Logger, tracer observability.Tracer) (*GRPCClient, error) {
	client := rpcclient.NewClient(rpcclient.Config{
		ServiceName: serviceURL,
		Timeout:     30 * time.Second,
		MaxRetries:  3,
		RetryDelay:  time.Second,
		PoolSize:    10,
		PoolTimeout: 5 * time.Second,
	}, logger, tracer)

	return &GRPCClient{
		client: client,
		logger: logger,
	}, nil
}

// Register registers a monitoring request with the Event Monitor Service
func (c *GRPCClient) Register(ctx context.Context, req *types.MonitoringRequest) error {
	var response types.RegisterResponse
	if err := c.client.Call(ctx, "register", req, &response); err != nil {
		return fmt.Errorf("failed to register monitoring request: %w", err)
	}

	if !response.Success {
		return fmt.Errorf("registration failed: %s", response.Message)
	}

	c.logger.Info(ctx, "Registered monitoring request with Event Monitor Service",
		observability.String("request_id", req.RequestID),
		observability.String("chain_id", req.ChainID),
		observability.String("contract_address", req.ContractAddr))

	return nil
}

// Unregister unregisters a monitoring request from the Event Monitor Service
func (c *GRPCClient) Unregister(ctx context.Context, requestID string) error {
	req := map[string]interface{}{
		"request_id": requestID,
	}

	var response types.UnregisterResponse
	if err := c.client.Call(ctx, "unregister", req, &response); err != nil {
		return fmt.Errorf("failed to unregister monitoring request: %w", err)
	}

	if !response.Success {
		return fmt.Errorf("unregistration failed: %s", response.Message)
	}

	c.logger.Info(ctx, "Unregistered monitoring request from Event Monitor Service",
		observability.String("request_id", requestID))

	return nil
}

// Close closes the gRPC client
func (c *GRPCClient) Close() {
	ctx := context.Background()
	_ = c.client.Close(ctx)
}
