package file

import (
	"context"

	"github.com/trigg3rX/triggerx-backend/pkg/docker/types"
)

// DownloaderAPI defines the methods we need from the downloader.
type FileManagerAPI interface {
	GetOrDownload(ctx context.Context, fileURL string) (*types.ExecutionContext, error)
	ValidateFile(filePath string) (*types.ValidationResult, error)
	ValidateContent(content []byte) (*types.ValidationResult, error)
	GetFileByKey(key string) (string, error)
	GetCacheStats() *types.CacheStats
	GetPerformanceStats() *types.PerformanceMetrics
	CleanupFile(filePath string) error
}
