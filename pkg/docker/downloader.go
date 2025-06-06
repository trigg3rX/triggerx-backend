package docker

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type Downloader struct {
	timeoutSeconds int
	client *http.Client
}

func NewDownloader(timeoutSeconds int) *Downloader {
	return &Downloader{
		timeoutSeconds: timeoutSeconds,
		client: &http.Client{
			Timeout: time.Duration(timeoutSeconds) * time.Second,
		},
	}
}

func (d *Downloader) DownloadFile(ctx context.Context, url string) (string, error) {
	tmpDir, err := os.MkdirTemp("", "ipfs-code")
	if err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := d.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to download file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	filePath := filepath.Join(tmpDir, "code.go")
	out, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create local file: %w", err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to save file: %w", err)
	}

	return filePath, nil
}