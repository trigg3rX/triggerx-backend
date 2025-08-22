package execution

import (
	"context"
	"fmt"

	"github.com/docker/docker/client"
	"github.com/trigg3rX/triggerx-backend/pkg/docker/config"
	"github.com/trigg3rX/triggerx-backend/pkg/docker/container"
	"github.com/trigg3rX/triggerx-backend/pkg/docker/file"
	"github.com/trigg3rX/triggerx-backend/pkg/docker/types"
	fs "github.com/trigg3rX/triggerx-backend/pkg/filesystem"
	httppkg "github.com/trigg3rX/triggerx-backend/pkg/http"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

type CodeExecutor struct {
	pipeline *ExecutionPipeline
	monitor  *ExecutionMonitor
	config   config.ExecutorConfig
	logger   logging.Logger
}

func NewCodeExecutor(ctx context.Context, cfg config.ExecutorConfig, httpClient *httppkg.HTTPClient, logger logging.Logger) (*CodeExecutor, error) {
	// Create Docker client with API version compatibility
	cli, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithVersion("1.44"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create docker client: %w", err)
	}

	// Create file manager
	fileMgr, err := file.NewFileManager(cfg, httpClient, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create file manager: %w", err)
	}

	// Create container manager
	containerMgr, err := container.NewContainerManager(cli, &fs.OSFileSystem{}, cfg, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create container manager: %w", err)
	}

	// Initialize container manager
	if err := containerMgr.Initialize(ctx); err != nil {
		return nil, fmt.Errorf("failed to initialize container manager: %w", err)
	}

	// Create execution pipeline
	pipeline := NewExecutionPipeline(cfg, fileMgr, containerMgr, logger)

	// Create execution monitor
	monitor := NewExecutionMonitor(pipeline, cfg, logger)

	return &CodeExecutor{
		pipeline: pipeline,
		monitor:  monitor,
		config:   cfg,
		logger:   logger,
	}, nil
}

func (e *CodeExecutor) Execute(ctx context.Context, fileURL string, fileLanguage string, noOfAttesters int) (*types.ExecutionResult, error) {
	e.logger.Infof("Executing code from URL: %s with %d attestations", fileURL, noOfAttesters)

	// Execute through pipeline
	result, err := e.pipeline.Execute(ctx, fileURL, fileLanguage, noOfAttesters)
	if err != nil {
		e.logger.Errorf("Execution failed: %v", err)
		return nil, err
	}

	e.logger.Infof("Execution completed successfully")
	return result, nil
}

func (e *CodeExecutor) GetHealthStatus() *HealthStatus {
	return e.monitor.GetHealthStatus()
}

func (e *CodeExecutor) GetStats() *types.PerformanceMetrics {
	return e.pipeline.GetStats()
}

func (e *CodeExecutor) GetPoolStats() map[types.Language]*types.PoolStats {
	return e.pipeline.containerMgr.GetPoolStats()
}

// InitializeLanguagePools initializes language-specific container pools
func (e *CodeExecutor) InitializeLanguagePools(ctx context.Context, languages []types.Language) error {
	return e.pipeline.containerMgr.InitializeLanguagePools(ctx, languages)
}

// GetSupportedLanguages returns all languages with active pools
func (e *CodeExecutor) GetSupportedLanguages() []types.Language {
	return e.pipeline.containerMgr.GetSupportedLanguages()
}

// IsLanguageSupported checks if a language is supported
func (e *CodeExecutor) IsLanguageSupported(language types.Language) bool {
	return e.pipeline.containerMgr.IsLanguageSupported(language)
}

func (e *CodeExecutor) GetActiveExecutions() []*types.ExecutionContext {
	return e.monitor.GetActiveExecutions()
}

func (e *CodeExecutor) GetExecutionByID(executionID string) (*types.ExecutionContext, bool) {
	return e.monitor.GetExecutionByID(executionID)
}

func (e *CodeExecutor) CancelExecution(executionID string) error {
	return e.monitor.CancelExecution(executionID)
}

func (e *CodeExecutor) GetAlerts(severity string, limit int) []Alert {
	return e.monitor.GetAlerts(severity, limit)
}

func (e *CodeExecutor) ClearAlerts() {
	e.monitor.ClearAlerts()
}

func (e *CodeExecutor) Close() error {
	e.logger.Info("Closing code executor")

	// Close monitor
	if e.monitor != nil {
		if err := e.monitor.Close(); err != nil {
			e.logger.Warnf("Failed to close monitor: %v", err)
		}
	}

	if e.pipeline.fileManager != nil {
		if err := e.pipeline.fileManager.Close(); err != nil {
			e.logger.Warnf("Failed to close file manager: %v", err)
		}
	}

	if e.pipeline.containerMgr != nil {
		if err := e.pipeline.containerMgr.Close(); err != nil {
			e.logger.Warnf("Failed to close container manager: %v", err)
		}
	}

	e.logger.Info("Code executor closed")
	return nil
}
