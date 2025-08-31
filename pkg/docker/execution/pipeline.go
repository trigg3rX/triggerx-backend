package execution

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/trigg3rX/triggerx-backend/pkg/docker/config"
	"github.com/trigg3rX/triggerx-backend/pkg/docker/types"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

// ContainerManager defines what the execution pipeline needs from a container manager
type ContainerManager interface {
	GetContainer(ctx context.Context, language types.Language) (*types.PooledContainer, error)
	ReturnContainer(container *types.PooledContainer) error
	ExecuteInContainer(ctx context.Context, containerID string, filePath string, language types.Language) (*types.ExecutionResult, string, error)
	MarkContainerAsFailed(containerID string, language types.Language, err error)
	KillExecProcess(ctx context.Context, execID string) error
	GetPoolStats() map[types.Language]*types.PoolStats
	InitializeLanguagePools(ctx context.Context, languages []types.Language) error
	GetSupportedLanguages() []types.Language
	IsLanguageSupported(language types.Language) bool
	Close() error
}

// FileManager defines what the execution pipeline needs from a file manager
type FileManager interface {
	GetOrDownload(ctx context.Context, fileURL string, fileLanguage string) (*types.ExecutionContext, error)
	Close() error
}

type executionPipeline struct {
	fileManager        FileManager
	containerMgr       ContainerManager
	config             config.ConfigProviderInterface
	logger             logging.Logger
	mutex              sync.RWMutex
	activeExecutions   map[string]*types.ExecutionContext
	stats              *types.PerformanceMetrics
	activeExecutionsWG sync.WaitGroup // Track active executions for graceful shutdown
	shutdownChan       chan struct{}  // Signal for shutdown
	closed             bool
}

func newExecutionPipeline(cfg config.ConfigProviderInterface, fileMgr FileManager, containerMgr ContainerManager, logger logging.Logger) *executionPipeline {
	return &executionPipeline{
		fileManager:      fileMgr,
		containerMgr:     containerMgr,
		config:           cfg,
		logger:           logger,
		activeExecutions: make(map[string]*types.ExecutionContext),
		shutdownChan:     make(chan struct{}),
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

func (ep *executionPipeline) execute(ctx context.Context, fileURL string, fileLanguage string, noOfAttesters int) (*types.ExecutionResult, error) {
	startTime := time.Now()
	executionID := generateExecutionID()

	ep.logger.Infof("Starting execution %s for file: %s", executionID, fileURL)

	// Check if pipeline is shutting down
	select {
	case <-ep.shutdownChan:
		return nil, fmt.Errorf("execution pipeline is shutting down")
	default:
	}

	// Create a cancellable context for this execution
	execCtx, cancelFunc := context.WithCancel(ctx)
	defer cancelFunc() // Ensure cleanup

	// Create execution context
	executionContext := &types.ExecutionContext{
		FileURL:       fileURL,
		FileLanguage:  fileLanguage,
		NoOfAttesters: noOfAttesters,
		TraceID:       executionID,
		StartedAt:     startTime,
		Metadata:      make(map[string]string),
		State: types.ExecutionState{
			CancelFunc: cancelFunc,
		},
	}

	// Track execution with WaitGroup for graceful shutdown
	ep.activeExecutionsWG.Add(1)
	defer ep.activeExecutionsWG.Done()

	// Track execution
	ep.mutex.Lock()
	ep.activeExecutions[executionID] = executionContext
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
	result, err := ep.executeStages(execCtx, executionContext)
	if err != nil {
		executionContext.CompletedAt = time.Now()
		ep.updateStats(false, time.Since(startTime), 0.0)
		return nil, fmt.Errorf("execution failed: %w", err)
	}

	executionContext.CompletedAt = time.Now()
	duration := time.Since(startTime)

	ep.logger.Infof("Execution %s completed successfully in %v", executionID, duration)
	return result, nil
}

func (ep *executionPipeline) executeStages(ctx context.Context, execCtx *types.ExecutionContext) (*types.ExecutionResult, error) {
	// Stage 1: Download and Validate
	ep.logger.Debugf("Stage 1: Downloading and validating file")
	fileCtx, err := ep.fileManager.GetOrDownload(ctx, execCtx.FileURL, execCtx.FileLanguage)
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

	// Determine language from file extension
	filePath := fileCtx.Metadata["file_path"]
	if filePath == "" {
		return nil, fmt.Errorf("file path not found in execution context")
	}

	language := types.GetLanguageFromFile(filePath)
	ep.logger.Debugf("Detected language: %s for file: %s", language, filePath)

	container, err := ep.containerMgr.GetContainer(ctx, language)
	if err != nil {
		return nil, fmt.Errorf("failed to get container: %w", err)
	}

	ep.logger.Debugf("Got container %s from %s pool", container.ID, container.Language)

	// Return container to pool synchronously for proper cleanup during shutdown
	defer func() {
		ep.logger.Debugf("Returning container %s to pool (sync)", container.ID)
		if err := ep.containerMgr.ReturnContainer(container); err != nil {
			ep.logger.Warnf("Failed to return container to pool: %v", err)
		}
	}()

	// Stage 3: Execute Code
	ep.logger.Debugf("Stage 3: Executing code in container %s", container.ID)
	result, execID, err := ep.containerMgr.ExecuteInContainer(ctx, container.ID, filePath, container.Language)
	if err != nil {
		// Mark container as failed if execution fails
		ep.logger.Warnf("Execution failed in container %s, marking as failed: %v", container.ID, err)
		ep.containerMgr.MarkContainerAsFailed(container.ID, container.Language, err)
		return nil, fmt.Errorf("failed to execute code: %w", err)
	}

	// Store exec ID and container ID for potential cancellation
	execCtx.State.ExecID = execID
	execCtx.State.ContainerID = container.ID

	// Check if execution was successful
	if !result.Success {
		// Mark container as failed if execution returned non-zero exit code
		ep.logger.Warnf("Execution failed in container %s with error: %v", container.ID, result.Error)
		ep.containerMgr.MarkContainerAsFailed(container.ID, container.Language, result.Error)
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

func (ep *executionPipeline) processResults(result *types.ExecutionResult, execCtx *types.ExecutionContext) *types.ExecutionResult {
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

func (ep *executionPipeline) calculateFees(execCtx *types.ExecutionContext) float64 {
	// Basic fee calculation
	duration := time.Since(execCtx.StartedAt)
	baseFee := ep.config.GetFeesConfig().FixedCost
	timeFee := duration.Seconds() * ep.config.GetFeesConfig().PricePerTG

	// Add complexity factor if available
	complexityFee := 0.0
	if _, ok := execCtx.Metadata["complexity"]; ok {
		// Parse complexity and apply factor
		// This is a simplified calculation
		complexityFee = 0.1 // Placeholder
	}

	return baseFee + timeFee + complexityFee
}

// func (ep *executionPipeline) cleanupExecution(execCtx *types.ExecutionContext) error {
// 	// Cleanup any temporary files
// 	// In this implementation, the file manager handles cleanup
// 	return nil
// }

func (ep *executionPipeline) getActiveExecutions() []*types.ExecutionContext {
	ep.mutex.RLock()
	defer ep.mutex.RUnlock()

	executions := make([]*types.ExecutionContext, 0, len(ep.activeExecutions))
	for _, exec := range ep.activeExecutions {
		executions = append(executions, exec)
	}

	return executions
}

// Close gracefully shuts down the execution pipeline
func (ep *executionPipeline) close() error {
	ep.mutex.Lock()
	defer ep.mutex.Unlock()

	if ep.closed {
		return nil
	}

	ep.logger.Info("Closing execution pipeline")

	// Signal shutdown to prevent new executions
	close(ep.shutdownChan)
	ep.closed = true

	// Cancel all active executions
	for executionID, execCtx := range ep.activeExecutions {
		ep.logger.Infof("Cancelling active execution: %s", executionID)
		if execCtx.State.CancelFunc != nil {
			execCtx.State.CancelFunc()
		}
	}

	// Wait for all active executions to complete with timeout
	done := make(chan struct{})
	go func() {
		ep.activeExecutionsWG.Wait()
		close(done)
	}()

	select {
	case <-done:
		ep.logger.Info("All active executions completed")
	case <-time.After(30 * time.Second):
		ep.logger.Warn("Timeout waiting for active executions to complete")
	}

	ep.logger.Info("Execution pipeline closed")
	return nil
}

func (ep *executionPipeline) cancelExecution(executionID string) error {
	ep.mutex.Lock()
	defer ep.mutex.Unlock()

	exec, exists := ep.activeExecutions[executionID]
	if !exists {
		return fmt.Errorf("execution not found: %s", executionID)
	}

	// Actually cancel the execution by calling the cancel function
	if exec.State.CancelFunc != nil {
		exec.State.CancelFunc()
		ep.logger.Infof("Execution %s cancelled - context cancellation will terminate Docker processes", executionID)
	} else {
		ep.logger.Warnf("Execution %s has no cancel function - marking as cancelled", executionID)
	}

	// Attempt to terminate the Docker exec process if we have the exec ID
	if exec.State.ExecID != "" {
		ep.logger.Infof("Attempting to terminate Docker exec process %s for execution %s", exec.State.ExecID, executionID)
		if err := ep.containerMgr.KillExecProcess(context.Background(), exec.State.ExecID); err != nil {
			ep.logger.Warnf("Failed to terminate exec process %s: %v", exec.State.ExecID, err)
		}
	}

	exec.CompletedAt = time.Now()

	ep.logger.Infof("Execution %s cancelled", executionID)
	return nil
}

func (ep *executionPipeline) getStats() *types.PerformanceMetrics {
	ep.mutex.RLock()
	defer ep.mutex.RUnlock()

	// Create a copy to avoid race conditions
	stats := *ep.stats
	return &stats
}

func (ep *executionPipeline) updateStats(success bool, duration time.Duration, complexity float64) {
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

func (ep *executionPipeline) calculateCost(duration time.Duration, complexity float64) float64 {
	// Basic cost calculation based on execution time and complexity
	timeCost := duration.Seconds() * ep.config.GetFeesConfig().PricePerTG
	complexityCost := complexity * ep.config.GetFeesConfig().PricePerTG
	return timeCost + complexityCost + ep.config.GetFeesConfig().FixedCost
}

func generateExecutionID() string {
	return fmt.Sprintf("exec_%d", time.Now().UnixNano())
}
