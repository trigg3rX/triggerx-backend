package nodeclient

import (
	"context"
	"encoding/json"
	"fmt"
)

// EthGetBlockByNumber fetches a block by number
func (c *NodeClient) EthGetBlockByNumber(ctx context.Context, blockNumber BlockNumber, fullTx bool) (*Block, error) {
	rpcParams := []interface{}{string(blockNumber), fullTx}

	result, err := c.call(ctx, "eth_getBlockByNumber", rpcParams)
	if err != nil {
		return nil, fmt.Errorf("eth_getBlockByNumber failed: %w", err)
	}

	// Handle null result (block not found)
	if len(result) == 0 || string(result) == "null" {
		return nil, nil
	}

	var block Block
	if err := json.Unmarshal(result, &block); err != nil {
		return nil, fmt.Errorf("failed to unmarshal block: %w", err)
	}

	return &block, nil
}

// EthGetBlockReceipts fetches all transaction receipts for a given block
func (c *NodeClient) EthGetBlockReceipts(ctx context.Context, blockNumber BlockNumber) ([]TransactionReceipt, error) {
	rpcParams := []interface{}{string(blockNumber)}

	result, err := c.call(ctx, "eth_getBlockReceipts", rpcParams)
	if err != nil {
		return nil, fmt.Errorf("eth_getBlockReceipts failed: %w", err)
	}

	// Handle empty result
	if len(result) == 0 || string(result) == "null" {
		return []TransactionReceipt{}, nil
	}

	var receipts []TransactionReceipt
	if err := json.Unmarshal(result, &receipts); err != nil {
		return nil, fmt.Errorf("failed to unmarshal block receipts: %w", err)
	}

	return receipts, nil
}

// EthGetLogs fetches event logs matching the specified filter
func (c *NodeClient) EthGetLogs(ctx context.Context, params EthGetLogsParams) ([]Log, error) {
	// Convert params to interface{} for JSON-RPC call
	rpcParams := []interface{}{params}

	result, err := c.call(ctx, "eth_getLogs", rpcParams)
	if err != nil {
		return nil, fmt.Errorf("eth_getLogs failed: %w", err)
	}

	// Handle empty result
	if len(result) == 0 || string(result) == "null" {
		return []Log{}, nil
	}

	var logs []Log
	if err := json.Unmarshal(result, &logs); err != nil {
		return nil, fmt.Errorf("failed to unmarshal logs: %w", err)
	}

	return logs, nil
}

// EthGetFilteredLogs fetches event logs matching the specified filter (alias for eth_getLogs)
func (c *NodeClient) EthGetFilteredLogs(ctx context.Context, params EthGetLogsParams) ([]Log, error) {
	// eth_getFilteredLogs is typically an alias for eth_getLogs
	return c.EthGetLogs(ctx, params)
}
