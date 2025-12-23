package taskmonitor

import (
	"context"
	"fmt"
	"time"

	"github.com/trigg3rX/triggerx-backend/internal/keeper/config"
	"github.com/trigg3rX/triggerx-backend/pkg/cryptography"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
	"github.com/trigg3rX/triggerx-backend/pkg/rpc/client"
)

// Client represents a client for communicating with the taskmonitor service
type Client struct {
	rpcClient *client.Client
	logger    observability.Logger
}

// NewClient creates a new taskmonitor client
func NewClient(logger observability.Logger) (*Client, error) {
	rpcUrl := config.GetTaskMonitorRPCUrl()
	if rpcUrl == "" {
		return nil, fmt.Errorf("task monitor RPC URL is not configured")
	}

	rpcClient := client.NewClient(client.Config{
		ServiceName: rpcUrl,
		Timeout:     30 * time.Second,
		MaxRetries:  3,
		RetryDelay:  time.Second,
		PoolSize:    10,
		PoolTimeout: 5 * time.Second,
	}, logger)

	return &Client{
		rpcClient: rpcClient,
		logger:    logger,
	}, nil
}

// ReportTaskStatusRequest represents the request to report task execution status
type ReportTaskStatusRequest struct {
	TaskID              int64  `json:"task_id" validate:"required"`
	KeeperAddress       string `json:"keeper_address" validate:"required"`
	ExecutionSuccessful bool   `json:"execution_successful"`          // Whether the task execution itself succeeded
	AggregatorSubmitted bool   `json:"aggregator_submitted"`          // Whether the aggregator submission succeeded
	Error               string `json:"error,omitempty"`               // Error message if any step failed
	ExecutionTxHash     string `json:"execution_tx_hash,omitempty"`   // Transaction hash from on-chain execution
	ProofCID            string `json:"proof_cid,omitempty"`           // IPFS CID of the proof data
	Signature           string `json:"signature" validate:"required"` // Keeper's signature for authentication
}

// ReportTaskStatusResponse represents the response from taskmonitor
type ReportTaskStatusResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

// ReportTaskStatus reports task execution status to taskmonitor
// This should be called after the aggregator submission attempt (regardless of success or failure)
func (c *Client) ReportTaskStatus(ctx context.Context, taskID int64, executionSuccessful, aggregatorSubmitted bool, executionTxHash, proofCID, errorMsg string) error {
	keeperAddress := config.GetKeeperAddress()

	// Create request data for signing (without signature field)
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
		ExecutionSuccessful: executionSuccessful,
		AggregatorSubmitted: aggregatorSubmitted,
		ExecutionTxHash:     executionTxHash,
		ProofCID:            proofCID,
		Error:               errorMsg,
	}

	// Sign the request data
	signature, err := cryptography.SignJSONMessage(signData, config.GetPrivateKeyConsensus())
	if err != nil {
		return fmt.Errorf("failed to sign status report: %w", err)
	}

	// Create request
	request := ReportTaskStatusRequest{
		TaskID:              taskID,
		KeeperAddress:       keeperAddress,
		ExecutionSuccessful: executionSuccessful,
		AggregatorSubmitted: aggregatorSubmitted,
		ExecutionTxHash:     executionTxHash,
		ProofCID:            proofCID,
		Error:               errorMsg,
		Signature:           signature,
	}

	// Make RPC call
	var response ReportTaskStatusResponse
	err = c.rpcClient.Call(ctx, "report-task-status", &request, &response)
	if err != nil {
		return fmt.Errorf("RPC call failed: %w", err)
	}

	if !response.Success {
		return fmt.Errorf("taskmonitor reported failure: %s", response.Message)
	}

	c.logger.Debug(ctx, "Task status reported successfully to taskmonitor",
		observability.Int64("task_id", taskID),
		observability.Bool("execution_successful", executionSuccessful),
		observability.Bool("aggregator_submitted", aggregatorSubmitted),
		observability.String("execution_tx_hash", executionTxHash),
		observability.String("proof_cid", proofCID))

	return nil
}

// Close closes the taskmonitor client
func (c *Client) Close(ctx context.Context) error {
	if c.rpcClient != nil {
		return c.rpcClient.Close(ctx)
	}
	return nil
}

// --- DEPRECATED: ---
// Since executor and validator are controlled by us, backward compatibility is unnecessary.
//
// type ReportTaskErrorRequest struct {
// 	TaskID        int64  `json:"task_id"`
// 	KeeperAddress string `json:"keeper_address"`
// 	Error         string `json:"error"`
// 	Signature     string `json:"signature"`
// }
//
// type ReportTaskErrorResponse struct {
// 	Success bool   `json:"success"`
// 	Message string `json:"message,omitempty"`
// }
//
// func (c *Client) ReportTaskError(ctx context.Context, taskID int64, errorMsg string) error {
// 	keeperAddress := config.GetKeeperAddress()
// 	signData := struct {
// 		TaskID        int64  `json:"task_id"`
// 		KeeperAddress string `json:"keeper_address"`
// 		Error         string `json:"error"`
// 	}{
// 		TaskID:        taskID,
// 		KeeperAddress: keeperAddress,
// 		Error:         errorMsg,
// 	}
// 	signature, err := cryptography.SignJSONMessage(signData, config.GetPrivateKeyConsensus())
// 	if err != nil {
// 		return fmt.Errorf("failed to sign error report: %w", err)
// 	}
// 	request := ReportTaskErrorRequest{
// 		TaskID:        taskID,
// 		KeeperAddress: keeperAddress,
// 		Error:         errorMsg,
// 		Signature:     signature,
// 	}
// 	var response ReportTaskErrorResponse
// 	err = c.rpcClient.Call(ctx, "report-task-error", &request, &response)
// 	if err != nil {
// 		return fmt.Errorf("RPC call failed: %w", err)
// 	}
// 	if !response.Success {
// 		return fmt.Errorf("taskmonitor reported failure: %s", response.Message)
// 	}
// 	c.logger.Info("Task error reported successfully to taskmonitor",
// 		"task_id", taskID,
// 		"error", errorMsg)
// 	return nil
// }
