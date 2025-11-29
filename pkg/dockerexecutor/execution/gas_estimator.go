package execution

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

// GasEstimator handles gas estimation for on-chain transactions
type GasEstimator struct {
	clients map[string]*ethclient.Client // Map of chainID -> eth client
	logger  logging.Logger
}

// NewGasEstimator creates a new gas estimator instance
func NewGasEstimator(logger logging.Logger) *GasEstimator {
	return &GasEstimator{
		clients: make(map[string]*ethclient.Client),
		logger:  logger,
	}
}

// getOrCreateClient gets or creates an eth client for the given chain ID
func (ge *GasEstimator) getOrCreateClient(ctx context.Context, chainID string) (*ethclient.Client, error) {
	// Check if client already exists
	if client, exists := ge.clients[chainID]; exists {
		return client, nil
	}

	// Map of chain IDs to RPC URLs (you should configure this based on your setup)
	rpcURLs := map[string]string{
		"1":     "https://eth-mainnet.g.alchemy.com/v2/YOUR_API_KEY", // Ethereum Mainnet
		"11155111": "https://eth-sepolia.g.alchemy.com/v2/YOUR_API_KEY", // Sepolia Testnet
		"137":   "https://polygon-mainnet.g.alchemy.com/v2/YOUR_API_KEY", // Polygon
		"80001": "https://polygon-mumbai.g.alchemy.com/v2/YOUR_API_KEY", // Mumbai Testnet
		"42161": "", //Arbitrum mainnet
		// Add more chains as needed
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

// EstimateGasForFunction estimates gas for calling a specific function on a contract
func (ge *GasEstimator) EstimateGasForFunction(
	ctx context.Context,
	chainID string,
	contractAddress string,
	functionName string,
	contractABI string,
	args []interface{},
) (uint64, *big.Int, error) {
	// Get or create client for the chain
	client, err := ge.getOrCreateClient(ctx, chainID)
	if err != nil {
		return 0, nil, err
	}

	// Parse the ABI
	parsedABI, err := abi.JSON(strings.NewReader(contractABI))
	if err != nil {
		return 0, nil, fmt.Errorf("failed to parse ABI: %w", err)
	}

	// Pack the function call data
	callData, err := parsedABI.Pack(functionName, args...)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to pack function call: %w", err)
	}

	// Create the call message
	to := common.HexToAddress(contractAddress)
	msg := ethereum.CallMsg{
		To:   &to,
		Data: callData,
	}

	// Estimate gas
	gasLimit, err := client.EstimateGas(ctx, msg)
	if err != nil {
		ge.logger.Warnf("Gas estimation failed, using default value: %v", err)
		// Use a default gas limit if estimation fails
		gasLimit = 300000 // Default fallback
	}

	// Get current gas price
	gasPrice, err := client.SuggestGasPrice(ctx)
	if err != nil {
		ge.logger.Warnf("Failed to get gas price, using default: %v", err)
		// Use a default gas price if suggestion fails (e.g., 50 gwei)
		gasPrice = big.NewInt(50000000000)
	}

	ge.logger.Debugf("Gas estimation for %s.%s: gasLimit=%d, gasPrice=%s", contractAddress, functionName, gasLimit, gasPrice.String())

	return gasLimit, gasPrice, nil
}

// CalculateGasCostInWei calculates the total gas cost in Wei
func (ge *GasEstimator) CalculateGasCostInWei(gasLimit uint64, gasPrice *big.Int) *big.Int {
	gasCost := new(big.Int).SetUint64(gasLimit)
	gasCost.Mul(gasCost, gasPrice)
	return gasCost
}

// Close closes all eth clients
func (ge *GasEstimator) Close() {
	for chainID, client := range ge.clients {
		client.Close()
		ge.logger.Debugf("Closed eth client for chain %s", chainID)
	}
	ge.clients = make(map[string]*ethclient.Client)
}
