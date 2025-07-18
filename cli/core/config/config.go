package config

import (
	"crypto/ecdsa"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/joho/godotenv"
	"github.com/trigg3rX/triggerx-backend-imua/pkg/env"

	blscommon "github.com/prysmaticlabs/prysm/v5/crypto/bls/common"
)

type Config struct {
	// Private Keys
	ecdsaPrivateKey          *ecdsa.PrivateKey
	blsPrivateKey            *blscommon.SecretKey
	ecdsaPrivateKeyStorePath string
	blsPrivateKeyStorePath   string

	// RPC URLs
	ethHttpRpcUrl string
	ethWsRpcUrl   string

	// Contract Addresses
	operatorStateRetrieverAddr common.Address
	avsRegistryCoordinatorAddr common.Address

	// Aggregator
	aggregatorServerIpPortAddr string
	aggregatorAddress          common.Address

	// Operator and AVS Configuration
	operatorAddress common.Address
	avsOwnerAddress common.Address
	avsAddress      common.Address

	// API Configuration
	nodeApiIpPortAddress string
	enableNodeApi        bool

	// Other Configuration
	production bool
}

var cfg Config

func Init() error {
	if err := godotenv.Load(); err != nil {
		return fmt.Errorf("error loading .env file: %w", err)
	}

	// ECDSA Private Key
	ecdsaPrivateKeyStr := env.GetEnvString("OPERATOR_PRIVATE_KEY", "")
	ecdsaPrivateKey, err := crypto.HexToECDSA(ecdsaPrivateKeyStr)
	if err != nil {
		return fmt.Errorf("error converting private key to ECDSA: %w", err)
	}

	// BLS Private Key - check if environment variable exists but don't load it here
	// The operator will handle BLS key loading using the appropriate library
	blsPrivateKeyStr := env.GetEnvString("BLS_PRIVATE_KEY", "")
	var blsPrivateKey *blscommon.SecretKey
	// Note: BLS key loading is handled in the operator using the imua-avs-sdk library
	// We just track whether the environment variable is available
	if blsPrivateKeyStr != "" {
		// BLS private key is available in environment, operator will load it
		blsPrivateKey = nil
	}

	// ECDSA Private Key Store Path
	ecdsaPrivateKeyStorePath := env.GetEnvString("ECDSA_PRIVATE_KEY_STORE_PATH", "")

	// BLS Private Key Store Path
	blsPrivateKeyStorePath := env.GetEnvString("BLS_PRIVATE_KEY_STORE_PATH", "")

	// RPC URLs
	ethHttpRpcUrl := env.GetEnvString("ETH_HTTP_RPC_URL", "")
	ethWsRpcUrl := env.GetEnvString("ETH_WS_RPC_URL", "")

	// Contract Addresses
	operatorStateRetrieverAddrStr := env.GetEnvString("OPERATOR_STATE_RETRIEVER_ADDR", "")
	var operatorStateRetrieverAddr common.Address
	if operatorStateRetrieverAddrStr != "" {
		operatorStateRetrieverAddr = common.HexToAddress(operatorStateRetrieverAddrStr)
	}

	avsRegistryCoordinatorAddrStr := env.GetEnvString("AVS_REGISTRY_COORDINATOR_ADDR", "")
	var avsRegistryCoordinatorAddr common.Address
	if avsRegistryCoordinatorAddrStr != "" {
		avsRegistryCoordinatorAddr = common.HexToAddress(avsRegistryCoordinatorAddrStr)
	}

	// Aggregator
	aggregatorServerIpPortAddr := env.GetEnvString("AGGREGATOR_SERVER_IP_PORT_ADDR", "")

	aggregatorAddressStr := env.GetEnvString("AGGREGATOR_ADDRESS", "")
	var aggregatorAddress common.Address
	if aggregatorAddressStr != "" {
		aggregatorAddress = common.HexToAddress(aggregatorAddressStr)
	}

	// Operator and AVS Configuration
	operatorAddressStr := env.GetEnvString("OPERATOR_ADDRESS", "")
	var operatorAddress common.Address
	if operatorAddressStr != "" {
		operatorAddress = common.HexToAddress(operatorAddressStr)
	} else {
		// Derive from private key if not provided
		operatorAddress = crypto.PubkeyToAddress(ecdsaPrivateKey.PublicKey)
	}

	avsOwnerAddressStr := env.GetEnvString("AVS_OWNER_ADDRESS", "")
	var avsOwnerAddress common.Address
	if avsOwnerAddressStr != "" {
		avsOwnerAddress = common.HexToAddress(avsOwnerAddressStr)
	}

	avsAddressStr := env.GetEnvString("AVS_ADDRESS", "")
	var avsAddress common.Address
	if avsAddressStr != "" {
		avsAddress = common.HexToAddress(avsAddressStr)
	}

	// API Configuration
	nodeApiIpPortAddress := env.GetEnvString("NODE_API_IP_PORT_ADDRESS", "")
	enableNodeApi := env.GetEnvString("ENABLE_NODE_API", "false") == "true"

	// Other Configuration
	production := env.GetEnvString("PRODUCTION", "false") == "true"

	cfg = Config{
		ecdsaPrivateKey:            ecdsaPrivateKey,
		blsPrivateKey:              blsPrivateKey,
		ecdsaPrivateKeyStorePath:   ecdsaPrivateKeyStorePath,
		blsPrivateKeyStorePath:     blsPrivateKeyStorePath,
		ethHttpRpcUrl:              ethHttpRpcUrl,
		ethWsRpcUrl:                ethWsRpcUrl,
		operatorStateRetrieverAddr: operatorStateRetrieverAddr,
		avsRegistryCoordinatorAddr: avsRegistryCoordinatorAddr,
		aggregatorServerIpPortAddr: aggregatorServerIpPortAddr,
		aggregatorAddress:          aggregatorAddress,
		operatorAddress:            operatorAddress,
		avsOwnerAddress:            avsOwnerAddress,
		avsAddress:                 avsAddress,
		nodeApiIpPortAddress:       nodeApiIpPortAddress,
		enableNodeApi:              enableNodeApi,
		production:                 production,
	}
	return nil
}

func GetEcdsaPrivateKey() *ecdsa.PrivateKey {
	return cfg.ecdsaPrivateKey
}

func GetBlsPrivateKey() *blscommon.SecretKey {
	return cfg.blsPrivateKey
}

func GetEcdsaPrivateKeyStorePath() string {
	return cfg.ecdsaPrivateKeyStorePath
}

func GetBlsPrivateKeyStorePath() string {
	return cfg.blsPrivateKeyStorePath
}

func GetEthHttpRpcUrl() string {
	return cfg.ethHttpRpcUrl
}

func GetEthWsRpcUrl() string {
	return cfg.ethWsRpcUrl
}

func GetOperatorStateRetrieverAddr() common.Address {
	return cfg.operatorStateRetrieverAddr
}

func GetAvsRegistryCoordinatorAddr() common.Address {
	return cfg.avsRegistryCoordinatorAddr
}

func GetAggregatorServerIpPortAddr() string {
	return cfg.aggregatorServerIpPortAddr
}

func GetAggregatorAddress() common.Address {
	return cfg.aggregatorAddress
}

func GetOperatorAddress() common.Address {
	return cfg.operatorAddress
}

func GetAvsOwnerAddress() common.Address {
	return cfg.avsOwnerAddress
}

func GetAvsAddress() common.Address {
	return cfg.avsAddress
}

func GetNodeApiIpPortAddress() string {
	return cfg.nodeApiIpPortAddress
}

func GetEnableNodeApi() bool {
	return cfg.enableNodeApi
}

func GetProduction() bool {
	return cfg.production
}

func HasBLSPrivateKeyInEnv() bool {
	return env.GetEnvString("BLS_PRIVATE_KEY", "") != ""
}

func GetBLSPrivateKeyHex() string {
	return env.GetEnvString("BLS_PRIVATE_KEY", "")
}
