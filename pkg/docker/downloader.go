package docker

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/trigg3rX/triggerx-backend-imua/pkg/logging"
	"github.com/trigg3rX/triggerx-backend-imua/pkg/retry"
)

type Downloader struct {
	client *retry.HTTPClient
}

func NewDownloader(logger logging.Logger) (*Downloader, error) {
	httpClient, err := retry.NewHTTPClient(retry.DefaultHTTPRetryConfig(), logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP client: %w", err)
	}

	return &Downloader{
		client: httpClient,
	}, nil
}

func (d *Downloader) DownloadFile(url string, logger logging.Logger) (string, error) {
	// Always use /tmp directory for Docker-in-Docker compatibility
	tmpDir, err := os.MkdirTemp("/tmp", "ipfs-code")
	if err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}

	// Set directory permissions to be accessible from Docker-in-Docker
	if err := os.Chmod(tmpDir, 0777); err != nil {
		logger.Warnf("Failed to set directory permissions: %v", err)
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := d.client.DoWithRetry(req)
	if err != nil {
		return "", fmt.Errorf("failed to download file: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			logger.Error("Error closing response body", "error", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	filePath := filepath.Join(tmpDir, "code.go")
	out, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create local file: %w", err)
	}
	defer func() {
		if err := out.Close(); err != nil {
			logger.Error("Error closing output file", "error", err)
		}
	}()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to save file: %w", err)
	}

	// Set file permissions to be accessible from Docker-in-Docker
	if err := os.Chmod(filePath, 0666); err != nil {
		logger.Warnf("Failed to set file permissions: %v", err)
	}

	logger.Infof("Downloaded code to: %s with permissions 0666", filePath)

	// Verify the file exists and has correct permissions
	if info, err := os.Stat(filePath); err == nil {
		logger.Infof("File %s exists with mode: %v", filePath, info.Mode())
	} else {
		logger.Warnf("Could not verify file: %v", err)
	}

	return filePath, nil
}
