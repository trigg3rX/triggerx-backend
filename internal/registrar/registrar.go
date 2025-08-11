package registrar

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/trigg3rX/triggerx-backend/internal/registrar/clients/database"
	"github.com/trigg3rX/triggerx-backend/internal/registrar/config"
	"github.com/trigg3rX/triggerx-backend/internal/registrar/events"
	"github.com/trigg3rX/triggerx-backend/internal/registrar/rewards"
	syncMgr "github.com/trigg3rX/triggerx-backend/internal/registrar/sync"
	"github.com/trigg3rX/triggerx-backend/pkg/client/redis"
	dbClient "github.com/trigg3rX/triggerx-backend/pkg/database"
	"github.com/trigg3rX/triggerx-backend/pkg/ipfs"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

const (
	defaultConnectTimeout = 30 * time.Second
	defaultBlockOverlap   = uint64(5)
)

// RegistrarService manages the event polling and WebSocket listening
type RegistrarService struct {
	logger logging.Logger

	// Event listener
	eventListener *events.ContractEventListener

	// State management
	stateManager *syncMgr.StateManager

	// Rewards service
	rewardsService *rewards.RewardsService

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

	// Initialize Redis client first
	redis, err := redis.NewRedisClient(logger, redis.RedisConfig{
		UpstashConfig: redis.UpstashConfig{
			URL:   config.GetUpstashRedisUrl(),
			Token: config.GetUpstashRedisRestToken(),
		},
		ConnectionSettings: redis.ConnectionSettings{
			PoolSize:      10,
			MaxIdleConns:  10,
			MinIdleConns:  1,
			MaxRetries:    3,
			DialTimeout:   5 * time.Second,
			ReadTimeout:   5 * time.Second,
			WriteTimeout:  5 * time.Second,
			PoolTimeout:   5 * time.Second,
			PingTimeout:   5 * time.Second,
			HealthTimeout: 5 * time.Second,
		},
	})
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to initialize Redis client: %w", err)
	}

	// Initialize state manager
	stateManager := syncMgr.NewStateManager(redis, logger)

	// Try to load existing state from Redis first
	initCtx, initCancel := context.WithTimeout(ctx, defaultConnectTimeout)
	lastEthBlock, err := stateManager.GetLastEthBlockUpdated(initCtx)
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

	lastBaseBlock, err := stateManager.GetLastBaseBlockUpdated(initCtx)
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
	initCancel()

	// Ensure state is initialized in Redis (will only set if not already present)
	if err := stateManager.InitializeState(ctx, lastEthBlock, lastBaseBlock, 0, time.Now().UTC()); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to initialize blockchain state: %w", err)
	}

	dbCfg := &dbClient.Config{
		Hosts:       []string{config.GetDatabaseHostAddress() + ":" + config.GetDatabaseHostPort()},
		Keyspace:    "triggerx",
		Timeout:     10 * time.Second,
		Retries:     3,
		ConnectWait: 5 * time.Second,
	}
	dbConn, err := dbClient.NewConnection(dbCfg, logger)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to initialize database client: %w", err)
	}

	// Initialize database client
	databaseClient := database.NewDatabaseClient(logger, dbConn)

	ipfsCfg := ipfs.NewConfig(config.GetPinataHost(), config.GetPinataJWT())
	ipfsClient, err := ipfs.NewClient(ipfsCfg, logger)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to initialize IPFS client: %w", err)
	}
	eventListener := events.NewContractEventListener(logger, events.GetDefaultConfig(), databaseClient, ipfsClient)

	// Initialize rewards service
	rewardsService := rewards.NewRewardsService(logger, stateManager, databaseClient)

	return &RegistrarService{
		logger:         logger,
		eventListener:  eventListener,
		stateManager:   stateManager,
		rewardsService: rewardsService,
		ctx:            ctx,
		cancel:         cancel,
		stopChan:       make(chan struct{}),
	}, nil
}

// Start begins the event monitoring service
func (s *RegistrarService) Start() error {
	s.logger.Info("Starting registrar service...")

	// Start Rewards Service (if initialized)
	if s.rewardsService != nil {
		go s.rewardsService.StartDailyRewardsPoints()
	} else {
		s.logger.Info("Rewards service not initialized (database client not available)")
	}

	// Start event listener
	if err := s.eventListener.Start(); err != nil {
		s.logger.Errorf("Failed to start event listener: %v", err)
		s.logger.Info("Falling back to polling mode")
	}

	s.logger.Info("Registrar service started successfully")
	return nil
}

// Stop gracefully stops the service
func (s *RegistrarService) Stop() error {
	s.logger.Info("Stopping registrar service...")

	// Signal all goroutines to stop
	s.cancel()
	close(s.stopChan)

	// Wait for all goroutines with timeout
	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		s.logger.Info("All goroutines stopped successfully")
	case <-time.After(30 * time.Second):
		s.logger.Warn("Timeout waiting for goroutines to stop")
	}

	// Stop event listener
	if err := s.eventListener.Stop(); err != nil {
		s.logger.Errorf("Error stopping event listener: %v", err)
	}

	s.logger.Info("Registrar service stopped")
	return nil
}
