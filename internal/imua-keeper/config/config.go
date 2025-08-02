package config

import (
	"fmt"
	"crypto/ecdsa"
	"time"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	sdkEcdsa "github.com/imua-xyz/imua-avs-sdk/crypto/ecdsa"
	"github.com/imua-xyz/imua-avs-sdk/crypto/bls"
	blscommon "github.com/prysmaticlabs/prysm/v5/crypto/bls/common"

	"github.com/trigg3rX/triggerx-backend/pkg/env"
)

const (
	version = "0.1.6"
	isImua  = true
)

const (
	avsName = "hello-world-avs-demo"
	semVer  = "0.0.1"
	maxRetries = 80
	retryDelay = 1 * time.Second
)

type Config struct {
	devMode bool

	// RPC URLs for Ethereum and Base
	ethRPCUrl  string
	ethWsUrl   string

	// API Keys for Alchemy and Etherscan
	alchemyAPIKey   string
	etherscanAPIKey string

	// Controller Key and Keeper Address
	privateKeyController *ecdsa.PrivateKey
	keeperAddress        string

	// Consensus Key and Address (BLS)
	consensusKeyPair blscommon.SecretKey

	// Public IP Address and Peer ID
	publicIPV4Address string
	peerID            string

	// Ports for Keeper API server, P2P connections, metrics and Grafana
	keeperRPCPort     string
	keeperP2PPort     string
	keeperMetricsPort string
	grafanaPort       string
	nodeApiPort       string

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

	l1Chain string
	l2Chain string

	// AVS Contract Address
	avsGovernanceAddress string
	taskExecutionAddress string

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
		ethRPCUrl:                env.GetEnvString("ETH_RPC_URL", ""),
		ethWsUrl:                 env.GetEnvString("ETH_WS_URL", ""),
		alchemyAPIKey:           env.GetEnvString("ALCHEMY_API_KEY", ""),
		publicIPV4Address:        env.GetEnvString("PUBLIC_IPV4_ADDRESS", ""),
		peerID:                   env.GetEnvString("PEER_ID", ""),
		keeperRPCPort:            env.GetEnvString("OPERATOR_RPC_PORT", "9011"),
		keeperP2PPort:            env.GetEnvString("OPERATOR_P2P_PORT", "9012"),
		keeperMetricsPort:        env.GetEnvString("OPERATOR_METRICS_PORT", "9013"),
		nodeApiPort:              env.GetEnvString("OPERATOR_NODE_API_PORT", "9014"),
		grafanaPort:              env.GetEnvString("GRAFANA_PORT", "3000"),
		aggregatorRPCUrl:         env.GetEnvString("OTHENTIC_CLIENT_RPC_ADDRESS", "https://aggregator.triggerx.network"),
		healthRPCUrl:             env.GetEnvString("HEALTH_IP_ADDRESS", "https://health.triggerx.network"),
		tlsProofHost:             "www.google.com",
		tlsProofPort:             "443",
		l1Chain:                  env.GetEnvString("L1_CHAIN", "17000"),
		l2Chain:                  env.GetEnvString("L2_CHAIN", "84532"),
		avsGovernanceAddress:     env.GetEnvString("TRIGGERX_AVS_ADDRESS", "0x72A5016ECb9EB01d7d54ae48bFFB62CA0B8e57a5"),
		othenticBootstrapID:      env.GetEnvString("OTHENTIC_BOOTSTRAP_ID", "12D3KooWBNFG1QjuF3UKAKvqhdXcxh9iBmj88cM5eU2EK5Pa91KB"),
	}

	blsKeyPassword := env.GetEnvString("OPERATOR_BLS_KEY_PASSWORD", "")
	blsKeyStorePath := env.GetEnvString("OPERATOR_BLS_KEY_STORE_PATH", "")
	blsKeyPair, err := bls.ReadPrivateKeyFromFile(blsKeyStorePath, blsKeyPassword)
	if err != nil {
		return fmt.Errorf("invalid bls key password: %s", err)
	}
	cfg.consensusKeyPair = blsKeyPair

	ecdsaKeyPassword := env.GetEnvString("OPERATOR_ECDSA_KEY_PASSWORD", "")
	ecdsaKeyStorePath := env.GetEnvString("OPERATOR_ECDSA_KEY_STORE_PATH", "")
	ecdsaPrivateKey, err := sdkEcdsa.ReadKey(ecdsaKeyStorePath, ecdsaKeyPassword)
	if err != nil {
		return fmt.Errorf("invalid ecdsa key password: %s", err)
	}
	cfg.privateKeyController = ecdsaPrivateKey

	ecdsaAddress, err := sdkEcdsa.GetAddressFromKeyStoreFile(ecdsaKeyStorePath)
	if err != nil {
		return fmt.Errorf("invalid ecdsa key password: %s", err)
	}
	cfg.keeperAddress = ecdsaAddress.Hex()
	
	if err := validateConfig(cfg); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}
	if !cfg.devMode {
		gin.SetMode(gin.ReleaseMode)
	}
	// isRegistered := checkKeeperRegistration()
	// if !isRegistered {
	// 	log.Println("Keeper address is not yet registered on L2. Please register the address before continuing. If registered, please wait for the registration to be confirmed.")
	// 	log.Fatal("Keeper address is not registered on L2")
	// }
	return nil
}

func validateConfig(cfg Config) error {
	if env.IsEmpty(cfg.ethRPCUrl) {
		return fmt.Errorf("invalid eth rpc url: %s", cfg.ethRPCUrl)
	}
	if env.IsEmpty(cfg.ethWsUrl) {
		return fmt.Errorf("invalid eth ws url: %s", cfg.ethWsUrl)
	}
	if !env.IsValidIPAddress(cfg.publicIPV4Address) {
		return fmt.Errorf("invalid public ipv4 address: %s", cfg.publicIPV4Address)
	}
	// if !env.IsValidEthAddress(cfg.keeperAddress) {
	// 	return fmt.Errorf("invalid keeper address: %s", cfg.keeperAddress)
	// }
	if !env.IsValidPeerID(cfg.peerID) {
		return fmt.Errorf("invalid peer id: %s", cfg.peerID)
	}
	return nil
}

func GetEthRPCUrl() string {
	return cfg.ethRPCUrl
}

func GetEthWsUrl() string {
	return cfg.ethWsUrl
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

func GetPrivateKeyController() *ecdsa.PrivateKey {
	return cfg.privateKeyController
}

func GetKeeperAddress() string {
	return cfg.keeperAddress
}

func GetConsensusKeyPair() blscommon.SecretKey {
	return cfg.consensusKeyPair
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

func GetOperatorNodeApiPort() string {
	return cfg.nodeApiPort
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

func GetAvsGovernanceAddress() string {
	return cfg.avsGovernanceAddress
}

func GetVersion() string {
	return version
}

func GetAvsName() string {
	return avsName
}

func GetSemVer() string {
	return semVer
}

func GetMaxRetries() int {
	return maxRetries
}

func GetRetryDelay() time.Duration {
	return retryDelay
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
