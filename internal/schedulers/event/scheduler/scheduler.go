package scheduler

import (
	"context"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/trigg3rX/triggerx-backend/internal/schedulers/event/client"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/event/config"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/event/metrics"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	schedulerTypes "github.com/trigg3rX/triggerx-backend/pkg/types"
)

const (
	blockConfirmations = 3                // Wait for 3 block confirmations
	pollInterval       = 10 * time.Second // Poll every 10 seconds for new blocks
	workerTimeout      = 30 * time.Second // Timeout for worker operations
	maxRetries         = 3                // Max retries for failed operations
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
	metrics      *metrics.Collector
	managerID    string
}

// JobWorker represents an individual worker watching a specific job
type JobWorker struct {
	job          *schedulerTypes.EventJobData
	client       *ethclient.Client
	logger       logging.Logger
	dbClient     *client.DBServerClient
	ctx          context.Context
	cancel       context.CancelFunc
	eventSig     common.Hash
	contractAddr common.Address
	lastBlock    uint64
	isRunning    bool
	mutex        sync.RWMutex
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

	scheduler := &EventBasedScheduler{
		ctx:          ctx,
		cancel:       cancel,
		logger:       logger,
		workers:      make(map[int64]*JobWorker),
		chainClients: make(map[string]*ethclient.Client),
		dbClient:     dbClient,
		metrics:      metrics.NewCollector(),
		managerID:    managerID,
	}

	// Initialize chain clients
	if err := scheduler.initChainClients(); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to initialize chain clients: %w", err)
	}

	// Start metrics collection
	scheduler.metrics.Start()

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

	// Check if job is already scheduled
	if _, exists := s.workers[jobData.JobID]; exists {
		return fmt.Errorf("job %d is already scheduled", jobData.JobID)
	}

	// Get chain client
	s.clientsMutex.RLock()
	client, exists := s.chainClients[jobData.TriggerChainID]
	s.clientsMutex.RUnlock()

	if !exists {
		return fmt.Errorf("unsupported chain ID: %s", jobData.TriggerChainID)
	}

	// Create job worker
	worker, err := s.createJobWorker(jobData, client)
	if err != nil {
		return fmt.Errorf("failed to create job worker: %w", err)
	}

	// Store worker
	s.workers[jobData.JobID] = worker

	// Start worker
	go worker.start()

	// Update metrics
	metrics.JobsScheduled.Inc()
	metrics.JobsRunning.Inc()

	s.logger.Info("Job scheduled successfully",
		"job_id", jobData.JobID,
		"trigger_chain", jobData.TriggerChainID,
		"contract", jobData.TriggerContractAddress,
		"event", jobData.TriggerEvent,
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

	// Get current block number
	currentBlock, err := client.BlockNumber(context.Background())
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to get current block number: %w", err)
	}

	worker := &JobWorker{
		job:          jobData,
		client:       client,
		logger:       s.logger,
		dbClient:     s.dbClient,
		ctx:          ctx,
		cancel:       cancel,
		eventSig:     eventSig,
		contractAddr: contractAddr,
		lastBlock:    currentBlock,
		isRunning:    false,
	}

	return worker, nil
}

// UnscheduleJob stops and removes a job worker
func (s *EventBasedScheduler) UnscheduleJob(jobID int64) error {
	s.workersMutex.Lock()
	defer s.workersMutex.Unlock()

	worker, exists := s.workers[jobID]
	if !exists {
		return fmt.Errorf("job %d is not scheduled", jobID)
	}

	// Stop worker
	worker.stop()

	// Remove from workers map
	delete(s.workers, jobID)

	// Update metrics
	metrics.JobsRunning.Dec()

	s.logger.Info("Job unscheduled successfully", "job_id", jobID)
	return nil
}

// start begins the job worker's event monitoring loop
func (w *JobWorker) start() {
	w.mutex.Lock()
	w.isRunning = true
	w.mutex.Unlock()

	w.logger.Info("Starting job worker",
		"job_id", w.job.JobID,
		"contract", w.job.TriggerContractAddress,
		"event", w.job.TriggerEvent,
	)

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-w.ctx.Done():
			w.logger.Info("Job worker stopped", "job_id", w.job.JobID)
			return
		case <-ticker.C:
			if err := w.checkForEvents(); err != nil {
				w.logger.Error("Error checking for events", "job_id", w.job.JobID, "error", err)
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
		} else {
			metrics.EventsProcessed.Inc()
		}
	}

	// Update last processed block
	w.lastBlock = safeBlock

	w.logger.Debug("Processed blocks",
		"job_id", w.job.JobID,
		"from_block", w.lastBlock+1,
		"to_block", safeBlock,
		"events_found", len(logs),
	)

	return nil
}

// processEvent processes a single event and triggers the action
func (w *JobWorker) processEvent(log types.Log) error {
	w.logger.Info("Event detected",
		"job_id", w.job.JobID,
		"tx_hash", log.TxHash.Hex(),
		"block", log.BlockNumber,
		"log_index", log.Index,
	)

	// TODO: Implement action execution logic
	// This should:
	// 1. Parse event data if needed
	// 2. Send task to manager/keeper for execution
	// 3. Handle response and update job status

	// For now, simulate action execution
	return w.simulateActionExecution(log)
}

// simulateActionExecution simulates the action execution (replace with actual implementation)
func (w *JobWorker) simulateActionExecution(log types.Log) error {
	// TODO: Replace with actual action execution logic
	w.logger.Info("Simulating action execution",
		"job_id", w.job.JobID,
		"target_chain", w.job.TargetChainID,
		"target_contract", w.job.TargetContractAddress,
		"target_function", w.job.TargetFunction,
	)

	// Simulate processing time
	time.Sleep(100 * time.Millisecond)

	return nil
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
	s.logger.Info("Stopping event-based scheduler")

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

	s.logger.Info("Event-based scheduler stopped")
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
		"connected_chains": len(s.chainClients),
		"supported_chains": []string{"11155420", "84532", "11155111"}, // OP Sepolia, Base Sepolia, Ethereum Sepolia
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
	}, nil
}
