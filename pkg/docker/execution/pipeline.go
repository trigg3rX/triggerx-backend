package execution

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/trigg3rX/triggerx-backend/pkg/docker/config"
	"github.com/trigg3rX/triggerx-backend/pkg/docker/container"
	"github.com/trigg3rX/triggerx-backend/pkg/docker/file"
	"github.com/trigg3rX/triggerx-backend/pkg/docker/types"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

type ExecutionPipeline struct {
	fileManager      *file.FileManager
	containerMgr     *container.Manager
	config           config.ExecutorConfig
	logger           logging.Logger
	mutex            sync.RWMutex
	activeExecutions map[string]*types.ExecutionContext
	stats            *types.PerformanceMetrics
}

func NewExecutionPipeline(cfg config.ExecutorConfig, fileMgr *file.FileManager, containerMgr *container.Manager, logger logging.Logger) *ExecutionPipeline {
	return &ExecutionPipeline{
		fileManager:      fileMgr,
		containerMgr:     containerMgr,
		config:           cfg,
		logger:           logger,
		activeExecutions: make(map[string]*types.ExecutionContext),
		stats: &types.PerformanceMetrics{
			TotalExecutions:      0,
			SuccessfulExecutions: 0,
			FailedExecutions:     0,
			AverageExecutionTime: 0,
			MinExecutionTime:     0,
			MaxExecutionTime:     0,
			TotalCost:            0.0,
			AverageCost:          0.0,
			LastExecution:        time.Time{},
		},
	}
}

func (ep *ExecutionPipeline) Execute(ctx context.Context, fileURL string, noOfAttesters int) (*types.ExecutionResult, error) {
	startTime := time.Now()
	executionID := generateExecutionID()

	ep.logger.Infof("Starting execution %s for file: %s", executionID, fileURL)

	// Create execution context
	execCtx := &types.ExecutionContext{
		FileURL:       fileURL,
		NoOfAttesters: noOfAttesters,
		TraceID:       executionID,
		StartedAt:     startTime,
		Metadata:      make(map[string]string),
	}

	// Track execution
	ep.mutex.Lock()
	ep.activeExecutions[executionID] = execCtx
	ep.mutex.Unlock()

	defer func() {
		// Remove from active executions
		ep.mutex.Lock()
		delete(ep.activeExecutions, executionID)
		ep.mutex.Unlock()

		// Update statistics
		duration := time.Since(startTime)
		ep.updateStats(true, duration, 0.0)
	}()

	// Execute pipeline stages
	result, err := ep.executeStages(ctx, execCtx)
	if err != nil {
		execCtx.CompletedAt = time.Now()
		ep.updateStats(false, time.Since(startTime), 0.0)
		return nil, fmt.Errorf("execution failed: %w", err)
	}

	execCtx.CompletedAt = time.Now()
	duration := time.Since(startTime)

	ep.logger.Infof("Execution %s completed successfully in %v", executionID, duration)
	return result, nil
}

func (ep *ExecutionPipeline) executeStages(ctx context.Context, execCtx *types.ExecutionContext) (*types.ExecutionResult, error) {
	// Stage 1: Download and Validate
	ep.logger.Debugf("Stage 1: Downloading and validating file")
	fileCtx, err := ep.fileManager.GetOrDownload(ctx, execCtx.FileURL)
	if err != nil {
		return nil, fmt.Errorf("failed to download file: %w", err)
	}

	// Check validation results
	if fileCtx.Metadata["validation_errors"] != "" {
		return &types.ExecutionResult{
			Success: false,
			Output:  "",
			Error:   fmt.Errorf("file validation failed: %s", fileCtx.Metadata["validation_errors"]),
		}, nil
	}

	// Stage 2: Get Container
	ep.logger.Debugf("Stage 2: Getting container from pool")
	container, err := ep.containerMgr.GetContainer(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get container: %w", err)
	}
	ep.logger.Debugf("Got container %s from pool", container.ID)

	// Return container to pool asynchronously
	defer func() {
		go func(containerID string, container *types.PooledContainer) {
			ep.logger.Debugf("Returning container %s to pool (async)", containerID)
			if err := ep.containerMgr.ReturnContainer(container); err != nil {
				ep.logger.Warnf("Failed to return container to pool: %v", err)
			}
		}(container.ID, container)
	}()

	// Stage 3: Execute Code
	ep.logger.Debugf("Stage 3: Executing code in container %s", container.ID)
	// Get file path from metadata
	filePath := fileCtx.Metadata["file_path"]
	if filePath == "" {
		return nil, fmt.Errorf("file path not found in execution context")
	}
	ep.logger.Debugf("File path from metadata: %s", filePath)

	result, err := ep.containerMgr.ExecuteInContainer(ctx, container.ID, filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to execute code: %w", err)
	}

	// Stage 4: Process Results
	ep.logger.Debugf("Stage 4: Processing results")
	finalResult := ep.processResults(result, execCtx)

	// Stage 5: Cleanup
	// ep.logger.Debugf("Stage 5: Cleaning up")
	// if err := ep.cleanupExecution(execCtx); err != nil {
	// 	ep.logger.Warnf("Failed to cleanup execution: %v", err)
	// }

	return finalResult, nil
}

func (ep *ExecutionPipeline) processResults(result *types.ExecutionResult, execCtx *types.ExecutionContext) *types.ExecutionResult {
	// Add execution metadata
	execCtx.Metadata["execution_time"] = result.Stats.ExecutionTime.String()
	execCtx.Metadata["static_complexity"] = fmt.Sprintf("%.6f", result.Stats.StaticComplexity)
	execCtx.Metadata["dynamic_complexity"] = fmt.Sprintf("%.6f", result.Stats.DynamicComplexity)

	// Calculate fees
	fees := ep.calculateFees(execCtx)
	execCtx.Metadata["fees"] = fmt.Sprintf("%.6f", fees)
	result.Stats.TotalCost = fees

	return result
}

func (ep *ExecutionPipeline) calculateFees(execCtx *types.ExecutionContext) float64 {
	// Basic fee calculation
	duration := time.Since(execCtx.StartedAt)
	baseFee := ep.config.Fees.FixedCost
	timeFee := duration.Seconds() * ep.config.Fees.PricePerTG

	// Add complexity factor if available
	complexityFee := 0.0
	if _, ok := execCtx.Metadata["complexity"]; ok {
		// Parse complexity and apply factor
		// This is a simplified calculation
		complexityFee = 0.1 // Placeholder
	}

	return baseFee + timeFee + complexityFee
}

// func (ep *ExecutionPipeline) cleanupExecution(execCtx *types.ExecutionContext) error {
// 	// Cleanup any temporary files
// 	// In this implementation, the file manager handles cleanup
// 	return nil
// }

func (ep *ExecutionPipeline) GetActiveExecutions() []*types.ExecutionContext {
	ep.mutex.RLock()
	defer ep.mutex.RUnlock()

	executions := make([]*types.ExecutionContext, 0, len(ep.activeExecutions))
	for _, exec := range ep.activeExecutions {
		executions = append(executions, exec)
	}

	return executions
}

func (ep *ExecutionPipeline) GetExecutionByID(executionID string) (*types.ExecutionContext, bool) {
	ep.mutex.RLock()
	defer ep.mutex.RUnlock()

	exec, exists := ep.activeExecutions[executionID]
	return exec, exists
}

func (ep *ExecutionPipeline) CancelExecution(executionID string) error {
	ep.mutex.Lock()
	defer ep.mutex.Unlock()

	exec, exists := ep.activeExecutions[executionID]
	if !exists {
		return fmt.Errorf("execution not found: %s", executionID)
	}

	exec.CompletedAt = time.Now()

	ep.logger.Infof("Execution %s cancelled", executionID)
	return nil
}

func (ep *ExecutionPipeline) GetStats() *types.PerformanceMetrics {
	ep.mutex.RLock()
	defer ep.mutex.RUnlock()

	// Create a copy to avoid race conditions
	stats := *ep.stats
	return &stats
}

func (ep *ExecutionPipeline) updateStats(success bool, duration time.Duration, complexity float64) {
	ep.mutex.Lock()
	defer ep.mutex.Unlock()

	ep.stats.TotalExecutions++
	ep.stats.LastExecution = time.Now()

	if success {
		ep.stats.SuccessfulExecutions++
	} else {
		ep.stats.FailedExecutions++
	}

	// Update execution time statistics
	if ep.stats.MinExecutionTime == 0 || duration < ep.stats.MinExecutionTime {
		ep.stats.MinExecutionTime = duration
	}
	if duration > ep.stats.MaxExecutionTime {
		ep.stats.MaxExecutionTime = duration
	}

	// Calculate average execution time - only if we have successful executions
	if ep.stats.SuccessfulExecutions > 0 {
		if ep.stats.SuccessfulExecutions == 1 {
			// First successful execution
			ep.stats.AverageExecutionTime = duration
		} else {
			// Calculate running average
			totalDuration := ep.stats.AverageExecutionTime * time.Duration(ep.stats.SuccessfulExecutions-1)
			totalDuration += duration
			ep.stats.AverageExecutionTime = totalDuration / time.Duration(ep.stats.SuccessfulExecutions)
		}
	}

	// Update cost statistics
	cost := ep.calculateCost(duration, complexity)
	ep.stats.TotalCost += cost

	// Calculate average cost - only if we have successful executions
	if ep.stats.SuccessfulExecutions > 0 {
		ep.stats.AverageCost = ep.stats.TotalCost / float64(ep.stats.SuccessfulExecutions)
	}
}

func (ep *ExecutionPipeline) calculateCost(duration time.Duration, complexity float64) float64 {
	// Basic cost calculation based on execution time and complexity
	timeCost := duration.Seconds() * ep.config.Fees.PricePerTG
	complexityCost := complexity * ep.config.Fees.PricePerTG
	return timeCost + complexityCost + ep.config.Fees.FixedCost
}

func generateExecutionID() string {
	return fmt.Sprintf("exec_%d", time.Now().UnixNano())
}
