package taskmonitor

import (
	"context"
	"fmt"
	"time"

	"github.com/trigg3rX/triggerx-backend/internal/taskmonitor/types"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
	rpcclient "github.com/trigg3rX/triggerx-backend/pkg/rpc/client"
	pkgTypes "github.com/trigg3rX/triggerx-backend/pkg/types"
)

// Client is an RPC client for TaskMonitor service
type Client struct {
	rpcClient *rpcclient.Client
	logger    observability.Logger
}

// NewClient creates a new TaskMonitor RPC client
func NewClient(serverAddress string, logger observability.Logger, tracer observability.Tracer) (*Client, error) {
	config := rpcclient.Config{
		ServiceName: serverAddress,
		Timeout:     30 * time.Second,
		MaxRetries:  3,
		RetryDelay:  1 * time.Second,
		PoolSize:    10,
		PoolTimeout: 5 * time.Second,
	}

	rpcClient := rpcclient.NewClient(config, logger, tracer)

	return &Client{
		rpcClient: rpcClient,
		logger:    logger,
	}, nil
}

// ReportConsensusEvent reports a consensus event (TaskSubmitted or TaskRejected) to TaskMonitor
// The EventMonitor fetches IPFS data (containing trace context) and passes it here
func (c *Client) ReportConsensusEvent(ctx context.Context, txHash string, isAccepted bool, ipfsData *pkgTypes.IPFSData, ipfsCID string) error {
	req := &types.ReportConsensusEventRequest{
		TxHash:     txHash,
		IsAccepted: isAccepted,
		IPFSData:   ipfsData,
		IPFSCID:    ipfsCID,
	}

	var resp types.ReportConsensusEventResponse
	if err := c.rpcClient.Call(ctx, "report-consensus-event", req, &resp); err != nil {
		return fmt.Errorf("failed to call task monitor: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("task monitor returned error: %s", resp.Message)
	}

	return nil
}

// Close closes the RPC client
func (c *Client) Close(ctx context.Context) error {
	return c.rpcClient.Close(ctx)
}
