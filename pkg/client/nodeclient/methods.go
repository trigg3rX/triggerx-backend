package nodeclient

import (
	"context"
	"encoding/json"
	"fmt"
)

// EthBlockNumber returns the number of the most recent block
func (c *NodeClient) EthBlockNumber(ctx context.Context) (string, error) {
	result, err := c.call(ctx, "eth_blockNumber", []interface{}{})
	if err != nil {
		return "", fmt.Errorf("eth_blockNumber failed: %w", err)
	}

	var blockNumber string
	if err := json.Unmarshal(result, &blockNumber); err != nil {
		return "", fmt.Errorf("failed to unmarshal block number: %w", err)
	}

	return blockNumber, nil
}

// EthChainId returns the chain ID of the network
func (c *NodeClient) EthChainId(ctx context.Context) (string, error) {
	result, err := c.call(ctx, "eth_chainId", []interface{}{})
	if err != nil {
		return "", fmt.Errorf("eth_chainId failed: %w", err)
	}

	var chainId string
	if err := json.Unmarshal(result, &chainId); err != nil {
		return "", fmt.Errorf("failed to unmarshal chain ID: %w", err)
	}

	return chainId, nil
}

// EthGasPrice returns the current gas price in wei
func (c *NodeClient) EthGasPrice(ctx context.Context) (string, error) {
	result, err := c.call(ctx, "eth_gasPrice", []interface{}{})
	if err != nil {
		return "", fmt.Errorf("eth_gasPrice failed: %w", err)
	}

	var gasPrice string
	if err := json.Unmarshal(result, &gasPrice); err != nil {
		return "", fmt.Errorf("failed to unmarshal gas price: %w", err)
	}

	return gasPrice, nil
}

// EthMaxPriorityFeePerGas returns the current max priority fee per gas in wei
func (c *NodeClient) EthMaxPriorityFeePerGas(ctx context.Context) (string, error) {
	result, err := c.call(ctx, "eth_maxPriorityFeePerGas", []interface{}{})
	if err != nil {
		return "", fmt.Errorf("eth_maxPriorityFeePerGas failed: %w", err)
	}

	var maxPriorityFeePerGas string
	if err := json.Unmarshal(result, &maxPriorityFeePerGas); err != nil {
		return "", fmt.Errorf("failed to unmarshal max priority fee per gas: %w", err)
	}

	return maxPriorityFeePerGas, nil
}

// EthGetCode returns the code at a given address
func (c *NodeClient) EthGetCode(ctx context.Context, address string, blockNumber BlockNumber) (string, error) {
	rpcParams := []interface{}{address, string(blockNumber)}

	result, err := c.call(ctx, "eth_getCode", rpcParams)
	if err != nil {
		return "", fmt.Errorf("eth_getCode failed: %w", err)
	}

	var code string
	if err := json.Unmarshal(result, &code); err != nil {
		return "", fmt.Errorf("failed to unmarshal code: %w", err)
	}

	return code, nil
}

// EthGetStorageAt returns the value from a storage position at a given address
func (c *NodeClient) EthGetStorageAt(ctx context.Context, address string, position string, blockNumber BlockNumber) (string, error) {
	rpcParams := []interface{}{address, position, string(blockNumber)}

	result, err := c.call(ctx, "eth_getStorageAt", rpcParams)
	if err != nil {
		return "", fmt.Errorf("eth_getStorageAt failed: %w", err)
	}

	var storageValue string
	if err := json.Unmarshal(result, &storageValue); err != nil {
		return "", fmt.Errorf("failed to unmarshal storage value: %w", err)
	}

	return storageValue, nil
}

// EthEstimateGas estimates the gas needed to execute a transaction
func (c *NodeClient) EthEstimateGas(ctx context.Context, params EthEstimateGasParams) (string, error) {
	// Convert params to map for JSON-RPC call
	rpcParams := []interface{}{params}

	result, err := c.call(ctx, "eth_estimateGas", rpcParams)
	if err != nil {
		return "", fmt.Errorf("eth_estimateGas failed: %w", err)
	}

	var gasEstimate string
	if err := json.Unmarshal(result, &gasEstimate); err != nil {
		return "", fmt.Errorf("failed to unmarshal gas estimate: %w", err)
	}

	return gasEstimate, nil
}

// EthSendRawTransaction broadcasts a signed transaction to the network
func (c *NodeClient) EthSendRawTransaction(ctx context.Context, signedTxData string) (string, error) {
	rpcParams := []interface{}{signedTxData}

	result, err := c.call(ctx, "eth_sendRawTransaction", rpcParams)
	if err != nil {
		return "", fmt.Errorf("eth_sendRawTransaction failed: %w", err)
	}

	var txHash string
	if err := json.Unmarshal(result, &txHash); err != nil {
		return "", fmt.Errorf("failed to unmarshal transaction hash: %w", err)
	}

	return txHash, nil
}

// EthSendRawTransactionSync broadcasts a signed transaction to the network and waits for it to be included in a block
func (c *NodeClient) EthSendRawTransactionSync(ctx context.Context, signedTxData string) (*TransactionReceipt, error) {
	rpcParams := []interface{}{signedTxData}

	result, err := c.call(ctx, "eth_sendRawTransactionSync", rpcParams)
	if err != nil {
		return nil, fmt.Errorf("eth_sendRawTransactionSync failed: %w", err)
	}

	// Handle null result
	if len(result) == 0 || string(result) == "null" {
		return nil, nil
	}

	var receipt TransactionReceipt
	if err := json.Unmarshal(result, &receipt); err != nil {
		return nil, fmt.Errorf("failed to unmarshal transaction receipt: %w", err)
	}

	return &receipt, nil
}

// GetConfig returns the client configuration
func (c *NodeClient) GetConfig() *Config {
	return c.config
}
