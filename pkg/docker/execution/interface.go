package execution

//go:generate mockgen -source=interface.go -destination=mock_execution.go -package=execution

import (
	"context"

	"github.com/trigg3rX/triggerx-backend/pkg/docker/types"
)

type ExecutionAPI interface {
	Execute(ctx context.Context, fileURL string, fileLanguage string, noOfAttesters int) (*types.ExecutionResult, error)
	GetHealthStatus() *HealthStatus
	GetStats() *types.PerformanceMetrics
	GetPoolStats() map[types.Language]*types.PoolStats
	InitializeLanguagePools(ctx context.Context, languages []types.Language) error
	GetSupportedLanguages() []types.Language
	IsLanguageSupported(language types.Language) bool
	GetActiveExecutions() []*types.ExecutionContext
	GetAlerts(severity string, limit int) []Alert
	ClearAlerts()
	CancelExecution(executionID string) error
	Close(ctx context.Context) error
}
