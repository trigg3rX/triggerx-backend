package config

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/trigg3rX/triggerx-backend/pkg/validator"
)

type Config struct {
	EthRPCUrl       string
	BaseRPCUrl      string
	AlchemyAPIKey   string
	EtherscanAPIKey string

	PrivateKeyConsensus  string
	PrivateKeyController string
	KeeperAddress        string
	ConsensusAddress     string

	PublicIPV4Address string
	PeerID            string

	OperatorRPCPort string

	DevMode bool

	KeeperRPCPort     string
	KeeperP2PPort     string
	KeeperMetricsPort string
	GrafanaPort       string

	PinataApiKey       string
	PinataSecretApiKey string
	IpfsHost           string

	AggregatorRPCAddress string
	HealthRPCAddress     string

	L1Chain string
	L2Chain string

	AVSGovernanceAddress     string
	AttestationCenterAddress string

	OthenticBootstrapID string
}

var cfg Config

func Init() error {
	if err := godotenv.Load(); err != nil {
		return fmt.Errorf("error loading .env file: %w", err)
	}
	cfg = Config{
		DevMode:                  os.Getenv("DEV_MODE") == "true",
		EthRPCUrl:                os.Getenv("ETH_RPC_URL"),
		BaseRPCUrl:               os.Getenv("BASE_RPC_URL"),
		AlchemyAPIKey:            os.Getenv("ALCHEMY_API_KEY"),
		EtherscanAPIKey:          os.Getenv("ETHERSCAN_API_KEY"),
		PrivateKeyConsensus:      os.Getenv("PRIVATE_KEY"),
		KeeperAddress:            os.Getenv("OPERATOR_ADDRESS"),
		ConsensusAddress:         crypto.PubkeyToAddress(crypto.ToECDSAUnsafe(common.FromHex(os.Getenv("PRIVATE_KEY"))).PublicKey).Hex(),
		PublicIPV4Address:        os.Getenv("PUBLIC_IPV4_ADDRESS"),
		PeerID:                   os.Getenv("PEER_ID"),
		OperatorRPCPort:          os.Getenv("OPERATOR_RPC_PORT"),
		KeeperRPCPort:            os.Getenv("KEEPER_RPC_PORT"),
		KeeperP2PPort:            os.Getenv("KEEPER_P2P_PORT"),
		KeeperMetricsPort:        os.Getenv("KEEPER_METRICS_PORT"),
		GrafanaPort:              os.Getenv("GRAFANA_PORT"),
		PinataApiKey:             os.Getenv("PINATA_API_KEY"),
		PinataSecretApiKey:       os.Getenv("PINATA_SECRET_API_KEY"),
		IpfsHost:                 os.Getenv("IPFS_HOST"),
		AggregatorRPCAddress:     os.Getenv("OTHENTIC_CLIENT_RPC_ADDRESS"),
		HealthRPCAddress:         os.Getenv("HEALTH_RPC_ADDRESS"),
		L1Chain:                  os.Getenv("L1_CHAIN"),
		L2Chain:                  os.Getenv("L2_CHAIN"),
		AVSGovernanceAddress:     os.Getenv("AVS_GOVERNANCE_ADDRESS"),
		AttestationCenterAddress: os.Getenv("ATTESTATION_CENTER_ADDRESS"),
		OthenticBootstrapID:      os.Getenv("OTHENTIC_BOOTSTRAP_ID"),
	}

	if err := validateConfig(cfg); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	if !cfg.DevMode {
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
	if validator.IsEmpty(cfg.EthRPCUrl) {
		return fmt.Errorf("invalid eth rpc url: %s", cfg.EthRPCUrl)
	}
	if validator.IsEmpty(cfg.BaseRPCUrl) {
		return fmt.Errorf("invalid base rpc url: %s", cfg.BaseRPCUrl)
	}
	if validator.IsEmpty(cfg.AlchemyAPIKey) {
		return fmt.Errorf("invalid alchemy api key: %s", cfg.AlchemyAPIKey)
	}
	// if validator.IsEmpty(cfg.EtherscanAPIKey) {
	// 	return fmt.Errorf("invalid etherscan api key: %s", cfg.EtherscanAPIKey)
	// }
	if !validator.IsValidPort(cfg.OperatorRPCPort) {
		return fmt.Errorf("invalid operator rpc port: %s", cfg.OperatorRPCPort)
	}
	if !validator.IsValidIPAddress(cfg.PublicIPV4Address) {
		return fmt.Errorf("invalid public ipv4 address: %s", cfg.PublicIPV4Address)
	}
	if !validator.IsValidPrivateKey(cfg.PrivateKeyConsensus) {
		return fmt.Errorf("invalid private key consensus: %s", cfg.PrivateKeyConsensus)
	}
	if !validator.IsValidAddress(cfg.KeeperAddress) {
		return fmt.Errorf("invalid keeper address: %s", cfg.KeeperAddress)
	}
	if !validator.IsValidAddress(cfg.ConsensusAddress) {
		return fmt.Errorf("invalid consensus address: %s", cfg.ConsensusAddress)
	}
	if !validator.IsValidPeerID(cfg.PeerID) {
		return fmt.Errorf("invalid peer id: %s", cfg.PeerID)
	}
	if !validator.IsValidPort(cfg.KeeperRPCPort) {
		cfg.KeeperRPCPort = "9005"
	}
	if !validator.IsValidPort(cfg.KeeperP2PPort) {
		cfg.KeeperP2PPort = "9006"
	}
	if !validator.IsValidPort(cfg.KeeperMetricsPort) {
		cfg.KeeperMetricsPort = "9009"
	}
	if !validator.IsValidPort(cfg.GrafanaPort) {
		cfg.GrafanaPort = "3000"
	}
	if !validator.IsValidRPCAddress(cfg.AggregatorRPCAddress) {
		cfg.AggregatorRPCAddress = "http://127.0.0.1:9001"
	}
	if !validator.IsValidRPCAddress(cfg.HealthRPCAddress) {
		cfg.HealthRPCAddress = "http://127.0.0.1:9004"
	}
	if validator.IsEmpty(cfg.L1Chain) {
		cfg.L1Chain = "17000"
	}
	if validator.IsEmpty(cfg.L2Chain) {
		cfg.L2Chain = "84532"
	}
	if !validator.IsValidAddress(cfg.AVSGovernanceAddress) {
		cfg.AVSGovernanceAddress = "0x12f45551f11df20b3ecbdf329138bdc65cc58ec0"
	}
	if !validator.IsValidAddress(cfg.AttestationCenterAddress) {
		cfg.AttestationCenterAddress = "0x9725fb95b5ec36c062a49ca2712b3b1ff66f04ed"
	}
	if !validator.IsValidPeerID(cfg.OthenticBootstrapID) {
		cfg.OthenticBootstrapID = "12D3KooWBNFG1QjuF3UKAKvqhdXcxh9iBmj88cM5eU2EK5Pa91KB"
	}
	if validator.IsEmpty(cfg.IpfsHost) {
		cfg.IpfsHost = "aquamarine-urgent-limpet-846.mypinata.cloud"
	}
	// if validator.IsEmpty(cfg.PinataApiKey) {
	// 	cfg.PinataApiKey = "9f5922013fb9e2dfbc13"
	// }
	// if validator.IsEmpty(cfg.PinataSecretApiKey) {
	// 	cfg.PinataSecretApiKey = "190e9f1c959861bce0aed5a0e6c74a45a225658f7fa4fdc70f3fe136b76587fb"
	// }
	return nil
}

func GetEthRPCUrl() string {
	return cfg.EthRPCUrl
}

func GetBaseRPCUrl() string {
	return cfg.BaseRPCUrl
}

func GetAlchemyAPIKey() string {
	return cfg.AlchemyAPIKey
}

func GetEtherscanAPIKey() string {
	return cfg.EtherscanAPIKey
}

func GetPrivateKeyConsensus() string {
	return cfg.PrivateKeyConsensus
}

func GetPrivateKeyController() string {
	return cfg.PrivateKeyController
}

func GetKeeperAddress() string {
	return cfg.KeeperAddress
}

func GetConsensusAddress() string {
	return cfg.ConsensusAddress
}

func GetPublicIPV4Address() string {
	return cfg.PublicIPV4Address
}

func GetPeerID() string {
	return cfg.PeerID
}

func GetOperatorRPCPort() string {
	return cfg.OperatorRPCPort
}

func IsDevMode() bool {
	return cfg.DevMode
}

func GetKeeperRPCPort() string {
	return cfg.KeeperRPCPort
}

func GetAggregatorRPCAddress() string {
	return cfg.AggregatorRPCAddress
}

func GetHealthRPCAddress() string {
	return cfg.HealthRPCAddress
}

func GetAvsGovernanceAddress() string {
	return cfg.AVSGovernanceAddress
}

func GetAttestationCenterAddress() string {
	return cfg.AttestationCenterAddress
}

func GetIpfsHost() string {
	return cfg.IpfsHost
}

func GetPinataApiKey() string {
	return cfg.PinataApiKey
}

func GetVersion() string {
	return "0.1.2"
}
