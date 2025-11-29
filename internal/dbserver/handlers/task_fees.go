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
)

func (h *Handler) CalculateTaskFees(ipfsURLs string, taskDefinitionID int, targetChainID, targetContractAddress, targetFunction, abi string) (*big.Int, error) {
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

			// Prepare metadata for fee calculation
			metadata := map[string]string{
				"task_definition_id":       fmt.Sprintf("%d", taskDefinitionID),
				"target_chain_id":          targetChainID,
				"target_contract_address":  targetContractAddress,
				"target_function":          targetFunction,
				"abi":                      abi,
			}

			// Use the Execute method with metadata for accurate fee calculation
			result, err := h.dockerExecutor.Execute(ctx, url, string(types.LanguageGo), 10, metadata)
			if err != nil {
				h.logger.Errorf("Error executing code: %v", err)
				return
			}

			if !result.Success {
				h.logger.Errorf("Code execution failed: %v", result.Error)
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
	traceID := h.getTraceID(c)
	h.logger.Infof("[GetTaskFees] trace_id=%s - Getting task fees", traceID)

	// Get query parameters
	ipfsURLs := c.Query("ipfs_url")
	taskDefID := c.DefaultQuery("task_definition_id", "0")
	targetChainID := c.Query("target_chain_id")
	targetContractAddress := c.Query("target_contract_address")
	targetFunction := c.Query("target_function")
	abi := c.Query("abi")

	// Parse task definition ID
	taskDefinitionID := 0
	if parsed, err := fmt.Sscanf(taskDefID, "%d", &taskDefinitionID); err != nil || parsed != 1 {
		h.logger.Warnf("[GetTaskFees] Invalid task_definition_id: %s, using 0", taskDefID)
	}

	totalFee, err := h.CalculateTaskFees(ipfsURLs, taskDefinitionID, targetChainID, targetContractAddress, targetFunction, abi)
	if err != nil {
		h.logger.Errorf("[GetTaskFees] Error calculating fees: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"total_fee": totalFee,
		"total_fee_wei": totalFee.String(),
	})
}
