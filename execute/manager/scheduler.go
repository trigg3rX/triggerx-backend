package manager

import (
	"fmt"
	"log"
	"sync"
	"time"

	"context"

	"strconv"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
	"github.com/robfig/cron/v3"
	"github.com/trigg3rX/triggerx-backend/pkg/database"
	"github.com/trigg3rX/triggerx-backend/pkg/network"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

var (
	ErrInvalidTimeframe = fmt.Errorf("invalid timeframe specified")
)

// JobScheduler enhanced with load balancing
type JobScheduler struct {
	jobs            map[string]*Job
	quorums         map[string]*Quorum
	jobQueue        chan *Job
	waitingQueue    []WaitingJob
	resources       SystemResources
	Cron            *cron.Cron
	ctx             context.Context
	cancel          context.CancelFunc
	mu              sync.RWMutex
	workersCount    int
	metricsInterval time.Duration
	waitingQueueMu  sync.RWMutex
	networkClient   *network.Messaging
	loadBalancer    *LoadBalancer
	dbClient        *database.Connection
}

// NewJobScheduler creates an enhanced scheduler with resource limits
func NewJobScheduler(workersCount int, dbClient *database.Connection) *JobScheduler {
	ctx, cancel := context.WithCancel(context.Background())
	cronInstance := cron.New(cron.WithSeconds())

	host, err := libp2p.New()
	if err != nil {
		log.Fatalf("Failed to create libp2p host: %v", err)
	}

	networkClient := network.NewMessaging(host, "task_manager")

	scheduler := &JobScheduler{
		jobs:          make(map[string]*Job),
		quorums:       make(map[string]*Quorum),
		jobQueue:      make(chan *Job, 1000),
		waitingQueue:  make([]WaitingJob, 0),
		Cron:          cronInstance,
		ctx:           ctx,
		cancel:        cancel,
		workersCount:  workersCount,
		networkClient: networkClient,
		loadBalancer:  NewLoadBalancer(),
		dbClient:      dbClient,
	}

	scheduler.startWorkers()
	go scheduler.loadBalancer.MonitorResources()
	go scheduler.processWaitingQueue()

	discovery := network.NewDiscovery(ctx, host, "task_manager")
	if err := discovery.SavePeerInfo(); err != nil {
		log.Printf("Failed to save task manager peer info: %v", err)
	}

	return scheduler
}

// Add method to fetch complete job data
func (js *JobScheduler) fetchCompleteJobData(jobID string) (*Job, error) {
	var jobData types.JobData

	// Query the database using the job ID
	err := js.dbClient.Session().Query(`
		SELECT job_id, jobType, user_address, chain_id, 
			   time_frame, time_interval, contract_address, 
			   target_function, arg_type, arguments, status, 
			   job_cost_prediction, script_function, script_ipfs_url
		FROM triggerx.job_data 
		WHERE job_id = ?`, jobID).Scan(
		&jobData.JobID, &jobData.JobType, &jobData.UserAddress,
		&jobData.ChainID, &jobData.TimeFrame, &jobData.TimeInterval,
		&jobData.ContractAddress, &jobData.TargetFunction,
		&jobData.ArgType, &jobData.Arguments, &jobData.Status,
		&jobData.JobCostPrediction, &jobData.ScriptFunction,
		&jobData.ScriptIpfsUrl)

	if err != nil {
		return nil, fmt.Errorf("failed to fetch job data: %v", err)
	}

	// Convert database model to Job struct
	job := &Job{
		JobID:             strconv.FormatInt(jobData.JobID, 10),
		ArgType:           strconv.Itoa(jobData.ArgType),
		ChainID:           strconv.Itoa(jobData.ChainID),
		ContractAddress:   jobData.ContractAddress,
		JobCostPrediction: float64(jobData.JobCostPrediction),
		Status:            "pending",
		TargetFunction:    jobData.TargetFunction,
		TimeFrame:         jobData.TimeFrame,
		TimeInterval:      int64(jobData.TimeInterval),
		UserID:            jobData.UserAddress,
		CreatedAt:         time.Now(),
		MaxRetries:        3, // Set default value
		CurrentRetries:    0,
		CodeURL:           jobData.ScriptIpfsUrl,
		Arguments:         make(map[string]interface{}),
	}

	// Convert arguments array to map
	for i, arg := range jobData.Arguments {
		job.Arguments[fmt.Sprintf("arg%d", i)] = arg
	}

	return job, nil
}

// AddJob enhanced with resource checking
func (js *JobScheduler) AddJob(jobID string) error {
	// Fetch complete job data from database
	job, err := js.fetchCompleteJobData(jobID)
	if err != nil {
		return fmt.Errorf("failed to fetch job data: %v", err)
	}

	if job.TimeFrame <= 0 {
		return ErrInvalidTimeframe
	}

	js.mu.Lock()
	defer js.mu.Unlock()

	if !js.loadBalancer.CheckResourceAvailability() {
		estimatedTime := js.loadBalancer.CalculateEstimatedWaitTime(js.jobs)
		js.loadBalancer.AddToWaitingQueue(job, estimatedTime)
		return nil
	}

	return js.scheduleJob(job)
}

// scheduleJob handles the actual job scheduling
func (js *JobScheduler) scheduleJob(job *Job) error {
	// Add to jobs map
	js.jobs[job.JobID] = job

	// Create cron spec
	cronSpec := fmt.Sprintf("@every %ds", job.TimeInterval)

	// Schedule initial execution
	time.AfterFunc(2*time.Second, func() {
		js.processJob(0, job)
	})

	// Schedule recurring executions
	_, err := js.Cron.AddFunc(cronSpec, func() {
		if time.Since(job.CreatedAt) > time.Duration(job.TimeFrame)*time.Second {
			return
		}

		js.mu.RLock()
		currentJob := js.jobs[job.JobID]
		shouldQueue := currentJob.Status != "processing" &&
			currentJob.Status != "completed" &&
			currentJob.Status != "failed"
		js.mu.RUnlock()

		if shouldQueue {
			js.jobQueue <- job
		}
	})

	if err != nil {
		return fmt.Errorf("failed to schedule job: %w", err)
	}

	log.Printf("Job %s scheduled successfully", job.JobID)
	return nil
}

// processWaitingQueue continuously checks and processes waiting jobs
func (js *JobScheduler) processWaitingQueue() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-js.ctx.Done():
			return
		case <-ticker.C:
			if js.loadBalancer.CheckResourceAvailability() {
				js.waitingQueueMu.Lock()
				if len(js.waitingQueue) > 0 {
					// Get next job from queue
					nextJob := js.waitingQueue[0]
					js.waitingQueue = js.waitingQueue[1:]
					js.waitingQueueMu.Unlock()

					// Schedule the job
					js.mu.Lock()
					err := js.scheduleJob(nextJob.Job)
					js.mu.Unlock()

					if err != nil {
						log.Printf("Failed to schedule waiting job %s: %v", nextJob.Job.JobID, err)
					} else {
						log.Printf("Successfully scheduled waiting job %s", nextJob.Job.JobID)
					}
				} else {
					js.waitingQueueMu.Unlock()
				}
			}
		}
	}
}

func (js *JobScheduler) transmitJobToKeeper(keeperName string, job *Job) error {
	// Ensure network client is initialized
	if js.networkClient == nil {
		return fmt.Errorf("network client not initialized")
	}

	// Get keeper peer info from registry
	registry, err := network.NewPeerRegistry()
	if err != nil {
		return fmt.Errorf("failed to create peer registry: %w", err)
	}

	// Get keeper service info
	keeperService, exists := registry.GetService(network.ServiceKeeper)
	if !exists {
		return fmt.Errorf("keeper service not found in registry")
	}

	// Convert keeper peer ID string to peer.ID
	peerID, err := peer.Decode(keeperService.PeerID)
	if err != nil {
		return fmt.Errorf("invalid peer ID for keeper: %w", err)
	}

	// Get keeper addresses
	if len(keeperService.Addresses) == 0 {
		return fmt.Errorf("no addresses found for keeper")
	}

	// Additional connection attempt with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Parse multiaddress
	maddr, err := multiaddr.NewMultiaddr(keeperService.Addresses[0])
	if err != nil {
		return fmt.Errorf("failed to parse multiaddress: %v", err)
	}

	// Attempt to connect to the peer before sending message
	if err := js.networkClient.GetHost().Connect(ctx, peer.AddrInfo{
		ID:    peerID,
		Addrs: []multiaddr.Multiaddr{maddr},
	}); err != nil {
		return fmt.Errorf("failed to connect to peer: %v", err)
	}

	// Prepare network message
	networkMessage := network.Message{
		From:      "task_manager",
		To:        keeperName,
		Content:   job,
		Type:      "JOB_TRANSMISSION",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	// Send the message
	err = js.networkClient.SendMessage(keeperName, peerID, networkMessage)
	if err != nil {
		return fmt.Errorf("failed to send job to keeper %s: %v", keeperName, err)
	}

	log.Printf("Job %s transmitted to keeper %s", job.JobID, keeperName)
	return nil
}

// func (js *JobScheduler) SetResourceLimits(cpuThreshold, memoryThreshold float64) {
// 	js.mu.Lock()
// 	defer js.mu.Unlock()

// 	js.loadBalancer.SetResourceLimits(cpuThreshold, memoryThreshold)
// }

// UpdateJob updates the status of a job in the scheduler
func (js *JobScheduler) UpdateJob(jobID int64) {
	js.mu.Lock()
	defer js.mu.Unlock()

	// Convert jobID to string to match our map key type
	jobIDStr := strconv.FormatInt(jobID, 10)

	// Check if job exists in our scheduler
	if job, exists := js.jobs[jobIDStr]; exists {
		// Fetch updated job data from database to ensure consistency
		updatedJob, err := js.fetchCompleteJobData(jobIDStr)
		if err != nil {
			log.Printf("Failed to fetch updated job data for job %s: %v", jobIDStr, err)
			return
		}

		// Update the job with new data while preserving runtime information
		updatedJob.CurrentRetries = job.CurrentRetries
		updatedJob.LastExecuted = job.LastExecuted
		updatedJob.Error = job.Error

		// Update the job in our map
		js.jobs[jobIDStr] = updatedJob

		log.Printf("Job %s status updated", jobIDStr)
	} else {
		log.Printf("Attempted to update non-existent job: %s", jobIDStr)
	}
}