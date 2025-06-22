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

	// Backend Service URLs
	aggregatorRPCUrl string
	healthRPCUrl     string

	l1Chain string
	l2Chain string

	// AVS Contract Address
	avsGovernanceAddress     string
	attestationCenterAddress string

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
		ethRPCUrl:                env.GetEnv("L1_RPC", ""),
		baseRPCUrl:               env.GetEnv("L2_RPC", ""),
		alchemyAPIKey:           env.GetEnv("ALCHEMY_API_KEY", ""),
		privateKeyConsensus:      env.GetEnv("PRIVATE_KEY", ""),
		privateKeyController:     env.GetEnv("OPERATOR_PRIVATE_KEY", ""),
		keeperAddress:            env.GetEnv("OPERATOR_ADDRESS", ""),
		consensusAddress:         crypto.PubkeyToAddress(crypto.ToECDSAUnsafe(common.FromHex(env.GetEnv("PRIVATE_KEY", ""))).PublicKey).Hex(),
		publicIPV4Address:        env.GetEnv("PUBLIC_IPV4_ADDRESS", ""),
		peerID:                   env.GetEnv("PEER_ID", ""),
		keeperRPCPort:            env.GetEnv("OPERATOR_RPC_PORT", "9011"),
		keeperP2PPort:            env.GetEnv("OPERATOR_P2P_PORT", "9012"),
		keeperMetricsPort:        env.GetEnv("OPERATOR_METRICS_PORT", "9013"),
		grafanaPort:              env.GetEnv("GRAFANA_PORT", "3000"),
		aggregatorRPCUrl:         env.GetEnv("OTHENTIC_CLIENT_RPC_ADDRESS", "https://aggregator.triggerx.network"),
		healthRPCUrl:             env.GetEnv("HEALTH_IP_ADDRESS", "https://health.triggerx.network"),
		tlsProofHost:             env.GetEnv("TLS_PROOF_HOST", "www.google.com"),
		tlsProofPort:             env.GetEnv("TLS_PROOF_PORT", "443"),
		l1Chain:                  env.GetEnv("L1_CHAIN", "17000"),
		l2Chain:                  env.GetEnv("L2_CHAIN", "84532"),
		avsGovernanceAddress:     env.GetEnv("AVS_GOVERNANCE_ADDRESS", "0x12f45551f11Df20b3EcBDf329138Bdc65cc58Ec0"),
		attestationCenterAddress: env.GetEnv("ATTESTATION_CENTER_ADDRESS", "0x9725fB95B5ec36c062A49ca2712b3B1ff66F04eD"),
		othenticBootstrapID:      env.GetEnv("OTHENTIC_BOOTSTRAP_ID", "12D3KooWBNFG1QjuF3UKAKvqhdXcxh9iBmj88cM5eU2EK5Pa91KB"),
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
	if !env.IsValidPeerID(cfg.peerID) {
		return fmt.Errorf("invalid peer id: %s", cfg.peerID)
	}
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

func GetAvsGovernanceAddress() string {
	return cfg.avsGovernanceAddress
}

func GetAttestationCenterAddress() string {
	return cfg.attestationCenterAddress
}

func GetVersion() string {
	return "0.1.5"
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
func GetTLSProofHost() string {
	return cfg.tlsProofHost
}

func GetTLSProofPort() string {
	return cfg.tlsProofPort
}

// SetKeeperAddress sets the keeper address in the config (for testing)
func SetKeeperAddress(addr string) {
	cfg.keeperAddress = addr
}
