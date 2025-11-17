package execution

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/google/uuid"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
	dockertypes "github.com/trigg3rX/triggerx-backend/pkg/dockerexecutor/types"
)

// ExecuteCustomScript handles custom script execution (TaskDefinitionID = 7)
// Returns: script output, storage updates, error
//
// Phase 1: Scripts execute without environment variable injection
// - Scripts can OUTPUT storage via stderr: STORAGE_SET:key=value
// - Scripts cannot READ previous storage (Phase 2 feature)
// - Execution metadata (job_id, task_id, timestamp) passed via script context
func (e *TaskExecutor) ExecuteCustomScript(
	ctx context.Context,
	targetData *types.TaskTargetData,
	triggerData *types.TaskTriggerData,
) (*types.CustomScriptOutput, map[string]string, error) {
	e.logger.Infof("[CustomScript] Starting execution for job %s", targetData.JobID.String())

	// Execute script in Docker (Phase 1: no env var injection)
	scriptURL := targetData.DynamicArgumentsScriptUrl
	scriptLanguage := targetData.ScriptLanguage
	if scriptLanguage == "" {
		scriptLanguage = string(dockertypes.LanguageTS) // Default
	}

	e.logger.Infof("[CustomScript] Executing %s script from: %s", scriptLanguage, scriptURL)

	// Use standard Execute method (env var injection deferred to Phase 2)
	result, err := e.validator.GetDockerExecutor().Execute(
		ctx,
		scriptURL,
		scriptLanguage,
		1, // noOfAttesters
	)
	if err != nil {
		return nil, nil, fmt.Errorf("docker execution failed: %w", err)
	}

	if !result.Success {
		return nil, nil, fmt.Errorf("script execution failed: %s", result.Error)
	}

	// Parse script output (JSON from stdout)
	var scriptOutput types.CustomScriptOutput
	err = json.Unmarshal([]byte(result.Output), &scriptOutput)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse script output: %w", err)
	}

	// Validate output
	if err := validateCustomScriptOutput(&scriptOutput); err != nil {
		return nil, nil, fmt.Errorf("invalid script output: %w", err)
	}

	e.logger.Infof("[CustomScript] Script output: shouldExecute=%v, targetContract=%s",
		scriptOutput.ShouldExecute, scriptOutput.TargetContract)

	// Extract storage updates from JSON output
	storageUpdates := scriptOutput.StorageUpdates
	if storageUpdates == nil {
		storageUpdates = make(map[string]string)
	}
	if len(storageUpdates) > 0 {
		e.logger.Infof("[CustomScript] Found %d storage updates", len(storageUpdates))
	}

	return &scriptOutput, storageUpdates, nil
}

// prepareCustomScriptEnv prepares environment variables for script execution
// Phase 2: Environment variable injection for storage and context
// Currently unused in Phase 1
/* func prepareCustomScriptEnv(
	targetData *types.TaskTargetData,
	triggerData *types.TaskTriggerData,
) map[string]string {
	env := make(map[string]string)

	// TriggerX context
	env["TRIGGERX_TIMESTAMP"] = fmt.Sprintf("%d", time.Now().Unix())
	env["TRIGGERX_JOB_ID"] = targetData.JobID.String()
	env["TRIGGERX_TASK_ID"] = fmt.Sprintf("%d", targetData.TaskID)
	env["TRIGGERX_EXECUTION_ID"] = generateExecutionID(targetData.JobID, time.Now())

	// Inject storage from task data
	if targetData.ScriptStorage != nil {
		for key, value := range targetData.ScriptStorage {
			env["TRIGGERX_STORAGE_"+key] = value
		}
	}

	return env
} */

// validateCustomScriptOutput validates the script output format
func validateCustomScriptOutput(output *types.CustomScriptOutput) error {
	if output.ShouldExecute {
		if output.TargetContract == "" {
			return fmt.Errorf("targetContract required when shouldExecute=true")
		}
		if output.Calldata == "" {
			return fmt.Errorf("calldata required when shouldExecute=true")
		}
		// Validate address format
		if !strings.HasPrefix(output.TargetContract, "0x") || len(output.TargetContract) != 42 {
			return fmt.Errorf("invalid targetContract address format")
		}
		// Validate calldata format
		if !strings.HasPrefix(output.Calldata, "0x") {
			return fmt.Errorf("calldata must be hex string starting with 0x")
		}
	}
	return nil
}

// parseStorageUpdates parses STORAGE_SET commands from stderr
// Format: STORAGE_SET:key=value
func parseStorageUpdates(stderr string) map[string]string {
	updates := make(map[string]string)
	lines := strings.Split(stderr, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "STORAGE_SET:") {
			// Remove prefix
			kvPair := strings.TrimPrefix(line, "STORAGE_SET:")
			// Split by first =
			parts := strings.SplitN(kvPair, "=", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				updates[key] = value
			}
		}
	}

	return updates
}

// generateExecutionID generates a unique execution identifier
func generateExecutionID(jobID *types.BigInt, timestamp time.Time) string {
	return fmt.Sprintf("exec_%s_%s", jobID.String(), uuid.New().String()[:8])
}

// generateExecutionProof generates cryptographic proof of execution
func generateExecutionProof(
	executionID string,
	jobID *types.BigInt,
	storage map[string]string,
	output *types.CustomScriptOutput,
	timestamp time.Time,
	performerAddress string,
) *types.ExecutionProof {
	// Calculate input hash
	inputData := fmt.Sprintf("%d:%s:%s",
		timestamp.Unix(),
		jobID.String(),
		hashStorage(storage),
	)
	inputHash := crypto.Keccak256Hash([]byte(inputData)).Hex()

	// Calculate output hash
	outputData := fmt.Sprintf("%t:%s:%s",
		output.ShouldExecute,
		output.TargetContract,
		output.Calldata,
	)
	outputHash := crypto.Keccak256Hash([]byte(outputData)).Hex()

	// Sign the proof
	signature := signProof(inputHash, outputHash)

	return &types.ExecutionProof{
		ExecutionID:      executionID,
		JobID:            jobID.String(),
		Timestamp:        timestamp.Unix(),
		InputHash:        inputHash,
		OutputHash:       outputHash,
		Signature:        signature,
		PerformerAddress: performerAddress,
	}
}

// Helper functions

func hashStorage(storage map[string]string) string {
	if len(storage) == 0 {
		return "0x0"
	}

	// Create deterministic hash
	data := ""
	for k, v := range storage {
		data += fmt.Sprintf("%s=%s;", k, v)
	}

	hash := sha256.Sum256([]byte(data))
	return "0x" + hex.EncodeToString(hash[:])
}

func signProof(inputHash, outputHash string) string {
	// TODO: Implement actual signature with keeper's private key
	// For now, just combine hashes
	combined := inputHash + outputHash
	hash := sha256.Sum256([]byte(combined))
	return "0x" + hex.EncodeToString(hash[:])
}
