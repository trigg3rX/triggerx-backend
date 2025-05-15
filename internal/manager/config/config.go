package config

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"os"

	"github.com/trigg3rX/triggerx-backend/pkg/validator"
)

type Config struct {
	FoundNextPerformer bool

	EtherscanApiKey string
	AlchemyApiKey   string
	IpfsHost        string

	DeployerPrivateKey string

	ManagerRPCPort       string
	DatabaseRPCAddress   string
	AggregatorRPCAddress string

	DevMode bool
}

var cfg Config

func Init() error {
	if err := godotenv.Load(); err != nil {
		return fmt.Errorf("error loading .env file: %w", err)
	}

	cfg = Config{
		DevMode:              os.Getenv("DEV_MODE") == "true",
		FoundNextPerformer:   false,
		EtherscanApiKey:      os.Getenv("ETHERSCAN_API_KEY"),
		AlchemyApiKey:        os.Getenv("ALCHEMY_API_KEY"),
		DeployerPrivateKey:   os.Getenv("MANAGER_PRIVATE_KEY"),
		DatabaseRPCAddress:   os.Getenv("DATABASE_RPC_ADDRESS"),
		ManagerRPCPort:       os.Getenv("MANAGER_RPC_PORT"),
		AggregatorRPCAddress: os.Getenv("OTHENTIC_CLIENT_RPC_ADDRESS"),
		IpfsHost:             os.Getenv("IPFS_HOST"),
	}

	if err := validateConfig(cfg); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	if !cfg.DevMode {
		gin.SetMode(gin.ReleaseMode)
	}

	return nil
}

func validateConfig(cfg Config) error {
	if validator.IsEmpty(cfg.EtherscanApiKey) {
		return fmt.Errorf("invalid Etherscan API Key")
	}

	if validator.IsEmpty(cfg.AlchemyApiKey) {
		return fmt.Errorf("invalid Alchemy API Key")
	}

	if !validator.IsValidRPCAddress(cfg.DatabaseRPCAddress) {
		return fmt.Errorf("invalid Database RPC Address")
	}

	if !validator.IsValidPort(cfg.ManagerRPCPort) {
		return fmt.Errorf("invalid Manager RPC Port")
	}

	if !validator.IsValidRPCAddress(cfg.AggregatorRPCAddress) {
		return fmt.Errorf("invalid Aggregator RPC Address")
	}

	if validator.IsEmpty(cfg.IpfsHost) {
		return fmt.Errorf("invalid IPFS Host")
	}

	if validator.IsEmpty(cfg.DeployerPrivateKey) {
		return fmt.Errorf("invalid Deployer Private Key")
	}

	return nil
}

func GetEtherscanApiKey() string {
	return cfg.EtherscanApiKey
}

func GetAlchemyApiKey() string {
	return cfg.AlchemyApiKey
}

func GetIpfsHost() string {
	return cfg.IpfsHost
}

func GetManagerRPCPort() string {
	return cfg.ManagerRPCPort
}

func GetDatabaseRPCAddress() string {
	return cfg.DatabaseRPCAddress
}

func GetAggregatorRPCAddress() string {
	return cfg.AggregatorRPCAddress
}

func GetDeployerPrivateKey() string {
	return cfg.DeployerPrivateKey
}

func IsDevMode() bool {
	return cfg.DevMode
}

func GetFoundNextPerformer() bool {
	return cfg.FoundNextPerformer
}

func SetFoundNextPerformer(foundNextPerformer bool) {
	cfg.FoundNextPerformer = foundNextPerformer
}
