// github.com/trigg3rX/triggerx-keeper/pkg/execution/execution.go
package execution

import (
    "context"
    "encoding/json"
    "log"
    "sync"
    "time"

    "github.com/libp2p/go-libp2p/core/peer"
    "github.com/trigg3rX/go-backend/pkg/network"
    "github.com/trigg3rX/go-backend/pkg/types"
)

// MessageType for internal message handling
type MessageType string

const (
    ExecuteJob MessageType = "EXECUTE_JOB"
    CancelJob  MessageType = "CANCEL_JOB"
    Heartbeat  MessageType = "HEARTBEAT"
)

// KeeperMessage wraps the message type with the job message
type KeeperMessage struct {
    Type    MessageType     `json:"type"`
    Message types.JobMessage `json:"message"`
}

// Keeper represents a node that executes jobs
type Keeper struct {
    name       string
    messaging  *network.Messaging
    jobs       map[string]*types.Job
    jobsMutex  sync.RWMutex
    ctx        context.Context
    cancel     context.CancelFunc
    managerID  peer.ID
}

// NewKeeper creates a new keeper instance
func NewKeeper(name string, messaging *network.Messaging, managerID peer.ID) *Keeper {
    ctx, cancel := context.WithCancel(context.Background())
    return &Keeper{
        name:      name,
        messaging: messaging,
        jobs:      make(map[string]*types.Job),
        ctx:       ctx,
        cancel:    cancel,
        managerID: managerID,
    }
}

// Start begins the keeper's operation
func (k *Keeper) Start() error {
    k.messaging.InitMessageHandling(k.handleMessage)
    go k.sendHeartbeat()
    return nil
}

// Stop gracefully shuts down the keeper
func (k *Keeper) Stop() {
    k.cancel()
}

func (k *Keeper) handleMessage(msg network.Message) {
    var keeperMsg KeeperMessage
    content, ok := msg.Content.(map[string]interface{})
    if !ok {
        log.Printf("Error: message content is not a map")
        return
    }

    // Convert map to JSON bytes for unmarshaling
    jsonBytes, err := json.Marshal(content)
    if err != nil {
        log.Printf("Error marshaling content: %v", err)
        return
    }

    if err := json.Unmarshal(jsonBytes, &keeperMsg); err != nil {
        log.Printf("Error unmarshaling keeper message: %v", err)
        return
    }

    switch keeperMsg.Type {
    case ExecuteJob:
        if keeperMsg.Message.Job != nil {
            k.executeJob(keeperMsg.Message.Job)
        }
    case CancelJob:
        if keeperMsg.Message.Job != nil {
            k.cancelJob(keeperMsg.Message.Job.JobID)
        }
    default:
        log.Printf("Unknown message type: %s", keeperMsg.Type)
    }
}

func (k *Keeper) executeJob(job *types.Job) {
    k.jobsMutex.Lock()
    k.jobs[job.JobID] = job
    k.jobsMutex.Unlock()

    log.Printf("Executing job %s on contract %s", job.JobID, job.ContractAddress)
    
    // Placeholder for actual blockchain interaction
    // This is where you'd implement the actual job execution logic
    
    job.Status = "completed"
    job.LastExecuted = time.Now()
    job.NextExecutionTime = time.Now().Add(time.Duration(job.TimeInterval) * time.Second)

    // Prepare result message
    resultMsg := types.JobMessage{
        Job: job,
        Timestamp: time.Now().UTC().Format(time.RFC3339),
    }
    
    // Send result back to manager
    k.messaging.SendMessage("manager", k.managerID, resultMsg)
}

func (k *Keeper) cancelJob(jobID string) {
    k.jobsMutex.Lock()
    if job, exists := k.jobs[jobID]; exists {
        job.Status = "cancelled"
        delete(k.jobs, jobID)
    }
    k.jobsMutex.Unlock()
    log.Printf("Cancelled job %s", jobID)
}

func (k *Keeper) sendHeartbeat() {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-k.ctx.Done():
            return
        case <-ticker.C:
            heartbeat := network.Message{
                Type: string(Heartbeat),
                From: k.name,
                Content: map[string]interface{}{
                    "timestamp": time.Now().UTC().Format(time.RFC3339),
                    "status":    "active",
                    "jobs_count": len(k.jobs),
                },
            }
            k.messaging.SendMessage("manager", k.managerID, heartbeat)
        }
    }
}