package workers

// import (
// 	"context"
// 	"fmt"
// 	"os"
// 	"time"

// 	"github.com/ethereum/go-ethereum"
// 	"github.com/ethereum/go-ethereum/common"
// 	gethtypes "github.com/ethereum/go-ethereum/core/types"
// 	"github.com/ethereum/go-ethereum/crypto"
// 	"github.com/ethereum/go-ethereum/ethclient"

// 	"github.com/trigg3rX/triggerx-backend/pkg/types"
// )

// type EventBasedWorker struct {
// 	jobID        int64
// 	scheduler    *JobScheduler
// 	chainID      int
// 	jobData      *types.Job
// 	client       *ethclient.Client
// 	subscription ethereum.Subscription
// 	BaseWorker
// }

// func NewEventBasedWorker(jobData *types.Job, scheduler *JobScheduler) *EventBasedWorker {
// 	return &EventBasedWorker{
// 		jobID:     jobData.JobID,
// 		scheduler: scheduler,
// 		chainID:   jobData.TriggerChainID,
// 		jobData:   jobData,
// 		BaseWorker: BaseWorker{
// 			status:     "pending",
// 			maxRetries: 3,
// 		},
// 	}
// }

// func (w *EventBasedWorker) Start(ctx context.Context) {
// 	wsURL := w.getAlchemyWSURL()

// 	w.scheduler.logger.Infof("Connecting to Alchemy at %s", wsURL)

// 	client, err := ethclient.Dial(wsURL)
// 	if err != nil {
// 		w.scheduler.logger.Errorf("Failed to connect to Alchemy for job %d: %v", w.jobID, err)
// 		w.status = "failed"
// 		w.Stop()
// 		return
// 	}
// 	w.client = client

// 	eventSigHash := crypto.Keccak256Hash([]byte(w.jobData.TriggerEvent))
// 	eventSignature := eventSigHash.Hex()

// 	query := ethereum.FilterQuery{
// 		Addresses: []common.Address{common.HexToAddress(w.jobData.TriggerContractAddress)},
// 		Topics:    [][]common.Hash{{common.HexToHash(eventSignature)}},
// 	}

// 	logs := make(chan gethtypes.Log)

// 	sub, err := client.SubscribeFilterLogs(ctx, query, logs)
// 	if err != nil {
// 		w.scheduler.logger.Errorf("Failed to subscribe to events for job %d: %v", w.jobID, err)
// 		w.status = "failed"
// 		w.Stop()
// 		return
// 	}
// 	w.subscription = sub

// 	var triggerData types.TriggerData
// 	triggerData.TimeInterval = w.jobData.TimeInterval
// 	triggerData.LastExecuted = time.Now()
// 	triggerData.ConditionParams = make(map[string]interface{})

// 	go func() {
// 		for {
// 			select {
// 			case <-ctx.Done():
// 				return

// 			case err := <-sub.Err():
// 				w.scheduler.logger.Errorf("Event subscription error for job %d: %v", w.jobID, err)
// 				w.status = "failed"
// 				return

// 			case log := <-logs:
// 				w.scheduler.logger.Infof("Event detected for job %d: %v", w.jobID, log.TxHash.Hex())

// 				triggerData.Timestamp = time.Now()
// 				triggerData.TriggerTxHash = log.TxHash.Hex()

// 				if err := w.executeTask(w.jobData, &triggerData); err != nil {
// 					w.handleError(err)
// 				}
// 				return
// 			}
// 		}
// 	}()
// }

// func (w *EventBasedWorker) Stop() {
// 	if w.subscription != nil {
// 		w.subscription.Unsubscribe()
// 	}
// 	if w.client != nil {
// 		w.client.Close()
// 	}
// 	w.scheduler.RemoveJob(w.jobID)
// }

// func (w *EventBasedWorker) getAlchemyWSURL() string {
// 	apiKey := os.Getenv("ALCHEMY_API_KEY")

// 	var network string
// 	switch w.chainID {
// 	case 17000:
// 		network = "eth-holesky"
// 	case 11155111:
// 		network = "eth-sepolia"
// 	case 84532:
// 		network = "base-sepolia"
// 	case 421614:
// 		network = "arb-sepolia"
// 	case 11155420:
// 		network = "opt-sepolia"
// 	default:
// 		network = "eth-holesky"
// 	}

// 	return fmt.Sprintf("wss://%s.g.alchemy.com/v2/%s", network, apiKey)
// }

// func (w *EventBasedWorker) executeTask(jobData *types.Job, triggerData *types.TriggerData) error {
// 	w.scheduler.logger.Infof("Executing event-based job: %d", w.jobID)

// 	taskData := &types.CreateTaskData{
// 		JobID:            w.jobID,
// 		TaskDefinitionID: jobData.TaskDefinitionID,
// 		TaskPerformerID:  0,
// 	}

// 	taskID, status, err := CreateTaskData(taskData)
// 	if err != nil {
// 		w.scheduler.logger.Errorf("Failed to create task data for job %d: %v", w.jobID, err)
// 		return err
// 	}

// 	triggerData.TaskID = taskID

// 	if !status {
// 		return fmt.Errorf("failed to create task data for job %d", w.jobID)
// 	}

// 	w.scheduler.logger.Infof("Task sent for job %d to performer", w.jobID)

// 	if err := w.handleLinkedJob(w.scheduler, jobData); err != nil {
// 		w.scheduler.logger.Errorf("Failed to execute linked job for job %d: %v", w.jobID, err)
// 	}

// 	return nil
// }

// func (w *EventBasedWorker) GetJobID() int64 {
// 	return w.jobID
// }

// func (w *EventBasedWorker) GetStatus() string {
// 	return w.status
// }

// func (w *EventBasedWorker) GetError() string {
// 	return w.error
// }

// func (w *EventBasedWorker) GetRetries() int {
// 	return w.currentRetry
// }
