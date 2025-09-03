package file

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/trigg3rX/triggerx-backend/pkg/docker/config"
	"github.com/trigg3rX/triggerx-backend/pkg/docker/types"
	httppkg "github.com/trigg3rX/triggerx-backend/pkg/http"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

type downloader struct {
	client    *httppkg.HTTPClient
	cache     *fileCache
	validator *codeValidator
	logger    logging.Logger
}

type downloadResult struct {
	FilePath   string
	Content    []byte
	Hash       string
	Size       int64
	IsCached   bool
	Validation *types.ValidationResult
}

func newDownloader(cfg config.CacheConfig, validationCfg config.ValidationConfig, httpClient *httppkg.HTTPClient, logger logging.Logger) (*downloader, error) {
	cache, err := newFileCache(cfg, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create file cache: %w", err)
	}

	validator := newCodeValidator(validationCfg, logger)

	return &downloader{
		client:    httpClient,
		cache:     cache,
		validator: validator,
		logger:    logger,
	}, nil
}

func (d *downloader) downloadFile(ctx context.Context, key string, url string) (*downloadResult, error) {
	// Get file from cache or download it
	var isCached bool
	filePath, err := d.cache.getOrDownloadFile(key, func() ([]byte, error) {
		return d.downloadContent(ctx, url)
	})
	if err != nil {
		return nil, fmt.Errorf("failed to download or store file in cache: %w", err)
	}
	isCached = true
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read downloaded file: %w", err)
	}
	d.logger.Infof("File downloaded and stored in cache: %s", key)

	// Validate content (either fresh or from cache)
	validation, err := d.validator.validateFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to validate content: %w", err)
	}

	if !validation.IsValid {
		d.logger.Warnf("File validation failed: %v", validation.Errors)
		return &downloadResult{
			Content:    content,
			Validation: validation,
		}, nil
	}

	fileInfo, err := os.Stat(filePath)
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
			d.logger.Error("Error closing response body", "error", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	d.logger.Debugf("Downloaded %d bytes", len(content))
	return content, nil
}

func (d *downloader) close() error {
	if d.cache != nil {
		return d.cache.close()
	}
	return nil
}
