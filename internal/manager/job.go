// job.go
package manager

import (
    "log"
    "math/rand"
    "time"
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

// processJob handles the execution of a job
func (js *JobScheduler) processJob(job *Job) {
    js.mu.Lock()
    if job.Status == "completed" || job.Status == "failed" {
        js.mu.Unlock()
        return
    }
    
    job.Status = "processing"
    job.LastExecuted = time.Now()
    js.mu.Unlock()

    // Simulate job execution with random success/failure
    time.Sleep(time.Duration(2+rand.Intn(3)) * time.Second)
    
    js.mu.Lock()
    defer js.mu.Unlock()

    if rand.Float64() < 0.8 { // 80% success rate
        job.Status = "completed"
        log.Printf("Job %s completed successfully", job.JobID)
    } else {
        job.CurrentRetries++
        if job.CurrentRetries >= job.MaxRetries {
            job.Status = "failed"
            job.Error = "maximum retries exceeded"
            log.Printf("Job %s failed after %d retries", job.JobID, job.MaxRetries)
        } else {
            job.Status = "pending"
            log.Printf("Job %s failed, will retry (%d/%d)", job.JobID, job.CurrentRetries, job.MaxRetries)
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
        go func(workerID int) {
            for {
                select {
                case job := <-js.jobQueue:
                    if job == nil {
                        return
                    }
                    js.processJob(job)
                case <-js.ctx.Done():
                    return
                }
            }
        }(i)
    }
}