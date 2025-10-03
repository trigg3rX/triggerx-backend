package connection

import (
	"context"
	"time"

	"github.com/trigg3rX/triggerx-backend/pkg/datastore/interfaces"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

// HealthChecker performs database health checks
type HealthChecker struct {
	connectionManager interfaces.Connection
	logger            logging.Logger
	interval          time.Duration
	stopChan          chan struct{}
}

// NewHealthChecker creates a new health checker
func NewHealthChecker(connectionManager interfaces.Connection, logger logging.Logger, interval time.Duration) *HealthChecker {
	return &HealthChecker{
		connectionManager: connectionManager,
		logger:            logger,
		interval:          interval,
		stopChan:          make(chan struct{}),
	}
}

// Start begins the health checking process
func (hc *HealthChecker) Start() {
	ticker := time.NewTicker(hc.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			if err := hc.connectionManager.HealthCheck(ctx); err != nil {
				hc.logger.Errorf("Health check failed: %v", err)
			}
			cancel()
		case <-hc.stopChan:
			hc.logger.Info("Health checker stopped")
			return
		}
	}
}

// Stop stops the health checking process
func (hc *HealthChecker) Stop() {
	close(hc.stopChan)
}
