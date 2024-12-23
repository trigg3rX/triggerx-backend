package keeper

import (
	"context"
	"fmt"

	// "fmt"
	// "math/big"

	// "math/big"
	// "os"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/trigg3rX/triggerx-backend/pkg/core/chainio"

	"github.com/Layr-Labs/eigensdk-go/chainio/clients"
	"github.com/Layr-Labs/eigensdk-go/chainio/clients/wallet"
	"github.com/Layr-Labs/eigensdk-go/chainio/txmgr"
	"github.com/Layr-Labs/eigensdk-go/crypto/bls"
	sdkecdsa "github.com/Layr-Labs/eigensdk-go/crypto/ecdsa"

	// "github.com/Layr-Labs/eigensdk-go/logging"
	sdklogging "github.com/Layr-Labs/eigensdk-go/logging"
	sdkmetrics "github.com/Layr-Labs/eigensdk-go/metrics"
	"github.com/Layr-Labs/eigensdk-go/metrics/collectors/economic"

	"github.com/Layr-Labs/eigensdk-go/nodeapi"
	"github.com/Layr-Labs/eigensdk-go/signerv2"
	sdktypes "github.com/Layr-Labs/eigensdk-go/types"

	rpccalls "github.com/Layr-Labs/eigensdk-go/metrics/collectors/rpc_calls"

	"github.com/trigg3rX/triggerx-backend/pkg/metrics"
	"github.com/trigg3rX/triggerx-backend/pkg/types"

	txtaskmanager "github.com/trigg3rX/triggerx-backend/pkg/avsinterface/bindings/TriggerXTaskManager"
	sdkcommon "github.com/trigg3rX/triggerx-backend/pkg/common"

	sdkelcontracts "github.com/Layr-Labs/eigensdk-go/chainio/clients/elcontracts"
	"github.com/Layr-Labs/eigensdk-go/chainio/clients/eth"
)

type Keeper struct {
	Config                     types.NodeConfig
	Logger                     sdklogging.Logger
	EthClient                  sdkcommon.EthClientInterface
	MetricsReg                 *prometheus.Registry
	Metrics                    *metrics.AvsAndEigenMetrics
	NodeApi                    *nodeapi.NodeApi
	AvsReader                  chainio.AvsReaderer
	AvsWriter                  chainio.AvsWriterer
	AvsSubscriber              chainio.AvsSubscriberer
	EigenlayerReader           sdkelcontracts.ChainReader
	EigenlayerWriter           sdkelcontracts.ChainWriter
	BlsKeypair                 *bls.KeyPair
	KeeperId                   sdktypes.OperatorId
	KeeperAddr                 common.Address
	NewTaskCreatedChan         chan *txtaskmanager.ContractTriggerXTaskManagerTaskCreated
	ValidatorServerIpPortAddr  string
	TriggerxServiceManagerAddr common.Address
}

func NewKeeperFromConfig(c types.NodeConfig) (*Keeper, error) {
	// Load the YAML config first
	config, err := loadConfig("triggerx_operator.yaml")
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Update the URLs from the YAML config
	c.EthRpcUrl = config.Environment.EthRpcUrl
	c.EthWsUrl = config.Environment.EthWsUrl

	// Update the addresses from the YAML config
	c.ServiceManagerAddress = config.Addresses.ServiceManagerAddress
	c.OperatorStateRetrieverAddress = config.Addresses.OperatorStateRetriever

	// Update metrics configuration from YAML
	if config.Prometheus.PortAddress != "" {
		c.EigenMetricsIpPortAddress = config.Prometheus.PortAddress
	}

	// Add validation for required fields
	if c.ServiceManagerAddress == "" {
		return nil, fmt.Errorf("ServiceManagerAddress is empty in configuration")
	}
	if c.OperatorStateRetrieverAddress == "" {
		return nil, fmt.Errorf("OperatorStateRetriever address is empty in configuration")
	}
	if c.EigenMetricsIpPortAddress == "" {
		return nil, fmt.Errorf("Prometheus metrics ip port address is empty in configuration")
	}
	if c.EthRpcUrl == "" {
		return nil, fmt.Errorf("EthRpcUrl is empty in configuration")
	}
	if c.EthWsUrl == "" {
		return nil, fmt.Errorf("EthWsUrl is empty in configuration")
	}

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
	reg := prometheus.NewRegistry()
	eigenMetrics := sdkmetrics.NewEigenMetrics(config.AVS_NAME, c.EigenMetricsIpPortAddress, reg, logger)
	avsAndEigenMetrics := metrics.NewAvsAndEigenMetrics(config.AVS_NAME, eigenMetrics, reg)

	// Setup Node Api
	nodeApi := nodeapi.NewNodeApi(config.AVS_NAME, config.SEM_VER, c.NodeApiIpPortAddress, logger)

	var ethRpcClient, ethWsClient sdkcommon.EthClientInterface
	if c.EnableMetrics {
		rpcCallsCollector := rpccalls.NewCollector(config.AVS_NAME, reg)
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

	// blsKeyPassword, ok := os.LookupEnv("OPERATOR_BLS_KEY_PASSWORD")
	// if !ok {
	// 	logger.Warnf("OPERATOR_BLS_KEY_PASSWORD env var not set. using empty string")
	// }
	blsKeyPassword := "pL6!oK5@iJ4#"
	blsKeyPair, err := bls.ReadPrivateKeyFromFile(c.BlsPrivateKeyStorePath, blsKeyPassword)
	if err != nil {
		logger.Errorf("Cannot parse bls private key", "err", err)
		return nil, err
	}
	// TODO(samlaf): should we add the chainId to the config instead?
	// this way we can prevent creating a signer that signs on mainnet by mistake
	// if the config says chainId=5, then we can only create a goerli signer
	chainId, err := ethRpcClient.ChainID(context.Background())
	if err != nil {
		logger.Error("Cannot get chainId", "err", err)
		return nil, err
	}

	// ecdsaKeyPassword, ok := os.LookupEnv("OPERATOR_ECDSA_KEY_PASSWORD")
	// if !ok {
	// 	logger.Warnf("OPERATOR_ECDSA_KEY_PASSWORD env var not set. using empty string")
	// }
	ecdsaKeyPassword := "pL6!oK5@iJ4#"

	signerV2, _, err := signerv2.SignerFromConfig(signerv2.Config{
		KeystorePath: c.EcdsaPrivateKeyStorePath,
		Password:     ecdsaKeyPassword,
	}, chainId)
	if err != nil {
		panic(err)
	}
	chainioConfig := clients.BuildAllConfig{
		EthHttpUrl:                 c.EthRpcUrl,
		EthWsUrl:                   c.EthWsUrl,
		RegistryCoordinatorAddr:    c.ServiceManagerAddress,
		OperatorStateRetrieverAddr: c.OperatorStateRetrieverAddress,
		AvsName:                    config.AVS_NAME,
		PromMetricsIpPortAddress:   c.EigenMetricsIpPortAddress,
	}
	operatorEcdsaPrivateKey, err := sdkecdsa.ReadKey(
		c.EcdsaPrivateKeyStorePath,
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

	avsReader, err := chainio.BuildAvsReader(
		common.HexToAddress(c.ServiceManagerAddress),
		common.HexToAddress(c.OperatorStateRetrieverAddress),
		ethRpcClient, logger)
	if err != nil {
		logger.Error("Cannot create AvsReader", "err", err)
		return nil, err
	}
	avsWriter, err := chainio.BuildAvsWriter(txMgr, common.HexToAddress(c.ServiceManagerAddress),
		common.HexToAddress(c.OperatorStateRetrieverAddress), ethRpcClient, logger,
	)
	if err != nil {
		logger.Error("Cannot create AvsWriter", "err", err)
		return nil, err
	}
	avsSubscriber, err := chainio.BuildAvsSubscriber(common.HexToAddress(c.ServiceManagerAddress),
		common.HexToAddress(c.OperatorStateRetrieverAddress), ethWsClient, logger,
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
		config.AVS_NAME, logger, common.HexToAddress(c.KeeperAddress), quorumNames)
	reg.MustRegister(economicMetricsCollector)

	keeper := &Keeper{
		Config:                     c,
		Logger:                     logger,
		EthClient:                  ethRpcClient,
		MetricsReg:                 reg,
		Metrics:                    avsAndEigenMetrics,
		NodeApi:                    nodeApi,
		AvsReader:                  avsReader,
		AvsWriter:                  avsWriter,
		AvsSubscriber:              avsSubscriber,
		EigenlayerReader:           *sdkClients.ElChainReader,
		EigenlayerWriter:           *sdkClients.ElChainWriter,
		BlsKeypair:                 blsKeyPair,
		KeeperId:                   [32]byte{0}, // this will be set after registration
		KeeperAddr:                 common.HexToAddress(c.KeeperAddress),
		NewTaskCreatedChan:         make(chan *txtaskmanager.ContractTriggerXTaskManagerTaskCreated),
		ValidatorServerIpPortAddr:  "",
		TriggerxServiceManagerAddr: common.HexToAddress(c.ServiceManagerAddress),
	}

	// OperatorId is set in contract during registration so we get it after registering operator.
	operatorId, err := sdkClients.AvsRegistryChainReader.GetOperatorId(&bind.CallOpts{}, keeper.KeeperAddr)
	if err != nil {
		logger.Error("Cannot get operator id", "err", err)
		return nil, err
	}
	keeper.KeeperId = operatorId
	logger.Info("Operator info",
		"operatorId", operatorId,
		"operatorAddr", c.KeeperAddress,
		"operatorG1Pubkey", keeper.BlsKeypair.GetPubKeyG1(),
		"operatorG2Pubkey", keeper.BlsKeypair.GetPubKeyG2(),
	)

	return keeper, nil

}

// func (o *Keeper) Start(ctx context.Context) error {
// 	operatorIsRegistered, err := o.avsReader.IsOperatorRegistered(&bind.CallOpts{}, o.keeperAddr)
// 	if err != nil {
// 		o.logger.Error("Error checking if operator is registered", "err", err)
// 		return err
// 	}
// 	if !operatorIsRegistered {
// 		// We bubble the error all the way up instead of using logger.Fatal because logger.Fatal prints a huge stack trace
// 		// that hides the actual error message. This error msg is more explicit and doesn't require showing a stack trace to the user.
// 		return fmt.Errorf("operator is not registered. Registering operator using the operator-cli before starting operator")
// 	}

// 	o.logger.Infof("Starting operator.")

// 	if o.config.EnableNodeApi {
// 		o.nodeApi.Start()
// 	}
// 	var metricsErrChan <-chan error
// 	if o.config.EnableMetrics {
// 		metricsErrChan = o.metrics.Start(ctx, o.metricsReg)
// 	} else {
// 		metricsErrChan = make(chan error, 1)
// 	}

// 	// TODO(samlaf): wrap this call with increase in avs-node-spec metric
// 	sub := o.avsSubscriber.SubscribeToNewTasks(o.newTaskCreatedChan)
// 	for {
// 		select {
// 		case <-ctx.Done():
// 			return nil
// 		case err := <-metricsErrChan:
// 			// TODO(samlaf); we should also register the service as unhealthy in the node api
// 			// https://eigen.nethermind.io/docs/spec/api/
// 			o.logger.Fatal("Error in metrics server", "err", err)
// 		case err := <-sub.Err():
// 			o.logger.Error("Error in websocket subscription", "err", err)
// 			// TODO(samlaf): write unit tests to check if this fixed the issues we were seeing
// 			sub.Unsubscribe()
// 			// TODO(samlaf): wrap this call with increase in avs-node-spec metric
// 			sub = o.avsSubscriber.SubscribeToNewTasks(o.newTaskCreatedChan)
// 		case newTaskCreatedLog := <-o.newTaskCreatedChan:
// 			o.metrics.IncNumTasksReceived()
// 			taskResponse := o.ProcessNewTaskCreatedLog(newTaskCreatedLog)
// 			signedTaskResponse, err := o.SignTaskResponse(taskResponse)
// 			if err != nil {
// 				continue
// 			}
// 			go o.aggregatorRpcClient.SendSignedTaskResponseToAggregator(signedTaskResponse)
// 		}
// 	}
// }

// // Takes a NewTaskCreatedLog struct as input and returns a TaskResponseHeader struct.
// // The TaskResponseHeader struct is the struct that is signed and sent to the contract as a task response.
// func (o *Operator) ProcessNewTaskCreatedLog(newTaskCreatedLog *cstaskmanager.ContractIncredibleSquaringTaskManagerNewTaskCreated) *cstaskmanager.IIncredibleSquaringTaskManagerTaskResponse {
// 	o.logger.Debug("Received new task", "task", newTaskCreatedLog)
// 	o.logger.Info("Received new task",
// 		"numberToBeSquared", newTaskCreatedLog.Task.NumberToBeSquared,
// 		"taskIndex", newTaskCreatedLog.TaskIndex,
// 		"taskCreatedBlock", newTaskCreatedLog.Task.TaskCreatedBlock,
// 		"quorumNumbers", newTaskCreatedLog.Task.QuorumNumbers,
// 		"QuorumThresholdPercentage", newTaskCreatedLog.Task.QuorumThresholdPercentage,
// 	)
// 	numberSquared := big.NewInt(0).Exp(newTaskCreatedLog.Task.NumberToBeSquared, big.NewInt(2), nil)
// 	taskResponse := &cstaskmanager.IIncredibleSquaringTaskManagerTaskResponse{
// 		ReferenceTaskIndex: newTaskCreatedLog.TaskIndex,
// 		NumberSquared:      numberSquared,
// 	}
// 	return taskResponse
// }

// func (o *Operator) SignTaskResponse(taskResponse *cstaskmanager.IIncredibleSquaringTaskManagerTaskResponse) (*aggregator.SignedTaskResponse, error) {
// 	taskResponseHash, err := core.GetTaskResponseDigest(taskResponse)
// 	if err != nil {
// 		o.logger.Error("Error getting task response header hash. skipping task (this is not expected and should be investigated)", "err", err)
// 		return nil, err
// 	}
// 	blsSignature := o.blsKeypair.SignMessage(taskResponseHash)
// 	signedTaskResponse := &aggregator.SignedTaskResponse{
// 		TaskResponse: *taskResponse,
// 		BlsSignature: *blsSignature,
// 		OperatorId:   o.operatorId,
// 	}
// 	o.logger.Debug("Signed task response", "signedTaskResponse", signedTaskResponse)
// 	return signedTaskResponse, nil
// }

// func (k *Keeper) registerKeeperOnStartup(
// 	operatorEcdsaPrivateKey *sdkecdsa.PrivateKey,
// 	mockTokenStrategyAddr common.Address,
// ) {
// 	err := k.RegisterOperatorWithEigenlayer()
// 	if err != nil {
// 		k.logger.Error("Error registering operator with eigenlayer", "err", err)
// 	} else {
// 		k.logger.Infof("Registered operator with eigenlayer")
// 	}

// 	amount := big.NewInt(1000)
// 	err = k.DepositIntoStrategy(mockTokenStrategyAddr, amount)
// 	if err != nil {
// 		k.logger.Fatal("Error depositing into strategy", "err", err)
// 	}
// 	k.logger.Infof("Deposited %s into strategy %s", amount, mockTokenStrategyAddr)

// 	err = k.RegisterOperatorWithAvs(operatorEcdsaPrivateKey)
// 	if err != nil {
// 		k.logger.Fatal("Error registering operator with avs", "err", err)
// 	}
// 	k.logger.Infof("Registered operator with avs")
// }
