package services

import (
	"bytes"
	"context"
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
var activeTunnel *localtunnel.Listener
var tunnelServer *http.Server

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

	// Create a shorter, cleaner subdomain based on keeper address
	// Take the last 8 characters of the address to make it shorter
	addressSuffix := keeperAddress
	if len(keeperAddress) > 8 {
		addressSuffix = keeperAddress[len(keeperAddress)-8:]
	}

	// Create a unique subdomain with a timestamp to avoid conflicts
	subdomain := fmt.Sprintf("triggerx-%s-%d", strings.ToLower(addressSuffix), time.Now().Unix())

	// Remove any non-alphanumeric characters that might cause issues
	subdomain = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			return r
		}
		return '-'
	}, subdomain)

	logger.Info("Using subdomain for tunnel", "subdomain", subdomain)

	// Create a localtunnel listener
	listener, err := localtunnel.Listen(localtunnel.Options{
		Subdomain: subdomain,
	})

	if err != nil {
		logger.Warn("Failed to create tunnel with specific subdomain, trying without subdomain",
			"error", err,
			"subdomain", subdomain)

		// Try again without a specific subdomain
		listener, err = localtunnel.Listen(localtunnel.Options{})
		if err != nil {
			return "", fmt.Errorf("failed to create localtunnel listener: %w", err)
		}
	}

	// Store the listener for later cleanup
	activeTunnel = listener

	// Get the tunnel URL
	tunnelURL := listener.Addr().String()

	// Ensure the URL has the http:// prefix
	if !strings.HasPrefix(tunnelURL, "http://") && !strings.HasPrefix(tunnelURL, "https://") {
		tunnelURL = "https://" + tunnelURL
	}

	logger.Info("Localtunnel established", "url", tunnelURL)

	// Start a simple HTTP server to handle requests through the tunnel
	mux := http.NewServeMux()

	// Add a health endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		// Add response headers to help with tunnel password bypass
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("User-Agent", "TriggerX-Keeper-Service")
		w.Header().Set("bypass-tunnel-reminder", "true")

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":         "healthy",
			"keeper_address": config.KeeperAddress,
			"timestamp":      time.Now().Unix(),
		})
	})

	// Add a root endpoint for testing
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Add response headers to help with tunnel password bypass
		w.Header().Set("User-Agent", "TriggerX-Keeper-Service")
		w.Header().Set("bypass-tunnel-reminder", "true")

		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "TriggerX Keeper Service")
	})

	// Create and start the server
	tunnelServer = &http.Server{
		Handler: mux,
	}

	// Start the server in a goroutine
	go func() {
		logger.Info("Starting HTTP server on tunnel")
		if err := tunnelServer.Serve(listener); err != nil && err != http.ErrServerClosed {
			logger.Error("Tunnel server error", "error", err)
		}
	}()

	return tunnelURL, nil
}

// CloseTunnel closes the active tunnel if one exists
func CloseTunnel() {
	// First, close the HTTP server if it exists
	if tunnelServer != nil {
		logger.Info("Shutting down tunnel HTTP server")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := tunnelServer.Shutdown(ctx); err != nil {
			logger.Error("Error shutting down tunnel server", "error", err)
		}
		tunnelServer = nil
	}

	// Then close the tunnel listener
	if activeTunnel != nil {
		logger.Info("Closing localtunnel")
		if err := activeTunnel.Close(); err != nil {
			logger.Error("Error closing tunnel", "error", err)
		} else {
			logger.Info("Tunnel closed successfully")
		}
		activeTunnel = nil
	}
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
