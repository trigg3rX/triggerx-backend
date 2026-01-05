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

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/trigg3rX/triggerx-backend/pkg/dockerexecutor/config"
	"github.com/trigg3rX/triggerx-backend/pkg/dockerexecutor/types"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
)

// ContainerManager defines what the execution pipeline needs from a container manager
type ContainerManager interface {
	GetContainer(ctx context.Context, language types.Language) (*types.PooledContainer, error)
	ReturnContainer(ctx context.Context, container *types.PooledContainer) error
	ExecuteInContainer(ctx context.Context, containerID string, filePath string, language types.Language) (*types.ExecutionResult, string, error)
	MarkContainerAsFailed(ctx context.Context, containerID string, language types.Language, err error)
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
	logger             observability.Logger
	mutex              sync.RWMutex
	activeExecutions   map[string]*types.ExecutionContext
	stats              *types.PerformanceMetrics
	activeExecutionsWG sync.WaitGroup // Track active executions for graceful shutdown
	shutdownChan       chan struct{}  // Signal for shutdown
	closed             bool
	gasEstimator       *GasEstimator // Shared gas estimator with 7-day cache
}

func newExecutionPipeline(cfg config.ConfigProviderInterface, fileMgr FileManager, containerMgr ContainerManager, logger observability.Logger) *executionPipeline {
	return &executionPipeline{
		fileManager:      fileMgr,
		containerMgr:     containerMgr,
		config:           cfg,
		logger:           logger,
		activeExecutions: make(map[string]*types.ExecutionContext),
		shutdownChan:     make(chan struct{}),
		gasEstimator:     NewGasEstimator(logger), // Shared gas estimator instance
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

func (ep *executionPipeline) execute(ctx context.Context, fileURL string, fileLanguage string, noOfAttesters int, alchemyAPIKey string, metadata map[string]string) (*types.ExecutionResult, error) {
	startTime := time.Now()
	executionID := generateExecutionID()

	ep.logger.Debug(ctx, "Starting execution", observability.String("executionID", executionID), observability.String("fileURL", fileURL))

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
	result, err := ep.executeStages(execCtx, executionContext, alchemyAPIKey)
	if err != nil {
		executionContext.CompletedAt = time.Now()
		ep.updateStats(false, time.Since(startTime), 0.0)
		return nil, fmt.Errorf("execution failed: %w", err)
	}

	executionContext.CompletedAt = time.Now()
	duration := time.Since(startTime)

	ep.logger.Debug(ctx, "Execution completed successfully", observability.String("executionID", executionID), observability.Duration("duration", duration))
	return result, nil
}

// executeSource accepts raw code and language, writes to a temp file, then runs through the same stages
func (ep *executionPipeline) executeSource(ctx context.Context, code string, language string, alchemyAPIKey string, metadata map[string]string) (*types.ExecutionResult, error) {
	startTime := time.Now()
	executionID := generateExecutionID()

	ep.logger.Debug(ctx, "Starting raw execution", observability.String("executionID", executionID), observability.String("language", language))

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

	result, err := ep.executeStages(execCtx, executionContext, alchemyAPIKey)
	if err != nil {
		executionContext.CompletedAt = time.Now()
		ep.updateStats(false, time.Since(startTime), 0.0)
		return nil, fmt.Errorf("execution failed: %w", err)
	}
	executionContext.CompletedAt = time.Now()
	return result, nil
}

func (ep *executionPipeline) executeStages(ctx context.Context, execCtx *types.ExecutionContext, alchemyAPIKey string) (*types.ExecutionResult, error) {
	// If task_definition_id is 2, 4, or 6, skip to Stage 4 (Process Results, e.g., fee calculation)
	if taskDefStr, ok := execCtx.Metadata["task_definition_id"]; ok {
		var taskDefinitionID int
		if _, err := fmt.Sscanf(taskDefStr, "%d", &taskDefinitionID); err == nil {
			if (taskDefinitionID == 1 || taskDefinitionID == 3 || taskDefinitionID == 5 || taskDefinitionID == 7) && execCtx.FileURL == "" {
				ep.logger.Debug(ctx, "Skipping to Stage 4: Only processing results for task_definition_id", observability.Int("taskDefinitionID", taskDefinitionID))
				// No execution result available, so construct a minimal ExecutionResult to allow fee calculation
				result := &types.ExecutionResult{
					Stats:   types.DockerResourceStats{},
					Output:  "",
					Success: true,
					Error:   nil,
				}
				finalResult := ep.processResults(ctx, result, execCtx, alchemyAPIKey)
				return finalResult, nil
			}
		}
	}
	// Stage 1: Prepare file (download/validate) unless already provided (inline code)
	var fileCtx *types.ExecutionContext
	if existingPath := execCtx.Metadata["file_path"]; existingPath != "" {
		ep.logger.Debug(ctx, "Stage 1: Using provided file path (inline source)", observability.String("existingPath", existingPath))
		fileCtx = execCtx
	} else {
		ep.logger.Debug(ctx, "Stage 1: Downloading and validating file")
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
	ep.logger.Debug(ctx, "Stage 2: Getting container from pool")

	// Determine language from file extension
	filePath := fileCtx.Metadata["file_path"]
	if filePath == "" {
		return nil, fmt.Errorf("file path not found in execution context")
	}

	language := types.GetLanguageFromFile(filePath)
	ep.logger.Debug(ctx, "Detected language", observability.String("language", string(language)), observability.String("filePath", filePath))

	container, err := ep.containerMgr.GetContainer(ctx, language)
	if err != nil {
		return nil, fmt.Errorf("failed to get container: %w", err)
	}

	ep.logger.Debug(ctx, "Got container from pool", observability.String("containerID", container.ID), observability.String("containerLanguage", string(container.Language)))

	// Return container to pool synchronously for proper cleanup during shutdown
	defer func() {
		ep.logger.Debug(ctx, "Returning container to pool (sync)", observability.String("containerID", container.ID))
		if err := ep.containerMgr.ReturnContainer(ctx, container); err != nil {
			ep.logger.Warn(ctx, "Failed to return container to pool", observability.String("containerID", container.ID), observability.String("language", string(container.Language)), observability.Error(err))
		}
	}()

	// Stage 3: Execute Code
	ep.logger.Debug(ctx, "Stage 3: Executing code in container", observability.String("containerID", container.ID))
	result, execID, err := ep.containerMgr.ExecuteInContainer(ctx, container.ID, filePath, container.Language)
	if err != nil {
		// Mark container as failed if execution fails
		ep.logger.Warn(ctx, "Execution failed in container, marking as failed", observability.String("containerID", container.ID), observability.String("filePath", filePath), observability.String("language", string(container.Language)), observability.Error(err))
		ep.containerMgr.MarkContainerAsFailed(ctx, container.ID, container.Language, err)
		return nil, fmt.Errorf("failed to execute code: %w", err)
	}

	// Store exec ID and container ID for potential cancellation
	execCtx.State.ExecID = execID
	execCtx.State.ContainerID = container.ID

	// Check if execution was successful
	if !result.Success {
		// Mark container as failed if execution returned non-zero exit code
		ep.logger.Warn(ctx, "Execution failed in container with error", observability.String("containerID", container.ID), observability.String("filePath", filePath), observability.String("language", string(container.Language)), observability.Error(result.Error))
		ep.containerMgr.MarkContainerAsFailed(ctx, container.ID, container.Language, result.Error)
	}

	// Stage 4: Process Results
	ep.logger.Debug(ctx, "Stage 4: Processing results")
	finalResult := ep.processResults(ctx, result, execCtx, alchemyAPIKey)

	// Stage 5: Cleanup
	// ep.logger.Debugf("Stage 5: Cleaning up")
	// if err := ep.cleanupExecution(execCtx); err != nil {
	// 	ep.logger.Warn(ctx, "Failed to cleanup execution", observability.Error(err))
	// }

	return finalResult, nil
}

func (ep *executionPipeline) processResults(ctx context.Context, result *types.ExecutionResult, execCtx *types.ExecutionContext, alchemyAPIKey string) *types.ExecutionResult {
	// Add execution metadata
	execCtx.Metadata["execution_time"] = result.Stats.ExecutionTime.String()
	execCtx.Metadata["static_complexity"] = fmt.Sprintf("%.6f", result.Stats.StaticComplexity)
	execCtx.Metadata["dynamic_complexity"] = fmt.Sprintf("%.6f", result.Stats.DynamicComplexity)

	// Extract arguments for dynamic tasks (2, 4, 6) and custom scripts (7)
	if taskDefStr, ok := execCtx.Metadata["task_definition_id"]; ok {
		var taskDefinitionID int
		if _, err := fmt.Sscanf(taskDefStr, "%d", &taskDefinitionID); err == nil {
			ep.logger.Debug(ctx, "task defination id in process", observability.Int("taskDefinitionID", taskDefinitionID))
			switch taskDefinitionID {
			case 2, 4, 6:
				// For dynamic tasks, the output is expected to be a JSON array of arguments
				// We store this in metadata to be used by calculateFees
				// Clean up output - sometimes it might have newlines or extra whitespace
				cleanOutput := strings.TrimSpace(result.Output)
				// Basic validation that it looks like a JSON array
				if strings.HasPrefix(cleanOutput, "[") && strings.HasSuffix(cleanOutput, "]") {
					execCtx.Metadata["on_chain_args"] = cleanOutput
					// } else {
					// If it's not a JSON array, we might want to log a warning or handle it
					// For now, we'll just log it
					// ep.logger.Warn(ctx, "Dynamic task output does not look like a JSON array", observability.String("cleanOutput", cleanOutput))
				}
			case 7:
				// For custom scripts, parse the JSON output to extract targetContract and calldata
				cleanOutput := strings.TrimSpace(result.Output)
				if cleanOutput != "" && strings.HasPrefix(cleanOutput, "{") {
					var scriptOutput struct {
						ShouldExecute  bool   `json:"shouldExecute"`
						TargetContract string `json:"targetContract"`
						Calldata       string `json:"calldata"`
					}
					if err := json.Unmarshal([]byte(cleanOutput), &scriptOutput); err == nil {
						if scriptOutput.ShouldExecute && scriptOutput.TargetContract != "" && scriptOutput.Calldata != "" {
							execCtx.Metadata["script_target_contract"] = scriptOutput.TargetContract
							execCtx.Metadata["script_calldata"] = scriptOutput.Calldata
							ep.logger.Debug(ctx, "Custom script output", observability.String("targetContract", scriptOutput.TargetContract), observability.String("calldata", scriptOutput.Calldata[:min(len(scriptOutput.Calldata), 66)]))
						}
					} else {
						ep.logger.Warn(ctx, "Failed to parse custom script output", observability.String("cleanOutput", cleanOutput[:min(len(cleanOutput), 200)]), observability.Error(err))
					}
				}
			}
		}
	}

	// Calculate fees (pass result to check for execution failures)
	fees, currentFees := ep.calculateFees(ctx, execCtx, result, alchemyAPIKey)
	execCtx.Metadata["fees"] = fees.String()
	execCtx.Metadata["current_fees"] = currentFees.String()
	result.Stats.TotalCost = fees
	result.Stats.CurrentTotalCost = currentFees

	return result
}

// getChainlinkETHUSDPrice fetches ETH/USD price from Chainlink oracle on Arbitrum One
// Returns price in USD (float64) and error
// We use Arbitrum One as a single source of truth since ETH/USD price is the same across all chains
func (ep *executionPipeline) getChainlinkETHUSDPrice(ctx context.Context, alchemyAPIKey string) (float64, error) {
	// Always use Arbitrum One's Chainlink ETH/USD price feed
	const arbitrumOneChainID = "42161"
	const arbitrumOnePriceFeedAddr = "0xb2A824043730FE05F3DA2efaFa1CBbe83fa548D6" // Arbitrum One ETH/USD

	// Get Ethereum client for Arbitrum One
	client, err := ep.gasEstimator.getOrCreateClient(ctx, arbitrumOneChainID, alchemyAPIKey)
	if err != nil {
		return 0, fmt.Errorf("failed to get eth client for Arbitrum One (chain %s): %w", arbitrumOneChainID, err)
	}

	// Chainlink Aggregator V3 ABI for latestRoundData()
	// function latestRoundData() external view returns (
	//     uint80 roundId,
	//     int256 answer,
	//     uint256 startedAt,
	//     uint256 updatedAt,
	//     uint80 answeredInRound
	// )
	// The answer is in 8 decimals (e.g., 300000000000 = $3000)
	priceFeedABI := `[{"inputs":[],"name":"latestRoundData","outputs":[{"internalType":"uint80","name":"roundId","type":"uint80"},{"internalType":"int256","name":"answer","type":"int256"},{"internalType":"uint256","name":"startedAt","type":"uint256"},{"internalType":"uint256","name":"updatedAt","type":"uint256"},{"internalType":"uint80","name":"answeredInRound","type":"uint80"}],"stateMutability":"view","type":"function"}]`

	// Parse ABI
	abiParsed, err := abi.JSON(strings.NewReader(priceFeedABI))
	if err != nil {
		return 0, fmt.Errorf("failed to parse Chainlink ABI: %w", err)
	}

	// Pack the function call
	callData, err := abiParsed.Pack("latestRoundData")
	if err != nil {
		return 0, fmt.Errorf("failed to pack latestRoundData call: %w", err)
	}

	// Call the contract
	priceFeedAddress := common.HexToAddress(arbitrumOnePriceFeedAddr)
	result, err := client.CallContract(ctx, ethereum.CallMsg{
		To:   &priceFeedAddress,
		Data: callData,
	}, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to call Chainlink contract on Arbitrum One: %w", err)
	}

	// Unpack the result
	// Unpack returns (roundId, answer, startedAt, updatedAt, answeredInRound)
	outputs, err := abiParsed.Unpack("latestRoundData", result)
	if err != nil {
		return 0, fmt.Errorf("failed to unpack Chainlink response: %w", err)
	}

	// Extract values from outputs
	if len(outputs) < 5 {
		return 0, fmt.Errorf("invalid Chainlink response: expected 5 values, got %d", len(outputs))
	}

	// uint80 is represented as *big.Int in Go
	var answer *big.Int
	var updatedAt *big.Int

	// Extract answer (index 1) and updatedAt (index 3)
	if answerBig, ok := outputs[1].(*big.Int); ok {
		answer = answerBig
	} else {
		return 0, fmt.Errorf("invalid answer type: expected *big.Int")
	}

	if updatedAtBig, ok := outputs[3].(*big.Int); ok {
		updatedAt = updatedAtBig
	}

	// Check if answer is valid (not zero or negative)
	if answer == nil || answer.Sign() <= 0 {
		return 0, fmt.Errorf("invalid Chainlink price: answer is zero or negative")
	}

	// Check if price is stale (updated more than 1 hour ago)
	if updatedAt != nil {
		now := time.Now().Unix()
		updatedAtUnix := updatedAt.Int64()
		if updatedAtUnix > 0 && now-updatedAtUnix > 3600 {
			ep.logger.Warn(ctx, "Chainlink price feed is stale", observability.Int64("updatedAtUnix", now-updatedAtUnix))
		}
	}

	// Convert answer from 8 decimals to USD price
	// answer is in 8 decimals, so divide by 1e8
	priceFloat := new(big.Float).SetInt(answer)
	decimals := new(big.Float).SetInt(big.NewInt(1e8))
	priceFloat.Quo(priceFloat, decimals)
	price, _ := priceFloat.Float64()

	ep.logger.Debug(ctx, "Fetched ETH/USD price from Chainlink on Arbitrum One", observability.String("arbitrumOnePriceFeedAddr", arbitrumOnePriceFeedAddr), observability.Float64("price", price))
	return price, nil
}

func (ep *executionPipeline) calculateFees(ctx context.Context, execCtx *types.ExecutionContext, result *types.ExecutionResult, alchemyAPIKey string) (*big.Int, *big.Int) {
	feesConfig := ep.config.GetFeesConfig()

	// Get task definition ID from metadata
	var taskDefinitionID int
	if taskDefStr, ok := execCtx.Metadata["task_definition_id"]; ok {
		if parsed, err := fmt.Sscanf(taskDefStr, "%d", &taskDefinitionID); err != nil || parsed != 1 {
			ep.logger.Warn(ctx, "Failed to parse task_definition_id", observability.String("taskDefStr", taskDefStr), observability.Error(err))
			taskDefinitionID = 0
		}
	}

	ep.logger.Debug(ctx, "task defination id in the calculate fees", observability.Int("taskDefinitionID", taskDefinitionID))

	// Calculate off-chain fees based on task definition ID
	var offChainFeeUSD float64
	switch taskDefinitionID {
	case 1, 3, 5:
		// Static tasks
		offChainFeeUSD = feesConfig.StaticOffChainFeeUSD
		ep.logger.Debug(ctx, "Using static off-chain fee", observability.Float64("offChainFeeUSD", offChainFeeUSD))
	case 2, 4, 6:
		// Dynamic tasks
		offChainFeeUSD = feesConfig.DynamicOffChainFeeUSD
		ep.logger.Debug(ctx, "Using dynamic off-chain fee", observability.Float64("offChainFeeUSD", offChainFeeUSD))
	case 7:
		// Custom script - uses dynamic off-chain fee (same as 2, 4, 6)
		offChainFeeUSD = feesConfig.DynamicOffChainFeeUSD
		ep.logger.Debug(ctx, "Using dynamic off-chain fee for custom script", observability.Float64("offChainFeeUSD", offChainFeeUSD))
	default:
		// Fallback to old calculation for unknown task types
		ep.logger.Warn(ctx, "Unknown task_definition_id, using legacy fee calculation", observability.Int("taskDefinitionID", taskDefinitionID))
		return ep.calculateLegacyFees(ctx, execCtx, feesConfig), big.NewInt(0)
	}

	// Fetch ETH to USD conversion rate from CoinGecko API
	coingeckoURL := "https://api.coingecko.com/api/v3/simple/price?ids=ethereum&vs_currencies=usd"
	resp, err := http.Get(coingeckoURL)
	if err != nil {
		ep.logger.Warn(ctx, "failed to fetch ETH-USD rate from CoinGecko", observability.String("url", coingeckoURL), observability.Int("taskDefinitionID", taskDefinitionID), observability.Error(err))
		return big.NewInt(0), big.NewInt(0)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			ep.logger.Error(ctx, "Error closing response body", observability.String("url", "https://api.coingecko.com/api/v3/simple/price?ids=ethereum&vs_currencies=usd"), observability.Error(err))
		}
	}()

	var coingeckoResp struct {
		Ethereum struct {
			USD float64 `json:"usd"`
		} `json:"ethereum"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&coingeckoResp); err != nil {
		ep.logger.Warn(ctx, "failed to decode CoinGecko response", observability.String("url", coingeckoURL), observability.Int("statusCode", resp.StatusCode), observability.Int("taskDefinitionID", taskDefinitionID), observability.Error(err))
		return big.NewInt(0), big.NewInt(0)
	}
	EthToUSDRate := coingeckoResp.Ethereum.USD

	// Check if ETH/USD rate is valid (not zero or negative)
	if EthToUSDRate <= 0 {
		ep.logger.Warn(ctx, "invalid ETH/USD rate from CoinGecko, fetching from Chainlink oracle on Arbitrum One", observability.Float64("EthToUSDRate", EthToUSDRate))
		// Fetch from Chainlink oracle on Arbitrum One as fallback
		ctx := context.Background()
		chainlinkPrice, err := ep.getChainlinkETHUSDPrice(ctx, alchemyAPIKey)
		if err != nil {
			ep.logger.Warn(ctx, "failed to fetch ETH/USD rate from Chainlink on Arbitrum One", observability.Error(err))
			return big.NewInt(0), big.NewInt(0)
		}
		EthToUSDRate = chainlinkPrice
		ep.logger.Debug(ctx, "Using Chainlink ETH/USD rate for Arbitrum One", observability.Float64("EthToUSDRate", EthToUSDRate))
	}

	// Convert off-chain fee from USD to Wei
	offChainFeeInEther := offChainFeeUSD / EthToUSDRate
	offChainFeeFloat := big.NewFloat(offChainFeeInEther)
	weiMultiplier := big.NewFloat(1e18) // 10^18
	offChainFeeFloat.Mul(offChainFeeFloat, weiMultiplier)
	offChainFeeWei, _ := offChainFeeFloat.Int(nil)

	// Safety check: ensure offChainFeeWei is not nil (could happen if float is Inf or NaN)
	if offChainFeeWei == nil {
		ep.logger.Warn(ctx, "offChainFeeWei conversion resulted in nil, using $3000 USD fallback", observability.Float64("offChainFeeUSD", offChainFeeUSD), observability.Float64("EthToUSDRate", EthToUSDRate), observability.Float64("offChainFeeInEther", offChainFeeInEther))
		// Use $3000 USD as fallback
		fallbackFeeInEther := 3000.0 / EthToUSDRate
		fallbackFeeFloat := big.NewFloat(fallbackFeeInEther)
		weiMultiplierFallback := big.NewFloat(1e18)
		fallbackFeeFloat.Mul(fallbackFeeFloat, weiMultiplierFallback)
		offChainFeeWei, _ = fallbackFeeFloat.Int(nil)
	}

	// For task definition ID 7 (custom script), calculate on-chain fees using fixed 1M gas
	// During actual execution (in keeper), real gas estimation with actual calldata is used
	if taskDefinitionID == 7 {
		var onChainFeeWei = big.NewInt(0)
		var currentOnChainFeeWei = big.NewInt(0)
		var aggregatorOnChainFeeWei = big.NewInt(0)

		chainID := execCtx.Metadata["target_chain_id"]
		ctx := context.Background()

		if chainID != "" {
			// Check if we have real calldata from script execution
			scriptTargetContract := execCtx.Metadata["script_target_contract"]
			scriptCalldata := execCtx.Metadata["script_calldata"]
			fromAddress := execCtx.Metadata["from_address"]

			var gasLimit uint64
			var gasPrice, currentGasPrice *big.Int

			if scriptTargetContract != "" && scriptCalldata != "" {
				// EXECUTION: Use real gas estimation with actual calldata
				ep.logger.Debug(ctx, "Custom script: using real calldata for gas estimation")

				var err error
				gasLimit, gasPrice, currentGasPrice, err = ep.gasEstimator.EstimateGasWithCalldata(
					ctx,
					chainID,
					scriptTargetContract,
					common.FromHex(scriptCalldata), // Properly decode hex string to bytes
					fromAddress,
					alchemyAPIKey,
				)
				if err != nil {
					ep.logger.Warn(ctx, "Failed to estimate gas with calldata, using fallback 600K gas", observability.String("chainID", chainID), observability.String("scriptTargetContract", scriptTargetContract), observability.String("scriptCalldata", scriptCalldata[:min(len(scriptCalldata), 100)]+"..."), observability.String("fromAddress", fromAddress), observability.Error(err))
					gasLimit = 600000
					gasPrice = big.NewInt(1000000000)
					currentGasPrice = gasPrice
				}

				// Add 20% buffer to gas limit (for execution safety margin)
				gasLimit = gasLimit * 120 / 100

				ep.logger.Debug(ctx, "Custom script execution", observability.Uint64("gasLimit", gasLimit), observability.String("gasPrice", gasPrice.String()), observability.String("currentGasPrice", currentGasPrice.String()))
			} else {
				// JOB CREATION: Use fixed 1M gas (conservative estimate)
				gasLimit = uint64(1000000)

				var err error
				gasPrice, err = ep.gasEstimator.GetGasPrice(ctx, chainID, alchemyAPIKey)
				if err != nil {
					ep.logger.Warn(ctx, "Failed to get gas price for custom script, using default", observability.String("chainID", chainID), observability.Error(err))
					gasPrice = big.NewInt(1000000000) // 1 gwei fallback
				}

				currentGasPrice, err = ep.gasEstimator.GetCurrentGasPrice(ctx, chainID, alchemyAPIKey)
				if err != nil {
					ep.logger.Warn(ctx, "Failed to get current gas price, using historical price", observability.String("chainID", chainID), observability.Error(err))
					currentGasPrice = gasPrice
				}

				ep.logger.Debug(ctx, "Custom script job creation", observability.Uint64("gasLimit", gasLimit), observability.String("gasPrice", gasPrice.String()), observability.String("currentGasPrice", currentGasPrice.String()))
			}

			// Calculate on-chain fee
			onChainFeeWei = ep.gasEstimator.CalculateGasCostInWei(gasLimit, gasPrice)
			currentOnChainFeeWei = ep.gasEstimator.CalculateGasCostInWei(gasLimit, currentGasPrice)
		} else {
			ep.logger.Debug(ctx, "Custom script: no chain ID provided, skipping on-chain fee calculation")
		}

		// Calculate aggregator fee (same as other task types)
		const aggregatorGasUsed = uint64(720000)
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

		// Use shared gasEstimator for aggregator fee
		ethClient, err := ep.gasEstimator.getOrCreateClient(ctx, baseAggregatorFeeChainID, alchemyAPIKey)
		var aggregatorGasPrice *big.Int
		if err != nil {
			ep.logger.Warn(ctx, "Failed to get eth client for aggregator gas estimation", observability.String("baseAggregatorFeeChainID", baseAggregatorFeeChainID), observability.String("targetChainID", targetChainID), observability.Error(err))
			aggregatorGasPrice = big.NewInt(1000000000) // 1 gwei fallback
		} else {
			aggregatorGasPrice, err = ethClient.SuggestGasPrice(ctx)
			if err != nil {
				ep.logger.Warn(ctx, "Failed to get current gas price for aggregator", observability.String("baseAggregatorFeeChainID", baseAggregatorFeeChainID), observability.String("targetChainID", targetChainID), observability.Error(err))
				aggregatorGasPrice = big.NewInt(1000000000)
			} else {
				ep.logger.Debug(ctx, "Current gas price for aggregator fee (custom script)", observability.String("baseAggregatorFeeChainID", baseAggregatorFeeChainID), observability.String("aggregatorGasPrice", aggregatorGasPrice.String()))
			}
		}

		aggregatorOnChainFeeWei = ep.gasEstimator.CalculateGasCostInWei(aggregatorGasUsed, aggregatorGasPrice)

		// Total fee = off-chain fee + on-chain fee + aggregator on-chain fee
		totalFeeWei := new(big.Int).Add(offChainFeeWei, onChainFeeWei)
		currentTotalFeeWei := new(big.Int).Add(offChainFeeWei, currentOnChainFeeWei)
		currentTotalFeeWei.Add(currentTotalFeeWei, aggregatorOnChainFeeWei)
		totalFeeWei.Add(totalFeeWei, aggregatorOnChainFeeWei)
		// Apply 20% buffer (same as other task types)
		totalFeeWei.Mul(totalFeeWei, big.NewInt(120))
		totalFeeWei.Div(totalFeeWei, big.NewInt(100))

		ep.logger.Debug(ctx, "Fee calculation for custom script (ID 7)", observability.Float64("offChainFeeUSD", offChainFeeUSD), observability.String("offChainFeeWei", offChainFeeWei.String()), observability.String("onChainFeeWei", onChainFeeWei.String()), observability.String("aggregatorOnChainFeeWei", aggregatorOnChainFeeWei.String()), observability.String("totalFeeWei", totalFeeWei.String()), observability.String("currentTotalFeeWei", currentTotalFeeWei.String()))

		return totalFeeWei, currentTotalFeeWei
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
			// Check if execution failed or timed out - skip gas estimation in these cases
			executionFailed := !result.Success
			isTimeout := false
			if result.Error != nil {
				errorMsg := result.Error.Error()
				isTimeout = strings.Contains(errorMsg, "timeout") || strings.Contains(errorMsg, "execution timeout")
			}

			if executionFailed || isTimeout {
				ep.logger.Warn(ctx, "Skipping gas estimation due to execution failure or timeout", observability.String("chainID", chainID), observability.String("contractAddr", contractAddr), observability.String("function", function), observability.Bool("execution_failed", executionFailed), observability.Bool("is_timeout", isTimeout), observability.Error(result.Error))
				// Use zero on-chain fee when execution fails
				onChainFeeWei = big.NewInt(0)
				currentOnChainFeeWei = big.NewInt(0)
			} else {
				// Parse arguments from metadata if available
				var args []interface{}
				if argsStr, ok := execCtx.Metadata["on_chain_args"]; ok && argsStr != "" {
					if err := json.Unmarshal([]byte(argsStr), &args); err != nil {
						ep.logger.Warn(ctx, "Failed to parse on_chain_args, using empty args", observability.String("argsStr", argsStr[:min(len(argsStr), 200)]), observability.String("chainID", chainID), observability.String("contractAddr", contractAddr), observability.Error(err))
						args = []interface{}{}
					}
				} else {
					args = []interface{}{}
				}

				// For dynamic tasks (2, 4, 6), check if args are empty - this indicates execution didn't produce output
				var taskDefinitionID int
				shouldSkipGasEstimation := false
				if taskDefStr, ok := execCtx.Metadata["task_definition_id"]; ok {
					if _, err := fmt.Sscanf(taskDefStr, "%d", &taskDefinitionID); err == nil {
						// Dynamic tasks require arguments from execution output
						if (taskDefinitionID == 2 || taskDefinitionID == 4 || taskDefinitionID == 6) && len(args) == 0 {
							ep.logger.Warn(ctx, "Skipping gas estimation: dynamic task execution produced no arguments", observability.String("chainID", chainID), observability.String("contractAddr", contractAddr), observability.String("function", function), observability.Int("task_definition_id", taskDefinitionID))
							shouldSkipGasEstimation = true
						}
					}
				}

				if shouldSkipGasEstimation {
					// Use zero on-chain fee when no arguments are available
					onChainFeeWei = big.NewInt(0)
					currentOnChainFeeWei = big.NewInt(0)
				} else {
					// Get from address if provided
					fromAddress := execCtx.Metadata["from_address"]

					// Estimate gas for the on-chain transaction
					gasLimit, gasPrice, currentGasPrice, err := ep.gasEstimator.EstimateGasForFunction(
						ctx,
						chainID,
						contractAddr,
						function,
						contractABI,
						args,
						fromAddress,
						alchemyAPIKey,
					)

					if err != nil {
						ep.logger.Warn(ctx, "Failed to estimate gas, using default on-chain fee", observability.String("chainID", chainID), observability.String("contractAddr", contractAddr), observability.String("function", function), observability.Any("args", args), observability.Error(err))
						// Use a default on-chain fee if estimation fails (e.g., 0.001 ETH)
						defaultOnChainFee := big.NewFloat(0.001)
						defaultOnChainFee.Mul(defaultOnChainFee, weiMultiplier)
						onChainFeeWei, _ = defaultOnChainFee.Int(nil)
						// Also set currentOnChainFeeWei to the same default to prevent nil pointer dereference
				currentOnChainFeeWei, _ = defaultOnChainFee.Int(nil)
			} else {
						// Calculate gas cost in Wei
						onChainFeeWei = ep.gasEstimator.CalculateGasCostInWei(gasLimit, gasPrice)
						currentOnChainFeeWei = ep.gasEstimator.CalculateGasCostInWei(gasLimit, currentGasPrice)
						ep.logger.Debug(ctx, "Gas estimation", observability.Uint64("gasLimit", gasLimit), observability.String("gasPrice", gasPrice.String()), observability.String("currentGasPrice", currentGasPrice.String()), observability.String("gasCost", onChainFeeWei.String()), observability.String("currentGasCost", currentOnChainFeeWei.String()))
					}
				}
			}
		} else {
			ep.logger.Warn(ctx, "Missing contract details for on-chain fee calculation", observability.String("chainID", chainID), observability.String("contractAddr", contractAddr), observability.String("function", function))
			onChainFeeWei = big.NewInt(0)
			currentOnChainFeeWei = big.NewInt(0)
		}
	} else {
		ep.logger.Debug(ctx, "No chain ID provided, skipping on-chain fee calculation")
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

	// Use shared gasEstimator to fetch current gas price
	ethClient, err := ep.gasEstimator.getOrCreateClient(ctx, baseAggregatorFeeChainID, alchemyAPIKey)
	var aggregatorGasPrice *big.Int
	if err != nil {
		ep.logger.Warn(ctx, "Failed to get eth client for aggregator gas estimation", observability.String("baseAggregatorFeeChainID", baseAggregatorFeeChainID), observability.String("targetChainID", targetChainID), observability.Error(err))
		aggregatorGasPrice = big.NewInt(1000000000) // 1 gwei fallback
	} else {
		aggregatorGasPrice, err = ethClient.SuggestGasPrice(ctx)
		if err != nil {
			ep.logger.Warn(ctx, "Failed to get current gas price for aggregator", observability.String("baseAggregatorFeeChainID", baseAggregatorFeeChainID), observability.String("targetChainID", targetChainID), observability.Error(err))
			aggregatorGasPrice = big.NewInt(1000000000)
		} else {
			ep.logger.Debug(ctx, "Current gas price for aggregator fee", observability.String("baseAggregatorFeeChainID", baseAggregatorFeeChainID), observability.String("aggregatorGasPrice", aggregatorGasPrice.String()))
		}
	}

	aggregatorOnChainFeeWei = ep.gasEstimator.CalculateGasCostInWei(aggregatorGasUsed, aggregatorGasPrice)

	// Log aggregator fee calculation
	ep.logger.Debug(ctx, "Aggregator on-chain fee", observability.String("baseAggregatorFeeChainID", baseAggregatorFeeChainID), observability.Uint64("gasUsed", aggregatorGasUsed), observability.String("aggregatorGasPrice", aggregatorGasPrice.String()), observability.String("aggregatorOnChainFeeWei", aggregatorOnChainFeeWei.String()))
	// Total fee = off-chain fee + on-chain fee + aggregator on-chain fee
	// Ensure all fee components are non-nil to prevent panic
	if offChainFeeWei == nil {
		offChainFeeWei = big.NewInt(0)
	}
	if onChainFeeWei == nil {
		onChainFeeWei = big.NewInt(0)
	}
	if currentOnChainFeeWei == nil {
		currentOnChainFeeWei = big.NewInt(0)
	}
	if aggregatorOnChainFeeWei == nil {
		aggregatorOnChainFeeWei = big.NewInt(0)
	}

	totalFeeWei := new(big.Int).Add(offChainFeeWei, onChainFeeWei)
	currentTotalFeeWei := new(big.Int).Add(offChainFeeWei, currentOnChainFeeWei)
	currentTotalFeeWei.Add(currentTotalFeeWei, aggregatorOnChainFeeWei)
	totalFeeWei.Add(totalFeeWei, aggregatorOnChainFeeWei)
	totalFeeWei.Mul(totalFeeWei, big.NewInt(120))
	totalFeeWei.Div(totalFeeWei, big.NewInt(100)) // 20% buffer

	ep.logger.Debug(ctx, "Fee calculation", observability.Int("taskDefinitionID", taskDefinitionID), observability.Float64("offChainFeeUSD", offChainFeeUSD), observability.String("offChainFeeWei", offChainFeeWei.String()), observability.String("onChainFeeWei", onChainFeeWei.String()), observability.String("currentOnChainFeeWei", currentOnChainFeeWei.String()), observability.String("aggregatorOnChainFeeWei", aggregatorOnChainFeeWei.String()), observability.String("totalFeeWei", totalFeeWei.String()), observability.String("currentTotalFeeWei", currentTotalFeeWei.String()))
	return totalFeeWei, currentTotalFeeWei
}

// calculateLegacyFees is the old fee calculation method for backward compatibility
func (ep *executionPipeline) calculateLegacyFees(ctx context.Context, execCtx *types.ExecutionContext, feesConfig config.ExecutionFeeConfig) *big.Int {
	var staticComplexity, dynamicComplexity float64

	// Try to get complexity from metadata first (set in processResults)
	if staticStr, ok := execCtx.Metadata["static_complexity"]; ok {
		if parsed, err := fmt.Sscanf(staticStr, "%f", &staticComplexity); err != nil || parsed != 1 {
			ep.logger.Warn(ctx, "Failed to parse static complexity", observability.String("staticStr", staticStr), observability.Error(err))
			staticComplexity = 0.0
		}
	}

	if dynamicStr, ok := execCtx.Metadata["dynamic_complexity"]; ok {
		if parsed, err := fmt.Sscanf(dynamicStr, "%f", &dynamicComplexity); err != nil || parsed != 1 {
			ep.logger.Warn(ctx, "Failed to parse dynamic complexity", observability.String("dynamicStr", dynamicStr), observability.Error(err))
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

	ep.logger.Debug(ctx, "Legacy fee calculation", observability.Float64("staticComplexity", staticComplexity), observability.Float64("dynamicComplexity", dynamicComplexity), observability.Float64("x", x), observability.Float64("feeInTG", feeInTG), observability.Float64("feeInEther", feeInEther), observability.String("feeWei", feeWei.String()))

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
func (ep *executionPipeline) close(ctx context.Context) error {
	ep.mutex.Lock()
	defer ep.mutex.Unlock()

	if ep.closed {
		return nil
	}

	ep.logger.Debug(ctx, "Closing execution pipeline")

	// Signal shutdown to prevent new executions
	close(ep.shutdownChan)
	ep.closed = true

	// Cancel all active executions
	for executionID, execCtx := range ep.activeExecutions {
		ep.logger.Debug(ctx, "Cancelling active execution", observability.String("executionID", executionID))
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
		ep.logger.Debug(ctx, "All active executions completed")
	case <-time.After(30 * time.Second):
		activeCount := len(ep.activeExecutions)
		ep.logger.Warn(ctx, "Timeout waiting for active executions to complete", observability.Int("activeExecutionsCount", activeCount))
	}

	ep.logger.Debug(ctx, "Execution pipeline closed")
	return nil
}

func (ep *executionPipeline) cancelExecution(ctx context.Context, executionID string) error {
	ep.mutex.Lock()
	defer ep.mutex.Unlock()

	exec, exists := ep.activeExecutions[executionID]
	if !exists {
		return fmt.Errorf("execution not found: %s", executionID)
	}

	// Actually cancel the execution by calling the cancel function
	if exec.State.CancelFunc != nil {
		exec.State.CancelFunc()
		ep.logger.Debug(ctx, "Execution cancelled - context cancellation will terminate Docker processes", observability.String("executionID", executionID))
	} else {
		ep.logger.Warn(ctx, "Execution has no cancel function - marking as cancelled", observability.String("executionID", executionID))
	}

	// Attempt to terminate the Docker exec process if we have the exec ID
	if exec.State.ExecID != "" {
		ep.logger.Debug(ctx, "Attempting to terminate Docker exec process", observability.String("execID", exec.State.ExecID), observability.String("executionID", executionID))
		if err := ep.containerMgr.KillExecProcess(context.Background(), exec.State.ExecID); err != nil {
			ep.logger.Warn(ctx, "Failed to terminate exec process", observability.String("execID", exec.State.ExecID), observability.String("executionID", executionID), observability.String("containerID", exec.State.ContainerID), observability.Error(err))
		}
	}

	exec.CompletedAt = time.Now()

	ep.logger.Debug(ctx, "Execution cancelled", observability.String("executionID", executionID))
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
