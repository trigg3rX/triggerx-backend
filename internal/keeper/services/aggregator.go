package services

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"os"
	"strings"

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

	client, err := rpc.Dial(config.AggregatorRPCAddress)
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

func ConnectToTaskManager(keeperAddress string, connectionAddress string) (bool, error) {
	taskManagerRPCAddress := fmt.Sprintf("%s/connect", config.ManagerRPCAddress)

	var payload types.UpdateKeeperConnectionData
	payload.KeeperAddress = keeperAddress
	payload.ConnectionAddress = connectionAddress

	// Ensure the connection address has the proper format for health checks
	if !strings.HasPrefix(payload.ConnectionAddress, "http://") && !strings.HasPrefix(payload.ConnectionAddress, "https://") {
		payload.ConnectionAddress = "https://" + payload.ConnectionAddress
	}

	logger.Info("Connecting to task manager",
		"keeper_address", keeperAddress,
		"connection_address", payload.ConnectionAddress,
		"task_manager", taskManagerRPCAddress)

	var response types.UpdateKeeperConnectionDataResponse

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return false, fmt.Errorf("failed to marshal payload: %w", err)
	}

	client := &http.Client{}

	req, err := http.NewRequest("POST", taskManagerRPCAddress, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return false, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return false, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	json.NewDecoder(resp.Body).Decode(&response)

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("task manager returned non-200 status code: %d", resp.StatusCode)
	}

	envFile := ".env"
	keeperIDLine := fmt.Sprintf("\nKEEPER_ID=%d", response.KeeperID)

	f, err := os.OpenFile(envFile, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return false, fmt.Errorf("failed to open .env file: %w", err)
	}
	defer f.Close()

	if _, err := f.WriteString(keeperIDLine); err != nil {
		return false, fmt.Errorf("failed to write keeper ID to .env: %w", err)
	}

	return true, nil
}
