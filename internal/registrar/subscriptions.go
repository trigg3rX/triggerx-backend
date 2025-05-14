package registrar

import (
	"context"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/trigg3rX/triggerx-backend/internal/registrar/config"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

var (
	logger = logging.GetLogger(logging.Development, logging.RegistrarProcess)

	ethClient  *ethclient.Client
	baseClient *ethclient.Client

	lastProcessedBlockEth  uint64
	lastProcessedBlockBase uint64
	blockProcessingMutex   sync.Mutex
)

func StartEventPolling(
	avsGovernanceAddress common.Address,
	attestationCenterAddress common.Address,
) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var err error
	ethClient, err = ethclient.Dial(config.EthRpcUrl)
	if err != nil {
		logger.Errorf("Failed to connect to Ethereum node: %v", err)
		return
	}
	logger.Debug("Ethereum node connected")
	defer ethClient.Close()

	baseClient, err = ethclient.Dial(config.BaseRpcUrl)
	if err != nil {
		logger.Errorf("Failed to connect to Base node: %v", err)
		return
	}
	logger.Debug("Base node connected")
	defer baseClient.Close()

	lastProcessedBlockEth, err = ethClient.BlockNumber(ctx)
	if err != nil {
		logger.Error("failed to get ETH latest block: %v", err)
	}

	lastProcessedBlockBase, err = baseClient.BlockNumber(ctx)
	if err != nil {
		logger.Error("failed to get BASE latest block: %v", err)
	}

	logger.Info("Starting event polling service...")

	ticker := time.NewTicker(config.PollingInterval)
	defer ticker.Stop()

	for range ticker.C {
		pollEvents(avsGovernanceAddress, attestationCenterAddress)
	}
}

func pollEvents(
	avsGovernanceAddress common.Address,
	attestationCenterAddress common.Address,
) {
	logger.Debug("Polling for new events...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	ethLatestBlock, err := ethClient.BlockNumber(ctx)
	if err != nil {
		logger.Errorf("Failed to get ETH latest block: %v", err)
		return
	}

	baseLatestBlock, err := baseClient.BlockNumber(ctx)
	if err != nil {
		logger.Errorf("Failed to get BASE latest block: %v", err)
		return
	}

	if ethLatestBlock > lastProcessedBlockEth {
		fromBlock := lastProcessedBlockEth + 1
		logger.Debugf("Polling AVSG from block %d to %d", fromBlock, ethLatestBlock)

		err = ProcessOperatorRegisteredEvents(ethClient, avsGovernanceAddress, fromBlock, ethLatestBlock)
		if err != nil {
			logger.Errorf("Failed to process OperatorRegistered events: %v", err)
		}

		err = ProcessOperatorUnregisteredEvents(ethClient, avsGovernanceAddress, fromBlock, ethLatestBlock)
		if err != nil {
			logger.Errorf("Failed to process OperatorUnregistered events: %v", err)
		}

		blockProcessingMutex.Lock()
		lastProcessedBlockEth = ethLatestBlock
		blockProcessingMutex.Unlock()
	}

	if baseLatestBlock > lastProcessedBlockBase {
		fromBlock := lastProcessedBlockBase
		overlap := uint64(5)
		if fromBlock > overlap {
			fromBlock -= overlap
		}

		toBlock := baseLatestBlock

		logger.Debugf("Polling AttC from block %d to %d", fromBlock, toBlock)

		err = ProcessTaskSubmittedEvents(baseClient, attestationCenterAddress, fromBlock, baseLatestBlock)
		if err != nil {
			logger.Errorf("Failed to process TaskSubmitted events: %v", err)
		}

		err = ProcessTaskRejectedEvents(baseClient, attestationCenterAddress, fromBlock, baseLatestBlock)
		if err != nil {
			logger.Errorf("Failed to process TaskRejected events: %v", err)
		}

		blockProcessingMutex.Lock()
		lastProcessedBlockBase = baseLatestBlock
		blockProcessingMutex.Unlock()
	}
}
