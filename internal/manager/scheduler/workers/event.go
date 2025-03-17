package workers

import (
	"context"
	"fmt"
	"os"
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
	BaseWorker
}

func NewEventBasedWorker(jobData types.HandleCreateJobData, scheduler JobScheduler) *EventBasedWorker {
	return &EventBasedWorker{
		jobID:     jobData.JobID,
		scheduler: scheduler,
		chainID:   jobData.TriggerChainID,
		jobData:   &jobData,
		BaseWorker: BaseWorker{
			status:     "pending",
			maxRetries: 3,
		},
	}
}

func (w *EventBasedWorker) Start(ctx context.Context) {
	wsURL := w.getAlchemyWSURL()

	w.scheduler.Logger().Infof("Connecting to Alchemy at %s", wsURL)

	client, err := ethclient.Dial(wsURL)
	if err != nil {
		w.scheduler.Logger().Errorf("Failed to connect to Alchemy for job %d: %v", w.jobID, err)
		w.status = "failed"
		w.Stop()
		return
	}
	w.client = client

	eventSigHash := crypto.Keccak256Hash([]byte(w.jobData.TriggerEvent))
	eventSignature := eventSigHash.Hex()

	query := ethereum.FilterQuery{
		Addresses: []common.Address{common.HexToAddress(w.jobData.TriggerContractAddress)},
		Topics:    [][]common.Hash{{common.HexToHash(eventSignature)}},
	}

	logs := make(chan gethtypes.Log)

	sub, err := client.SubscribeFilterLogs(ctx, query, logs)
	if err != nil {
		w.scheduler.Logger().Errorf("Failed to subscribe to events for job %d: %v", w.jobID, err)
		w.status = "failed"
		w.Stop()
		return
	}
	w.subscription = sub

	var triggerData types.TriggerData
	triggerData.TimeInterval = w.jobData.TimeInterval
	triggerData.LastExecuted = time.Now()
	triggerData.ConditionParams = make(map[string]interface{})

	// Calculate end time if timeframe is specified
	var endTime time.Time
	if w.jobData.TimeFrame > 0 {
		endTime = time.Now().Add(time.Duration(w.jobData.TimeFrame) * time.Second)
	}

	go func() {
		// Use a flag to prevent multiple Stop() calls
		var stopped bool
		defer func() {
			if !stopped {
				stopped = true
				w.Stop()
			}
		}()

		for {
			select {
			case <-ctx.Done():
				return

			case err := <-sub.Err():
				w.scheduler.Logger().Errorf("Event subscription error for job %d: %v", w.jobID, err)
				w.handleError(err)
				if w.status == "failed" {
					return
				}

				// Try to reconnect if we're in recurring mode and haven't failed
				if w.jobData.Recurring {
					w.reconnectSubscription(ctx, query, logs)
					if w.status == "failed" {
						return
					}
				} else {
					return
				}

			case log := <-logs:
				// Check if we've exceeded the timeframe
				if w.jobData.TimeFrame > 0 && time.Now().After(endTime) {
					w.scheduler.Logger().Infof("Timeframe reached for job %d, stopping worker", w.jobID)
					return
				}

				w.scheduler.Logger().Infof("Event detected for job %d: %v", w.jobID, log.TxHash.Hex())

				triggerData.Timestamp = time.Now()
				triggerData.TriggerTxHash = log.TxHash.Hex()

				if err := w.executeTask(w.jobData, &triggerData); err != nil {
					w.handleError(err)
					if w.status == "failed" || !w.jobData.Recurring {
						return
					}
					continue
				}

				// If it's not recurring, exit after first execution
				if !w.jobData.Recurring {
					return
				}

				// Update last execution time
				triggerData.LastExecuted = time.Now()
				w.scheduler.Logger().Infof("Job %d executed. Continuing to listen for events due to recurring flag", w.jobID)
			}
		}
	}()
}

// Add a helper method to reconnect subscription after errors
func (w *EventBasedWorker) reconnectSubscription(ctx context.Context, query ethereum.FilterQuery, logs chan gethtypes.Log) {
	w.scheduler.Logger().Infof("Attempting to reconnect subscription for job %d", w.jobID)

	// Unsubscribe from existing subscription
	if w.subscription != nil {
		w.subscription.Unsubscribe()
		w.subscription = nil
	}

	// Close existing client
	if w.client != nil {
		w.client.Close()
		w.client = nil
	}

	// Create new client
	wsURL := w.getAlchemyWSURL()
	client, err := ethclient.Dial(wsURL)
	if err != nil {
		w.scheduler.Logger().Errorf("Failed to reconnect to Alchemy for job %d: %v", w.jobID, err)
		w.status = "failed"
		return
	}
	w.client = client

	// Create new subscription
	sub, err := client.SubscribeFilterLogs(ctx, query, logs)
	if err != nil {
		w.scheduler.Logger().Errorf("Failed to resubscribe to events for job %d: %v", w.jobID, err)
		w.status = "failed"
		return
	}
	w.subscription = sub

	w.scheduler.Logger().Infof("Successfully reconnected subscription for job %d", w.jobID)
}

func (w *EventBasedWorker) Stop() {
	if w.subscription != nil {
		w.subscription.Unsubscribe()
		w.subscription = nil
	}

	if w.client != nil {
		w.client.Close()
		w.client = nil
	}

	w.scheduler.RemoveJob(w.jobID)
}

func (w *EventBasedWorker) getAlchemyWSURL() string {
	apiKey := os.Getenv("ALCHEMY_API_KEY")

	var network string
	switch w.chainID {
	case "17000":
		network = "eth-holesky"
	case "11155111":
		network = "eth-sepolia"
	case "84532":
		network = "base-sepolia"
	case "421614":
		network = "arb-sepolia"
	case "11155420":
		network = "opt-sepolia"
	default:
		network = "eth-holesky"
	}

	return fmt.Sprintf("wss://%s.g.alchemy.com/v2/%s", network, apiKey)
}

func (w *EventBasedWorker) executeTask(jobData *types.HandleCreateJobData, triggerData *types.TriggerData) error {
	w.scheduler.Logger().Infof("Executing event-based job: %d", w.jobID)

	taskData := &types.CreateTaskData{
		JobID:            w.jobID,
		TaskDefinitionID: jobData.TaskDefinitionID,
		TaskPerformerID:  0,
	}

	performerData, err := services.GetPerformerData()
	if err != nil {
		w.scheduler.Logger().Errorf("Failed to get performer data for job %d: %v", w.jobID, err)
		return err
	}

	taskData.TaskPerformerID = performerData.KeeperID

	taskID, status, err := services.CreateTaskData(taskData)
	if err != nil {
		w.scheduler.Logger().Errorf("Failed to create task data for job %d: %v", w.jobID, err)
		return err
	}

	triggerData.TaskID = taskID

	if !status {
		return fmt.Errorf("failed to create task data for job %d", w.jobID)
	}

	status, err = services.SendTaskToPerformer(jobData, triggerData, performerData.ConnectionAddress)
	if err != nil {
		w.scheduler.Logger().Errorf("Error sending task to performer: %v", err)
		return err
	}

	w.scheduler.Logger().Infof("Task sent for job %d to performer", w.jobID)

	if err := w.handleLinkedJob(w.scheduler, jobData); err != nil {
		w.scheduler.Logger().Errorf("Failed to execute linked job for job %d: %v", w.jobID, err)
	}

	if !status {
		return fmt.Errorf("failed to send task to performer for job %d", w.jobID)
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
	}
}
