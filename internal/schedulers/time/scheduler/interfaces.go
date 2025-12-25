package scheduler

import (
	"context"
)

// TaskDispatcherClient defines the interface for task dispatcher RPC client operations.
// This interface allows for dependency injection and easier testing.
type TaskDispatcherClient interface {
	// Call makes an RPC call to the task dispatcher service.
	// method is the RPC method name (e.g., "submit-task").
	// request is the request payload.
	// response is where the response will be unmarshaled.
	Call(ctx context.Context, method string, request interface{}, response interface{}) error

	// Close closes the client connection pool and releases resources.
	Close(ctx context.Context) error
}

// Scheduler defines the public interface for the time-based scheduler.
// This allows for easier testing and potential alternative implementations.
type Scheduler interface {
	// Start begins the scheduler's main polling and execution loop.
	Start(ctx context.Context)

	// Stop gracefully stops the scheduler and releases resources.
	Stop(ctx context.Context)

	// GetStats returns current scheduler statistics and metrics.
	GetStats() map[string]interface{}
}
