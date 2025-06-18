package handlers

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/metrics"
	"github.com/trigg3rX/triggerx-backend/pkg/docker"
)

func (h *Handler) CalculateTaskFees(ipfsURLs string) (float64, error) {
	if ipfsURLs == "" {
		return 0, fmt.Errorf("missing IPFS URLs")
	}

	trackDBOp := metrics.TrackDBOperation("read", "task_fees")
	urlList := strings.Split(ipfsURLs, ",")
	totalFee := 0.0
	var mu sync.Mutex
	var wg sync.WaitGroup

	ctx := context.Background()

	executor, err := docker.NewCodeExecutor(ctx, docker.ExecutorConfig{
		Docker: h.docker,
	}, h.logger)
	if err != nil {
		trackDBOp(err)
		return 0, fmt.Errorf("failed to create code executor: %v", err)
	}

	for _, ipfsURL := range urlList {
		ipfsURL = strings.TrimSpace(ipfsURL)
		wg.Add(1)

		go func(url string) {
			defer wg.Done()

			codePath, err := executor.Downloader.DownloadFile(ctx, url, h.logger)
			if err != nil {
				h.logger.Errorf("Error downloading IPFS file: %v", err)
				return
			}
			defer func() {
				if err := os.RemoveAll(filepath.Dir(codePath)); err != nil {
					h.logger.Errorf("Error removing temporary directory: %v", err)
				}
			}()

			containerID, err := executor.DockerManager.CreateContainer(ctx, codePath)
			if err != nil {
				h.logger.Errorf("Error creating container: %v", err)
				return
			}
			if err := executor.DockerManager.CleanupContainer(ctx, containerID); err != nil {
				h.logger.Errorf("Error removing container: %v", err)
			}

			result, err := executor.MonitorExecution(ctx, executor.DockerManager.Cli, containerID, 10)
			if err != nil {
				h.logger.Errorf("Error monitoring resources: %v", err)
				return
			}

			mu.Lock()
			totalFee += result.Stats.TotalCost
			mu.Unlock()
		}(ipfsURL)
	}

	wg.Wait()
	trackDBOp(nil)
	return totalFee, nil
}

func (h *Handler) GetTaskFees(c *gin.Context) {
	ipfsURLs := c.Query("ipfs_url")

	totalFee, err := h.CalculateTaskFees(ipfsURLs)
	if err != nil {
		h.logger.Errorf("[GetTaskFees] Error calculating fees: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"total_fee": totalFee,
	})
}
