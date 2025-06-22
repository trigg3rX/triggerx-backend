package aggregator

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/Layr-Labs/eigensdk-go/crypto/bls"
	"github.com/consensys/gnark-crypto/ecc/bn254"
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
	publicKey, ok := privateKey.Public().(*ecdsa.PublicKey)
	if !ok {
		c.logger.Error("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
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
		c.logger.Error("Failed to encode task data", "error", err)
		return false, fmt.Errorf("failed to encode task data: %w", err)
	}
	messageHash := crypto.Keccak256(dataPacked)

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
		PerformerAddress: performerAddress,
		Signature:        serializedSignature,
		SignatureType:    "ecdsa",
		TargetChainID:    84532,
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

func (c *AggregatorClient) SendTaskToValidatorsBLS(ctx context.Context, taskResult *types.BroadcastDataForValidators) (bool, error) {
	c.logger.Debug("Sending task result to aggregator",
		"taskDefinitionId", taskResult.TaskDefinitionID,
		"proofOfTask", taskResult.ProofOfTask)

	// BLS
	keyPair, err := getKeyPair(c.config.SenderPrivateKey)
	if err != nil {
		c.logger.Error("error creating key pair:", err)
		return false, fmt.Errorf("error creating key pair: %w", err)
	}

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
		c.logger.Error("Failed to encode task data", "error", err)
		return false, fmt.Errorf("failed to encode task data: %w", err)
	}
	messageHash := crypto.Keccak256Hash(dataPacked)

	// BLS
	messagePoint, err := hashToPoint(messageHash.Hex(), getDomain())
	if err != nil {
		c.logger.Error("error hashing message to point:", err)
		return false, fmt.Errorf("error hashing message to point: %w", err)
	}

	// Manual BLS signing: multiply G1 point by private key (like JavaScript mcl.mul)
	sig := new(bn254.G1Affine).ScalarMultiplication(messagePoint, keyPair.PrivKey.BigInt(new(big.Int)))

	// Extract x and y coordinates from the G1Affine point
	xBigInt := sig.X.BigInt(new(big.Int))
	yBigInt := sig.Y.BigInt(new(big.Int))

	// Create G1Point using the proper constructor
	g1Point := bls.NewG1Point(xBigInt, yBigInt)
	signature := &bls.Signature{
		G1Point: g1Point,
	}

	// Extract x,y coordinates from G1Point (like JavaScript g1ToHex)
	xBig := signature.G1Point.X.BigInt(new(big.Int))
	yBig := signature.G1Point.Y.BigInt(new(big.Int))

	// Convert to hex with proper formatting (32 bytes each for BN254)
	xBytes := make([]byte, 32)
	yBytes := make([]byte, 32)
	xBig.FillBytes(xBytes)
	yBig.FillBytes(yBytes)

	xHex := "0x" + hex.EncodeToString(xBytes)
	yHex := "0x" + hex.EncodeToString(yBytes)

	signatureJSON := map[string]string{
		"x": xHex,
		"y": yHex,
	}
	signatureJSONString, err := json.Marshal(signatureJSON)
	if err != nil {
		c.logger.Error("error marshaling signature to JSON:", err)
		return false, fmt.Errorf("error marshaling signature to JSON: %w", err)
	}

	// BLS: Prepare parameters using consistent structure
	params := CallParams{
		ProofOfTask:      taskResult.ProofOfTask,
		Data:             "0x" + hex.EncodeToString([]byte(taskResult.Data)),
		TaskDefinitionID: taskResult.TaskDefinitionID,
		PerformerAddress: c.config.SenderAddress,
		Signature:        string(signatureJSONString),
		SignatureType:    "bls",
		TargetChainID:    84532,
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
