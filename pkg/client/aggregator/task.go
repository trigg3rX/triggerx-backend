package aggregator

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// SendTaskToValidators sends a task result to the validators
func (c *AggregatorClient) SendTaskToValidators(ctx context.Context, taskResult *types.BroadcastDataForValidators) (bool, error) {
	c.logger.Debug(ctx, "Sending task result to aggregator",
		observability.Int("taskDefinitionId", taskResult.TaskDefinitionID),
		observability.String("proofOfTask", taskResult.ProofOfTask))

	privateKey, err := crypto.HexToECDSA(c.config.SenderPrivateKey)
	if err != nil {
		c.logger.Error(ctx, "Failed to convert private key to ECDSA", observability.Error(err))
		return false, fmt.Errorf("failed to convert private key to ECDSA: %w", err)
	}
	publicKey, ok := privateKey.Public().(*ecdsa.PublicKey)
	if !ok {
		c.logger.Error(ctx, "Cannot assert type: publicKey is not of type *ecdsa.PublicKey")
	}
	performerAddress := crypto.PubkeyToAddress(*publicKey).Hex()

	// Prepare ABI arguments
	arguments := abi.Arguments{
		{Type: abi.Type{T: abi.StringTy}},
		{Type: abi.Type{T: abi.BytesTy}},
		{Type: abi.Type{T: abi.AddressTy}},
		{Type: abi.Type{T: abi.UintTy}},
	}

	dataPacked, err := arguments.Pack(
		taskResult.ProofOfTask,
		taskResult.Data,
		common.HexToAddress(c.config.SenderAddress),
		big.NewInt(int64(taskResult.TaskDefinitionID)),
	)
	if err != nil {
		c.logger.Error(ctx, "Failed to encode task data", observability.Error(err))
		return false, fmt.Errorf("failed to encode task data: %w", err)
	}
	messageHash := crypto.Keccak256(dataPacked)

	sig, err := crypto.Sign(messageHash, privateKey)
	if err != nil {
		c.logger.Error(ctx, "Failed to sign task data", observability.Error(err))
		return false, fmt.Errorf("failed to sign task data: %w", err)
	}
	sig[64] += 27
	serializedSignature := hexutil.Encode(sig)

	var targetChainID int
	switch taskResult.TargetChainID {
		case 42161, 8453:
			targetChainID = 8453
		default:
			targetChainID = 84532
	}

	// Prepare parameters using consistent structure
	params := CallParams{
		ProofOfTask:      taskResult.ProofOfTask,
		Data:             "0x" + hex.EncodeToString(taskResult.Data),
		TaskDefinitionID: taskResult.TaskDefinitionID,
		PerformerAddress: performerAddress,
		Signature:        serializedSignature,
		SignatureType:    "ecdsa",
		TargetChainID:    targetChainID,
	}

	var response interface{}
	err = c.executeWithRetry(ctx, "sendTask", &response, params)
	if err != nil {
		c.logger.Error(ctx, "Failed to send task result", observability.Error(err))
		return false, fmt.Errorf("failed to send task result: %w", err)
	}

	c.logger.Debug(ctx, "Successfully sent task result to aggregator",
		observability.Int("taskDefinitionId", taskResult.TaskDefinitionID),
		observability.String("proofOfTask", taskResult.ProofOfTask),
		observability.Any("response", response))

	return true, nil
}
