package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/trigg3rX/triggerx-backend/execute/manager"
	"github.com/trigg3rX/triggerx-backend/pkg/api"
	"github.com/trigg3rX/triggerx-backend/pkg/database"
	"github.com/trigg3rX/triggerx-backend/pkg/events"
)

// toUint converts various types to uint
func toUint(v interface{}) uint {
	switch val := v.(type) {
	case uint:
		return val
	case int:
		return uint(val)
	case float64:
		return uint(val)
	case string:
		// Convert string to uint, returning 0 if conversion fails
		if uintVal, err := strconv.ParseUint(val, 10, 64); err == nil {
			return uint(uintVal)
		}
	default:
		return 0
	}
	return 0
}

func main() {
	log.SetOutput(os.Stdout)
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)

	// Initialize database connection
	dbConfig := &database.Config{
		Hosts:       []string{"localhost"}, // or your Cassandra host
		Timeout:     time.Second * 30,
		Retries:     3,
		ConnectWait: time.Second * 20,
	}

	db, err := database.NewConnection(dbConfig)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize the API server
	server := api.NewServer(db)

	// Initialize the job scheduler with 5 workers
	jobScheduler := manager.NewJobScheduler(5, db)
	jobScheduler.Cron.Start()
	defer jobScheduler.Stop()

	// Subscribe to job creation events
	server.GetEventBus().Subscribe(events.JobCreated, func(event events.Event) {
		jobEvent, ok := event.Payload.(events.JobCreatedEvent)
		if !ok {
			log.Printf("Error: Invalid event payload type")
			return
		}

		// Convert job ID to string and add to scheduler
		jobID := strconv.FormatInt(jobEvent.JobID, 10)
		if err := jobScheduler.AddJob(jobID); err != nil {
			log.Printf("Failed to add job %s: %v", jobID, err)
			return
		}

		log.Printf("Added new job %s to scheduler from event", jobID)
	})

	// Subscribe to job updated events and update the job in the scheduler
	server.GetEventBus().Subscribe(events.JobUpdated, func(event events.Event) {
		jobEvent, ok := event.Payload.(events.JobUpdatedEvent)
		if !ok {
			log.Printf("Error: Invalid event payload type")
			return
		}
		status := "inactive"
		if jobEvent.Status {
			status = "active"
		}
		jobScheduler.UpdateJob(jobEvent.JobID, status)
	})

	// Start the status monitoring goroutine
	statusTicker := time.NewTicker(10 * time.Second)
	defer statusTicker.Stop()

	go func() {
		for range statusTicker.C {
			queueStatus := jobScheduler.GetQueueStatus()
			systemMetrics := jobScheduler.GetSystemMetrics()

			log.Printf("System Status:")
			log.Printf("  Active Jobs: %d", queueStatus["active_jobs"])
			log.Printf("  Waiting Jobs: %d", queueStatus["waiting_jobs"])
			log.Printf("  CPU Usage: %.2f%%", systemMetrics.CPUUsage)
			log.Printf("  Memory Usage: %.2f%%", systemMetrics.MemoryUsage)
		}
	}()

	// API endpoints
	http.HandleFunc("/system/metrics", func(w http.ResponseWriter, r *http.Request) {
		metrics := jobScheduler.GetSystemMetrics()
		json.NewEncoder(w).Encode(metrics)
	})

	http.HandleFunc("/queue/status", func(w http.ResponseWriter, r *http.Request) {
		status := jobScheduler.GetQueueStatus()
		json.NewEncoder(w).Encode(status)
	})

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

	// Start the API server
	go func() {
		if err := server.Start("8082"); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Start HTTP server for manager endpoints
	managerAddr := ":8081"
	fmt.Printf("Manager server starting on %s\n", managerAddr)
	log.Fatal(http.ListenAndServe(managerAddr, nil))
}
