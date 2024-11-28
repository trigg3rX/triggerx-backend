package main

import (
    "context"
    "log"
    "os"
    "os/signal"
    "syscall"

    "github.com/libp2p/go-libp2p"
    "github.com/trigg3rX/go-backend/pkg/network" 
    "github.com/trigg3rX/triggerx-keeper/execute/handler"
    "github.com/trigg3rX/go-backend/execute/manager" // Adjust import path
)

func main() {
    // Setup logging
    log.SetOutput(os.Stdout)
    log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)

    // Create context
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    // Create libp2p host
    host, err := libp2p.New()
    if err != nil {
        log.Fatalf("Failed to create libp2p host: %v", err)
    }
    defer host.Close()

    // Create network messaging
    keeperName := "keeper-node-1"
    messaging := network.NewMessaging(host, keeperName)

    // Create job handler
    jobHandler := handler.NewJobHandler()

    // Setup message handling
    messaging.InitMessageHandling(func(msg network.Message) {
        // Type assertion and job handling
        if msg.Type != "JOB_TRANSMISSION" {
            log.Printf("Received non-job message: %s", msg.Type)
            return
        }

        job, ok := msg.Content.(map[string]interface{})
        if !ok {
            log.Printf("Invalid job content type")
            return
        }

        // Convert map to Job struct
        jobData, err := convertMapToJob(job)
        if err != nil {
            log.Printf("Job conversion error: %v", err)
            return
        }

        // Handle job
        if err := jobHandler.HandleJob(jobData); err != nil {
            log.Printf("Job handling error: %v", err)
        }
    })

    // Save peer info for discovery
    discovery := network.NewDiscovery(ctx, host, keeperName)
    if err := discovery.SavePeerInfo(); err != nil {
        log.Printf("Failed to save peer info: %v", err)
    }

    // Wait for interrupt
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
    
    log.Printf("🚀 Keeper node started. Listening on %s", host.Addrs())
    <-sigChan
}

// Helper function to convert map to Job struct
func convertMapToJob(jobMap map[string]interface{}) (*manager.Job, error) {
    job := &manager.Job{
        JobID:           toString(jobMap["JobID"]),
        ArgType:         toString(jobMap["ArgType"]),
        ChainID:         toString(jobMap["ChainID"]),
        ContractAddress: toString(jobMap["ContractAddress"]),
        TargetFunction:  toString(jobMap["TargetFunction"]),
        Status:          toString(jobMap["Status"]),
    }

    // Convert arguments
    if args, ok := jobMap["Arguments"].(map[string]interface{}); ok {
        job.Arguments = args
    }

    return job, nil
}

func toString(v interface{}) string {
    if s, ok := v.(string); ok {
        return s
    }
    return ""
}