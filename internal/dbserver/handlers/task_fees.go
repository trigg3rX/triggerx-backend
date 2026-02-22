package handlers

import (
	"context"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/trigg3rX/triggerx-backend/internal/dbserver/config"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/metrics"
	"github.com/trigg3rX/triggerx-backend/pkg/dockerexecutor/types"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
)

func (h *Handler) CalculateTaskFees(ctx context.Context, ipfsURLs string, taskDefinitionID int, targetChainID, targetContractAddress, targetFunction, abi, args, fromAddress, jobOwnerAddress string) (*big.Int, *big.Int, error) {
	// Span: Fee calculation
	ctx, calcSpan := h.tracer.Start(ctx, "fee.calculate",
		observability.WithSpanKind(trace.SpanKindInternal),
		observability.WithAttributes(
			attribute.Int("task_definition_id", taskDefinitionID),
			attribute.String("target_chain_id", targetChainID),
		),
	)
	defer calcSpan.End()

	// TaskDefinitionID 2, 4, 6 require ipfsURL(s) for dynamic argument scripts
	// TaskDefinitionID 7 (Custom Script) does NOT require IPFS script execution during job creation
	// For ID 7, fee estimation uses fixed 1M gas (handled in pipeline.go calculateFees)
	needsIPFS := taskDefinitionID == 2 || taskDefinitionID == 4 || taskDefinitionID == 6

	if needsIPFS && ipfsURLs == "" {
		err := fmt.Errorf("missing IPFS URLs")
		calcSpan.RecordError(err)
		calcSpan.SetStatus(codes.Error, err.Error())
		return big.NewInt(0), big.NewInt(0), err
	}

	trackDBOp := metrics.TrackDBOperation("read", "task_fees")
	totalFee := big.NewInt(0)
	currentTotalFee := big.NewInt(0)

	var mu sync.Mutex
	var wg sync.WaitGroup

	if needsIPFS {
		urlList := strings.Split(ipfsURLs, ",")
		calcSpan.SetAttributes(attribute.Int("ipfs_url_count", len(urlList)))

		for _, ipfsURL := range urlList {
			ipfsURL = strings.TrimSpace(ipfsURL)
			wg.Add(1)

			go func(url, from string) {
				defer wg.Done()

				// Span: Docker execution for each IPFS URL
				_, dockerSpan := h.tracer.Start(ctx, "docker.execute",
					observability.WithSpanKind(trace.SpanKindClient),
					observability.WithAttributes(
						attribute.String("ipfs_url", url),
						attribute.String("language", string(types.LanguageGo)),
						attribute.Int("timeout", 10),
					),
				)
				defer dockerSpan.End()

				metadata := map[string]string{
					"task_definition_id":      fmt.Sprintf("%d", taskDefinitionID),
					"target_chain_id":         targetChainID,
					"target_contract_address": targetContractAddress,
					"target_function":         targetFunction,
					"abi":                     abi,
					"on_chain_args":           args,
					"from_address":            from,
					"job_owner_address":       jobOwnerAddress,
				}

				// Dynamic argument scripts (2, 4, 6) use Go
				result, err := h.dockerExecutor.Execute(ctx, url, string(types.LanguageGo), 10, config.GetAlchemyAPIKey(), metadata)
				if err != nil {
					dockerSpan.RecordError(err)
					dockerSpan.SetStatus(codes.Error, "docker execution failed")
					h.logger.Error(ctx, "Error executing code", observability.Error(err))
					return
				}

				if !result.Success {
					execErr := fmt.Errorf("code execution failed: %v", result.Error)
					dockerSpan.RecordError(execErr)
					dockerSpan.SetStatus(codes.Error, "code execution failed")
					h.logger.Error(ctx, "Code execution failed", observability.Error(execErr))
					return
				}

				dockerSpan.SetStatus(codes.Ok, "")
				mu.Lock()
				totalFee.Add(totalFee, result.Stats.TotalCost)
				currentTotalFee.Add(currentTotalFee, result.Stats.CurrentTotalCost)
				mu.Unlock()
			}(ipfsURL, fromAddress)
		}
		wg.Wait()
	} else {
		// Span: Docker execution without IPFS
		_, dockerSpan := h.tracer.Start(ctx, "docker.execute",
			observability.WithSpanKind(trace.SpanKindClient),
			observability.WithAttributes(
				attribute.String("language", string(types.LanguageGo)),
				attribute.Int("timeout", 10),
				attribute.Bool("needs_ipfs", false),
			),
		)

		// No IPFS required; just invoke Execute with empty code/url and rely on metadata for fee calculation
		metadata := map[string]string{
			"task_definition_id":      fmt.Sprintf("%d", taskDefinitionID),
			"target_chain_id":         targetChainID,
			"target_contract_address": targetContractAddress,
			"target_function":         targetFunction,
			"abi":                     abi,
			"on_chain_args":           args,
			"from_address":            fromAddress,
			"job_owner_address":       jobOwnerAddress,
		}
		result, err := h.dockerExecutor.Execute(ctx, "", string(types.LanguageGo), 10, config.GetAlchemyAPIKey(), metadata)
		if err != nil {
			dockerSpan.RecordError(err)
			dockerSpan.SetStatus(codes.Error, "docker execution failed")
			dockerSpan.End()
			h.logger.Error(ctx, "Error executing code", observability.Error(err))
			return big.NewInt(0), big.NewInt(0), err
		}
		if !result.Success {
			execErr := fmt.Errorf("code execution failed: %v", result.Error)
			dockerSpan.RecordError(execErr)
			dockerSpan.SetStatus(codes.Error, "code execution failed")
			dockerSpan.End()
			h.logger.Error(ctx, "Code execution failed", observability.Error(execErr))
			return big.NewInt(0), big.NewInt(0), fmt.Errorf("code execution failed")
		}
		dockerSpan.SetStatus(codes.Ok, "")
		dockerSpan.End()
		totalFee.Set(result.Stats.TotalCost)
		currentTotalFee.Set(result.Stats.CurrentTotalCost)
	}

	calcSpan.SetStatus(codes.Ok, "")
	trackDBOp(nil)
	return totalFee, currentTotalFee, nil
}

func (h *Handler) GetTaskFees(c *gin.Context) {

	// Get query parameters
	ipfsURLs := c.Query("ipfs_url")
	taskDefID := c.DefaultQuery("task_definition_id", "0")
	targetChainID := c.Query("target_chain_id")
	targetContractAddress := c.Query("target_contract_address")
	targetFunction := c.Query("target_function")
	abi := c.Query("abi")

	args := c.Query("args")
	jobOwnerAddress := c.Query("job_owner_address")

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

	totalFee, currentTotalFee, err := h.CalculateTaskFees(context.WithoutCancel(c.Request.Context()), ipfsURLs, taskDefinitionID, targetChainID, targetContractAddress, targetFunction, abi, args, fromAddress, jobOwnerAddress)
	if err != nil {
		h.logger.Warn(c.Request.Context(), "[GetTaskFees] Failed to calculate fees", observability.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"total_fee":             totalFee,
		"total_fee_wei":         totalFee.String(),
		"current_total_fee":     currentTotalFee,
		"current_total_fee_wei": currentTotalFee.String(),
	})
	h.logger.Debug(c.Request.Context(), "[GetTaskFees] Calculated task fees", observability.Int("task_definition_id", taskDefinitionID))
}
