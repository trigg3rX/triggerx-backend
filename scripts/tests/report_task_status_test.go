//go:build integration
// +build integration

// This is an integration test for the ReportTaskStatus functionality.
// Run with: go test -tags=integration ./scripts/tests/... -v
//
// Prerequisites:
// 1. TaskMonitor service running (or mock server)
// 2. Environment variables set (KEEPER_ADDRESS, PRIVATE_KEY_CONSENSUS, TASK_MONITOR_RPC_URL)

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/trigg3rX/triggerx-backend/pkg/cryptography"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/rpc/client"
)

// ReportTaskStatusRequest mirrors the request type
type ReportTaskStatusRequest struct {
	TaskID              int64  `json:"task_id"`
	KeeperAddress       string `json:"keeper_address"`
	ExecutionSuccessful bool   `json:"execution_successful"`
	AggregatorSubmitted bool   `json:"aggregator_submitted"`
	ExecutionTxHash     string `json:"execution_tx_hash,omitempty"`
	ProofCID            string `json:"proof_cid,omitempty"`
	Error               string `json:"error,omitempty"`
	Signature           string `json:"signature"`
}

// ReportTaskStatusResponse mirrors the response type
type ReportTaskStatusResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

func main() {
	// Configuration - set these via environment variables or modify here for testing
	taskMonitorURL := getEnvOrDefault("TASK_MONITOR_RPC_URL", "http://localhost:8080")
	keeperAddress := getEnvOrDefault("KEEPER_ADDRESS", "0x1234567890abcdef1234567890abcdef12345678")
	privateKey := getEnvOrDefault("PRIVATE_KEY_CONSENSUS", "")

	if privateKey == "" {
		log.Fatal("PRIVATE_KEY_CONSENSUS environment variable is required")
	}

	logger := logging.NewNoOpLogger()

	// Create RPC client
	rpcClient := client.NewClient(client.Config{
		ServiceName: taskMonitorURL,
		Timeout:     30 * time.Second,
		MaxRetries:  3,
		RetryDelay:  time.Second,
		PoolSize:    5,
		PoolTimeout: 5 * time.Second,
	}, logger)
	defer rpcClient.Close()

	ctx := context.Background()

	fmt.Println("=== Testing ReportTaskStatus API ===")
	fmt.Printf("TaskMonitor URL: %s\n", taskMonitorURL)
	fmt.Printf("Keeper Address: %s\n\n", keeperAddress)

	// Test 1: Report successful task
	fmt.Println("--- Test 1: Report Successful Task ---")
	testReportSuccess(ctx, rpcClient, keeperAddress, privateKey)

	// Test 2: Report failed task
	fmt.Println("\n--- Test 2: Report Failed Task ---")
	testReportFailure(ctx, rpcClient, keeperAddress, privateKey)

	// Test 3: Report failed task without CID (early failure)
	fmt.Println("\n--- Test 3: Report Early Failure (no CID) ---")
	testReportEarlyFailure(ctx, rpcClient, keeperAddress, privateKey)

	fmt.Println("\n=== All tests completed ===")
}

func testReportSuccess(ctx context.Context, rpcClient *client.Client, keeperAddress, privateKey string) {
	taskID := int64(time.Now().Unix()) // Use timestamp as unique task ID for testing

	// Create signing data (must match server expectations)
	signData := struct {
		TaskID              int64  `json:"task_id"`
		KeeperAddress       string `json:"keeper_address"`
		ExecutionSuccessful bool   `json:"execution_successful"`
		AggregatorSubmitted bool   `json:"aggregator_submitted"`
		ExecutionTxHash     string `json:"execution_tx_hash,omitempty"`
		ProofCID            string `json:"proof_cid,omitempty"`
		Error               string `json:"error,omitempty"`
	}{
		TaskID:              taskID,
		KeeperAddress:       keeperAddress,
		ExecutionSuccessful: true,
		AggregatorSubmitted: true,
		ExecutionTxHash:     "0xabc123def456",
		ProofCID:            "QmTestSuccessCID12345",
		Error:               "",
	}

	signature, err := cryptography.SignJSONMessage(signData, privateKey)
	if err != nil {
		log.Printf("Failed to sign request: %v", err)
		return
	}

	request := ReportTaskStatusRequest{
		TaskID:              taskID,
		KeeperAddress:       keeperAddress,
		ExecutionSuccessful: true,
		AggregatorSubmitted: true,
		ExecutionTxHash:     "0xabc123def456",
		ProofCID:            "QmTestSuccessCID12345",
		Error:               "",
		Signature:           signature,
	}

	var response ReportTaskStatusResponse
	err = rpcClient.Call(ctx, "report-task-status", &request, &response)
	if err != nil {
		log.Printf("RPC call failed: %v", err)
		return
	}

	fmt.Printf("TaskID: %d\n", taskID)
	fmt.Printf("Response - Success: %v, Message: %s\n", response.Success, response.Message)
}

func testReportFailure(ctx context.Context, rpcClient *client.Client, keeperAddress, privateKey string) {
	taskID := int64(time.Now().Unix()) + 1 // Different task ID

	signData := struct {
		TaskID              int64  `json:"task_id"`
		KeeperAddress       string `json:"keeper_address"`
		ExecutionSuccessful bool   `json:"execution_successful"`
		AggregatorSubmitted bool   `json:"aggregator_submitted"`
		ExecutionTxHash     string `json:"execution_tx_hash,omitempty"`
		ProofCID            string `json:"proof_cid,omitempty"`
		Error               string `json:"error,omitempty"`
	}{
		TaskID:              taskID,
		KeeperAddress:       keeperAddress,
		ExecutionSuccessful: true,
		AggregatorSubmitted: false,
		ExecutionTxHash:     "0xdef789abc123",
		ProofCID:            "QmTestFailureCID67890",
		Error:               "aggregator submission failed: connection timeout",
	}

	signature, err := cryptography.SignJSONMessage(signData, privateKey)
	if err != nil {
		log.Printf("Failed to sign request: %v", err)
		return
	}

	request := ReportTaskStatusRequest{
		TaskID:              taskID,
		KeeperAddress:       keeperAddress,
		ExecutionSuccessful: true,
		AggregatorSubmitted: false,
		ExecutionTxHash:     "0xdef789abc123",
		ProofCID:            "QmTestFailureCID67890",
		Error:               "aggregator submission failed: connection timeout",
		Signature:           signature,
	}

	var response ReportTaskStatusResponse
	err = rpcClient.Call(ctx, "report-task-status", &request, &response)
	if err != nil {
		log.Printf("RPC call failed: %v", err)
		return
	}

	fmt.Printf("TaskID: %d\n", taskID)
	fmt.Printf("Response - Success: %v, Message: %s\n", response.Success, response.Message)
}

func testReportEarlyFailure(ctx context.Context, rpcClient *client.Client, keeperAddress, privateKey string) {
	taskID := int64(time.Now().Unix()) + 2 // Different task ID

	signData := struct {
		TaskID              int64  `json:"task_id"`
		KeeperAddress       string `json:"keeper_address"`
		ExecutionSuccessful bool   `json:"execution_successful"`
		AggregatorSubmitted bool   `json:"aggregator_submitted"`
		ExecutionTxHash     string `json:"execution_tx_hash,omitempty"`
		ProofCID            string `json:"proof_cid,omitempty"`
		Error               string `json:"error,omitempty"`
	}{
		TaskID:              taskID,
		KeeperAddress:       keeperAddress,
		ExecutionSuccessful: false,
		AggregatorSubmitted: false,
		ExecutionTxHash:     "", // No tx hash - failed before tx submission
		ProofCID:            "", // No CID - failed before IPFS upload
		Error:               "action execution failed: contract reverted",
	}

	signature, err := cryptography.SignJSONMessage(signData, privateKey)
	if err != nil {
		log.Printf("Failed to sign request: %v", err)
		return
	}

	request := ReportTaskStatusRequest{
		TaskID:              taskID,
		KeeperAddress:       keeperAddress,
		ExecutionSuccessful: false,
		AggregatorSubmitted: false,
		ExecutionTxHash:     "",
		ProofCID:            "",
		Error:               "action execution failed: contract reverted",
		Signature:           signature,
	}

	var response ReportTaskStatusResponse
	err = rpcClient.Call(ctx, "report-task-status", &request, &response)
	if err != nil {
		log.Printf("RPC call failed: %v", err)
		return
	}

	fmt.Printf("TaskID: %d\n", taskID)
	fmt.Printf("Response - Success: %v, Message: %s\n", response.Success, response.Message)
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
