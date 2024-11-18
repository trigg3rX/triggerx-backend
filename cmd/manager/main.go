// cmd/main.go
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/trigg3rX/go-backend/internal/manager/manager" // Replace with your actual project path
)

// JobRequest represents the API request structure for creating a job
type JobRequest struct {
	ArgType           string                 `json:"arg_type"`
	Arguments         map[string]interface{} `json:"arguments"`
	ChainID           string                 `json:"chain_id"`
	ContractAddress   string                 `json:"contract_address"`
	JobCostPrediction float64                `json:"job_cost_prediction"`
	Stake             float64                `json:"stake"`
	TargetFunction    string                 `json:"target_function"`
	TimeFrame         int                    `json:"time_frame"`
	TimeInterval      int                    `json:"time_interval"`
	UserID            string                 `json:"user_id"`
}

var jobScheduler *scheduler.JobScheduler

func main() {
	// Initialize the job scheduler with 5 worker threads
	jobScheduler = scheduler.NewJobScheduler(5)
	
	// Start the cron scheduler
	jobScheduler.Cron.Start()
	defer jobScheduler.Stop()

	// API endpoints
	http.HandleFunc("/job/create", handleCreateJob)
	http.HandleFunc("/job/status", handleJobStatus)
	http.HandleFunc("/quorum/status", handleQuorumStatus)

	// Start HTTP server
	serverAddr := ":8080"
	fmt.Printf("Server starting on %s\n", serverAddr)
	log.Fatal(http.ListenAndServe(serverAddr, nil))
}

func handleCreateJob(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req JobRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Create a new job
	job := &scheduler.Job{
		JobID:             fmt.Sprintf("job_%d", time.Now().UnixNano()),
		ArgType:           req.ArgType,
		Arguments:         req.Arguments,
		ChainID:           req.ChainID,
		ContractAddress:   req.ContractAddress,
		JobCostPrediction: req.JobCostPrediction,
		Stake:             req.Stake,
		Status:            "pending",
		TargetFunction:    req.TargetFunction,
		TimeFrame:         req.TimeFrame,
		TimeInterval:      req.TimeInterval,
		UserID:            req.UserID,
		CreatedAt:         time.Now(),
	}

	// Add job to scheduler
	if err := jobScheduler.AddJob(job); err != nil {
		http.Error(w, fmt.Sprintf("Failed to schedule job: %v", err), http.StatusInternalServerError)
		return
	}

	// Return success response
	response := map[string]string{
		"job_id": job.JobID,
		"status": "scheduled",
	}
	json.NewEncoder(w).Encode(response)
}

func handleJobStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	jobID := r.URL.Query().Get("job_id")
	if jobID == "" {
		http.Error(w, "Job ID is required", http.StatusBadRequest)
		return
	}

	job, exists := jobScheduler.GetJob(jobID)
	if !exists {
		http.Error(w, "Job not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(job)
}

func handleQuorumStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	status := jobScheduler.GetQuorumsStatus()
	json.NewEncoder(w).Encode(status)
}