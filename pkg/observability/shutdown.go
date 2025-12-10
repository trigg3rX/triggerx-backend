package observability

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// ShutdownTimeout is the default timeout for graceful shutdown
var ShutdownTimeout = 30 * time.Second

// RegisterShutdownHooks registers signal handlers for graceful shutdown
// It returns a channel that will be closed when a shutdown signal is received
func RegisterShutdownHooks() <-chan os.Signal {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	return sigChan
}

// ShutdownWithTimeout shuts down the observability instance with a timeout
func ShutdownWithTimeout(obs *Observability, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	return obs.Shutdown(ctx)
}

// ShutdownOnSignal waits for a shutdown signal and then gracefully shuts down
// This is a convenience function that combines signal handling and shutdown
func ShutdownOnSignal(obs *Observability, timeout time.Duration) error {
	sigChan := RegisterShutdownHooks()
	<-sigChan

	return ShutdownWithTimeout(obs, timeout)
}
