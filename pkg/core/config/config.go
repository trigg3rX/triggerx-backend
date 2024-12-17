package config

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"os"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/urfave/cli"

	"github.com/Layr-Labs/eigensdk-go/chainio/clients/wallet"
	"github.com/Layr-Labs/eigensdk-go/chainio/txmgr"
	// "github.com/Layr-Labs/eigensdk-go/crypto/bls"
	sdklogging "github.com/Layr-Labs/eigensdk-go/logging"
	"github.com/Layr-Labs/eigensdk-go/signerv2"
	"github.com/ethereum/go-ethereum/ethclient"

	sdkutils "github.com/Layr-Labs/eigensdk-go/utils"
	"encoding/json"
	"gopkg.in/yaml.v3"
)

const AVS_NAME = "TriggerX"
const SEM_VER = "0.0.1"

type Config struct {
	EcdsaPrivateKey						*ecdsa.PrivateKey		// Keeper node's private key - used for signing transactions
	// BlsPrivateKey						*bls.PrivateKey			// BLS private key - used for signing BLS messages
	Logger								sdklogging.Logger
	// EigenMetricsIpPortAddress 			string
	// Needed for Communication:
	EthHttpRpcUrl						string					// Ethereum HTTP RPC URL
	EthWsRpcUrl							string					// Ethereum WebSocket RPC URL
	EthHttpClient						ethclient.Client		// Ethereum HTTP client
	EthWsClient							ethclient.Client		// Ethereum WebSocket client
	SignerFn							signerv2.SignerFn `json:"-"`
	TxMgr								txmgr.TxManager			// Transaction manager
	// Contracts	
	OperatorStateRetrieverAddr			common.Address			// Operator state retriever address
	TriggerXServiceManagerAddr			common.Address			// TriggerX service manager address
	// DB API address
	DBApiIpPortAddr						string					// DB API IP port address
}

type ConfigRaw struct {
	Environment                	sdklogging.LogLevel `yaml:"environment"`
	EthRpcUrl                  	string              `yaml:"eth_rpc_url"`
	EthWsUrl                   	string              `yaml:"eth_ws_url"`
	DBApiIpPortAddr				string				`yaml:"db_api_ip_port_addr"`	// DB API IP port address
}

type TriggerXDeploymentRaw struct {
	Addresses TriggerXContractsRaw `json:"addresses"`
}
type TriggerXContractsRaw struct {
	ServiceManagerAddr    string `json:"serviceManager"`
	OperatorStateRetrieverAddr string `json:"operatorStateRetriever"`
}

func NewConfig(ctx *cli.Context) (*Config, error) {
	var configRaw ConfigRaw
	configFilePath := ctx.GlobalString(ConfigFileFlag.Name)
	if configFilePath != "" {
		yamlFile, err := os.ReadFile(configFilePath)
		if err != nil {
			return nil, err
		}
		if err := yaml.Unmarshal(yamlFile, &configRaw); err != nil {
			return nil, err
		}
	}

	var triggerXDeploymentRaw TriggerXDeploymentRaw
	triggerXDeploymentFilePath := ctx.GlobalString(TriggerXDeploymentFileFlag.Name)
	if _, err := os.Stat(triggerXDeploymentFilePath); errors.Is(err, os.ErrNotExist) {
		panic("Path " + triggerXDeploymentFilePath + " does not exist")
	}
	jsonFile, err := os.ReadFile(triggerXDeploymentFilePath)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(jsonFile, &triggerXDeploymentRaw); err != nil {
		return nil, err
	}

	logger, err := sdklogging.NewZapLogger(configRaw.Environment)
	if err != nil {
		return nil, err
	}

	ethRpcClient, err := ethclient.Dial(configRaw.EthRpcUrl)
	if err != nil {
		logger.Errorf("Cannot create http ethclient", "err", err)
		return nil, err
	}

	ethWsClient, err := ethclient.Dial(configRaw.EthWsUrl)
	if err != nil {
		logger.Errorf("Cannot create ws ethclient", "err", err)
		return nil, err
	}

	ecdsaPrivateKeyString := ctx.GlobalString(EcdsaPrivateKeyFlag.Name)
	if ecdsaPrivateKeyString[:2] == "0x" {
		ecdsaPrivateKeyString = ecdsaPrivateKeyString[2:]
	}
	ecdsaPrivateKey, err := crypto.HexToECDSA(ecdsaPrivateKeyString)
	if err != nil {
		logger.Errorf("Cannot parse ecdsa private key", "err", err)
		return nil, err
	}

	keeperNodeAddr, err := sdkutils.EcdsaPrivateKeyToAddress(ecdsaPrivateKey)
	if err != nil {
		logger.Error("Cannot get operator address", "err", err)
		return nil, err
	}


	chainId, err := ethRpcClient.ChainID(context.Background())
	if err != nil {
		logger.Error("Cannot get chainId", "err", err)
		return nil, err
	}

	signerV2, _, err := signerv2.SignerFromConfig(signerv2.Config{PrivateKey: ecdsaPrivateKey}, chainId)
	if err != nil {
		panic(err)
	}
	skWallet, err := wallet.NewPrivateKeyWallet(ethRpcClient, signerV2, keeperNodeAddr, logger)
	if err != nil {
		panic(err)
	}
	txMgr := txmgr.NewSimpleTxManager(skWallet, ethRpcClient, logger, keeperNodeAddr)

	config := &Config{
		EcdsaPrivateKey:            ecdsaPrivateKey,
		Logger:                     logger,
		EthWsRpcUrl:                configRaw.EthWsUrl,
		EthHttpRpcUrl:              configRaw.EthRpcUrl,
		EthHttpClient:              *ethRpcClient,
		EthWsClient:                *ethWsClient,
		OperatorStateRetrieverAddr: common.HexToAddress(triggerXDeploymentRaw.Addresses.OperatorStateRetrieverAddr),
		TriggerXServiceManagerAddr: common.HexToAddress(triggerXDeploymentRaw.Addresses.ServiceManagerAddr),
		SignerFn:                 	signerV2,
		TxMgr:                    	txMgr,
	}
	config.validate()
	return config, nil
}

func (c *Config) validate() {
	if c.OperatorStateRetrieverAddr == common.HexToAddress("") {
		panic("Config: BLSOperatorStateRetrieverAddr is required")
	}
	if c.TriggerXServiceManagerAddr == common.HexToAddress("") {
		panic("Config: TriggerXServiceManagerAddr is required")
	}
}

var (
	ConfigFileFlag = cli.StringFlag{
		Name:     "config",
		Required: true,
		Usage:    "Load configuration from `FILE`",
	}
	TriggerXDeploymentFileFlag = cli.StringFlag{
		Name:     "triggerx-deployment",
		Required: true,
		Usage:    "Load triggerx contract addresses from `FILE`",
	}
	EcdsaPrivateKeyFlag = cli.StringFlag{
		Name:     "ecdsa-private-key",
		Usage:    "Ethereum private key",
		Required: true,
		EnvVar:   "ECDSA_PRIVATE_KEY",
	}
)

var requiredFlags = []cli.Flag{
	ConfigFileFlag,
	TriggerXDeploymentFileFlag,
	EcdsaPrivateKeyFlag,
}

var optionalFlags = []cli.Flag{}

func init() {
	Flags = append(requiredFlags, optionalFlags...)
}

var Flags []cli.Flag