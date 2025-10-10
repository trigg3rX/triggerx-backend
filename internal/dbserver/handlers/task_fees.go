package handlers

import (
	"context"
	"fmt"
	"math/big"

	"net/http"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/metrics"
	"github.com/trigg3rX/triggerx-backend/pkg/dockerexecutor/types"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

func (h *Handler) CalculateTaskFees(ipfsURLs string, logger logging.Logger) (*big.Int, error) {
	if ipfsURLs == "" {
		return big.NewInt(0), fmt.Errorf("missing IPFS URLs")
	}

	trackDBOp := metrics.TrackDBOperation("read", "task_fees")
	urlList := strings.Split(ipfsURLs, ",")
	totalFee := big.NewInt(0)
	var mu sync.Mutex
	var wg sync.WaitGroup

	ctx := context.Background()

	for _, ipfsURL := range urlList {
		ipfsURL = strings.TrimSpace(ipfsURL)
		wg.Add(1)

		go func(url string) {
			defer wg.Done()

			// Use the Execute method directly which handles all the Docker-in-Docker compatibility
			result, err := h.dockerExecutor.Execute(ctx, url, string(types.LanguageGo), 10)
			if err != nil {
				logger.Errorf("Error executing code: %v", err)
				return
			}

			if !result.Success {
				logger.Errorf("Code execution failed: %v", result.Error)
				return
			}

			mu.Lock()
			totalFee.Add(totalFee, result.Stats.TotalCost)
			mu.Unlock()
		}(ipfsURL)
	}

	wg.Wait()
	trackDBOp(nil)
	return totalFee, nil
}

func (h *Handler) GetTaskFees(c *gin.Context) {
	logger := h.getLogger(c)
	logger.Debugf("GET [GetTaskFees] Getting task fees")
	ipfsURLs := c.Query("ipfs_url")

	totalFee, err := h.CalculateTaskFees(ipfsURLs, logger)
	if err != nil {
		logger.Errorf("Error calculating fees: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"total_fee": totalFee,
	})
}
