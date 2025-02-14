package manager

import (
	"context"

	"github.com/robfig/cron/v3"
)

type Worker interface {
	Start(ctx context.Context)
	Stop()
	GetJobID() int64
}

// TimeBasedWorker handles cron-based jobs
type TimeBasedWorker struct {
	jobID     int64
	scheduler *JobScheduler
	cron      *cron.Cron
	schedule  string
}

func NewTimeBasedWorker(jobID int64, schedule string, scheduler *JobScheduler) *TimeBasedWorker {
	return &TimeBasedWorker{
		jobID:     jobID,
		scheduler: scheduler,
		cron:      cron.New(cron.WithSeconds()),
		schedule:  schedule,
	}
}

func (w *TimeBasedWorker) Start(ctx context.Context) {
	w.cron.AddFunc(w.schedule, func() {
		w.scheduler.logger.Infof("Executing time-based job: %d", w.jobID)
	})
	w.cron.Start()

	go func() {
		<-ctx.Done()
		w.Stop()
	}()
}

func (w *TimeBasedWorker) Stop() {
	w.cron.Stop()
	w.scheduler.RemoveJob(w.jobID)
}

func (w *TimeBasedWorker) GetJobID() int64 {
	return w.jobID
}
