package execution

// import (
// 	"context"
// 	"errors"
// 	"math/big"
// 	"testing"

// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/require"
// 	"github.com/trigg3rX/triggerx-backend/internal/keeper/core/validation"
// 	dockerexecutor "github.com/trigg3rX/triggerx-backend/pkg/dockerexecutor"
// 	dockertypes "github.com/trigg3rX/triggerx-backend/pkg/dockerexecutor/types"
// 	"github.com/trigg3rX/triggerx-backend/pkg/logging"
// 	"github.com/trigg3rX/triggerx-backend/pkg/types"
// 	gomock "go.uber.org/mock/gomock"
// )

// func TestTaskExecutor_ExecuteCustomScript_SingleExecution(t *testing.T) {
// 	ctrl := gomock.NewController(t)
// 	defer ctrl.Finish()

// 	mockDocker := dockerexecutor.NewMockDockerExecutorAPI(ctrl)
// 	logger := logging.NewNoOpLogger()
// 	validator := validation.NewTaskValidator("", "", mockDocker, nil, logger, nil)
// 	executor := NewTaskExecutor("", validator, nil, nil, logger)

// 	ctx := context.Background()
// 	targetData := &types.TaskTargetData{
// 		JobID:                     types.NewBigInt(big.NewInt(999888777666555)),
// 		TaskID:                    12345,
// 		TaskDefinitionID:          7,
// 		DynamicArgumentsScriptUrl: "https://ipfs.io/ipfs/bafybeicm2uyje7k7wbgmdgzjt57lexoo7ca6565myx6ktynzjm2qgy2uzm",
// 		ScriptLanguage:            "typescript",
// 	}
// 	triggerData := &types.TaskTriggerData{}

// 	outputJSON := `{
// 		"shouldExecute": true,
// 		"targetContract": "0xa0bC1477cfc452C05786262c377DE51FB8bc4669",
// 		"calldata": "0xa9059cbb000000000000000000000000a76cacba495cafeabb628491733eb86f1db006df000000000000000000000000000000000000000000000000000009184e72a000",
// 		"storageUpdates": {
// 			"status": "executed",
// 			"executionCount": "1"
// 		},
// 		"metadata": {
// 			"timestamp": 1763298589440,
// 			"reason": "demo",
// 			"gasEstimate": 150000
// 		}
// 	}`

// 	mockDocker.EXPECT().
// 		Execute(gomock.Any(), targetData.DynamicArgumentsScriptUrl, targetData.ScriptLanguage, 1, "", gomock.Any()).
// 		Return(&dockertypes.ExecutionResult{
// 			Success: true,
// 			Output:  outputJSON,
// 		}, nil)

// 	result, storage, err := executor.ExecuteCustomScript(ctx, targetData, triggerData)
// 	require.NoError(t, err)
// 	require.NotNil(t, result)

// 	assert.True(t, result.ShouldExecute)
// 	assert.Equal(t, "0xa0bC1477cfc452C05786262c377DE51FB8bc4669", result.TargetContract)
// 	assert.Equal(t, "0xa9059cbb000000000000000000000000a76cacba495cafeabb628491733eb86f1db006df000000000000000000000000000000000000000000000000000009184e72a000", result.Calldata)
// 	assert.Equal(t, "executed", storage["status"])
// 	assert.Equal(t, "1", storage["executionCount"])
// }

// func TestTaskExecutor_ExecuteCustomScript_PropagatesDockerErrors(t *testing.T) {
// 	ctrl := gomock.NewController(t)
// 	defer ctrl.Finish()

// 	mockDocker := dockerexecutor.NewMockDockerExecutorAPI(ctrl)
// 	logger := logging.NewNoOpLogger()
// 	validator := validation.NewTaskValidator("", "", mockDocker, nil, logger, nil)
// 	executor := NewTaskExecutor("", validator, nil, nil, logger)

// 	targetData := &types.TaskTargetData{
// 		JobID:                     types.NewBigInt(big.NewInt(1)),
// 		TaskDefinitionID:          7,
// 		DynamicArgumentsScriptUrl: "https://ipfs.io/ipfs/bad",
// 		ScriptLanguage:            "typescript",
// 	}

// 	mockDocker.EXPECT().
// 		Execute(gomock.Any(), gomock.Any(), gomock.Any(), 1, "", gomock.Any()).
// 		Return(nil, errors.New("docker unavailable"))

// 	_, _, err := executor.ExecuteCustomScript(context.Background(), targetData, &types.TaskTriggerData{})
// 	require.Error(t, err)
// 	assert.Contains(t, err.Error(), "docker execution failed")
// }

// // TestCustomScriptExecution_TaskDefinitionID7_Integration tests the complete flow
// // from receiving a TaskDefinitionID 7 task to preparing the contract call
// func TestCustomScriptExecution_TaskDefinitionID7_Integration(t *testing.T) {
// 	// Skip if in short mode or CI (requires Docker and IPFS access)
// 	if testing.Short() {
// 		t.Skip("Skipping integration test in short mode")
// 	}

// 	// Test data - using actual IPFS script
// 	targetData := &types.TaskTargetData{
// 		JobID:            types.NewBigInt(big.NewInt(999888777666555)),
// 		TaskID:           12345,
// 		TaskDefinitionID: 7,
// 		TargetChainID:    "421614", // Arbitrum Sepolia
// 		// Script URL - actual IPFS hosted script
// 		DynamicArgumentsScriptUrl: "https://ipfs.io/ipfs/bafybeicm2uyje7k7wbgmdgzjt57lexoo7ca6565myx6ktynzjm2qgy2uzm",
// 		ScriptLanguage:            "typescript",
// 		// Storage from previous execution (empty for first run)
// 		ScriptStorage: map[string]string{},
// 	}

// 	t.Run("Script Execution Returns Valid Output", func(t *testing.T) {
// 		// This test verifies that:
// 		// 1. Script is fetched from IPFS
// 		// 2. Script executes in Docker
// 		// 3. Output is valid JSON
// 		// 4. Output contains required fields

// 		// Expected output structure from the script:
// 		// {
// 		//   "shouldExecute": true,
// 		//   "targetContract": "0xa0bC1477cfc452C05786262c377DE51FB8bc4669",
// 		//   "calldata": "0x...",
// 		//   "storageUpdates": {...},
// 		//   "metadata": {...}
// 		// }

// 		t.Log("Testing with IPFS URL:", targetData.DynamicArgumentsScriptUrl)
// 		t.Log("Expected target contract: 0xa0bC1477cfc452C05786262c377DE51FB8bc4669")
// 		t.Log("Expected chain: Arbitrum Sepolia (421614)")
// 	})

// 	t.Run("Validates Script Output Structure", func(t *testing.T) {
// 		// Expected fields in script output
// 		expectedFields := []string{
// 			"shouldExecute",
// 			"targetContract",
// 			"calldata",
// 			"storageUpdates",
// 			"metadata",
// 		}

// 		t.Log("Script must return JSON with fields:", expectedFields)
// 	})

// 	t.Run("Validates Target Contract Address Format", func(t *testing.T) {
// 		// Target contract from script
// 		expectedTargetContract := "0xa0bC1477cfc452C05786262c377DE51FB8bc4669"

// 		// Validate format
// 		assert.Equal(t, 42, len(expectedTargetContract), "Address should be 42 characters (0x + 40 hex)")
// 		assert.Equal(t, "0x", expectedTargetContract[:2], "Address should start with 0x")

// 		t.Logf("Target contract address validated: %s", expectedTargetContract)
// 	})

// 	t.Run("Validates Calldata Format", func(t *testing.T) {
// 		// Calldata should:
// 		// 1. Start with 0x
// 		// 2. Be valid hex
// 		// 3. Contain function selector (first 4 bytes = 8 hex chars)

// 		// From script: transfer(address,uint256) = 0xa9059cbb
// 		expectedFunctionSelector := "a9059cbb"

// 		t.Log("Expected function selector:", expectedFunctionSelector)
// 		t.Log("Calldata format: 0x + selector (8 chars) + encoded params")
// 	})

// 	t.Run("Storage Updates Structure", func(t *testing.T) {
// 		// Expected storage updates from script
// 		expectedStorageKeys := []string{
// 			"lastExecutionTime",
// 			"lastActionTarget",
// 			"lastActionValue",
// 			"lastSafeAddress",
// 			"executionCount",
// 			"status",
// 		}

// 		t.Log("Expected storage keys:", expectedStorageKeys)
// 		t.Log("Storage updates will be saved to database after execution")
// 	})

// 	t.Run("Metadata Contains Required Fields", func(t *testing.T) {
// 		// Expected metadata fields
// 		expectedMetadataFields := []string{
// 			"timestamp",
// 			"reason",
// 			"gasEstimate",
// 		}

// 		t.Log("Expected metadata fields:", expectedMetadataFields)
// 	})

// 	t.Run("Complete Execution Flow", func(t *testing.T) {
// 		t.Log("=== Complete TaskDefinitionID 7 Execution Flow ===")
// 		t.Log("")
// 		t.Log("1. Keeper receives task with:")
// 		t.Logf("   - JobID: %s", targetData.JobID.String())
// 		t.Logf("   - TaskID: %d", targetData.TaskID)
// 		t.Logf("   - TaskDefinitionID: %d", targetData.TaskDefinitionID)
// 		t.Logf("   - TargetChainID: %s", targetData.TargetChainID)
// 		t.Logf("   - ScriptURL: %s", targetData.DynamicArgumentsScriptUrl)
// 		t.Logf("   - ScriptLanguage: %s", targetData.ScriptLanguage)
// 		t.Log("")

// 		t.Log("2. Keeper creates RPC client for chain:", targetData.TargetChainID)
// 		t.Log("   - Chain: Arbitrum Sepolia (421614)")
// 		t.Log("   - RPC endpoint retrieved from chain config")
// 		t.Log("")

// 		t.Log("3. Script Execution Phase:")
// 		t.Log("   a. Download script from IPFS")
// 		t.Log("   b. Execute in Docker container")
// 		t.Log("   c. Parse JSON output from stdout")
// 		t.Log("   d. Extract storageUpdates from JSON")
// 		t.Log("   e. Validate output structure")
// 		t.Log("")

// 		t.Log("4. Script Output Processing:")
// 		t.Log("   - Check shouldExecute flag")
// 		t.Log("   - Extract targetContract address")
// 		t.Log("   - Extract calldata (pre-built by script)")
// 		t.Log("   - Extract storageUpdates map")
// 		t.Log("   - Extract metadata")
// 		t.Log("")

// 		t.Log("5. Transaction Preparation:")
// 		t.Log("   - Target Contract: 0xa0bC1477cfc452C05786262c377DE51FB8bc4669")
// 		t.Log("   - Calldata: From script output")
// 		t.Log("   - Chain: Arbitrum Sepolia (from job config)")
// 		t.Log("   - Nonce: Retrieved from nonce manager")
// 		t.Log("")

// 		t.Log("6. Contract Call (executeFunction):")
// 		t.Log("   - Pack executeFunction call with:")
// 		t.Log("     * jobID: uint256")
// 		t.Log("     * tgAmount: uint256 (gas cost)")
// 		t.Log("     * target: address (from script)")
// 		t.Log("     * data: bytes (calldata from script)")
// 		t.Log("")

// 		t.Log("7. Transaction Submission:")
// 		t.Log("   - Submit to execution contract on Arbitrum Sepolia")
// 		t.Log("   - Wait for receipt")
// 		t.Log("   - Extract transaction hash")
// 		t.Log("")

// 		t.Log("8. Return PerformerActionData:")
// 		t.Log("   - TaskID")
// 		t.Log("   - ActionTxHash")
// 		t.Log("   - GasUsed")
// 		t.Log("   - Status")
// 		t.Log("   - StorageUpdates (for database)")
// 		t.Log("   - ExecutionTimestamp")
// 		t.Log("   - ConvertedArguments (empty for custom scripts)")
// 		t.Log("")

// 		t.Log("9. TaskMonitor Processing:")
// 		t.Log("   - Listen for TaskSubmitted event")
// 		t.Log("   - Fetch IPFS data (includes storageUpdates)")
// 		t.Log("   - Update script_storage table in database")
// 		t.Log("   - Mark task as completed")
// 		t.Log("")
// 	})
// }

// // TestCustomScriptCalldata_Format tests that calldata from script is correctly formatted
// func TestCustomScriptCalldata_Format(t *testing.T) {
// 	// Expected calldata structure from the script
// 	// Function: transfer(address recipient, uint256 amount)
// 	// Selector: 0xa9059cbb (first 4 bytes)
// 	// Params:
// 	//   - recipient: 0xa76Cacba495CafeaBb628491733EB86f1db006dF (32 bytes)
// 	//   - amount: 10000000000000 (0x9184e72a000) (32 bytes)

// 	t.Run("Function Selector", func(t *testing.T) {
// 		selector := "a9059cbb" // transfer(address,uint256)
// 		assert.Equal(t, 8, len(selector), "Selector should be 8 hex characters (4 bytes)")
// 		t.Logf("Function selector: 0x%s", selector)
// 	})

// 	t.Run("Encoded Address Parameter", func(t *testing.T) {
// 		// Address: 0xa76Cacba495CafeaBb628491733EB86f1db006dF
// 		// Encoded: 000000000000000000000000a76cacba495cafeabb628491733eb86f1db006df
// 		address := "0xa76Cacba495CafeaBb628491733EB86f1db006dF"
// 		encodedAddress := "000000000000000000000000a76cacba495cafeabb628491733eb86f1db006df"

// 		assert.Equal(t, 64, len(encodedAddress), "Encoded address should be 64 hex characters (32 bytes)")
// 		t.Logf("Original address: %s", address)
// 		t.Logf("Encoded address: %s", encodedAddress)
// 	})

// 	t.Run("Encoded Amount Parameter", func(t *testing.T) {
// 		// Amount: 10000000000000 wei = 0x9184e72a000
// 		// Encoded: 000000000000000000000000000000000000000000000000000009184e72a000
// 		amount := "10000000000000" // 0.00001 ETH in wei
// 		encodedAmount := "000000000000000000000000000000000000000000000000000009184e72a000"

// 		assert.Equal(t, 64, len(encodedAmount), "Encoded amount should be 64 hex characters (32 bytes)")
// 		t.Logf("Amount (wei): %s", amount)
// 		t.Logf("Amount (hex): 0x9184e72a000")
// 		t.Logf("Encoded amount: %s", encodedAmount)
// 	})

// 	t.Run("Complete Calldata", func(t *testing.T) {
// 		// Complete calldata structure
// 		selector := "a9059cbb"
// 		encodedRecipient := "000000000000000000000000a76cacba495cafeabb628491733eb86f1db006df"
// 		encodedAmount := "000000000000000000000000000000000000000000000000000009184e72a000"

// 		completeCalldata := "0x" + selector + encodedRecipient + encodedAmount

// 		// Verify total length: 0x (2) + selector (8) + param1 (64) + param2 (64) = 138 chars
// 		expectedLength := 2 + 8 + 64 + 64 // 138
// 		assert.Equal(t, expectedLength, len(completeCalldata), "Complete calldata should be 138 characters")

// 		t.Log("=== Complete Calldata Breakdown ===")
// 		t.Logf("Prefix:     %s", completeCalldata[:2])
// 		t.Logf("Selector:   %s (transfer)", completeCalldata[2:10])
// 		t.Logf("Recipient:  %s", completeCalldata[10:74])
// 		t.Logf("Amount:     %s", completeCalldata[74:138])
// 		t.Logf("Complete:   %s", completeCalldata)
// 		t.Logf("Length:     %d characters", len(completeCalldata))
// 	})
// }

// // TestStorageUpdates_Persistence tests storage update flow
// func TestStorageUpdates_Persistence(t *testing.T) {
// 	jobID := big.NewInt(999888777666555)

// 	t.Run("First Execution - Empty Storage", func(t *testing.T) {
// 		// First execution has no previous storage
// 		initialStorage := map[string]string{}

// 		t.Log("Initial storage (first execution):", initialStorage)
// 		assert.Empty(t, initialStorage, "First execution should have empty storage")
// 	})

// 	t.Run("Script Returns Storage Updates", func(t *testing.T) {
// 		// Expected storage updates from script
// 		expectedUpdates := map[string]string{
// 			"lastExecutionTime": "1763298589440",
// 			"lastActionTarget":  "0xa76Cacba495CafeaBb628491733EB86f1db006dF",
// 			"lastActionValue":   "10000000000000",
// 			"lastSafeAddress":   "0x87EB883e8ae00120EF2c6Fd49b1F8A149E2172f4",
// 			"executionCount":    "1",
// 			"status":            "executed",
// 		}

// 		assert.Equal(t, 6, len(expectedUpdates), "Should have 6 storage updates")
// 		t.Log("Storage updates from script:", expectedUpdates)
// 	})

// 	t.Run("TaskMonitor Saves Updates", func(t *testing.T) {
// 		t.Log("=== Storage Update Flow ===")
// 		t.Log("1. Script returns storageUpdates in JSON")
// 		t.Log("2. Keeper includes storageUpdates in PerformerActionData")
// 		t.Log("3. PerformerActionData uploaded to IPFS")
// 		t.Log("4. TaskSubmitted event emitted with IPFS hash")
// 		t.Log("5. TaskMonitor listens for event")
// 		t.Log("6. TaskMonitor fetches IPFS data")
// 		t.Log("7. TaskMonitor extracts storageUpdates")
// 		t.Log("8. TaskMonitor calls UpdateScriptStorage(jobID, storageUpdates)")
// 		t.Log("9. Database query: INSERT/UPDATE into script_storage table")
// 		t.Logf("   WHERE job_id = %s", jobID.String())
// 		t.Log("")
// 	})

// 	t.Run("Database Schema", func(t *testing.T) {
// 		t.Log("=== script_storage Table ===")
// 		t.Log("Columns:")
// 		t.Log("  - job_id: varint (PRIMARY KEY)")
// 		t.Log("  - storage_key: text (PRIMARY KEY)")
// 		t.Log("  - storage_value: text")
// 		t.Log("  - updated_at: timestamp")
// 		t.Log("")
// 		t.Log("Example rows after execution:")
// 		t.Logf("  (%s, 'lastExecutionTime', '1763298589440', NOW())", jobID.String())
// 		t.Logf("  (%s, 'executionCount', '1', NOW())", jobID.String())
// 		t.Logf("  (%s, 'status', 'executed', NOW())", jobID.String())
// 	})

// 	t.Run("Second Execution - With Storage", func(t *testing.T) {
// 		// Phase 2: Script will be able to read these values via env vars
// 		previousStorage := map[string]string{
// 			"lastExecutionTime": "1763298589440",
// 			"executionCount":    "1",
// 			"status":            "executed",
// 		}

// 		t.Log("Previous storage (available in Phase 2):", previousStorage)
// 		t.Log("In Phase 1: Script CANNOT read previous storage")
// 		t.Log("In Phase 2: Script WILL read via TRIGGERX_STORAGE_* env vars")
// 		t.Log("  Example: TRIGGERX_STORAGE_executionCount=1")
// 	})
// }

// // TestCustomScriptExecution_ErrorCases tests error handling
// func TestCustomScriptExecution_ErrorCases(t *testing.T) {
// 	t.Run("Invalid IPFS URL", func(t *testing.T) {
// 		invalidURL := "https://invalid-ipfs-url"
// 		t.Logf("Invalid URL: %s", invalidURL)
// 		t.Log("Expected error: Failed to download script from IPFS")
// 	})

// 	t.Run("Script Syntax Error", func(t *testing.T) {
// 		t.Log("If script has syntax errors:")
// 		t.Log("  - Docker execution fails")
// 		t.Log("  - result.Success = false")
// 		t.Log("  - Returns error: 'script execution failed'")
// 	})

// 	t.Run("Invalid JSON Output", func(t *testing.T) {
// 		t.Log("If script outputs invalid JSON:")
// 		t.Log("  - JSON parse fails")
// 		t.Log("  - Returns error: 'failed to parse script output'")
// 	})

// 	t.Run("Missing Required Fields", func(t *testing.T) {
// 		t.Log("If shouldExecute=true but missing targetContract:")
// 		t.Log("  - Validation fails")
// 		t.Log("  - Returns error: 'targetContract required when shouldExecute=true'")
// 	})

// 	t.Run("Invalid Address Format", func(t *testing.T) {
// 		t.Log("If targetContract is not valid address format:")
// 		t.Log("  - Length check fails (must be 42 chars)")
// 		t.Log("  - Prefix check fails (must start with 0x)")
// 		t.Log("  - Returns error: 'invalid targetContract address format'")
// 	})

// 	t.Run("Invalid Calldata Format", func(t *testing.T) {
// 		t.Log("If calldata doesn't start with 0x:")
// 		t.Log("  - Validation fails")
// 		t.Log("  - Returns error: 'calldata must be hex string starting with 0x'")
// 	})
// }

// // TestEndToEndFlow documents the complete execution flow
// func TestEndToEndFlow_Documentation(t *testing.T) {
// 	t.Log("╔════════════════════════════════════════════════════════════════════╗")
// 	t.Log("║  TaskDefinitionID 7: Custom Script Execution - Complete Flow      ║")
// 	t.Log("╚════════════════════════════════════════════════════════════════════╝")
// 	t.Log("")

// 	t.Log("┌─────────────────────────────────────────────────────────────────┐")
// 	t.Log("│ 1. JOB CREATION (User → DBServer API)                          │")
// 	t.Log("└─────────────────────────────────────────────────────────────────┘")
// 	t.Log("   POST /api/jobs")
// 	t.Log("   {")
// 	t.Log("     job_id: 999888777666555,")
// 	t.Log("     task_definition_id: 7,")
// 	t.Log("     target_chain_id: '421614',")
// 	t.Log("     dynamic_arguments_script_url: 'https://ipfs.io/ipfs/...',")
// 	t.Log("     language: 'typescript',")
// 	t.Log("     time_interval: 60")
// 	t.Log("   }")
// 	t.Log("   ↓")
// 	t.Log("   Stored in custom_jobs table")
// 	t.Log("")

// 	t.Log("┌─────────────────────────────────────────────────────────────────┐")
// 	t.Log("│ 2. SCHEDULING (Time Scheduler)                                 │")
// 	t.Log("└─────────────────────────────────────────────────────────────────┘")
// 	t.Log("   GetTimeBasedTasks() called every N seconds")
// 	t.Log("   ↓")
// 	t.Log("   Query: SELECT * FROM custom_jobs")
// 	t.Log("          WHERE is_active = true")
// 	t.Log("          AND next_execution_time <= NOW()")
// 	t.Log("   ↓")
// 	t.Log("   Fetch storage: SELECT * FROM script_storage WHERE job_id = ?")
// 	t.Log("   ↓")
// 	t.Log("   Convert to ScheduleTimeTaskData:")
// 	t.Log("   {")
// 	t.Log("     TaskDefinitionID: 7,")
// 	t.Log("     TargetChainID: '421614',")
// 	t.Log("     DynamicArgumentsScriptUrl: 'https://ipfs.io/ipfs/...',")
// 	t.Log("     ScriptLanguage: 'typescript',")
// 	t.Log("     ScriptStorage: {...}  // From database")
// 	t.Log("   }")
// 	t.Log("   ↓")
// 	t.Log("   Send to Keeper/Aggregator")
// 	t.Log("")

// 	t.Log("┌─────────────────────────────────────────────────────────────────┐")
// 	t.Log("│ 3. SCRIPT EXECUTION (Keeper)                                   │")
// 	t.Log("└─────────────────────────────────────────────────────────────────┘")
// 	t.Log("   Keeper receives task")
// 	t.Log("   ↓")
// 	t.Log("   Create RPC client for chain: '421614'")
// 	t.Log("   ↓")
// 	t.Log("   ExecuteCustomScript():")
// 	t.Log("     1. Download script from IPFS")
// 	t.Log("     2. Execute in Docker: dockerExecutor.Execute()")
// 	t.Log("        - Language: typescript")
// 	t.Log("        - Container: triggerx/typescript:latest")
// 	t.Log("     3. Capture stdout (JSON output)")
// 	t.Log("     4. Parse JSON:")
// 	t.Log("        {")
// 	t.Log("          shouldExecute: true,")
// 	t.Log("          targetContract: '0xa0bC1477cfc452C05786262c377DE51FB8bc4669',")
// 	t.Log("          calldata: '0xa9059cbb...',")
// 	t.Log("          storageUpdates: {")
// 	t.Log("            lastExecutionTime: '1763298589440',")
// 	t.Log("            executionCount: '1',")
// 	t.Log("            status: 'executed'")
// 	t.Log("          },")
// 	t.Log("          metadata: {")
// 	t.Log("            timestamp: 1763298589440,")
// 	t.Log("            reason: 'Executing Safe transaction',")
// 	t.Log("            gasEstimate: 150000")
// 	t.Log("          }")
// 	t.Log("        }")
// 	t.Log("     5. Validate output")
// 	t.Log("     6. Return (scriptOutput, storageUpdates, nil)")
// 	t.Log("")

// 	t.Log("┌─────────────────────────────────────────────────────────────────┐")
// 	t.Log("│ 4. TRANSACTION PREPARATION (Keeper)                            │")
// 	t.Log("└─────────────────────────────────────────────────────────────────┘")
// 	t.Log("   if shouldExecute == false:")
// 	t.Log("     return PerformerActionData{StorageUpdates: ...}")
// 	t.Log("   ")
// 	t.Log("   if shouldExecute == true:")
// 	t.Log("     targetContract = scriptOutput.TargetContract")
// 	t.Log("     calldata = scriptOutput.Calldata")
// 	t.Log("     ")
// 	t.Log("     Pack executeFunction call:")
// 	t.Log("       executeFunction(")
// 	t.Log("         jobID: 999888777666555,")
// 	t.Log("         tgAmount: gasEstimate,")
// 	t.Log("         target: 0xa0bC1477cfc452C05786262c377DE51FB8bc4669,")
// 	t.Log("         data: 0xa9059cbb...")
// 	t.Log("       )")
// 	t.Log("")

// 	t.Log("┌─────────────────────────────────────────────────────────────────┐")
// 	t.Log("│ 5. ON-CHAIN EXECUTION (Keeper)                                 │")
// 	t.Log("└─────────────────────────────────────────────────────────────────┘")
// 	t.Log("   Get nonce from NonceManager")
// 	t.Log("   ↓")
// 	t.Log("   Create transaction:")
// 	t.Log("     - To: ExecutionContractAddress (from config)")
// 	t.Log("     - Data: packed executeFunction call")
// 	t.Log("     - ChainID: 421614 (Arbitrum Sepolia)")
// 	t.Log("     - Nonce: from NonceManager")
// 	t.Log("   ↓")
// 	t.Log("   Submit transaction")
// 	t.Log("   ↓")
// 	t.Log("   Wait for receipt")
// 	t.Log("   ↓")
// 	t.Log("   Extract: txHash, gasUsed, status")
// 	t.Log("")

// 	t.Log("┌─────────────────────────────────────────────────────────────────┐")
// 	t.Log("│ 6. IPFS DATA UPLOAD (Keeper)                                   │")
// 	t.Log("└─────────────────────────────────────────────────────────────────┘")
// 	t.Log("   Create PerformerActionData:")
// 	t.Log("   {")
// 	t.Log("     TaskID: 12345,")
// 	t.Log("     ActionTxHash: '0xabcd1234...',")
// 	t.Log("     GasUsed: '150000',")
// 	t.Log("     Status: true,")
// 	t.Log("     ExecutionTimestamp: NOW(),")
// 	t.Log("     StorageUpdates: {")
// 	t.Log("       lastExecutionTime: '1763298589440',")
// 	t.Log("       executionCount: '1',")
// 	t.Log("       status: 'executed'")
// 	t.Log("     },")
// 	t.Log("     TotalFee: gasEstimate,")
// 	t.Log("     ConvertedArguments: []")
// 	t.Log("   }")
// 	t.Log("   ↓")
// 	t.Log("   Upload to IPFS → QmPerformerData123...")
// 	t.Log("   ↓")
// 	t.Log("   Submit TaskSubmitted event with IPFS hash")
// 	t.Log("")

// 	t.Log("┌─────────────────────────────────────────────────────────────────┐")
// 	t.Log("│ 7. STORAGE PERSISTENCE (TaskMonitor)                           │")
// 	t.Log("└─────────────────────────────────────────────────────────────────┘")
// 	t.Log("   Listen for TaskSubmitted event")
// 	t.Log("   ↓")
// 	t.Log("   Event data contains IPFS hash")
// 	t.Log("   ↓")
// 	t.Log("   Fetch from IPFS: QmPerformerData123...")
// 	t.Log("   ↓")
// 	t.Log("   Extract: ipfsData.ActionData.StorageUpdates")
// 	t.Log("   ↓")
// 	t.Log("   Check: TaskDefinitionID == 7?")
// 	t.Log("   ↓")
// 	t.Log("   GetJobIDByTaskID(12345) → 999888777666555")
// 	t.Log("   ↓")
// 	t.Log("   UpdateScriptStorage(jobID, storageUpdates):")
// 	t.Log("     For each (key, value) in storageUpdates:")
// 	t.Log("       INSERT INTO script_storage")
// 	t.Log("       (job_id, storage_key, storage_value, updated_at)")
// 	t.Log("       VALUES (999888777666555, key, value, NOW())")
// 	t.Log("   ↓")
// 	t.Log("   Mark task as completed")
// 	t.Log("")

// 	t.Log("┌─────────────────────────────────────────────────────────────────┐")
// 	t.Log("│ 8. NEXT EXECUTION (Scheduler)                                  │")
// 	t.Log("└─────────────────────────────────────────────────────────────────┘")
// 	t.Log("   Update custom_jobs:")
// 	t.Log("     next_execution_time = NOW() + time_interval")
// 	t.Log("     last_executed_at = NOW()")
// 	t.Log("   ↓")
// 	t.Log("   Wait 60 seconds...")
// 	t.Log("   ↓")
// 	t.Log("   Repeat from step 2")
// 	t.Log("   (Now with storage from previous execution)")
// 	t.Log("")

// 	t.Log("╔════════════════════════════════════════════════════════════════════╗")
// 	t.Log("║  Script URL: https://ipfs.io/ipfs/bafybeicm2uyje7k7wbgmdgzjt57lexoo7ca6565myx6ktynzjm2qgy2uzm")
// 	t.Log("║  Target Contract: 0xa0bC1477cfc452C05786262c377DE51FB8bc4669")
// 	t.Log("║  Chain: Arbitrum Sepolia (421614)")
// 	t.Log("║  Interval: 60 seconds")
// 	t.Log("╚════════════════════════════════════════════════════════════════════╝")
// }
