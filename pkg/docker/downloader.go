package docker

import (
	"context"
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

func (d *Downloader) DownloadFile(ctx context.Context, url string, logger logging.Logger) (string, error) {
	tmpDir, err := os.MkdirTemp("", "ipfs-code")
	if err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
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

	return filePath, nil
}
