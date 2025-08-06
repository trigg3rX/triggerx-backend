package file

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/trigg3rX/triggerx-backend/pkg/docker/config"
	"github.com/trigg3rX/triggerx-backend/pkg/docker/types"
	httppkg "github.com/trigg3rX/triggerx-backend/pkg/http"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

type Downloader struct {
	client    *httppkg.HTTPClient
	cache     *FileCache
	validator *CodeValidator
	logger    logging.Logger
}

type DownloadResult struct {
	FilePath   string
	Content    []byte
	Hash       string
	Size       int64
	IsCached   bool
	Validation *types.ValidationResult
}

func NewDownloader(cfg config.CacheConfig, validationCfg config.ValidationConfig, logger logging.Logger) (*Downloader, error) {
	httpClient, err := httppkg.NewHTTPClient(httppkg.DefaultHTTPRetryConfig(), logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP client: %w", err)
	}

	cache, err := NewFileCache(cfg, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create file cache: %w", err)
	}

	validator := NewCodeValidator(validationCfg, logger)

	return &Downloader{
		client:    httpClient,
		cache:     cache,
		validator: validator,
		logger:    logger,
	}, nil
}

func (d *Downloader) DownloadFile(key string, url string) (*DownloadResult, error) {
	// Try to get from cache first
	cachedPath, err := d.cache.GetByKey(key)
	var filePath string
	isCached := false
	var content []byte

	if err == nil {
		// File found in cache
		filePath = cachedPath
		isCached = true
		content, err = os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read cached file: %w", err)
		}
		d.logger.Infof("File found in cache: %s", key)
	} else {
		// File not in cache, download and store it
		filePath, err = d.cache.GetOrDownload(key, func() ([]byte, error) {
			return d.downloadContent(url)
		})
		if err != nil {
			return nil, fmt.Errorf("failed to download or store file in cache: %w", err)
		}
		content, err = os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read downloaded file: %w", err)
		}
		d.logger.Infof("File downloaded and stored in cache: %s", key)
	}

	// Validate content
	validation, err := d.validator.ValidateContent(content)
	if err != nil {
		return nil, fmt.Errorf("failed to validate content: %w", err)
	}

	if !validation.IsValid {
		d.logger.Warnf("File validation failed: %v", validation.Errors)
		return &DownloadResult{
			Content:    content,
			Validation: validation,
		}, nil
	}

	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	return &DownloadResult{
		FilePath:   filePath,
		Content:    content,
		Hash:       key,
		Size:       fileInfo.Size(),
		IsCached:   isCached,
		Validation: validation,
	}, nil
}

func (d *Downloader) downloadContent(url string) ([]byte, error) {
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

func (d *Downloader) GetCacheStats() *types.CacheStats {
	return d.cache.GetStats()
}

func (d *Downloader) ValidateFile(filePath string) (*types.ValidationResult, error) {
	return d.validator.ValidateFile(filePath)
}

func (d *Downloader) ValidateContent(content []byte) (*types.ValidationResult, error) {
	return d.validator.ValidateContent(content)
}

func (d *Downloader) Close() error {
	if d.cache != nil {
		return d.cache.Close()
	}
	return nil
}
