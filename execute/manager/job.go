// job.go
package manager

import (
    "log"
    "math/rand"
    "time"
	"os"
	"encoding/json"
	
	"github.com/trigg3rX/go-backend/pkg/network"
)

type JobMessage struct {
    Job       *Job   `json:"job"`
    Timestamp string `json:"timestamp"`
}

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
    
    // Get a random keeper from the quorum
    quorum := js.quorums["default"]
    if len(quorum.ActiveNodes) == 0 {
        job.Status = "failed"
        job.Error = "no active keepers available"
        js.mu.Unlock()
        return
    }
    
    // Select random keeper
    keeperName := quorum.ActiveNodes[rand.Intn(len(quorum.ActiveNodes))]
    js.mu.Unlock()

    // Load keeper information
    peerInfos := make(map[string]network.PeerInfo)
    if err := js.loadPeerInfo(&peerInfos); err != nil {
        log.Printf("Failed to load peer info: %v", err)
        return
    }

    keeperInfo, exists := peerInfos[keeperName]
    if !exists {
        log.Printf("Keeper %s not found in peer info", keeperName)
        return
    }

    // Connect to the keeper
    peerID, err := js.discovery.ConnectToPeer(keeperInfo)
    if err != nil {
        log.Printf("Failed to connect to keeper %s: %v", keeperName, err)
        return
    }

    // Prepare job message
    jobMsg := JobMessage{
        Job:       job,
        Timestamp: time.Now().UTC().Format(time.RFC3339),
    }

    // Send job to keeper
    err = js.messaging.SendMessage(keeperName, *peerID, jobMsg)
    if err != nil {
        log.Printf("Failed to send job to keeper %s: %v", keeperName, err)
        js.mu.Lock()
        job.Status = "failed"
        job.Error = err.Error()
        js.mu.Unlock()
        return
    }

    log.Printf("Job %s sent to keeper %s", job.JobID, keeperName)
}

func (js *JobScheduler) loadPeerInfo(peerInfos *map[string]network.PeerInfo) error {
    file, err := os.Open(network.PeerInfoFilePath)
    if err != nil {
        return err
    }
    defer file.Close()

    decoder := json.NewDecoder(file)
    return decoder.Decode(peerInfos)
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