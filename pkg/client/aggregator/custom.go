package aggregator

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"

	"github.com/trigg3rX/triggerx-backend/pkg/cryptography"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// SendTaskToPerformer sends a task to the specified performer through the aggregator
func (c *AggregatorClient) SendTaskToPerformer(ctx context.Context, taskData *types.SendTaskDataToKeeper) (bool, error) {
	c.logger.Debug("Sending task to performer",
		"performerID", taskData.PerformerData.KeeperID,
		"TaskID", taskData.TargetData.TaskID)

	jsonData, err := json.Marshal(taskData)
	if err != nil {
		c.logger.Error("Failed to marshal task data", "error", err)
		return false, fmt.Errorf("%w: %v", ErrMarshalFailed, err)
	}

	// Prepare ABI arguments
	arguments := abi.Arguments{
		{Type: abi.Type{T: abi.StringTy}},
		{Type: abi.Type{T: abi.BytesTy}},
		{Type: abi.Type{T: abi.AddressTy}},
		{Type: abi.Type{T: abi.UintTy}},
	}

	dataPacked, err := arguments.Pack(
		"proofOfTask",
		jsonData,
		common.HexToAddress(taskData.PerformerData.KeeperAddress),
		big.NewInt(int64(taskData.TargetData.TaskDefinitionID)),
	)
	if err != nil {
		c.logger.Error("Failed to encode task data", "error", err)
		return false, fmt.Errorf("%w: failed to encode data: %v", ErrMarshalFailed, err)
	}

	signature, err := cryptography.SignJSONMessage(dataPacked, c.config.SenderPrivateKey)
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
		TaskDefinitionID: taskData.TargetData.TaskDefinitionID,
		PerformerAddress: taskData.PerformerData.KeeperAddress,
		Signature:        signature,
	}

	var result interface{}
	err = c.executeWithRetry(ctx, "sendCustomMessage", &result, params)
	if err != nil {
		c.logger.Error("Failed to send custom task", "error", err)
		return false, fmt.Errorf("failed to send custom task: %w", err)
	}

	c.logger.Info("Task sent successfully",
		"performerID", taskData.PerformerData.KeeperID,
		"TaskID", taskData.TargetData.TaskID,
		"result", result)
	return true, nil
}
