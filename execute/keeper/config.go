package keeper

import (
	// "context"
	// "encoding/hex"
	// "encoding/json"
	"fmt"

	// "math/big"
	// "io/ioutil"

	"gopkg.in/yaml.v2"

	// "github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	// "github.com/ethereum/go-ethereum/ethclient"
	"github.com/prometheus/client_golang/prometheus"

	// "github.com/trigg3rX/triggerx-backend/pkg/core/chainio"

	// "github.com/Layr-Labs/eigensdk-go/chainio/clients"
	// "github.com/Layr-Labs/eigensdk-go/chainio/clients/wallet"
	// "github.com/Layr-Labs/eigensdk-go/chainio/txmgr"
	"github.com/Layr-Labs/eigensdk-go/crypto/bls"
	// sdkecdsa "github.com/Layr-Labs/eigensdk-go/crypto/ecdsa"

	// "github.com/trigg3rX/triggerx-backend/pkg/core"

	sdklogging "github.com/Layr-Labs/eigensdk-go/logging"
	// sdkmetrics "github.com/Layr-Labs/eigensdk-go/metrics"
	// "github.com/Layr-Labs/eigensdk-go/metrics/collectors/economic"

	// "github.com/Layr-Labs/eigensdk-go/nodeapi"
	// "github.com/Layr-Labs/eigensdk-go/signerv2"
	sdktypes "github.com/Layr-Labs/eigensdk-go/types"

	// rpccalls "github.com/Layr-Labs/eigensdk-go/metrics/collectors/rpc_calls"

	"github.com/trigg3rX/triggerx-backend/pkg/metrics"
	"github.com/trigg3rX/triggerx-backend/pkg/types"

	txtaskmanager "github.com/trigg3rX/triggerx-contracts/bindings/contracts/TriggerXTaskManager"
	sdkcommon "github.com/trigg3rX/triggerx-backend/pkg/common"

	// sdkelcontracts "github.com/Layr-Labs/eigensdk-go/chainio/clients/elcontracts"
	// "github.com/Layr-Labs/eigensdk-go/chainio/clients/eth"

	"os"
	"path/filepath"
)

type Keeper struct {
	Config                     types.NodeConfig
	Logger                     sdklogging.Logger
	EthClient                  sdkcommon.EthClientInterface
	MetricsReg                 *prometheus.Registry
	Metrics                    *metrics.AvsAndEigenMetrics
	BlsKeypair                 *bls.KeyPair
	KeeperId                   sdktypes.OperatorId
	KeeperAddr                 common.Address
	NewTaskCreatedChan         chan *txtaskmanager.ContractTriggerXTaskManagerTaskCreated
	ValidatorServerIpPortAddr  string
	TriggerxServiceManagerAddr common.Address
}

type OperatorStatus struct {
	EcdsaAddress      string
	PubkeysRegistered bool
	G1Pubkey          string
	G2Pubkey          string
	RegisteredWithAvs bool
	OperatorId        string
}

func expandHomeDir(path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("empty path provided")
	}

	if path[0] == '~' {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get home directory: %w", err)
		}
		return filepath.Join(home, path[1:]), nil
	}
	return path, nil
}

func NewKeeperFromConfigFile(configPath string) (*Keeper, error) {
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
	// Validate required fields
	if c.ServiceManagerAddress == "" {
		return nil, fmt.Errorf("ServiceManagerAddress is empty in configuration")
	}
	if c.OperatorStateRetrieverAddress == "" {
		return nil, fmt.Errorf("OperatorStateRetriever address is empty in configuration")
	}
	if c.MetricsIpPortAddress == "" {
		return nil, fmt.Errorf("Prometheus metrics ip port address is empty in configuration")
	}
	if c.EthRpcUrl == "" {
		return nil, fmt.Errorf("EthRpcUrl is empty in configuration")
	}
	if c.EthWsUrl == "" {
		return nil, fmt.Errorf("EthWsUrl is empty in configuration")
	}

// 	var logLevel sdklogging.LogLevel
// 	if c.Production {
// 		logLevel = sdklogging.Production
// 	} else {
// 		logLevel = sdklogging.Development
// 	}
// 	logger, err := sdklogging.NewZapLogger(logLevel)
// 	if err != nil {
// 		return nil, err
// 	}
// 	reg := prometheus.NewRegistry()
// 	eigenMetrics := sdkmetrics.NewEigenMetrics(c.AvsName, c.EigenMetricsIpPortAddress, reg, logger)
// 	avsAndEigenMetrics := metrics.NewAvsAndEigenMetrics(c.AvsName, eigenMetrics, reg)

// 	// Setup Node Api
// 	nodeApi := nodeapi.NewNodeApi(c.AvsName, c.SemVer, c.NodeApiIpPortAddress, logger)

// 	var ethRpcClient, ethWsClient sdkcommon.EthClientInterface
// 	if c.EnableMetrics {
// 		rpcCallsCollector := rpccalls.NewCollector(c.AvsName, reg)
// 		ethRpcClient, err = eth.NewInstrumentedClient(c.EthRpcUrl, rpcCallsCollector)
// 		if err != nil {
// 			logger.Errorf("Cannot create http ethclient", "err", err)
// 			return nil, err
// 		}
// 		ethWsClient, err = eth.NewInstrumentedClient(c.EthWsUrl, rpcCallsCollector)
// 		if err != nil {
// 			logger.Errorf("Cannot create ws ethclient", "err", err)
// 			return nil, err
// 		}
// 	} else {
// 		ethRpcClient, err = ethclient.Dial(c.EthRpcUrl)
// 		if err != nil {
// 			logger.Errorf("Cannot create http ethclient", "err", err)
// 			return nil, err
// 		}
// 		ethWsClient, err = ethclient.Dial(c.EthWsUrl)
// 		if err != nil {
// 			logger.Errorf("Cannot create ws ethclient", "err", err)
// 			return nil, err
// 		}
// 	}

// 	// blsKeyPassword, ok := os.LookupEnv("OPERATOR_BLS_KEY_PASSWORD")
// 	// if !ok {
// 	// 	logger.Warnf("OPERATOR_BLS_KEY_PASSWORD env var not set. using empty string")
// 	// }
// 	blsKeyPassword := "pL6!oK5@iJ4#"
// 	blsKeyPath, err := expandHomeDir(c.BlsPrivateKeyStorePath)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to process BLS keystore path: %w", err)
// 	}

// 	blsKeyPair, err := bls.ReadPrivateKeyFromFile(blsKeyPath, blsKeyPassword)
// 	if err != nil {
// 		logger.Errorf("Cannot parse bls private key", "err", err)
// 		return nil, err
// 	}
// 	// TODO(samlaf): should we add the chainId to the config instead?
// 	// this way we can prevent creating a signer that signs on mainnet by mistake
// 	// if the config says chainId=5, then we can only create a goerli signer
// 	chainId, err := ethRpcClient.ChainID(context.Background())
// 	if err != nil {
// 		logger.Error("Cannot get chainId", "err", err)
// 		return nil, err
// 	}

// 	// ecdsaKeyPassword, ok := os.LookupEnv("OPERATOR_ECDSA_KEY_PASSWORD")
// 	// if !ok {
// 	// 	logger.Warnf("OPERATOR_ECDSA_KEY_PASSWORD env var not set. using empty string")
// 	// }
// 	ecdsaKeyPassword := "pL6!oK5@iJ4#"
// 	ecdsaKeyPath, err := expandHomeDir(c.EcdsaPrivateKeyStorePath)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to process ECDSA keystore path: %w", err)
// 	}

// 	signerV2, _, err := signerv2.SignerFromConfig(signerv2.Config{
// 		KeystorePath: ecdsaKeyPath,
// 		Password:     ecdsaKeyPassword,
// 	}, chainId)
// 	if err != nil {
// 		panic(err)
// 	}
// 	chainioConfig := clients.BuildAllConfig{
// 		EthHttpUrl:                 c.EthRpcUrl,
// 		EthWsUrl:                   c.EthWsUrl,
// 		RegistryCoordinatorAddr:    c.ServiceManagerAddress,
// 		OperatorStateRetrieverAddr: c.OperatorStateRetrieverAddress,
// 		AvsName:                    c.AvsName,
// 		PromMetricsIpPortAddress:   c.EigenMetricsIpPortAddress,
// 	}
// 	operatorEcdsaPrivateKey, err := sdkecdsa.ReadKey(
// 		ecdsaKeyPath,
// 		ecdsaKeyPassword,
// 	)
// 	if err != nil {
// 		return nil, err
// 	}
// 	sdkClients, err := clients.BuildAll(chainioConfig, operatorEcdsaPrivateKey, logger)
// 	if err != nil {
// 		panic(err)
// 	}
// 	skWallet, err := wallet.NewPrivateKeyWallet(ethRpcClient, signerV2, common.HexToAddress(c.KeeperAddress), logger)
// 	if err != nil {
// 		panic(err)
// 	}
// 	txMgr := txmgr.NewSimpleTxManager(skWallet, ethRpcClient, logger, common.HexToAddress(c.KeeperAddress))
// 	avsReader, err := chainio.BuildAvsReader(
// 		common.HexToAddress(c.ServiceManagerAddress),
// 		common.HexToAddress(c.OperatorStateRetrieverAddress),
// 		ethRpcClient, logger)
// 	if err != nil {
// 		logger.Error("Cannot create AvsReader", "err", err)
// 		return nil, err
// 	}
// 	if avsReader == nil {
// 		return nil, fmt.Errorf("avsReader was not properly initialized")
// 	}
// 	avsWriter, err := chainio.BuildAvsWriter(txMgr, common.HexToAddress(c.ServiceManagerAddress),
// 		common.HexToAddress(c.OperatorStateRetrieverAddress), ethRpcClient, logger,
// 	)
// 	if err != nil {
// 		logger.Error("Cannot create AvsWriter", "err", err)
// 		return nil, err
// 	}
// 	avsSubscriber, err := chainio.BuildAvsSubscriber(common.HexToAddress(c.ServiceManagerAddress),
// 		common.HexToAddress(c.OperatorStateRetrieverAddress), ethWsClient, logger,
// 	)
// 	if err != nil {
// 		logger.Error("Cannot create AvsSubscriber", "err", err)
// 		return nil, err
// 	}

// 	// We must register the economic metrics separately because they are exported metrics (from jsonrpc or subgraph calls)
// 	// and not instrumented metrics: see https://prometheus.io/docs/instrumenting/writing_clientlibs/#overall-structure
// 	quorumNames := map[sdktypes.QuorumNum]string{
// 		0: "quorum0",
// 	}
// 	economicMetricsCollector := economic.NewCollector(
// 		sdkClients.ElChainReader, sdkClients.AvsRegistryChainReader,
// 		c.AvsName, logger, common.HexToAddress(c.KeeperAddress), quorumNames)
// 	reg.MustRegister(economicMetricsCollector)

// 	keeper := &Keeper{
// 		Config:                     c,
// 		Logger:                     logger,
// 		EthClient:                  ethRpcClient,
// 		MetricsReg:                 reg,
// 		Metrics:                    avsAndEigenMetrics,
// 		NodeApi:                    nodeApi,
// 		AvsReader:                  avsReader,
// 		AvsWriter:                  avsWriter,
// 		AvsSubscriber:              avsSubscriber,
// 		EigenlayerReader:           *sdkClients.ElChainReader,
// 		EigenlayerWriter:           *sdkClients.ElChainWriter,
// 		BlsKeypair:                 blsKeyPair,
// 		KeeperId:                   [32]byte{0}, // this will be set after registration
// 		KeeperAddr:                 common.HexToAddress(c.KeeperAddress),
// 		NewTaskCreatedChan:         make(chan *txtaskmanager.ContractTriggerXTaskManagerTaskCreated),
// 		ValidatorServerIpPortAddr:  "",
// 		TriggerxServiceManagerAddr: common.HexToAddress(c.ServiceManagerAddress),
// 	}

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

	// return keeper, nil
	return nil, nil
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
