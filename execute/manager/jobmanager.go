// github.com/trigg3rX/go-backend/execute/manager/jobmanager.go
package manager

import (
    "encoding/json"
    "fmt"
    "log"
    "os"
    "strings"
    "sync"
    "time"

    "context"
    "github.com/libp2p/go-libp2p"
    "github.com/libp2p/go-libp2p/core/peer"
    "github.com/shirou/gopsutil/v3/cpu"
    "github.com/shirou/gopsutil/v3/mem"
    "github.com/robfig/cron/v3"
    "github.com/trigg3rX/go-backend/pkg/network"
)

var (
    ErrInvalidTimeframe = fmt.Errorf("invalid timeframe specified")
)

// SystemResources tracks system resource usage
type SystemResources struct {
    CPUUsage    float64
    MemoryUsage float64
    MaxCPU      float64
    MaxMemory   float64
}

// WaitingJob represents a job waiting in queue
type WaitingJob struct {
    Job           *Job
    EstimatedTime time.Time
}

// JobScheduler enhanced with load balancing
type JobScheduler struct {
    jobs              map[string]*Job
    quorums           map[string]*Quorum
    jobQueue          chan *Job
    waitingQueue      []WaitingJob
    resources         SystemResources
    Cron              *cron.Cron
    ctx               context.Context
    cancel            context.CancelFunc
    mu                sync.RWMutex
    workersCount      int
    metricsInterval   time.Duration
    waitingQueueMu    sync.RWMutex
    networkClient *network.Messaging 
}

// NewJobScheduler creates an enhanced scheduler with resource limits
func NewJobScheduler(workersCount int) *JobScheduler {
    ctx, cancel := context.WithCancel(context.Background())
    cronInstance := cron.New(cron.WithSeconds())
    
    host, err := libp2p.New()
    if err != nil {
        log.Fatalf("Failed to create libp2p host: %v", err)
    }

    networkClient := network.NewMessaging(host, "task_manager")
    
    scheduler := &JobScheduler{
        jobs:             make(map[string]*Job),
        quorums:          make(map[string]*Quorum),
        jobQueue:         make(chan *Job, 1000),
        waitingQueue:     make([]WaitingJob, 0),
        resources: SystemResources{
            MaxCPU:    10.0, // 10% CPU threshold
            MaxMemory: 80.0, // 80% Memory threshold
        },
        Cron:            cronInstance,
        ctx:             ctx,
        cancel:          cancel,
        workersCount:    workersCount,
        metricsInterval: 5 * time.Second,
        networkClient: networkClient,
    }

    
        scheduler.initializeQuorums()
        scheduler.startWorkers()
        go scheduler.monitorResources()
        go scheduler.processWaitingQueue()
        
    

    discovery := network.NewDiscovery(ctx, host, "task_manager")
    if err := discovery.SavePeerInfo(); err != nil {
        log.Printf("Failed to save task manager peer info: %v", err)
    }

    return scheduler
}

func (js *JobScheduler) transmitJobToKeeper(keeperName string, job *Job) error {
    // Ensure network client is initialized
    if js.networkClient == nil {
        return fmt.Errorf("network client not initialized")
    }

    // Load peer information
    peerInfos, err := js.loadPeerInfo()
    if err != nil {
        return fmt.Errorf("failed to load peer info: %v", err)
    }

    // Find the specific keeper's peer information
    peerInfo, exists := peerInfos[keeperName]
    if !exists {
        return fmt.Errorf("keeper %s not found in peer information", keeperName)
    }

    // Convert peer address to peer ID
    peerID, err := peer.Decode(strings.Split(peerInfo.Address, "/p2p/")[1])
    if err != nil {
        return fmt.Errorf("invalid peer ID for keeper %s: %v", keeperName, err)
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

// Helper method to load peer information
func (js *JobScheduler) loadPeerInfo() (map[string]network.PeerInfo, error) {
    file, err := os.Open(network.PeerInfoFilePath)
    if err != nil {
        return nil, fmt.Errorf("unable to open peer info file: %v", err)
    }
    defer file.Close()

    var peerInfos map[string]network.PeerInfo
    decoder := json.NewDecoder(file)
    if err := decoder.Decode(&peerInfos); err != nil {
        return nil, fmt.Errorf("unable to decode peer info: %v", err)
    }

    return peerInfos, nil
}

// monitorResources continuously monitors system resources
func (js *JobScheduler) monitorResources() {
    ticker := time.NewTicker(js.metricsInterval)
    defer ticker.Stop()

    for {
        select {
        case <-js.ctx.Done():
            return
        case <-ticker.C:
            cpuPercent, err := cpu.Percent(time.Second, false)
            if err == nil && len(cpuPercent) > 0 {
                js.resources.CPUUsage = cpuPercent[0]
            }

            memInfo, err := mem.VirtualMemory()
            if err == nil {
                js.resources.MemoryUsage = memInfo.UsedPercent
            }

            // Log current resource usage
            log.Printf("System Resources - CPU: %.2f%%, Memory: %.2f%%",
                js.resources.CPUUsage, js.resources.MemoryUsage)
        }
    }
}

// checkResourceAvailability verifies if system can handle new jobs
func (js *JobScheduler) checkResourceAvailability() bool {
    return js.resources.CPUUsage < js.resources.MaxCPU &&
           js.resources.MemoryUsage < js.resources.MaxMemory
}

// AddJob enhanced with resource checking
func (js *JobScheduler) AddJob(job *Job) error {
    if job.TimeFrame <= 0 {
        return ErrInvalidTimeframe
    }

    js.mu.Lock()
    defer js.mu.Unlock()

    // Check system resources
    if !js.checkResourceAvailability() {
        // Calculate estimated time for resource availability
        estimatedTime := js.calculateEstimatedWaitTime()
        
        // Add to waiting queue
        js.waitingQueueMu.Lock()
        js.waitingQueue = append(js.waitingQueue, WaitingJob{
            Job:           job,
            EstimatedTime: estimatedTime,
        })
        js.waitingQueueMu.Unlock()

        log.Printf("System at capacity. Job %s added to waiting queue. Estimated start time: %v",
            job.JobID, estimatedTime)
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

// calculateEstimatedWaitTime estimates when resources might be available
func (js *JobScheduler) calculateEstimatedWaitTime() time.Time {
    // Find the job that will finish soonest
    var earliestCompletion time.Time
    now := time.Now()
    earliestCompletion = now.Add(30 * time.Second) // Default wait time

    js.mu.RLock()
    for _, job := range js.jobs {
        if job.Status == "processing" {
            expectedCompletion := job.CreatedAt.Add(time.Duration(job.TimeFrame) * time.Second)
            if earliestCompletion.After(expectedCompletion) {
                earliestCompletion = expectedCompletion
            }
        }
    }
    js.mu.RUnlock()

    return earliestCompletion
}

func (js *JobScheduler) SetResourceLimits(cpuThreshold, memoryThreshold float64) {
    js.mu.Lock()
    defer js.mu.Unlock()
    
    js.resources.MaxCPU = cpuThreshold
    js.resources.MaxMemory = memoryThreshold
}

// GetJobDetails returns detailed information about a specific job
func (js *JobScheduler) GetJobDetails(jobID string) (map[string]interface{}, error) {
    js.mu.RLock()
    defer js.mu.RUnlock()

    job, exists := js.jobs[jobID]
    if !exists {
        return nil, fmt.Errorf("job %s not found", jobID)
    }

    return map[string]interface{}{
        "job_id":            job.JobID,
        "status":            job.Status,
        "created_at":        job.CreatedAt,
        "last_executed":     job.LastExecuted,
        "current_retries":   job.CurrentRetries,
        "time_frame":        job.TimeFrame,
        "time_interval":     job.TimeInterval,
        "error":            job.Error,
    }, nil
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
            if js.checkResourceAvailability() {
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