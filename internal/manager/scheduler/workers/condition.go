package workers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/trigg3rX/triggerx-backend/internal/manager/scheduler/services"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

type ConditionBasedWorker struct {
	jobID     int64
	scheduler JobScheduler
	jobData   *types.HandleCreateJobData
	ticker    *time.Ticker
	done      chan bool
	BaseWorker
}

func NewConditionBasedWorker(jobData types.HandleCreateJobData, scheduler JobScheduler) *ConditionBasedWorker {
	return &ConditionBasedWorker{
		jobID:     jobData.JobID,
		scheduler: scheduler,
		jobData:   &jobData,
		done:      make(chan bool),
		BaseWorker: BaseWorker{
			status:     "pending",
			maxRetries: 3,
		},
	}
}

func (w *ConditionBasedWorker) Start(ctx context.Context) {
	w.status = "running"
	w.ticker = time.NewTicker(1 * time.Second)
	w.done = make(chan bool) // Ensure we have a fresh channel

	w.scheduler.Logger().Infof("Starting condition-based job %d", w.jobID)
	w.scheduler.Logger().Infof("Listening to %s", w.jobData.ScriptTriggerFunction)

	var triggerData types.TriggerData
	triggerData.TimeInterval = w.jobData.TimeInterval
	triggerData.LastExecuted = time.Now().UTC()
	triggerData.TriggerTxHash = ""
	triggerData.ConditionParams = make(map[string]interface{})

	// Calculate end time if timeframe is specified
	var endTime time.Time
	if w.jobData.TimeFrame > 0 {
		endTime = time.Now().UTC().Add(time.Duration(w.jobData.TimeFrame) * time.Second)
	}

	go func() {
		// Use a flag to prevent multiple Stop() calls
		var stopped bool
		defer func() {
			if !stopped {
				stopped = true
				w.Stop()
			}
		}()

		for {
			select {
			case <-ctx.Done():
				return

			case <-w.done:
				return

			case <-w.ticker.C:
				// Check if we've exceeded the timeframe
				if w.jobData.TimeFrame > 0 && time.Now().UTC().After(endTime) {
					w.scheduler.Logger().Infof("Timeframe reached for job %d, stopping worker", w.jobID)
					return
				}

				satisfied, conditionParams, err := w.checkCondition()
				if err != nil {
					w.error = err.Error()
					w.currentRetry++

					if w.currentRetry >= w.maxRetries {
						w.status = "failed"
						return
					}
					continue
				}

				if satisfied {
					w.scheduler.Logger().Infof("Condition satisfied for job %d", w.jobID)

					triggerData.Timestamp = time.Now().UTC()

					// Update condition parameters with the latest values
					if conditionParams != nil {
						triggerData.ConditionParams = conditionParams
					}

					if err := w.executeTask(w.jobData, &triggerData); err != nil {
						w.handleError(err)
						// If we're in recurring mode and hit an error, we continue checking
						if !w.jobData.Recurring || w.status == "failed" {
							return
						}
						continue
					}

					// If it's not recurring, exit after first execution
					if !w.jobData.Recurring {
						return
					}

					// Update last execution time
					triggerData.LastExecuted = time.Now().UTC()
					w.scheduler.Logger().Infof("Job %d executed. Continuing to monitor condition due to recurring flag", w.jobID)

					// Optional: add a small pause after execution before checking again
					time.Sleep(5 * time.Second)
				}
			}
		}
	}()
}

func (w *ConditionBasedWorker) Stop() {
	if w.ticker != nil {
		w.ticker.Stop()
		w.ticker = nil
	}

	// Only close the done channel if it's not nil and not already closed
	if w.done != nil {
		// Use a recover to catch panic if channel is already closed
		defer func() {
			if r := recover(); r != nil {
				w.scheduler.Logger().Warnf("Attempted to close already closed channel for job %d: %v", w.jobID, r)
			}
		}()

		select {
		case <-w.done:
			// Channel is already closed, do nothing
		default:
			close(w.done)
		}
	}

	w.scheduler.RemoveJob(w.jobID)
}

func (w *ConditionBasedWorker) handleError(err error) {
	w.error = err.Error()
	w.currentRetry++

	if w.currentRetry >= w.maxRetries {
		w.status = "failed"
	}
}

func (w *ConditionBasedWorker) checkCondition() (bool, map[string]interface{}, error) {
	resp, err := http.Get(w.jobData.ScriptTriggerFunction)
	if err != nil {
		return false, nil, fmt.Errorf("failed to fetch API data: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, nil, fmt.Errorf("failed to read response body: %v", err)
	}

	w.scheduler.Logger().Infof("API response: %s", string(body))

	// Parse the response to determine if the condition is satisfied
	// The IPFS script will print multiple lines, we need to extract the condition satisfaction
	responseStr := string(body)

	// Check if the response contains "Condition satisfied: true"
	// This is a simple approach - in a production environment, we might want to parse this more robustly
	if responseStr == "" {
		return false, nil, fmt.Errorf("empty response from condition script")
	}

	// Check for "Condition satisfied: true" in the response
	satisfied := false
	var conditionParams map[string]interface{}

	// For a more robust approach, we could try to parse the actual structured data
	// For example, if the response is JSON, we could unmarshal it
	type ConditionResult struct {
		Satisfied bool      `json:"Satisfied"`
		Timestamp time.Time `json:"Timestamp"`
		Response  float64   `json:"Response"`
	}

	// Try to unmarshal if it looks like JSON
	if len(responseStr) > 0 && (responseStr[0] == '{' || responseStr[0] == '[') {
		var result ConditionResult
		if err := json.Unmarshal(body, &result); err == nil {
			w.scheduler.Logger().Infof("Parsed condition result: satisfied=%v, response=%v", result.Satisfied, result.Response)

			// Store only the three requested condition parameters
			conditionParams = map[string]interface{}{
				"satisfied": result.Satisfied,
				"timestamp": result.Timestamp.Format(time.RFC3339),
				"response":  result.Response,
			}

			return result.Satisfied, conditionParams, nil
		}
		// If JSON parsing failed, fall back to string checking
	}

	// Simple string-based check as fallback
	if containsString(responseStr, "Condition satisfied: true") {
		satisfied = true
		// Create basic condition params for string-based check with only the three requested fields
		conditionParams = map[string]interface{}{
			"satisfied": satisfied,
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"response":  responseStr, // In string-based mode, use the raw response as the response value
		}
	} else {
		// Even if not satisfied, still provide the three parameters
		conditionParams = map[string]interface{}{
			"satisfied": satisfied,
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"response":  responseStr,
		}
	}

	w.scheduler.Logger().Infof("Condition satisfied (string check): %v", satisfied)
	return satisfied, conditionParams, nil
}

// Helper function to check if a string contains another string
func containsString(s, substr string) bool {
	return strings.Contains(s, substr)
}

func (w *ConditionBasedWorker) executeTask(jobData *types.HandleCreateJobData, triggerData *types.TriggerData) error {
	w.scheduler.Logger().Infof("Executing condition-based job: %d", w.jobID)

	taskData := &types.CreateTaskData{
		JobID:            w.jobID,
		TaskDefinitionID: jobData.TaskDefinitionID,
		TaskPerformerID:  0,
	}

	performerData, err := services.GetPerformer()
	if err != nil {
		w.scheduler.Logger().Errorf("Failed to get performer data for job %d: %v", w.jobID, err)
		return err
	}

	taskData.TaskPerformerID = performerData.KeeperID

	taskID, status, err := services.CreateTaskData(taskData)
	if err != nil {
		w.scheduler.Logger().Errorf("Failed to create task data for job %d: %v", w.jobID, err)
		return err
	}

	triggerData.TaskID = taskID

	if !status {
		return fmt.Errorf("failed to create task data for job %d", w.jobID)
	}

	status, err = services.SendTaskToPerformer(jobData, triggerData, performerData)
	if err != nil {
		w.scheduler.Logger().Errorf("Error sending task to performer: %v", err)
		return err
	}

	w.scheduler.Logger().Infof("Task sent for job %d to performer", w.jobID)

	if err := w.handleLinkedJob(w.scheduler, jobData); err != nil {
		w.scheduler.Logger().Errorf("Failed to execute linked job for job %d: %v", w.jobID, err)
	}

	if !status {
		return fmt.Errorf("failed to send task to performer for job %d", w.jobID)
	}

	return nil
}

func (w *ConditionBasedWorker) GetJobID() int64 {
	return w.jobID
}

func (w *ConditionBasedWorker) GetStatus() string {
	return w.status
}

func (w *ConditionBasedWorker) GetError() string {
	return w.error
}

func (w *ConditionBasedWorker) GetRetries() int {
	return w.currentRetry
}

// Add UpdateLastExecutedTime method to allow updating the last execution timestamp
func (w *ConditionBasedWorker) UpdateLastExecutedTime(timestamp time.Time) {
	// Update the jobData with the new timestamp
	if w.jobData != nil {
		w.jobData.LastExecutedAt = timestamp
		w.scheduler.Logger().Infof("Updated LastExecutedAt for job %d to %v", w.jobID, timestamp)
	}
}
