package client

import (
	"context"
	"fmt"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/trigg3rX/triggerx-backend/pkg/observability"
	"github.com/trigg3rX/triggerx-backend/pkg/rpc/tracing"
)

// ConnectionPool manages a pool of gRPC connections
type ConnectionPool struct {
	address     string
	maxSize     int
	timeout     time.Duration
	logger      observability.Logger
	tracer      observability.Tracer
	serviceName string
	connections chan *grpc.ClientConn
	mu          sync.RWMutex
	closed      bool
}

// NewConnectionPool creates a new connection pool
func NewConnectionPool(maxSize int, timeout time.Duration, logger observability.Logger, tracer observability.Tracer, serviceName string) *ConnectionPool {
	return &ConnectionPool{
		maxSize:     maxSize,
		timeout:     timeout,
		logger:      logger,
		tracer:      tracer,
		serviceName: serviceName,
		connections: make(chan *grpc.ClientConn, maxSize),
	}
}

// GetConnection gets a connection from the pool
func (p *ConnectionPool) GetConnection(ctx context.Context, address string) (*grpc.ClientConn, error) {
	p.mu.Lock()

	if p.closed {
		p.mu.Unlock()
		return nil, fmt.Errorf("connection pool is closed")
	}

	// Update address if needed
	if p.address == "" {
		p.address = address
	}

	// Try to get existing connection
	select {
	case conn := <-p.connections:
		p.mu.Unlock()
		// Test connection health (outside lock to avoid deadlock)
		if p.isConnectionHealthy(conn) {
			return conn, nil
		}
		// Connection is unhealthy, close it and create new one
		if err := conn.Close(); err != nil {
			p.logger.Error(ctx, "Failed to close unhealthy connection", observability.Error(err))
		}
		// Recursively try again (will create new connection if pool not full)
		return p.GetConnection(ctx, address)
	default:
		// No connection available, create new one if under limit
		poolSize := len(p.connections)
		if poolSize < p.maxSize {
			p.mu.Unlock()
			return p.createConnection(ctx, address)
		}
		// Pool is full, unlock and wait
		p.mu.Unlock()
	}

	// Wait for available connection with pool timeout (outside lock)
	poolCtx, cancel := context.WithTimeout(ctx, p.timeout)
	defer cancel()

	select {
	case conn := <-p.connections:
		// Verify connection is healthy before returning
		if p.isConnectionHealthy(conn) {
			return conn, nil
		}
		// Connection is unhealthy, close it and try again
		if err := conn.Close(); err != nil {
			p.logger.Error(ctx, "Failed to close unhealthy connection", observability.Error(err))
		}
		// Recursively try again
		return p.GetConnection(ctx, address)
	case <-poolCtx.Done():
		if poolCtx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("timeout waiting for connection from pool (timeout: %v)", p.timeout)
		}
		return nil, fmt.Errorf("timeout waiting for connection: %w", poolCtx.Err())
	}
}

// ReturnConnection returns a connection to the pool
func (p *ConnectionPool) ReturnConnection(ctx context.Context, conn *grpc.ClientConn, failed bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		if err := conn.Close(); err != nil {
			p.logger.Error(ctx, "Failed to close connection", observability.Error(err))
		}
		return
	}

	if failed {
		// Connection failed, close it
		if err := conn.Close(); err != nil {
			p.logger.Error(ctx, "Failed to close connection", observability.Error(err))
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
			p.logger.Error(ctx, "Failed to close connection", observability.Error(err))
		}
	}
}

// createConnection creates a new gRPC connection
func (p *ConnectionPool) createConnection(ctx context.Context, address string) (*grpc.ClientConn, error) {
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	// Add trace interceptor if tracer is available
	if p.tracer != nil && p.serviceName != "" {
		opts = append(opts, grpc.WithUnaryInterceptor(
			tracing.TraceClientInterceptor(p.tracer, p.serviceName),
		))
	}

	conn, err := grpc.NewClient(address, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to dial gRPC server: %w", err)
	}

	p.logger.Debug(ctx, "Created new gRPC connection", observability.String("address", address))
	return conn, nil
}

// isConnectionHealthy checks if a connection is healthy
func (p *ConnectionPool) isConnectionHealthy(conn *grpc.ClientConn) bool {
	// Simple health check - check connection state
	state := conn.GetState()
	return state.String() == "READY"
}

// Close closes the connection pool
func (p *ConnectionPool) Close(ctx context.Context) error {
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
			p.logger.Error(ctx, "Failed to close connection", observability.Error(err))
		}
	}

	return nil
}
