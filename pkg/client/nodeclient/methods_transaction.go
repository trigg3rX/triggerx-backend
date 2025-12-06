package nodeclient

import (
	"context"
	"encoding/json"
	"fmt"
)

// EthGetTransactionByHash fetches a transaction by hash
func (c *NodeClient) EthGetTransactionByHash(ctx context.Context, txHash string) (*Transaction, error) {
	rpcParams := []interface{}{txHash}

	result, err := c.call(ctx, "eth_getTransactionByHash", rpcParams)
	if err != nil {
		return nil, fmt.Errorf("eth_getTransactionByHash failed: %w", err)
	}

	// Handle null result (transaction not found)
	if len(result) == 0 || string(result) == "null" {
		return nil, nil
	}

	var tx Transaction
	if err := json.Unmarshal(result, &tx); err != nil {
		return nil, fmt.Errorf("failed to unmarshal transaction: %w", err)
	}

	return &tx, nil
}

// EthGetTransactionReceipt fetches a transaction receipt by hash
func (c *NodeClient) EthGetTransactionReceipt(ctx context.Context, txHash string) (*TransactionReceipt, error) {
	rpcParams := []interface{}{txHash}

	result, err := c.call(ctx, "eth_getTransactionReceipt", rpcParams)
	if err != nil {
		return nil, fmt.Errorf("eth_getTransactionReceipt failed: %w", err)
	}

	// Handle null result (transaction not found)
	if len(result) == 0 || string(result) == "null" {
		return nil, nil
	}

	var receipt TransactionReceipt
	if err := json.Unmarshal(result, &receipt); err != nil {
		return nil, fmt.Errorf("failed to unmarshal transaction receipt: %w", err)
	}

	return &receipt, nil
}
