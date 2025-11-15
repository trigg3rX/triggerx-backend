package config

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/trigg3rX/triggerx-backend/pkg/env"
)

const (
	version = "1.0.1"
	isImua  = false
)

type Config struct {
	devMode bool

	// RPC URLs for Ethereum and Base
	ethRPCUrl  string
	baseRPCUrl string

	// API Keys for Alchemy and Etherscan
	alchemyAPIKey   string
	etherscanAPIKey string

	// Controller Key and Keeper Address
	privateKeyController string
	keeperAddress        string

	// Consensus Key and Address
	privateKeyConsensus string
	consensusAddress    string

	// Public IP Address and Peer ID
	publicIPV4Address string
	peerID            string

	// Ports for Keeper API server, P2P connections, metrics and Grafana
	keeperRPCPort     string
	keeperP2PPort     string
	keeperMetricsPort string
	grafanaPort       string

	// IPFS configuration
	ipfsHost  string
	pinataJWT string

	// TLS Proof configuration
	tlsProofHost string
	tlsProofPort string

	// Manager Signing Address
	managerSigningAddress string

	// Backend Service URLs
	aggregatorRPCUrl string
	healthRPCUrl     string
	taskMonitorRPCUrl string

	l1Chain string
	l2Chain string

	// AVS Contract Address
	avsGovernanceAddress     string
	attestationCenterAddress string
	taskExecutionAddress     string

	// Othentic Bootstrap ID
	othenticBootstrapID string
}

var cfg Config

func Init() error {
	if err := godotenv.Load(); err != nil {
		return fmt.Errorf("error loading .env file: %w", err)
	}
	cfg = Config{
		devMode:                  env.GetEnvBool("DEV_MODE", false),
		ethRPCUrl:                env.GetEnvString("L1_RPC", ""),
		baseRPCUrl:               env.GetEnvString("L2_RPC", ""),
		privateKeyConsensus:      env.GetEnvString("PRIVATE_KEY", ""),
		privateKeyController:     env.GetEnvString("OPERATOR_PRIVATE_KEY", ""),
		keeperAddress:            env.GetEnvString("OPERATOR_ADDRESS", ""),
		consensusAddress:         crypto.PubkeyToAddress(crypto.ToECDSAUnsafe(common.FromHex(env.GetEnvString("PRIVATE_KEY", ""))).PublicKey).Hex(),
		publicIPV4Address:        env.GetEnvString("PUBLIC_IPV4_ADDRESS", ""),
		peerID:                   env.GetEnvString("PEER_ID", ""),
		keeperRPCPort:            env.GetEnvString("OPERATOR_RPC_PORT", "9011"),
		keeperP2PPort:            env.GetEnvString("OPERATOR_P2P_PORT", "9012"),
		keeperMetricsPort:        env.GetEnvString("OPERATOR_METRICS_PORT", "9013"),
		grafanaPort:              env.GetEnvString("GRAFANA_PORT", "3000"),
		aggregatorRPCUrl:         env.GetEnvString("OTHENTIC_CLIENT_RPC_ADDRESS", "https://aggregator.triggerx.network"),
		healthRPCUrl:             env.GetEnvString("HEALTH_IP_ADDRESS", "https://health.triggerx.network"),
		taskMonitorRPCUrl:        env.GetEnvString("TASK_MONITOR_RPC_URL", "https://task.triggerx.network"),
		tlsProofHost:             "www.google.com",
		tlsProofPort:             "443",
		// l1Chain:                  env.GetEnvString("L1_CHAIN", "11155111"),
		// l2Chain:                  env.GetEnvString("L2_CHAIN", "84532"),
		// avsGovernanceAddress:     env.GetEnvString("TEST_AVS_GOVERNANCE_ADDRESS", "0xaaE90bE86cec5E6c34D584917FFfCE7C379fFEE1"),
		// attestationCenterAddress: env.GetEnvString("TEST_ATTESTATION_CENTER_ADDRESS", "0x21B099554F6D27E47D57991D2B44251DaFa9323b"),
		l1Chain:                  env.GetEnvString("L1_CHAIN", "1"),
		l2Chain:                  env.GetEnvString("L2_CHAIN", "8453"),
		avsGovernanceAddress:     env.GetEnvString("AVS_GOVERNANCE_ADDRESS", "0x875B5ff698B74B26f39C223c4996871F28AcDdea"),
		attestationCenterAddress: env.GetEnvString("ATTESTATION_CENTER_ADDRESS", "0x6DFee10D13d5B43AaF97bDA908C1D76d4313aF5f"),
		othenticBootstrapID:      env.GetEnvString("OTHENTIC_BOOTSTRAP_ID", "12D3KooWBNFG1QjuF3UKAKvqhdXcxh9iBmj88cM5eU2EK5Pa91KB"),
	}
	if err := validateConfig(cfg); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}
	if !cfg.devMode {
		gin.SetMode(gin.ReleaseMode)
	}
	isRegistered := checkKeeperRegistration()
	if !isRegistered {
		log.Println("Keeper address is not yet registered on L2. Please register the address before continuing. If registered, please wait for the registration to be confirmed.")
		log.Fatal("Keeper address is not registered on L2")
	}
	return nil
}

func validateConfig(cfg Config) error {
	if env.IsEmpty(cfg.ethRPCUrl) {
		return fmt.Errorf("invalid eth rpc url: %s", cfg.ethRPCUrl)
	}
	if env.IsEmpty(cfg.baseRPCUrl) {
		return fmt.Errorf("invalid base rpc url: %s", cfg.baseRPCUrl)
	}
	if !env.IsValidIPAddress(cfg.publicIPV4Address) {
		return fmt.Errorf("invalid public ipv4 address: %s", cfg.publicIPV4Address)
	}
	if !env.IsValidPrivateKey(cfg.privateKeyConsensus) {
		return fmt.Errorf("invalid private key consensus: %s", cfg.privateKeyConsensus)
	}
	if !env.IsValidEthAddress(cfg.keeperAddress) {
		return fmt.Errorf("invalid keeper address: %s", cfg.keeperAddress)
	}
	// if !env.IsValidPeerID(cfg.peerID) {
	// 	return fmt.Errorf("invalid peer id: %s", cfg.peerID)
	// }
	return nil
}

func GetEthRPCUrl() string {
	return cfg.ethRPCUrl
}

func GetBaseRPCUrl() string {
	return cfg.baseRPCUrl
}

// Only sets it if there was no key in env file
func SetAlchemyAPIKey(key string) {
	if !env.IsEmpty(cfg.alchemyAPIKey) {
		return
	}
	cfg.alchemyAPIKey = key
}

func GetAlchemyAPIKey() string {
	return cfg.alchemyAPIKey
}

func SetEtherscanAPIKey(key string) {
	cfg.etherscanAPIKey = key
}

func GetEtherscanAPIKey() string {
	return cfg.etherscanAPIKey
}

func GetPrivateKeyConsensus() string {
	return cfg.privateKeyConsensus
}

func GetPrivateKeyController() string {
	return cfg.privateKeyController
}

func GetKeeperAddress() string {
	return cfg.keeperAddress
}

func GetConsensusAddress() string {
	return cfg.consensusAddress
}

func GetPublicIPV4Address() string {
	return cfg.publicIPV4Address
}

func GetPeerID() string {
	return cfg.peerID
}

func GetOperatorRPCPort() string {
	return cfg.keeperRPCPort
}

func IsDevMode() bool {
	return cfg.devMode
}

func GetKeeperRPCPort() string {
	return cfg.keeperRPCPort
}

func GetAggregatorRPCUrl() string {
	return cfg.aggregatorRPCUrl
}

func GetHealthRPCUrl() string {
	return cfg.healthRPCUrl
}

func GetTaskMonitorRPCUrl() string {
	return cfg.taskMonitorRPCUrl
}

func GetAvsGovernanceAddress() string {
	return cfg.avsGovernanceAddress
}

func GetAttestationCenterAddress() string {
	return cfg.attestationCenterAddress
}

func GetVersion() string {
	return version
}

func IsImua() bool {
	return isImua
}

// IPFS configuration
func SetIpfsHost(host string) {
	cfg.ipfsHost = host
}

func GetIpfsHost() string {
	return cfg.ipfsHost
}

func SetPinataJWT(jwt string) {
	cfg.pinataJWT = jwt
}

func GetPinataJWT() string {
	return cfg.pinataJWT
}

// TLS Proof configuration
func SetTLSProofHost(host string) {
	cfg.tlsProofHost = host
}

func SetTLSProofPort(port string) {
	cfg.tlsProofPort = port
}

func GetTLSProofHost() string {
	return cfg.tlsProofHost
}

func GetTLSProofPort() string {
	return cfg.tlsProofPort
}

// Manager Signing Address
func SetManagerSigningAddress(addr string) {
	cfg.managerSigningAddress = addr
}

func GetManagerSigningAddress() string {
	return cfg.managerSigningAddress
}

func SetTaskExecutionAddress(addr string) {
	cfg.taskExecutionAddress = addr
}

func GetTaskExecutionAddress() string {
	return cfg.taskExecutionAddress
}

// SetKeeperAddress sets the keeper address in the config (for testing)
func SetKeeperAddress(addr string) {
	cfg.keeperAddress = addr
}
