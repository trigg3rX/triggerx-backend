package file

//go:generate mockgen -source=interface.go -destination=mock_file_manager.go -package=file

import (
	"context"

	"github.com/trigg3rX/triggerx-backend/pkg/docker/types"
)

// DownloaderAPI defines the methods we need from the downloader.
type FileManagerAPI interface {
	GetOrDownload(ctx context.Context, fileURL string, fileLanguage string) (*types.ExecutionContext, error)
	GetCacheStats() *types.CacheStats
	GetPerformanceStats() *types.PerformanceMetrics
	Close() error
}
