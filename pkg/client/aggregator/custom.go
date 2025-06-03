package aggregator

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"

	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// CustomTaskData represents the data for a custom task
type CustomTaskData struct {
	JobData       *types.HandleCreateJobData `json:"jobData"`
	TriggerData   *types.TriggerData         `json:"triggerData"`
	PerformerData types.GetPerformerData     `json:"performerData"`
}

// SendTaskToPerformer sends a task to the specified performer through the aggregator
func (c *AggregatorClient) SendTaskToPerformer(ctx context.Context, jobData *types.HandleCreateJobData, triggerData *types.TriggerData, performerData types.GetPerformerData) (bool, error) {
	c.logger.Debug("Sending task to performer",
		"performerID", performerData.KeeperID,
		"jobID", jobData.JobID)

	// Pack task data
	taskData := &CustomTaskData{
		JobData:       jobData,
		TriggerData:   triggerData,
		PerformerData: performerData,
	}

	jsonData, err := json.Marshal(taskData)
	if err != nil {
		c.logger.Error("Failed to marshal task data", "error", err)
		return false, fmt.Errorf("%w: %v", ErrMarshalFailed, err)
	}

	// Prepare ABI arguments for additional encoding if needed
	abiArguments := abi.Arguments{
		{Type: abi.Type{T: abi.StringTy}},
		{Type: abi.Type{T: abi.BytesTy}},
		{Type: abi.Type{T: abi.AddressTy}},
		{Type: abi.Type{T: abi.UintTy}},
	}

	performerAddress := performerData.KeeperAddress

	abiPackedData, err := abiArguments.Pack(
		"proofOfTask",
		jsonData,
		common.HexToAddress(performerAddress),
		big.NewInt(0),
	)
	if err != nil {
		c.logger.Error("Failed to encode task data", "error", err)
		return false, fmt.Errorf("%w: failed to encode data: %v", ErrMarshalFailed, err)
	}

	signature, err := c.signMessage(abiPackedData)
	if err != nil {
		c.logger.Error("Failed to sign task data", "error", err)
		return false, err
	}

	c.logger.Debug("Task data signed successfully", "signature", signature)

	// Prepare parameters using the same structure as task.go
	params := struct {
		ProofOfTask      string `json:"proofOfTask"`
		Data             string `json:"data"`
		TaskDefinitionID int    `json:"taskDefinitionId"`
		PerformerAddress string `json:"performerAddress"`
		Signature        string `json:"signature"`
	}{
		ProofOfTask:      "proofOfTask",
		Data:             "0x" + hex.EncodeToString(jsonData),
		TaskDefinitionID: 0,
		PerformerAddress: performerAddress,
		Signature:        signature,
	}

	var result interface{}
	err = c.executeWithRetry(ctx, "sendCustomMessage", &result, params)
	if err != nil {
		c.logger.Error("Failed to send custom task", "error", err)
		return false, fmt.Errorf("failed to send custom task: %w", err)
	}

	c.logger.Info("Task sent successfully",
		"performerID", performerData.KeeperID,
		"jobID", jobData.JobID,
		"result", result)
	return true, nil
}
