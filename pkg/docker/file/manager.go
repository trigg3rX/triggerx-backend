package file

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/trigg3rX/triggerx-backend/pkg/docker/config"
	"github.com/trigg3rX/triggerx-backend/pkg/docker/types"
	httppkg "github.com/trigg3rX/triggerx-backend/pkg/http"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

type FileManager struct {
	downloader *downloader
	config     config.ExecutorConfig
	logger     logging.Logger
	mutex      sync.RWMutex
	stats      *types.PerformanceMetrics
}

func NewFileManager(cfg config.ExecutorConfig, httpClient *httppkg.HTTPClient, logger logging.Logger) (*FileManager, error) {
	downloader, err := newDownloader(cfg.Cache, cfg.Validation, httpClient, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create downloader: %w", err)
	}

	return &FileManager{
		downloader: downloader,
		config:     cfg,
		logger:     logger,
		stats: &types.PerformanceMetrics{
			TotalExecutions:      0,
			SuccessfulExecutions: 0,
			FailedExecutions:     0,
			AverageExecutionTime: 0,
			MinExecutionTime:     0,
			MaxExecutionTime:     0,
			TotalCost:            0.0,
			AverageCost:          0.0,
			LastExecution:        time.Time{},
		},
	}, nil
}

func (fm *FileManager) GetOrDownload(ctx context.Context, fileURL string) (*types.ExecutionContext, error) {
	startTime := time.Now()
	fm.logger.Debugf("Processing file: %s", fileURL)

	// Download and validate file
	result, err := fm.downloader.downloadFile(ctx, fileURL, fileURL)
	if err != nil {
		fm.updateStats(false, time.Since(startTime), 0.0)
		return nil, fmt.Errorf("failed to download file: %w", err)
	}

	// Check validation results
	if !result.Validation.IsValid {
		fm.logger.Warnf("File validation failed: %v", result.Validation.Errors)
		fm.updateStats(false, time.Since(startTime), 0.0)
		return &types.ExecutionContext{
			FileURL:   fileURL,
			StartedAt: startTime,
			Metadata: map[string]string{
				"validation_errors": fmt.Sprintf("%v", result.Validation.Errors),
				"is_cached":         fmt.Sprintf("%v", result.IsCached),
			},
		}, nil
	}

	// Create execution context
	execCtx := &types.ExecutionContext{
		FileURL:   fileURL,
		StartedAt: startTime,
		Metadata: map[string]string{
			"file_path":  result.FilePath,
			"file_hash":  result.Hash,
			"file_size":  fmt.Sprintf("%d", result.Size),
			"is_cached":  fmt.Sprintf("%v", result.IsCached),
			"complexity": fmt.Sprintf("%.2f", result.Validation.Complexity),
			"warnings":   fmt.Sprintf("%v", result.Validation.Warnings),
		},
	}

	// Update statistics
	fm.updateStats(true, time.Since(startTime), result.Validation.Complexity)

	fm.logger.Debugf("File processed successfully (cached: %v, size: %d bytes)", result.IsCached, result.Size)

	return execCtx, nil
}

func (fm *FileManager) GetPerformanceStats() *types.PerformanceMetrics {
	fm.mutex.RLock()
	defer fm.mutex.RUnlock()

	// Create a copy to avoid race conditions
	stats := *fm.stats
	return &stats
}

func (fm *FileManager) GetCacheStats() *types.CacheStats {
	return fm.downloader.cache.getCacheStats()
}

func (fm *FileManager) updateStats(success bool, duration time.Duration, complexity float64) {
	fm.mutex.Lock()
	defer fm.mutex.Unlock()

	fm.stats.TotalExecutions++
	fm.stats.LastExecution = time.Now()

	if success {
		fm.stats.SuccessfulExecutions++
	} else {
		fm.stats.FailedExecutions++
	}

	// Update execution time statistics
	if fm.stats.MinExecutionTime == 0 || duration < fm.stats.MinExecutionTime {
		fm.stats.MinExecutionTime = duration
	}
	if duration > fm.stats.MaxExecutionTime {
		fm.stats.MaxExecutionTime = duration
	}

	// Calculate average execution time
	if fm.stats.SuccessfulExecutions > 0 {
		totalDuration := fm.stats.AverageExecutionTime * time.Duration(fm.stats.SuccessfulExecutions-1)
		totalDuration += duration
		fm.stats.AverageExecutionTime = totalDuration / time.Duration(fm.stats.SuccessfulExecutions)
	} else {
		fm.stats.AverageExecutionTime = duration
	}

	// Update cost statistics (basic calculation)
	cost := fm.calculateCost(duration, complexity)
	fm.stats.TotalCost += cost
	if fm.stats.SuccessfulExecutions > 0 {
		fm.stats.AverageCost = fm.stats.TotalCost / float64(fm.stats.SuccessfulExecutions)
	} else {
		fm.stats.AverageCost = cost
	}
}

func (fm *FileManager) calculateCost(duration time.Duration, complexity float64) float64 {
	// Basic cost calculation based on execution time and complexity
	timeCost := duration.Seconds() * fm.config.Fees.PricePerTG
	complexityCost := complexity * fm.config.Fees.PricePerTG
	return timeCost + complexityCost + fm.config.Fees.FixedCost
}

func (fm *FileManager) Close() error {
	var errors []error

	if fm.downloader != nil {
		if err := fm.downloader.close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close downloader: %w", err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("errors during cleanup: %v", errors)
	}

	return nil
}
