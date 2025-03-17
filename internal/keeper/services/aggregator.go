package services

import (
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common/hexutil"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rpc"

	"github.com/trigg3rX/triggerx-backend/internal/keeper/config"
)

func SendTask(proofOfTask string, data string, taskDefinitionId int) {
	privateKey, err := crypto.HexToECDSA(config.PrivateKeyController)
	if err != nil {
		logger.Errorf("Error converting private key", "error", err)
	}
	publicKey, ok := privateKey.Public().(*ecdsa.PublicKey)
	if !ok {
		logger.Error("Cannot assert type: publicKey is not of type *ecdsa.PublicKey")
	}
	performerAddress := crypto.PubkeyToAddress(*publicKey).Hex()

	arguments := abi.Arguments{
		{Type: abi.Type{T: abi.StringTy}},
		{Type: abi.Type{T: abi.BytesTy}},
		{Type: abi.Type{T: abi.AddressTy}},
		{Type: abi.Type{T: abi.UintTy}},
	}

	dataPacked, err := arguments.Pack(
		proofOfTask,
		[]byte(data),
		common.HexToAddress(config.KeeperAddress),
		big.NewInt(int64(taskDefinitionId)),
	)
	if err != nil {
		logger.Errorf("Error encoding data", "error", err)
		return
	}
	messageHash := crypto.Keccak256Hash(dataPacked)

	sig, err := crypto.Sign(messageHash.Bytes(), privateKey)
	if err != nil {
		logger.Errorf("Error signing message", "error", err)
		return
	}
	sig[64] += 27
	serializedSignature := hexutil.Encode(sig)
	logger.Infof("Serialized signature", "signature", serializedSignature)

	aggregatorURL := config.AggregatorIPAddress
	if !strings.HasPrefix(aggregatorURL, "http://") && !strings.HasPrefix(aggregatorURL, "https://") {
		aggregatorURL = "http://" + aggregatorURL
	}

	client, err := rpc.Dial(aggregatorURL)
	if err != nil {
		logger.Errorf("Error dialing RPC", "error", err)
		return
	}

	// Convert taskDefinitionId to integer for RPC call
	taskDefID, err := strconv.Atoi(fmt.Sprintf("%d", taskDefinitionId))
	if err != nil {
		logger.Errorf("Error converting taskDefinitionId", "error", err)
		return
	}

	response := makeRPCRequest(client, proofOfTask, "0x"+hex.EncodeToString([]byte(data)), taskDefID, performerAddress, serializedSignature)
	logger.Infof("API response:", "response", response)
}

func makeRPCRequest(client *rpc.Client, proofOfTask string, data string, taskDefinitionID int, performerAddress string, performerSignature string) interface{} {
	var result interface{}

	err := client.Call(&result, "sendTask", proofOfTask, data, taskDefinitionID, performerAddress, performerSignature)
	if err != nil {
		logger.Errorf("Error making RPC request", "error", err)
	}
	return result
}
