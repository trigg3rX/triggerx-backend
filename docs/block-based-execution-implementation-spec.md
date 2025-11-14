# Block-Based Script Execution - Implementation Specification

## Executive Summary

This document provides detailed specifications for implementing per-block script execution in TriggerX, addressing three key questions:

1. **Script Format**: How users provide scripts and what format to expect
2. **Keeper Code**: How keepers execute and verify scripts
3. **Verification Integration**: How script execution integrates with existing AttestationCenter

**Approach**: Start with basic execution and verification, add fraud proofs and slashing later.

---

## Table of Contents

1. [Script Format Specification](#1-script-format-specification)
2. [Keeper Execution Architecture](#2-keeper-execution-architecture)
3. [Verification with AttestationCenter](#3-verification-with-attestationcenter)
4. [Data Structures](#4-data-structures)
5. [Implementation Phases](#5-implementation-phases)
6. [Code Examples](#6-code-examples)

---

## 1. Script Format Specification

### 1.1 User-Provided Script Structure

Users provide scripts in a **standardized directory structure**:

```
my-block-script/
├── index.go                 # Main script (or index.py, index.ts, index.js)
├── script.json              # Configuration file
├── .env.example             # Example environment variables
└── README.md                # Optional documentation
```

#### Option A: IPFS CID (Recommended)

```bash
# User uploads to IPFS
$ ipfs add -r my-block-script/
added QmHash123... my-block-script/index.go
added QmHash456... my-block-script/script.json
added QmHash789... my-block-script

# User provides CID when creating task
CID: QmHash789...
```

#### Option B: GitHub URL (Alternative)

```
URL: https://raw.githubusercontent.com/user/repo/main/scripts/my-script/index.go
```

---

### 1.2 Script Configuration (`script.json`)

Every script requires a `script.json` configuration file:

```json
{
  "version": "1.0",
  "runtime": {
    "language": "go",
    "memory": 256,
    "timeout": 30
  },
  "execution": {
    "cadence": 1,
    "deterministic": true
  },
  "outputs": {
    "format": "json",
    "schema": {
      "shouldExecute": "boolean",
      "params": "array",
      "metadata": "object"
    }
  },
  "dependencies": {
    "packages": [],
    "preInstall": ""
  },
  "secrets": ["API_KEY", "RPC_URL"],
  "storage": {
    "enabled": true,
    "keys": ["lastBlock", "lastPrice"]
  }
}
```

**Fields:**

| Field | Type | Description | Required |
|-------|------|-------------|----------|
| `version` | string | Script format version | ✅ Yes |
| `runtime.language` | string | `"go"`, `"python"`, `"typescript"`, `"javascript"` | ✅ Yes |
| `runtime.memory` | number | Memory in MB (128-512) | ✅ Yes |
| `runtime.timeout` | number | Timeout in seconds (10-60) | ✅ Yes |
| `execution.cadence` | number | Execute every N blocks (1 = every block) | ✅ Yes |
| `execution.deterministic` | boolean | If true, use deterministic inputs | ✅ Yes |
| `outputs.format` | string | Output format: `"json"` | ✅ Yes |
| `outputs.schema` | object | Expected output structure | ✅ Yes |
| `dependencies.packages` | array | List of dependencies | ❌ No |
| `dependencies.preInstall` | string | Pre-install command | ❌ No |
| `secrets` | array | Required secret keys | ❌ No |
| `storage.enabled` | boolean | Enable persistent storage | ❌ No |
| `storage.keys` | array | Storage keys used | ❌ No |

---

### 1.3 Script Code Structure

#### **1.3.1 Go Script Example**

```go
package main

import (
    "encoding/json"
    "fmt"
    "os"
    "strconv"
)

// Output structure - REQUIRED format
type ScriptOutput struct {
    ShouldExecute bool          `json:"shouldExecute"`
    Params        []interface{} `json:"params"`
    Metadata      Metadata      `json:"metadata"`
}

type Metadata struct {
    BlockNumber uint64 `json:"blockNumber"`
    Timestamp   uint64 `json:"timestamp"`
    Reason      string `json:"reason"`
}

// Context - Injected by TriggerX keeper
type ExecutionContext struct {
    BlockNumber uint64
    Timestamp   uint64
    JobID       string
    ChainID     string
}

func main() {
    // 1. Get execution context (injected as env vars)
    ctx := getContext()

    // 2. Get secrets (if needed)
    apiKey := os.Getenv("SECRET_API_KEY")
    if apiKey == "" {
        exitWithError("API_KEY not configured")
    }

    // 3. Get storage (if enabled)
    lastBlock := getStorage("lastBlock")
    if lastBlock == "" {
        lastBlock = "0"
    }

    // 4. Your custom logic here
    shouldExecute, params, reason := checkCondition(ctx, apiKey, lastBlock)

    // 5. Update storage
    setStorage("lastBlock", strconv.FormatUint(ctx.BlockNumber, 10))

    // 6. Output result in required JSON format
    output := ScriptOutput{
        ShouldExecute: shouldExecute,
        Params:        params,
        Metadata: Metadata{
            BlockNumber: ctx.BlockNumber,
            Timestamp:   ctx.Timestamp,
            Reason:      reason,
        },
    }

    outputJSON, _ := json.Marshal(output)
    fmt.Println(string(outputJSON))
}

// Helper: Get execution context from environment variables
func getContext() ExecutionContext {
    blockNumber, _ := strconv.ParseUint(os.Getenv("TRIGGERX_BLOCK_NUMBER"), 10, 64)
    timestamp, _ := strconv.ParseUint(os.Getenv("TRIGGERX_TIMESTAMP"), 10, 64)

    return ExecutionContext{
        BlockNumber: blockNumber,
        Timestamp:   timestamp,
        JobID:       os.Getenv("TRIGGERX_JOB_ID"),
        ChainID:     os.Getenv("TRIGGERX_CHAIN_ID"),
    }
}

// Helper: Get from persistent storage
func getStorage(key string) string {
    // TriggerX provides storage via HTTP API
    // For now, use env var: TRIGGERX_STORAGE_<KEY>
    return os.Getenv("TRIGGERX_STORAGE_" + key)
}

// Helper: Set persistent storage
func setStorage(key, value string) {
    // Output storage updates in special format
    fmt.Fprintf(os.Stderr, "STORAGE_SET:%s=%s\n", key, value)
}

// Helper: Exit with error
func exitWithError(msg string) {
    output := ScriptOutput{
        ShouldExecute: false,
        Metadata: Metadata{
            Reason: msg,
        },
    }
    outputJSON, _ := json.Marshal(output)
    fmt.Println(string(outputJSON))
    os.Exit(0)
}

// Your custom logic
func checkCondition(ctx ExecutionContext, apiKey, lastBlock string) (bool, []interface{}, string) {
    // Example: Check if price has changed significantly
    // This is where your business logic goes

    // Dummy example:
    if ctx.BlockNumber%10 == 0 {
        return true, []interface{}{ctx.BlockNumber, "0x123"}, "Block divisible by 10"
    }

    return false, nil, "Condition not met"
}
```

#### **1.3.2 Python Script Example**

```python
#!/usr/bin/env python3
import json
import os
import sys

class ExecutionContext:
    def __init__(self):
        self.block_number = int(os.getenv('TRIGGERX_BLOCK_NUMBER', '0'))
        self.timestamp = int(os.getenv('TRIGGERX_TIMESTAMP', '0'))
        self.job_id = os.getenv('TRIGGERX_JOB_ID', '')
        self.chain_id = os.getenv('TRIGGERX_CHAIN_ID', '')

def get_storage(key):
    """Get value from persistent storage"""
    return os.getenv(f'TRIGGERX_STORAGE_{key}', '')

def set_storage(key, value):
    """Set value in persistent storage"""
    print(f'STORAGE_SET:{key}={value}', file=sys.stderr)

def main():
    # 1. Get execution context
    ctx = ExecutionContext()

    # 2. Get secrets
    api_key = os.getenv('SECRET_API_KEY')
    if not api_key:
        output = {
            'shouldExecute': False,
            'params': [],
            'metadata': {
                'reason': 'API_KEY not configured'
            }
        }
        print(json.dumps(output))
        return

    # 3. Get storage
    last_block = get_storage('lastBlock') or '0'

    # 4. Your custom logic
    should_execute, params, reason = check_condition(ctx, api_key, last_block)

    # 5. Update storage
    set_storage('lastBlock', str(ctx.block_number))

    # 6. Output result
    output = {
        'shouldExecute': should_execute,
        'params': params,
        'metadata': {
            'blockNumber': ctx.block_number,
            'timestamp': ctx.timestamp,
            'reason': reason
        }
    }

    print(json.dumps(output))

def check_condition(ctx, api_key, last_block):
    """Your custom logic here"""
    # Example: Execute every 10 blocks
    if ctx.block_number % 10 == 0:
        return True, [ctx.block_number, '0x123'], 'Block divisible by 10'

    return False, [], 'Condition not met'

if __name__ == '__main__':
    main()
```

#### **1.3.3 TypeScript Script Example**

```typescript
import * as fs from 'fs';

interface ScriptOutput {
  shouldExecute: boolean;
  params: any[];
  metadata: {
    blockNumber: number;
    timestamp: number;
    reason: string;
  };
}

interface ExecutionContext {
  blockNumber: number;
  timestamp: number;
  jobId: string;
  chainId: string;
}

function getContext(): ExecutionContext {
  return {
    blockNumber: parseInt(process.env.TRIGGERX_BLOCK_NUMBER || '0'),
    timestamp: parseInt(process.env.TRIGGERX_TIMESTAMP || '0'),
    jobId: process.env.TRIGGERX_JOB_ID || '',
    chainId: process.env.TRIGGERX_CHAIN_ID || '',
  };
}

function getStorage(key: string): string {
  return process.env[`TRIGGERX_STORAGE_${key}`] || '';
}

function setStorage(key: string, value: string): void {
  console.error(`STORAGE_SET:${key}=${value}`);
}

async function main() {
  // 1. Get execution context
  const ctx = getContext();

  // 2. Get secrets
  const apiKey = process.env.SECRET_API_KEY;
  if (!apiKey) {
    const output: ScriptOutput = {
      shouldExecute: false,
      params: [],
      metadata: {
        blockNumber: ctx.blockNumber,
        timestamp: ctx.timestamp,
        reason: 'API_KEY not configured',
      },
    };
    console.log(JSON.stringify(output));
    return;
  }

  // 3. Get storage
  const lastBlock = getStorage('lastBlock') || '0';

  // 4. Your custom logic
  const { shouldExecute, params, reason } = checkCondition(ctx, apiKey, lastBlock);

  // 5. Update storage
  setStorage('lastBlock', ctx.blockNumber.toString());

  // 6. Output result
  const output: ScriptOutput = {
    shouldExecute,
    params,
    metadata: {
      blockNumber: ctx.blockNumber,
      timestamp: ctx.timestamp,
      reason,
    },
  };

  console.log(JSON.stringify(output));
}

function checkCondition(
  ctx: ExecutionContext,
  apiKey: string,
  lastBlock: string
): { shouldExecute: boolean; params: any[]; reason: string } {
  // Your custom logic here
  if (ctx.blockNumber % 10 === 0) {
    return {
      shouldExecute: true,
      params: [ctx.blockNumber, '0x123'],
      reason: 'Block divisible by 10',
    };
  }

  return {
    shouldExecute: false,
    params: [],
    reason: 'Condition not met',
  };
}

main().catch((error) => {
  console.error('Script error:', error);
  process.exit(1);
});
```

---

### 1.4 Output Format (REQUIRED)

**All scripts MUST output JSON to stdout in this format:**

```json
{
  "shouldExecute": true,
  "params": [
    "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb",
    1000000000000000000,
    "0xabcdef..."
  ],
  "metadata": {
    "blockNumber": 12345678,
    "timestamp": 1709876543,
    "reason": "Price threshold exceeded",
    "gasEstimate": 150000,
    "apiCalls": [
      {
        "url": "https://api.coingecko.com/...",
        "status": 200
      }
    ]
  }
}
```

**Field Definitions:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `shouldExecute` | boolean | ✅ Yes | Whether to execute the transaction |
| `params` | array | ✅ Yes | Parameters to pass to target function (can be empty if shouldExecute=false) |
| `metadata` | object | ✅ Yes | Additional execution metadata |
| `metadata.blockNumber` | number | ✅ Yes | Block number when executed |
| `metadata.timestamp` | number | ✅ Yes | Timestamp when executed |
| `metadata.reason` | string | ✅ Yes | Human-readable reason for decision |
| `metadata.gasEstimate` | number | ❌ No | Estimated gas for transaction |
| `metadata.apiCalls` | array | ❌ No | API calls made (for debugging) |

**Error Output:**

If script encounters an error:

```json
{
  "shouldExecute": false,
  "params": [],
  "metadata": {
    "blockNumber": 12345678,
    "timestamp": 1709876543,
    "reason": "Error: API rate limit exceeded",
    "error": true
  }
}
```

---

### 1.5 Environment Variables (Injected by Keeper)

**Standard Context Variables:**

| Variable | Type | Example | Description |
|----------|------|---------|-------------|
| `TRIGGERX_BLOCK_NUMBER` | uint64 | `12345678` | Current block number |
| `TRIGGERX_TIMESTAMP` | uint64 | `1709876543` | Current block timestamp |
| `TRIGGERX_JOB_ID` | string | `"0x123..."` | Job ID for this task |
| `TRIGGERX_CHAIN_ID` | string | `"84532"` | Chain ID (Base Sepolia) |
| `TRIGGERX_EXECUTION_MODE` | string | `"performer"` or `"validator"` | Who is executing |

**Secret Variables (User-configured):**

| Variable | Example | Description |
|----------|---------|-------------|
| `SECRET_API_KEY` | `sk_test_123` | User's API key |
| `SECRET_RPC_URL` | `https://...` | Custom RPC endpoint |
| `SECRET_*` | Any value | Any user-defined secret |

**Storage Variables (Read-only):**

| Variable | Example | Description |
|----------|---------|-------------|
| `TRIGGERX_STORAGE_lastBlock` | `"12345"` | Previous value of `lastBlock` |
| `TRIGGERX_STORAGE_lastPrice` | `"2500.50"` | Previous value of `lastPrice` |
| `TRIGGERX_STORAGE_*` | Any value | Any stored key-value |

**Storage Updates (Write via stderr):**

Scripts write storage updates to stderr:

```
STORAGE_SET:lastBlock=12345678
STORAGE_SET:lastPrice=2500.50
```

Keeper parses these and persists to database.

---

### 1.6 Script Validation Rules

Before accepting a script, TriggerX validates:

**1. Configuration File:**
- ✅ `script.json` exists
- ✅ All required fields present
- ✅ Valid language (`go`, `python`, `typescript`, `javascript`)
- ✅ Memory within limits (128-512 MB)
- ✅ Timeout within limits (10-60 seconds)

**2. Code Structure:**
- ✅ Entry point exists (`index.go`, `index.py`, `index.ts`, `index.js`)
- ✅ No malicious patterns (basic static analysis)
- ✅ File size < 10 MB

**3. Complexity Analysis:**
- ✅ Complexity score < threshold (use existing `file/validator.go` logic)

**4. Output Format:**
- ✅ Script outputs valid JSON
- ✅ JSON contains required fields (`shouldExecute`, `params`, `metadata`)

---

## 2. Keeper Execution Architecture

### 2.1 New Task Definition Type

Add to existing task types:

```go
const (
    TaskDefinitionTimeStatic        = 1  // Existing
    TaskDefinitionTimeDynamic       = 2  // Existing
    TaskDefinitionEventStatic       = 3  // Existing
    TaskDefinitionEventDynamic      = 4  // Existing
    TaskDefinitionConditionStatic   = 5  // Existing
    TaskDefinitionConditionDynamic  = 6  // Existing
    TaskDefinitionBlockBased        = 7  // NEW: Block-based execution
)
```

### 2.2 Enhanced TaskTargetData

Update `pkg/types/schedulers.go`:

```go
type TaskTargetData struct {
    JobID                     *BigInt `json:"job_id"`
    TaskID                    int64   `json:"task_id"`
    TaskDefinitionID          int     `json:"task_definition_id"`
    TargetChainID             string  `json:"target_chain_id"`
    TargetContractAddress     string  `json:"target_contract_address"`
    TargetFunction            string  `json:"target_function"`
    ABI                       string  `json:"abi"`
    ArgType                   int     `json:"arg_type"`
    Arguments                 []string `json:"arguments"`
    DynamicArgumentsScriptUrl string  `json:"dynamic_arguments_script_url"`
    IsImua                    bool    `json:"is_imua"`

    // NEW: Block-based execution fields
    ScriptSource              string  `json:"script_source"`  // "ipfs" or "url"
    ScriptCID                 string  `json:"script_cid"`     // IPFS CID if source=ipfs
    ScriptURL                 string  `json:"script_url"`     // URL if source=url
    ScriptLanguage            string  `json:"script_language"` // "go", "python", etc.
    ExecutionCadence          uint64  `json:"execution_cadence"` // Execute every N blocks
    LastExecutedBlock         uint64  `json:"last_executed_block"` // Track last execution
}
```

### 2.3 Keeper Block Subscription

**New Component: `internal/keeper/core/blockmonitor/monitor.go`**

```go
package blockmonitor

import (
    "context"
    "time"

    "github.com/ethereum/go-ethereum/ethclient"
    "github.com/trigg3rX/triggerx-backend/pkg/logging"
)

type BlockMonitor struct {
    client  *ethclient.Client
    logger  logging.Logger
    chainID string

    handlers []BlockHandler
}

type BlockHandler interface {
    OnNewBlock(ctx context.Context, blockNumber uint64, timestamp uint64) error
}

func NewBlockMonitor(rpcURL string, chainID string, logger logging.Logger) (*BlockMonitor, error) {
    client, err := ethclient.Dial(rpcURL)
    if err != nil {
        return nil, err
    }

    return &BlockMonitor{
        client:  client,
        logger:  logger,
        chainID: chainID,
        handlers: []BlockHandler{},
    }, nil
}

func (bm *BlockMonitor) RegisterHandler(handler BlockHandler) {
    bm.handlers = append(bm.handlers, handler)
}

func (bm *BlockMonitor) Start(ctx context.Context) error {
    bm.logger.Info("Starting block monitor for chain", bm.chainID)

    // Subscribe to new block headers
    headers := make(chan *types.Header)
    sub, err := bm.client.SubscribeNewHead(ctx, headers)
    if err != nil {
        return fmt.Errorf("failed to subscribe to new heads: %w", err)
    }
    defer sub.Unsubscribe()

    for {
        select {
        case err := <-sub.Err():
            bm.logger.Errorf("Block subscription error: %v", err)
            return err

        case header := <-headers:
            blockNumber := header.Number.Uint64()
            timestamp := header.Time

            bm.logger.Debugf("New block: %d (timestamp: %d)", blockNumber, timestamp)

            // Notify all handlers
            for _, handler := range bm.handlers {
                go func(h BlockHandler) {
                    if err := h.OnNewBlock(ctx, blockNumber, timestamp); err != nil {
                        bm.logger.Errorf("Handler error: %v", err)
                    }
                }(handler)
            }

        case <-ctx.Done():
            bm.logger.Info("Block monitor stopped")
            return nil
        }
    }
}
```

### 2.4 Block-Based Task Executor

**New Component: `internal/keeper/core/execution/block_executor.go`**

```go
package execution

import (
    "context"
    "encoding/json"
    "fmt"
    "time"

    "github.com/trigg3rX/triggerx-backend/pkg/types"
    dockertypes "github.com/trigg3rX/triggerx-backend/pkg/dockerexecutor/types"
)

type BlockBasedExecutor struct {
    taskExecutor *TaskExecutor
    logger       logging.Logger
}

// OnNewBlock implements BlockHandler interface
func (bbe *BlockBasedExecutor) OnNewBlock(ctx context.Context, blockNumber uint64, timestamp uint64) error {
    // 1. Get all block-based tasks for this chain
    tasks, err := bbe.getBlockBasedTasks()
    if err != nil {
        return fmt.Errorf("failed to get block-based tasks: %w", err)
    }

    bbe.logger.Infof("Processing %d block-based tasks for block %d", len(tasks), blockNumber)

    // 2. Execute each task
    for _, task := range tasks {
        // Check if we should execute based on cadence
        if !bbe.shouldExecuteThisBlock(task, blockNumber) {
            bbe.logger.Debugf("Skipping task %d (cadence: every %d blocks)", task.TaskID, task.ExecutionCadence)
            continue
        }

        // Execute in goroutine for parallel processing
        go bbe.executeBlockTask(ctx, task, blockNumber, timestamp)
    }

    return nil
}

func (bbe *BlockBasedExecutor) shouldExecuteThisBlock(task *types.TaskTargetData, currentBlock uint64) bool {
    if task.ExecutionCadence == 0 {
        task.ExecutionCadence = 1 // Default: every block
    }

    // Execute every N blocks
    blocksSinceLastExecution := currentBlock - task.LastExecutedBlock
    return blocksSinceLastExecution >= task.ExecutionCadence
}

func (bbe *BlockBasedExecutor) executeBlockTask(
    ctx context.Context,
    task *types.TaskTargetData,
    blockNumber uint64,
    timestamp uint64,
) {
    bbe.logger.Infof("Executing block task %d at block %d", task.TaskID, blockNumber)

    // 1. Prepare execution context
    execCtx := bbe.prepareExecutionContext(task, blockNumber, timestamp)

    // 2. Execute script in Docker
    result, err := bbe.executeScript(ctx, task, execCtx)
    if err != nil {
        bbe.logger.Errorf("Script execution failed for task %d: %v", task.TaskID, err)
        return
    }

    // 3. Parse script output
    output, err := bbe.parseScriptOutput(result.Output)
    if err != nil {
        bbe.logger.Errorf("Failed to parse script output for task %d: %v", task.TaskID, err)
        return
    }

    // 4. Check if we should execute
    if !output.ShouldExecute {
        bbe.logger.Infof("Task %d: shouldExecute=false, reason: %s", task.TaskID, output.Metadata.Reason)
        return
    }

    // 5. Generate execution proof
    proof := bbe.generateExecutionProof(task, execCtx, output, result)

    // 6. Submit transaction
    err = bbe.submitTransaction(ctx, task, output, proof)
    if err != nil {
        bbe.logger.Errorf("Failed to submit transaction for task %d: %v", task.TaskID, err)
        return
    }

    // 7. Broadcast execution data to validators (via Othentic)
    err = bbe.broadcastToValidators(task, execCtx, output, proof)
    if err != nil {
        bbe.logger.Errorf("Failed to broadcast to validators for task %d: %v", task.TaskID, err)
    }

    bbe.logger.Infof("Successfully executed task %d at block %d", task.TaskID, blockNumber)
}

type ExecutionContext struct {
    BlockNumber uint64
    Timestamp   uint64
    JobID       string
    ChainID     string
    Mode        string // "performer" or "validator"
}

func (bbe *BlockBasedExecutor) prepareExecutionContext(
    task *types.TaskTargetData,
    blockNumber uint64,
    timestamp uint64,
) *ExecutionContext {
    return &ExecutionContext{
        BlockNumber: blockNumber,
        Timestamp:   timestamp,
        JobID:       task.JobID.String(),
        ChainID:     task.TargetChainID,
        Mode:        "performer",
    }
}

func (bbe *BlockBasedExecutor) executeScript(
    ctx context.Context,
    task *types.TaskTargetData,
    execCtx *ExecutionContext,
) (*dockertypes.ExecutionResult, error) {
    // 1. Get script source
    scriptURL := bbe.getScriptURL(task)

    // 2. Prepare environment variables
    envVars := map[string]string{
        "TRIGGERX_BLOCK_NUMBER":   fmt.Sprintf("%d", execCtx.BlockNumber),
        "TRIGGERX_TIMESTAMP":      fmt.Sprintf("%d", execCtx.Timestamp),
        "TRIGGERX_JOB_ID":         execCtx.JobID,
        "TRIGGERX_CHAIN_ID":       execCtx.ChainID,
        "TRIGGERX_EXECUTION_MODE": execCtx.Mode,
    }

    // 3. Add secrets (from database)
    secrets, err := bbe.getSecrets(task.JobID)
    if err == nil {
        for key, value := range secrets {
            envVars["SECRET_"+key] = value
        }
    }

    // 4. Add storage (from database)
    storage, err := bbe.getStorage(task.JobID)
    if err == nil {
        for key, value := range storage {
            envVars["TRIGGERX_STORAGE_"+key] = value
        }
    }

    // 5. Execute in Docker with environment variables
    dockerExec := bbe.taskExecutor.validator.GetDockerExecutor()

    // TODO: Modify DockerExecutor to accept env vars
    result, err := dockerExec.ExecuteWithEnv(
        ctx,
        scriptURL,
        task.ScriptLanguage,
        envVars,
    )

    if err != nil {
        return nil, fmt.Errorf("docker execution failed: %w", err)
    }

    // 6. Parse storage updates from stderr
    bbe.parseStorageUpdates(result.Stderr, task.JobID)

    return result, nil
}

type ScriptOutput struct {
    ShouldExecute bool                   `json:"shouldExecute"`
    Params        []interface{}          `json:"params"`
    Metadata      ScriptOutputMetadata   `json:"metadata"`
}

type ScriptOutputMetadata struct {
    BlockNumber uint64                 `json:"blockNumber"`
    Timestamp   uint64                 `json:"timestamp"`
    Reason      string                 `json:"reason"`
    GasEstimate uint64                 `json:"gasEstimate,omitempty"`
    ApiCalls    []ApiCallMetadata      `json:"apiCalls,omitempty"`
}

type ApiCallMetadata struct {
    URL    string `json:"url"`
    Status int    `json:"status"`
}

func (bbe *BlockBasedExecutor) parseScriptOutput(output string) (*ScriptOutput, error) {
    var scriptOutput ScriptOutput
    err := json.Unmarshal([]byte(output), &scriptOutput)
    if err != nil {
        return nil, fmt.Errorf("invalid JSON output: %w", err)
    }

    // Validate required fields
    if scriptOutput.Metadata.BlockNumber == 0 {
        return nil, fmt.Errorf("missing blockNumber in metadata")
    }
    if scriptOutput.Metadata.Timestamp == 0 {
        return nil, fmt.Errorf("missing timestamp in metadata")
    }
    if scriptOutput.Metadata.Reason == "" {
        return nil, fmt.Errorf("missing reason in metadata")
    }

    return &scriptOutput, nil
}

type ExecutionProof struct {
    JobID        string `json:"jobId"`
    BlockNumber  uint64 `json:"blockNumber"`
    Timestamp    uint64 `json:"timestamp"`
    InputHash    string `json:"inputHash"`    // keccak256(blockNumber, timestamp)
    OutputHash   string `json:"outputHash"`   // keccak256(shouldExecute, params)
    Signature    string `json:"signature"`    // Performer's signature
    PerformerAddress string `json:"performerAddress"`
}

func (bbe *BlockBasedExecutor) generateExecutionProof(
    task *types.TaskTargetData,
    execCtx *ExecutionContext,
    output *ScriptOutput,
    result *dockertypes.ExecutionResult,
) *ExecutionProof {
    // 1. Calculate inputHash
    inputData := fmt.Sprintf("%d:%d", execCtx.BlockNumber, execCtx.Timestamp)
    inputHash := crypto.Keccak256Hash([]byte(inputData)).Hex()

    // 2. Calculate outputHash
    outputJSON, _ := json.Marshal(struct {
        ShouldExecute bool          `json:"shouldExecute"`
        Params        []interface{} `json:"params"`
    }{
        ShouldExecute: output.ShouldExecute,
        Params:        output.Params,
    })
    outputHash := crypto.Keccak256Hash(outputJSON).Hex()

    // 3. Sign (inputHash + outputHash)
    signature := bbe.signProof(inputHash, outputHash)

    return &ExecutionProof{
        JobID:            task.JobID.String(),
        BlockNumber:      execCtx.BlockNumber,
        Timestamp:        execCtx.Timestamp,
        InputHash:        inputHash,
        OutputHash:       outputHash,
        Signature:        signature,
        PerformerAddress: bbe.getPerformerAddress(),
    }
}

func (bbe *BlockBasedExecutor) submitTransaction(
    ctx context.Context,
    task *types.TaskTargetData,
    output *ScriptOutput,
    proof *ExecutionProof,
) error {
    // 1. Get contract ABI
    contractABI, method, err := bbe.taskExecutor.getContractMethodAndABI(task.TargetFunction, task)
    if err != nil {
        return fmt.Errorf("failed to get contract method: %w", err)
    }

    // 2. Convert params to ABI types
    convertedArgs, err := bbe.taskExecutor.processArguments(output.Params, method.Inputs, contractABI)
    if err != nil {
        return fmt.Errorf("failed to process arguments: %w", err)
    }

    // 3. Pack target contract calldata
    callData, err := contractABI.Pack(method.Name, convertedArgs...)
    if err != nil {
        return fmt.Errorf("failed to pack calldata: %w", err)
    }

    // 4. Pack execution contract call with proof
    // NEW: executeBlockBasedFunction(jobId, tgAmount, target, data, proof)
    executionInput, err := bbe.packExecutionWithProof(task, callData, proof)
    if err != nil {
        return fmt.Errorf("failed to pack execution: %w", err)
    }

    // 5. Submit transaction
    // ... (similar to existing executeAction logic)

    return nil
}

func (bbe *BlockBasedExecutor) broadcastToValidators(
    task *types.TaskTargetData,
    execCtx *ExecutionContext,
    output *ScriptOutput,
    proof *ExecutionProof,
) error {
    // Package execution data for Othentic
    executionData := BroadcastExecutionData{
        JobID:       task.JobID.String(),
        BlockNumber: execCtx.BlockNumber,
        Timestamp:   execCtx.Timestamp,
        ScriptURL:   bbe.getScriptURL(task),
        Output:      output,
        Proof:       proof,
    }

    // Send to Othentic network for validation
    // TODO: Implement Othentic broadcast
    return bbe.sendToOthentic(executionData)
}
```

---

## 3. Verification with AttestationCenter

### 3.1 Current AttestationCenter Architecture

Based on your description, AttestationCenter on L2 (Base):
- Receives attestations from validators
- BLS keys already registered
- Validates attestations and emits events

### 3.2 Integration with Script Execution

**Flow:**

```
Performer executes script at block N → Generates proof → Submits tx → Broadcasts to Othentic
                                                              ↓
Validator 1 re-executes script → Compares output → Attestation (APPROVE/REJECT)
Validator 2 re-executes script → Compares output → Attestation (APPROVE/REJECT)
Validator 3 re-executes script → Compares output → Attestation (APPROVE/REJECT)
                                                              ↓
                            Othentic aggregates BLS signatures
                                                              ↓
                            AttestationCenter receives aggregated attestation
                                                              ↓
                        If > threshold APPROVE → Execution confirmed
                        If > threshold REJECT → Slash performer (later phase)
```

### 3.3 Validator Re-Execution

**New Component: `internal/keeper/core/validation/block_validator.go`**

```go
package validation

type BlockTaskValidator struct {
    validator     *TaskValidator
    logger        logging.Logger
    archiveClient *ethclient.Client // Archive node for historical state
}

// OnValidationRequest is called when Othentic requests validation
func (btv *BlockTaskValidator) OnValidationRequest(ctx context.Context, req *ValidationRequest) error {
    btv.logger.Infof("Validating execution for job %s at block %d", req.JobID, req.BlockNumber)

    // 1. Get task data
    task, err := btv.getTaskData(req.JobID)
    if err != nil {
        return fmt.Errorf("failed to get task data: %w", err)
    }

    // 2. Re-execute script with SAME inputs
    execCtx := &ExecutionContext{
        BlockNumber: req.BlockNumber,
        Timestamp:   req.Timestamp,
        JobID:       req.JobID,
        ChainID:     task.TargetChainID,
        Mode:        "validator",
    }

    result, err := btv.executeScript(ctx, task, execCtx)
    if err != nil {
        btv.logger.Errorf("Re-execution failed: %v", err)
        return btv.submitAttestation(req, false, "Re-execution failed")
    }

    // 3. Parse output
    output, err := btv.parseScriptOutput(result.Output)
    if err != nil {
        btv.logger.Errorf("Failed to parse output: %v", err)
        return btv.submitAttestation(req, false, "Invalid output format")
    }

    // 4. Compare outputs
    performerOutputHash := req.Proof.OutputHash
    validatorOutputHash := btv.calculateOutputHash(output)

    if performerOutputHash != validatorOutputHash {
        btv.logger.Warnf("Output mismatch! Performer: %s, Validator: %s",
            performerOutputHash, validatorOutputHash)
        return btv.submitAttestation(req, false, "Output mismatch")
    }

    // 5. Verify on-chain transaction (if shouldExecute=true)
    if output.ShouldExecute {
        valid, err := btv.verifyTransaction(ctx, req, task, output)
        if err != nil || !valid {
            btv.logger.Errorf("Transaction verification failed: %v", err)
            return btv.submitAttestation(req, false, "Transaction verification failed")
        }
    }

    // 6. All checks passed
    btv.logger.Infof("Validation passed for job %s at block %d", req.JobID, req.BlockNumber)
    return btv.submitAttestation(req, true, "Validation passed")
}

func (btv *BlockTaskValidator) submitAttestation(
    req *ValidationRequest,
    approved bool,
    reason string,
) error {
    attestation := Attestation{
        JobID:       req.JobID,
        BlockNumber: req.BlockNumber,
        Approved:    approved,
        Reason:      reason,
        Validator:   btv.getValidatorAddress(),
        Signature:   btv.signAttestation(req, approved),
    }

    // Send to Othentic
    return btv.sendToOthentic(attestation)
}
```

### 3.4 AttestationCenter Contract Updates

**Current AttestationCenter.sol (assumed structure):**

```solidity
contract AttestationCenter {
    // BLS key registry
    mapping(address => bytes) public blsPublicKeys;

    // Attestation events
    event AttestationReceived(
        address indexed validator,
        bytes32 indexed taskHash,
        bool approved
    );

    function submitAttestation(
        bytes32 taskHash,
        bool approved,
        bytes calldata blsSignature
    ) external {
        // Verify BLS signature
        require(verifyBLSSignature(msg.sender, taskHash, blsSignature), "Invalid signature");

        // Emit attestation
        emit AttestationReceived(msg.sender, taskHash, approved);
    }
}
```

**Enhanced AttestationCenter.sol for Script Execution:**

```solidity
contract AttestationCenter {
    // ... existing code ...

    // NEW: Script execution attestations
    struct ScriptExecutionAttestation {
        uint256 jobId;
        uint64 blockNumber;
        bytes32 outputHash;
        bool approved;
        address[] validators;
        bytes aggregatedSignature;  // BLS aggregated signature
        uint256 timestamp;
    }

    mapping(bytes32 => ScriptExecutionAttestation) public scriptAttestations;

    event ScriptExecutionAttested(
        uint256 indexed jobId,
        uint64 indexed blockNumber,
        bytes32 outputHash,
        bool approved,
        uint8 validatorCount
    );

    // NEW: Submit aggregated attestation (called by Othentic)
    function submitScriptAttestation(
        uint256 jobId,
        uint64 blockNumber,
        bytes32 outputHash,
        bool approved,
        address[] calldata validators,
        bytes calldata aggregatedSignature
    ) external {
        require(msg.sender == othentic, "Only Othentic can submit");

        // Verify aggregated BLS signature
        require(
            verifyAggregatedBLSSignature(validators, outputHash, aggregatedSignature),
            "Invalid aggregated signature"
        );

        // Create attestation hash
        bytes32 attestationHash = keccak256(abi.encodePacked(
            jobId,
            blockNumber,
            outputHash
        ));

        // Store attestation
        scriptAttestations[attestationHash] = ScriptExecutionAttestation({
            jobId: jobId,
            blockNumber: blockNumber,
            outputHash: outputHash,
            approved: approved,
            validators: validators,
            aggregatedSignature: aggregatedSignature,
            timestamp: block.timestamp
        });

        emit ScriptExecutionAttested(
            jobId,
            blockNumber,
            outputHash,
            approved,
            uint8(validators.length)
        );

        // If rejected, trigger slashing (future phase)
        if (!approved) {
            // TODO: Call slashing contract
        }
    }

    // Query if execution was validated
    function isExecutionValidated(
        uint256 jobId,
        uint64 blockNumber,
        bytes32 outputHash
    ) external view returns (bool) {
        bytes32 attestationHash = keccak256(abi.encodePacked(
            jobId,
            blockNumber,
            outputHash
        ));

        ScriptExecutionAttestation memory attestation = scriptAttestations[attestationHash];
        return attestation.approved && attestation.validators.length >= minValidatorThreshold;
    }
}
```

### 3.5 TaskExecutionHub Integration

**Update TaskExecutionHub.sol:**

```solidity
contract TaskExecutionHub {
    // ... existing code ...

    AttestationCenter public attestationCenter;

    // NEW: Execute block-based function with proof
    function executeBlockBasedFunction(
        uint256 jobId,
        uint256 tgAmount,
        address target,
        bytes calldata data,
        ExecutionProof calldata proof
    ) external payable onlyKeeper nonReentrant {
        // 1. Verify proof hasn't been used
        require(!usedProofs[proof.outputHash], "Proof already used");

        // 2. Verify block window (must be recent)
        require(block.number - proof.blockNumber <= MAX_BLOCK_DELAY, "Proof too old");

        // 3. Check if execution was validated by AttestationCenter
        bool validated = attestationCenter.isExecutionValidated(
            jobId,
            proof.blockNumber,
            bytes32(proof.outputHash)
        );

        // For now, we don't require validation (add in later phase)
        // require(validated, "Execution not validated");

        // 4. Execute as usual
        (uint256 chainId, ) = jobRegistry.unpackJobId(jobId);
        require(chainId == block.chainid, "Job is from a different chain");

        address jobOwner = jobRegistry.getJobOwner(jobId);
        require(jobOwner != address(0), "Job not found");

        triggerGasRegistry.deductTGBalance(jobOwner, tgAmount);

        _executeFunction(target, data);

        // 5. Mark proof as used
        usedProofs[proof.outputHash] = true;

        emit BlockBasedExecutionCompleted(jobId, proof.blockNumber, proof.outputHash);
    }

    struct ExecutionProof {
        uint64 blockNumber;
        uint64 timestamp;
        bytes32 inputHash;
        bytes32 outputHash;
        bytes signature;
    }

    mapping(bytes32 => bool) public usedProofs;

    event BlockBasedExecutionCompleted(
        uint256 indexed jobId,
        uint64 indexed blockNumber,
        bytes32 outputHash
    );
}
```

---

## 4. Data Structures

### 4.1 Database Schema Additions

**New Table: `block_tasks`**

```sql
CREATE TABLE block_tasks (
    id BIGSERIAL PRIMARY KEY,
    job_id NUMERIC(78, 0) NOT NULL,
    task_id BIGINT NOT NULL,
    script_source VARCHAR(10) NOT NULL, -- 'ipfs' or 'url'
    script_cid TEXT,                    -- IPFS CID
    script_url TEXT,                    -- URL
    script_language VARCHAR(20) NOT NULL,
    execution_cadence INT NOT NULL DEFAULT 1,
    last_executed_block BIGINT DEFAULT 0,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    FOREIGN KEY (job_id) REFERENCES jobs(job_id)
);

CREATE INDEX idx_block_tasks_job_id ON block_tasks(job_id);
CREATE INDEX idx_block_tasks_last_executed ON block_tasks(last_executed_block);
```

**New Table: `script_secrets`**

```sql
CREATE TABLE script_secrets (
    id BIGSERIAL PRIMARY KEY,
    job_id NUMERIC(78, 0) NOT NULL,
    secret_key VARCHAR(255) NOT NULL,
    secret_value_encrypted TEXT NOT NULL,  -- Encrypted with server key
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    FOREIGN KEY (job_id) REFERENCES jobs(job_id),
    UNIQUE(job_id, secret_key)
);

CREATE INDEX idx_script_secrets_job_id ON script_secrets(job_id);
```

**New Table: `script_storage`**

```sql
CREATE TABLE script_storage (
    id BIGSERIAL PRIMARY KEY,
    job_id NUMERIC(78, 0) NOT NULL,
    storage_key VARCHAR(255) NOT NULL,
    storage_value TEXT NOT NULL,
    updated_at TIMESTAMP DEFAULT NOW(),
    FOREIGN KEY (job_id) REFERENCES jobs(job_id),
    UNIQUE(job_id, storage_key)
);

CREATE INDEX idx_script_storage_job_id ON script_storage(job_id);
```

**New Table: `execution_proofs`**

```sql
CREATE TABLE execution_proofs (
    id BIGSERIAL PRIMARY KEY,
    job_id NUMERIC(78, 0) NOT NULL,
    block_number BIGINT NOT NULL,
    timestamp BIGINT NOT NULL,
    input_hash VARCHAR(66) NOT NULL,
    output_hash VARCHAR(66) NOT NULL,
    performer_address VARCHAR(42) NOT NULL,
    signature TEXT NOT NULL,
    tx_hash VARCHAR(66),
    created_at TIMESTAMP DEFAULT NOW(),
    FOREIGN KEY (job_id) REFERENCES jobs(job_id)
);

CREATE INDEX idx_execution_proofs_job_id ON execution_proofs(job_id);
CREATE INDEX idx_execution_proofs_block ON execution_proofs(block_number);
CREATE INDEX idx_execution_proofs_output_hash ON execution_proofs(output_hash);
```

**New Table: `validator_attestations`**

```sql
CREATE TABLE validator_attestations (
    id BIGSERIAL PRIMARY KEY,
    job_id NUMERIC(78, 0) NOT NULL,
    block_number BIGINT NOT NULL,
    validator_address VARCHAR(42) NOT NULL,
    approved BOOLEAN NOT NULL,
    reason TEXT,
    signature TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    FOREIGN KEY (job_id) REFERENCES jobs(job_id)
);

CREATE INDEX idx_validator_attestations_job_block ON validator_attestations(job_id, block_number);
CREATE INDEX idx_validator_attestations_validator ON validator_attestations(validator_address);
```

### 4.2 Go Type Definitions

**New file: `pkg/types/block_execution.go`**

```go
package types

type BlockTask struct {
    JobID             *BigInt `json:"job_id" db:"job_id"`
    TaskID            int64   `json:"task_id" db:"task_id"`
    ScriptSource      string  `json:"script_source" db:"script_source"`
    ScriptCID         string  `json:"script_cid" db:"script_cid"`
    ScriptURL         string  `json:"script_url" db:"script_url"`
    ScriptLanguage    string  `json:"script_language" db:"script_language"`
    ExecutionCadence  int     `json:"execution_cadence" db:"execution_cadence"`
    LastExecutedBlock uint64  `json:"last_executed_block" db:"last_executed_block"`
    CreatedAt         string  `json:"created_at" db:"created_at"`
    UpdatedAt         string  `json:"updated_at" db:"updated_at"`
}

type ScriptSecret struct {
    JobID              *BigInt `json:"job_id" db:"job_id"`
    SecretKey          string  `json:"secret_key" db:"secret_key"`
    SecretValueEncrypted string `json:"-" db:"secret_value_encrypted"`
    CreatedAt          string  `json:"created_at" db:"created_at"`
}

type ScriptStorage struct {
    JobID        *BigInt `json:"job_id" db:"job_id"`
    StorageKey   string  `json:"storage_key" db:"storage_key"`
    StorageValue string  `json:"storage_value" db:"storage_value"`
    UpdatedAt    string  `json:"updated_at" db:"updated_at"`
}

type ExecutionProof struct {
    JobID            *BigInt `json:"job_id" db:"job_id"`
    BlockNumber      uint64  `json:"block_number" db:"block_number"`
    Timestamp        uint64  `json:"timestamp" db:"timestamp"`
    InputHash        string  `json:"input_hash" db:"input_hash"`
    OutputHash       string  `json:"output_hash" db:"output_hash"`
    PerformerAddress string  `json:"performer_address" db:"performer_address"`
    Signature        string  `json:"signature" db:"signature"`
    TxHash           string  `json:"tx_hash" db:"tx_hash"`
    CreatedAt        string  `json:"created_at" db:"created_at"`
}

type ValidatorAttestation struct {
    JobID            *BigInt `json:"job_id" db:"job_id"`
    BlockNumber      uint64  `json:"block_number" db:"block_number"`
    ValidatorAddress string  `json:"validator_address" db:"validator_address"`
    Approved         bool    `json:"approved" db:"approved"`
    Reason           string  `json:"reason" db:"reason"`
    Signature        string  `json:"signature" db:"signature"`
    CreatedAt        string  `json:"created_at" db:"created_at"`
}

type ScriptOutput struct {
    ShouldExecute bool                 `json:"shouldExecute"`
    Params        []interface{}        `json:"params"`
    Metadata      ScriptOutputMetadata `json:"metadata"`
}

type ScriptOutputMetadata struct {
    BlockNumber uint64            `json:"blockNumber"`
    Timestamp   uint64            `json:"timestamp"`
    Reason      string            `json:"reason"`
    GasEstimate uint64            `json:"gasEstimate,omitempty"`
    ApiCalls    []ApiCallMetadata `json:"apiCalls,omitempty"`
}

type ApiCallMetadata struct {
    URL    string `json:"url"`
    Status int    `json:"status"`
}
```

---

## 5. Implementation Phases

### Phase 1: Basic Execution (No Verification)

**Goal**: Get block-based execution working with single performer, no validation.

**Tasks:**
1. ✅ Add Task Definition Type 7
2. ✅ Implement BlockMonitor for block subscription
3. ✅ Implement BlockBasedExecutor for script execution
4. ✅ Add database tables (block_tasks, script_secrets, script_storage)
5. ✅ Implement script output parsing
6. ✅ Test with simple Go/Python scripts on testnet

**Success Criteria:**
- Performer can execute block-based tasks every N blocks
- Scripts run in Docker with deterministic inputs
- Structured JSON output parsed correctly
- Transactions submitted successfully

---

### Phase 2: Secrets & Storage

**Goal**: Add secrets management and persistent storage.

**Tasks:**
1. ✅ Implement secrets encryption/decryption
2. ✅ Inject secrets as environment variables
3. ✅ Implement storage get/set via environment variables
4. ✅ Parse storage updates from stderr
5. ✅ Add API endpoints for managing secrets

**Success Criteria:**
- Users can set secrets via API
- Scripts can access secrets securely
- Scripts can persist state between executions

---

### Phase 3: Validation & Attestation

**Goal**: Add multi-validator verification without slashing.

**Tasks:**
1. ✅ Implement validator re-execution logic
2. ✅ Add output comparison
3. ✅ Integrate with Othentic for attestation broadcast
4. ✅ Update AttestationCenter.sol with script execution logic
5. ✅ Store attestations in database

**Success Criteria:**
- Validators re-execute scripts
- Attestations submitted to AttestationCenter
- Output mismatches detected
- No slashing yet (just logging)

---

### Phase 4: Fraud Proofs & Slashing (Future)

**Goal**: Add slashing for malicious performers.

**Tasks:**
1. ❌ Implement fraud proof generation
2. ❌ Add slashing logic to AttestationCenter
3. ❌ Implement automatic slashing triggers
4. ❌ Add appeal mechanism

**Success Criteria:**
- Malicious performers slashed automatically
- Honest performers rewarded
- Slashing criteria clear and fair

---

## 6. Code Examples

### 6.1 Complete End-to-End Example

**User's Script (Go):**

File: `my-liquidation-checker/index.go`

```go
package main

import (
    "encoding/json"
    "fmt"
    "math/big"
    "os"
)

// ... (same structure as shown in section 1.3.1)

func checkCondition(ctx ExecutionContext, apiKey, lastBlock string) (bool, []interface{}, string) {
    // Example: Check if liquidation is needed

    // 1. Get user's collateral ratio from on-chain
    ratio := getUserCollateralRatio("0x123...")

    // 2. Check if below threshold
    if ratio < 1.2 {
        // Liquidation needed
        return true, []interface{}{
            "0x123...",  // user address
            big.NewInt(1000000000000000000),  // debt amount
        }, "Collateral ratio below 1.2"
    }

    return false, nil, "Collateral ratio healthy"
}

func getUserCollateralRatio(user string) float64 {
    // Call on-chain contract or API
    // For demo, return dummy value
    return 1.15
}
```

**Configuration:**

File: `my-liquidation-checker/script.json`

```json
{
  "version": "1.0",
  "runtime": {
    "language": "go",
    "memory": 256,
    "timeout": 30
  },
  "execution": {
    "cadence": 1,
    "deterministic": true
  },
  "outputs": {
    "format": "json",
    "schema": {
      "shouldExecute": "boolean",
      "params": "array",
      "metadata": "object"
    }
  }
}
```

**Keeper Execution:**

```go
// Keeper detects new block
blockNumber := 12345678
timestamp := 1709876543

// Execute script
executor := NewBlockBasedExecutor(...)
result, err := executor.executeBlockTask(context.Background(), task, blockNumber, timestamp)

// Parse output
// {"shouldExecute":true,"params":["0x123...",1000000000000000000],"metadata":{...}}

// If shouldExecute=true, submit transaction
// TaskExecutionHub.executeBlockBasedFunction(jobId, tgAmount, liquidationContract, calldata, proof)
```

---

## 7. Summary

### What Users Provide

1. **Script files** (Go/Python/TypeScript/JavaScript)
2. **script.json** configuration
3. **IPFS CID** or **GitHub URL**
4. **Secrets** (via API after task creation)

### What Keeper Does

1. **Monitors new blocks** via BlockMonitor
2. **Executes scripts** in Docker with deterministic inputs
3. **Parses JSON output** (shouldExecute, params, metadata)
4. **Submits transaction** if shouldExecute=true
5. **Broadcasts to validators** for verification

### What Validators Do

1. **Receive execution data** from Othentic
2. **Re-execute script** with same inputs
3. **Compare outputs** (outputHash)
4. **Submit attestation** (APPROVE/REJECT)

### What AttestationCenter Does

1. **Receives attestations** from Othentic
2. **Verifies BLS signatures**
3. **Stores validation results**
4. **Emits events** for transparency

### Integration Points

1. **Block subscription** → BlockMonitor
2. **Script execution** → DockerExecutor (enhanced with env vars)
3. **Attestation** → Othentic → AttestationCenter
4. **On-chain execution** → TaskExecutionHub.executeBlockBasedFunction()

---

## Next Steps

1. ✅ **Review this specification** with team
2. ✅ **Start Phase 1**: Basic execution without validation
3. ✅ **Test with simple scripts** on Base Sepolia testnet
4. ✅ **Iterate** based on findings
5. ❌ **Add validation** in Phase 3
6. ❌ **Add slashing** in Phase 4

---

**Document Version:** 1.0
**Date:** 2025-11-14
**Author:** TriggerX Engineering Team
**Status:** Specification Ready for Implementation
