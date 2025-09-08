package client

import (
	"context"
	"fmt"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

// ConnectionPool manages a pool of gRPC connections
type ConnectionPool struct {
	address string
	maxSize int
	timeout time.Duration
	logger  logging.Logger

	connections chan *grpc.ClientConn
	mu          sync.RWMutex
	closed      bool
}

// NewConnectionPool creates a new connection pool
func NewConnectionPool(maxSize int, timeout time.Duration, logger logging.Logger) *ConnectionPool {
	return &ConnectionPool{
		maxSize:     maxSize,
		timeout:     timeout,
		logger:      logger,
		connections: make(chan *grpc.ClientConn, maxSize),
	}
}

// GetConnection gets a connection from the pool
func (p *ConnectionPool) GetConnection(ctx context.Context, address string) (*grpc.ClientConn, error) {
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
		if err := conn.Close(); err != nil {
			p.logger.Errorf("Failed to close connection: %v", err)
		}
	default:
		// No connection available, create new one if under limit
		if len(p.connections) < p.maxSize {
			return p.createConnection(address)
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
func (p *ConnectionPool) ReturnConnection(conn *grpc.ClientConn, failed bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		if err := conn.Close(); err != nil {
			p.logger.Errorf("Failed to close connection: %v", err)
		}
		return
	}

	if failed {
		// Connection failed, close it
		if err := conn.Close(); err != nil {
			p.logger.Errorf("Failed to close connection: %v", err)
		}
		return
	}

	// Return connection to pool
	select {
	case p.connections <- conn:
		// Successfully returned to pool
	default:
		// Pool is full, close connection
		if err := conn.Close(); err != nil {
			p.logger.Errorf("Failed to close connection: %v", err)
		}
	}
}

// createConnection creates a new gRPC connection
func (p *ConnectionPool) createConnection(address string) (*grpc.ClientConn, error) {
	conn, err := grpc.NewClient(address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to dial gRPC server: %w", err)
	}

	p.logger.Debug("Created new gRPC connection", "address", address)
	return conn, nil
}

// isConnectionHealthy checks if a connection is healthy
func (p *ConnectionPool) isConnectionHealthy(conn *grpc.ClientConn) bool {
	// Simple health check - check connection state
	state := conn.GetState()
	return state.String() == "READY"
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
		if err := conn.Close(); err != nil {
			p.logger.Errorf("Failed to close connection: %v", err)
		}
	}

	return nil
}
