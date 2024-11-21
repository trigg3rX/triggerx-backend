package manager

import (
    "context"
	"errors"
    "fmt"
    "log"
    "math/rand"
    "sync"
    "time"
    "runtime"
    "github.com/robfig/cron/v3"
)

type Job struct {
    JobID             string                 `json:"job_id"`
    ArgType           string                 `json:"arg_type"`
    Arguments         map[string]interface{} `json:"arguments"`
    ChainID           string                 `json:"chain_id"`
    ContractAddress   string                 `json:"contract_address"`
    JobCostPrediction float64                `json:"job_cost_prediction"`
    Stake             float64                `json:"stake"`
    Status            string                 `json:"status"`
    TargetFunction    string                 `json:"target_function"`
    TimeFrame         int                    `json:"time_frame"`
    TimeInterval      int                    `json:"time_interval"`
    UserID            string                 `json:"user_id"`
    CreatedAt         time.Time              `json:"created_at"`
    LastError         *JobError              `json:"last_error,omitempty"`
    MaxRetries        int                    `json:"max_retries"`
    RetryCount        int                    `json:"retry_count"`
}

type JobError struct {
    Timestamp time.Time
    Error     error
    RetryCount int
}

var (
    ErrJobNotFound     = errors.New("job not found")
    ErrQuorumNotFound  = errors.New("no available quorum")
    ErrJobProcessing   = errors.New("job is already processing")
    ErrInvalidTimeframe = errors.New("invalid timeframe")
)

type QuorumStatus struct {
    ID           string  `json:"id"`
    HeadID       string  `json:"head_id"`
    ActiveKeepers int    `json:"active_keepers"`
    CurrentLoad  float64 `json:"current_load"`
    MaxLoad      float64 `json:"max_load"`
}

// Keeper represents a node that can execute jobs
type Keeper struct {
    ID               string
    CPUUsage         float64
    MemoryUsage      float64
    ActiveJobs       int
    MaxJobs          int
    IsQuorumHead     bool
    LastHeartbeat    time.Time
}

// Quorum represents a group of keepers
type Quorum struct {
    ID       string
    Keepers  []*Keeper
    Head     *Keeper
    MaxLoad  float64
    mu       sync.RWMutex
}

// JobScheduler manages job scheduling and load balancing
type JobScheduler struct {
    jobs              map[string]*Job
    quorums           map[string]*Quorum
    jobQueue          chan *Job
    Cron              *cron.Cron           
    ctx               context.Context
    cancel            context.CancelFunc
    mu                sync.RWMutex
    workersCount      int
    metricsInterval   time.Duration
}


// SystemMetrics represents system resource usage
type SystemMetrics struct {
    CPUUsage    float64
    MemoryUsage float64
}

// NewJobScheduler creates a new job scheduler instance
func NewJobScheduler(workersCount int) *JobScheduler {
    ctx, cancel := context.WithCancel(context.Background())
    cronInstance := cron.New(cron.WithSeconds())
    
    scheduler := &JobScheduler{
        jobs:             make(map[string]*Job),
        quorums:          make(map[string]*Quorum),
        jobQueue:         make(chan *Job, 1000),
        Cron:            cronInstance,
        ctx:             ctx,
        cancel:          cancel,
        workersCount:    workersCount,
        metricsInterval: 5 * time.Second,
    }

    scheduler.initializeQuorums()
    scheduler.startWorkers()
    go scheduler.collectMetrics()
    go scheduler.cleanupExpiredJobs()
    
    return scheduler
}
// initializeQuorums creates sample quorums with keepers
func (js *JobScheduler) initializeQuorums() {
    for i := 0; i < 3; i++ {
        quorum := &Quorum{
            ID:      fmt.Sprintf("quorum_%d", i),
            Keepers: make([]*Keeper, 4),
            MaxLoad: 800.0, // 80% max load threshold
        }

        // Initialize keepers with more realistic initial values
        for j := 0; j < 4; j++ {
            keeper := &Keeper{
                ID:            fmt.Sprintf("keeper_%d_%d", i, j),
                MaxJobs:       5,
                ActiveJobs:    0,
                CPUUsage:      40.0 + float64(rand.Intn(20)), // Initial CPU 40-60%
                MemoryUsage:   40.0 + float64(rand.Intn(20)), // Initial Memory 40-60%
                LastHeartbeat: time.Now(),
            }
            quorum.Keepers[j] = keeper
        }

        js.selectQuorumHead(quorum)
        js.quorums[quorum.ID] = quorum
    }
}

// GetJob returns a job by its ID
func (js *JobScheduler) GetJob(jobID string) (*Job, bool) {
    js.mu.RLock()
    defer js.mu.RUnlock()
    
    job, exists := js.jobs[jobID]
    return job, exists
}

func (js *JobScheduler) GetQuorumsStatus() []QuorumStatus {
    js.mu.RLock()
    defer js.mu.RUnlock()

    status := make([]QuorumStatus, 0, len(js.quorums))
    
    for _, quorum := range js.quorums {
        quorum.mu.RLock()
        activeKeepers := 0
        for _, keeper := range quorum.Keepers {
            if time.Since(keeper.LastHeartbeat) < 30*time.Second {
                activeKeepers++
            }
        }
        
        currentLoad := js.calculateQuorumLoad(quorum)
        
        status = append(status, QuorumStatus{
            ID:           quorum.ID,
            HeadID:      quorum.Head.ID,
            ActiveKeepers: activeKeepers,
            CurrentLoad:  currentLoad,
            MaxLoad:      quorum.MaxLoad,
        })
        quorum.mu.RUnlock()
    }
    
    return status
}

// selectQuorumHead randomly selects a keeper as quorum head
func (js *JobScheduler) selectQuorumHead(quorum *Quorum) {
    quorum.mu.Lock()
    defer quorum.mu.Unlock()

    // Reset current head if exists
    if quorum.Head != nil {
        quorum.Head.IsQuorumHead = false
    }

    // Randomly select new head
    headIndex := rand.Intn(len(quorum.Keepers))
    quorum.Head = quorum.Keepers[headIndex]
    quorum.Head.IsQuorumHead = true
}

// AddJob adds a new job to the scheduler
func (js *JobScheduler) AddJob(job *Job) error {
    if job.TimeFrame <= 0 {
        return ErrInvalidTimeframe
    }

    if job.MaxRetries == 0 {
        job.MaxRetries = 3 // Set default max retries
    }

    js.mu.Lock()
    defer js.mu.Unlock()

    // Add to jobs map
    js.jobs[job.JobID] = job
    
    // Create cron spec based on time interval
    cronSpec := fmt.Sprintf("@every %ds", job.TimeInterval)
    
    // Schedule initial execution after a short delay
    time.AfterFunc(2*time.Second, func() {
        js.processJob(job)
    })
    
    // Schedule recurring executions
    _, err := js.Cron.AddFunc(cronSpec, func() {
        // Check if the job should still be executed
        if time.Since(job.CreatedAt) > time.Duration(job.TimeFrame)*time.Second {
            log.Printf("Job %s exceeded timeframe, stopping execution", job.JobID)
            return
        }

        // Only queue if job isn't already processing
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

func (js *JobScheduler) handleJobError(job *Job, err error) {
    js.mu.Lock()
    defer js.mu.Unlock()

    jobError := &JobError{
        Timestamp:  time.Now(),
        Error:      err,
        RetryCount: job.RetryCount,
    }

    job.LastError = jobError
    job.RetryCount++

    // Check if we should retry
    if job.RetryCount <= job.MaxRetries {
        // Calculate exponential backoff
        backoff := time.Duration(1<<uint(job.RetryCount-1)) * time.Second
        
        // Schedule retry
        job.Status = "pending_retry"
        time.AfterFunc(backoff, func() {
            js.jobQueue <- job
        })
        
        log.Printf("Scheduled retry %d/%d for job %s in %v", 
            job.RetryCount, job.MaxRetries, job.JobID, backoff)
    } else {
        job.Status = "failed"
        log.Printf("Job %s failed after %d retries. Last error: %v", 
            job.JobID, job.MaxRetries, err)
    }
}

// startWorkers initializes the worker pool
func (js *JobScheduler) startWorkers() {
    for i := 0; i < js.workersCount; i++ {
        go js.worker()
    }
}

// worker processes jobs from the queue
func (js *JobScheduler) worker() {
    for {
        select {
        case <-js.ctx.Done():
            return
        case job := <-js.jobQueue:
            // Recover from panics in job processing
            defer func() {
                if r := recover(); r != nil {
                    log.Printf("Recovered from panic in worker processing job %s: %v", job.JobID, r)
                    js.handleJobError(job, fmt.Errorf("job processing panicked: %v", r))
                }
            }()

            // Check if job is still within its timeframe
            if time.Since(job.CreatedAt) <= time.Duration(job.TimeFrame)*time.Second {
                js.processJob(job)
            } else {
                log.Printf("Job %s exceeded timeframe, marking as expired", job.JobID)
                js.mu.Lock()
                if currentJob, exists := js.jobs[job.JobID]; exists {
                    currentJob.Status = "expired"
                }
                js.mu.Unlock()
            }
        }
    }
}

func (js *JobScheduler) cleanupExpiredJobs() {
    ticker := time.NewTicker(1 * time.Minute)
    defer ticker.Stop()

    for {
        select {
        case <-js.ctx.Done():
            return
        case <-ticker.C:
            js.mu.Lock()
            for _, job := range js.jobs {
                if time.Since(job.CreatedAt) > time.Duration(job.TimeFrame)*time.Second {
                    if job.Status != "completed" && job.Status != "failed" {
                        job.Status = "expired"
                    }
                }
            }
            js.mu.Unlock()
        }
    }
}

// processJob handles the execution of a job
func (js *JobScheduler) processJob(job *Job) {
    js.mu.Lock()
    currentJob, exists := js.jobs[job.JobID]
    if !exists {
        js.mu.Unlock()
        log.Printf("Job %s not found for processing", job.JobID)
        return
    }

    // Check if job is already processing
    if currentJob.Status == "processing" {
        js.mu.Unlock()
        log.Printf("Job %s is already processing", job.JobID)
        return
    }

    // Update job status to processing
    currentJob.Status = "processing"
    js.mu.Unlock()

    // Find best quorum based on load balancing
    quorum := js.selectBestQuorum()
    if quorum == nil {
        js.handleJobError(currentJob, ErrQuorumNotFound)
        return
    }

    // Send job to quorum head
    err := js.sendJobToQuorumHead(currentJob, quorum)
    if err != nil {
        js.handleJobError(currentJob, fmt.Errorf("failed to send job to quorum head: %w", err))
        return
    }

    // Update job status and quorum metrics
    js.mu.Lock()
    currentJob.Status = "completed"
    currentJob.LastError = nil // Clear any previous errors
    js.mu.Unlock()

    quorum.mu.Lock()
    quorum.Head.ActiveJobs++
    quorum.mu.Unlock()

    log.Printf("Job %s successfully processed by quorum %s", currentJob.JobID, quorum.ID)
}

// selectBestQuorum chooses the best quorum based on load balancing metrics
func (js *JobScheduler) selectBestQuorum() *Quorum {
    js.mu.RLock()
    defer js.mu.RUnlock()

    var bestQuorum *Quorum
    lowestLoad := float64(100)

    for _, quorum := range js.quorums {
        quorum.mu.RLock()
        load := js.calculateQuorumLoad(quorum)
        if load < lowestLoad && load < quorum.MaxLoad {
            lowestLoad = load
            bestQuorum = quorum
        }
        quorum.mu.RUnlock()
    }

    return bestQuorum
}

// calculateQuorumLoad calculates the current load of a quorum
func (js *JobScheduler) calculateQuorumLoad(quorum *Quorum) float64 {
    var totalCPU, totalMemory float64
    activeKeepers := 0

    for _, keeper := range quorum.Keepers {
        if time.Since(keeper.LastHeartbeat) < 30*time.Second {
            totalCPU += keeper.CPUUsage
            totalMemory += keeper.MemoryUsage
            activeKeepers++
        }
    }

    if activeKeepers == 0 {
        return 1000.0 // Return very high load if no active keepers
    }

    // Calculate average load based on active keepers
    avgCPU := totalCPU / float64(activeKeepers)
    avgMemory := totalMemory / float64(activeKeepers)
    
    // Consider active jobs in load calculation
    jobLoad := float64(quorum.Head.ActiveJobs) / float64(quorum.Head.MaxJobs) * 100
    
    // Weight the different factors (CPU: 40%, Memory: 30%, Jobs: 30%)
    return (avgCPU * 0.4) + (avgMemory * 0.3) + (jobLoad * 0.3)
}

// executeJob handles the actual execution of the job on the selected quorum
func (js *JobScheduler) sendJobToQuorumHead(job *Job, quorum *Quorum) error {
    quorum.mu.RLock()
    defer quorum.mu.RUnlock()

    if quorum.Head == nil {
        return fmt.Errorf("no head found for quorum %s", quorum.ID)
    }

    // Simulate job execution by the quorum head
    log.Printf("Sending job %s to quorum head %s", job.JobID, quorum.Head.ID)

    // Simulated execution: Add some delay to mimic job processing
    processingTime := time.Duration(rand.Intn(3)+1) * time.Second
    time.Sleep(processingTime)

    // Randomly simulate job success or failure
    if rand.Intn(10) < 8 { // 80% success rate
        log.Printf("Job %s executed successfully by quorum head %s", job.JobID, quorum.Head.ID)
        return nil
    } else {
        log.Printf("Job %s failed execution by quorum head %s", job.JobID, quorum.Head.ID)
        return fmt.Errorf("execution failed for job %s", job.JobID)
    }
}

// collectMetrics periodically updates system metrics
func (js *JobScheduler) collectMetrics() {
    ticker := time.NewTicker(js.metricsInterval)
    defer ticker.Stop()

    for {
        select {
        case <-js.ctx.Done():
            return
        case <-ticker.C:
            js.updateSystemMetrics()
        }
    }
}

// updateSystemMetrics updates CPU and memory metrics for all keepers
func (js *JobScheduler) updateSystemMetrics() {
    js.mu.Lock()
    defer js.mu.Unlock()

    for _, quorum := range js.quorums {
        quorum.mu.Lock()
        for _, keeper := range quorum.Keepers {
            // Simulate realistic system metrics
            var m runtime.MemStats
            runtime.ReadMemStats(&m)
            
            // Generate more realistic CPU usage (40-80% range)
            keeper.CPUUsage = 40.0 + float64(rand.Intn(40))
            
            // Calculate memory usage (40-70% range)
            keeper.MemoryUsage = 40.0 + float64(rand.Intn(30))
            
            // Update heartbeat
            keeper.LastHeartbeat = time.Now()
        }
        quorum.mu.Unlock()
    }
}

// Stop gracefully shuts down the scheduler
func (js *JobScheduler) Stop() {
    js.cancel()
    js.Cron.Stop()  // Note the capital C
    close(js.jobQueue)
}

// Example usage
func main() {
    // Create a new scheduler with 5 workers
    scheduler := NewJobScheduler(5)
    
    // Start the cron scheduler
    scheduler.Cron.Start()  // Note the capital C

    // Example job
    job := &Job{
        JobID:             "job_1",
        ArgType:           "contract_call",
        Arguments:         map[string]interface{}{"function": "transfer"},
        ChainID:           "chain_1",
        ContractAddress:   "0x123...",
        JobCostPrediction: 0.5,
        Stake:            1.0,
        Status:           "pending",
        TargetFunction:   "execute",
        TimeFrame:        60,
        TimeInterval:     300,
        UserID:           "user_1",
        CreatedAt:        time.Now(),
    }

    // Add job to scheduler
    if err := scheduler.AddJob(job); err != nil {
        log.Fatal(err)
    }

    // Keep the application running
    select {}
}