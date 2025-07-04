package validation

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	ethtypes "github.com/ethereum/go-ethereum/core/types"
)

const timeTolerance = 1100 * time.Millisecond

func (v *TaskValidator) getBlockTimestamp(receipt *ethtypes.Receipt, rpcURL string) (time.Time, error) {
	blockNumberHex := fmt.Sprintf("0x%x", receipt.BlockNumber)
	reqBody := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "eth_getBlockByNumber",
		"params": []interface{}{
			blockNumberHex,
			false,
		},
		"id": 1,
	}
	reqBytes, err := json.Marshal(reqBody)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to marshal eth_getBlockReceipts request: %v", err)
	}

	httpReq, err := http.NewRequest("POST", rpcURL, bytes.NewBuffer(reqBytes))
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to create eth_getBlockReceipts request: %v", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	httpClient := &http.Client{Timeout: 10 * time.Second}
	resp, err := httpClient.Do(httpReq)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to call eth_getBlockReceipts: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return time.Time{}, fmt.Errorf("eth_getBlockReceipts returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var rpcResp struct {
		Result json.RawMessage `json:"result"`
		Error  interface{}     `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&rpcResp); err != nil {
		return time.Time{}, fmt.Errorf("failed to decode eth_getBlockReceipts response: %v", err)
	}
	if rpcResp.Error != nil {
		return time.Time{}, fmt.Errorf("eth_getBlockReceipts error: %v", rpcResp.Error)
	}

	var block map[string]interface{}
	if err := json.Unmarshal(rpcResp.Result, &block); err != nil {
		return time.Time{}, fmt.Errorf("failed to unmarshal block: %v", err)
	}

	timestampHex, ok := block["timestamp"].(string)
	if !ok {
		return time.Time{}, fmt.Errorf("block timestamp is not a string")
	}
	timestampInt, err := strconv.ParseInt(timestampHex, 0, 64)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse block timestamp hex: %v", err)
	}
	txTimestamp := time.Unix(timestampInt, 0).UTC()

	return txTimestamp, nil
}

// func (v *TaskValidator) fetchContractABI(contractAddress string) (string, error) {
// 	blockscoutUrl := fmt.Sprintf(
// 		"https://optimism-sepolia.blockscout.com/api?module=contract&action=getabi&address=%s",
// 		contractAddress)

// 	resp, err := http.Get(blockscoutUrl)
// 	if err == nil && resp.StatusCode == http.StatusOK {
// 		defer func() {
// 			_ = resp.Body.Close()
// 		}()

// 		body, err := io.ReadAll(resp.Body)
// 		if err == nil {
// 			var response struct {
// 				Status  string `json:"status"`
// 				Message string `json:"message"`
// 				Result  string `json:"result"`
// 			}

// 			err = json.Unmarshal(body, &response)
// 			if err == nil && response.Status == "1" {
// 				return response.Result, nil
// 			}
// 		}
// 	}
// 	etherscanUrl := fmt.Sprintf(
// 		"https://api-sepolia-optimism.etherscan.io/api?module=contract&action=getabi&address=%s&apikey=%s",
// 		contractAddress, v.etherscanAPIKey)

// 	resp, err = http.Get(etherscanUrl)
// 	if err != nil {
// 		return "", fmt.Errorf("failed to fetch ABI from both APIs: %v", err)
// 	}
// 	defer func() {
// 		_ = resp.Body.Close()
// 	}()

// 	if resp.StatusCode != http.StatusOK {
// 		return "", fmt.Errorf("failed to fetch ABI from both APIs, Etherscan status code: %d", resp.StatusCode)
// 	}

// 	body, err := io.ReadAll(resp.Body)
// 	if err != nil {
// 		return "", fmt.Errorf("failed to read Etherscan response body: %v", err)
// 	}

// 	var response struct {
// 		Status  string `json:"status"`
// 		Message string `json:"message"`
// 		Result  string `json:"result"`
// 	}

// 	err = json.Unmarshal(body, &response)
// 	if err != nil {
// 		return "", fmt.Errorf("failed to parse Etherscan JSON response: %v", err)
// 	}

// 	if response.Status != "1" {
// 		return "", fmt.Errorf("error from both APIs, Etherscan error: %s", response.Message)
// 	}

// 	return response.Result, nil
// }
