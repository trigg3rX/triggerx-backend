package services

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/localtunnel/go-localtunnel"

	"github.com/trigg3rX/triggerx-backend/internal/keeper/config"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

var logger = logging.GetLogger(logging.Development, logging.KeeperProcess)

func Init() {
	config.Init()
	logger.Info("Config Initialized")
}

type Params struct {
	proofOfTask      string
	data             string
	taskDefinitionId int
	performerAddress string
	signature        string
}

func SendTask(proofOfTask string, data string, taskDefinitionId int) {
	privateKey, err := crypto.HexToECDSA(config.PrivateKeyPerformer)
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
		common.HexToAddress(performerAddress),
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

	client, err := rpc.Dial(config.OTHENTIC_CLIENT_RPC_ADDRESS)
	if err != nil {
		logger.Errorf("Error dialing RPC", "error", err)
	}

	params := Params{
		proofOfTask:      proofOfTask,
		data:             "0x" + hex.EncodeToString([]byte(data)),
		taskDefinitionId: taskDefinitionId,
		performerAddress: performerAddress,
		signature:        serializedSignature,
	}

	response := makeRPCRequest(client, params)
	logger.Infof("API response:", "response", response)
}

func makeRPCRequest(client *rpc.Client, params Params) interface{} {
	var result interface{}

	err := client.Call(&result, "sendTask", params.proofOfTask, params.data, params.taskDefinitionId, params.performerAddress, params.signature)
	if err != nil {
		logger.Errorf("Error making RPC request", "error", err)
	}
	return result
}

func ConnectToTaskManager(keeperAddress string, connectionAddress string) (bool, error) {
	taskManagerIPAddress := os.Getenv("TASK_MANAGER_IP_ADDRESS")
	taskManagerPort := os.Getenv("TASK_MANAGER_RPC_PORT")
	if taskManagerIPAddress == "" || taskManagerPort == "" {
		return false, fmt.Errorf("values missing in .env file")
	}

	taskManagerRPCAddress := fmt.Sprintf("http://%s:%s/connect", taskManagerIPAddress, taskManagerPort)

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

func SetupTunnel(port string, keeperAddress string) (string, error) {
	logger.Info("Setting up localtunnel for port", "port", port)

	// Convert port string to int
	portInt := 0
	_, err := fmt.Sscanf(port, "%d", &portInt)
	if err != nil {
		return "", fmt.Errorf("invalid port format: %w", err)
	}

	// Create a localtunnel client
	tunnel, err := localtunnel.New(portInt, "localhost", localtunnel.Options{
		Subdomain:      keeperAddress,
		MaxConnections: 10,
	})
	if err != nil {
		return "", fmt.Errorf("failed to create localtunnel: %w", err)
	}

	// Get the tunnel URL
	tunnelURL := tunnel.URL()

	// Ensure the URL has the http:// prefix
	if !strings.HasPrefix(tunnelURL, "http://") && !strings.HasPrefix(tunnelURL, "https://") {
		tunnelURL = "https://" + tunnelURL
	}

	logger.Info("Localtunnel established", "url", tunnelURL)

	// The tunnel starts automatically when created with New()
	// We'll keep a reference to close it later if needed
	// This could be stored in a package variable or context

	return tunnelURL, nil
}

func GetPublicIP() (string, error) {
	ipServices := []string{
		"https://api.ipify.org",
		"https://ifconfig.me/ip",
		"https://api.ipify.org?format=text",
		"https://checkip.amazonaws.com",
	}

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	var lastErr error
	for _, service := range ipServices {
		resp, err := client.Get(service)
		if err != nil {
			lastErr = err
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			lastErr = fmt.Errorf("service %s returned status: %d", service, resp.StatusCode)
			continue
		}

		ipBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			lastErr = err
			continue
		}

		ip := strings.TrimSpace(string(ipBytes))
		if ip != "" {
			return ip, nil
		}
	}

	return "", fmt.Errorf("failed to get public IP: %v", lastErr)
}
