package main

import (
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "time"
    "github.com/trigg3rX/go-backend/internal/manager"
)

func main() {
    // Initialize the job scheduler with 5 workers
    jobScheduler := manager.NewJobScheduler(5)
    jobScheduler.Cron.Start()
    defer jobScheduler.Stop()

    // Create multiple test jobs
    for i := 1; i <= 10; i++ {
        job := &manager.Job{
            JobID:             fmt.Sprintf("job_%d", i),
            ArgType:           "contract_call",
            Arguments:         map[string]interface{}{"function": "transfer"},
            ChainID:           "chain_1",
            ContractAddress:   fmt.Sprintf("0x123...%d", i),
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
        } else {
            log.Printf("Added job %s to scheduler", job.JobID)
        }

        // Small delay between job additions
        time.Sleep(2 * time.Second)
    }

    // API endpoints
    http.HandleFunc("/system/metrics", func(w http.ResponseWriter, r *http.Request) {
        metrics := jobScheduler.GetSystemMetrics()
        json.NewEncoder(w).Encode(metrics)
    })

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

    // Start HTTP server
    serverAddr := ":8080"
    fmt.Printf("Server starting on %s\n", serverAddr)
    log.Fatal(http.ListenAndServe(serverAddr, nil))
}
