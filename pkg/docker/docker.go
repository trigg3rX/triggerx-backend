package docker

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/trigg3rX/triggerx-backend/pkg/docker/config"
	"github.com/trigg3rX/triggerx-backend/pkg/docker/execution"
	"github.com/trigg3rX/triggerx-backend/pkg/docker/types"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

// DockerManager is the main entry point for the Docker package
// It provides a unified interface for all Docker operations
type DockerManager struct {
	executor    *execution.CodeExecutor
	config      config.ExecutorConfig
	logger      logging.Logger
	mutex       sync.RWMutex
	initialized bool
	closed      bool
}

// NewDockerManager creates a new Docker manager with the specified configuration
func NewDockerManager(cfg config.ExecutorConfig, logger logging.Logger) (*DockerManager, error) {
	if logger == nil {
		return nil, fmt.Errorf("logger cannot be nil")
	}

	dm := &DockerManager{
		config: cfg,
		logger: logger,
	}

	return dm, nil
}

// Initialize sets up the Docker manager and all its components
func (dm *DockerManager) Initialize(ctx context.Context) error {
	dm.mutex.Lock()
	defer dm.mutex.Unlock()

	if dm.initialized {
		return fmt.Errorf("docker manager already initialized")
	}

	dm.logger.Info("Initializing Docker manager")

	// Create the code executor
	executor, err := execution.NewCodeExecutor(ctx, dm.config, dm.logger)
	if err != nil {
		return fmt.Errorf("failed to create code executor: %w", err)
	}

	dm.executor = executor
	dm.initialized = true

	dm.logger.Info("Docker manager initialized successfully")
	return nil
}

// Execute runs code from the specified URL with the given number of attestations
func (dm *DockerManager) Execute(ctx context.Context, fileURL string, noOfAttesters int) (*types.ExecutionResult, error) {
	dm.mutex.RLock()
	if !dm.initialized {
		dm.mutex.RUnlock()
		return nil, fmt.Errorf("docker manager not initialized")
	}
	if dm.closed {
		dm.mutex.RUnlock()
		return nil, fmt.Errorf("docker manager is closed")
	}
	dm.mutex.RUnlock()

	dm.logger.Infof("Executing code from URL: %s with %d attestations", fileURL, noOfAttesters)

	result, err := dm.executor.Execute(ctx, fileURL, noOfAttesters)
	if err != nil {
		dm.logger.Errorf("Execution failed: %v", err)
		return nil, fmt.Errorf("execution failed: %w", err)
	}

	dm.logger.Infof("Execution completed successfully")
	return result, nil
}

// GetHealthStatus returns the current health status of the Docker manager
func (dm *DockerManager) GetHealthStatus() *execution.HealthStatus {
	dm.mutex.RLock()
	defer dm.mutex.RUnlock()

	if !dm.initialized || dm.closed {
		return &execution.HealthStatus{
			Status:    "unavailable",
			Score:     0.0,
			LastCheck: time.Now(),
			Alerts:    []execution.Alert{},
			Metrics:   &types.PerformanceMetrics{},
		}
	}

	return dm.executor.GetHealthStatus()
}

// GetStats returns performance metrics for the Docker manager
func (dm *DockerManager) GetStats() *types.PerformanceMetrics {
	dm.mutex.RLock()
	defer dm.mutex.RUnlock()

	if !dm.initialized || dm.closed {
		return &types.PerformanceMetrics{}
	}

	return dm.executor.GetStats()
}

// GetPoolStats returns container pool statistics
func (dm *DockerManager) GetPoolStats() *types.PoolStats {
	dm.mutex.RLock()
	defer dm.mutex.RUnlock()

	if !dm.initialized || dm.closed {
		return &types.PoolStats{}
	}

	return dm.executor.GetPoolStats()
}

// GetActiveExecutions returns all currently active executions
func (dm *DockerManager) GetActiveExecutions() []*types.ExecutionContext {
	dm.mutex.RLock()
	defer dm.mutex.RUnlock()

	if !dm.initialized || dm.closed {
		return []*types.ExecutionContext{}
	}

	return dm.executor.GetActiveExecutions()
}

// GetExecutionByID returns a specific execution by its ID
func (dm *DockerManager) GetExecutionByID(executionID string) (*types.ExecutionContext, bool) {
	dm.mutex.RLock()
	defer dm.mutex.RUnlock()

	if !dm.initialized || dm.closed {
		return nil, false
	}

	return dm.executor.GetExecutionByID(executionID)
}

// CancelExecution cancels a running execution
func (dm *DockerManager) CancelExecution(executionID string) error {
	dm.mutex.RLock()
	defer dm.mutex.RUnlock()

	if !dm.initialized {
		return fmt.Errorf("docker manager not initialized")
	}
	if dm.closed {
		return fmt.Errorf("docker manager is closed")
	}

	return dm.executor.CancelExecution(executionID)
}

// GetAlerts returns alerts from the monitoring system
func (dm *DockerManager) GetAlerts(severity string, limit int) []execution.Alert {
	dm.mutex.RLock()
	defer dm.mutex.RUnlock()

	if !dm.initialized || dm.closed {
		return []execution.Alert{}
	}

	return dm.executor.GetAlerts(severity, limit)
}

// ClearAlerts clears all alerts from the monitoring system
func (dm *DockerManager) ClearAlerts() {
	dm.mutex.RLock()
	defer dm.mutex.RUnlock()

	if !dm.initialized || dm.closed {
		return
	}

	dm.executor.ClearAlerts()
}

// IsInitialized returns whether the Docker manager is initialized
func (dm *DockerManager) IsInitialized() bool {
	dm.mutex.RLock()
	defer dm.mutex.RUnlock()
	return dm.initialized
}

// IsClosed returns whether the Docker manager is closed
func (dm *DockerManager) IsClosed() bool {
	dm.mutex.RLock()
	defer dm.mutex.RUnlock()
	return dm.closed
}

// GetConfig returns a copy of the current configuration
func (dm *DockerManager) GetConfig() config.ExecutorConfig {
	dm.mutex.RLock()
	defer dm.mutex.RUnlock()
	return dm.config
}

// UpdateConfig updates the configuration (requires reinitialization)
func (dm *DockerManager) UpdateConfig(newConfig config.ExecutorConfig) error {
	dm.mutex.Lock()
	defer dm.mutex.Unlock()

	if dm.initialized {
		return fmt.Errorf("cannot update config while initialized, close manager first")
	}

	dm.config = newConfig
	dm.logger.Info("Configuration updated")
	return nil
}

// Close shuts down the Docker manager and cleans up resources
func (dm *DockerManager) Close() error {
	dm.mutex.Lock()
	defer dm.mutex.Unlock()

	if dm.closed {
		return nil
	}

	dm.logger.Info("Closing Docker manager")

	if dm.executor != nil {
		if err := dm.executor.Close(); err != nil {
			dm.logger.Warnf("Failed to close executor: %v", err)
		}
	}

	dm.closed = true
	dm.logger.Info("Docker manager closed")
	return nil
}
