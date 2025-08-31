package docker

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/trigg3rX/triggerx-backend/pkg/docker/config"
	"github.com/trigg3rX/triggerx-backend/pkg/docker/execution"
	"github.com/trigg3rX/triggerx-backend/pkg/docker/types"
	httppkg "github.com/trigg3rX/triggerx-backend/pkg/http"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

// CodeExecutor defines what the DockerManager needs from a code executor
type CodeExecutor interface {
	Execute(ctx context.Context, fileURL string, fileLanguage string, noOfAttesters int) (*types.ExecutionResult, error)
	GetHealthStatus() *execution.HealthStatus
	GetStats() *types.PerformanceMetrics
	GetPoolStats() map[types.Language]*types.PoolStats
	InitializeLanguagePools(ctx context.Context, languages []types.Language) error
	GetSupportedLanguages() []types.Language
	IsLanguageSupported(language types.Language) bool
	GetActiveExecutions() []*types.ExecutionContext
	GetAlerts(severity string, limit int) []execution.Alert
	ClearAlerts()
	CancelExecution(executionID string) error
	Close() error
}

// DockerManager is the main entry point for the Docker package
// It provides a unified interface for all Docker operations
type DockerManager struct {
	executor    CodeExecutor
	config      config.ConfigProviderInterface
	logger      logging.Logger
	mutex       sync.RWMutex
	initialized bool
	closed      bool
}

// NewDockerManager creates a new Docker manager with the specified configuration
func NewDockerManager(
	executor CodeExecutor,
	cfg config.ConfigProviderInterface,
	logger logging.Logger,
) (*DockerManager, error) {
	if logger == nil {
		return nil, fmt.Errorf("logger cannot be nil")
	}
	if cfg == nil {
		return nil, fmt.Errorf("config provider cannot be nil")
	}
	if executor == nil {
		return nil, fmt.Errorf("executor cannot be nil")
	}

	dm := &DockerManager{
		executor: executor,
		config:   cfg,
		logger:   logger,
	}

	return dm, nil
}

// NewDockerManagerFromFile creates a new Docker manager from a configuration file
// This is a convenience function for backward compatibility
func NewDockerManagerFromFile(configFilePath string, logger logging.Logger) (*DockerManager, error) {
	if logger == nil {
		return nil, fmt.Errorf("logger cannot be nil")
	}

	configProvider, err := config.NewConfigProvider(configFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create config provider: %w", err)
	}

	// Create default implementations
	httpClient, err := httppkg.NewHTTPClient(httppkg.DefaultHTTPRetryConfig(), logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP client: %w", err)
	}
	executor, err := execution.NewCodeExecutor(context.Background(), configProvider, httpClient, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create default executor: %w", err)
	}

	return NewDockerManager(executor, configProvider, logger)
}

// Initialize sets up the Docker manager and all its components
func (dm *DockerManager) Initialize(ctx context.Context) error {
	dm.mutex.Lock()
	defer dm.mutex.Unlock()

	if dm.initialized {
		return fmt.Errorf("docker manager already initialized")
	}

	dm.logger.Info("Initializing Docker manager")

	// Initialize language-specific container pools
	supportedLanguages := dm.config.GetSupportedLanguages()
	if err := dm.executor.InitializeLanguagePools(ctx, supportedLanguages); err != nil {
		return fmt.Errorf("failed to initialize language pools: %w", err)
	}

	dm.initialized = true

	dm.logger.Infof("Docker manager initialized successfully with %d language pools", len(supportedLanguages))
	return nil
}

// Execute runs code from the specified URL with the given number of attestations
func (dm *DockerManager) Execute(ctx context.Context, fileURL string, fileLanguage string, noOfAttesters int) (*types.ExecutionResult, error) {
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

	result, err := dm.executor.Execute(ctx, fileURL, fileLanguage, noOfAttesters)
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

// GetPoolStats returns statistics for all language pools
func (dm *DockerManager) GetAllPoolStats() map[types.Language]*types.PoolStats {
	dm.mutex.RLock()
	defer dm.mutex.RUnlock()

	if !dm.initialized || dm.closed {
		return make(map[types.Language]*types.PoolStats)
	}

	return dm.executor.GetPoolStats()
}

// GetLanguageStats returns statistics for a specific language pool
func (dm *DockerManager) GetPoolStats(language types.Language) *types.PoolStats {
	dm.mutex.RLock()
	defer dm.mutex.RUnlock()

	if !dm.initialized || dm.closed {
		return nil
	}

	stats := dm.executor.GetPoolStats()
	if stats == nil {
		return nil
	}

	return stats[language]
}

// GetLanguageStats returns statistics for a specific language pool
func (dm *DockerManager) GetLanguageStats(language types.Language) (*types.PoolStats, bool) {
	dm.mutex.RLock()
	defer dm.mutex.RUnlock()

	if !dm.initialized || dm.closed {
		return nil, false
	}

	stats := dm.executor.GetPoolStats()
	if stats == nil {
		return nil, false
	}

	poolStats, exists := stats[language]
	return poolStats, exists
}

// GetSupportedLanguages returns all languages with active pools
func (dm *DockerManager) GetSupportedLanguages() []types.Language {
	dm.mutex.RLock()
	defer dm.mutex.RUnlock()

	if !dm.initialized || dm.closed {
		return []types.Language{}
	}

	return dm.executor.GetSupportedLanguages()
}

// IsLanguageSupported checks if a language is supported
func (dm *DockerManager) IsLanguageSupported(language types.Language) bool {
	dm.mutex.RLock()
	defer dm.mutex.RUnlock()

	if !dm.initialized || dm.closed {
		return false
	}

	return dm.executor.IsLanguageSupported(language)
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

// GetConfig returns the current configuration provider
func (dm *DockerManager) GetConfig() config.ConfigProviderInterface {
	dm.mutex.RLock()
	defer dm.mutex.RUnlock()
	return dm.config
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
