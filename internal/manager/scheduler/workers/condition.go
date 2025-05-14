package workers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
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
	w.done = make(chan bool)

	w.scheduler.Logger().Infof("Starting condition-based job %d", w.jobID)
	w.scheduler.Logger().Infof("Listening to %s", w.jobData.ScriptTriggerFunction)

	var triggerData types.TriggerData
	triggerData.TimeInterval = w.jobData.TimeInterval
	triggerData.LastExecuted = time.Now().UTC()
	triggerData.TriggerTxHash = ""

	var endTime time.Time
	if w.jobData.TimeFrame > 0 {
		endTime = time.Now().UTC().Add(time.Duration(w.jobData.TimeFrame) * time.Second)
	}

	go func() {
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
				if w.jobData.TimeFrame > 0 && time.Now().UTC().After(endTime) {
					w.scheduler.Logger().Infof("Timeframe reached for job %d, stopping worker", w.jobID)
					return
				}

				satisfied, err := w.checkCondition()
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
					triggerData.ConditionParams = make(map[string]interface{})

					if err := w.executeTask(w.jobData, &triggerData); err != nil {
						w.handleError(err)
						if !w.jobData.Recurring || w.status == "failed" {
							return
						}
						continue
					}

					if !w.jobData.Recurring {
						return
					}

					triggerData.LastExecuted = time.Now().UTC()
					w.scheduler.Logger().Infof("Job %d executed. Continuing to monitor condition due to recurring flag", w.jobID)

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

	if w.done != nil {
		defer func() {
			if r := recover(); r != nil {
				w.scheduler.Logger().Warnf("Attempted to close already closed channel for job %d: %v", w.jobID, r)
			}
		}()

		select {
		case <-w.done:
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

func (w *ConditionBasedWorker) checkCondition() (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	w.scheduler.Logger().Infof("Fetching condition script from: %s", w.jobData.ScriptTriggerFunction)

	req, err := http.NewRequestWithContext(ctx, "GET", w.jobData.ScriptTriggerFunction, nil)
	if err != nil {
		w.scheduler.Logger().Errorf("Failed to create request: %v", err)
		return false, fmt.Errorf("failed to create request: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		w.scheduler.Logger().Errorf("Failed to fetch condition script: %v", err)
		return false, fmt.Errorf("failed to fetch condition script: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errMsg := fmt.Sprintf("API returned status %d", resp.StatusCode)
		w.scheduler.Logger().Errorf(errMsg)
		return false, fmt.Errorf(errMsg)
	}

	scriptContent, err := io.ReadAll(resp.Body)
	if err != nil {
		w.scheduler.Logger().Errorf("Failed to read script content: %v", err)
		return false, fmt.Errorf("failed to read script content: %v", err)
	}

	tempFile, err := ioutil.TempFile("", "condition-*.go")
	if err != nil {
		w.scheduler.Logger().Errorf("Failed to create temp file: %v", err)
		return false, fmt.Errorf("failed to create temp file: %v", err)
	}
	defer func() {
		if err := os.Remove(tempFile.Name()); err != nil {
			w.scheduler.Logger().Warnf("Failed to remove temp file %s: %v", tempFile.Name(), err)
		}
	}()

	if _, err := tempFile.Write(scriptContent); err != nil {
		w.scheduler.Logger().Errorf("Failed to write script: %v", err)
		return false, fmt.Errorf("failed to write script: %v", err)
	}
	if err := tempFile.Close(); err != nil {
		w.scheduler.Logger().Errorf("Failed to close temp file: %v", err)
		return false, fmt.Errorf("failed to close temp file: %v", err)
	}

	w.scheduler.Logger().Infof("Temporary script file created: %s", tempFile.Name())

	tempDir, err := ioutil.TempDir("", "condition-build")
	if err != nil {
		w.scheduler.Logger().Errorf("Failed to create temp build dir: %v", err)
		return false, fmt.Errorf("failed to create temp build dir: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			w.scheduler.Logger().Warnf("Failed to remove temp dir %s: %v", tempDir, err)
		}
	}()

	outputBinary := filepath.Join(tempDir, "condition")
	cmd := exec.CommandContext(ctx, "go", "build", "-o", outputBinary, tempFile.Name())
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	w.scheduler.Logger().Infof("Compiling condition script...")

	if err := cmd.Run(); err != nil {
		compileError := fmt.Sprintf("Compilation failed: %v\nStderr: %s", err, stderr.String())
		w.scheduler.Logger().Errorf(compileError)
		return false, fmt.Errorf(compileError)
	}

	w.scheduler.Logger().Infof("Successfully compiled condition script to: %s", outputBinary)

	runCmd := exec.CommandContext(ctx, outputBinary)
	w.scheduler.Logger().Infof("Executing condition script...")

	stdout, err := runCmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			execError := fmt.Sprintf("Script execution failed: %v\nStderr: %s", err, exitErr.Stderr)
			w.scheduler.Logger().Errorf(execError)
			return false, fmt.Errorf(execError)
		}
		w.scheduler.Logger().Errorf("Failed to run script: %v", err)
		return false, fmt.Errorf("failed to run script: %v", err)
	}

	output := string(stdout)
	w.scheduler.Logger().Infof("Condition script output:\n%s", output)

	switch {
	case strings.Contains(output, "Condition satisfied: true"):
		w.scheduler.Logger().Infof("Condition SATISFIED")
		return true, nil
	case strings.Contains(output, "Condition satisfied: false"):
		w.scheduler.Logger().Infof("Condition NOT SATISFIED")
		return false, nil
	default:
		var result struct{ Satisfied bool }
		if json.Unmarshal(stdout, &result) == nil {
			if result.Satisfied {
				w.scheduler.Logger().Infof("Condition SATISFIED (from JSON)")
			} else {
				w.scheduler.Logger().Infof("Condition NOT SATISFIED (from JSON)")
			}
			return result.Satisfied, nil
		}
		errMsg := fmt.Sprintf("Unable to determine condition from output: %s", output)
		w.scheduler.Logger().Errorf(errMsg)
		return false, fmt.Errorf(errMsg)
	}
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

func (w *ConditionBasedWorker) UpdateLastExecutedTime(timestamp time.Time) {
	if w.jobData != nil {
		w.jobData.LastExecutedAt = timestamp
		w.scheduler.Logger().Infof("Updated LastExecutedAt for job %d to %v", w.jobID, timestamp)
	}
}
