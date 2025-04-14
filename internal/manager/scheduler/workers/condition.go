package workers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
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
					triggerData.ConditionParams = conditionParams

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
	responseStr := string(body)

	if responseStr == "" {
		return false, nil, fmt.Errorf("empty response from condition script")
	}

	// Create default condition parameters
	timestamp := time.Now().UTC()
	satisfied := false
	var response interface{} = nil

	// First, check if the response is already in JSON format
	if len(responseStr) > 0 && (responseStr[0] == '{' || responseStr[0] == '[') {
		type ConditionResult struct {
			Satisfied bool        `json:"Satisfied"`
			Timestamp time.Time   `json:"Timestamp"`
			Response  interface{} `json:"Response"`
		}

		var result ConditionResult
		if err := json.Unmarshal(body, &result); err == nil {
			w.scheduler.Logger().Infof("Parsed JSON condition result: satisfied=%v", result.Satisfied)

			conditionParams := map[string]interface{}{
				"satisfied": result.Satisfied,
				"timestamp": result.Timestamp.Format(time.RFC3339),
				"response":  result.Response,
			}

			return result.Satisfied, conditionParams, nil
		}
	}

	// Second, check if we received the output of the executed code
	if strings.Contains(responseStr, "Condition satisfied:") {
		// This appears to be executed program output
		w.scheduler.Logger().Infof("Received execution output from condition script")

		// Check if condition is satisfied
		satisfied = strings.Contains(responseStr, "Condition satisfied: true")

		// Try to extract response value from the output
		responseMatch := strings.Index(responseStr, "Response:")
		if responseMatch != -1 {
			responseLine := responseStr[responseMatch:]
			endOfLine := strings.Index(responseLine, "\n")
			if endOfLine != -1 {
				responseLine = responseLine[:endOfLine]
				// Extract value from the line
				parts := strings.Split(responseLine, ":")
				if len(parts) > 1 {
					responseValue := strings.TrimSpace(parts[1])
					// Try to parse as numeric first
					if responseFloat, err := strconv.ParseFloat(responseValue, 64); err == nil {
						response = responseFloat
					} else {
						// If not numeric, use as string
						response = responseValue
					}
				}
			}
		}

		// Try to extract timestamp if present
		timestampMatch := strings.Index(responseStr, "Timestamp:")
		if timestampMatch != -1 {
			timestampLine := responseStr[timestampMatch:]
			endOfLine := strings.Index(timestampLine, "\n")
			if endOfLine != -1 {
				timestampLine = timestampLine[:endOfLine]
				parts := strings.Split(timestampLine, ":")
				if len(parts) > 1 {
					timestampStr := strings.TrimSpace(parts[1])
					if parsedTime, err := time.Parse(time.RFC3339, timestampStr); err == nil {
						timestamp = parsedTime
					}
				}
			}
		}
	} else {
		// This appears to be source code
		w.scheduler.Logger().Infof("Received source code, will attempt to execute it")

		// Create a temporary directory to store and execute the code
		tempDir, err := os.MkdirTemp("", "condition-script")
		if err != nil {
			w.scheduler.Logger().Errorf("Failed to create temp directory: %v", err)
			return false, nil, fmt.Errorf("failed to create temp directory: %v", err)
		}
		defer os.RemoveAll(tempDir)

		// Write the source code to a file
		scriptPath := filepath.Join(tempDir, "condition_script.go")
		if err := os.WriteFile(scriptPath, body, 0644); err != nil {
			w.scheduler.Logger().Errorf("Failed to write script to file: %v", err)
			return false, nil, fmt.Errorf("failed to write script to file: %v", err)
		}

		// Create a command to execute the Go script
		cmd := exec.Command("go", "run", scriptPath)
		outputBytes, err := cmd.CombinedOutput()
		if err != nil {
			w.scheduler.Logger().Errorf("Failed to execute script: %v, output: %s", err, string(outputBytes))
			return false, nil, fmt.Errorf("failed to execute script: %v, output: %s", err, string(outputBytes))
		}

		output := string(outputBytes)
		w.scheduler.Logger().Infof("Script execution output: %s", output)

		// Parse the output to extract satisfied, timestamp, and response
		satisfied = strings.Contains(output, "Condition satisfied: true")

		// Extract response
		responseMatch := strings.Index(output, "Response:")
		if responseMatch != -1 {
			responseLine := output[responseMatch:]
			endOfLine := strings.Index(responseLine, "\n")
			if endOfLine != -1 {
				responseLine = responseLine[:endOfLine]
				parts := strings.Split(responseLine, ":")
				if len(parts) > 1 {
					responseValue := strings.TrimSpace(parts[1])
					if responseFloat, err := strconv.ParseFloat(responseValue, 64); err == nil {
						response = responseFloat
					} else {
						response = responseValue
					}
				}
			}
		}

		// Extract timestamp
		timestampMatch := strings.Index(output, "Timestamp:")
		if timestampMatch != -1 {
			timestampLine := output[timestampMatch:]
			endOfLine := strings.Index(timestampLine, "\n")
			if endOfLine != -1 {
				timestampLine = timestampLine[:endOfLine]
				parts := strings.Split(timestampLine, ":")
				if len(parts) > 1 {
					timestampStr := strings.TrimSpace(parts[1])
					if parsedTime, err := time.Parse(time.RFC3339, timestampStr); err == nil {
						timestamp = parsedTime
					}
				}
			}
		}
	}

	// If response is still nil, provide a default value
	if response == nil {
		if satisfied {
			response = "Condition satisfied"
		} else {
			response = "Condition not satisfied"
		}
	}

	conditionParams := map[string]interface{}{
		"satisfied": satisfied,
		"timestamp": timestamp.Format(time.RFC3339),
		"response":  response,
	}

	w.scheduler.Logger().Infof("Condition result: satisfied=%v", satisfied)
	return satisfied, conditionParams, nil
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
