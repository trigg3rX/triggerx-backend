package file

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/trigg3rX/triggerx-backend/pkg/dockerexecutor/config"
	"github.com/trigg3rX/triggerx-backend/pkg/dockerexecutor/types"
	fs "github.com/trigg3rX/triggerx-backend/pkg/filesystem"
	httppkg "github.com/trigg3rX/triggerx-backend/pkg/http"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
)

type downloader struct {
	client    httppkg.HTTPClientInterface
	cache     *fileCache
	validator *codeValidator
	logger    observability.Logger
	fs        fs.FileSystemAPI
}

type downloadResult struct {
	FilePath   string
	Content    []byte
	Hash       string
	Size       int64
	IsCached   bool
	Validation *types.ValidationResult
}

func newDownloader(ctx context.Context, cfg config.FileCacheConfig, validationCfg config.ValidationConfig, httpClient httppkg.HTTPClientInterface, logger observability.Logger, fs fs.FileSystemAPI) (*downloader, error) {
	cache, err := newFileCache(ctx, cfg, logger, fs)
	if err != nil {
		return nil, fmt.Errorf("failed to create file cache: %w", err)
	}

	validator := newCodeValidator(validationCfg, logger, fs)

	return &downloader{
		client:    httpClient,
		cache:     cache,
		validator: validator,
		logger:    logger,
		fs:        fs,
	}, nil
}

func (d *downloader) downloadFile(ctx context.Context, key string, url string, fileLanguage string) (*downloadResult, error) {
	// Get file from cache or download it
	var isCached bool
	filePath, err := d.cache.getOrDownloadFile(ctx, key, fileLanguage, func() ([]byte, error) {
		return d.downloadContent(ctx, url)
	})
	if err != nil {
		return nil, fmt.Errorf("failed to download or store file in cache: %w", err)
	}
	isCached = true
	content, err := d.fs.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read downloaded file: %w", err)
	}

	// Validate content (either fresh or from cache)
	validation, err := d.validator.validateFile(ctx, filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to validate content: %w", err)
	}

	if !validation.IsValid {
		d.logger.Warn(ctx, "File validation failed", observability.Any("validationErrors", validation.Errors))
		return &downloadResult{
			Content:    content,
			Validation: validation,
		}, nil
	}

	fileInfo, err := d.fs.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	return &downloadResult{
		FilePath:   filePath,
		Content:    content,
		Hash:       key,
		Size:       fileInfo.Size(),
		IsCached:   isCached,
		Validation: validation,
	}, nil
}

func (d *downloader) downloadContent(ctx context.Context, url string) ([]byte, error) {
	resp, err := d.client.Get(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("failed to download file: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			d.logger.Error(ctx, "Error closing response body", observability.Error(err))
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	d.logger.Debug(ctx, "File downloaded", observability.Int("size", len(content)))
	return content, nil
}

func (d *downloader) close() error {
	if d.cache != nil {
		return d.cache.close()
	}
	return nil
}
