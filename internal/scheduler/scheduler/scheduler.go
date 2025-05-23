package scheduler

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"

	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// Worker interface defines the methods that all workers must implement
type Worker interface {
	Start(ctx context.Context)
	Stop()
	GetJobData() types.HandleCreateJobData
}

type TimeBasedScheduler struct {
	ctx    context.Context
	cancel context.CancelFunc
	mu     sync.RWMutex
	logger logging.Logger

	cronScheduler *cron.Cron
	workers       map[int64]Worker
	workerCtx     context.Context
	workerCancel  context.CancelFunc
}

// NewTimeBasedScheduler creates a new instance of TimeBasedScheduler
func NewTimeBasedScheduler(logger logging.Logger) (*TimeBasedScheduler, error) {
	ctx, cancel := context.WithCancel(context.Background())

	scheduler := &TimeBasedScheduler{
		ctx:           ctx,
		cancel:        cancel,
		logger:        logger,
		cronScheduler: cron.New(cron.WithSeconds()),
		workers:       make(map[int64]Worker),
		workerCtx:     ctx,
		workerCancel:  cancel,
	}

	scheduler.cronScheduler.Start()
	return scheduler, nil
}

// RegisterRoutes registers the HTTP routes for the scheduler
func (s *TimeBasedScheduler) RegisterRoutes(router *gin.Engine) {
	api := router.Group("/api/v1")
	{
		api.POST("/jobs", s.handleCreateJob)
		api.DELETE("/jobs/:id", s.handleDeleteJob)
		api.GET("/jobs/:id", s.handleGetJob)
		api.GET("/jobs", s.handleListJobs)

		// api.GET("/jobs/status/:id", s.handleGetJobStatus)
		// api.GET("/jobs/metrics", s.handleGetJobMetrics)
	}
}

// handleCreateJob handles the creation of a new time-based job
func (s *TimeBasedScheduler) handleCreateJob(c *gin.Context) {
	var jobData types.HandleCreateJobData
	if err := c.ShouldBindJSON(&jobData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := s.StartTimeBasedJob(jobData); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Job created successfully"})
}

// handleDeleteJob handles the deletion of a job
func (s *TimeBasedScheduler) handleDeleteJob(c *gin.Context) {
	jobID := c.Param("id")
	if jobID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Job ID is required"})
		return
	}

	id, err := strconv.ParseInt(jobID, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid job ID"})
		return
	}

	s.RemoveJob(id)
	c.JSON(http.StatusOK, gin.H{"message": "Job deleted successfully"})
}

// handleGetJob handles getting a specific job
func (s *TimeBasedScheduler) handleGetJob(c *gin.Context) {
	jobID := c.Param("id")
	if jobID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Job ID is required"})
		return
	}

	id, err := strconv.ParseInt(jobID, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid job ID"})
		return
	}

	worker := s.GetWorker(id)
	if worker == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Job not found"})
		return
	}

	c.JSON(http.StatusOK, worker.GetJobData())
}

// handleListJobs handles listing all jobs
func (s *TimeBasedScheduler) handleListJobs(c *gin.Context) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	jobs := make([]types.HandleCreateJobData, 0, len(s.workers))
	for _, w := range s.workers {
		jobs = append(jobs, w.GetJobData())
	}

	c.JSON(http.StatusOK, jobs)
}

func (s *TimeBasedScheduler) StartTimeBasedJob(jobData types.HandleCreateJobData) error {
	s.mu.Lock()
	worker := NewTimeBasedWorker(jobData, fmt.Sprintf("@every %ds", jobData.TimeInterval), s)
	s.workers[jobData.JobID] = worker
	s.mu.Unlock()

	go worker.Start(s.workerCtx)

	s.logger.Infof("Started time-based job %d", jobData.JobID)
	return nil
}

func (s *TimeBasedScheduler) RemoveJob(jobID int64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if worker, exists := s.workers[jobID]; exists {
		worker.Stop()
		delete(s.workers, jobID)
		s.logger.Infof("Removed job %d", jobID)
	}
}

func (s *TimeBasedScheduler) GetWorker(jobID int64) Worker {
	s.mu.RLock()
	defer s.mu.RUnlock()

	worker, exists := s.workers[jobID]
	if !exists {
		return nil
	}

	return worker
}

func (s *TimeBasedScheduler) Logger() logging.Logger {
	return s.logger
}

func (s *TimeBasedScheduler) GetCronScheduler() *cron.Cron {
	return s.cronScheduler
}
