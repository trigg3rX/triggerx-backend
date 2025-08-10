package client

import (
	"context"
	"fmt"
	"net/rpc"
	"time"

	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/retry"
	rpcpkg "github.com/trigg3rX/triggerx-backend/pkg/rpc"
)

// Client represents a unified RPC client
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

// NewClient creates a new RPC client
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

// Call makes an RPC call to the specified service and method
func (c *Client) Call(ctx context.Context, method string, request interface{}, response interface{}) error {
	retryCfg := retry.DefaultRetryConfig()
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
	retryCfg.LogRetryAttempt = true
	// Retry on network/rpc errors; if the method returns a business error, don't retry
	retryCfg.ShouldRetry = func(err error) bool {
		if err == nil {
			return false
		}
		// Basic heuristic: retry on rpc errors and timeouts
		return true
	}

	_, err := retry.Retry(ctx, func() (struct{}, error) {
		// Get service info from registry or use direct connection
		var serviceInfo *rpcpkg.ServiceInfo
		var err error

		if c.registry != nil {
			serviceInfo, err = c.getServiceFromRegistry(ctx)
			if err != nil {
				return struct{}{}, fmt.Errorf("failed to get service from registry: %w", err)
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
			return struct{}{}, fmt.Errorf("failed to get connection: %w", err)
		}

		// Make RPC call
		callErr := c.makeRPCCall(ctx, conn, method, request, response)
		// Return connection to pool (mark failed if callErr != nil)
		c.pool.ReturnConnection(conn, callErr != nil)

		return struct{}{}, callErr
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

// makeRPCCall makes the actual RPC call
func (c *Client) makeRPCCall(ctx context.Context, conn *rpc.Client, method string, request interface{}, response interface{}) error {
	// Create call context with timeout
	callCtx, cancel := context.WithTimeout(ctx, c.config.Timeout)
	defer cancel()

	// Make the call
	call := conn.Go(method, request, response, nil)

	select {
	case <-callCtx.Done():
		return fmt.Errorf("RPC call timeout: %w", callCtx.Err())
	case <-call.Done:
		return call.Error
	}
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
