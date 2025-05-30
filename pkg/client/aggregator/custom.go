package aggregator

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rpc"

	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// SendTaskToPerformer sends a task to the specified performer through the aggregator
func (c *AggregatorClient) SendTaskToPerformer(jobData *types.HandleCreateJobData, triggerData *types.TriggerData, performerData types.GetPerformerData) (bool, error) {
	c.logger.Debug("Sending task to performer",
		"performerID", performerData.KeeperID,
		"jobID", jobData.JobID)

	// Pack task data
	data := map[string]interface{}{
		"jobData":       jobData,
		"triggerData":   triggerData,
		"performerData": performerData,
	}

	jsonData, err := json.Marshal(data)
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

	performerAddress := c.getPerformerAddress()

	dataPacked, err := arguments.Pack(
		"proofOfTask",
		jsonData,
		common.HexToAddress(performerAddress),
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

	// Connect to RPC
	client, err := rpc.Dial(c.config.AggregatorRPCAddress)
	if err != nil {
		c.logger.Error("Failed to connect to RPC", "error", err)
		return false, fmt.Errorf("%w: failed to dial: %v", ErrRPCFailed, err)
	}
	defer client.Close()

	// Prepare and send RPC request
	params := taskParams{
		proofOfTask:      "proofOfTask",
		data:             "0x" + hex.EncodeToString(jsonData),
		taskDefinitionID: 0,
		performerAddress: performerAddress,
		signature:        signature,
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.config.RequestTimeout)
	defer cancel()

	var result interface{}
	err = client.CallContext(ctx, &result, "sendCustomMessage", params.data, params.taskDefinitionID)
	if err != nil {
		c.logger.Error("RPC request failed", "error", err)
		return false, fmt.Errorf("%w: %v", ErrRPCFailed, err)
	}

	c.logger.Info("Task sent successfully",
		"performerID", performerData.KeeperID,
		"jobID", jobData.JobID,
		"result", result)
	return true, nil
}