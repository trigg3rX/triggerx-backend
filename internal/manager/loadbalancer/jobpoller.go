package loadbalancer

import (
	"context"
	"sort"
	"time"

	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// DatabaseClient defines the interface for database operations
type DatabaseClient interface {
	GetPendingJobs() ([]types.JobData, error)
	GetJobsForExecution(window time.Duration) ([]types.JobData, error)
	AssignJobToManager(jobID int64, managerID string) error
	UpdateJobStatus(jobID int64, status string) error
	GetJobByID(jobID int64) (*types.HandleCreateJobData, error)
}

type JobPoller struct {
	managerID    string
	dbClient     DatabaseClient
	loadBalancer *LoadBalancer
	logger       logging.Logger
	batchSize    int
	pollInterval time.Duration
	execWindow   time.Duration
}

func NewJobPoller(managerID string, dbClient DatabaseClient, lb *LoadBalancer) *JobPoller {
	return &JobPoller{
		managerID:    managerID,
		dbClient:     dbClient,
		loadBalancer: lb,
		logger:       logging.GetLogger(logging.Development, logging.ManagerProcess),
		batchSize:    10, // Process 10 jobs at a time
		pollInterval: 5 * time.Second,
		execWindow:   5 * time.Second,
	}
}

// Start begins polling for new jobs
func (jp *JobPoller) Start(ctx context.Context) {
	ticker := time.NewTicker(jp.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			jp.pollJobs()
		}
	}
}

func (jp *JobPoller) pollJobs() {
	// Get jobs that need execution in the next window
	jobs, err := jp.dbClient.GetJobsForExecution(jp.execWindow)
	if err != nil {
		jp.logger.Errorf("Failed to get jobs for execution: %v", err)
		return
	}

	if len(jobs) == 0 {
		jp.logger.Debug("No jobs need execution in the next window")
		return
	}

	// Sort jobs by priority and execution time
	sort.Slice(jobs, func(i, j int) bool {
		if jobs[i].Priority != jobs[j].Priority {
			return jobs[i].Priority > jobs[j].Priority
		}
		// If priorities are equal, sort by execution time
		return jobs[i].NextExecutionTime.Before(jobs[j].NextExecutionTime)
	})

	// Process jobs in batches
	for i := 0; i < len(jobs); i += jp.batchSize {
		end := i + jp.batchSize
		if end > len(jobs) {
			end = len(jobs)
		}

		batch := jobs[i:end]
		jp.processJobBatch(batch)
	}
}

func (jp *JobPoller) processJobBatch(jobs []types.JobData) {
	for _, job := range jobs {
		// Check if system has enough resources
		if !jp.loadBalancer.CheckResourceAvailability() {
			jp.logger.Warnf("System resources exceeded, skipping job %d", job.JobID)
			continue
		}

		// Get the least loaded manager
		managerID, err := jp.loadBalancer.GetLeastLoadedManager()
		if err != nil || managerID == "" {
			jp.logger.Warn("No healthy managers available")
			continue
		}

		// Assign job to manager
		if err := jp.dbClient.AssignJobToManager(job.JobID, managerID); err != nil {
			jp.logger.Errorf("Failed to assign job %d to manager %s: %v", job.JobID, managerID, err)
			continue
		}

		// Update job status
		if err := jp.dbClient.UpdateJobStatus(job.JobID, "assigned"); err != nil {
			jp.logger.Errorf("Failed to update job %d status: %v", job.JobID, err)
			continue
		}

		// Update load balancer's view of manager load
		jp.loadBalancer.UpdateManagerLoad(managerID, 1)

		jp.logger.Infof("Assigned job %d to manager %s for execution at %v",
			job.JobID, managerID, job.NextExecutionTime)
	}
}

// UpdateJobStatus updates the status of a job in the database
func (jp *JobPoller) UpdateJobStatus(jobID int64, status string) error {
	return jp.dbClient.UpdateJobStatus(jobID, status)
}

// GetJobDetails retrieves detailed information about a specific job
func (jp *JobPoller) GetJobDetails(jobID int64) (*types.HandleCreateJobData, error) {
	return jp.dbClient.GetJobByID(jobID)
}
