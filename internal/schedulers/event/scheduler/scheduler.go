package scheduler

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/trigg3rX/triggerx-backend/internal/cache"
	redisx "github.com/trigg3rX/triggerx-backend/internal/redis"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/event/client"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/event/config"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/event/metrics"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	schedulerTypes "github.com/trigg3rX/triggerx-backend/pkg/types"
)

const (
	blockConfirmations   = 3                // Wait for 3 block confirmations
	pollInterval         = 10 * time.Second // Poll every 10 seconds for new blocks
	workerTimeout        = 30 * time.Second // Timeout for worker operations
	maxRetries           = 3                // Max retries for failed operations
	performerLockTTL     = 15 * time.Minute // Lock duration for job execution
	blockCacheTTL        = 2 * time.Minute  // Cache TTL for block data
	eventCacheTTL        = 10 * time.Minute // Cache TTL for event data
	duplicateEventWindow = 30 * time.Second // Window to prevent duplicate event processing
)

// EventBasedScheduler manages individual job workers
type EventBasedScheduler struct {
	ctx          context.Context
	cancel       context.CancelFunc
	logger       logging.Logger
	workers      map[int64]*JobWorker // jobID -> worker
	workersMutex sync.RWMutex
	chainClients map[string]*ethclient.Client // chainID -> client
	clientsMutex sync.RWMutex
	dbClient     *client.DBServerClient
	cache        cache.Cache
	metrics      *metrics.Collector
	managerID    string
	maxWorkers   int
}

// JobWorker represents an individual worker watching a specific job
type JobWorker struct {
	job          *schedulerTypes.EventJobData
	client       *ethclient.Client
	logger       logging.Logger
	dbClient     *client.DBServerClient
	cache        cache.Cache
	ctx          context.Context
	cancel       context.CancelFunc
	eventSig     common.Hash
	contractAddr common.Address
	lastBlock    uint64
	isRunning    bool
	mutex        sync.RWMutex
	managerID    string
}

// JobScheduleRequest represents the request to schedule a new job
type JobScheduleRequest struct {
	JobID                         int64    `json:"job_id" binding:"required"`
	TimeFrame                     int64    `json:"time_frame"`
	Recurring                     bool     `json:"recurring"`
	TriggerChainID                string   `json:"trigger_chain_id" binding:"required"`
	TriggerContractAddress        string   `json:"trigger_contract_address" binding:"required"`
	TriggerEvent                  string   `json:"trigger_event" binding:"required"`
	TargetChainID                 string   `json:"target_chain_id" binding:"required"`
	TargetContractAddress         string   `json:"target_contract_address" binding:"required"`
	TargetFunction                string   `json:"target_function" binding:"required"`
	ABI                           string   `json:"abi"`
	ArgType                       int      `json:"arg_type"`
	Arguments                     []string `json:"arguments"`
	DynamicArgumentsScriptIPFSUrl string   `json:"dynamic_arguments_script_ipfs_url"`
}

// NewEventBasedScheduler creates a new instance of EventBasedScheduler
func NewEventBasedScheduler(managerID string, logger logging.Logger, dbClient *client.DBServerClient) (*EventBasedScheduler, error) {
	ctx, cancel := context.WithCancel(context.Background())

	maxWorkers := config.GetMaxWorkers()

	// Initialize cache with enhanced Redis support
	if err := cache.InitWithLogger(logger); err != nil {
		logger.Warnf("Failed to initialize cache: %v", err)
	}

	cacheInstance, err := cache.GetCache()
	if err != nil {
		logger.Warnf("Cache not available, running without cache: %v", err)
		cacheInstance = nil
	} else {
		// Log cache type and Redis availability
		cacheInfo := cache.GetCacheInfo()
		logger.Infof("Cache initialized: type=%s, redis_available=%v",
			cacheInfo["type"], cacheInfo["redis_available"])
	}

	scheduler := &EventBasedScheduler{
		ctx:          ctx,
		cancel:       cancel,
		logger:       logger,
		workers:      make(map[int64]*JobWorker),
		chainClients: make(map[string]*ethclient.Client),
		dbClient:     dbClient,
		cache:        cacheInstance,
		metrics:      metrics.NewCollector(),
		managerID:    managerID,
		maxWorkers:   maxWorkers,
	}

	// Initialize chain clients
	if err := scheduler.initChainClients(); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to initialize chain clients: %w", err)
	}

	// Start metrics collection
	scheduler.metrics.Start()

	// Add scheduler startup event to Redis stream (Redis is already initialized in main.go)
	if redisx.IsAvailable() {
		startupEvent := map[string]interface{}{
			"event_type":       "scheduler_startup",
			"manager_id":       managerID,
			"max_workers":      maxWorkers,
			"cache_available":  cacheInstance != nil,
			"redis_available":  redisx.IsAvailable(),
			"supported_chains": len(scheduler.chainClients),
			"started_at":       time.Now().Unix(),
		}

		if err := redisx.AddJobToStream(redisx.JobsReadyEventStream, startupEvent); err != nil {
			logger.Warnf("Failed to add scheduler startup event to Redis stream: %v", err)
		} else {
			logger.Info("Scheduler startup event added to Redis stream")
		}
	}

	scheduler.logger.Info("Event-based scheduler initialized",
		"max_workers", maxWorkers,
		"manager_id", managerID,
		"cache_available", cacheInstance != nil,
		"redis_available", redisx.IsAvailable(),
		"connected_chains", len(scheduler.chainClients),
	)

	return scheduler, nil
}

// initChainClients initializes RPC clients for supported chains
func (s *EventBasedScheduler) initChainClients() error {
	s.clientsMutex.Lock()
	defer s.clientsMutex.Unlock()

	for chainID, rpcURL := range config.GetChainRPCUrls() {
		client, err := ethclient.Dial(rpcURL)
		if err != nil {
			s.logger.Error("Failed to connect to chain", "chain_id", chainID, "rpc_url", rpcURL, "error", err)
			continue
		}

		// Test connection
		_, err = client.ChainID(context.Background())
		if err != nil {
			s.logger.Error("Failed to get chain ID", "chain_id", chainID, "error", err)
			client.Close()
			continue
		}

		s.chainClients[chainID] = client
		s.logger.Info("Connected to chain", "chain_id", chainID, "rpc_url", rpcURL)
	}

	if len(s.chainClients) == 0 {
		return fmt.Errorf("no chain clients initialized successfully")
	}

	return nil
}

// ScheduleJob creates and starts a new job worker
func (s *EventBasedScheduler) ScheduleJob(jobData *schedulerTypes.EventJobData) error {
	s.workersMutex.Lock()
	defer s.workersMutex.Unlock()

	startTime := time.Now()

	// Check if job is already scheduled
	if _, exists := s.workers[jobData.JobID]; exists {
		// Add job scheduling failure event to Redis stream
		if redisx.IsAvailable() {
			failureEvent := map[string]interface{}{
				"event_type":       "job_schedule_failed",
				"job_id":           jobData.JobID,
				"manager_id":       s.managerID,
				"error":            "job already scheduled",
				"trigger_chain_id": jobData.TriggerChainID,
				"contract_address": jobData.TriggerContractAddress,
				"trigger_event":    jobData.TriggerEvent,
				"failed_at":        startTime.Unix(),
			}
			err := redisx.AddJobToStream(redisx.JobsRetryEventStream, failureEvent)
			if err != nil {
				s.logger.Warnf("Failed to add job scheduling failure event to Redis stream: %v", err)
			}
		}
		return fmt.Errorf("job %d is already scheduled", jobData.JobID)
	}

	// Check if we've reached the maximum number of workers
	if len(s.workers) >= s.maxWorkers {
		// Add job scheduling failure event to Redis stream
		if redisx.IsAvailable() {
			failureEvent := map[string]interface{}{
				"event_type":       "job_schedule_failed",
				"job_id":           jobData.JobID,
				"manager_id":       s.managerID,
				"error":            fmt.Sprintf("maximum workers (%d) reached", s.maxWorkers),
				"current_workers":  len(s.workers),
				"max_workers":      s.maxWorkers,
				"trigger_chain_id": jobData.TriggerChainID,
				"contract_address": jobData.TriggerContractAddress,
				"trigger_event":    jobData.TriggerEvent,
				"failed_at":        startTime.Unix(),
			}
			err := redisx.AddJobToStream(redisx.JobsRetryEventStream, failureEvent)
			if err != nil {
				s.logger.Warnf("Failed to add job scheduling failure event to Redis stream: %v", err)
			}
		}
		return fmt.Errorf("maximum number of workers (%d) reached, cannot schedule job %d", s.maxWorkers, jobData.JobID)
	}

	// Get chain client
	s.clientsMutex.RLock()
	client, exists := s.chainClients[jobData.TriggerChainID]
	s.clientsMutex.RUnlock()

	if !exists {
		// Add job scheduling failure event to Redis stream
		if redisx.IsAvailable() {
			failureEvent := map[string]interface{}{
				"event_type":       "job_schedule_failed",
				"job_id":           jobData.JobID,
				"manager_id":       s.managerID,
				"error":            fmt.Sprintf("unsupported chain ID: %s", jobData.TriggerChainID),
				"trigger_chain_id": jobData.TriggerChainID,
				"contract_address": jobData.TriggerContractAddress,
				"trigger_event":    jobData.TriggerEvent,
				"failed_at":        startTime.Unix(),
			}
			err := redisx.AddJobToStream(redisx.JobsRetryEventStream, failureEvent)
			if err != nil {
				s.logger.Warnf("Failed to add job scheduling failure event to Redis stream: %v", err)
			}
		}
		return fmt.Errorf("unsupported chain ID: %s", jobData.TriggerChainID)
	}

	// Create job worker
	worker, err := s.createJobWorker(jobData, client)
	if err != nil {
		// Add job scheduling failure event to Redis stream
		if redisx.IsAvailable() {
			failureEvent := map[string]interface{}{
				"event_type":       "job_schedule_failed",
				"job_id":           jobData.JobID,
				"manager_id":       s.managerID,
				"error":            fmt.Sprintf("failed to create worker: %v", err),
				"trigger_chain_id": jobData.TriggerChainID,
				"contract_address": jobData.TriggerContractAddress,
				"trigger_event":    jobData.TriggerEvent,
				"failed_at":        startTime.Unix(),
			}
			if err := redisx.AddJobToStream(redisx.JobsRetryEventStream, failureEvent); err != nil {
				s.logger.Warnf("Failed to add job scheduling failure event to Redis stream: %v", err)
			}
		}
		return fmt.Errorf("failed to create job worker: %w", err)
	}

	// Store worker
	s.workers[jobData.JobID] = worker

	// Start worker
	go worker.start()

	// Update metrics
	metrics.JobsScheduled.Inc()
	metrics.JobsRunning.Inc()

	duration := time.Since(startTime)

	// Add comprehensive job scheduling success event to Redis stream
	if redisx.IsAvailable() {
		jobContext := map[string]interface{}{
			"event_type":               "job_scheduled",
			"job_id":                   jobData.JobID,
			"manager_id":               s.managerID,
			"trigger_chain_id":         jobData.TriggerChainID,
			"trigger_contract_address": jobData.TriggerContractAddress,
			"trigger_event":            jobData.TriggerEvent,
			"target_chain_id":          jobData.TargetChainID,
			"target_contract_address":  jobData.TargetContractAddress,
			"target_function":          jobData.TargetFunction,
			"recurring":                jobData.Recurring,
			"active_workers":           len(s.workers),
			"max_workers":              s.maxWorkers,
			"cache_available":          s.cache != nil,
			"scheduled_at":             startTime.Unix(),
			"duration_ms":              duration.Milliseconds(),
			"status":                   "scheduled",
		}

		if err := redisx.AddJobToStream(redisx.JobsReadyEventStream, jobContext); err != nil {
			s.logger.Warnf("Failed to add job scheduling event to Redis stream: %v", err)
		}
	}

	s.logger.Info("Job scheduled successfully",
		"job_id", jobData.JobID,
		"trigger_chain", jobData.TriggerChainID,
		"contract", jobData.TriggerContractAddress,
		"event", jobData.TriggerEvent,
		"target_chain", jobData.TargetChainID,
		"target_contract", jobData.TargetContractAddress,
		"target_function", jobData.TargetFunction,
		"active_workers", len(s.workers),
		"max_workers", s.maxWorkers,
		"duration", duration,
	)

	return nil
}

// createJobWorker creates a new job worker instance
func (s *EventBasedScheduler) createJobWorker(jobData *schedulerTypes.EventJobData, client *ethclient.Client) (*JobWorker, error) {
	ctx, cancel := context.WithCancel(s.ctx)

	// Validate contract address
	if !common.IsHexAddress(jobData.TriggerContractAddress) {
		cancel()
		return nil, fmt.Errorf("invalid contract address: %s", jobData.TriggerContractAddress)
	}
	contractAddr := common.HexToAddress(jobData.TriggerContractAddress)

	// Calculate event signature
	eventSig := crypto.Keccak256Hash([]byte(jobData.TriggerEvent))

	// Get current block number (with caching)
	currentBlock, err := s.getCachedOrFetchBlockNumber(client, jobData.TriggerChainID)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to get current block number: %w", err)
	}

	worker := &JobWorker{
		job:          jobData,
		client:       client,
		logger:       s.logger,
		dbClient:     s.dbClient,
		cache:        s.cache,
		ctx:          ctx,
		cancel:       cancel,
		eventSig:     eventSig,
		contractAddr: contractAddr,
		lastBlock:    currentBlock,
		isRunning:    false,
		managerID:    s.managerID,
	}

	return worker, nil
}

// getCachedOrFetchBlockNumber gets block number from cache or fetches from chain
func (s *EventBasedScheduler) getCachedOrFetchBlockNumber(client *ethclient.Client, chainID string) (uint64, error) {
	cacheKey := fmt.Sprintf("block_number_%s", chainID)

	if s.cache != nil {
		if cached, err := s.cache.Get(cacheKey); err == nil {
			var blockNum uint64
			if _, err := fmt.Sscanf(cached, "%d", &blockNum); err == nil {
				return blockNum, nil
			}
		}
	}

	// Fetch from chain
	currentBlock, err := client.BlockNumber(context.Background())
	if err != nil {
		return 0, err
	}

	// Cache the result
	if s.cache != nil {
		err := s.cache.Set(cacheKey, fmt.Sprintf("%d", currentBlock), blockCacheTTL)
		if err != nil {
			s.logger.Errorf("Error caching block number: %v", err)
		}
	}

	return currentBlock, nil
}

// UnscheduleJob stops and removes a job worker
func (s *EventBasedScheduler) UnscheduleJob(jobID int64) error {
	s.workersMutex.Lock()
	defer s.workersMutex.Unlock()

	startTime := time.Now()

	worker, exists := s.workers[jobID]
	if !exists {
		// Add job unscheduling failure event to Redis stream
		if redisx.IsAvailable() {
			failureEvent := map[string]interface{}{
				"event_type": "job_unschedule_failed",
				"job_id":     jobID,
				"manager_id": s.managerID,
				"error":      "job not found",
				"failed_at":  startTime.Unix(),
			}
			err := redisx.AddJobToStream(redisx.JobsRetryEventStream, failureEvent)
			if err != nil {
				s.logger.Warnf("Failed to add job unscheduling failure event to Redis stream: %v", err)
			}
		}
		return fmt.Errorf("job %d is not scheduled", jobID)
	}

	// Capture job details before stopping
	jobDetails := map[string]interface{}{
		"trigger_chain_id":         worker.job.TriggerChainID,
		"trigger_contract_address": worker.job.TriggerContractAddress,
		"trigger_event":            worker.job.TriggerEvent,
		"target_chain_id":          worker.job.TargetChainID,
		"target_contract_address":  worker.job.TargetContractAddress,
		"target_function":          worker.job.TargetFunction,
		"last_processed_block":     worker.lastBlock,
		"was_running":              worker.IsRunning(),
	}

	// Stop worker
	worker.stop()

	// Remove from workers map
	delete(s.workers, jobID)

	// Update metrics
	metrics.JobsRunning.Dec()

	duration := time.Since(startTime)

	// Add comprehensive job unscheduling success event to Redis stream
	if redisx.IsAvailable() {
		jobContext := map[string]interface{}{
			"event_type":               "job_unscheduled",
			"job_id":                   jobID,
			"manager_id":               s.managerID,
			"trigger_chain_id":         jobDetails["trigger_chain_id"],
			"trigger_contract_address": jobDetails["trigger_contract_address"],
			"trigger_event":            jobDetails["trigger_event"],
			"target_chain_id":          jobDetails["target_chain_id"],
			"target_contract_address":  jobDetails["target_contract_address"],
			"target_function":          jobDetails["target_function"],
			"last_processed_block":     jobDetails["last_processed_block"],
			"was_running":              jobDetails["was_running"],
			"remaining_workers":        len(s.workers),
			"max_workers":              s.maxWorkers,
			"unscheduled_at":           startTime.Unix(),
			"duration_ms":              duration.Milliseconds(),
			"status":                   "unscheduled",
		}

		if err := redisx.AddJobToStream(redisx.JobsReadyEventStream, jobContext); err != nil {
			s.logger.Warnf("Failed to add job unscheduling event to Redis stream: %v", err)
		}
	}

	s.logger.Info("Job unscheduled successfully",
		"job_id", jobID,
		"trigger_chain", jobDetails["trigger_chain_id"],
		"contract", jobDetails["trigger_contract_address"],
		"event", jobDetails["trigger_event"],
		"was_running", jobDetails["was_running"],
		"remaining_workers", len(s.workers),
		"duration", duration,
	)

	return nil
}

// start begins the job worker's event monitoring loop
func (w *JobWorker) start() {
	startTime := time.Now()

	w.mutex.Lock()
	w.isRunning = true
	w.mutex.Unlock()

	// Try to acquire performer lock
	lockKey := fmt.Sprintf("event_job_%d_%s", w.job.JobID, w.job.TriggerChainID)
	lockAcquired := false

	if w.cache != nil {
		acquired, err := w.cache.AcquirePerformerLock(lockKey, performerLockTTL)
		if err != nil {
			w.logger.Warnf("Failed to acquire performer lock for job %d: %v", w.job.JobID, err)

			// Add lock failure event to Redis stream
			if redisx.IsAvailable() {
				lockFailureEvent := map[string]interface{}{
					"event_type":       "worker_lock_failed",
					"job_id":           w.job.JobID,
					"manager_id":       w.managerID,
					"lock_key":         lockKey,
					"error":            err.Error(),
					"trigger_chain_id": w.job.TriggerChainID,
					"contract_address": w.job.TriggerContractAddress,
					"trigger_event":    w.job.TriggerEvent,
					"failed_at":        startTime.Unix(),
				}
				err := redisx.AddJobToStream(redisx.JobsRetryEventStream, lockFailureEvent)
				if err != nil {
					w.logger.Warnf("Failed to add worker lock failure event to Redis stream: %v", err)
				}
			}
		} else if !acquired {
			w.logger.Warnf("Job %d is already being monitored by another worker, stopping", w.job.JobID)

			// Add lock conflict event to Redis stream
			if redisx.IsAvailable() {
				lockConflictEvent := map[string]interface{}{
					"event_type":       "worker_lock_conflict",
					"job_id":           w.job.JobID,
					"manager_id":       w.managerID,
					"lock_key":         lockKey,
					"trigger_chain_id": w.job.TriggerChainID,
					"contract_address": w.job.TriggerContractAddress,
					"trigger_event":    w.job.TriggerEvent,
					"conflict_at":      startTime.Unix(),
				}
				err := redisx.AddJobToStream(redisx.JobsRetryEventStream, lockConflictEvent)
				if err != nil {
					w.logger.Warnf("Failed to add worker lock conflict event to Redis stream: %v", err)
				}
			}
			return
		} else {
			lockAcquired = true
		}

		defer func() {
			if lockAcquired {
				if err := w.cache.ReleasePerformerLock(lockKey); err != nil {
					w.logger.Warnf("Failed to release performer lock for job %d: %v", w.job.JobID, err)
				}
			}
		}()
	}

	// Add worker start event to Redis stream
	if redisx.IsAvailable() {
		workerStartEvent := map[string]interface{}{
			"event_type":               "worker_started",
			"job_id":                   w.job.JobID,
			"manager_id":               w.managerID,
			"trigger_chain_id":         w.job.TriggerChainID,
			"trigger_contract_address": w.job.TriggerContractAddress,
			"trigger_event":            w.job.TriggerEvent,
			"target_chain_id":          w.job.TargetChainID,
			"target_contract_address":  w.job.TargetContractAddress,
			"target_function":          w.job.TargetFunction,
			"starting_block":           w.lastBlock,
			"lock_acquired":            lockAcquired,
			"cache_available":          w.cache != nil,
			"started_at":               startTime.Unix(),
			"poll_interval_seconds":    pollInterval.Seconds(),
		}

		if err := redisx.AddJobToStream(redisx.JobsReadyEventStream, workerStartEvent); err != nil {
			w.logger.Warnf("Failed to add worker start event to Redis stream: %v", err)
		}
	}

	w.logger.Info("Starting job worker",
		"job_id", w.job.JobID,
		"contract", w.job.TriggerContractAddress,
		"event", w.job.TriggerEvent,
		"starting_block", w.lastBlock,
		"lock_acquired", lockAcquired,
	)

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-w.ctx.Done():
			stopTime := time.Now()
			duration := stopTime.Sub(startTime)

			// Add worker stop event to Redis stream
			if redisx.IsAvailable() {
				workerStopEvent := map[string]interface{}{
					"event_type":       "worker_stopped",
					"job_id":           w.job.JobID,
					"manager_id":       w.managerID,
					"trigger_chain_id": w.job.TriggerChainID,
					"contract_address": w.job.TriggerContractAddress,
					"trigger_event":    w.job.TriggerEvent,
					"final_block":      w.lastBlock,
					"runtime_seconds":  duration.Seconds(),
					"stopped_at":       stopTime.Unix(),
					"graceful_stop":    true,
				}

				if err := redisx.AddJobToStream(redisx.JobsReadyEventStream, workerStopEvent); err != nil {
					w.logger.Warnf("Failed to add worker stop event to Redis stream: %v", err)
				}
			}

			w.logger.Info("Job worker stopped",
				"job_id", w.job.JobID,
				"runtime", duration,
				"final_block", w.lastBlock,
			)
			return
		case <-ticker.C:
			if err := w.checkForEvents(); err != nil {
				w.logger.Error("Error checking for events", "job_id", w.job.JobID, "error", err)
				metrics.JobsFailed.Inc()

				// Add error event to Redis stream
				if redisx.IsAvailable() {
					errorEvent := map[string]interface{}{
						"event_type":       "worker_error",
						"job_id":           w.job.JobID,
						"manager_id":       w.managerID,
						"error":            err.Error(),
						"trigger_chain_id": w.job.TriggerChainID,
						"contract_address": w.job.TriggerContractAddress,
						"trigger_event":    w.job.TriggerEvent,
						"current_block":    w.lastBlock,
						"error_at":         time.Now().Unix(),
					}
					if err := redisx.AddJobToStream(redisx.JobsRetryEventStream, errorEvent); err != nil {
						w.logger.Warnf("Failed to add worker error event to Redis stream: %v", err)
					}
				}
			}
		}
	}
}

// checkForEvents checks for new events since the last processed block
func (w *JobWorker) checkForEvents() error {
	// Get current block number
	currentBlock, err := w.client.BlockNumber(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get current block number: %w", err)
	}

	// Calculate safe block (with confirmations)
	safeBlock := currentBlock
	if currentBlock > blockConfirmations {
		safeBlock = currentBlock - blockConfirmations
	}

	// Check if there are new blocks to process
	if safeBlock <= w.lastBlock {
		return nil // No new blocks to process
	}

	// Check cache for recent events in this block range to avoid reprocessing
	blockRangeKey := fmt.Sprintf("events_%d_%d_%d", w.job.JobID, w.lastBlock+1, safeBlock)
	if w.cache != nil {
		if _, err := w.cache.Get(blockRangeKey); err == nil {
			w.logger.Debug("Block range already processed", "job_id", w.job.JobID, "from", w.lastBlock+1, "to", safeBlock)
			w.lastBlock = safeBlock
			return nil
		}
	}

	// Query logs for events
	query := ethereum.FilterQuery{
		FromBlock: new(big.Int).SetUint64(w.lastBlock + 1),
		ToBlock:   new(big.Int).SetUint64(safeBlock),
		Addresses: []common.Address{w.contractAddr},
		Topics:    [][]common.Hash{{w.eventSig}},
	}

	logs, err := w.client.FilterLogs(context.Background(), query)
	if err != nil {
		return fmt.Errorf("failed to filter logs: %w", err)
	}

	// Process each event
	for _, log := range logs {
		metrics.EventsDetected.Inc()
		if err := w.processEvent(log); err != nil {
			w.logger.Error("Failed to process event",
				"job_id", w.job.JobID,
				"tx_hash", log.TxHash.Hex(),
				"block", log.BlockNumber,
				"error", err,
			)
			metrics.JobsFailed.Inc()
		} else {
			metrics.EventsProcessed.Inc()
		}
	}

	// Cache that this block range has been processed
	if w.cache != nil {
		processedData := map[string]interface{}{
			"job_id":       w.job.JobID,
			"from_block":   w.lastBlock + 1,
			"to_block":     safeBlock,
			"events_found": len(logs),
			"processed_at": time.Now().Unix(),
		}
		if jsonData, err := json.Marshal(processedData); err == nil {
			if err := w.cache.Set(blockRangeKey, string(jsonData), eventCacheTTL); err != nil {
				w.logger.Errorf("Failed to set block range cache: %v", err)
			}
		}
	}

	// Update last processed block
	w.lastBlock = safeBlock

	w.logger.Debug("Processed blocks",
		"job_id", w.job.JobID,
		"from_block", w.lastBlock+1-uint64(len(logs)),
		"to_block", safeBlock,
		"events_found", len(logs),
	)

	return nil
}

// processEvent processes a single event and triggers the action
func (w *JobWorker) processEvent(log types.Log) error {
	startTime := time.Now()

	w.logger.Info("Event detected",
		"job_id", w.job.JobID,
		"tx_hash", log.TxHash.Hex(),
		"block", log.BlockNumber,
		"log_index", log.Index,
	)

	// Check for duplicate event processing
	eventKey := fmt.Sprintf("event_%s_%d", log.TxHash.Hex(), log.Index)
	if w.cache != nil {
		if cachedValue, err := w.cache.Get(eventKey); err == nil {
			w.logger.Debug("Event already processed, skipping",
				"tx_hash", log.TxHash.Hex(),
				"cached_at", cachedValue,
			)

			// Add duplicate event detection to Redis stream
			if redisx.IsAvailable() {
				duplicateEvent := map[string]interface{}{
					"event_type":   "event_duplicate_detected",
					"job_id":       w.job.JobID,
					"manager_id":   w.managerID,
					"chain_id":     w.job.TriggerChainID,
					"contract":     w.job.TriggerContractAddress,
					"event":        w.job.TriggerEvent,
					"tx_hash":      log.TxHash.Hex(),
					"block_number": log.BlockNumber,
					"log_index":    log.Index,
					"cached_at":    cachedValue,
					"detected_at":  startTime.Unix(),
				}
				if err := redisx.AddJobToStream(redisx.JobsReadyEventStream, duplicateEvent); err != nil {
					w.logger.Warnf("Failed to add event duplicate detection to Redis stream: %v", err)
				}
			}
			return nil
		}

		// Mark event as processed
		if err := w.cache.Set(eventKey, time.Now().Format(time.RFC3339), duplicateEventWindow); err != nil {
			w.logger.Errorf("Failed to set event cache: %v", err)
		}
	}

	// Create comprehensive event context for Redis streaming
	eventContext := map[string]interface{}{
		"event_type":               "event_detected",
		"job_id":                   w.job.JobID,
		"manager_id":               w.managerID,
		"trigger_chain_id":         w.job.TriggerChainID,
		"trigger_contract_address": w.job.TriggerContractAddress,
		"trigger_event":            w.job.TriggerEvent,
		"target_chain_id":          w.job.TargetChainID,
		"target_contract_address":  w.job.TargetContractAddress,
		"target_function":          w.job.TargetFunction,
		"tx_hash":                  log.TxHash.Hex(),
		"block_number":             log.BlockNumber,
		"log_index":                log.Index,
		"gas_used":                 log.BlockHash.Hex(), // Block hash for reference
		"cache_available":          w.cache != nil,
		"detected_at":              startTime.Unix(),
		"status":                   "processing",
	}

	// Add event detection to Redis stream
	if redisx.IsAvailable() {
		if err := redisx.AddJobToStream(redisx.JobsReadyEventStream, eventContext); err != nil {
			w.logger.Warnf("Failed to add event detection to Redis stream: %v", err)
		}
	}

	// Execute the action
	executionSuccess := w.performActionExecution(log)

	duration := time.Since(startTime)

	// Update event context with completion info
	eventContext["duration_ms"] = duration.Milliseconds()
	eventContext["completed_at"] = time.Now().Unix()

	if executionSuccess {
		eventContext["event_type"] = "event_completed"
		eventContext["status"] = "completed"

		if redisx.IsAvailable() {
			if err := redisx.AddJobToStream(redisx.JobsReadyEventStream, eventContext); err != nil {
				w.logger.Warnf("Failed to add event completion to Redis stream: %v", err)
			}
		}

		w.logger.Info("Event processed successfully",
			"job_id", w.job.JobID,
			"tx_hash", log.TxHash.Hex(),
			"block", log.BlockNumber,
			"target_chain", w.job.TargetChainID,
			"target_function", w.job.TargetFunction,
			"duration", duration,
		)
	} else {
		eventContext["event_type"] = "event_failed"
		eventContext["status"] = "failed"
		eventContext["error"] = "action execution failed"

		if redisx.IsAvailable() {
			if err := redisx.AddJobToStream(redisx.JobsRetryEventStream, eventContext); err != nil {
				w.logger.Warnf("Failed to add event failure to Redis stream: %v", err)
			}
		}

		w.logger.Error("Event processing failed",
			"job_id", w.job.JobID,
			"tx_hash", log.TxHash.Hex(),
			"block", log.BlockNumber,
			"target_chain", w.job.TargetChainID,
			"target_function", w.job.TargetFunction,
			"duration", duration,
		)
	}

	return nil
}

// performActionExecution handles the actual action execution logic
func (w *JobWorker) performActionExecution(log types.Log) bool {
	// TODO: Implement actual action execution logic
	// This should:
	// 1. Parse event data if needed
	// 2. Send task to manager/keeper for execution
	// 3. Handle response and update job status

	// Simulate action execution for now
	w.logger.Info("Simulating action execution",
		"job_id", w.job.JobID,
		"target_chain", w.job.TargetChainID,
		"target_contract", w.job.TargetContractAddress,
		"target_function", w.job.TargetFunction,
	)

	// Simulate processing time
	time.Sleep(100 * time.Millisecond)

	// Simulate 95% success rate
	return time.Now().UnixNano()%100 < 95
}

// stop gracefully stops the job worker
func (w *JobWorker) stop() {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if w.isRunning {
		w.cancel()
		w.isRunning = false
		w.logger.Info("Job worker stopped", "job_id", w.job.JobID)
	}
}

// IsRunning returns whether the worker is currently running
func (w *JobWorker) IsRunning() bool {
	w.mutex.RLock()
	defer w.mutex.RUnlock()
	return w.isRunning
}

// Start begins the scheduler's main loop (for compatibility)
func (s *EventBasedScheduler) Start(ctx context.Context) {
	s.logger.Info("Event-based scheduler ready for job scheduling", "manager_id", s.managerID)

	// Keep the service alive
	<-ctx.Done()
	s.logger.Info("Scheduler context cancelled, stopping all workers")
	s.Stop()
}

// Stop gracefully stops all job workers
func (s *EventBasedScheduler) Stop() {
	startTime := time.Now()

	s.logger.Info("Stopping event-based scheduler")

	// Capture statistics before shutdown
	s.workersMutex.RLock()
	totalWorkers := len(s.workers)
	runningWorkers := 0
	workerDetails := make([]map[string]interface{}, 0, totalWorkers)

	for jobID, worker := range s.workers {
		isRunning := worker.IsRunning()
		if isRunning {
			runningWorkers++
		}

		workerDetails = append(workerDetails, map[string]interface{}{
			"job_id":           jobID,
			"is_running":       isRunning,
			"trigger_chain_id": worker.job.TriggerChainID,
			"contract_address": worker.job.TriggerContractAddress,
			"trigger_event":    worker.job.TriggerEvent,
			"last_block":       worker.lastBlock,
		})
	}
	s.workersMutex.RUnlock()

	s.clientsMutex.RLock()
	connectedChains := len(s.chainClients)
	s.clientsMutex.RUnlock()

	s.cancel()

	// Stop all workers
	s.workersMutex.Lock()
	for jobID, worker := range s.workers {
		worker.stop()
		s.logger.Info("Stopped worker", "job_id", jobID)
	}
	s.workers = make(map[int64]*JobWorker)
	s.workersMutex.Unlock()

	// Close chain clients
	s.clientsMutex.Lock()
	for chainID, client := range s.chainClients {
		client.Close()
		s.logger.Info("Closed chain client", "chain_id", chainID)
	}
	s.chainClients = make(map[string]*ethclient.Client)
	s.clientsMutex.Unlock()

	duration := time.Since(startTime)

	// Add comprehensive scheduler shutdown event to Redis stream
	if redisx.IsAvailable() {
		shutdownEvent := map[string]interface{}{
			"event_type":        "scheduler_shutdown",
			"manager_id":        s.managerID,
			"total_workers":     totalWorkers,
			"running_workers":   runningWorkers,
			"connected_chains":  connectedChains,
			"cache_available":   s.cache != nil,
			"worker_details":    workerDetails,
			"shutdown_at":       startTime.Unix(),
			"duration_ms":       duration.Milliseconds(),
			"graceful_shutdown": true,
		}

		if err := redisx.AddJobToStream(redisx.JobsReadyEventStream, shutdownEvent); err != nil {
			s.logger.Warnf("Failed to add scheduler shutdown event to Redis stream: %v", err)
		} else {
			s.logger.Info("Scheduler shutdown event added to Redis stream")
		}
	}

	s.logger.Info("Event-based scheduler stopped",
		"total_workers_stopped", totalWorkers,
		"running_workers_stopped", runningWorkers,
		"chains_disconnected", connectedChains,
		"duration", duration,
	)
}

// GetStats returns current scheduler statistics
func (s *EventBasedScheduler) GetStats() map[string]interface{} {
	s.workersMutex.RLock()
	defer s.workersMutex.RUnlock()

	s.clientsMutex.RLock()
	defer s.clientsMutex.RUnlock()

	runningWorkers := 0
	for _, worker := range s.workers {
		if worker.IsRunning() {
			runningWorkers++
		}
	}

	return map[string]interface{}{
		"manager_id":       s.managerID,
		"total_workers":    len(s.workers),
		"running_workers":  runningWorkers,
		"max_workers":      s.maxWorkers,
		"connected_chains": len(s.chainClients),
		"supported_chains": []string{"11155420", "84532", "11155111"}, // OP Sepolia, Base Sepolia, Ethereum Sepolia
		"cache_available":  s.cache != nil,
	}
}

// GetJobWorkerStats returns statistics for a specific job worker
func (s *EventBasedScheduler) GetJobWorkerStats(jobID int64) (map[string]interface{}, error) {
	s.workersMutex.RLock()
	defer s.workersMutex.RUnlock()

	worker, exists := s.workers[jobID]
	if !exists {
		return nil, fmt.Errorf("job %d not found", jobID)
	}

	return map[string]interface{}{
		"job_id":           worker.job.JobID,
		"is_running":       worker.IsRunning(),
		"trigger_chain_id": worker.job.TriggerChainID,
		"contract_address": worker.job.TriggerContractAddress,
		"trigger_event":    worker.job.TriggerEvent,
		"last_block":       worker.lastBlock,
		"manager_id":       worker.managerID,
	}, nil
}
