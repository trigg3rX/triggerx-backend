package file

import (
	"fmt"
	"io"
	"net/http"

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

func newDownloader(cfg config.FileCacheConfig, validationCfg config.ValidationConfig, httpClient httppkg.HTTPClientInterface, logger logging.Logger, fs fs.FileSystemAPI) (*downloader, error) {
	cache, err := newFileCache(cfg, logger, fs)
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

func (d *downloader) downloadFile(key string, url string) (*downloadResult, error) {
	// Get file from cache or download it
	var isCached bool
	filePath, err := d.cache.getOrDownloadFile(key, func() ([]byte, error) {
		return d.downloadContent(url)
	})
	if err != nil {
		return nil, fmt.Errorf("failed to download or store file in cache: %w", err)
	}
	isCached = true
	content, err := d.fs.ReadFile(filePath)
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

func (d *downloader) downloadContent(url string) ([]byte, error) {
	resp, err := d.client.Get(url)
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
