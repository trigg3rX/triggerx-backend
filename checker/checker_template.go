package checker

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/trigg3rX/go-backend/execute/manager"
)

// CustomLogicFunc defines the signature for custom processing logic
type CustomLogicFunc func(*manager.Job, interface{}) (interface{}, error)

// IntervalChecker manages job execution timing and validation
type IntervalChecker struct{}

// NewIntervalChecker creates a new instance of IntervalChecker
func NewIntervalChecker() *IntervalChecker {
	return &IntervalChecker{}
}

// PriceResponse represents the structure of the price API response
type PriceResponse struct {
	Data struct {
		Value int64 `json:"value"`
	} `json:"data"`
}

// Checker is a comprehensive method that validates job execution and processes dynamic arguments
func (ic *IntervalChecker) Checker(job *manager.Job, customLogic CustomLogicFunc) (bool, map[string]interface{}) {
	// Initialize payload to store results
	payload := make(map[string]interface{})

	// Check job interval
	if !ic.ValidateJobInterval(job) {
		log.Printf("Job %s not ready for execution", job.JobID)
		return false, payload
	}

	// Check job time frame
	if isValid, reason := ic.validateJobTimeFrame(job); !isValid {
		log.Printf("Job %s time frame validation failed: %s", job.JobID, reason)
		return false, payload
	}

	// Fetch price from the API
	price, err := ic.fetchPrice()
	if err != nil {
		log.Printf("Failed to fetch price for job %s: %v", job.JobID, err)
		return false, payload
	}
	payload["price"] = price

	// Execute custom logic if provided
	if customLogic != nil {
		result, err := customLogic(job, price)
		if err != nil {
			log.Printf("Custom logic failed for job %s: %v", job.JobID, err)
			return false, payload
		}
		payload["customResult"] = result
	}

	// Log successful processing
	log.Printf("Job %s processed successfully", job.JobID)
	return true, payload
}

// ValidateJobInterval checks if a job is ready to be executed based on its time interval
func (ic *IntervalChecker) ValidateJobInterval(job *manager.Job) bool {
	if job.LastExecuted.IsZero() {
		return true
	}
	timeSinceLastExecution := time.Since(job.LastExecuted)
	if timeSinceLastExecution.Seconds() < float64(job.TimeInterval) {
		log.Printf("Job %s not ready. Time since last execution: %.2f seconds (required: %d seconds)",
			job.JobID, timeSinceLastExecution.Seconds(), job.TimeInterval)
		return false
	}
	return true
}

// validateJobTimeFrame checks if the job is within its allowed execution time frame
func (ic *IntervalChecker) validateJobTimeFrame(job *manager.Job) (bool, string) {
	if job.TimeFrame <= 0 {
		return true, ""
	}
	jobAge := time.Since(job.CreatedAt)
	if jobAge.Seconds() > float64(job.TimeFrame) {
		reason := fmt.Sprintf("job exceeded time frame: created %v ago, max allowed %d seconds",
			jobAge.Round(time.Second), job.TimeFrame)
		return false, reason
	}
	return true, ""
}

// fetchPrice retrieves the current price from the specified API endpoint
func (ic *IntervalChecker) fetchPrice() (int64, error) {
	apiEndpoint := "http://localhost:3005/get-price"
	resp, err := http.Get(apiEndpoint)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch price: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("received non-OK HTTP status: %s", resp.Status)
	}

	var priceResp PriceResponse
	if err := json.NewDecoder(resp.Body).Decode(&priceResp); err != nil {
		return 0, fmt.Errorf("failed to decode price response: %w", err)
	}

	return priceResp.Data.Value, nil
}

// CalculateNextExecutionTime determines the next time the job should be executed
func (ic *IntervalChecker) CalculateNextExecutionTime(job *manager.Job) time.Time {
	if job.LastExecuted.IsZero() {
		return time.Now()
	}
	return job.LastExecuted.Add(time.Duration(job.TimeInterval) * time.Second)
}