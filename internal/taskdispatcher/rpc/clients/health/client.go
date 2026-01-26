package health

import (
	"context"
	"fmt"
	"time"

	"github.com/trigg3rX/triggerx-backend/pkg/observability"
	rpcclient "github.com/trigg3rX/triggerx-backend/pkg/rpc/client"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// Client handles gRPC communication with the health service
type Client struct {
	client *rpcclient.Client
	logger observability.Logger
}

// NewClient creates a new health service gRPC client
func NewClient(serviceURL string, logger observability.Logger, tracer observability.Tracer) (*Client, error) {
	client := rpcclient.NewClient(rpcclient.Config{
		ServiceName: serviceURL,
		Timeout:     30 * time.Second,
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

// GetPerformerData gets a performer using the dynamic selection system via gRPC
func (c *Client) GetPerformerData(ctx context.Context, network types.Network) (types.PerformerData, error) {
	c.logger.Debug(ctx, "Getting performer data from health service via gRPC", observability.String("network", string(network)))

	req := types.GetPerformerRequest{
		Network: network,
	}

	var response types.GetPerformerResponse
	err := c.client.Call(ctx, "get-performer", req, &response)
	if err != nil {
		c.logger.Error(ctx, "Failed to get performer data via gRPC", observability.Error(err))
		return types.PerformerData{}, fmt.Errorf("gRPC call failed: %w", err)
	}

	if !response.Success {
		c.logger.Error(ctx, "Health service returned error", observability.String("error", response.Error))
		return types.PerformerData{}, fmt.Errorf("health service error: %s", response.Error)
	}

	c.logger.Info(ctx, "Selected performer from health service via gRPC",
		observability.Int("operator_id", response.Performer.OperatorID),
		observability.String("keeper_address", response.Performer.KeeperAddress),
		observability.String("network", string(response.Performer.Network)))

	return response.Performer, nil
}

// Close closes the gRPC client
func (c *Client) Close(ctx context.Context) error {
	return c.client.Close(ctx)
}
