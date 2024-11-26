package main

import (
	"log"
    "fmt"
    "time"
    "os"
    "os/signal"
    "syscall"

	"github.com/trigg3rX/go-backend/pkg/network"
	"github.com/trigg3rX/go-backend/execute/manager"
	"github.com/trigg3rX/triggerx-keeper/pkg/execution"
	"github.com/libp2p/go-libp2p"
	"context"
	"github.com/trigg3rX/go-backend/pkg/types"
)

const (
    WORKERS_COUNT     = 5
    KEEPER_COUNT     = 3
)

func main() {
    // Create libp2p host
    host, err := libp2p.New(
        libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/9000"),
    )
    if err != nil {
        log.Fatalf("Failed to create libp2p host: %v", err)
    }
    defer host.Close()

    // Initialize networking components
    messaging := network.NewMessaging(host, "keeper")
    discovery := network.NewDiscovery(context.Background(), host, "keeper")

    // Initialize the job scheduler
    scheduler := manager.NewJobScheduler(WORKERS_COUNT, messaging, discovery)
    scheduler.SetResourceLimits(80.0, 80.0) // Set CPU and Memory thresholds to 80%

    // Initialize keepers
    keepers := make([]*execution.Keeper, KEEPER_COUNT)
    for i := 0; i < KEEPER_COUNT; i++ {
        keeperName := fmt.Sprintf("keeper-%d", i+1)
        keepers[i] = execution.NewKeeper(keeperName, messaging, host.ID())
        keepers[i].Start()
    }

    // Example: Schedule a test job
    testJob := &types.Job{
        JobID:           "test-job-1",
        ArgType:         "string",
        Arguments:       map[string]interface{}{"param1": "value1"},
        ChainID:         "chain_1",
        ContractAddress: "0x123...",
        TimeFrame:       3600,    // 1 hour
        TimeInterval:    60,      // Execute every 60 seconds
        UserID:          "user1",
        CreatedAt:       time.Now(),
        MaxRetries:      3,
        Status:          "pending",
    }

    err = scheduler.AddJob(testJob)
    if err != nil {
        log.Printf("Failed to add test job: %v", err)
    }

    // Set up graceful shutdown
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

    // Start monitoring system metrics
    go monitorSystem(scheduler)

    // Wait for shutdown signal
    <-sigChan
    log.Println("Shutting down...")
    
    // Cleanup
    scheduler.Stop()
}

// monitorSystem periodically logs system status
func monitorSystem(scheduler *manager.JobScheduler) {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            metrics := scheduler.GetSystemMetrics()
            queueStatus := scheduler.GetQueueStatus()
            
            log.Printf("System Status:")
            log.Printf("  CPU Usage: %.2f%%", metrics.CPUUsage)
            log.Printf("  Memory Usage: %.2f%%", metrics.MemoryUsage)
            log.Printf("  Active Jobs: %d", queueStatus["active_jobs"])
            log.Printf("  Waiting Jobs: %d", queueStatus["waiting_jobs"])
        }
    }
}