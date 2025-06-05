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

// TimeBasedTaskData represents the data for a time based task
type TimeBasedTaskData struct {
	TaskDefinitionID int                 `json:"taskDefinitionId"`
	TimeJobData   *types.ScheduleTimeJobData `json:"timeJobData"`
	PerformerData types.GetPerformerData     `json:"performerData"`
}

type TaskData struct {
	TaskDefinitionID int                 `json:"taskDefinitionId"`
	TaskTargetData *types.TaskTargetData `json:"taskTargetData"`
	TriggerData    *types.TriggerData    `json:"triggerData"`
	PerformerData  types.GetPerformerData `json:"performerData"`
}

// SendTaskToPerformer sends a task to the specified performer through the aggregator
func (c *AggregatorClient) SendTimeBasedTaskToPerformer(
	ctx context.Context, 
	timeJobData *types.ScheduleTimeJobData,
	performerData types.GetPerformerData,
) (bool, error) {
	c.logger.Debug("Sending time based task to performer",
		"performerID", performerData.KeeperID,
		"jobID", timeJobData.JobID)

	// Pack task data
	taskData := &TimeBasedTaskData{
		TaskDefinitionID: timeJobData.TaskDefinitionID,
		TimeJobData:   timeJobData,
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

	abiPackedData, err := abiArguments.Pack(
		"proofOfTask",
		jsonData,
		common.HexToAddress(performerData.KeeperAddress),
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
		PerformerAddress: performerData.KeeperAddress,
		Signature:        signature,
	}

	var result interface{}
	err = c.executeWithRetry(ctx, "sendCustomMessage", &result, params)
	if err != nil {
		c.logger.Error("Failed to send time based task", "error", err)
		return false, fmt.Errorf("failed to send time based task: %w", err)
	}

	c.logger.Info("Task sent successfully",
		"performerID", performerData.KeeperID,
		"jobID", timeJobData.JobID,
		"result", result)
	return true, nil
}

// SendTaskToPerformer sends a task to the specified performer through the aggregator
func (c *AggregatorClient) SendTaskToPerformer(ctx context.Context, taskTargetData *types.TaskTargetData, triggerData *types.TriggerData, performerData types.GetPerformerData) (bool, error) {
	c.logger.Debug("Sending task to performer",
		"performerID", performerData.KeeperID,
		"jobID", taskTargetData.JobID)

	// Pack task data
	taskData := &TaskData{
		TaskDefinitionID: taskTargetData.TaskDefinitionID,
		TaskTargetData: taskTargetData,
		TriggerData:    triggerData,
		PerformerData:  performerData,
	}

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
		common.HexToAddress(performerData.KeeperAddress),
		big.NewInt(0),
	)
	if err != nil {
		c.logger.Error("Failed to encode task data", "error", err)
		return false, fmt.Errorf("%w: failed to encode data: %v", ErrMarshalFailed, err)
	}

	signature, err := c.signMessage(dataPacked)
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
		PerformerAddress: performerData.KeeperAddress,
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
		"jobID", taskTargetData.JobID,
		"result", result)
	return true, nil
}
