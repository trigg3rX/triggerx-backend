package execution

import (
	"context"
	"fmt"
	"time"

	"github.com/docker/docker/client"
	"github.com/trigg3rX/triggerx-backend/pkg/dockerexecutor/config"
	"github.com/trigg3rX/triggerx-backend/pkg/dockerexecutor/container"
	"github.com/trigg3rX/triggerx-backend/pkg/dockerexecutor/file"
	"github.com/trigg3rX/triggerx-backend/pkg/dockerexecutor/types"
	fs "github.com/trigg3rX/triggerx-backend/pkg/filesystem"
	httppkg "github.com/trigg3rX/triggerx-backend/pkg/http"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
)

// Events during code execution pipeline
type ExecutionEvent struct {
	Type     string // "started", "completed", "failed", "timeout"
	TraceID  string
	Stage    types.ExecutionStage
	Duration time.Duration
	Error    error
	Metadata map[string]interface{}
}

type codeExecutor struct {
	pipeline *executionPipeline
	monitor  *executionMonitor
	config   config.ConfigProviderInterface
	logger   observability.Logger
}

func NewCodeExecutor(ctx context.Context, cfg config.ConfigProviderInterface, httpClient *httppkg.HTTPClient, logger observability.Logger) (*codeExecutor, error) {
	// Create Docker client with API version compatibility
	cli, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithVersion("1.44"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create docker client: %w", err)
	}

	// Create file manager
	fileMgr, err := file.NewFileManager(ctx, cfg, httpClient, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create file manager: %w", err)
	}

	// Create container manager
	containerMgr, err := container.NewContainerManager(cli, &fs.OSFileSystem{}, cfg, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create container manager: %w", err)
	}

	// Create adapters for the pipeline
	fileManagerAdapter := NewFileManagerAdapter(fileMgr)
	containerManagerAdapter := NewContainerManagerAdapter(containerMgr)

	// Create execution pipeline
	pipeline := newExecutionPipeline(cfg, fileManagerAdapter, containerManagerAdapter, logger)

	// Create execution monitor
	monitor := newExecutionMonitor(ctx, pipeline, cfg, logger)

	return &codeExecutor{
		pipeline: pipeline,
		monitor:  monitor,
		config:   cfg,
		logger:   logger,
	}, nil
}

func (e *codeExecutor) Execute(ctx context.Context, fileURL string, fileLanguage string, noOfAttesters int, alchemyAPIKey string, metadata ...map[string]string) (*types.ExecutionResult, error) {
	e.logger.Info(ctx, "Executing code from URL", observability.String("fileURL", fileURL), observability.Int("noOfAttesters", noOfAttesters))

	// Extract metadata if provided
	var metadataMap map[string]string
	if len(metadata) > 0 {
		metadataMap = metadata[0]
	}

	// Execute through pipeline
	result, err := e.pipeline.execute(ctx, fileURL, fileLanguage, noOfAttesters, alchemyAPIKey, metadataMap)
	if err != nil {
		e.logger.Error(ctx, "Execution failed", observability.Error(err))
		return nil, err
	}

	e.logger.Info(ctx, "Execution completed successfully")
	return result, nil
}

// ExecuteSource executes raw source code by writing it to a temp file internally
func (e *codeExecutor) ExecuteSource(ctx context.Context, code string, language string, alchemyAPIKey string, metadata ...map[string]string) (*types.ExecutionResult, error) {
	e.logger.Info(ctx, "Executing raw source for language", observability.String("language", language))

	// Extract metadata if provided
	var metadataMap map[string]string
	if len(metadata) > 0 {
		metadataMap = metadata[0]
	}

	result, err := e.pipeline.executeSource(ctx, code, language, alchemyAPIKey, metadataMap)
	if err != nil {
		e.logger.Error(ctx, "Execution (raw) failed", observability.Error(err))
		return nil, err
	}
	e.logger.Info(ctx, "Execution (raw) completed successfully")
	return result, nil
}

func (e *codeExecutor) GetHealthStatus() *HealthStatus {
	return e.monitor.getHealthStatus()
}

func (e *codeExecutor) GetStats() *types.PerformanceMetrics {
	return e.pipeline.getStats()
}

func (e *codeExecutor) GetPoolStats(ctx context.Context) map[types.Language]*types.PoolStats {
	return e.pipeline.containerMgr.GetPoolStats()
}

// InitializeLanguagePools initializes language-specific container pools
func (e *codeExecutor) InitializeLanguagePools(ctx context.Context, languages []types.Language) error {
	return e.pipeline.containerMgr.InitializeLanguagePools(ctx, languages)
}

// GetSupportedLanguages returns all languages with active pools
func (e *codeExecutor) GetSupportedLanguages(ctx context.Context) []types.Language {
	return e.pipeline.containerMgr.GetSupportedLanguages()
}

// IsLanguageSupported checks if a language is supported
func (e *codeExecutor) IsLanguageSupported(language types.Language) bool {
	return e.pipeline.containerMgr.IsLanguageSupported(language)
}

func (e *codeExecutor) GetActiveExecutions() []*types.ExecutionContext {
	return e.monitor.getActiveExecutions()
}

func (e *codeExecutor) CancelExecution(ctx context.Context, executionID string) error {
	return e.monitor.cancelExecution(ctx, executionID)
}

func (e *codeExecutor) GetAlerts(severity string, limit int) []Alert {
	return e.monitor.getAlerts(severity, limit)
}

func (e *codeExecutor) ClearAlerts(ctx context.Context) {
	e.monitor.clearAlerts(ctx)
}

func (e *codeExecutor) Close(ctx context.Context) error {
	e.logger.Info(ctx, "Closing code executor")

	// Close pipeline first to ensure all active executions complete
	if e.pipeline != nil {
		if err := e.pipeline.close(ctx); err != nil {
			e.logger.Warn(ctx, "Failed to close pipeline", observability.Error(err))
		}
	}

	// Close monitor
	if e.monitor != nil {
		if err := e.monitor.close(ctx); err != nil {
			e.logger.Warn(ctx, "Failed to close monitor", observability.Error(err))
		}
	}

	if e.pipeline.fileManager != nil {
		if err := e.pipeline.fileManager.Close(); err != nil {
			e.logger.Warn(ctx, "Failed to close file manager", observability.Error(err))
		}
	}

	if e.pipeline.containerMgr != nil {
		if err := e.pipeline.containerMgr.Close(ctx); err != nil {
			e.logger.Warn(ctx, "Failed to close container manager", observability.Error(err))
		}
	}

	e.logger.Info(ctx, "Code executor closed")
	return nil
}
