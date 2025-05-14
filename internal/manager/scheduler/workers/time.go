package workers

import (
	"context"
	"fmt"
	"time"

	"github.com/robfig/cron/v3"

	"github.com/trigg3rX/triggerx-backend/pkg/types"
	"github.com/trigg3rX/triggerx-backend/internal/manager/scheduler/services"
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
	scheduler JobScheduler
	cron      *cron.Cron
	schedule  string
	jobData   *types.HandleCreateJobData
	startTime time.Time
	BaseWorker
}

func NewTimeBasedWorker(jobData types.HandleCreateJobData, schedule string, scheduler JobScheduler) *TimeBasedWorker {
	return &TimeBasedWorker{
		jobID:     jobData.JobID,
		scheduler: scheduler,
		cron:      cron.New(cron.WithSeconds()),
		schedule:  schedule,
		jobData:   &jobData,
		startTime: time.Now().UTC(),
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

	var triggerData types.TriggerData
	triggerData.TimeInterval = w.jobData.TimeInterval
	triggerData.TriggerTxHash = ""
	triggerData.ConditionParams = make(map[string]interface{})

	w.status = "running"

	// time.AfterFunc(1*time.Second, func() {
	// 	triggerData.Timestamp = time.Now().UTC()
	// 	triggerData.LastExecuted = time.Now().UTC()

	// 	if err := w.executeTask(w.jobData, &triggerData); err != nil {
	// 		w.handleError(err)
	// 	}
	// })

	w.cron.AddFunc(w.schedule, func() {
		if w.jobData.TimeFrame > 0 && time.Since(w.startTime) > time.Duration(w.jobData.TimeFrame)*time.Second {
			w.Stop()
			return
		}

		if w.status != "running" {
			return
		}

		triggerData.Timestamp = time.Now().UTC()

		if err := w.executeTask(w.jobData, &triggerData); err != nil {
			w.handleError(err)
		}

		triggerData.LastExecuted = time.Now().UTC()
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

func (w *TimeBasedWorker) executeTask(jobData *types.HandleCreateJobData, triggerData *types.TriggerData) error {
	w.scheduler.Logger().Infof("Executing time-based job: %d", w.jobID)

	taskData := &types.CreateTaskData{
		JobID:            w.jobID,
		TaskDefinitionID: jobData.TaskDefinitionID,
		TaskPerformerID:  0,
	}

	performerData, err := services.GetPerformer()
	if err != nil {
		w.scheduler.Logger().Errorf("Failed to get performer data for job %d: %v", w.jobID, err)
		return err
	}

	taskData.TaskPerformerID = performerData.KeeperID

	w.scheduler.Logger().Infof("Task data: %d | %d | %d", taskData.JobID, taskData.TaskDefinitionID, taskData.TaskPerformerID)

	taskID, status, err := services.CreateTaskData(taskData)
	if err != nil {
		w.scheduler.Logger().Errorf("Failed to create task data for job %d: %v", w.jobID, err)
		return err
	}

	triggerData.TaskID = taskID

	if !status {
		return fmt.Errorf("failed to create task data for job %d", w.jobID)
	}

	w.scheduler.Logger().Infof("Task ID %d created for job %v", taskID, w.jobID)

	status, err = services.SendTaskToPerformer(jobData, triggerData, performerData)
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

func (w *TimeBasedWorker) GetStatus() string {
	return w.status
}

func (w *TimeBasedWorker) GetError() string {
	return w.error
}

func (w *TimeBasedWorker) GetRetries() int {
	return w.currentRetry
}

func (w *BaseWorker) handleLinkedJob(scheduler JobScheduler, jobData *types.HandleCreateJobData) error {
	if jobData.LinkJobID <= 0 {
		return nil
	}

	scheduler.UpdateJobChainStatus(jobData.JobID, "completed")
	scheduler.Logger().Infof("Found linked job %d for job %d", jobData.LinkJobID, jobData.JobID)

	linkedJob, err := scheduler.GetJobDetails(jobData.LinkJobID)
	if err != nil {
		return fmt.Errorf("failed to fetch linked job %d details: %v", jobData.LinkJobID, err)
	}

	switch {
	case linkedJob.TaskDefinitionID == 1 || linkedJob.TaskDefinitionID == 2:
		return scheduler.StartTimeBasedJob(*linkedJob)
	case linkedJob.TaskDefinitionID == 3 || linkedJob.TaskDefinitionID == 4:
		return scheduler.StartEventBasedJob(*linkedJob)
	case linkedJob.TaskDefinitionID == 5 || linkedJob.TaskDefinitionID == 6:
		return scheduler.StartConditionBasedJob(*linkedJob)
	default:
		return fmt.Errorf("invalid job type for linked job %d", linkedJob.JobID)
	}
}

func (w *TimeBasedWorker) UpdateLastExecutedTime(timestamp time.Time) {
	if w.jobData != nil {
		w.jobData.LastExecutedAt = timestamp
		w.scheduler.Logger().Infof("Updated LastExecutedAt for job %d to %v", w.jobID, timestamp)
	}
}
