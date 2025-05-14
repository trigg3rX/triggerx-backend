package config

import (
	"context"
	"math/big"
	"os"
	"regexp"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

const AttestationCenterABI = `[{
  "inputs": [
    {
      "internalType": "address",
      "name": "_operator",
      "type": "address"
    }
  ],
  "name": "operatorsIdsByAddress",
  "outputs": [
    {
      "internalType": "uint256",
      "name": "",
      "type": "uint256"
    }
  ],
  "stateMutability": "view",
  "type": "function"
}]`

var (
	logger = logging.GetLogger(logging.Development, logging.KeeperProcess)

	EthRPCUrl            string
	AlchemyAPIKey        string
	PrivateKeyConsensus  string
	PrivateKeyController string
	KeeperAddress        string
	KeeperRPCPort        string
	PublicIPV4Address    string
	PeerID               string
	KeeperP2PPort        string
	KeeperMetricsPort    string
	GrafanaPort          string
	L2RPC                string

	PinataApiKey             string
	PinataSecretApiKey       string
	IpfsHost                 string
	AggregatorIPAddress      string
	HealthIPAddress          string
	L1Chain                  string
	L2Chain                  string
	AVSGovernanceAddress     string
	AttestationCenterAddress string
	OthenticBootstrapID      string
)

func validateHexAddress(address string) bool {
	matched, _ := regexp.MatchString("^0x[0-9a-fA-F]{40}$", address)
	return matched
}

func validatePrivateKey(key string) bool {
	matched, _ := regexp.MatchString("^[0-9a-fA-F]{64}$", key)
	return matched
}

func validatePeerID(peerID string) bool {
	if len(peerID) < 46 {
		return false
	}
	return true
}

func validateIPAddress(ip string) bool {
	ipPattern := `^((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$`
	matched, _ := regexp.MatchString(ipPattern, ip)
	return matched
}

func validatePort(port string) bool {
	matched, _ := regexp.MatchString("^([1-9][0-9]{0,3}|[1-5][0-9]{4}|6[0-4][0-9]{3}|65[0-4][0-9]{2}|655[0-2][0-9]|6553[0-5])$", port)
	return matched
}

func validateRPCUrl(url string) bool {
	if url == "" {
		return false
	}

	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return false
	}

	return true
}

func checkKeeperRegistration() bool {
	client, err := ethclient.Dial(L2RPC)
	if err != nil {
		logger.Error("Failed to connect to L2 network", "error", err)
		return false
	}
	defer client.Close()

	parsedABI, err := abi.JSON(strings.NewReader(AttestationCenterABI))
	if err != nil {
		logger.Error("Failed to parse AttestationCenter ABI", "error", err)
		return false
	}

	keeperAddr := common.HexToAddress(KeeperAddress)
	data, err := parsedABI.Pack("operatorsIdsByAddress", keeperAddr)
	if err != nil {
		logger.Error("Failed to pack function call data", "error", err)
		return false
	}

	attestationCenterAddr := common.HexToAddress(AttestationCenterAddress)
	result, err := client.CallContract(context.Background(), ethereum.CallMsg{
		To:   &attestationCenterAddr,
		Data: data,
	}, nil)
	if err != nil {
		logger.Error("Failed to call AttestationCenter contract", "error", err)
		return false
	}

	if len(result) == 0 {
		logger.Error("Empty result from contract call")
		return false
	}

	operatorID := new(big.Int).SetBytes(result)

	if operatorID.Cmp(big.NewInt(0)) == 0 {
		logger.Error("Keeper address is not registered on L2", "address", KeeperAddress)
		return false
	}

	// logger.Info("Keeper address is registered on L2")
	return true
}

func Init() {
	err := godotenv.Load()
	if err != nil {
		logger.Error("Error loading .env file", "error", err)
	}

	loadAndValidateEnvVars()

	if !checkKeeperRegistration() {
		logger.Fatal("Keeper address is not registered on L2. Please register the address before continuing.")
	}

	gin.SetMode(gin.ReleaseMode)
}

func loadAndValidateEnvVars() {
	PinataApiKey = os.Getenv("PINATA_API_KEY")
	PinataSecretApiKey = os.Getenv("PINATA_SECRET_API_KEY")
	IpfsHost = os.Getenv("IPFS_HOST")
	AggregatorIPAddress = os.Getenv("OTHENTIC_CLIENT_RPC_ADDRESS")
	HealthIPAddress = os.Getenv("HEALTH_IP_ADDRESS")
	L1Chain = os.Getenv("L1_CHAIN")
	L2Chain = os.Getenv("L2_CHAIN")
	AVSGovernanceAddress = os.Getenv("AVS_GOVERNANCE_ADDRESS")
	AttestationCenterAddress = os.Getenv("ATTESTATION_CENTER_ADDRESS")
	OthenticBootstrapID = os.Getenv("OTHENTIC_BOOTSTRAP_ID")

	if PinataApiKey == "" || PinataSecretApiKey == "" || IpfsHost == "" ||
		AggregatorIPAddress == "" || HealthIPAddress == "" || L1Chain == "" ||
		L2Chain == "" || AVSGovernanceAddress == "" ||
		AttestationCenterAddress == "" || OthenticBootstrapID == "" {
		logger.Fatal("Required environment variables are missing in .env file")
	}

	loadAndValidateUserEnvVars()
}

func loadAndValidateUserEnvVars() {
	EthRPCUrl = os.Getenv("L1_RPC")
	if !validateRPCUrl(EthRPCUrl) {
		logger.Fatal("L1_RPC URL is missing or invalid. Please set a valid Ethereum RPC URL")
	}

	L2RPC = os.Getenv("L2_RPC")
	if !validateRPCUrl(L2RPC) {
		logger.Fatal("L2_RPC URL is missing or invalid. Please set a valid L2 RPC URL")
	}

	if strings.Contains(EthRPCUrl, "alchemy.com") {
		AlchemyAPIKey = strings.TrimPrefix(EthRPCUrl, "https://eth-holesky.g.alchemy.com/v2/")
	}

	PrivateKeyConsensus = os.Getenv("PRIVATE_KEY")
	if PrivateKeyConsensus == "" {
		logger.Fatal("PRIVATE_KEY is missing in .env file")
	}

	if strings.HasPrefix(PrivateKeyConsensus, "0x") {
		PrivateKeyConsensus = strings.TrimPrefix(PrivateKeyConsensus, "0x")
	}
	if !validatePrivateKey(PrivateKeyConsensus) {
		logger.Fatal("PRIVATE_KEY is invalid. It should be 64 hex characters without 0x prefix")
	}

	PrivateKeyController = os.Getenv("OPERATOR_PRIVATE_KEY")

	KeeperAddress = os.Getenv("OPERATOR_ADDRESS")
	if KeeperAddress == "" {
		logger.Fatal("OPERATOR_ADDRESS is missing in .env file")
	}
	if !validateHexAddress(KeeperAddress) {
		logger.Fatal("OPERATOR_ADDRESS is invalid. It should be a valid Ethereum address (0x followed by 40 hex characters)")
	}

	PublicIPV4Address = os.Getenv("PUBLIC_IPV4_ADDRESS")
	if PublicIPV4Address == "" {
		logger.Fatal("PUBLIC_IPV4_ADDRESS is missing in .env file")
	}
	if !validateIPAddress(PublicIPV4Address) {
		logger.Fatal("PUBLIC_IPV4_ADDRESS is invalid. It should be a valid IPv4 address")
	}

	PeerID = os.Getenv("PEER_ID")
	if PeerID == "" {
		logger.Fatal("PEER_ID is missing in .env file")
	}
	if !validatePeerID(PeerID) {
		logger.Fatal("PEER_ID is invalid. It should be at least 46 characters long")
	}

	KeeperRPCPort = os.Getenv("OPERATOR_RPC_PORT")
	if KeeperRPCPort == "" {
		KeeperRPCPort = "9005"
	} else if !validatePort(KeeperRPCPort) {
		logger.Fatal("OPERATOR_RPC_PORT is invalid. It should be a number between 1 and 65535")
	}

	KeeperP2PPort = os.Getenv("OPERATOR_P2P_PORT")
	if KeeperP2PPort == "" {
		KeeperP2PPort = "9006"
	} else if !validatePort(KeeperP2PPort) {
		logger.Fatal("OPERATOR_P2P_PORT is invalid. It should be a number between 1 and 65535")
	}

	KeeperMetricsPort = os.Getenv("OPERATOR_METRICS_PORT")
	if KeeperMetricsPort == "" {
		KeeperMetricsPort = "9009"
	} else if !validatePort(KeeperMetricsPort) {
		logger.Fatal("OPERATOR_METRICS_PORT is invalid. It should be a number between 1 and 65535")
	}

	GrafanaPort = os.Getenv("GRAFANA_PORT")
	if GrafanaPort == "" {
		GrafanaPort = "3000"
	} else if !validatePort(GrafanaPort) {
		logger.Fatal("GRAFANA_PORT is invalid. It should be a number between 1 and 65535")
	}
	
}
