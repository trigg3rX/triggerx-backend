package validation

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func (v *TaskValidator) fetchContractABI(contractAddress string) (string, error) {
	// First try using Blockscout API for Optimism Sepolia network
	blockscoutUrl := fmt.Sprintf(
		"https://optimism-sepolia.blockscout.com/api?module=contract&action=getabi&address=%s",
		contractAddress)

	resp, err := http.Get(blockscoutUrl)
	if err == nil && resp.StatusCode == http.StatusOK {
		defer func() {
			_ = resp.Body.Close()
		}()

		body, err := io.ReadAll(resp.Body)
		if err == nil {
			var response struct {
				Status  string `json:"status"`
				Message string `json:"message"`
				Result  string `json:"result"`
			}

			err = json.Unmarshal(body, &response)
			if err == nil && response.Status == "1" {
				return response.Result, nil
			}
		}
	}
	OptimismAPIKey := os.Getenv("OPTIMISM_API_KEY")
	if OptimismAPIKey == "" {
		return "", fmt.Errorf("OPTIMISM environment variable not set")
	}
	// If we reach here, Blockscout API failed, try Etherscan API as fallback
	etherscanUrl := fmt.Sprintf(
		"https://api-sepolia-optimism.etherscan.io/api?module=contract&action=getabi&address=%s&apikey=%s",
		contractAddress, OptimismAPIKey)

	resp, err = http.Get(etherscanUrl)
	if err != nil {
		return "", fmt.Errorf("failed to fetch ABI from both APIs: %v", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch ABI from both APIs, Etherscan status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read Etherscan response body: %v", err)
	}

	var response struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Result  string `json:"result"`
	}

	err = json.Unmarshal(body, &response)
	if err != nil {
		return "", fmt.Errorf("failed to parse Etherscan JSON response: %v", err)
	}

	if response.Status != "1" {
		return "", fmt.Errorf("error from both APIs, Etherscan error: %s", response.Message)
	}

	return response.Result, nil
}
