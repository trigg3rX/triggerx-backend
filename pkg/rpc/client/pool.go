package client

import (
	"context"
	"fmt"
	"net/rpc"
	"sync"
	"time"

	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

// ConnectionPool manages a pool of RPC connections
type ConnectionPool struct {
	address     string
	maxSize     int
	timeout     time.Duration
	logger      logging.Logger

	connections chan *rpc.Client
	mu          sync.RWMutex
	closed      bool
}

// NewConnectionPool creates a new connection pool
func NewConnectionPool(maxSize int, timeout time.Duration, logger logging.Logger) *ConnectionPool {
	return &ConnectionPool{
		maxSize:     maxSize,
		timeout:     timeout,
		logger:      logger,
		connections: make(chan *rpc.Client, maxSize),
	}
}

// GetConnection gets a connection from the pool
func (p *ConnectionPool) GetConnection(ctx context.Context, address string) (*rpc.Client, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return nil, fmt.Errorf("connection pool is closed")
	}

	// Update address if needed
	if p.address == "" {
		p.address = address
	}

	// Try to get existing connection
	select {
	case conn := <-p.connections:
		// Test connection health
		if p.isConnectionHealthy(conn) {
			return conn, nil
		}
		// Connection is unhealthy, close it and create new one
		conn.Close()
	default:
		// No connection available, create new one if under limit
		if len(p.connections) < p.maxSize {
			return p.createConnection(ctx, address)
		}
	}

	// Wait for available connection
	select {
	case conn := <-p.connections:
		return conn, nil
	case <-ctx.Done():
		return nil, fmt.Errorf("timeout waiting for connection: %w", ctx.Err())
	}
}

// ReturnConnection returns a connection to the pool
func (p *ConnectionPool) ReturnConnection(conn *rpc.Client, failed bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		conn.Close()
		return
	}

	if failed {
		// Connection failed, close it
		conn.Close()
		return
	}

	// Return connection to pool
	select {
	case p.connections <- conn:
		// Successfully returned to pool
	default:
		// Pool is full, close connection
		conn.Close()
	}
}

// createConnection creates a new RPC connection
func (p *ConnectionPool) createConnection(ctx context.Context, address string) (*rpc.Client, error) {
	conn, err := rpc.DialHTTP("tcp", address)
	if err != nil {
		return nil, fmt.Errorf("failed to dial RPC server: %w", err)
	}

	p.logger.Debug("Created new RPC connection", "address", address)
	return conn, nil
}

// isConnectionHealthy checks if a connection is healthy
func (p *ConnectionPool) isConnectionHealthy(conn *rpc.Client) bool {
	// Simple health check - try to call a method that should always exist
	var response interface{}
	err := conn.Call("HealthCheck", &struct{}{}, &response)
	return err == nil
}

// Close closes the connection pool
func (p *ConnectionPool) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return nil
	}

	p.closed = true
	close(p.connections)

	// Close all connections
	for conn := range p.connections {
		conn.Close()
	}

	return nil
}