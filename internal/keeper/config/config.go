package config

import (
	"fmt"
	// "log"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/trigg3rX/triggerx-backend-imua/pkg/env"
)

type Config struct {
	devMode bool

	// API Keys for Alchemy and Etherscan
	alchemyAPIKey   string
	etherscanAPIKey string

	// Controller Key and Keeper Address
	privateKeyController string
	keeperAddress      string

	// Public IP Address and Peer ID
	publicIPV4Address string
	peerID            string

	// Ports for Keeper API server and P2P connections
	keeperRPCPort string
	keeperP2PPort string

	// IPFS configuration
	ipfsHost  string
	pinataJWT string

	// TLS Proof configuration
	tlsProofHost string
	tlsProofPort string

	// Backend Service URLs
	aggregatorRPCUrl string
	healthRPCUrl     string

	// AVS Contract Address
	triggerxAvsContract string

	// Imua Protocol Configuration
	ethRpcUrl                 string
	ethWsUrl                  string
	blsPrivateKeyStorePath    string
}

var cfg Config

func Init() error {
	if err := godotenv.Load(); err != nil {
		return fmt.Errorf("error loading .env file: %w", err)
	}
	cfg = Config{
		devMode:                   env.GetEnvBool("DEV_MODE", false),
		alchemyAPIKey:             env.GetEnvString("ALCHEMY_API_KEY", ""),
		etherscanAPIKey:           env.GetEnvString("ETHERSCAN_API_KEY", ""),
		privateKeyController:      env.GetEnvString("PRIVATE_KEY_CONTROLLER", ""),
		keeperAddress:           env.GetEnvString("OPERATOR_ADDRESS", ""),
		publicIPV4Address:         env.GetEnvString("PUBLIC_IPV4_ADDRESS", ""),
		peerID:                    env.GetEnvString("PEER_ID", ""),
		keeperRPCPort:             env.GetEnvString("KEEPER_RPC_PORT", ""),
		keeperP2PPort:             env.GetEnvString("KEEPER_P2P_PORT", ""),
		triggerxAvsContract:       env.GetEnvString("TRIGGERX_AVS_CONTRACT", ""),
		aggregatorRPCUrl:          env.GetEnvString("AGGREGATOR_RPC_URL", ""),
		healthRPCUrl:              env.GetEnvString("HEALTH_RPC_URL", ""),
		ethRpcUrl:                 env.GetEnvString("ETH_RPC_URL", ""),
		ethWsUrl:                  env.GetEnvString("ETH_WS_URL", ""),
		blsPrivateKeyStorePath:    env.GetEnvString("BLS_PRIVATE_KEY_STORE_PATH", ""),
	}
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
	if env.IsEmpty(cfg.alchemyAPIKey) {
		return fmt.Errorf("ALCHEMY_API_KEY is empty")
	}
	if !env.IsValidIPAddress(cfg.publicIPV4Address) {
		return fmt.Errorf("invalid public ipv4 address: %s", cfg.publicIPV4Address)
	}
	if !env.IsValidPrivateKey(cfg.privateKeyController) {
		return fmt.Errorf("invalid private key controller: %s", cfg.privateKeyController)
	}
	if !env.IsValidEthAddress(cfg.keeperAddress) {
		return fmt.Errorf("invalid operator address: %s", cfg.keeperAddress)
	}
	if !env.IsValidEthAddress(cfg.triggerxAvsContract) {
		return fmt.Errorf("invalid triggerx avs contract: %s", cfg.triggerxAvsContract)
	}
	if env.IsEmpty(cfg.ethRpcUrl) {
		return fmt.Errorf("ETH_RPC_URL is empty")
	}
	if env.IsEmpty(cfg.ethWsUrl) {
		return fmt.Errorf("ETH_WS_URL is empty")
	}
	return nil
}

func IsDevMode() bool {
	return cfg.devMode
}

func GetAlchemyAPIKey() string {
	return cfg.alchemyAPIKey
}

func GetEtherscanAPIKey() string {
	return cfg.etherscanAPIKey
}

func GetPrivateKeyController() string {
	return cfg.privateKeyController
}

func GetKeeperAddress() string {
	return cfg.keeperAddress
}

func GetPublicIPV4Address() string {
	return cfg.publicIPV4Address
}

func GetTriggerxAvsContract() string {
	return cfg.triggerxAvsContract
}

func GetAggregatorRPCUrl() string {
	return cfg.aggregatorRPCUrl
}

func GetHealthRPCUrl() string {
	return cfg.healthRPCUrl
}

func SetIPFSConfig(ipfsHost string, pinataJWT string) {
	cfg.ipfsHost = ipfsHost
	cfg.pinataJWT = pinataJWT
}

func GetIpfsHost() string {
	return cfg.ipfsHost
}

func GetPinataJWT() string {
	return cfg.pinataJWT
}

func SetTLSProofConfig(tlsProofHost string, tlsProofPort string) {
	cfg.tlsProofHost = tlsProofHost
	cfg.tlsProofPort = tlsProofPort
}

func GetTLSProofHost() string {
	return cfg.tlsProofHost
}

func GetTLSProofPort() string {
	return cfg.tlsProofPort
}

func GetEthRpcUrl() string {
	return cfg.ethRpcUrl
}

func GetEthWsUrl() string {
	return cfg.ethWsUrl
}

func GetBlsPrivateKeyStorePath() string {
	return cfg.blsPrivateKeyStorePath
}
