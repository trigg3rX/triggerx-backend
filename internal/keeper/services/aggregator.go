package services

import (
	"crypto/ecdsa"
	"encoding/hex"
	"math/big"

	"github.com/ethereum/go-ethereum/common/hexutil"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rpc"

	"github.com/trigg3rX/triggerx-backend/internal/keeper/config"
)

type Params struct {
	proofOfTask      string
	data             string
	taskDefinitionId int
	performerAddress string
	signature        string
}

func SendTask(proofOfTask string, data string, taskDefinitionId int) {
	privateKey, err := crypto.HexToECDSA(config.PrivateKeyController)
	if err != nil {
		logger.Errorf("Error converting private key", "error", err)
	}
	publicKey, ok := privateKey.Public().(*ecdsa.PublicKey)
	if !ok {
		logger.Errorf("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
	}
	performerAddress := crypto.PubkeyToAddress(*publicKey).Hex()

	// performerAddress := common.HexToAddress(config.KeeperAddress)

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
	logger.Infof("Serialized signature: %s", serializedSignature)

	client, err := rpc.Dial(config.AggregatorIPAddress)
	if err != nil {
		logger.Errorf("Error dialing RPC", "error", err)
		return
	}

	params := Params{
		proofOfTask:      proofOfTask,
		data:             "0x" + hex.EncodeToString([]byte(data)),
		taskDefinitionId: taskDefinitionId,
		performerAddress: performerAddress,
		signature:        serializedSignature,
	}

	response := makeRPCRequest(client, params)
	logger.Infof("API response: %v", response)
}

func makeRPCRequest(client *rpc.Client, params Params) interface{} {
	var result interface{}

	err := client.Call(&result, "sendTask", params.proofOfTask, params.data, params.taskDefinitionId, params.performerAddress, params.signature)
	if err != nil {
		logger.Errorf("Error making RPC request: %v", err)
	}
	return result
}