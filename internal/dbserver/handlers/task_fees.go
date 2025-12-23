package handlers

import (
	"context"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/config"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/metrics"
	"github.com/trigg3rX/triggerx-backend/pkg/dockerexecutor/types"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
)

func (h *Handler) CalculateTaskFees(ctx context.Context, ipfsURLs string, taskDefinitionID int, targetChainID, targetContractAddress, targetFunction, abi, args, fromAddress string) (*big.Int, *big.Int, error) {
	// TaskDefinitionID 2, 4, 6 require ipfsURL(s) for dynamic argument scripts
	// TaskDefinitionID 7 (Custom Script) does NOT require IPFS script execution during job creation
	// For ID 7, fee estimation uses fixed 1M gas (handled in pipeline.go calculateFees)
	needsIPFS := taskDefinitionID == 2 || taskDefinitionID == 4 || taskDefinitionID == 6

	if needsIPFS && ipfsURLs == "" {
		return big.NewInt(0), big.NewInt(0), fmt.Errorf("missing IPFS URLs")
	}

	trackDBOp := metrics.TrackDBOperation("read", "task_fees")
	totalFee := big.NewInt(0)
	currentTotalFee := big.NewInt(0)

	var mu sync.Mutex
	var wg sync.WaitGroup

	if needsIPFS {
		urlList := strings.Split(ipfsURLs, ",")
		for _, ipfsURL := range urlList {
			ipfsURL = strings.TrimSpace(ipfsURL)
			wg.Add(1)

			go func(url, from string) {
				defer wg.Done()

				metadata := map[string]string{
					"task_definition_id":      fmt.Sprintf("%d", taskDefinitionID),
					"target_chain_id":         targetChainID,
					"target_contract_address": targetContractAddress,
					"target_function":         targetFunction,
					"abi":                     abi,
					"on_chain_args":           args,
					"from_address":            from,
				}

				// Dynamic argument scripts (2, 4, 6) use Go
				result, err := h.dockerExecutor.Execute(ctx, url, string(types.LanguageGo), 10, config.GetAlchemyAPIKey(), metadata)
				if err != nil {
					h.logger.Error(ctx, "Error executing code", observability.Error(err))
					return
				}

				if !result.Success {
					h.logger.Error(ctx, "Code execution failed", observability.Error(fmt.Errorf("code execution failed: %v", result.Error)))
					return
				}

				mu.Lock()
				totalFee.Add(totalFee, result.Stats.TotalCost)
				currentTotalFee.Add(currentTotalFee, result.Stats.CurrentTotalCost)
				mu.Unlock()
			}(ipfsURL, fromAddress)
		}
		wg.Wait()
	} else {
		// No IPFS required; just invoke Execute with empty code/url and rely on metadata for fee calculation
		metadata := map[string]string{
			"task_definition_id":      fmt.Sprintf("%d", taskDefinitionID),
			"target_chain_id":         targetChainID,
			"target_contract_address": targetContractAddress,
			"target_function":         targetFunction,
			"abi":                     abi,
			"on_chain_args":           args,
			"from_address":            fromAddress,
		}
		result, err := h.dockerExecutor.Execute(ctx, "", string(types.LanguageGo), 10, config.GetAlchemyAPIKey(), metadata)
		if err != nil {
			h.logger.Error(ctx, "Error executing code", observability.Error(err))
			return big.NewInt(0), big.NewInt(0), err
		}
		if !result.Success {
			h.logger.Error(ctx, "Code execution failed", observability.Error(fmt.Errorf("code execution failed: %v", result.Error)))
			return big.NewInt(0), big.NewInt(0), fmt.Errorf("code execution failed")
		}
		totalFee.Set(result.Stats.TotalCost)
		currentTotalFee.Set(result.Stats.CurrentTotalCost)
	}

	trackDBOp(nil)
	return totalFee, currentTotalFee, nil
}

func (h *Handler) GetTaskFees(c *gin.Context) {
	traceID := h.getTraceID(c)
	h.logger.Info(c.Request.Context(), "[GetTaskFees] Getting task fees", observability.String("trace_id", traceID))

	// Get query parameters
	ipfsURLs := c.Query("ipfs_url")
	taskDefID := c.DefaultQuery("task_definition_id", "0")
	targetChainID := c.Query("target_chain_id")
	targetContractAddress := c.Query("target_contract_address")
	targetFunction := c.Query("target_function")
	abi := c.Query("abi")

	args := c.Query("args")

	// Default to testnet address unless targetChainID is 42161 or 8453 (mainnet/arbitrum)
	fromAddress := config.GetTestTaskExecutionAddress()
	if targetChainID == "42161" || targetChainID == "8453" {
		fromAddress = config.GetTaskExecutionAddress()
	}

	// Parse task definition ID
	taskDefinitionID := 0
	if parsed, err := fmt.Sscanf(taskDefID, "%d", &taskDefinitionID); err != nil || parsed != 1 {
		h.logger.Warn(c.Request.Context(), "[GetTaskFees] Validation failed: Invalid task_definition_id", observability.String("task_definition_id", taskDefID))
	}

	totalFee, currentTotalFee, err := h.CalculateTaskFees(c.Request.Context(), ipfsURLs, taskDefinitionID, targetChainID, targetContractAddress, targetFunction, abi, args, fromAddress)
	if err != nil {
		h.logger.Error(c.Request.Context(), "[GetTaskFees] Error calculating fees", observability.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"total_fee":             totalFee,
		"total_fee_wei":         totalFee.String(),
		"current_total_fee":     currentTotalFee,
		"current_total_fee_wei": currentTotalFee.String(),
	})
}
