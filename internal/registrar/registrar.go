package registrar

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/trigg3rX/triggerx-backend/internal/registrar/config"
	"github.com/trigg3rX/triggerx-backend/internal/registrar/events"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

const (
	defaultConnectTimeout = 30 * time.Second
	defaultBlockOverlap   = uint64(5)
)

// BlockState tracks the last processed block numbers
type BlockState struct {
	lastProcessedBlockEth  uint64
	lastProcessedBlockBase uint64
	mu                     sync.RWMutex
}

func (bs *BlockState) updateEthBlock(block uint64) {
	bs.mu.Lock()
	defer bs.mu.Unlock()
	bs.lastProcessedBlockEth = block
}

func (bs *BlockState) updateBaseBlock(block uint64) {
	bs.mu.Lock()
	defer bs.mu.Unlock()
	bs.lastProcessedBlockBase = block
}

func (bs *BlockState) getEthBlock() uint64 {
	bs.mu.RLock()
	defer bs.mu.RUnlock()
	return bs.lastProcessedBlockEth
}

func (bs *BlockState) getBaseBlock() uint64 {
	bs.mu.RLock()
	defer bs.mu.RUnlock()
	return bs.lastProcessedBlockBase
}

// RegistrarService manages the event polling and processing
type RegistrarService struct {
	logger     logging.Logger
	ethClient  *ethclient.Client
	baseClient *ethclient.Client
	blockState *BlockState
	stopChan   chan struct{}
	wg         sync.WaitGroup
}

// NewRegistrarService creates a new instance of RegistrarService
func NewRegistrarService(logger logging.Logger) (*RegistrarService, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultConnectTimeout)
	defer cancel()

	ethClient, err := ethclient.Dial(config.GetEthRPCURL())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Ethereum node: %w", err)
	}

	baseClient, err := ethclient.Dial(config.GetBaseRPCURL())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Base node: %w", err)
	}

	lastEthBlock, err := ethClient.BlockNumber(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get ETH latest block: %w", err)
	}

	lastBaseBlock, err := baseClient.BlockNumber(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get BASE latest block: %w", err)
	}

	return &RegistrarService{
		logger:     logger,
		ethClient:  ethClient,
		baseClient: baseClient,
		blockState: &BlockState{
			lastProcessedBlockEth:  lastEthBlock,
			lastProcessedBlockBase: lastBaseBlock,
		},
		stopChan: make(chan struct{}),
	}, nil
}

// Start begins the event polling service
func (s *RegistrarService) Start() {
	s.logger.Info("Starting registrar service...")

	s.wg.Add(1)
	go s.pollEvents()
}

// Stop gracefully stops the service
func (s *RegistrarService) Stop() {
	s.logger.Info("Stopping registrar service...")
	close(s.stopChan)
	s.wg.Wait()

	if s.ethClient != nil {
		s.ethClient.Close()
	}
	if s.baseClient != nil {
		s.baseClient.Close()
	}

	s.logger.Info("Registrar service stopped")
}

func (s *RegistrarService) pollEvents() {
	defer s.wg.Done()

	s.logger.Info("Polling for new events...")
	s.logger.Infof("Polling interval: %s", config.GetPollingInterval())

	ticker := time.NewTicker(config.GetPollingInterval())
	defer ticker.Stop()

	for {
		select {
		case <-s.stopChan:
			return
		case <-ticker.C:
			s.processPendingEvents()
		}
	}
}

func (s *RegistrarService) processPendingEvents() {
	s.logger.Debug("Polling for new events...")

	ctx, cancel := context.WithTimeout(context.Background(), defaultConnectTimeout)
	defer cancel()

	if err := s.processEthEvents(ctx); err != nil {
		s.logger.Error("Failed to process ETH events", "error", err)
	}

	if err := s.processBaseEvents(ctx); err != nil {
		s.logger.Error("Failed to process BASE events", "error", err)
	}
}

func (s *RegistrarService) processEthEvents(ctx context.Context) error {
	ethLatestBlock, err := s.ethClient.BlockNumber(ctx)
	if err != nil {
		return fmt.Errorf("failed to get ETH latest block: %w", err)
	}

	lastProcessed := s.blockState.getEthBlock()
	if ethLatestBlock <= lastProcessed {
		return nil
	}

	fromBlock := lastProcessed + 1
	s.logger.Debug("Processing ETH events",
		"fromBlock", fromBlock,
		"toBlock", ethLatestBlock,
	)

	avsAddr := common.HexToAddress(config.GetAvsGovernanceAddress())

	if err := events.ProcessOperatorRegisteredEvents(s.ethClient, avsAddr, fromBlock, ethLatestBlock, s.logger); err != nil {
		return fmt.Errorf("failed to process OperatorRegistered events: %w", err)
	}

	if err := events.ProcessOperatorUnregisteredEvents(s.ethClient, avsAddr, fromBlock, ethLatestBlock, s.logger); err != nil {
		return fmt.Errorf("failed to process OperatorUnregistered events: %w", err)
	}

	s.blockState.updateEthBlock(ethLatestBlock)
	return nil
}

func (s *RegistrarService) processBaseEvents(ctx context.Context) error {
	baseLatestBlock, err := s.baseClient.BlockNumber(ctx)
	if err != nil {
		return fmt.Errorf("failed to get BASE latest block: %w", err)
	}

	lastProcessed := s.blockState.getBaseBlock()
	if baseLatestBlock <= lastProcessed {
		return nil
	}

	fromBlock := lastProcessed
	if fromBlock > defaultBlockOverlap {
		fromBlock -= defaultBlockOverlap
	}

	s.logger.Debug("Processing BASE events",
		"fromBlock", fromBlock,
		"toBlock", baseLatestBlock,
	)

	attAddr := common.HexToAddress(config.GetAttestationCenterAddress())

	if err := events.ProcessTaskSubmittedEvents(s.baseClient, attAddr, fromBlock, baseLatestBlock, s.logger); err != nil {
		return fmt.Errorf("failed to process TaskSubmitted events: %w", err)
	}

	if err := events.ProcessTaskRejectedEvents(s.baseClient, attAddr, fromBlock, baseLatestBlock, s.logger); err != nil {
		return fmt.Errorf("failed to process TaskRejected events: %w", err)
	}

	s.blockState.updateBaseBlock(baseLatestBlock)
	return nil
}
