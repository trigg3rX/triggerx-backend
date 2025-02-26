package services

import (
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common/hexutil"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rpc"

	"github.com/trigg3rX/triggerx-backend/internal/keeper/config"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
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
	}
	messageHash := crypto.Keccak256Hash(dataPacked)

	sig, err := crypto.Sign(messageHash.Bytes(), privateKey)
	if err != nil {
		logger.Errorf("Error signing message", "error", err)
	}
	sig[64] += 27
	serializedSignature := hexutil.Encode(sig)
	logger.Infof("Serialized signature", "signature", serializedSignature)

	client, err := rpc.Dial(config.AggregatorIPAddress)
	if err != nil {
		logger.Errorf("Error dialing RPC", "error", err)
	}

	params := types.PerformerData{
		ProofOfTask:      proofOfTask,
		Data:             "0x" + hex.EncodeToString([]byte(data)),
		TaskDefinitionID: fmt.Sprintf("%d", taskDefinitionId),
		PerformerAddress: performerAddress,
		PerformerSignature: serializedSignature,
	}

	response := makeRPCRequest(client, params)
	logger.Infof("API response:", "response", response)
}

func makeRPCRequest(client *rpc.Client, params types.PerformerData) interface{} {
	var result interface{}

	err := client.Call(&result, "sendTask", params.ProofOfTask, params.Data, params.TaskDefinitionID, params.PerformerAddress, params.PerformerSignature)
	if err != nil {
		logger.Errorf("Error making RPC request", "error", err)
	}
	return result
}
