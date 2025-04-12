package workers

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	gethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/trigg3rX/triggerx-backend/internal/manager/scheduler/services"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

type EventBasedWorker struct {
	jobID        int64
	scheduler    JobScheduler
	chainID      string
	jobData      *types.HandleCreateJobData
	client       *ethclient.Client
	subscription ethereum.Subscription
	logsChan     chan gethtypes.Log
	stopChan     chan struct{}
	wg           sync.WaitGroup
	BaseWorker
}

func NewEventBasedWorker(jobData types.HandleCreateJobData, scheduler JobScheduler) *EventBasedWorker {
	return &EventBasedWorker{
		jobID:     jobData.JobID,
		scheduler: scheduler,
		chainID:   jobData.TriggerChainID,
		jobData:   &jobData,
		stopChan:  make(chan struct{}),
		logsChan:  make(chan gethtypes.Log, 100),
		BaseWorker: BaseWorker{
			status:     "pending",
			maxRetries: 3,
		},
	}
}

func (w *EventBasedWorker) Start(ctx context.Context) {
	w.wg.Add(1)
	defer w.wg.Done()

	wsURL := w.getAlchemyWSURL()
	w.scheduler.Logger().Infof("Starting event worker for job %d on chain %s", w.jobID, w.chainID)
	w.scheduler.Logger().Debugf("Connecting to Alchemy at %s", wsURL)

	// Initial connection
	if err := w.connect(ctx); err != nil {
		w.scheduler.Logger().Errorf("Initial connection failed for job %d: %v", w.jobID, err)
		w.status = "failed"
		w.Stop()
		return
	}

	// Calculate end time if timeframe is specified
	var endTime time.Time
	if w.jobData.TimeFrame > 0 {
		endTime = time.Now().UTC().Add(time.Duration(w.jobData.TimeFrame) * time.Second)
		w.scheduler.Logger().Debugf("Job %d has timeframe until %v", w.jobID, endTime)
	}

	// Prepare trigger data
	var triggerData types.TriggerData
	triggerData.TimeInterval = w.jobData.TimeInterval
	triggerData.LastExecuted = time.Now().UTC()
	triggerData.ConditionParams = make(map[string]interface{})

	// Main event loop
	// Main event loop
	for {
		select {
		case <-ctx.Done():
			w.scheduler.Logger().Infof("Context canceled for job %d", w.jobID)
			return

		case <-w.stopChan:
			w.scheduler.Logger().Infof("Stop signal received for job %d", w.jobID)
			return

		case err := <-w.subscription.Err():
			w.scheduler.Logger().Errorf("Subscription error for job %d: %v", w.jobID, err)
			w.handleError(err)

			if w.status == "failed" {
				return
			}

			// Try to reconnect
			if err := w.connect(ctx); err != nil {
				w.scheduler.Logger().Errorf("Reconnection failed for job %d: %v", w.jobID, err)
				w.handleError(err)
				if w.status == "failed" {
					return
				}
			}

		case log := <-w.logsChan:
			// Check timeframe
			if w.jobData.TimeFrame > 0 && time.Now().UTC().After(endTime) {
				w.scheduler.Logger().Infof("Timeframe reached for job %d, stopping worker", w.jobID)
				return
			}

			w.scheduler.Logger().Infof("Event detected for job %d: %v", w.jobID, log.TxHash.Hex())

			// Handle the event exactly once
			w.handleEvent(log, &triggerData)

			// For non-recurring jobs, exit immediately after handling the event
			if !w.jobData.Recurring {
				w.scheduler.Logger().Infof("Non-recurring job %d completed, stopping worker", w.jobID)
				return
			}

			// For recurring jobs, continue listening for new events
		}
	}
}

func (w *EventBasedWorker) connect(ctx context.Context) error {
	// Clean up any existing connection
	w.cleanup()

	wsURL := w.getAlchemyWSURL()
	w.scheduler.Logger().Debugf("Connecting to Alchemy for job %d at %s", w.jobID, wsURL)

	// Create new client with timeout
	connectCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	client, err := ethclient.DialContext(connectCtx, wsURL)
	if err != nil {
		return fmt.Errorf("failed to dial client: %w", err)
	}

	// Verify connection
	if _, err := client.BlockNumber(connectCtx); err != nil {
		client.Close()
		return fmt.Errorf("connection test failed: %w", err)
	}

	// Prepare event subscription
	eventSigHash := crypto.Keccak256Hash([]byte(w.jobData.TriggerEvent))
	query := ethereum.FilterQuery{
		Addresses: []common.Address{common.HexToAddress(w.jobData.TriggerContractAddress)},
		Topics:    [][]common.Hash{{eventSigHash}},
	}

	sub, err := client.SubscribeFilterLogs(ctx, query, w.logsChan)
	if err != nil {
		client.Close()
		return fmt.Errorf("failed to subscribe to logs: %w", err)
	}

	w.client = client
	w.subscription = sub

	w.scheduler.Logger().Infof("Successfully subscribed to events for job %d", w.jobID)
	return nil
}

func (w *EventBasedWorker) handleEvent(log gethtypes.Log, triggerData *types.TriggerData) {
	triggerData.Timestamp = time.Now().UTC()
	triggerData.TriggerTxHash = log.TxHash.Hex()

	if err := w.executeTask(w.jobData, triggerData); err != nil {
		w.scheduler.Logger().Errorf("Task execution failed for job %d: %v", w.jobID, err)
		w.handleError(err)
		return
	}

	// Update last execution time
	triggerData.LastExecuted = time.Now().UTC()
	w.scheduler.Logger().Infof("Job %d executed successfully", w.jobID)
}

func (w *EventBasedWorker) Stop() {
	close(w.stopChan)
	w.cleanup()
	w.wg.Wait()
	w.scheduler.RemoveJob(w.jobID)
}

func (w *EventBasedWorker) cleanup() {
	if w.subscription != nil {
		w.subscription.Unsubscribe()
		w.subscription = nil
	}

	if w.client != nil {
		w.client.Close()
		w.client = nil
	}
}

func (w *EventBasedWorker) getAlchemyWSURL() string {
	apiKey := os.Getenv("ALCHEMY_API_KEY")
	if apiKey == "" {
		w.scheduler.Logger().Error("ALCHEMY_API_KEY environment variable not set")
		return ""
	}

	networkMap := map[string]string{
		"17000":    "eth-holesky",
		"11155111": "eth-sepolia",
		"84532":    "base-sepolia",
		"421614":   "arb-sepolia",
		"11155420": "opt-sepolia",
	}

	network, ok := networkMap[w.chainID]
	if !ok {
		w.scheduler.Logger().Warnf("Unknown chain ID %s, defaulting to eth-holesky", w.chainID)
		network = "eth-holesky"
	}

	return fmt.Sprintf("wss://%s.g.alchemy.com/v2/%s", network, apiKey)
}

func (w *EventBasedWorker) executeTask(jobData *types.HandleCreateJobData, triggerData *types.TriggerData) error {
	w.scheduler.Logger().Infof("Executing task for job %d", w.jobID)

	taskData := &types.CreateTaskData{
		JobID:            w.jobID,
		TaskDefinitionID: jobData.TaskDefinitionID,
		TaskPerformerID:  0,
	}

	performerData, err := services.GetPerformer()
	if err != nil {
		return fmt.Errorf("failed to get performer: %w", err)
	}

	taskData.TaskPerformerID = performerData.KeeperID

	taskID, status, err := services.CreateTaskData(taskData)
	if err != nil || !status {
		return fmt.Errorf("failed to create task data: %w", err)
	}

	triggerData.TaskID = taskID

	if status, err = services.SendTaskToPerformer(jobData, triggerData, performerData); err != nil || !status {
		return fmt.Errorf("failed to send task to performer: %w", err)
	}

	if err := w.handleLinkedJob(w.scheduler, jobData); err != nil {
		w.scheduler.Logger().Errorf("Linked job execution failed: %v", err)
	}

	return nil
}

func (w *EventBasedWorker) GetJobID() int64 {
	return w.jobID
}

func (w *EventBasedWorker) GetStatus() string {
	return w.status
}

func (w *EventBasedWorker) GetError() string {
	return w.error
}

func (w *EventBasedWorker) GetRetries() int {
	return w.currentRetry
}

func (w *EventBasedWorker) handleError(err error) {
	w.error = err.Error()
	w.currentRetry++

	if w.currentRetry >= w.maxRetries {
		w.status = "failed"
		w.scheduler.Logger().Errorf("Job %d reached max retries (%d)", w.jobID, w.maxRetries)
	} else {
		w.scheduler.Logger().Warnf("Job %d error (retry %d/%d): %v",
			w.jobID, w.currentRetry, w.maxRetries, err)
	}
}

func (w *EventBasedWorker) UpdateLastExecutedTime(timestamp time.Time) {
	if w.jobData != nil {
		w.jobData.LastExecutedAt = timestamp
		w.scheduler.Logger().Debugf("Updated LastExecutedAt for job %d to %v", w.jobID, timestamp)
	}
}
