package client

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/trigg3rX/triggerx-backend/pkg/observability"
	rpcpkg "github.com/trigg3rX/triggerx-backend/pkg/rpc"
	rpcproto "github.com/trigg3rX/triggerx-backend/pkg/rpc/proto"
)

// Client represents a gRPC client
type Client struct {
	config   Config
	tracer   observability.Tracer
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

// NewClientWithTracing creates a new gRPC client with trace propagation support
func NewClient(config Config, logger observability.Logger, tracer observability.Tracer) *Client {
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
		tracer: tracer,
		pool:   NewConnectionPool(config.PoolSize, config.PoolTimeout, logger, tracer, config.ServiceName),
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
		// Return connection to pool (mark failed if it's a connection error)
		isConnectionError := isConnectionFailure(callErr)
		c.pool.ReturnConnection(ctx, conn, isConnectionError)

		return callErr
	}, retryCfg)

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
			// For non-proto messages, serialize as JSON
			jsonData, err := json.Marshal(request)
			if err != nil {
				return fmt.Errorf("failed to serialize request as JSON: %w", err)
			}
			payload = &anypb.Any{
				TypeUrl: "application/json",
				Value:   jsonData,
			}
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
		} else {
			// For non-proto messages, deserialize from JSON
			if grpcResp.Result.TypeUrl == "application/json" {
				if err := json.Unmarshal(grpcResp.Result.Value, response); err != nil {
					return fmt.Errorf("failed to deserialize response from JSON: %w", err)
				}
			} else {
				return fmt.Errorf("unsupported response type: %s", grpcResp.Result.TypeUrl)
			}
		}
	}

	return nil
}

// Close closes the client and connection pool
func (c *Client) Close(ctx context.Context) error {
	return c.pool.Close(ctx)
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

// isConnectionFailure determines if an error indicates a connection failure
// that requires closing the connection
func isConnectionFailure(err error) bool {
	if err == nil {
		return false
	}

	// Check if it's a gRPC status error
	if st, ok := status.FromError(err); ok {
		code := st.Code()
		// These codes typically indicate connection issues:
		// - Unavailable: service is unavailable (connection lost)
		// - DeadlineExceeded: timeout (could indicate connection issues)
		// - Unimplemented: method not found (not a connection issue, but we'll close it to be safe)
		if code == codes.Unavailable || code == codes.DeadlineExceeded {
			return true
		}
		// Internal errors with connection-related messages
		if code == codes.Internal {
			errMsg := strings.ToLower(err.Error())
			if strings.Contains(errMsg, "connection") ||
				strings.Contains(errMsg, "transport") ||
				strings.Contains(errMsg, "broken pipe") ||
				strings.Contains(errMsg, "connection reset") {
				return true
			}
		}
	}

	// Check error message for connection-related issues
	errMsg := strings.ToLower(err.Error())
	connectionErrorIndicators := []string{
		"connection refused",
		"connection reset",
		"broken pipe",
		"connection closed",
		"transport is closing",
		"connection lost",
		"no connection available",
		"context deadline exceeded", // Often indicates connection timeout
	}

	for _, indicator := range connectionErrorIndicators {
		if strings.Contains(errMsg, indicator) {
			return true
		}
	}

	return false
}
