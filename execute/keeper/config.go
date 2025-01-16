package keeper

import (
	"context"
	// "encoding/hex"
	// "encoding/json"
	"fmt"

	// "math/big"
	// "io/ioutil"

	"gopkg.in/yaml.v2"

	// "github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/Layr-Labs/eigensdk-go/chainio/clients"
	"github.com/Layr-Labs/eigensdk-go/chainio/clients/eth"
	"github.com/Layr-Labs/eigensdk-go/chainio/clients/wallet"
	"github.com/Layr-Labs/eigensdk-go/chainio/txmgr"
	"github.com/Layr-Labs/eigensdk-go/crypto/bls"

	sdkecdsa "github.com/Layr-Labs/eigensdk-go/crypto/ecdsa"

	chainio "github.com/Layr-Labs/eigensdk-go/chainio/clients/avsregistry"

	sdklogging "github.com/Layr-Labs/eigensdk-go/logging"
	sdkmetrics "github.com/Layr-Labs/eigensdk-go/metrics"

	"github.com/Layr-Labs/eigensdk-go/metrics/collectors/economic"

	// "github.com/Layr-Labs/eigensdk-go/nodeapi"
	"github.com/Layr-Labs/eigensdk-go/signerv2"
	sdktypes "github.com/Layr-Labs/eigensdk-go/types"

	rpccalls "github.com/Layr-Labs/eigensdk-go/metrics/collectors/rpc_calls"

	"github.com/trigg3rX/triggerx-backend/pkg/metrics"
	"github.com/trigg3rX/triggerx-backend/pkg/types"

	sdkcommon "github.com/trigg3rX/triggerx-backend/pkg/common"
	// txtaskmanager "github.com/trigg3rX/triggerx-contracts/bindings/contracts/TriggerXTaskManager"

	// sdkelcontracts "github.com/Layr-Labs/eigensdk-go/chainio/clients/elcontracts"
	// sdketh "github.com/Layr-Labs/eigensdk-go/chainio/clients/eth"

	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

type Keeper struct {
	EcdsaAddress      common.Address
	PubkeysRegistered bool
	BlsKeypair        *bls.KeyPair
	RegisteredWithAvs bool
	KeeperId          string

	// Config                     types.NodeConfig
	Logger        sdklogging.Logger
	EthClient     sdkcommon.EthClientInterface
	EthWsClient   sdkcommon.EthClientInterface
	MetricsReg    *prometheus.Registry
	Metrics       *metrics.AvsAndEigenMetrics
	AvsReader     *chainio.ChainReader
	AvsWriter     *chainio.ChainWriter
	AvsSubscriber *chainio.ChainSubscriber

	ValidatorServerIpPortAddr  string
	TriggerxServiceManagerAddr common.Address
}

func handleHomeDirPath(path string) string {
	if len(path) >= 2 && path[:2] == "~/" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(homeDir, path[2:])
	}
	return path
}

func NewKeeperFromConfigFile(configPath string) (*Keeper, error) {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		return nil, fmt.Errorf("error loading .env file: %w", err)
	}

	// Read and parse the config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	var config types.NodeConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("error parsing config file: %w", err)
	}

	return NewKeeperFromConfig(config)
}

func NewKeeperFromConfig(c types.NodeConfig) (*Keeper, error) {
	var logLevel sdklogging.LogLevel
	if c.Production {
		logLevel = sdklogging.Production
	} else {
		logLevel = sdklogging.Development
	}
	logger, err := sdklogging.NewZapLogger(logLevel)
	if err != nil {
		return nil, err
	}

	err = ValidateConfig(c)
	if err != nil {
		logger.Errorf("Invalid config: %v", err)
		return nil, err
	}

	reg := prometheus.NewRegistry()
	eigenMetrics := sdkmetrics.NewEigenMetrics(c.AvsName, c.MetricsIpPortAddress, reg, logger)
	avsAndEigenMetrics := metrics.NewAvsAndEigenMetrics(c.AvsName, eigenMetrics, reg)

	var ethRpcClient, ethWsClient sdkcommon.EthClientInterface
	if c.EnableMetrics {
		rpcCallsCollector := rpccalls.NewCollector(c.AvsName, reg)
		ethRpcClient, err = eth.NewInstrumentedClient(c.EthRpcUrl, rpcCallsCollector)
		if err != nil {
			logger.Errorf("Cannot create http ethclient", "err", err)
			return nil, err
		}
		ethWsClient, err = eth.NewInstrumentedClient(c.EthWsUrl, rpcCallsCollector)
		if err != nil {
			logger.Errorf("Cannot create ws ethclient", "err", err)
			return nil, err
		}
	} else {
		ethRpcClient, err = ethclient.Dial(c.EthRpcUrl)
		if err != nil {
			logger.Errorf("Cannot create http ethclient", "err", err)
			return nil, err
		}
		ethWsClient, err = ethclient.Dial(c.EthWsUrl)
		if err != nil {
			logger.Errorf("Cannot create ws ethclient", "err", err)
			return nil, err
		}
	}

	blsKeyPassword, ok := os.LookupEnv("OPERATOR_BLS_KEY_PASSWORD")
	if !ok {
		logger.Warnf("OPERATOR_BLS_KEY_PASSWORD env var not set. using empty string")
	}
	// blsKeyPassword := "pL6!oK5@iJ4#"
	blsKeyPath := handleHomeDirPath(c.BlsPrivateKeyStorePath)
	if blsKeyPath == "" {
		return nil, fmt.Errorf("invalid bls keystore path")
	}

	blsKeyPair, err := bls.ReadPrivateKeyFromFile(blsKeyPath, blsKeyPassword)
	if err != nil {
		logger.Errorf("Cannot parse bls private key", "err", err)
		return nil, err
	}
	chainId, err := ethRpcClient.ChainID(context.Background())
	if err != nil {
		logger.Error("Cannot get chainId", "err", err)
		return nil, err
	}

	ecdsaKeyPassword, ok := os.LookupEnv("OPERATOR_ECDSA_KEY_PASSWORD")
	if !ok {
		logger.Warnf("OPERATOR_ECDSA_KEY_PASSWORD env var not set. using empty string")
	}
	// ecdsaKeyPassword := "pL6!oK5@iJ4#"
	ecdsaKeyPath := handleHomeDirPath(c.EcdsaPrivateKeyStorePath)
	if ecdsaKeyPath == "" {
		return nil, fmt.Errorf("invalid ecdsa keystore path")
	}

	signerV2, _, err := signerv2.SignerFromConfig(signerv2.Config{
		KeystorePath: ecdsaKeyPath,
		Password:     ecdsaKeyPassword,
	}, chainId)
	if err != nil {
		panic(err)
	}

	// logger.Info("signerV2", "signerV2", signerV2)
	chainioConfig := clients.BuildAllConfig{
		EthHttpUrl:                 c.EthRpcUrl,
		EthWsUrl:                   c.EthWsUrl,
		RegistryCoordinatorAddr:    c.RegistryCoordinatorAddress,
		OperatorStateRetrieverAddr: c.OperatorStateRetrieverAddress,
		AvsName:                    c.AvsName,
		PromMetricsIpPortAddress:   c.MetricsIpPortAddress,
	}
	operatorEcdsaPrivateKey, err := sdkecdsa.ReadKey(
		ecdsaKeyPath,
		ecdsaKeyPassword,
	)
	if err != nil {
		return nil, err
	}
	sdkClients, err := clients.BuildAll(chainioConfig, operatorEcdsaPrivateKey, logger)
	if err != nil {
		panic(err)
	}
	skWallet, err := wallet.NewPrivateKeyWallet(ethRpcClient, signerV2, common.HexToAddress(c.KeeperAddress), logger)
	if err != nil {
		panic(err)
	}
	txMgr := txmgr.NewSimpleTxManager(skWallet, ethRpcClient, logger, common.HexToAddress(c.KeeperAddress))
	avsReader, err := chainio.NewReaderFromConfig(
		chainio.Config{
			RegistryCoordinatorAddress:    common.HexToAddress(c.RegistryCoordinatorAddress),
			OperatorStateRetrieverAddress: common.HexToAddress(c.OperatorStateRetrieverAddress),
		},
		ethRpcClient,
		logger,
	)
	if err != nil {
		logger.Error("Cannot create AvsReader", "err", err)
		return nil, err
	}

	avsWriter, err := chainio.NewWriterFromConfig(
		chainio.Config{
			RegistryCoordinatorAddress:    common.HexToAddress(c.RegistryCoordinatorAddress),
			OperatorStateRetrieverAddress: common.HexToAddress(c.OperatorStateRetrieverAddress),
		},
		ethRpcClient,
		txMgr,
		logger,
	)
	if err != nil {
		logger.Error("Cannot create AvsWriter", "err", err)
		return nil, err
	}

	avsSubscriber, err := chainio.NewSubscriberFromConfig(
		chainio.Config{
			RegistryCoordinatorAddress:    common.HexToAddress(c.RegistryCoordinatorAddress),
			OperatorStateRetrieverAddress: common.HexToAddress(c.OperatorStateRetrieverAddress),
		},
		ethWsClient,
		logger,
	)
	if err != nil {
		logger.Error("Cannot create AvsSubscriber", "err", err)
		return nil, err
	}

	// We must register the economic metrics separately because they are exported metrics (from jsonrpc or subgraph calls)
	// and not instrumented metrics: see https://prometheus.io/docs/instrumenting/writing_clientlibs/#overall-structure
	quorumNames := map[sdktypes.QuorumNum]string{
		0: "quorum0",
	}
	economicMetricsCollector := economic.NewCollector(
		sdkClients.ElChainReader, sdkClients.AvsRegistryChainReader,
		c.AvsName, logger, common.HexToAddress(c.KeeperAddress), quorumNames)
	reg.MustRegister(economicMetricsCollector)

	keeper := &Keeper{
		EcdsaAddress:      common.HexToAddress(c.KeeperAddress),
		PubkeysRegistered: false,
		BlsKeypair:        blsKeyPair,
		RegisteredWithAvs: false,
		KeeperId:          "",
		// Config: c,
		Logger:                     logger,
		EthClient:                  ethRpcClient,
		EthWsClient:                ethWsClient,
		MetricsReg:                 reg,
		Metrics:                    avsAndEigenMetrics,
		AvsReader:                  avsReader,
		AvsWriter:                  avsWriter,
		AvsSubscriber:              avsSubscriber,
		ValidatorServerIpPortAddr:  "",
		TriggerxServiceManagerAddr: common.HexToAddress(c.ServiceManagerAddress),
	}

	// 	// OperatorId is set in contract during registration so we get it after registering operator.
	// 	operatorId, err := sdkClients.AvsRegistryChainReader.GetOperatorId(&bind.CallOpts{}, keeper.KeeperAddr)
	// 	if err != nil {
	// 		logger.Error("Cannot get operator id", "err", err)
	// 		return nil, err
	// 	}
	// 	keeper.KeeperId = operatorId
	// 	logger.Info("Operator info",
	// 		"operatorId", operatorId,
	// 		"operatorAddr", c.KeeperAddress,
	// 		"operatorG1Pubkey", keeper.BlsKeypair.GetPubKeyG1(),
	// 		"operatorG2Pubkey", keeper.BlsKeypair.GetPubKeyG2(),
	// 	)

	return keeper, nil
	// return nil, nil
}

// func (k *Keeper) Start(ctx context.Context) error {
// 	// Check if operator is registered using AvsReader interface
// 	operatorIsRegistered, err := k.AvsReader.GetOperatorId(&bind.CallOpts{}, k.KeeperAddr)
// 	if err != nil {
// 		k.Logger.Error("Error checking if operator is registered", "err", err)
// 		return err
// 	}
// 	if operatorIsRegistered == [32]byte{} {
// 		return fmt.Errorf("operator is not registered. Register operator using the operator-cli before starting operator")
// 	}

// 	k.Logger.Info("Starting keeper node")

// 	// Start Node API if enabled
// 	if k.Config.EnableNodeApi {
// 		k.NodeApi.Start()
// 	}

// 	// Start metrics if enabled
// 	var metricsErrChan <-chan error
// 	if k.Config.EnableMetrics {
// 		metricsErrChan = k.Metrics.Start(ctx, k.MetricsReg)
// 	} else {
// 		metricsErrChan = make(chan error)
// 	}

// 	// Subscribe to new tasks using AvsSubscriber interface
// 	sub := k.AvsSubscriber.SubscribeToNewTasks(k.NewTaskCreatedChan)
// 	defer sub.Unsubscribe()

// 	// Main event loop
// 	for {
// 		select {
// 		case <-ctx.Done():
// 			k.Logger.Info("Context cancelled, shutting down")
// 			return nil
// 		case err := <-metricsErrChan:
// 			k.Logger.Error("Metrics error", "err", err)
// 			return err
// 		case err := <-sub.Err():
// 			k.Logger.Error("Subscription error", "err", err)
// 			// Resubscribe on error
// 			sub.Unsubscribe()
// 			sub = k.AvsSubscriber.SubscribeToNewTasks(k.NewTaskCreatedChan)
// 		case task := <-k.NewTaskCreatedChan:
// 			k.handleNewTask(task)
// 		}
// 	}
// }

// func (k *Keeper) handleNewTask(task *txtaskmanager.ContractTriggerXTaskManagerTaskCreated) {
// 	k.Metrics.IncNumTasksReceived()

// 	// bleh bleh bleh
// }

// func (k *Keeper) SignTaskResponse(taskResponse *txtaskmanager.ContractTriggerXTaskManagerTaskResponse) (*aggregator.SignedTaskResponse, error) {
// 	taskResponseHash, err := core.GetTaskResponseDigest(taskResponse)
// 	if err != nil {
// 		k.Logger.Error("Error getting task response header hash. skipping task (this is not expected and should be investigated)", "err", err)
// 		return nil, err
// 	}
// 	blsSignature := k.BlsKeypair.SignMessage(taskResponseHash)
// 	signedTaskResponse := &aggregator.SignedTaskResponse{
// 		TaskResponse: *taskResponse,
// 		BlsSignature: *blsSignature,
// 		OperatorId:   k.KeeperId,
// 	}
// 	k.Logger.Debug("Signed task response", "signedTaskResponse", signedTaskResponse)
// 	return signedTaskResponse, nil
// }

func ValidateConfig(c types.NodeConfig) error {
	// Check if required fields are present and valid
	if c.AvsName == "" {
		return fmt.Errorf("avs_name is required")
	}

	if c.SemVer == "" {
		return fmt.Errorf("sem_ver is required")
	}

	// Validate Ethereum addresses
	if !common.IsHexAddress(c.KeeperAddress) {
		return fmt.Errorf("invalid keeper address: %s", c.KeeperAddress)
	}

	if !common.IsHexAddress(c.AvsDirectoryAddress) {
		return fmt.Errorf("invalid avs directory address: %s", c.AvsDirectoryAddress)
	}

	if !common.IsHexAddress(c.StrategyManagerAddress) {
		return fmt.Errorf("invalid strategy manager address: %s", c.StrategyManagerAddress)
	}

	if !common.IsHexAddress(c.ServiceManagerAddress) {
		return fmt.Errorf("invalid service manager address: %s", c.ServiceManagerAddress)
	}

	if !common.IsHexAddress(c.OperatorStateRetrieverAddress) {
		return fmt.Errorf("invalid operator state retriever address: %s", c.OperatorStateRetrieverAddress)
	}

	// Validate and expand keystore paths
	ecdsaPath := handleHomeDirPath(c.EcdsaPrivateKeyStorePath)
	if ecdsaPath == "" {
		return fmt.Errorf("invalid ecdsa keystore path")
	}
	if _, err := os.Stat(ecdsaPath); os.IsNotExist(err) {
		return fmt.Errorf("ecdsa keystore file does not exist at path: %s", ecdsaPath)
	}

	blsPath := handleHomeDirPath(c.BlsPrivateKeyStorePath)
	if blsPath == "" {
		return fmt.Errorf("invalid bls keystore path")
	}
	if _, err := os.Stat(blsPath); os.IsNotExist(err) {
		return fmt.Errorf("bls keystore file does not exist at path: %s", blsPath)
	}

	// Basic URL format validation
	if c.EthRpcUrl == "" {
		return fmt.Errorf("eth rpc url is required")
	}

	if c.EthWsUrl == "" {
		return fmt.Errorf("eth websocket url is required")
	}

	// Port address validation
	if c.MetricsIpPortAddress == "" {
		return fmt.Errorf("port address is required")
	}

	return nil
}

// func (k *Keeper) PrintOperatorStatus() error {
// 	fmt.Println("Printing operator status")
// 	operatorId, err := k.AvsReader.GetOperatorId(&bind.CallOpts{}, k.KeeperAddr)
// 	if err != nil {
// 		return err
// 	}
// 	pubkeysRegistered := operatorId != [32]byte{}
// 	registeredWithAvs := k.KeeperId != [32]byte{}
// 	operatorStatus := OperatorStatus{
// 		EcdsaAddress:      k.KeeperAddr.String(),
// 		PubkeysRegistered: pubkeysRegistered,
// 		G1Pubkey:          k.BlsKeypair.GetPubKeyG1().String(),
// 		G2Pubkey:          k.BlsKeypair.GetPubKeyG2().String(),
// 		RegisteredWithAvs: registeredWithAvs,
// 		OperatorId:        hex.EncodeToString(k.KeeperId[:]),
// 	}
// 	operatorStatusJson, err := json.MarshalIndent(operatorStatus, "", " ")
// 	if err != nil {
// 		return err
// 	}
// 	fmt.Println(string(operatorStatusJson))
// 	return nil
// }
