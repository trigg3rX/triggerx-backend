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

	"github.com/trigg3rX/triggerx-backend/internal/keeper/utils"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// ValidateConditionBasedJob validates a condition-based job by executing the condition script
// and checking if it returns true
func (v *TaskValidator) ValidateConditionBasedTask(ipfsData types.IPFSData) (bool, error) {
	targetData := ipfsData.TargetData
	triggerData := ipfsData.TriggerData
	v.logger.Infof("Validating condition-based job %d (taskDefID: %d)", targetData.TaskID, targetData.TaskDefinitionID)

	// For non-recurring jobs, check if job has already been executed and shouldn't run again
	if !triggerData.Recurring && !triggerData.TriggerTimestamp.IsZero() {
		v.logger.Infof("Job %d is non-recurring and has already been executed on %s",
			targetData.TaskID, triggerData.TriggerTimestamp.Format(time.RFC3339))
		return false, nil
	}

	// Check if the ScriptTriggerFunction is provided
	if triggerData.ConditionSourceUrl == "" {
		return false, fmt.Errorf("missing ScriptTriggerFunction for condition-based job %d", targetData.TaskID)
	}

	// Fetch and execute the condition script
	v.logger.Infof("Fetching condition script from IPFS: %s", triggerData.ConditionSourceUrl)
	scriptContent, err := utils.FetchDataFromUrl(triggerData.ConditionSourceUrl)
	if err != nil {
		return false, fmt.Errorf("failed to fetch condition script: %v", err)
	}
	v.logger.Infof("Successfully fetched condition script for job %d", targetData.TaskID)

	// Check if job is within its timeframe before executing script
	if triggerData.TriggerTimestamp.After(triggerData.ExpirationTime) {
		v.logger.Infof("Job %d is outside its timeframe (created: %s, timeframe: %d seconds)", targetData.TaskID, triggerData.TriggerTimestamp.Format(time.RFC3339), triggerData.TimeInterval)
		return false, nil
	}

	// Create a temporary file for the script
	tempFile, err := os.CreateTemp("", "condition-*.go")
	if err != nil {
		return false, fmt.Errorf("failed to create temporary file: %v", err)
	}
	defer func() {
		if err := os.Remove(tempFile.Name()); err != nil {
			v.logger.Warnf("Failed to remove temporary file %s: %v", tempFile.Name(), err)
		}
	}()

	if _, err := tempFile.Write([]byte(scriptContent)); err != nil {
		return false, fmt.Errorf("failed to write script to file: %v", err)
	}
	if err := tempFile.Close(); err != nil {
		return false, fmt.Errorf("failed to close temporary file: %v", err)
	}

	// Create a temp directory for the script's build output
	tempDir, err := os.MkdirTemp("", "condition-build")
	if err != nil {
		return false, fmt.Errorf("failed to create temporary build directory: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			v.logger.Warnf("Failed to remove temporary directory %s: %v", tempDir, err)
		}
	}()

	// Compile the script
	v.logger.Infof("Compiling condition script for job %d", targetData.TaskID)
	outputBinary := filepath.Join(tempDir, "condition")
	cmd := exec.Command("go", "build", "-o", outputBinary, tempFile.Name())
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return false, fmt.Errorf("failed to compile condition script: %v, stderr: %s", err, stderr.String())
	}

	// Run the compiled script
	v.logger.Infof("Executing condition script for job %d", targetData.TaskID)
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
			v.logger.Infof("Condition script reported satisfaction for job %d", targetData.TaskID)
			return true, nil
		} else if strings.Contains(line, "Condition satisfied: false") {
			v.logger.Infof("Condition script reported non-satisfaction for job %d", targetData.TaskID)
			return false, nil
		}
	}

	// If no explicit condition found, try parsing as JSON
	var conditionResult struct {
		Satisfied bool `json:"satisfied"`
	}
	if err := json.Unmarshal(stdout, &conditionResult); err != nil {
		v.logger.Warnf("Could not determine condition result from output for job %d: %s", targetData.TaskID, string(stdout))
		return false, fmt.Errorf("could not determine condition result from output: %s", string(stdout))
	}

	v.logger.Infof("Condition script for job %d returned satisfied: %v", targetData.TaskID, conditionResult.Satisfied)
	return conditionResult.Satisfied, nil
}
