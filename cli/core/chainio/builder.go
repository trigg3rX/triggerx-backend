package chainio

import (
	gethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/imua-xyz/imua-avs-sdk/client/txmgr"
	"github.com/imua-xyz/imua-avs-sdk/logging"
	"github.com/imua-xyz/imua-avs-sdk/signer"
	"github.com/trigg3rX/triggerx-backend-imua/cli/core/chainio/eth"
)

type BuildAllConfig struct {
	EthHttpUrl string
	EthWsUrl   string
	AvsAddr    string
	AvsName    string
}

type Clients struct {
	AvsRegistryChainSubscriber *AvsRegistryChainSubscriber
	ChainReader                *ChainReader
	ChainWriter                *ChainWriter
	EthHttpClient              *eth.Client
	EthWsClient                *eth.Client
}

func BuildAll(
	config BuildAllConfig,
	signerAddr gethcommon.Address,
	signerFn signer.SignerFn,
	logger logging.Logger,
) (*Clients, error) {
	config.validate(logger)

	// creating two types of Eth clients: HTTP and WS
	ethHttpClient, err := eth.NewClient(config.EthHttpUrl)
	if err != nil {
		logger.Error("Failed to create Eth Http client", "err", err)
		return nil, err
	}

	ethWsClient, err := eth.NewClient(config.EthWsUrl)
	if err != nil {
		logger.Error("Failed to create Eth WS client", "err", err)
		return nil, err
	}

	txMgr := txmgr.NewSimpleTxManager(ethHttpClient, logger, signerFn, signerAddr)
	// creating  clients: Reader, Writer and Subscriber
	chainReader, chainWriter, avsRegistrySubscriber, err := config.buildClients(
		ethHttpClient,
		txMgr,
		logger,
	)
	if err != nil {
		logger.Error("Failed to create Reader, Writer and Subscriber", "err", err)
		return nil, err
	}

	return &Clients{
		ChainReader:                chainReader,
		ChainWriter:                chainWriter,
		AvsRegistryChainSubscriber: avsRegistrySubscriber,
		EthHttpClient:              ethHttpClient,
		EthWsClient:                ethWsClient,
	}, nil

}

func (config *BuildAllConfig) buildClients(
	ethHttpClient eth.EthClient,
	txMgr txmgr.TxManager,
	logger logging.Logger,
) (*ChainReader, *ChainWriter, *AvsRegistryChainSubscriber, error) {
	contractBindings, err := NewContractBindings(
		gethcommon.HexToAddress(config.AvsAddr),
		ethHttpClient,
		logger,
	)
	if err != nil {
		logger.Error("Failed to create ContractBindings", "err", err)
		return nil, nil, nil, err
	}

	// get the Reader for the chain contracts
	chainReader := NewChainReader(
		*contractBindings.AVSManager,
		logger,
		ethHttpClient,
	)

	chainWriter := NewChainWriter(
		*contractBindings.AVSManager,
		chainReader,
		ethHttpClient,
		logger,
		txMgr,
	)
	if err != nil {
		logger.Error("Failed to create ChainWriter", "err", err)
		return nil, nil, nil, err
	}

	avsRegistrySubscriber, err := BuildAvsRegistryChainSubscriber(
		contractBindings.AvsAddr,
		ethHttpClient,
		logger,
	)
	if err != nil {
		logger.Error("Failed to create ChainSubscriber", "err", err)
		return nil, nil, nil, err
	}
	return chainReader, chainWriter, avsRegistrySubscriber, err
}

// Very basic validation that makes sure all fields are nonempty
// we might eventually want more sophisticated validation, based on regexp,
// or use something like https://json-schema.org/ (?)
func (config *BuildAllConfig) validate(logger logging.Logger) {
	if config.EthHttpUrl == "" {
		logger.Fatalf("BuildAllConfig.validate: Missing eth http url")
	}
	if config.EthWsUrl == "" {
		logger.Fatalf("BuildAllConfig.validate: Missing eth ws url")
	}
	if config.AvsAddr == "" {
		logger.Fatalf("BuildAllConfig.validate: Missing bls registry coordinator address")
	}
	if config.AvsName == "" {
		logger.Fatalf("BuildAllConfig.validate: Missing avs name")
	}
}