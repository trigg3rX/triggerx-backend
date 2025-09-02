package alchemy

import (
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

// GetTransactionByHash retrieves transaction details by hash
func (c *Client) GetTransactionByHash(txHash string, chainID string) (*Transaction, bool, error) {
	params := []string{txHash}
	response, err := c.makeRequest("eth_getTransactionByHash", params, chainID)
	if err != nil {
		return nil, false, err
	}

	if response.Result == nil {
		return nil, false, fmt.Errorf("transaction not found")
	}

	// Convert result to JSON and then unmarshal to Transaction
	resultJSON, err := json.Marshal(response.Result)
	if err != nil {
		return nil, false, fmt.Errorf("failed to marshal result: %w", err)
	}

	var transaction Transaction
	if err := json.Unmarshal(resultJSON, &transaction); err != nil {
		return nil, false, fmt.Errorf("failed to unmarshal transaction: %w", err)
	}

	return &transaction, false, nil
}

// GetTransactionReceipt retrieves transaction receipt by hash
func (c *Client) GetTransactionReceipt(txHash string, chainID string) (*TransactionReceipt, error) {
	params := []string{txHash}
	response, err := c.makeRequest("eth_getTransactionReceipt", params, chainID)
	if err != nil {
		return nil, err
	}

	if response.Result == nil {
		return nil, fmt.Errorf("transaction receipt not found")
	}

	resultJSON, err := json.Marshal(response.Result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	var receipt TransactionReceipt
	if err := json.Unmarshal(resultJSON, &receipt); err != nil {
		return nil, fmt.Errorf("failed to unmarshal receipt: %w", err)
	}

	return &receipt, nil
}

// GetGasPrice retrieves the current gas price
func (c *Client) GetGasPrice(chainID string) (*big.Int, error) {
	response, err := c.makeRequest("eth_gasPrice", []interface{}{}, chainID)
	if err != nil {
		return nil, err
	}

	if response.Result == nil {
		return nil, fmt.Errorf("gas price not available")
	}

	gasPriceHex, ok := response.Result.(string)
	if !ok {
		return nil, fmt.Errorf("invalid gas price format")
	}

	// Remove "0x" prefix and convert hex to big.Int
	if len(gasPriceHex) > 2 && gasPriceHex[:2] == "0x" {
		gasPriceHex = gasPriceHex[2:]
	}

	gasPrice := new(big.Int)
	gasPrice.SetString(gasPriceHex, 16)

	return gasPrice, nil
}

// EstimateGas estimates the gas required for a transaction
func (c *Client) EstimateGas(from, to, data string, value *big.Int, chainID string) (*big.Int, error) {
	params := map[string]interface{}{
		"from": from,
		"to":   to,
		"data": data,
	}

	if value != nil && value.Sign() > 0 {
		params["value"] = "0x" + value.Text(16)
	}

	response, err := c.makeRequest("eth_estimateGas", []interface{}{params}, chainID)
	if err != nil {
		return nil, err
	}

	if response.Result == nil {
		return nil, fmt.Errorf("gas estimation failed")
	}

	gasHex, ok := response.Result.(string)
	if !ok {
		return nil, fmt.Errorf("invalid gas estimate format")
	}

	// Remove "0x" prefix and convert hex to big.Int
	if len(gasHex) > 2 && gasHex[:2] == "0x" {
		gasHex = gasHex[2:]
	}

	gas := new(big.Int)
	gas.SetString(gasHex, 16)

	return gas, nil
}

// SendRawTransaction sends a signed raw transaction
func (c *Client) SendRawTransaction(signedTx string, chainID string) (string, error) {
	params := []string{signedTx}
	response, err := c.makeRequest("eth_sendRawTransaction", params, chainID)
	if err != nil {
		return "", err
	}

	if response.Result == nil {
		return "", fmt.Errorf("transaction submission failed")
	}

	txHash, ok := response.Result.(string)
	if !ok {
		return "", fmt.Errorf("invalid transaction hash format")
	}

	return txHash, nil
}

// SendTransaction sends a transaction (requires account to be unlocked)
func (c *Client) SendTransaction(tx *SendTransactionRequest, chainID string) (string, error) {
	response, err := c.makeRequest("eth_sendTransaction", []interface{}{tx}, chainID)
	if err != nil {
		return "", err
	}

	if response.Result == nil {
		return "", fmt.Errorf("transaction submission failed")
	}

	txHash, ok := response.Result.(string)
	if !ok {
		return "", fmt.Errorf("invalid transaction hash format")
	}

	return txHash, nil
}

// GetBlockByNumber retrieves block details by block number
func (c *Client) GetBlockByNumber(blockNumber string, fullTx bool, chainID string) (*Block, error) {
	if blockNumber == "" {
		blockNumber = "latest"
	}

	params := []interface{}{blockNumber, fullTx}
	response, err := c.makeRequest("eth_getBlockByNumber", params, chainID)
	if err != nil {
		return nil, err
	}

	if response.Result == nil {
		return nil, fmt.Errorf("block not found")
	}

	resultJSON, err := json.Marshal(response.Result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	var block Block
	if err := json.Unmarshal(resultJSON, &block); err != nil {
		return nil, fmt.Errorf("failed to unmarshal block: %w", err)
	}

	return &block, nil
}

// GetBlockByHash retrieves block details by block hash
func (c *Client) GetBlockByHash(blockHash string, fullTx bool, chainID string) (*Block, error) {
	params := []interface{}{blockHash, fullTx}
	response, err := c.makeRequest("eth_getBlockByHash", params, chainID)
	if err != nil {
		return nil, err
	}

	if response.Result == nil {
		return nil, fmt.Errorf("block not found")
	}

	resultJSON, err := json.Marshal(response.Result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	var block Block
	if err := json.Unmarshal(resultJSON, &block); err != nil {
		return nil, fmt.Errorf("failed to unmarshal block: %w", err)
	}

	return &block, nil
}

// GetBlockNumber retrieves the latest block number
func (c *Client) GetBlockNumber(chainID string) (*big.Int, error) {
	response, err := c.makeRequest("eth_blockNumber", []interface{}{}, chainID)
	if err != nil {
		return nil, err
	}

	if response.Result == nil {
		return nil, fmt.Errorf("block number not available")
	}

	blockNumberHex, ok := response.Result.(string)
	if !ok {
		return nil, fmt.Errorf("invalid block number format")
	}

	// Remove "0x" prefix and convert hex to big.Int
	if len(blockNumberHex) > 2 && blockNumberHex[:2] == "0x" {
		blockNumberHex = blockNumberHex[2:]
	}

	blockNumber := new(big.Int)
	blockNumber.SetString(blockNumberHex, 16)

	return blockNumber, nil
}

// GetBalance retrieves the balance of an account
func (c *Client) GetBalance(address, blockTag string, chainID string) (*big.Int, error) {
	if blockTag == "" {
		blockTag = "latest"
	}

	params := []string{address, blockTag}
	response, err := c.makeRequest("eth_getBalance", params, chainID)
	if err != nil {
		return nil, err
	}

	if response.Result == nil {
		return nil, fmt.Errorf("balance not available")
	}

	balanceHex, ok := response.Result.(string)
	if !ok {
		return nil, fmt.Errorf("invalid balance format")
	}

	// Remove "0x" prefix and convert hex to big.Int
	if len(balanceHex) > 2 && balanceHex[:2] == "0x" {
		balanceHex = balanceHex[2:]
	}

	balance := new(big.Int)
	balance.SetString(balanceHex, 16)

	return balance, nil
}

func (c *Client) GetPendingNonce(address common.Address, chainID string) (uint64, error) {
	response, err := c.makeRequest("eth_getPendingNonce", []interface{}{address}, chainID)
	if err != nil {
		return 0, err
	}

	if response.Result == nil {
		return 0, fmt.Errorf("pending nonce not available")
	}

	nonceHex, ok := response.Result.(uint64)
	if !ok {
		return 0, fmt.Errorf("invalid pending nonce format")
	}

	return nonceHex, nil
}