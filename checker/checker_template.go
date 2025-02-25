package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// CustomLogicFunc defines the signature for custom processing logic
type CustomLogicFunc func(*types.HandleCreateJobData, interface{}) (interface{}, error)

// IntervalChecker manages job execution timing and validation
type IntervalChecker struct{}

// NewIntervalChecker creates a new instance of IntervalChecker
func NewIntervalChecker() *IntervalChecker {
	return &IntervalChecker{}
}

func main() {
	// Create a new checker
	checker := NewIntervalChecker()

	// Create a dummy job for testing
	job := &types.HandleCreateJobData{
		JobID:        999,
		TimeInterval: 60,
		LastExecutedAt: time.Now().Add(-time.Minute),
		CreatedAt:    time.Now().Add(-time.Hour),
	}

	// Run the checker and get results
	success, payload := checker.Checker(job, nil)

	// Structure the output exactly as job_handler.go expects
	output := struct {
		Success bool                   `json:"success"`
		Payload map[string]interface{} `json:"payload"`
	}{
		Success: success,
		Payload: payload,
	}

	// Convert to JSON and write to stdout
	jsonData, err := json.Marshal(output)
	if err != nil {
		errorOutput := struct {
			Success bool   `json:"success"`
			Error   string `json:"error"`
		}{
			Success: false,
			Error:   err.Error(),
		}
		jsonData, _ = json.Marshal(errorOutput)
	}

	fmt.Print(string(jsonData))
}

// Checker is a comprehensive method that validates job execution and processes dynamic arguments
func (ic *IntervalChecker) Checker(job *types.HandleCreateJobData, customLogic CustomLogicFunc) (bool, map[string]interface{}) {
	payload := make(map[string]interface{})

	// Validate interval
	if !ic.ValidateJobInterval(job) {
		payload["error"] = "job interval validation failed"
		return false, payload
	}

	// Validate time frame
	if isValid, msg := ic.ValidateJobTimeFrame(job); !isValid {
		payload["error"] = msg
		return false, payload
	}

	return true, payload
}

// ValidateJobInterval checks if enough time has passed since last execution
func (ic *IntervalChecker) ValidateJobInterval(job *types.HandleCreateJobData) bool {
	if job.LastExecutedAt.IsZero() {
		return true
	}

	elapsed := time.Since(job.LastExecutedAt)
	return elapsed.Seconds() >= float64(job.TimeInterval)
}

// ValidateJobTimeFrame checks if the job is still within its allowed time frame
func (ic *IntervalChecker) ValidateJobTimeFrame(job *types.HandleCreateJobData) (bool, string) {
	if job.TimeFrame <= 0 {
		return true, ""
	}

	age := time.Since(job.CreatedAt)
	if age.Seconds() > float64(job.TimeFrame) {
		return false, fmt.Sprintf("job expired: age %v exceeds time frame %d seconds", age.Round(time.Second), job.TimeFrame)
	}

	return true, ""
}
