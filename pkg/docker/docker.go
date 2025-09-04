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
	Close(ctx context.Context) error
}

// DockerExecutor is the main entry point for the Docker package
// It provides a unified interface for all Docker operations
type DockerExecutor struct {
	executor    CodeExecutor
	config      config.ConfigProviderInterface
	logger      logging.Logger
	mutex       sync.RWMutex
	initialized bool
	closed      bool
}

// NewDockerManager creates a new Docker manager with the specified configuration
func NewDockerExecutor(
	executor CodeExecutor,
	cfg config.ConfigProviderInterface,
	logger logging.Logger,
) (*DockerExecutor, error) {
	if logger == nil {
		return nil, fmt.Errorf("logger cannot be nil")
	}
	if cfg == nil {
		return nil, fmt.Errorf("config provider cannot be nil")
	}
	if executor == nil {
		return nil, fmt.Errorf("executor cannot be nil")
	}

	dm := &DockerExecutor{
		executor: executor,
		config:   cfg,
		logger:   logger,
	}

	return dm, nil
}

// NewDockerManagerFromFile creates a new Docker manager from a configuration file
// This is a convenience function for backward compatibility
func NewDockerExecutorFromFile(configFilePath string, logger logging.Logger) (*DockerExecutor, error) {
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

	return NewDockerExecutor(executor, configProvider, logger)
}

// Initialize sets up the Docker manager and all its components
func (de *DockerExecutor) Initialize(ctx context.Context) error {
	de.mutex.Lock()
	defer de.mutex.Unlock()

	if de.initialized {
		return fmt.Errorf("docker manager already initialized")
	}

	de.logger.Info("Initializing Docker manager")

	// Initialize language-specific container pools
	supportedLanguages := de.config.GetSupportedLanguages()
	if err := de.executor.InitializeLanguagePools(ctx, supportedLanguages); err != nil {
		return fmt.Errorf("failed to initialize language pools: %w", err)
	}

	de.initialized = true

	de.logger.Infof("Docker manager initialized successfully with %d language pools", len(supportedLanguages))
	return nil
}

// Execute runs code from the specified URL with the given number of attestations
func (de *DockerExecutor) Execute(ctx context.Context, fileURL string, fileLanguage string, noOfAttesters int) (*types.ExecutionResult, error) {
	de.mutex.RLock()
	if !de.initialized {
		de.mutex.RUnlock()
		return nil, fmt.Errorf("docker manager not initialized")
	}
	if de.closed {
		de.mutex.RUnlock()
		return nil, fmt.Errorf("docker manager is closed")
	}
	de.mutex.RUnlock()

	de.logger.Infof("Executing code from URL: %s with %d attestations", fileURL, noOfAttesters)

	result, err := de.executor.Execute(ctx, fileURL, fileLanguage, noOfAttesters)
	if err != nil {
		de.logger.Errorf("Execution failed: %v", err)
		return nil, fmt.Errorf("execution failed: %w", err)
	}

	de.logger.Infof("Execution completed successfully")
	return result, nil
}

// GetHealthStatus returns the current health status of the Docker manager
func (de *DockerExecutor) GetHealthStatus() *execution.HealthStatus {
	de.mutex.RLock()
	defer de.mutex.RUnlock()

	if !de.initialized || de.closed {
		return &execution.HealthStatus{
			Status:    "unavailable",
			Score:     0.0,
			LastCheck: time.Now(),
			Alerts:    []execution.Alert{},
			Metrics:   &types.PerformanceMetrics{},
		}
	}

	return de.executor.GetHealthStatus()
}

// GetStats returns performance metrics for the Docker manager
func (de *DockerExecutor) GetStats() *types.PerformanceMetrics {
	de.mutex.RLock()
	defer de.mutex.RUnlock()

	if !de.initialized || de.closed {
		return &types.PerformanceMetrics{}
	}

	return de.executor.GetStats()
}

// GetPoolStats returns statistics for all language pools
func (de *DockerExecutor) GetAllPoolStats() map[types.Language]*types.PoolStats {
	de.mutex.RLock()
	defer de.mutex.RUnlock()

	if !de.initialized || de.closed {
		return make(map[types.Language]*types.PoolStats)
	}

	return de.executor.GetPoolStats()
}

// GetLanguageStats returns statistics for a specific language pool
func (de *DockerExecutor) GetPoolStats(language types.Language) *types.PoolStats {
	de.mutex.RLock()
	defer de.mutex.RUnlock()

	if !de.initialized || de.closed {
		return nil
	}

	stats := de.executor.GetPoolStats()
	if stats == nil {
		return nil
	}

	return stats[language]
}

// GetLanguageStats returns statistics for a specific language pool
func (de *DockerExecutor) GetLanguageStats(language types.Language) (*types.PoolStats, bool) {
	de.mutex.RLock()
	defer de.mutex.RUnlock()

	if !de.initialized || de.closed {
		return nil, false
	}

	stats := de.executor.GetPoolStats()
	if stats == nil {
		return nil, false
	}

	poolStats, exists := stats[language]
	return poolStats, exists
}

// GetSupportedLanguages returns all languages with active pools
func (de *DockerExecutor) GetSupportedLanguages() []types.Language {
	de.mutex.RLock()
	defer de.mutex.RUnlock()

	if !de.initialized || de.closed {
		return []types.Language{}
	}

	return de.executor.GetSupportedLanguages()
}

// IsLanguageSupported checks if a language is supported
func (de *DockerExecutor) IsLanguageSupported(language types.Language) bool {
	de.mutex.RLock()
	defer de.mutex.RUnlock()

	if !de.initialized || de.closed {
		return false
	}

	return de.executor.IsLanguageSupported(language)
}

// GetActiveExecutions returns all currently active executions
func (de *DockerExecutor) GetActiveExecutions() []*types.ExecutionContext {
	de.mutex.RLock()
	defer de.mutex.RUnlock()

	if !de.initialized || de.closed {
		return []*types.ExecutionContext{}
	}

	return de.executor.GetActiveExecutions()
}

// CancelExecution cancels a running execution
func (de *DockerExecutor) CancelExecution(executionID string) error {
	de.mutex.RLock()
	defer de.mutex.RUnlock()

	if !de.initialized {
		return fmt.Errorf("docker manager not initialized")
	}
	if de.closed {
		return fmt.Errorf("docker manager is closed")
	}

	return de.executor.CancelExecution(executionID)
}

// GetAlerts returns alerts from the monitoring system
func (de *DockerExecutor) GetAlerts(severity string, limit int) []execution.Alert {
	de.mutex.RLock()
	defer de.mutex.RUnlock()

	if !de.initialized || de.closed {
		return []execution.Alert{}
	}

	return de.executor.GetAlerts(severity, limit)
}

// ClearAlerts clears all alerts from the monitoring system
func (de *DockerExecutor) ClearAlerts() {
	de.mutex.RLock()
	defer de.mutex.RUnlock()

	if !de.initialized || de.closed {
		return
	}

	de.executor.ClearAlerts()
}

// GetConfig returns the current configuration provider
func (de *DockerExecutor) GetConfig() config.ConfigProviderInterface {
	de.mutex.RLock()
	defer de.mutex.RUnlock()
	return de.config
}

// Close shuts down the Docker manager and cleans up resources
func (de *DockerExecutor) Close(ctx context.Context) error {
	de.mutex.Lock()
	defer de.mutex.Unlock()

	if de.closed {
		return nil
	}

	de.logger.Info("Closing Docker manager")

	if de.executor != nil {
		if err := de.executor.Close(ctx); err != nil {
			de.logger.Warnf("Failed to close executor: %v", err)
		}
	}

	de.closed = true
	de.logger.Info("Docker manager closed")
	return nil
}
