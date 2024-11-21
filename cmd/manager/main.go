package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/trigg3rX/go-backend/internal/manager"
)

// APIJobDetails represents the structure of the job details from the external API
type APIJobDetails struct {
	JobID              int      `json:"job_id"`
	JobType            int      `json:"jobType"`
	UserID             int      `json:"user_id"`
	ChainID            int      `json:"chain_id"`
	TimeFrame          int      `json:"time_frame"`
	TimeInterval       int      `json:"time_interval"`
	ContractAddress    string   `json:"contract_address"`
	TargetFunction     string   `json:"target_function"`
	ArgType            int      `json:"arg_type"`
	Arguments          []string `json:"arguments"`
	Status             bool     `json:"status"`
	JobCostPrediction  float64  `json:"job_cost_prediction"`
}

// fetchJobDetailsFromAPI retrieves job details from an external API
func fetchJobDetailsFromAPI(url string) (*APIJobDetails, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Add("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %v", err)
	}

	var jobDetails APIJobDetails
	err = json.Unmarshal(body, &jobDetails)
	if err != nil {
		return nil, fmt.Errorf("error parsing JSON: %v", err)
	}

	return &jobDetails, nil
}

// convertAPIJobToSchedulerJob converts external API job to internal scheduler job
func convertAPIJobToSchedulerJob(apiJob *APIJobDetails) *manager.Job {
    // Convert arguments to map[string]interface{}
    arguments := make(map[string]interface{})
    for i, arg := range apiJob.Arguments {
        arguments[fmt.Sprintf("arg_%d", i)] = arg
    }

    return &manager.Job{
        JobID:             fmt.Sprintf("job_%d", apiJob.JobID),
        ArgType:           fmt.Sprintf("%d", apiJob.ArgType),
        Arguments:         arguments,
        ChainID:           fmt.Sprintf("%d", apiJob.ChainID),
        ContractAddress:   apiJob.ContractAddress,
        JobCostPrediction: apiJob.JobCostPrediction,
        Stake:             0.0, // Set a default stake or derive from API if needed
        Status:            "pending",
        TargetFunction:    apiJob.TargetFunction,
        TimeFrame:         apiJob.TimeFrame,
        TimeInterval:      apiJob.TimeInterval,
        UserID:            fmt.Sprintf("%d", apiJob.UserID),
        CreatedAt:         time.Now(),
        MaxRetries:        3, // Default max retries
    }
}
func main() {
	// Initialize the job scheduler with 5 worker threads
	jobScheduler := manager.NewJobScheduler(5)
	
	// Start the cron scheduler
	jobScheduler.Cron.Start()
    defer jobScheduler.Stop()

	// API URL for fetching job details
	apiURL := "http://192.168.1.53:8080/api/jobs/2"

	// Fetch job details from external API
	apiJobDetails, err := fetchJobDetailsFromAPI(apiURL)
	if err != nil {
		log.Fatalf("Failed to fetch job details: %v", err)
	}

	// Convert API job to scheduler job
	schedulerJob := convertAPIJobToSchedulerJob(apiJobDetails)

	// Add job to scheduler
	if err := jobScheduler.AddJob(schedulerJob); err != nil {
		log.Fatalf("Failed to add job to scheduler: %v", err)
	}

	log.Printf("Job %s added to scheduler successfully", schedulerJob.JobID)

	// HTTP server setup (similar to previous example)
	http.HandleFunc("/job/create", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Decode incoming job request
		var req manager.Job
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Add job to scheduler
		if err := jobScheduler.AddJob(&req); err != nil {
			http.Error(w, fmt.Sprintf("Failed to schedule job: %v", err), http.StatusInternalServerError)
			return
		}

		// Return success response
		response := map[string]string{
			"job_id": req.JobID,
			"status": "scheduled",
		}
		json.NewEncoder(w).Encode(response)
	})

	http.HandleFunc("/job/status", func(w http.ResponseWriter, r *http.Request) {
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
	})

	http.HandleFunc("/quorum/status", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		status := jobScheduler.GetQuorumsStatus()
		json.NewEncoder(w).Encode(status)
	})

	// Start HTTP server
	serverAddr := ":8080"
	fmt.Printf("Server starting on %s\n", serverAddr)
	log.Fatal(http.ListenAndServe(serverAddr, nil))
}