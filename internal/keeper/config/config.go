package config

import (
	"fmt"
	"os"
	"strings"
	"regexp"
	
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

var (
	logger = logging.GetLogger(logging.Development, logging.KeeperProcess)

	// User Entered Information
	EthRPCUrl                   string
	AlchemyAPIKey               string
	PrivateKeyConsensus         string
	PrivateKeyController        string
	KeeperAddress               string
	KeeperRPCPort               string
	PublicIPV4Address           string
	PeerID                      string
	KeeperP2PPort               string
	KeeperMetricsPort           string
	GrafanaPort                 string
	L2RPC                       string
	// Provided Information
	PinataApiKey                string
	PinataSecretApiKey          string
	IpfsHost                    string
	AggregatorIPAddress         string
	HealthIPAddress             string
	L1Chain                     string
	L2Chain                     string
	AVSGovernanceAddress        string
	AttestationCenterAddress    string
	OthenticBootstrapID         string
)

// validateHexAddress ensures an Ethereum address is valid (0x followed by 40 hex chars)
func validateHexAddress(address string) bool {
	matched, _ := regexp.MatchString("^0x[0-9a-fA-F]{40}$", address)
	return matched
}

// validatePrivateKey ensures a private key is 64 hex chars (without 0x prefix)
func validatePrivateKey(key string) bool {
	// Private key should be 64 hex chars without 0x prefix
	matched, _ := regexp.MatchString("^[0-9a-fA-F]{64}$", key)
	return matched
}

// validatePeerID ensures the peer ID matches expected format
func validatePeerID(peerID string) bool {
	// PeerID should be at least 46 characters
	if len(peerID) < 46 {
		return false
	}
	return true
}

// validateIPAddress ensures the IP address is valid
func validateIPAddress(ip string) bool {
	ipPattern := `^((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$`
	matched, _ := regexp.MatchString(ipPattern, ip)
	return matched
}

// validatePort ensures the port is a number between 1 and 65535
func validatePort(port string) bool {
	matched, _ := regexp.MatchString("^([1-9][0-9]{0,3}|[1-5][0-9]{4}|6[0-4][0-9]{3}|65[0-4][0-9]{2}|655[0-2][0-9]|6553[0-5])$", port)
	return matched
}

// validateRPCUrl ensures the RPC URL is valid
func validateRPCUrl(url string) bool {
	if url == "" {
		return false
	}
	
	// Basic check for http(s):// prefix
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return false
	}
	
	return true
}

func Init() {
	err := godotenv.Load()
	if err != nil {
		logger.Error("Error loading .env file", "error", err)
	}

	// Load and validate required environment variables
	loadAndValidateEnvVars()

	gin.SetMode(gin.ReleaseMode)
}

func loadAndValidateEnvVars() {
	// Load provided information first
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

	// Check required provided variables
	if PinataApiKey == "" || PinataSecretApiKey == "" || IpfsHost == "" || 
	   AggregatorIPAddress == "" || HealthIPAddress == "" || L1Chain == "" || 
	   L2Chain == "" || AVSGovernanceAddress == "" || 
	   AttestationCenterAddress == "" || OthenticBootstrapID == "" {
		logger.Fatal("Required environment variables are missing in .env file")
	}

	// Validate and load user entered information
	loadAndValidateUserEnvVars()
}

func loadAndValidateUserEnvVars() {
	// RPC URLs
	EthRPCUrl = os.Getenv("L1_RPC")
	if !validateRPCUrl(EthRPCUrl) {
		logger.Fatal("L1_RPC URL is missing or invalid. Please set a valid Ethereum RPC URL")
	}
	
	L2RPC = os.Getenv("L2_RPC")
	if !validateRPCUrl(L2RPC) {
		logger.Fatal("L2_RPC URL is missing or invalid. Please set a valid L2 RPC URL")
	}

	// Extract Alchemy API Key if present
	if strings.Contains(EthRPCUrl, "alchemy.com") {
		AlchemyAPIKey = strings.TrimPrefix(EthRPCUrl, "https://eth-holesky.g.alchemy.com/v2/")
	}

	// Private keys
	PrivateKeyConsensus = os.Getenv("PRIVATE_KEY")
	if PrivateKeyConsensus == "" {
		logger.Fatal("PRIVATE_KEY is missing in .env file")
	}
	// Remove 0x prefix if present
	if strings.HasPrefix(PrivateKeyConsensus, "0x") {
		PrivateKeyConsensus = strings.TrimPrefix(PrivateKeyConsensus, "0x")
	}
	if !validatePrivateKey(PrivateKeyConsensus) {
		logger.Fatal("PRIVATE_KEY is invalid. It should be 64 hex characters without 0x prefix")
	}

	PrivateKeyController = os.Getenv("OPERATOR_PRIVATE_KEY")
	// Keeper Address
	KeeperAddress = os.Getenv("OPERATOR_ADDRESS")
	if KeeperAddress == "" {
		logger.Fatal("OPERATOR_ADDRESS is missing in .env file")
	}
	if !validateHexAddress(KeeperAddress) {
		logger.Fatal("OPERATOR_ADDRESS is invalid. It should be a valid Ethereum address (0x followed by 40 hex characters)")
	}

	// P2P Configuration
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

	// Port configurations
	KeeperRPCPort = os.Getenv("OPERATOR_RPC_PORT")
	if KeeperRPCPort == "" {
		KeeperRPCPort = "9005" // Default
	} else if !validatePort(KeeperRPCPort) {
		logger.Fatal("OPERATOR_RPC_PORT is invalid. It should be a number between 1 and 65535")
	}

	KeeperP2PPort = os.Getenv("OPERATOR_P2P_PORT")
	if KeeperP2PPort == "" {
		KeeperP2PPort = "9006" // Default
	} else if !validatePort(KeeperP2PPort) {
		logger.Fatal("OPERATOR_P2P_PORT is invalid. It should be a number between 1 and 65535")
	}

	KeeperMetricsPort = os.Getenv("OPERATOR_METRICS_PORT")
	if KeeperMetricsPort == "" {
		KeeperMetricsPort = "9009" // Default
	} else if !validatePort(KeeperMetricsPort) {
		logger.Fatal("OPERATOR_METRICS_PORT is invalid. It should be a number between 1 and 65535")
	}

	GrafanaPort = os.Getenv("GRAFANA_PORT")
	if GrafanaPort == "" {
		GrafanaPort = "3000" // Default
	} else if !validatePort(GrafanaPort) {
		logger.Fatal("GRAFANA_PORT is invalid. It should be a number between 1 and 65535")
	}

	// Log configuration summary
	logConfigSummary()
}

func logConfigSummary() {
	logger.Info("Configuration loaded successfully")
	logger.Info(fmt.Sprintf("Operator Address: %s", KeeperAddress))
	logger.Info(fmt.Sprintf("P2P Configuration: %s:%s (Peer ID: %s...)", 
		PublicIPV4Address, KeeperP2PPort, PeerID))
	logger.Info(fmt.Sprintf("RPC Port: %s, Metrics Port: %s", KeeperRPCPort, KeeperMetricsPort))
	logger.Info(fmt.Sprintf("Grafana Port: %s", GrafanaPort))
}