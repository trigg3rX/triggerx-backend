package execution

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/trigg3rX/triggerx-backend/pkg/dockerexecutor/config"
	"github.com/trigg3rX/triggerx-backend/pkg/dockerexecutor/types"
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
	Close(ctx context.Context) error
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

func (ep *executionPipeline) execute(ctx context.Context, fileURL string, fileLanguage string, noOfAttesters int, metadata map[string]string) (*types.ExecutionResult, error) {
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

	// Initialize metadata if nil
	if metadata == nil {
		metadata = make(map[string]string)
	}

	// Create execution context
	executionContext := &types.ExecutionContext{
		FileURL:       fileURL,
		FileLanguage:  fileLanguage,
		NoOfAttesters: noOfAttesters,
		TraceID:       executionID,
		StartedAt:     startTime,
		Metadata:      metadata,
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

// executeSource accepts raw code and language, writes to a temp file, then runs through the same stages
func (ep *executionPipeline) executeSource(ctx context.Context, code string, language string, metadata map[string]string) (*types.ExecutionResult, error) {
	startTime := time.Now()
	executionID := generateExecutionID()

	ep.logger.Infof("Starting raw execution %s for language: %s", executionID, language)

	select {
	case <-ep.shutdownChan:
		return nil, fmt.Errorf("execution pipeline is shutting down")
	default:
	}

	execCtx, cancelFunc := context.WithCancel(ctx)
	defer cancelFunc()

	// Validate language before proceeding
	langType := types.Language(strings.ToLower(language))
	if !ep.containerMgr.IsLanguageSupported(langType) {
		return nil, fmt.Errorf("unsupported language: %s", language)
	}

	// Create a temporary file with appropriate extension
	var ext string
	switch langType {
	case types.LanguageGo:
		ext = ".go"
	case types.LanguagePy:
		ext = ".py"
	case types.LanguageJS, types.LanguageNode:
		ext = ".js"
	case types.LanguageTS:
		ext = ".ts"
	default:
		ext = ".go"
	}

	tmpFile, err := os.CreateTemp("", "tx-src-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPathNoExt := tmpFile.Name()
	_ = tmpFile.Close()
	tmpPath := tmpPathNoExt + ext
	// Use restrictive permissions (0600) to prevent world-readable access to sensitive code
	if err := os.WriteFile(tmpPath, []byte(code), 0600); err != nil {
		return nil, fmt.Errorf("failed to write temp source: %w", err)
	}
	defer func() { _ = os.Remove(tmpPath) }()

	// Initialize metadata if nil and add file path
	if metadata == nil {
		metadata = make(map[string]string)
	}
	metadata["file_path"] = tmpPath

	// Build a minimal execution context compatible with executeStages
	executionContext := &types.ExecutionContext{
		FileURL:       "inline://source",
		FileLanguage:  language,
		NoOfAttesters: 1,
		TraceID:       executionID,
		StartedAt:     startTime,
		Metadata:      metadata,
		State: types.ExecutionState{
			CancelFunc: cancelFunc,
		},
	}

	ep.activeExecutionsWG.Add(1)
	defer ep.activeExecutionsWG.Done()

	ep.mutex.Lock()
	ep.activeExecutions[executionID] = executionContext
	ep.mutex.Unlock()
	defer func() {
		ep.mutex.Lock()
		delete(ep.activeExecutions, executionID)
		ep.mutex.Unlock()
		duration := time.Since(startTime)
		ep.updateStats(true, duration, 0.0)
	}()

	result, err := ep.executeStages(execCtx, executionContext)
	if err != nil {
		executionContext.CompletedAt = time.Now()
		ep.updateStats(false, time.Since(startTime), 0.0)
		return nil, fmt.Errorf("execution failed: %w", err)
	}
	executionContext.CompletedAt = time.Now()
	return result, nil
}

func (ep *executionPipeline) executeStages(ctx context.Context, execCtx *types.ExecutionContext) (*types.ExecutionResult, error) {
	// If task_definition_id is 2, 4, or 6, skip to Stage 4 (Process Results, e.g., fee calculation)
	if taskDefStr, ok := execCtx.Metadata["task_definition_id"]; ok {
		var taskDefinitionID int
		if _, err := fmt.Sscanf(taskDefStr, "%d", &taskDefinitionID); err == nil {
			if taskDefinitionID == 1 || taskDefinitionID == 3 || taskDefinitionID == 5 || taskDefinitionID == 7 {
				ep.logger.Debugf("Skipping to Stage 4: Only processing results for task_definition_id=%d", taskDefinitionID)
				// No execution result available, so construct a minimal ExecutionResult to allow fee calculation
				result := &types.ExecutionResult{
					Stats:   types.DockerResourceStats{},
					Output:  "",
					Success: true,
					Error:   nil,
				}
				finalResult := ep.processResults(result, execCtx)
				return finalResult, nil
			}
		}
	}
	// Stage 1: Prepare file (download/validate) unless already provided (inline code)
	var fileCtx *types.ExecutionContext
	if existingPath := execCtx.Metadata["file_path"]; existingPath != "" {
		ep.logger.Debugf("Stage 1: Using provided file path (inline source): %s", existingPath)
		fileCtx = execCtx
	} else {
		ep.logger.Debugf("Stage 1: Downloading and validating file")
		var err error
		fileCtx, err = ep.fileManager.GetOrDownload(ctx, execCtx.FileURL, execCtx.FileLanguage)
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

	// Extract arguments for dynamic tasks (2, 4, 6)
	if taskDefStr, ok := execCtx.Metadata["task_definition_id"]; ok {
		var taskDefinitionID int
		if _, err := fmt.Sscanf(taskDefStr, "%d", &taskDefinitionID); err == nil {
			if taskDefinitionID == 2 || taskDefinitionID == 4 || taskDefinitionID == 6 {
				// For dynamic tasks, the output is expected to be a JSON array of arguments
				// We store this in metadata to be used by calculateFees
				// Clean up output - sometimes it might have newlines or extra whitespace
				cleanOutput := strings.TrimSpace(result.Output)
				// Basic validation that it looks like a JSON array
				if strings.HasPrefix(cleanOutput, "[") && strings.HasSuffix(cleanOutput, "]") {
					execCtx.Metadata["on_chain_args"] = cleanOutput
				} else {
					// If it's not a JSON array, we might want to log a warning or handle it
					// For now, we'll just log it
					// ep.logger.Warnf("Dynamic task output does not look like a JSON array: %s", cleanOutput)
				}
			}
		}
	}

	// Calculate fees
	fees, currentFees := ep.calculateFees(execCtx)
	execCtx.Metadata["fees"] = fees.String()
	execCtx.Metadata["current_fees"] = currentFees.String()
	result.Stats.TotalCost = fees
	result.Stats.CurrentTotalCost = currentFees

	return result
}

func (ep *executionPipeline) calculateFees(execCtx *types.ExecutionContext) (*big.Int, *big.Int) {
	feesConfig := ep.config.GetFeesConfig()

	// Get task definition ID from metadata
	var taskDefinitionID int
	if taskDefStr, ok := execCtx.Metadata["task_definition_id"]; ok {
		if parsed, err := fmt.Sscanf(taskDefStr, "%d", &taskDefinitionID); err != nil || parsed != 1 {
			ep.logger.Warnf("Failed to parse task_definition_id: %s", taskDefStr)
			taskDefinitionID = 0
		}
	}

	// Calculate off-chain fees based on task definition ID
	var offChainFeeUSD float64
	switch taskDefinitionID {
	case 1, 3, 5:
		// Static tasks
		offChainFeeUSD = feesConfig.StaticOffChainFeeUSD
		ep.logger.Debugf("Using static off-chain fee: $%.6f USD", offChainFeeUSD)
	case 2, 4, 6:
		// Dynamic tasks
		offChainFeeUSD = feesConfig.DynamicOffChainFeeUSD
		ep.logger.Debugf("Using dynamic off-chain fee: $%.6f USD", offChainFeeUSD)
	case 7:
		// Custom script - only off-chain, no on-chain
		offChainFeeUSD = feesConfig.CustomScriptFeeUSD
		ep.logger.Debugf("Using custom script off-chain fee: $%.6f USD", offChainFeeUSD)
	default:
		// Fallback to old calculation for unknown task types
		ep.logger.Warnf("Unknown task_definition_id: %d, using legacy fee calculation", taskDefinitionID)
		return ep.calculateLegacyFees(execCtx, feesConfig), big.NewInt(0)
	}

	// Fetch ETH to USD conversion rate from CoinGecko API
	resp, err := http.Get("https://api.coingecko.com/api/v3/simple/price?ids=ethereum&vs_currencies=usd")
	if err != nil {
		ep.logger.Warnf("failed to fetch ETH-USD rate from CoinGecko: %v", err)
		return big.NewInt(0), big.NewInt(0)
	}
	defer resp.Body.Close()

	var coingeckoResp struct {
		Ethereum struct {
			USD float64 `json:"usd"`
		} `json:"ethereum"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&coingeckoResp); err != nil {
		ep.logger.Warnf("failed to decode CoinGecko response: %v", err)
		return big.NewInt(0), big.NewInt(0)
	}
	EthToUSDRate := coingeckoResp.Ethereum.USD

	// Convert off-chain fee from USD to Wei
	offChainFeeInEther := offChainFeeUSD / EthToUSDRate
	offChainFeeFloat := big.NewFloat(offChainFeeInEther)
	weiMultiplier := big.NewFloat(1e18) // 10^18
	offChainFeeFloat.Mul(offChainFeeFloat, weiMultiplier)
	offChainFeeWei, _ := offChainFeeFloat.Int(nil)

	// For task definition ID 7 (custom script), only return off-chain fees
	if taskDefinitionID == 7 {
		ep.logger.Debugf("Fee calculation for custom script (ID 7): offchain_fee_usd=%.6f, offchain_fee_wei=%s",
			offChainFeeUSD, offChainFeeWei.String())
		return offChainFeeWei, offChainFeeWei
	}

	// Initialize on-chain related fee variables to zero by default
	var onChainFeeWei = big.NewInt(0)
	var currentOnChainFeeWei = big.NewInt(0)
	var aggregatorOnChainFeeWei = big.NewInt(0)

	// For other task types (1-6), calculate on-chain fees via gas estimation
	if chainID, hasChain := execCtx.Metadata["target_chain_id"]; hasChain && chainID != "" {
		contractAddr := execCtx.Metadata["target_contract_address"]
		function := execCtx.Metadata["target_function"]
		contractABI := execCtx.Metadata["abi"]

		if contractAddr != "" && function != "" && contractABI != "" {
			// Parse arguments from metadata if available
			var args []interface{}
			if argsStr, ok := execCtx.Metadata["on_chain_args"]; ok && argsStr != "" {
				if err := json.Unmarshal([]byte(argsStr), &args); err != nil {
					ep.logger.Warnf("Failed to parse on_chain_args: %v, using empty args", err)
					args = []interface{}{}
				}
			} else {
				args = []interface{}{}
			}

			// Get from address if provided
			fromAddress := execCtx.Metadata["from_address"]

			// Estimate gas for the on-chain transaction
			ctx := context.Background()
			gasEstimator := NewGasEstimator(ep.logger)
			defer gasEstimator.Close()

			gasLimit, gasPrice, currentGasPrice, err := gasEstimator.EstimateGasForFunction(
				ctx,
				chainID,
				contractAddr,
				function,
				contractABI,
				args,
				fromAddress,
			)

			if err != nil {
				ep.logger.Warnf("Failed to estimate gas: %v, using default on-chain fee", err)
				// Use a default on-chain fee if estimation fails (e.g., 0.001 ETH)
				defaultOnChainFee := big.NewFloat(0.001)
				defaultOnChainFee.Mul(defaultOnChainFee, weiMultiplier)
				onChainFeeWei, _ = defaultOnChainFee.Int(nil)
			} else {
				// Calculate gas cost in Wei
				onChainFeeWei = gasEstimator.CalculateGasCostInWei(gasLimit, gasPrice)
				currentOnChainFeeWei = gasEstimator.CalculateGasCostInWei(gasLimit, currentGasPrice)
				ep.logger.Debugf("Gas estimation: gasLimit=%d, gasPrice=%s, currentGasPrice=%s, gasCost=%s Wei, currentGasCost=%s Wei",
					gasLimit, gasPrice.String(), currentGasPrice.String(), onChainFeeWei.String(), currentOnChainFeeWei.String())
			}
		} else {
			ep.logger.Warnf("Missing contract details for on-chain fee calculation")
			onChainFeeWei = big.NewInt(0)
			currentOnChainFeeWei = big.NewInt(0)
		}
	} else {
		ep.logger.Debugf("No chain ID provided, skipping on-chain fee calculation")
		onChainFeeWei = big.NewInt(0)
		currentOnChainFeeWei = big.NewInt(0)
	}

	// Aggregator onchain fee calculation (always added to total fee)
	// Gas used is static at 720000, gas price is fetched from Base chain, chosen according to our chain id
	const aggregatorGasUsed = uint64(720000)

	// Logic:
	// If target_chain_id is mainnet (e.g. 8453, or any set you want to consider mainnet-extendable), use Base mainnet (8453) for aggregator fee.
	// If target_chain_id is testnet (e.g. 84532), use Base testnet/Sepolia (84532) for aggregator fee.
	// This mapping can be extended as needed.
	// NOTE: Only chain IDs 8453 (mainnet) and 84532 (testnet) are considered for Base aggregator fee.
	// If unknown, default to testnet (84532, safer to overestimate on test).

	const (
		baseMainnetChainID = "8453"
		baseTestnetChainID = "84532"
	)
	targetChainID := execCtx.Metadata["target_chain_id"]
	var baseAggregatorFeeChainID string
	if targetChainID == "42161" || targetChainID == "8453" {
		baseAggregatorFeeChainID = baseMainnetChainID
	} else {
		baseAggregatorFeeChainID = baseTestnetChainID
	}

	ctx := context.Background()
	gasEstimator := NewGasEstimator(ep.logger)
	defer gasEstimator.Close()

	// Use eth client to fetch current gas price (not historical/cached)
	ethClient, err := gasEstimator.getOrCreateClient(ctx, baseAggregatorFeeChainID)
	var aggregatorGasPrice *big.Int
	if err != nil {
		ep.logger.Warnf("Failed to get eth client for aggregator gas estimation on chain %s: %v, using default", baseAggregatorFeeChainID, err)
		aggregatorGasPrice = big.NewInt(1000000000) // 1 gwei fallback
	} else {
		aggregatorGasPrice, err = ethClient.SuggestGasPrice(ctx)
		if err != nil {
			ep.logger.Warnf("Failed to get current gas price for aggregator on chain %s: %v, using default", baseAggregatorFeeChainID, err)
			aggregatorGasPrice = big.NewInt(1000000000)
		} else {
			ep.logger.Debugf("Current gas price for aggregator fee, chain %s: %s Wei", baseAggregatorFeeChainID, aggregatorGasPrice.String())
		}
	}

	aggregatorOnChainFeeWei = gasEstimator.CalculateGasCostInWei(aggregatorGasUsed, aggregatorGasPrice)

	// Log aggregator fee calculation
	ep.logger.Debugf("Aggregator on-chain fee: baseAggregatorFeeChainID=%s, gasUsed=%d, currentGasPrice=%s Wei, aggregatorFee=%s Wei",
		baseAggregatorFeeChainID, aggregatorGasUsed, aggregatorGasPrice.String(), aggregatorOnChainFeeWei.String())

	// Total fee = off-chain fee + on-chain fee + aggregator on-chain fee
	totalFeeWei := new(big.Int).Add(offChainFeeWei, onChainFeeWei)
	currentTotalFeeWei := new(big.Int).Add(offChainFeeWei, currentOnChainFeeWei)
	currentTotalFeeWei.Add(currentTotalFeeWei, aggregatorOnChainFeeWei)
	totalFeeWei.Add(totalFeeWei, aggregatorOnChainFeeWei)
	totalFeeWei.Mul(totalFeeWei, big.NewInt(120))
	totalFeeWei.Div(totalFeeWei, big.NewInt(100)) // 20% buffer

	ep.logger.Infof("Fee calculation: task_definition_id=%d, offchain_fee_usd=%.6f, offchain_fee_wei=%s, onchain_fee_wei=%s, current_onchain_fee_wei=%s, aggregator_onchain_fee_wei=%s, total_fee_wei=%s, current_total_fee_wei=%s",
		taskDefinitionID, offChainFeeUSD, offChainFeeWei.String(), onChainFeeWei.String(), currentOnChainFeeWei.String(), aggregatorOnChainFeeWei.String(), totalFeeWei.String(), currentTotalFeeWei.String())
	return totalFeeWei, currentTotalFeeWei
}

// calculateLegacyFees is the old fee calculation method for backward compatibility
func (ep *executionPipeline) calculateLegacyFees(execCtx *types.ExecutionContext, feesConfig config.ExecutionFeeConfig) *big.Int {
	var staticComplexity, dynamicComplexity float64

	// Try to get complexity from metadata first (set in processResults)
	if staticStr, ok := execCtx.Metadata["static_complexity"]; ok {
		if parsed, err := fmt.Sscanf(staticStr, "%f", &staticComplexity); err != nil || parsed != 1 {
			ep.logger.Warnf("Failed to parse static complexity: %s", staticStr)
			staticComplexity = 0.0
		}
	}

	if dynamicStr, ok := execCtx.Metadata["dynamic_complexity"]; ok {
		if parsed, err := fmt.Sscanf(dynamicStr, "%f", &dynamicComplexity); err != nil || parsed != 1 {
			ep.logger.Warnf("Failed to parse dynamic complexity: %s", dynamicStr)
			dynamicComplexity = 0.0
		}
	}

	// Calculate x = static_complexity * factor + dynamic_complexity * factor + transaction_cost
	x := (staticComplexity * feesConfig.StaticComplexityFactor) +
		(dynamicComplexity * feesConfig.DynamicComplexityFactor) +
		feesConfig.TransactionCost

	// Calculate fee = [(0.1% of x) + x] TG
	// 0.1% = 0.001
	feeInTG := (feesConfig.FixedCost*x + x)

	// Convert TG to Ether using price per TG
	feeInEther := feeInTG * feesConfig.PricePerTG

	// Convert Ether to Wei (1 Ether = 10^18 Wei)
	// Use big.Float for precision
	feeFloat := big.NewFloat(feeInEther)
	weiMultiplier := big.NewFloat(1e18) // 10^18
	feeFloat.Mul(feeFloat, weiMultiplier)

	// Convert to big.Int (Wei)
	feeWei, _ := feeFloat.Int(nil)

	ep.logger.Debugf("Legacy fee calculation: static_complexity=%.6f, dynamic_complexity=%.6f, x=%.6f, fee_in_tg=%.6f, fee_in_ether=%.6f, fee_in_wei=%s",
		staticComplexity, dynamicComplexity, x, feeInTG, feeInEther, feeWei.String())

	return feeWei
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
	feesConfig := ep.config.GetFeesConfig()

	// Use the same formula as calculateFees but return as float64 for statistics
	// For statistics, we'll use a simplified version with just the complexity parameter
	// In a real scenario, you might want to pass both static and dynamic complexity separately

	// Calculate x = complexity * factor + transaction_cost
	// Note: This is a simplified version for statistics. The actual fee calculation
	// uses separate static and dynamic complexity values
	x := (complexity * feesConfig.StaticComplexityFactor) + feesConfig.TransactionCost

	// Calculate fee = [(0.1% of x) + x] TG
	// 0.1% = 0.001
	feeInTG := (0.001*x + x) // TG

	// Convert TG to Ether using price per TG
	feeInEther := feeInTG * feesConfig.PricePerTG

	return feeInEther
}

func generateExecutionID() string {
	return fmt.Sprintf("exec_%d", time.Now().UnixNano())
}
