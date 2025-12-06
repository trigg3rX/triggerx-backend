package execution

import (
	"context"
	"fmt"
	"math/big"
	// "os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

const (
	// cacheValidityDuration is how long the cached gas price is valid (7 days)
	cacheValidityDuration = 7 * 24 * time.Hour
	// percentile75 is the 75th percentile we use for gas price prediction
	percentile75 = 0.75
	// blocksPerDay is approximate blocks per day (varies by chain, using Ethereum as baseline ~7200 blocks/day)
	blocksPerDay = 7200
	// blocksFor7Days is approximate blocks for 7 days
	blocksFor7Days = 7 * blocksPerDay
)

// gasPriceCache stores cached gas price data for a chain
type gasPriceCache struct {
	gasPrice  *big.Int
	updatedAt time.Time
	mu        sync.RWMutex
}

// GasEstimator handles gas estimation for on-chain transactions
type GasEstimator struct {
	clients       map[string]*ethclient.Client // Map of chainID -> eth client
	logger        logging.Logger
	gasPriceCache map[string]*gasPriceCache // Map of chainID -> cached gas price
	cacheMu       sync.RWMutex              // Mutex for gas price cache
}

// NewGasEstimator creates a new gas estimator instance
func NewGasEstimator(logger logging.Logger) *GasEstimator {
	return &GasEstimator{
		clients:       make(map[string]*ethclient.Client),
		logger:        logger,
		gasPriceCache: make(map[string]*gasPriceCache),
	}
}

// convertArgsToABITypes converts JSON arguments to proper ABI types
func (ge *GasEstimator) convertArgsToABITypes(args []interface{}, inputs abi.Arguments) ([]interface{}, error) {
	if len(args) != len(inputs) {
		return nil, fmt.Errorf("argument count mismatch: got %d, expected %d", len(args), len(inputs))
	}

	converted := make([]interface{}, len(args))
	for i, arg := range args {
		argType := inputs[i].Type

		// Convert based on type
		switch argType.T {
		case abi.AddressTy:
			// Convert string to address
			strVal, ok := arg.(string)
			if !ok {
				return nil, fmt.Errorf("argument %d: expected string for address, got %T", i, arg)
			}
			converted[i] = common.HexToAddress(strVal)

		case abi.UintTy, abi.IntTy:
			// Convert to appropriate integer type based on size
			var bigInt *big.Int
			switch v := arg.(type) {
			case string:
				bigInt = new(big.Int)
				// Try to parse as decimal first, then hex if that fails
				if _, ok := bigInt.SetString(v, 10); !ok {
					if strings.HasPrefix(v, "0x") {
						if _, ok := bigInt.SetString(v[2:], 16); !ok {
							return nil, fmt.Errorf("argument %d: invalid number format: %s", i, v)
						}
					} else {
						return nil, fmt.Errorf("argument %d: invalid number format: %s", i, v)
					}
				}
			case float64:
				bigInt = big.NewInt(int64(v))
			case int:
				bigInt = big.NewInt(int64(v))
			case int64:
				bigInt = big.NewInt(v)
			default:
				return nil, fmt.Errorf("argument %d: unsupported type for uint/int: %T", i, arg)
			}

			// For small integer types (uint8, uint16, etc.), convert to native Go types
			// For larger types (uint256), use *big.Int
			if argType.Size <= 64 {
				// Convert to native Go integer type
				val := bigInt.Uint64()
				switch argType.Size {
				case 8:
					if argType.T == abi.UintTy {
						converted[i] = uint8(val)
					} else {
						converted[i] = int8(val)
					}
				case 16:
					if argType.T == abi.UintTy {
						converted[i] = uint16(val)
					} else {
						converted[i] = int16(val)
					}
				case 32:
					if argType.T == abi.UintTy {
						converted[i] = uint32(val)
					} else {
						converted[i] = int32(val)
					}
				case 64:
					if argType.T == abi.UintTy {
						converted[i] = val
					} else {
						converted[i] = int64(val)
					}
				default:
					// Fallback to *big.Int for other sizes
					converted[i] = bigInt
				}
			} else {
				// For uint256 and larger, use *big.Int
				converted[i] = bigInt
			}

		case abi.BytesTy, abi.FixedBytesTy:
			// Convert string to bytes
			strVal, ok := arg.(string)
			if !ok {
				return nil, fmt.Errorf("argument %d: expected string for bytes, got %T", i, arg)
			}
			// Remove 0x prefix if present
			if strings.TrimPrefix(strVal, "0x") != strVal {
				// Convert hex string to bytes
				bytes := common.Hex2Bytes(strVal)
				converted[i] = bytes
			} else {
				return nil, fmt.Errorf("argument %d: invalid hex string: %s", i, strVal)
			}

		case abi.BoolTy:
			boolVal, ok := arg.(bool)
			if !ok {
				return nil, fmt.Errorf("argument %d: expected bool, got %T", i, arg)
			}
			converted[i] = boolVal

		case abi.StringTy:
			strVal, ok := arg.(string)
			if !ok {
				return nil, fmt.Errorf("argument %d: expected string, got %T", i, arg)
			}
			converted[i] = strVal

		default:
			// For other types, pass as-is and let the ABI packer handle it
			converted[i] = arg
		}
	}

	return converted, nil
}

// getOrCreateClient gets or creates an eth client for the given chain ID
func (ge *GasEstimator) getOrCreateClient(ctx context.Context, chainID string, alchemyAPIKey string) (*ethclient.Client, error) {
	// Check if client already exists
	if client, exists := ge.clients[chainID]; exists {
		return client, nil
	}
	// Read Alchemy API key from environment (if set)
	// alchemyAPIKey = os.Getenv("ALCHEMY_API_KEY")

	// Map of chain IDs to RPC URLs (you should configure this based on your setup)
	rpcURLs := map[string]string{
		// Ethereum
		"1":        fmt.Sprintf("https://eth-mainnet.g.alchemy.com/v2/%s", alchemyAPIKey), // Ethereum Mainnet
		"11155111": fmt.Sprintf("https://eth-sepolia.g.alchemy.com/v2/%s", alchemyAPIKey), // Ethereum Sepolia
		// Optimism
		"10":       fmt.Sprintf("https://opt-mainnet.g.alchemy.com/v2/%s", alchemyAPIKey), // Optimism Mainnet
		"11155420": fmt.Sprintf("https://opt-sepolia.g.alchemy.com/v2/%s", alchemyAPIKey), // Optimism Sepolia
		// Polygon
		"137":   fmt.Sprintf("https://polygon-mainnet.g.alchemy.com/v2/%s", alchemyAPIKey), // Polygon Mainnet
		"80001": fmt.Sprintf("https://polygon-mumbai.g.alchemy.com/v2/%s", alchemyAPIKey),  // Polygon Mumbai
		// Arbitrum
		"42161":  fmt.Sprintf("https://arb-mainnet.g.alchemy.com/v2/%s", alchemyAPIKey), // Arbitrum One
		"421614": fmt.Sprintf("https://arb-sepolia.g.alchemy.com/v2/%s", alchemyAPIKey), // Arbitrum Sepolia
		// Base
		"8453":  fmt.Sprintf("https://base-mainnet.g.alchemy.com/v2/%s", alchemyAPIKey), // Base Mainnet
		"84532": fmt.Sprintf("https://base-sepolia.g.alchemy.com/v2/%s", alchemyAPIKey), // Base Sepolia
	}
	// Fail early if missing key
	if alchemyAPIKey == "" {
		return nil, fmt.Errorf("ALCHEMY_API_KEY must be set in env for all supported chains")
	}

	rpcURL, ok := rpcURLs[chainID]
	if !ok {
		return nil, fmt.Errorf("unsupported chain ID: %s", chainID)
	}

	client, err := ethclient.DialContext(ctx, rpcURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to chain %s: %w", chainID, err)
	}

	ge.clients[chainID] = client
	return client, nil
}

// calculatePercentile calculates the percentile value from a sorted slice of big.Int values
func calculatePercentile(values []*big.Int, percentile float64) *big.Int {
	if len(values) == 0 {
		return big.NewInt(0)
	}

	// Sort values in ascending order
	sorted := make([]*big.Int, len(values))
	copy(sorted, values)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Cmp(sorted[j]) < 0
	})

	// Calculate index for percentile
	index := int(float64(len(sorted)) * percentile)
	if index >= len(sorted) {
		index = len(sorted) - 1
	}

	return new(big.Int).Set(sorted[index])
}

// fetchHistoricalGasPrices fetches historical gas prices for the last 7 days and calculates 75th percentile
func (ge *GasEstimator) fetchHistoricalGasPrices(ctx context.Context, chainID string, alchemyAPIKey string) (*big.Int, error) {
	client, err := ge.getOrCreateClient(ctx, chainID, alchemyAPIKey)
	if err != nil {
		return nil, err
	}

	// Get latest block number
	latestBlock, err := client.BlockNumber(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest block number: %w", err)
	}

	// Calculate block range for 7 days (approximately)
	// We'll fetch in chunks since FeeHistory has limits
	var allMaxGasPrices []*big.Int
	blockCount := uint64(1024) // Max blocks per FeeHistory call (varies by RPC, 1024 is safe)
	fromBlock := latestBlock

	// Fetch historical fee data in chunks
	// We need approximately blocksFor7Days blocks, but fetch in smaller chunks
	totalBlocksNeeded := uint64(blocksFor7Days)
	blocksFetched := uint64(0)
	maxChunks := 20 // Safety limit to avoid infinite loops

	for blocksFetched < totalBlocksNeeded && len(allMaxGasPrices) < maxChunks*int(blockCount) {
		if fromBlock < blockCount {
			break // Can't go below block 0
		}

		toBlock := big.NewInt(int64(fromBlock))
		fromBlockNum := fromBlock - blockCount + 1

		// Request fee history with reward percentiles to get max gas prices
		// We request [0, 100] percentiles to get min and max
		feeHistory, err := client.FeeHistory(ctx, blockCount, toBlock, []float64{0, 100})
		if err != nil {
			ge.logger.Warnf("Failed to fetch fee history for blocks %d-%d: %v, using current gas price", fromBlockNum, fromBlock, err)
			// Fallback to current gas price if historical fetch fails
			currentGasPrice, err := client.SuggestGasPrice(ctx)
			if err != nil {
				return big.NewInt(1000000000), nil // Default 1 gwei
			}
			return currentGasPrice, nil
		}

		// Extract max gas prices from fee history
		// For EIP-1559 chains, we use baseFee + maxPriorityFee
		// For legacy chains, baseFee might not be accurate, but we'll use it as approximation
		for i, baseFee := range feeHistory.BaseFee {
			if baseFee == nil || baseFee.Cmp(big.NewInt(0)) <= 0 {
				continue
			}

			// Get max reward (priority fee) for this block if available
			var maxGasPrice *big.Int
			if i < len(feeHistory.Reward) && len(feeHistory.Reward[i]) > 0 {
				// EIP-1559: maxGasPrice = baseFee + maxPriorityFee
				maxPriorityFee := feeHistory.Reward[i][len(feeHistory.Reward[i])-1] // Last element is max (100th percentile)
				if maxPriorityFee != nil && maxPriorityFee.Cmp(big.NewInt(0)) >= 0 {
					maxGasPrice = new(big.Int).Add(baseFee, maxPriorityFee)
				} else {
					// If maxPriorityFee is invalid, use baseFee * 2 as approximation
					maxGasPrice = new(big.Int).Mul(baseFee, big.NewInt(2))
				}
			} else {
				// Legacy chain or no reward data: use baseFee * 2 as approximation
				// (baseFee alone might be too low for legacy chains)
				maxGasPrice = new(big.Int).Mul(baseFee, big.NewInt(2))
			}

			if maxGasPrice.Cmp(big.NewInt(0)) > 0 {
				allMaxGasPrices = append(allMaxGasPrices, maxGasPrice)
			}
		}

		blocksFetched += blockCount
		fromBlock = fromBlockNum - 1

		// Safety check
		if fromBlock < blockCount {
			break
		}
	}

	if len(allMaxGasPrices) == 0 {
		ge.logger.Warnf("No historical gas prices found for chain %s, using current gas price", chainID)
		currentGasPrice, err := client.SuggestGasPrice(ctx)
		if err != nil {
			return big.NewInt(1000000000), nil // Default 1 gwei
		}
		return currentGasPrice, nil
	}

	// Calculate 75th percentile of max gas prices
	percentile75GasPrice := calculatePercentile(allMaxGasPrices, percentile75)
	ge.logger.Infof("Calculated 75th percentile gas price for chain %s: %s (from %d samples)", chainID, percentile75GasPrice.String(), len(allMaxGasPrices))

	return percentile75GasPrice, nil
}

// getOrUpdateCachedGasPrice gets cached gas price or updates it if expired
func (ge *GasEstimator) getOrUpdateCachedGasPrice(ctx context.Context, chainID string, alchemyAPIKey string) (*big.Int, error) {
	ge.cacheMu.RLock()
	cache, exists := ge.gasPriceCache[chainID]
	ge.cacheMu.RUnlock()

	// Check if cache exists and is still valid
	if exists {
		cache.mu.RLock()
		cachedPrice := cache.gasPrice
		cachedTime := cache.updatedAt
		cache.mu.RUnlock()

		if time.Since(cachedTime) < cacheValidityDuration {
			ge.logger.Debugf("Using cached gas price for chain %s: %s (age: %v)", chainID, cachedPrice.String(), time.Since(cachedTime))
			return new(big.Int).Set(cachedPrice), nil
		}
	}

	// Cache expired or doesn't exist, fetch new historical data
	ge.logger.Infof("Cache expired or missing for chain %s, fetching historical gas prices...", chainID)
	gasPrice, err := ge.fetchHistoricalGasPrices(ctx, chainID, alchemyAPIKey)
	if err != nil {
		// If historical fetch fails, try to use current gas price
		ge.logger.Warnf("Failed to fetch historical gas prices for chain %s: %v, using current gas price", chainID, err)
		client, err := ge.getOrCreateClient(ctx, chainID, alchemyAPIKey)
		if err != nil {
			return big.NewInt(1000000000), nil // Default fallback
		}
		currentGasPrice, err := client.SuggestGasPrice(ctx)
		if err != nil {
			return big.NewInt(1000000000), nil // Default fallback
		}
		gasPrice = currentGasPrice
	}

	// Update cache
	ge.cacheMu.Lock()
	if ge.gasPriceCache[chainID] == nil {
		ge.gasPriceCache[chainID] = &gasPriceCache{}
	}
	cache = ge.gasPriceCache[chainID]
	cache.mu.Lock()
	cache.gasPrice = new(big.Int).Set(gasPrice)
	cache.updatedAt = time.Now()
	cache.mu.Unlock()
	ge.cacheMu.Unlock()

	ge.logger.Infof("Updated gas price cache for chain %s: %s", chainID, gasPrice.String())
	return new(big.Int).Set(gasPrice), nil
}

// EstimateGasForFunction estimates gas for calling a specific function on a contract
func (ge *GasEstimator) EstimateGasForFunction(
	ctx context.Context,
	chainID string,
	contractAddress string,
	functionName string,
	contractABI string,
	args []interface{},
	fromAddress string,
	alchemyAPIKey string,
) (uint64, *big.Int, *big.Int, error) {
	// Get or create client for the chain
	client, err := ge.getOrCreateClient(ctx, chainID, alchemyAPIKey)
	if err != nil {
		return 0, nil, nil, err
	}

	// Parse the ABI
	parsedABI, err := abi.JSON(strings.NewReader(contractABI))
	if err != nil {
		return 0, nil, nil, fmt.Errorf("failed to parse ABI: %w", err)
	}

	// Get the method from ABI
	method, ok := parsedABI.Methods[functionName]
	if !ok {
		return 0, nil, nil, fmt.Errorf("method %s not found in ABI", functionName)
	}

	// Convert args to proper types based on ABI
	convertedArgs, err := ge.convertArgsToABITypes(args, method.Inputs)
	if err != nil {
		return 0, nil, nil, fmt.Errorf("failed to convert arguments: %w", err)
	}

	// Pack the function call data
	callData, err := parsedABI.Pack(functionName, convertedArgs...)
	if err != nil {
		return 0, nil, nil, fmt.Errorf("failed to pack function call: %w", err)
	}

	// Create the call message
	to := common.HexToAddress(contractAddress)

	// Use provided from address if available, otherwise use a default
	var from common.Address
	if fromAddress != "" {
		from = common.HexToAddress(fromAddress)
	}
	msg := ethereum.CallMsg{
		To:   &to,
		Data: callData,
		From: from,
	}

	// Estimate gas
	gasLimit, err := client.EstimateGas(ctx, msg)
	if err != nil {
		ge.logger.Warnf("Gas estimation failed, using default value: %v", err)
		// Use a default gas limit if estimation fails
		gasLimit = 1000000 // Default fallback
	}

	// Get cached historical gas price (75th percentile of last 7 days)
	gasPrice, err := ge.getOrUpdateCachedGasPrice(ctx, chainID, alchemyAPIKey)
	if err != nil {
		ge.logger.Warnf("Failed to get cached gas price, using current gas price: %v", err)
		// Fallback to current gas price if cache fails
		gasPrice, err = client.SuggestGasPrice(ctx)
		if err != nil {
			ge.logger.Warnf("Failed to get gas price, using default: %v", err)
			// Use a default gas price if suggestion fails (e.g., 1 gwei)
			gasPrice = big.NewInt(1000000000)
		}
	}

	//Get current gas price
	currentGasPrice, err := client.SuggestGasPrice(ctx)
	if err != nil {
		ge.logger.Warnf("Failed to get current gas price, using default: %v", err)
		// Use a default gas price if suggestion fails (e.g., 1 gwei)
		currentGasPrice = big.NewInt(1000000000)
	}

	ge.logger.Debugf("Gas estimation for %s.%s: gasLimit=%d, gasPrice=%s (75th percentile from last 7 days)", contractAddress, functionName, gasLimit, gasPrice.String())

	return gasLimit, gasPrice,currentGasPrice, nil
}

// GetGasPrice gets the cached historical gas price (75th percentile of last 7 days) for a chain
func (ge *GasEstimator) GetGasPrice(ctx context.Context, chainID string, alchemyAPIKey string) (*big.Int, error) {
	// Get cached historical gas price (75th percentile of last 7 days)
	gasPrice, err := ge.getOrUpdateCachedGasPrice(ctx, chainID, alchemyAPIKey)
	if err != nil {
		ge.logger.Warnf("Failed to get cached gas price, using current gas price: %v", err)
		// Fallback to current gas price if cache fails
		client, err := ge.getOrCreateClient(ctx, chainID, alchemyAPIKey)
		if err != nil {
			return big.NewInt(1000000000), nil // Default fallback
		}
		gasPrice, err = client.SuggestGasPrice(ctx)
		if err != nil {
			ge.logger.Warnf("Failed to get gas price, using default: %v", err)
			// Use a default gas price if suggestion fails (e.g., 1 gwei)
			gasPrice = big.NewInt(1000000000)
		}
	}

	return gasPrice, nil
}

// CalculateGasCostInWei calculates the total gas cost in Wei
func (ge *GasEstimator) CalculateGasCostInWei(gasLimit uint64, gasPrice *big.Int) *big.Int {
	gasCost := new(big.Int).SetUint64(gasLimit)
	gasCost.Mul(gasCost, gasPrice)
	return gasCost
}

// Close closes all eth clients and clears the cache
func (ge *GasEstimator) Close() {
	for chainID, client := range ge.clients {
		client.Close()
		ge.logger.Debugf("Closed eth client for chain %s", chainID)
	}
	ge.clients = make(map[string]*ethclient.Client)

	// Clear gas price cache
	ge.cacheMu.Lock()
	ge.gasPriceCache = make(map[string]*gasPriceCache)
	ge.cacheMu.Unlock()
}
