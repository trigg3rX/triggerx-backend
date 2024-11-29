// github.com/trigg3rX/go-backend/execute/manager/job.go
package manager

import (
    "log"
    "math/rand"
    "time"
    "fmt"
)

// Job represents a scheduled task with its properties
type Job struct {
    JobID             string
    ArgType           string
    Arguments         map[string]interface{}
    ChainID           string
    ContractAddress   string
    JobCostPrediction float64
    Stake             float64
    Status            string
    TargetFunction    string
    TimeFrame         int64  // in seconds
    TimeInterval      int64  // in seconds
    UserID            string
    CreatedAt         time.Time
    MaxRetries        int
    CurrentRetries    int
    LastExecuted      time.Time
    NextExecutionTime time.Time
    Error            string
}

// Quorum represents a group of nodes that can execute jobs
type Quorum struct {
    QuorumID    string
    NodeCount   int
    ActiveNodes []string
    Status      string
    ChainID     string
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

func init() {
    // Initialize random seed
    rand.Seed(time.Now().UnixNano())
}

// initializeQuorums sets up initial quorums for the scheduler
func (js *JobScheduler) initializeQuorums() {
    defaultQuorum := &Quorum{
        QuorumID:    "default",
        NodeCount:   3,
        ActiveNodes: []string{"node1", "node2", "node3"},
        Status:      "active",
        ChainID:     "chain_1",
        CreatedAt:   time.Now(),
        UpdatedAt:   time.Now(),
    }

    js.mu.Lock()
    js.quorums["default"] = defaultQuorum
    js.mu.Unlock()
}

func (js *JobScheduler) selectRandomKeeper() (string, error) {
    // Acquire a read lock to safely access quorums
    js.mu.RLock()
    defer js.mu.RUnlock()

    // Check if any quorums exist
    if len(js.quorums) == 0 {
        return "", fmt.Errorf("no quorums available")
    }

    // Iterate through available quorums
    for _, quorum := range js.quorums {
        // Ensure the quorum has active nodes
        if len(quorum.ActiveNodes) > 0 {
            // Randomly select a keeper from the active nodes in this quorum
            //randomIndex := rand.Intn(len(quorum.ActiveNodes))
            return quorum.ActiveNodes[0], nil
        }
    }

    // If no active nodes are found in any quorum
    return "", fmt.Errorf("no active keepers found")
}
// processJob handles the execution of a job
func (js *JobScheduler) processJob(workerID int, job *Job) {
    js.mu.Lock()
    if job.Status == "completed" || job.Status == "failed" {
        js.mu.Unlock()
        return
    }

    job.Status = "processing"
    job.LastExecuted = time.Now()
    js.mu.Unlock()

    

    // Enhanced logging with worker and job details
    log.Printf("[Worker %d] Starting to process Job %s (Target: %s, ChainID: %s)", 
        workerID, job.JobID, job.TargetFunction, job.ChainID)

        selectedKeeper, err := js.selectRandomKeeper()
        if err != nil {
            log.Printf("Failed to select keeper for job %s: %v", job.JobID, err)
            // Handle failure - maybe retry or mark job as failed
            return
        }

        err = js.transmitJobToKeeper(selectedKeeper, job)
    if err != nil {
        log.Printf("Job transmission failed: %v", err)
    }

    // Simulate job execution with random success/failure
    executionTime := time.Duration(2+rand.Intn(3)) * time.Second
    log.Printf("[Worker %d] ⏳ Job %s will take approximately %v to complete", 
        workerID, job.JobID, executionTime)
    time.Sleep(executionTime)

    js.mu.Lock()
    defer js.mu.Unlock()

    if rand.Float64() < 0.8 { // 80% success rate
        job.Status = "completed"
        log.Printf("[Worker %d] Successfully completed Job %s", workerID, job.JobID)
    } else {
        job.CurrentRetries++
        if job.CurrentRetries >= job.MaxRetries {
            job.Status = "failed"
            job.Error = "maximum retries exceeded"
            log.Printf("[Worker %d] Job %s failed after %d retries. Error: %s", 
                workerID, job.JobID, job.MaxRetries, job.Error)
        } else {
            job.Status = "pending"
            log.Printf("[Worker %d] Job %s failed, scheduling retry (%d/%d)", 
                workerID, job.JobID, job.CurrentRetries, job.MaxRetries)
        }
    }
    
}


// GetSystemMetrics returns current system metrics
func (js *JobScheduler) GetSystemMetrics() SystemResources {
    js.mu.RLock()
    defer js.mu.RUnlock()
    return js.resources
}

// GetQueueStatus returns the current status of job queues
func (js *JobScheduler) GetQueueStatus() map[string]interface{} {
    js.mu.RLock()
    js.waitingQueueMu.RLock()
    defer js.mu.RUnlock()
    defer js.waitingQueueMu.RUnlock()

    return map[string]interface{}{
        "active_jobs":     len(js.jobs),
        "waiting_jobs":    len(js.waitingQueue),
        "cpu_usage":       js.resources.CPUUsage,
        "memory_usage":    js.resources.MemoryUsage,
    }
}

// Stop gracefully shuts down the scheduler
func (js *JobScheduler) Stop() {
    js.cancel()
    js.Cron.Stop()
}

// startWorkers initializes worker goroutines
func (js *JobScheduler) startWorkers() {
    for i := 0; i < js.workersCount; i++ {
        workerID := i
        go func(workerID int) {
            log.Printf("🔧 Worker %d initialized and ready to process jobs", workerID)
            for {
                select {
                case job, ok := <-js.jobQueue:
                    if !ok {
                        log.Printf("Worker %d: Job queue closed", workerID)
                        return
                    }
                    if job == nil {
                        log.Printf("Worker %d: Received nil job", workerID)
                        continue
                    }
                    js.processJob(workerID, job)
                case <-js.ctx.Done():
                    log.Printf("Worker %d: Context cancelled, shutting down", workerID)
                    return
                }
            }
        }(workerID)
    }
}