package dockerexecutor

//go:generate mockgen -source=interface.go -destination=mock_docker_executor.go -package=dockerexecutor . DockerExecutorAPI

import (
	"context"

	"github.com/trigg3rX/triggerx-backend/pkg/dockerexecutor/config"
	"github.com/trigg3rX/triggerx-backend/pkg/dockerexecutor/execution"
	"github.com/trigg3rX/triggerx-backend/pkg/dockerexecutor/types"
)

type DockerExecutorAPI interface {
	Initialize(ctx context.Context) error
	Execute(ctx context.Context, fileURL string, fileLanguage string, noOfAttesters int) (*types.ExecutionResult, error)
	ExecuteSource(ctx context.Context, code string, language string) (*types.ExecutionResult, error)
	GetHealthStatus() *execution.HealthStatus
	GetExecutionFeeConfig() config.ExecutionFeeConfig
	GetStats() *types.PerformanceMetrics
	GetAllPoolStats() map[types.Language]*types.PoolStats
	GetPoolStats(language types.Language) *types.PoolStats
	GetLanguageStats(language types.Language) (*types.PoolStats, bool)
	GetSupportedLanguages() []types.Language
	IsLanguageSupported(language types.Language) bool
	GetActiveExecutions() []*types.ExecutionContext
	CancelExecution(executionID string) error
	GetAlerts(severity string, limit int) []execution.Alert
	ClearAlerts()
	Close(ctx context.Context) error
}
