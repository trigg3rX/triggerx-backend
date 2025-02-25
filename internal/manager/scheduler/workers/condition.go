package workers

// import (
// 	"context"
// 	"fmt"
// 	"io"
// 	"net/http"
// 	"time"

// 	"github.com/trigg3rX/triggerx-backend/pkg/types"
// )

// type ConditionBasedWorker struct {
// 	jobID     int64
// 	scheduler *JobScheduler
// 	jobData   *types.Job
// 	ticker    *time.Ticker
// 	done      chan bool
// 	BaseWorker
// }

// func NewConditionBasedWorker(jobData *types.Job, scheduler *JobScheduler) *ConditionBasedWorker {
// 	return &ConditionBasedWorker{
// 		jobID:     jobData.JobID,
// 		scheduler: scheduler,
// 		jobData:   jobData,
// 		done:      make(chan bool),
// 		BaseWorker: BaseWorker{
// 			status:     "pending",
// 			maxRetries: 3,
// 		},
// 	}
// }

// func (w *ConditionBasedWorker) Start(ctx context.Context) {
// 	w.status = "running"
// 	w.ticker = time.NewTicker(1 * time.Second)

// 	w.scheduler.logger.Infof("Starting condition-based job %d", w.jobID)
// 	w.scheduler.logger.Infof("Listening to %s", w.jobData.ScriptIPFSUrl)

// 	var triggerData types.TriggerData
// 	triggerData.TimeInterval = w.jobData.TimeInterval
// 	triggerData.LastExecuted = time.Now()
// 	triggerData.TriggerTxHash = ""

// 	go func() {
// 		defer w.Stop()

// 		for {
// 			select {
// 			case <-ctx.Done():
// 				return

// 			case <-w.done:
// 				return

// 			case <-w.ticker.C:
// 				satisfied, err := w.checkCondition()
// 				if err != nil {
// 					w.error = err.Error()
// 					w.currentRetry++

// 					if w.currentRetry >= w.maxRetries {
// 						w.status = "failed"
// 						w.Stop()
// 						return
// 					}
// 					continue
// 				}

// 				if satisfied {
// 					w.scheduler.logger.Infof("Condition satisfied for job %d", w.jobID)

// 					triggerData.Timestamp = time.Now()
// 					triggerData.ConditionParams = make(map[string]interface{})

// 					if err := w.executeTask(w.jobData, &triggerData); err != nil {
// 						w.handleError(err)
// 					}
// 					return
// 				}
// 			}
// 		}
// 	}()
// }

// func (w *ConditionBasedWorker) Stop() {
// 	if w.ticker != nil {
// 		w.ticker.Stop()
// 	}
// 	close(w.done)
// 	w.scheduler.RemoveJob(w.jobID)
// }

// func (w *ConditionBasedWorker) checkCondition() (bool, error) {
// 	resp, err := http.Get(w.jobData.ScriptIPFSUrl)
// 	if err != nil {
// 		return false, fmt.Errorf("failed to fetch API data: %v", err)
// 	}
// 	defer resp.Body.Close()

// 	body, err := io.ReadAll(resp.Body)
// 	if err != nil {
// 		return false, fmt.Errorf("failed to read response body: %v", err)
// 	}

// 	w.scheduler.logger.Infof("API response: %s", string(body))

// 	return true, nil
// }

// func (w *ConditionBasedWorker) executeTask(jobData *types.Job, triggerData *types.TriggerData) error {
// 	w.scheduler.logger.Infof("Executing condition-based job: %d", w.jobID)

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

// func (w *ConditionBasedWorker) GetJobID() int64 {
// 	return w.jobID
// }

// func (w *ConditionBasedWorker) GetStatus() string {
// 	return w.status
// }

// func (w *ConditionBasedWorker) GetError() string {
// 	return w.error
// }

// func (w *ConditionBasedWorker) GetRetries() int {
// 	return w.currentRetry
// }
