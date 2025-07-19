package registrar

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"

	// "github.com/trigg3rX/triggerx-backend/internal/registrar/clients/database"
	redisClient "github.com/trigg3rX/triggerx-backend/internal/registrar/clients/redis"
	"github.com/trigg3rX/triggerx-backend/internal/registrar/config"
	"github.com/trigg3rX/triggerx-backend/internal/registrar/events"
	// "github.com/trigg3rX/triggerx-backend/internal/registrar/rewards"
	syncMgr "github.com/trigg3rX/triggerx-backend/internal/registrar/sync"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

const (
	defaultConnectTimeout = 30 * time.Second
	defaultBlockOverlap   = uint64(5)
)

// RegistrarService manages the event polling and WebSocket listening
type RegistrarService struct {
	logger logging.Logger

	// Blockchain clients
	ethClient  *ethclient.Client
	baseClient *ethclient.Client
	optClient  *ethclient.Client

	// Redis client
	redis *redisClient.Client

	// Event listener
	eventListener *events.ContractEventListener

	// Clients
	// databaseClient *database.DatabaseClient

	// State management
	stateManager      *syncMgr.StateManager
	checkpointManager *syncMgr.CheckpointManager
	backfillManager   *syncMgr.BackfillManager

	// Rewards service
	// rewardsService *rewards.RewardsService

	// Event handlers
	// operatorHandler *events.OperatorEventHandler
	// taskHandler     *events.TaskEventHandler

	// Lifecycle management
	ctx      context.Context
	cancel   context.CancelFunc
	wg       sync.WaitGroup
	stopChan chan struct{}
}

// NewRegistrarService creates a new instance of RegistrarService
func NewRegistrarService(logger logging.Logger) (*RegistrarService, error) {
	ctx, cancel := context.WithCancel(context.Background())

	// Initialize HTTP clients for fallback polling
	ethClient, err := ethclient.Dial(config.GetChainRPCUrl(true, "17000"))
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to connect to Ethereum node: %w", err)
	}

	baseClient, err := ethclient.Dial(config.GetChainRPCUrl(true, "84532"))
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to connect to Base node: %w", err)
	}

	optClient, err := ethclient.Dial(config.GetChainRPCUrl(true, "11155420"))
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to connect to Optimism node: %w", err)
	}

	// Initialize Redis client first
	redis, err := redisClient.NewClient(logger)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to initialize Redis client: %w", err)
	}

	// Initialize state manager
	stateManager := syncMgr.NewStateManager(redis, logger)

	// Initialize checkpoint manager
	checkpointManager := syncMgr.NewCheckpointManager(redis, logger)

	// Initialize backfill manager
	backfillManager := syncMgr.NewBackfillManager(ethClient, baseClient, optClient, logger)

	// Try to load existing state from Redis first
	initCtx, initCancel := context.WithTimeout(ctx, defaultConnectTimeout)
	lastEthBlock, err := stateManager.GetLastPolledEthBlock(initCtx)
	if err != nil || lastEthBlock == 0 {
		// Fallback to current blockchain block if Redis is empty
		logger.Info("No ETH block found in Redis, getting current block from blockchain")
		lastEthBlock, err = ethClient.BlockNumber(initCtx)
		if err != nil {
			initCancel()
			cancel()
			return nil, fmt.Errorf("failed to get ETH latest block: %w", err)
		}
	} else {
		logger.Infof("Loaded ETH block %d from Redis", lastEthBlock)
	}

	lastBaseBlock, err := stateManager.GetLastPolledBaseBlock(initCtx)
	if err != nil || lastBaseBlock == 0 {
		logger.Info("No BASE block found in Redis, getting current block from blockchain")
		lastBaseBlock, err = baseClient.BlockNumber(initCtx)
		if err != nil {
			initCancel()
			cancel()
			return nil, fmt.Errorf("failed to get BASE latest block: %w", err)
		}
	} else {
		logger.Infof("Loaded BASE block %d from Redis", lastBaseBlock)
	}

	lastOptBlock, err := stateManager.GetLastPolledOptBlock(initCtx)
	if err != nil || lastOptBlock == 0 {
		logger.Info("No OPT block found in Redis, getting current block from blockchain")
		lastOptBlock, err = optClient.BlockNumber(initCtx)
		if err != nil {
			initCancel()
			cancel()
			return nil, fmt.Errorf("failed to get OPT latest block: %w", err)
		}
	} else {
		logger.Infof("Loaded OPT block %d from Redis", lastOptBlock)
	}
	initCancel()

	// Initialize event listener with configuration
	eventConfig := events.GetDefaultConfig()
	eventListener := events.NewContractEventListener(logger, eventConfig)

	// TODO: Initialize event handlers properly - need constructor functions

	// Ensure state is initialized in Redis (will only set if not already present)
	if err := stateManager.InitializeState(ctx, lastEthBlock, lastBaseBlock, lastOptBlock); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to initialize blockchain state: %w", err)
	}

	return &RegistrarService{
		logger:            logger,
		ethClient:         ethClient,
		baseClient:        baseClient,
		optClient:         optClient,
		redis:             redis,
		eventListener:     eventListener,
		stateManager:      stateManager,
		checkpointManager: checkpointManager,
		backfillManager:   backfillManager,
		ctx:               ctx,
		cancel:            cancel,
		stopChan:          make(chan struct{}),
	}, nil
}

// Start begins the event monitoring service
func (s *RegistrarService) Start() error {
	s.logger.Info("Starting registrar service...")

	// Start event listener
	if err := s.eventListener.Start(); err != nil {
		s.logger.Errorf("Failed to start event listener: %v", err)
		s.logger.Info("Falling back to polling mode")
	}

	// Start event processor
	s.wg.Add(1)
	go s.processEvents()

	// Start fallback polling (in case event listener fails)
	s.wg.Add(1)
	go s.pollEventsFallback()

	// Start periodic checkpoints
	s.wg.Add(1)
	go s.checkpointManager.StartPeriodicCheckpoints(s.ctx, s.stateManager)

	// Perform initial backfill check for missing blocks
	go func() {
		if err := s.backfillManager.BackfillMissingBlocks(s.ctx, s.stateManager); err != nil {
			s.logger.Errorf("Initial backfill failed: %v", err)
		}
	}()

	s.logger.Info("Registrar service started successfully")
	return nil
}

// Stop gracefully stops the service
func (s *RegistrarService) Stop() error {
	s.logger.Info("Stopping registrar service...")

	// Signal all goroutines to stop
	s.cancel()
	close(s.stopChan)

	// Stop event listener
	if err := s.eventListener.Stop(); err != nil {
		s.logger.Errorf("Error stopping event listener: %v", err)
	}

	// Wait for all goroutines to finish
	s.wg.Wait()

	// Close HTTP clients
	if s.ethClient != nil {
		s.ethClient.Close()
	}
	if s.baseClient != nil {
		s.baseClient.Close()
	}
	if s.optClient != nil {
		s.optClient.Close()
	}

	s.logger.Info("Registrar service stopped")
	return nil
}

// processEvents processes incoming events from the event listener
func (s *RegistrarService) processEvents() {
	defer s.wg.Done()

	s.logger.Info("Starting event processor...")

	for {
		select {
		case <-s.ctx.Done():
			return
		default:
			// Process events using the event listener's status
			status := s.eventListener.GetStatus()
			s.logger.Debugf("Event listener status: %+v", status)

			// Sleep briefly to avoid tight loop
			time.Sleep(100 * time.Millisecond)
		}
	}
}

// pollEventsFallback provides fallback polling when event listener fails
func (s *RegistrarService) pollEventsFallback() {
	defer s.wg.Done()

	s.logger.Info("Starting fallback polling...")
	ticker := time.NewTicker(config.GetPollingInterval())
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-s.stopChan:
			return
		case <-ticker.C:
			// Check if event listener is running
			status := s.eventListener.GetStatus()
			if running, ok := status["running"].(bool); !ok || !running {
				s.logger.Warn("Event listener not running, using fallback polling")
				s.pollAllChains()
			}
		}
	}
}

// pollAllChains polls events for all chains
func (s *RegistrarService) pollAllChains() {
	chains := []string{"17000", "84532", "11155420"}
	for _, chainID := range chains {
		s.pollChainEvents(chainID)
	}
}

// pollChainEvents polls events for a specific chain (fallback method)
func (s *RegistrarService) pollChainEvents(chainID string) {
	ctx, cancel := context.WithTimeout(s.ctx, defaultConnectTimeout)
	defer cancel()

	switch chainID {
	case "17000":
		if err := s.processEthEvents(ctx); err != nil {
			s.logger.Errorf("Failed to process ETH events: %v", err)
		}
	case "84532":
		if err := s.processBaseEvents(ctx); err != nil {
			s.logger.Errorf("Failed to process BASE events: %v", err)
		}
	case "11155420":
		if err := s.processOptEvents(ctx); err != nil {
			s.logger.Errorf("Failed to process OPT events: %v", err)
		}
	}
}

// Legacy polling methods (kept for fallback)
func (s *RegistrarService) processEthEvents(ctx context.Context) error {
	return s.processChainEventsRange(ctx, "17000", s.ethClient, s.stateManager.GetLastPolledEthBlock, s.stateManager.SetLastPolledEthBlock, 1)
}

func (s *RegistrarService) processBaseEvents(ctx context.Context) error {
	return s.processChainEventsRange(ctx, "84532", s.baseClient, s.stateManager.GetLastPolledBaseBlock, s.stateManager.SetLastPolledBaseBlock, defaultBlockOverlap)
}

func (s *RegistrarService) processOptEvents(ctx context.Context) error {
	return s.processChainEventsRange(ctx, "11155420", s.optClient, s.stateManager.GetLastPolledOptBlock, s.stateManager.SetLastPolledOptBlock, defaultBlockOverlap)
}

// processChainEventsRange handles event processing for a chain using the unified architecture
func (s *RegistrarService) processChainEventsRange(
	ctx context.Context,
	chainID string,
	client *ethclient.Client,
	getLastBlock func(context.Context) (uint64, error),
	setLastBlock func(context.Context, uint64) error,
	blockOverlap uint64,
) error {
	latestBlock, err := client.BlockNumber(ctx)
	if err != nil {
		return fmt.Errorf("failed to get latest block for chain %s: %w", chainID, err)
	}

	lastProcessed, err := getLastBlock(ctx)
	if err != nil {
		return fmt.Errorf("failed to get last processed block for chain %s: %w", chainID, err)
	}

	if latestBlock <= lastProcessed {
		return nil
	}

	fromBlock := lastProcessed
	if chainID != "17000" && fromBlock > blockOverlap { // ETH doesn't use overlap
		fromBlock -= blockOverlap
	} else if chainID == "17000" {
		fromBlock = lastProcessed + 1
	}

	s.logger.Debugf("Polling %s events from block %d to %d", chainID, fromBlock, latestBlock)

	// Update block state directly in fallback mode
	if err := setLastBlock(ctx, latestBlock); err != nil {
		return fmt.Errorf("failed to update %s block state: %w", chainID, err)
	}

	return nil
}

// GetStatus returns the current status of the registrar service
func (s *RegistrarService) GetStatus() map[string]interface{} {
	chainStatus := s.eventListener.GetStatus()
	ethBlock, err := s.stateManager.GetLastPolledEthBlock(s.ctx)
	if err != nil {
		s.logger.Errorf("Failed to get last processed ETH block: %v", err)
	}
	baseBlock, err := s.stateManager.GetLastPolledBaseBlock(s.ctx)
	if err != nil {
		s.logger.Errorf("Failed to get last processed BASE block: %v", err)
	}
	optBlock, err := s.stateManager.GetLastPolledOptBlock(s.ctx)
	if err != nil {
		s.logger.Errorf("Failed to get last processed OPT block: %v", err)
	}

	// Get checkpoint and backfill health
	checkpointHealth := s.checkpointManager.GetCheckpointHealth(s.ctx)
	backfillHealth := s.backfillManager.GetBackfillHealth()

	return map[string]interface{}{
		"service": "registrar",
		"chains":  chainStatus,
		"block_state": map[string]interface{}{
			"ethereum": ethBlock,
			"base":     baseBlock,
			"optimism": optBlock,
		},
		"websocket_active": len(chainStatus) > 0,
		"checkpoints":      checkpointHealth,
		"backfill":         backfillHealth,
		"redis_health":     s.stateManager.GetStateHealth(s.ctx),
	}
}
