package aggregator

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// SendTaskToValidators sends a task result to the validators
func (c *AggregatorClient) SendTaskToValidators(ctx context.Context, taskResult *types.BroadcastDataForValidators) (bool, error) {
	c.logger.Debug("Sending task result to aggregator",
		"taskDefinitionId", taskResult.TaskDefinitionID,
		"proofOfTask", taskResult.ProofOfTask)

	privateKey, err := crypto.HexToECDSA(c.config.SenderPrivateKey)
	if err != nil {
		c.logger.Error("Failed to convert private key to ECDSA", "error", err)
		return false, fmt.Errorf("failed to convert private key to ECDSA: %w", err)
	}

	// Prepare ABI arguments
	arguments := abi.Arguments{
		{Type: abi.Type{T: abi.StringTy}},
		{Type: abi.Type{T: abi.BytesTy}},
		{Type: abi.Type{T: abi.AddressTy}},
		{Type: abi.Type{T: abi.UintTy}},
	}

	encodedData, err := arguments.Pack(
		taskResult.ProofOfTask,
		taskResult.Data,
		common.HexToAddress(taskResult.PerformerAddress),
		big.NewInt(int64(taskResult.TaskDefinitionID)),
	)
	if err != nil {
		c.logger.Error("Failed to encode task data", "error", err)
		return false, fmt.Errorf("failed to encode task data: %w", err)
	}
	messageHash := crypto.Keccak256(encodedData)

	// Sign the task data
	sig, err := crypto.Sign(messageHash, privateKey)
	if err != nil {
		c.logger.Error("Failed to sign task data", "error", err)
		return false, fmt.Errorf("failed to sign task data: %w", err)
	}
	sig[64] += 27
	serializedSignature := hexutil.Encode(sig)

	c.logger.Debug("Task data signed successfully", "signature", sig)

	// Prepare parameters using consistent structure
	params := CallParams{
		ProofOfTask:      taskResult.ProofOfTask,
		Data:             "0x" + hex.EncodeToString(taskResult.Data),
		TaskDefinitionID: taskResult.TaskDefinitionID,
		PerformerAddress: taskResult.PerformerAddress,
		Signature:        serializedSignature,
	}

	var response interface{}
	err = c.executeWithRetry(ctx, "sendTask", &response, params)
	if err != nil {
		c.logger.Error("Failed to send task result", "error", err)
		return false, fmt.Errorf("failed to send task result: %w", err)
	}

	c.logger.Info("Successfully sent task result to aggregator",
		"taskDefinitionId", taskResult.TaskDefinitionID,
		"proofOfTask", taskResult.ProofOfTask,
		"response", response)

	return true, nil
}
