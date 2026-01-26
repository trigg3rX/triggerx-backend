package client

import (
	"context"
	"fmt"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
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
		// Wait for connection to be ready if needed (outside lock to avoid deadlock)
		readyConn, err := p.waitForConnectionReady(ctx, conn, address)
		if err != nil {
			// Connection failed, recursively try again
			return p.GetConnection(ctx, address)
		}
		return readyConn, nil
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
		// Wait for connection to be ready if needed
		readyConn, err := p.waitForConnectionReady(ctx, conn, address)
		if err != nil {
			// Connection failed, recursively try again
			return p.GetConnection(ctx, address)
		}
		return readyConn, nil
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

	// Check connection health before returning to pool
	// If connection is not healthy, close it instead of returning to pool
	if !p.isConnectionHealthy(conn) {
		p.logger.Warn(ctx, "Connection is unhealthy, closing instead of returning to pool",
			observability.String("state", conn.GetState().String()))
		if err := conn.Close(); err != nil {
			p.logger.Error(ctx, "Failed to close unhealthy connection", observability.Error(err))
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
// Note: grpc.NewClient creates a lazy connection that doesn't connect until first use
// We return it immediately and let gRPC handle connection asynchronously
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

// waitForConnectionReady waits for a connection to become READY
// Returns the connection if ready, or error if it fails or times out
func (p *ConnectionPool) waitForConnectionReady(ctx context.Context, conn *grpc.ClientConn, address string) (*grpc.ClientConn, error) {
	state := conn.GetState()
	
	// If already READY, return immediately
	if state == connectivity.Ready {
		return conn, nil
	}
	
	// If in SHUTDOWN or TRANSIENT_FAILURE, close and return error
	if state == connectivity.Shutdown || state == connectivity.TransientFailure {
		if err := conn.Close(); err != nil {
			p.logger.Error(ctx, "Failed to close failed connection", observability.Error(err))
		}
		return nil, fmt.Errorf("connection is in failed state: %v", state)
	}
	
	// For IDLE or CONNECTING, wait for it to become READY
	// Use a reasonable timeout (5 seconds) for connection establishment
	connectCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	
	// Wait for state change to READY
	for {
		currentState := conn.GetState()
		if currentState == connectivity.Ready {
			return conn, nil
		}
		
		// If connection failed, close it and return error
		if currentState == connectivity.Shutdown || currentState == connectivity.TransientFailure {
			if err := conn.Close(); err != nil {
				p.logger.Error(ctx, "Failed to close failed connection", observability.Error(err))
			}
			return nil, fmt.Errorf("connection failed to establish: state=%v", currentState)
		}
		
		// Wait for state change
		if !conn.WaitForStateChange(connectCtx, currentState) {
			// Context expired - connection might still be connecting
			// Close it and let the caller create a new one
			if err := conn.Close(); err != nil {
				p.logger.Error(ctx, "Failed to close connection", observability.Error(err))
			}
			return nil, fmt.Errorf("timeout waiting for connection to be ready: %w", connectCtx.Err())
		}
	}
}

// isConnectionHealthy checks if a connection is healthy
func (p *ConnectionPool) isConnectionHealthy(conn *grpc.ClientConn) bool {
	state := conn.GetState()
	
	// Only READY state is considered healthy
	// Other states indicate the connection is not usable:
	// - IDLE: initial state, not connected
	// - CONNECTING: attempting to connect
	// - READY: connected and ready for RPCs
	// - TRANSIENT_FAILURE: connection lost, will attempt to reconnect
	// - SHUTDOWN: connection is closed
	return state == connectivity.Ready
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
