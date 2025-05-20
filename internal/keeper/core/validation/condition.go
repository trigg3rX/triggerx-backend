package validation

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
	"io/ioutil"

	"github.com/trigg3rX/triggerx-backend/pkg/types"
	"github.com/trigg3rX/triggerx-backend/internal/keeper/config"
	"github.com/trigg3rX/triggerx-backend/internal/keeper/client/ipfs"
)

// ValidateConditionBasedJob validates a condition-based job by executing the condition script
// and checking if it returns true
func (v *TaskValidator) ValidateConditionBasedTask(job *types.HandleCreateJobData, ipfsData *types.IPFSData) (bool, error) {
	// Ensure this is a condition-based job
	if job.TaskDefinitionID != 5 && job.TaskDefinitionID != 6 {
		return false, fmt.Errorf("not a condition-based job: task definition ID %d", job.TaskDefinitionID)
	}

	v.logger.Infof("Validating condition-based job %d (taskDefID: %d)", job.JobID, job.TaskDefinitionID)

	// For non-recurring jobs, check if job has already been executed and shouldn't run again
	if !job.Recurring && !job.LastExecutedAt.IsZero() {
		v.logger.Infof("Job %d is non-recurring and has already been executed on %s",
			job.JobID, job.LastExecutedAt.Format(time.RFC3339))
		return false, nil
	}

	// Check if the ScriptTriggerFunction is provided
	if job.ScriptTriggerFunction == "" {
		return false, fmt.Errorf("missing ScriptTriggerFunction for condition-based job %d", job.JobID)
	}

	// Fetch and execute the condition script
	v.logger.Infof("Fetching condition script from IPFS: %s", job.ScriptTriggerFunction)
	scriptContent, err := ipfs.FetchIPFSContent(config.GetIpfsHost(), job.ScriptTriggerFunction)
	if err != nil {
		return false, fmt.Errorf("failed to fetch condition script: %v", err)
	}
	v.logger.Infof("Successfully fetched condition script for job %d", job.JobID)

	// Check if job is within its timeframe before executing script
	now := time.Now().UTC()
	if job.TimeFrame > 0 {
		endTime := job.CreatedAt.Add(time.Duration(job.TimeFrame) * time.Second)
		if now.After(endTime) {
			v.logger.Infof("Job %d is outside its timeframe (created: %s, timeframe: %d seconds)",
				job.JobID, job.CreatedAt.Format(time.RFC3339), job.TimeFrame)
			return false, nil
		}
	}

	// Create a temporary file for the script
	tempFile, err := ioutil.TempFile("", "condition-*.go")
	if err != nil {
		return false, fmt.Errorf("failed to create temporary file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	if _, err := tempFile.Write([]byte(scriptContent)); err != nil {
		return false, fmt.Errorf("failed to write script to file: %v", err)
	}
	if err := tempFile.Close(); err != nil {
		return false, fmt.Errorf("failed to close temporary file: %v", err)
	}

	// Create a temp directory for the script's build output
	tempDir, err := ioutil.TempDir("", "condition-build")
	if err != nil {
		return false, fmt.Errorf("failed to create temporary build directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Compile the script
	v.logger.Infof("Compiling condition script for job %d", job.JobID)
	outputBinary := filepath.Join(tempDir, "condition")
	cmd := exec.Command("go", "build", "-o", outputBinary, tempFile.Name())
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return false, fmt.Errorf("failed to compile condition script: %v, stderr: %s", err, stderr.String())
	}

	// Run the compiled script
	v.logger.Infof("Executing condition script for job %d", job.JobID)
	result := exec.Command(outputBinary)
	stdout, err := result.Output()
	if err != nil {
		return false, fmt.Errorf("failed to run condition script: %v", err)
	}

	// Parse the output to determine if condition is satisfied
	// Look for a line containing "Condition satisfied: true" or "Condition satisfied: false"
	lines := strings.Split(string(stdout), "\n")
	for _, line := range lines {
		if strings.Contains(line, "Condition satisfied: true") {
			v.logger.Infof("Condition script reported satisfaction for job %d", job.JobID)
			return true, nil
		} else if strings.Contains(line, "Condition satisfied: false") {
			v.logger.Infof("Condition script reported non-satisfaction for job %d", job.JobID)
			return false, nil
		}
	}

	// If no explicit condition found, try parsing as JSON
	var conditionResult struct {
		Satisfied bool `json:"satisfied"`
	}
	if err := json.Unmarshal(stdout, &conditionResult); err != nil {
		v.logger.Warnf("Could not determine condition result from output for job %d: %s", job.JobID, string(stdout))
		return false, fmt.Errorf("could not determine condition result from output: %s", string(stdout))
	}

	v.logger.Infof("Condition script for job %d returned satisfied: %v", job.JobID, conditionResult.Satisfied)
	return conditionResult.Satisfied, nil
}