package manager

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	gethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/robfig/cron/v3"

	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

type Worker interface {
	Start(ctx context.Context)
	Stop()
	GetJobID() int64
	GetStatus() string
	GetError() string
	GetRetries() int
}

type BaseWorker struct {
	status       string
	error        string
	currentRetry int
	maxRetries   int
}

type TimeBasedWorker struct {
	jobID     int64
	scheduler *JobScheduler
	cron      *cron.Cron
	schedule  string
	jobData   *types.Job
	startTime time.Time
	BaseWorker
}

type EventBasedWorker struct {
	jobID           int64
	scheduler       *JobScheduler
	chainID         int
	jobData         *types.Job
	client          *ethclient.Client
	subscription    ethereum.Subscription
	BaseWorker
}

type ConditionBasedWorker struct {
	jobID         int64
	scheduler     *JobScheduler
	jobData       *types.Job
	ticker        *time.Ticker
	done          chan bool
	BaseWorker
}

func NewTimeBasedWorker(jobData *types.Job, schedule string, scheduler *JobScheduler) *TimeBasedWorker {
	return &TimeBasedWorker{
		jobID:     jobData.JobID,
		scheduler: scheduler,
		cron:      cron.New(cron.WithSeconds()),
		schedule:  schedule,
		jobData:   jobData,
		startTime: time.Now(),
		BaseWorker: BaseWorker{
			status:     "pending",
			maxRetries: 3,
		},
	}
}

func (w *TimeBasedWorker) Start(ctx context.Context) {
	if w.status == "completed" || w.status == "failed" {
		return
	}

	w.status = "running"

	time.AfterFunc(2*time.Second, func() {
		if err := w.executeTask(w.jobData); err != nil {
			w.handleError(err)
		}
	})

	w.cron.AddFunc(w.schedule, func() {
		if w.jobData.TimeFrame > 0 && time.Since(w.startTime) > time.Duration(w.jobData.TimeFrame)*time.Second {
			w.status = "completed"
			w.Stop()
			return
		}

		if w.status != "running" {
			return
		}

		if err := w.executeTask(w.jobData); err != nil {
			w.handleError(err)
		}
	})
	w.cron.Start()

	go func() {
		<-ctx.Done()
		w.Stop()
	}()
}

func (w *TimeBasedWorker) handleError(err error) {
	w.error = err.Error()
	w.currentRetry++

	if w.currentRetry >= w.maxRetries {
		w.status = "failed"
		w.Stop()
	}
}

func (w *TimeBasedWorker) Stop() {
	w.cron.Stop()
	w.scheduler.RemoveJob(w.jobID)
}

func (w *TimeBasedWorker) GetJobID() int64 {
	return w.jobID
}

func (w *TimeBasedWorker) executeTask(jobData *types.Job) error {
	w.scheduler.logger.Infof("Executing time-based job: %d", w.jobID)

	performer, err := GetPerformer()
	if err != nil {
		w.scheduler.logger.Errorf("Failed to get performer for job %d: %v", w.jobID, err)
		return err
	}

	taskData := &types.TaskData{
		JobID:           w.jobID,
		TaskDefinitionID: 1,
		TaskPerformer:   performer.KeeperAddress,
		IsApproved:      true,
	}

	w.scheduler.logger.Infof("Task data: %d | %d | %s", taskData.JobID, taskData.TaskDefinitionID, taskData.TaskPerformer)

	status, err := CreateTaskData(taskData)
	if err != nil {
		w.scheduler.logger.Errorf("Failed to create task data for job %d: %v", w.jobID, err)
		return err
	}

	if !status {
		return fmt.Errorf("failed to create task data for job %d", w.jobID)
	}

	w.scheduler.logger.Infof("Task data created for job %v", w.jobID)

	_, err = SendTaskToPerformer(jobData, taskData)
	if err != nil {
		w.scheduler.logger.Errorf("Error sending task to performer: %v", err)
		return err
	}

	w.scheduler.logger.Infof("Task sent for job %d to performer %s", w.jobID, performer.KeeperAddress)

	return nil
}

func NewEventBasedWorker(jobData *types.Job, scheduler *JobScheduler) *EventBasedWorker {
	return &EventBasedWorker{
		jobID:           jobData.JobID,
		scheduler:       scheduler,
		chainID:         jobData.ChainID,
		jobData:         jobData,
		BaseWorker: BaseWorker{
			status:     "pending",
			maxRetries: 3,
		},
	}
}

func (w *EventBasedWorker) Start(ctx context.Context) {
	wsURL := w.getAlchemyWSURL()

	w.scheduler.logger.Infof("Connecting to Alchemy at %s", wsURL)

	client, err := ethclient.Dial(wsURL)
	if err != nil {
		w.scheduler.logger.Errorf("Failed to connect to Alchemy for job %d: %v", w.jobID, err)
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
		w.scheduler.logger.Errorf("Failed to subscribe to events for job %d: %v", w.jobID, err)
		w.status = "failed"
		w.Stop()
		return
	}
	w.subscription = sub

	go func() {
		for {
			select {
			case <-ctx.Done():
				return

			case err := <-sub.Err():
				w.scheduler.logger.Errorf("Event subscription error for job %d: %v", w.jobID, err)
				w.status = "failed"
				return

			case log := <-logs:
				w.scheduler.logger.Infof("Event detected for job %d: %v", w.jobID, log.TxHash.Hex())

				w.executeTask(w.jobData)
				return
			}
		}
	}()
}

func (w *EventBasedWorker) Stop() {
	if w.subscription != nil {
		w.subscription.Unsubscribe()
	}
	if w.client != nil {
		w.client.Close()
	}
	w.scheduler.RemoveJob(w.jobID)
}

func (w *EventBasedWorker) GetJobID() int64 {
	return w.jobID
}

func (w *EventBasedWorker) getAlchemyWSURL() string {
	apiKey := os.Getenv("ALCHEMY_API_KEY")

	var network string
	switch w.chainID {
	// case 1:
	// 	network = "eth-mainnet"
	// case 10:
	// 	network = "opt-mainnet"
	// case 8453:
	// 	network = "base-mainnet"
	// case 42161:
	// 	network = "arb-mainnet"
	case 17000:
		network = "eth-holesky"
	case 11155111:
		network = "eth-sepolia"
	case 84532:
		network = "base-sepolia"
	case 421614:
		network = "arb-sepolia"
	case 11155420:
		network = "opt-sepolia"
	default:
		network = "eth-holesky"
	}

	return fmt.Sprintf("wss://%s.g.alchemy.com/v2/%s", network, apiKey)
}

func (w *EventBasedWorker) executeTask(jobData *types.Job) error {
	w.scheduler.logger.Infof("Executing event-based job: %d", w.jobID)

	performer, err := GetPerformer()
	if err != nil {
		w.scheduler.logger.Errorf("Failed to get performer for job %d: %v", w.jobID, err)
		return err
	}

	taskData := &types.TaskData{
		JobID:           w.jobID,
		TaskDefinitionID: 2,
		TaskPerformer:   performer.KeeperAddress,
		IsApproved:      true,
	}

	status, err := CreateTaskData(taskData)
	if err != nil {
		w.scheduler.logger.Errorf("Failed to create task data for job %d: %v", w.jobID, err)
		return err
	}

	if !status {
		return fmt.Errorf("failed to create task data for job %d", w.jobID)
	}

	_, err = SendTaskToPerformer(jobData, taskData)
	if err != nil {
		w.scheduler.logger.Errorf("Error sending task to performer: %v", err)
		return err
	}

	w.scheduler.logger.Infof("Task sent for job %d to performer %s", w.jobID, performer)

	return nil
}

func NewConditionBasedWorker(jobData *types.Job, scheduler *JobScheduler) *ConditionBasedWorker {
	return &ConditionBasedWorker{
		jobID:         jobData.JobID,
		scheduler:     scheduler,
		jobData:       jobData,
		done:          make(chan bool),
		BaseWorker: BaseWorker{
			status:     "pending",
			maxRetries: 3,
		},
	}
}

func (w *ConditionBasedWorker) Start(ctx context.Context) {
	w.status = "running"
	w.ticker = time.NewTicker(1 * time.Second)

	w.scheduler.logger.Infof("Starting condition-based job %d", w.jobID)
	w.scheduler.logger.Infof("Listening to %s", w.jobData.ScriptIPFSUrl)

	go func() {
		defer w.Stop()

		for {
			select {
			case <-ctx.Done():
				return

			case <-w.done:
				return

			case <-w.ticker.C:
				satisfied, err := w.checkCondition()
				if err != nil {
					w.error = err.Error()
					w.currentRetry++

					if w.currentRetry >= w.maxRetries {
						w.status = "failed"
						w.Stop()
						return
					}
					continue
				}

				if satisfied {
					w.scheduler.logger.Infof("Condition satisfied for job %d", w.jobID)
					w.status = "completed"
					w.executeTask(w.jobData)
					return
				}
			}
		}
	}()
}

func (w *ConditionBasedWorker) Stop() {
	if w.ticker != nil {
		w.ticker.Stop()
	}
	close(w.done)
	w.scheduler.RemoveJob(w.jobID)
}

func (w *ConditionBasedWorker) GetJobID() int64 {
	return w.jobID
}

func (w *ConditionBasedWorker) checkCondition() (bool, error) {
	resp, err := http.Get(w.jobData.ScriptIPFSUrl)
	if err != nil {
		return false, fmt.Errorf("failed to fetch API data: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, fmt.Errorf("failed to read response body: %v", err)
	}

	w.scheduler.logger.Infof("API response: %s", string(body))

	return true, nil
}

func (w *ConditionBasedWorker) executeTask(jobData *types.Job) error {
	w.scheduler.logger.Infof("Executing condition-based job: %d", w.jobID)
	
	performer, err := GetPerformer()
	if err != nil {
		w.scheduler.logger.Errorf("Failed to get performer for job %d: %v", w.jobID, err)
		return err
	}

	taskData := &types.TaskData{
		JobID:           w.jobID,
		TaskDefinitionID: 3,
		TaskPerformer:   performer.KeeperAddress,
		IsApproved:      true,
	}

	status, err := CreateTaskData(taskData)
	if err != nil {
		w.scheduler.logger.Errorf("Failed to create task data for job %d: %v", w.jobID, err)
		return err
	}

	if !status {
		return fmt.Errorf("failed to create task data for job %d", w.jobID)
	}

	_, err = SendTaskToPerformer(jobData, taskData)
	if err != nil {
		w.scheduler.logger.Errorf("Error sending task to performer: %v", err)
		return err
	}

	w.scheduler.logger.Infof("Task sent for job %d to performer %s", w.jobID, performer)

	return nil
}

func (w *ConditionBasedWorker) GetStatus() string {
	return w.status
}

func (w *ConditionBasedWorker) GetError() string {
	return w.error
}

func (w *ConditionBasedWorker) GetRetries() int {
	return w.currentRetry
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

func (w *TimeBasedWorker) GetStatus() string {
	return w.status
}

func (w *TimeBasedWorker) GetError() string {
	return w.error
}

func (w *TimeBasedWorker) GetRetries() int {
	return w.currentRetry
}
