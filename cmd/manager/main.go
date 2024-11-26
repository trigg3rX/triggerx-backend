// github.com/trigg3rX/go-backend/cmd/manager/main.go
package main

import (
    "context"
    "encoding/json"
    "flag"
    "fmt"
    "log"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/libp2p/go-libp2p"
    "github.com/trigg3rX/go-backend/execute/manager"
    "github.com/trigg3rX/go-backend/pkg/network"
    "github.com/trigg3rX/go-backend/pkg/types"
)

func main() {
    // Command line flags
    schedulerName := flag.String("name", "scheduler", "Name for this scheduler node")
    listenAddr := flag.String("listen", "/ip4/0.0.0.0/tcp/9000", "Listen address for p2p connections")
    httpAddr := flag.String("http", ":8080", "HTTP API address")
    flag.Parse()

    // Create a context that will be canceled on interrupt
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    // Create libp2p host
    host, err := libp2p.New(
        libp2p.ListenAddrStrings(*listenAddr),
    )
    if err != nil {
        log.Fatalf("Failed to create libp2p host: %v", err)
    }
    defer host.Close()

    // Initialize network components
    discovery := network.NewDiscovery(ctx, host, *schedulerName)
    messaging := network.NewMessaging(host, *schedulerName)

    // Save scheduler's peer info
    if err := discovery.SavePeerInfo(); err != nil {
        log.Printf("Warning: Failed to save peer info: %v", err)
    }

    // Initialize the job scheduler with network components
    jobScheduler := manager.NewJobScheduler(5, messaging, discovery)
    jobScheduler.Cron.Start()
    defer jobScheduler.Stop()

    // Create example jobs
    createExampleJobs(jobScheduler)

    // Set up HTTP API
    setupHTTPAPI(jobScheduler, *httpAddr)

    // Log scheduler address
    log.Printf("Scheduler is listening on:")
    log.Printf("  P2P: %s/p2p/%s", host.Addrs()[0], host.ID())
    log.Printf("  HTTP API: %s", *httpAddr)

    // Wait for interrupt signal
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
    
    // Block until signal is received
    <-sigChan
    fmt.Println("\nReceived interrupt signal, shutting down...")
}

func createExampleJobs(jobScheduler *manager.JobScheduler) {
    for i := 1; i <= 5; i++ {
        job := &types.Job{
            JobID:             fmt.Sprintf("job_%d", i),
            ArgType:           "contract_call",
            Arguments:         map[string]interface{}{"function": "transfer", "amount": 100 * i},
            ChainID:           "chain_1",
            ContractAddress:   fmt.Sprintf("0x123abc%d", i),
            JobCostPrediction: 0.5,
            Stake:            1.0,
            Status:           "pending",
            TargetFunction:   "execute",
            TimeFrame:        60,
            TimeInterval:     10,
            UserID:           fmt.Sprintf("user_%d", i),
            CreatedAt:        time.Now(),
            MaxRetries:       3,
        }

        if err := jobScheduler.AddJob(job); err != nil {
            log.Printf("Failed to add job %s: %v", job.JobID, err)
        }
    }
}

func setupHTTPAPI(jobScheduler *manager.JobScheduler, addr string) {
    // System metrics endpoint
    http.HandleFunc("/system/metrics", func(w http.ResponseWriter, r *http.Request) {
        metrics := jobScheduler.GetSystemMetrics()
        json.NewEncoder(w).Encode(metrics)
    })

    // Queue status endpoint
    http.HandleFunc("/queue/status", func(w http.ResponseWriter, r *http.Request) {
        status := jobScheduler.GetQueueStatus()
        json.NewEncoder(w).Encode(status)
    })

    // Job details endpoint
    http.HandleFunc("/job/", func(w http.ResponseWriter, r *http.Request) {
        jobID := r.URL.Path[len("/job/"):]
        if jobID == "" {
            http.Error(w, "Job ID required", http.StatusBadRequest)
            return
        }

        details, err := jobScheduler.GetJobDetails(jobID)
        if err != nil {
            http.Error(w, err.Error(), http.StatusNotFound)
            return
        }

        json.NewEncoder(w).Encode(details)
    })

    // Start HTTP server in a goroutine
    go func() {
        log.Printf("Starting HTTP server on %s", addr)
        if err := http.ListenAndServe(addr, nil); err != nil {
            log.Fatalf("HTTP server failed: %v", err)
        }
    }()
}