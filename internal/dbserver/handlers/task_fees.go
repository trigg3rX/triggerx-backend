package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend-imua/internal/dbserver/metrics"
	"github.com/trigg3rX/triggerx-backend-imua/pkg/docker"
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

	executor, err := docker.NewCodeExecutor(ctx, h.executor, h.logger)
	if err != nil {
		trackDBOp(err)
		return 0, fmt.Errorf("failed to create code executor: %v", err)
	}

	for _, ipfsURL := range urlList {
		ipfsURL = strings.TrimSpace(ipfsURL)
		wg.Add(1)

		go func(url string) {
			defer wg.Done()

			// Use the Execute method directly which handles all the Docker-in-Docker compatibility
			result, err := executor.Execute(ctx, url, 10)
			if err != nil {
				h.logger.Errorf("Error executing code: %v", err)
				return
			}

			if !result.Success {
				h.logger.Errorf("Code execution failed: %v", result.Error)
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
	traceID := h.getTraceID(c)
	h.logger.Infof("[GetTaskFees] trace_id=%s - Getting task fees", traceID)
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
