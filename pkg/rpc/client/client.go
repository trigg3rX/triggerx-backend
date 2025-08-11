package client

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	rpcpkg "github.com/trigg3rX/triggerx-backend/pkg/rpc"
	rpcproto "github.com/trigg3rX/triggerx-backend/pkg/rpc/proto"
)

// Client represents a gRPC client
type Client struct {
	config   Config
	logger   logging.Logger
	registry rpcpkg.ServiceRegistry
	pool     *ConnectionPool
}

// Config holds client configuration
type Config struct {
	ServiceName string
	Timeout     time.Duration
	MaxRetries  int
	RetryDelay  time.Duration
	PoolSize    int
	PoolTimeout time.Duration
}

// NewClient creates a new gRPC client
func NewClient(config Config, logger logging.Logger) *Client {
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}
	if config.MaxRetries == 0 {
		config.MaxRetries = 3
	}
	if config.RetryDelay == 0 {
		config.RetryDelay = time.Second
	}
	if config.PoolSize == 0 {
		config.PoolSize = 10
	}
	if config.PoolTimeout == 0 {
		config.PoolTimeout = 5 * time.Second
	}

	return &Client{
		config: config,
		logger: logger,
		pool:   NewConnectionPool(config.PoolSize, config.PoolTimeout, logger),
	}
}

// SetRegistry sets the service registry for discovery
func (c *Client) SetRegistry(registry rpcpkg.ServiceRegistry) {
	c.registry = registry
}

// Call makes a gRPC call to the specified service and method
func (c *Client) Call(ctx context.Context, method string, request interface{}, response interface{}) error {
	retryCfg := DefaultRetryConfig()
	retryCfg.MaxRetries = c.config.MaxRetries
	if retryCfg.MaxRetries <= 0 {
		retryCfg.MaxRetries = 3
	}
	retryCfg.InitialDelay = c.config.RetryDelay
	if retryCfg.InitialDelay <= 0 {
		retryCfg.InitialDelay = 500 * time.Millisecond
	}
	retryCfg.MaxDelay = 5 * time.Second
	retryCfg.BackoffFactor = 2.0
	retryCfg.JitterFactor = 0.2

	err := RetryWithBackoff(ctx, func() error {
		// Get service info from registry or use direct connection
		var serviceInfo *rpcpkg.ServiceInfo
		var err error

		if c.registry != nil {
			serviceInfo, err = c.getServiceFromRegistry(ctx)
			if err != nil {
				return fmt.Errorf("failed to get service from registry: %w", err)
			}
		} else {
			// Use direct connection if no registry
			serviceInfo = &rpcpkg.ServiceInfo{
				Address: c.config.ServiceName, // Assume it's a direct address
			}
		}

		// Get connection from pool
		conn, err := c.pool.GetConnection(ctx, serviceInfo.Address)
		if err != nil {
			return fmt.Errorf("failed to get connection: %w", err)
		}

		// Make gRPC call
		callErr := c.makeGRPCCall(ctx, conn, method, request, response)
		// Return connection to pool (mark failed if callErr != nil)
		c.pool.ReturnConnection(conn, callErr != nil)

		return callErr
	}, retryCfg, c.logger)

	return err
}

// getServiceFromRegistry gets service info from registry
func (c *Client) getServiceFromRegistry(ctx context.Context) (*rpcpkg.ServiceInfo, error) {
	serviceInfo, err := c.registry.GetService(ctx, c.config.ServiceName)
	if err != nil {
		return nil, err
	}

	// Check if service is healthy
	if serviceInfo.Health.Status != "healthy" {
		return nil, fmt.Errorf("service %s is not healthy: %s", c.config.ServiceName, serviceInfo.Health.Status)
	}

	return serviceInfo, nil
}

// makeGRPCCall makes the actual gRPC call
func (c *Client) makeGRPCCall(ctx context.Context, conn *grpc.ClientConn, method string, request interface{}, response interface{}) error {
	// Create call context with timeout
	callCtx, cancel := context.WithTimeout(ctx, c.config.Timeout)
	defer cancel()

	// Create gRPC client
	client := rpcproto.NewGenericServiceClient(conn)

	// Convert request to Any
	var payload *anypb.Any
	if request != nil {
		if protoMsg, ok := request.(proto.Message); ok {
			var err error
			payload, err = anypb.New(protoMsg)
			if err != nil {
				return fmt.Errorf("failed to serialize request: %w", err)
			}
		} else {
			// For non-proto messages, we'll store them as-is for now
			// In a real implementation, you might want to serialize them differently
			payload = &anypb.Any{}
		}
	}

	// Create gRPC request
	grpcReq := &rpcproto.RPCRequest{
		Method:  method,
		Payload: payload,
	}

	// Make the call
	grpcResp, err := client.Call(callCtx, grpcReq)
	if err != nil {
		return fmt.Errorf("gRPC call failed: %w", err)
	}

	// Check for error in response
	if grpcResp.Error != "" {
		return fmt.Errorf("service error: %s", grpcResp.Error)
	}

	// Convert response from Any
	if response != nil && grpcResp.Result != nil {
		if protoMsg, ok := response.(proto.Message); ok {
			if err := grpcResp.Result.UnmarshalTo(protoMsg); err != nil {
				return fmt.Errorf("failed to deserialize response: %w", err)
			}
		}
		// TODO: For non-proto messages, we'll store them as-is for now
		// In a real implementation, you might want to deserialize them differently
	}

	return nil
}

// Close closes the client and connection pool
func (c *Client) Close() error {
	return c.pool.Close()
}

// HealthCheck performs a health check on the service
func (c *Client) HealthCheck(ctx context.Context) (*rpcpkg.HealthStatus, error) {
	var health rpcpkg.HealthStatus
	err := c.Call(ctx, "HealthCheck", &struct{}{}, &health)
	if err != nil {
		return nil, err
	}
	return &health, nil
}

// GetMethods returns available methods from the service
func (c *Client) GetMethods(ctx context.Context) ([]rpcpkg.RPCMethod, error) {
	var methods []rpcpkg.RPCMethod
	err := c.Call(ctx, "GetMethods", &struct{}{}, &methods)
	if err != nil {
		return nil, err
	}
	return methods, nil
}
