# Custom Script Execution Feature - Implementation Guide

**Feature**: Custom Script Jobs (TaskDefinitionID = 7)
**Status**: Phase 1 Complete, Phase 2 Pending
**Last Updated**: 2025-11-16

---

## Table of Contents
1. [Overview](#overview)
2. [Phase 1 Implementation (COMPLETED)](#phase-1-implementation-completed)
3. [Phase 2 Requirements (PENDING)](#phase-2-requirements-pending)
4. [Architecture](#architecture)
5. [Database Schema](#database-schema)
6. [API Usage](#api-usage)
7. [Code Locations](#code-locations)

---

## Overview

Custom script execution allows users to create jobs that run custom TypeScript/Go scripts at regular intervals. Scripts can:
- Execute arbitrary logic (API calls, calculations, etc.)
- Decide whether to execute an on-chain transaction
- Generate dynamic calldata for contract interactions
- Persist state across executions using key-value storage
- Provide execution metadata for fraud-proof verification (Phase 2)

### Key Characteristics
- **Time-based execution**: Scripts run every N seconds (similar to TaskDefinitionID 1/2)
- **Optimistic execution**: Tasks execute without pre-verification, can be challenged later (Phase 2)
- **Fraud-proof system**: Validators can re-execute and challenge incorrect results (Phase 2)
- **Persistent storage**: Scripts can read/write key-value pairs that persist across executions

---

## Phase 1 Implementation (COMPLETED)

### What Was Implemented

#### 1. Database Schema & Queries
**Files Created:**
- `internal/dbserver/repository/schema/custom_jobs_schema.cql`
- `internal/dbserver/repository/queries/custom_job_queries.go`

**Tables Created:**
```sql
-- Main custom jobs table
CREATE TABLE triggerx.custom_jobs (
    job_id varint PRIMARY KEY,
    task_definition_id int,
    recurring boolean,
    custom_script_url text,        -- IPFS URL for the script
    time_interval bigint,           -- Execution interval in seconds
    script_language text,           -- 'typescript' or 'go'
    script_hash text,               -- SHA-256 hash for verification
    next_execution_time timestamp,  -- When to execute next
    max_execution_time int,         -- Timeout limit
    challenge_period bigint,        -- Fraud-proof challenge window
    expiration_time timestamp,
    last_executed_at timestamp,
    is_completed boolean,
    is_active boolean
);

-- Storage for persistent key-value pairs
CREATE TABLE triggerx.script_storage (
    job_id varint,
    storage_key text,
    storage_value text,
    updated_at timestamp,
    PRIMARY KEY (job_id, storage_key)
);
```

#### 2. Repository Layer
**Files Created:**
- `internal/dbserver/repository/custom_job_repository.go`
- `internal/dbserver/repository/script_storage_repository.go`

**Key Methods:**
- `CreateCustomJob(jobData *CustomJobData)` - Create new custom job
- `GetCustomJobsDueForExecution(currentTime)` - Fetch jobs ready to run
- `UpdateNextExecutionTime(jobID, nextTime)` - Update schedule
- `GetStorageByJobID(jobID)` - Fetch all storage for a job
- `UpsertStorage(jobID, key, value)` - Update storage key

#### 3. Scheduler Integration
**File Modified:** `internal/dbserver/handlers/time_jobs.go`

**Changes:**
- Extended `GetTimeBasedTasks()` to fetch custom jobs alongside time-based jobs
- Converts `CustomJobData` to `ScheduleTimeTaskData` format
- Fetches and includes storage in task data sent to keepers
- Custom jobs now flow through the same scheduler pipeline as regular time jobs

**Flow:**
```
Time Scheduler → Custom Job Repository → Fetch Storage →
Convert to Task Format → Send to Keeper
```

#### 4. Job Creation Handler
**File Modified:** `internal/dbserver/handlers/job_create.go`

**Changes:**
- Updated validation: TaskDefinitionID now accepts 1-7 (was 1-6)
- Added TaskDefinitionID=7 to script validation (IPFS code validation)
- Added case 7 handler in job creation switch statement

**Example Job Creation:**
```go
case 7:
    customJobData := commonTypes.CustomJobData{
        JobID:              jobID,
        TaskDefinitionID:   7,
        CustomScriptUrl:    tempJobs[i].DynamicArgumentsScriptUrl,
        TimeInterval:       tempJobs[i].TimeInterval,
        ScriptLanguage:     tempJobs[i].Language,
        NextExecutionTime:  nextExecutionTime,
        // ...
    }
    h.customJobRepository.CreateCustomJob(&customJobData)
```

#### 5. Keeper Execution Logic
**Files Created/Modified:**
- `internal/keeper/core/execution/custom_executor.go` (NEW)
- `internal/keeper/core/execution/action.go` (MODIFIED)

**Key Function:**
```go
func (e *TaskExecutor) ExecuteCustomScript(
    ctx context.Context,
    targetData *types.TaskTargetData,
    triggerData *types.TaskTriggerData,
) (*types.CustomScriptOutput, map[string]string, error)
```

**Execution Flow:**
1. Receive task with storage data from scheduler
2. Execute script in Docker container
3. Parse JSON output from stdout
4. Extract storage updates from JSON response
5. Validate output (address format, calldata format)
6. Return script output + storage updates

**Script Output Format:**
```json
{
  "shouldExecute": true,
  "targetContract": "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb",
  "calldata": "0xa9059cbb000000000000...",
  "storageUpdates": {
    "lastPrice": "1234.56",
    "executionCount": "42"
  },
  "metadata": {
    "timestamp": 1234567890,
    "reason": "Price threshold exceeded"
  }
}
```

#### 6. TaskMonitor Storage Updates
**Files Modified:**
- `internal/taskmonitor/clients/database/task.go`
- `internal/taskmonitor/clients/database/queries/task.go`
- `internal/taskmonitor/events/task.go`

**New Functions:**
- `UpdateScriptStorage(jobID, storageUpdates)` - Persist storage to DB
- `GetJobIDByTaskID(taskID)` - Helper to get job ID
- `GetTaskDefinitionIDByTaskID(taskID)` - Helper to get task definition

**Integration Point:**
After task is submitted on-chain and IPFS data is fetched, TaskMonitor checks if `TaskDefinitionID == 7` and updates storage:

```go
if taskData.TaskDefinitionID == 7 && ipfsData.ActionData.StorageUpdates != nil {
    jobID, _ := h.db.GetJobIDByTaskID(taskData.TaskID)
    h.db.UpdateScriptStorage(jobID, ipfsData.ActionData.StorageUpdates)
}
```

#### 7. Type Definitions
**Files Created/Modified:**
- `pkg/types/custom_execution.go` (NEW)
- `pkg/types/database_models.go` (MODIFIED)
- `pkg/types/schedulers.go` (MODIFIED)
- `pkg/types/keeper.go` (MODIFIED)

**Key Types:**
```go
// Script output format
type CustomScriptOutput struct {
    ShouldExecute  bool                       `json:"shouldExecute"`
    TargetContract string                     `json:"targetContract,omitempty"`
    Calldata       string                     `json:"calldata,omitempty"`
    Metadata       CustomScriptOutputMetadata `json:"metadata"`
    StorageUpdates map[string]string          `json:"storageUpdates,omitempty"`
}

// Job data
type CustomJobData struct {
    JobID              *BigInt
    TaskDefinitionID   int
    CustomScriptUrl    string
    TimeInterval       int64
    ScriptLanguage     string
    NextExecutionTime  time.Time
    // ...
}

// Task target data (extended)
type TaskTargetData struct {
    // ... existing fields ...
    ScriptStorage  map[string]string `json:"script_storage,omitempty"`
    ScriptLanguage string            `json:"script_language,omitempty"`
}

// Performer action data (extended)
type PerformerActionData struct {
    // ... existing fields ...
    StorageUpdates map[string]string `json:"storage_updates,omitempty"`
}
```

### Phase 1 Limitations

**Storage Access:**
- ❌ Scripts CANNOT read previous storage values (no env var injection)
- ✅ Scripts CAN write storage updates via JSON output
- Storage is fetched by DBServer and passed to keeper (keeper doesn't access DB directly)

**Execution:**
- Uses standard Docker `Execute()` method (no custom env vars)
- Storage updates returned in JSON instead of via stderr parsing
- Environment variable injection deferred to Phase 2

**Verification:**
- No fraud-proof generation yet
- No challenge mechanism
- No validator re-execution
- Execution metadata collected but not used for verification

---

## Phase 2 Requirements (PENDING)

### 1. Environment Variable Injection

**Goal:** Allow scripts to READ previous storage values

**Implementation Required:**
1. **Docker Executor Enhancement:**
   - Add `ExecuteWithEnv()` method to `DockerExecutorAPI` interface
   - OR extend existing `Execute()` to accept environment variables
   - Update `pkg/dockerexecutor/interface.go`

2. **Storage Injection:**
   - Uncomment `prepareCustomScriptEnv()` in `custom_executor.go`
   - Inject storage as environment variables: `TRIGGERX_STORAGE_lastPrice=1234.56`
   - Inject execution context: `TRIGGERX_JOB_ID`, `TRIGGERX_TASK_ID`, `TRIGGERX_TIMESTAMP`

**File Locations:**
- `pkg/dockerexecutor/interface.go` - Add method to interface
- `pkg/dockerexecutor/dockerexecutor.go` - Implement ExecuteWithEnv
- `internal/keeper/core/execution/custom_executor.go` - Uncomment and use prepareCustomScriptEnv()

**Example:**
```go
// Phase 2: Use environment variable injection
envVars := prepareCustomScriptEnv(targetData, triggerData)
result, err := e.validator.GetDockerExecutor().ExecuteWithEnv(
    ctx,
    scriptURL,
    scriptLanguage,
    envVars,
)
```

### 2. Fraud-Proof System

**Goal:** Enable validators to challenge incorrect executions

**Components to Implement:**

#### A. Execution Proof Generation
**File:** `internal/keeper/core/execution/custom_executor.go`

Uncomment and implement:
- `generateExecutionID()` - Unique identifier for each execution
- `generateExecutionProof()` - Cryptographic proof of execution
- `hashStorage()` - Deterministic storage hash
- `signProof()` - Sign with keeper's private key

**Proof Structure:**
```go
type ExecutionProof struct {
    ExecutionID      string  // exec_123_abc123
    JobID            string
    Timestamp        int64
    InputHash        string  // Hash of (timestamp, jobID, storage)
    OutputHash       string  // Hash of (shouldExecute, targetContract, calldata)
    Signature        string  // Keeper's signature
    PerformerAddress string
}
```

#### B. Execution Tracking
**Database Table:** `custom_script_executions` (already defined in schema)

**Implementation:**
- Store each execution with inputs, outputs, metadata
- Track challenge period deadlines
- Record verification status

**Repository Method:**
```go
func (r *customExecutionRepository) CreateExecution(
    execution *CustomScriptExecution,
) error
```

#### C. Challenge Mechanism

**New RPC Endpoint:**
```go
// Challenger submits alternative execution result
POST /api/challenge
{
    "execution_id": "exec_123_abc123",
    "challenger_address": "0x...",
    "challenge_reason": "incorrect_output",
    "challenger_output_hash": "0x...",
    "challenger_signature": "0x..."
}
```

**Challenge Flow:**
1. Challenger re-executes script with same inputs (from execution record)
2. If output differs, submit challenge with proof
3. Validators vote on correctness
4. If challenger wins, performer is slashed

**Files to Create:**
- `internal/challenger/executor.go` - Re-execution logic
- `internal/challenger/verifier.go` - Compare outputs
- `internal/dbserver/handlers/challenge.go` - Challenge submission endpoint

#### D. Validator Re-Execution

**Validator Flow:**
1. Receive challenge notification
2. Fetch execution record with inputs
3. Re-execute script deterministically
4. Compare output with performer's result
5. Vote on-chain (approve or reject)

**Determinism Requirements:**
- API calls: Use recorded responses from metadata
- Contract calls: Execute at specific block numbers (from metadata)
- Timestamps: Use recorded execution timestamp
- Storage: Use input snapshot

**Implementation:**
```go
func (v *Validator) ReExecuteForVerification(
    executionID string,
) (*ValidationResult, error) {
    // 1. Fetch execution record
    execution := fetchExecution(executionID)

    // 2. Prepare deterministic environment
    env := prepareDeterministicEnv(execution.InputStorage, execution.Metadata)

    // 3. Execute script
    result := executeScript(execution.ScriptURL, env)

    // 4. Compare output
    return compareOutputs(result, execution.OutputHash)
}
```

#### E. Metadata Recording

**Script Output Extension:**
```json
{
  "shouldExecute": true,
  "targetContract": "0x...",
  "calldata": "0x...",
  "storageUpdates": {...},
  "metadata": {
    "timestamp": 1234567890,
    "reason": "Price threshold exceeded",
    "gasEstimate": 100000,
    "apiCalls": [
      {
        "url": "https://api.example.com/price",
        "blockNumber": 12345678,
        "response": {"price": 1234.56},
        "statusCode": 200,
        "timestamp": 1234567890
      }
    ],
    "contractCalls": [
      {
        "contract": "0x...",
        "function": "getPrice",
        "blockNumber": 12345678,  // CRITICAL for determinism
        "response": "1234560000000000000000",
        "chainId": "11155111"
      }
    ]
  }
}
```

**Storage:**
- Metadata stored in `custom_script_executions.execution_metadata` as JSON
- Used by validators for deterministic re-execution

### 3. Secrets Management

**Goal:** Securely inject API keys and sensitive data into scripts

**Options:**

#### Option A: On-Chain Encryption
- User encrypts secrets with aggregator's public key
- Store encrypted secrets in custom job data
- Aggregator decrypts when preparing execution
- Inject as environment variables

#### Option B: Off-Chain Secret Store
- Use HashiCorp Vault or similar
- User grants aggregator access to specific secrets
- Secrets never touch blockchain or IPFS
- Higher trust requirement

**Implementation:**
```go
// Encrypted secrets in job data
type CustomJobData struct {
    // ... existing fields ...
    EncryptedSecrets map[string]string `json:"encrypted_secrets"`
}

// Decryption in execution
func (e *TaskExecutor) injectSecrets(env map[string]string, jobData *CustomJobData) {
    for key, encryptedValue := range jobData.EncryptedSecrets {
        decrypted := e.decryptor.Decrypt(encryptedValue)
        env["SECRET_"+key] = decrypted
    }
}
```

**Security Considerations:**
- Secrets never logged or stored in plaintext
- Secrets not included in fraud-proof metadata
- Use separate encryption key per user/job
- Implement key rotation

### 4. Additional Enhancements

#### A. Script Hash Verification
**Purpose:** Ensure script hasn't been modified since job creation

**Implementation:**
```go
func (e *TaskExecutor) verifyScriptHash(scriptURL string, expectedHash string) error {
    content := fetchScript(scriptURL)
    actualHash := sha256(content)
    if actualHash != expectedHash {
        return fmt.Errorf("script hash mismatch")
    }
}
```

#### B. Gas Estimation
**Purpose:** Predict execution costs accurately

**Implementation:**
- Script returns `gasEstimate` in metadata
- Compare with actual gas used
- Adjust job cost predictions

#### C. Execution Timeouts
**Purpose:** Prevent infinite loops

**Implementation:**
- Use `max_execution_time` from custom job data
- Docker executor enforces timeout
- Failed executions marked as timeout errors

#### D. Rate Limiting
**Purpose:** Prevent abuse of external APIs

**Implementation:**
- Track API call frequency per job
- Limit concurrent executions
- Throttle based on user tier

---

## Architecture

### Data Flow Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                        Phase 1: Complete                          │
└─────────────────────────────────────────────────────────────────┘

1. JOB CREATION
   User → DBServer API → CreateJobData (case 7)
       → CustomJobRepository.CreateCustomJob()
       → Cassandra: custom_jobs table

2. SCHEDULING
   Time Scheduler → GetTimeBasedTasks()
       → CustomJobRepository.GetCustomJobsDueForExecution()
       → ScriptStorageRepository.GetStorageByJobID()
       → Convert to ScheduleTimeTaskData (with storage)
       → Send to Keeper

3. EXECUTION (Keeper)
   Keeper receives task → TaskExecutor.executeAction()
       → case 7: ExecuteCustomScript()
       → Docker.Execute(scriptURL, language)
       → Parse JSON output (stdout)
       → Extract storageUpdates from JSON
       → Build calldata for target contract
       → Execute on-chain transaction
       → Return PerformerActionData (with storageUpdates)

4. STORAGE PERSISTENCE (TaskMonitor)
   TaskMonitor listens to on-chain events
       → TaskSubmitted event detected
       → Fetch IPFS data (includes storageUpdates)
       → if TaskDefinitionID == 7:
           → GetJobIDByTaskID()
           → UpdateScriptStorage(jobID, storageUpdates)
       → Cassandra: script_storage table updated

┌─────────────────────────────────────────────────────────────────┐
│                        Phase 2: Pending                           │
└─────────────────────────────────────────────────────────────────┘

5. FRAUD-PROOF GENERATION (Keeper)
   After execution → generateExecutionProof()
       → Hash inputs (timestamp, storage, jobID)
       → Hash outputs (shouldExecute, targetContract, calldata)
       → Sign with keeper private key
       → Store in custom_script_executions table
       → Upload proof to IPFS

6. CHALLENGE SUBMISSION (Challenger)
   Challenger → Re-execute script with same inputs
       → Compare outputs
       → if different: Submit challenge on-chain
       → Execution enters challenge period

7. VALIDATION (Validators)
   Validators → Fetch execution record
       → Re-execute deterministically (using metadata)
       → Vote on correctness
       → Consensus determines outcome
       → Slash performer if incorrect
```

### Component Interactions

```
┌──────────────┐         ┌──────────────┐         ┌──────────────┐
│   DBServer   │         │  Scheduler   │         │    Keeper    │
│              │         │  (Cron)      │         │              │
└──────┬───────┘         └──────┬───────┘         └──────┬───────┘
       │                        │                        │
       │  Create Custom Job     │                        │
       │  (TaskDefID=7)         │                        │
       │◄───────────────────────┤                        │
       │                        │                        │
       │  Store in DB           │                        │
       ├───────────────────────►│                        │
       │                        │                        │
       │                        │  Poll for tasks        │
       │                        │  (every N seconds)     │
       │                        │◄───────────────────────┤
       │                        │                        │
       │  Fetch custom jobs     │                        │
       │  + storage             │                        │
       │◄───────────────────────┤                        │
       │                        │                        │
       │  Return tasks          │                        │
       ├───────────────────────►│                        │
       │                        │                        │
       │                        │  Send task to keeper   │
       │                        ├───────────────────────►│
       │                        │                        │
       │                        │                        │  Execute script
       │                        │                        │  in Docker
       │                        │                        │
       │                        │  Return action data    │
       │                        │  (with storageUpdates) │
       │                        │◄───────────────────────┤
       │                        │                        │
       │                        │  Submit on-chain       │
       │                        │                        │
       │                        │                        │
       │  On-chain event        │                        │
       │  (TaskSubmitted)       │                        │
       │◄───────────────────────┘                        │
       │                                                  │
       │  Update storage in DB                           │
       │                                                  │
       └──────────────────────────────────────────────────┘
```

---

## Database Schema

### Tables Created (Phase 1)

#### 1. custom_jobs
```sql
CREATE TABLE triggerx.custom_jobs (
    job_id varint PRIMARY KEY,
    task_definition_id int,
    recurring boolean,
    custom_script_url text,
    time_interval bigint,
    script_language text,
    script_hash text,
    next_execution_time timestamp,
    max_execution_time int,
    challenge_period bigint,
    expiration_time timestamp,
    last_executed_at timestamp,
    is_completed boolean,
    is_active boolean
);
```

**Indexes:**
```sql
CREATE INDEX custom_jobs_next_execution_idx
ON triggerx.custom_jobs (next_execution_time);

CREATE INDEX custom_jobs_active_idx
ON triggerx.custom_jobs (is_active);
```

#### 2. script_storage
```sql
CREATE TABLE triggerx.script_storage (
    job_id varint,
    storage_key text,
    storage_value text,
    updated_at timestamp,
    PRIMARY KEY (job_id, storage_key)
);
```

**Purpose:** Persistent key-value storage for scripts

### Tables Defined (Phase 2 - Not Yet Used)

#### 3. custom_script_executions
```sql
CREATE TABLE triggerx.custom_script_executions (
    execution_id text PRIMARY KEY,
    job_id varint,
    task_id bigint,
    scheduled_time timestamp,
    actual_time timestamp,
    performer_address text,

    -- Inputs (for re-execution)
    input_timestamp bigint,
    input_storage text,      -- JSON snapshot
    input_hash text,

    -- Outputs
    should_execute boolean,
    target_contract text,
    calldata text,
    output_hash text,

    -- Metadata
    execution_metadata text, -- JSON (API calls, contract calls)
    script_hash text,
    signature text,

    -- Result
    tx_hash text,
    execution_status text,   -- 'success', 'failed', 'no_execution'
    execution_error text,

    -- Verification
    verification_status text,  -- 'pending', 'verified', 'challenged'
    challenge_deadline timestamp,
    is_challenged boolean,
    challenge_count int,

    created_at timestamp
);
```

#### 4. execution_challenges
```sql
CREATE TABLE triggerx.execution_challenges (
    challenge_id text PRIMARY KEY,
    execution_id text,
    challenger_address text,
    challenge_reason text,

    -- Challenger's claimed output
    challenger_output_hash text,
    challenger_should_execute boolean,
    challenger_target_contract text,
    challenger_calldata text,
    challenger_signature text,

    -- Resolution
    resolution_status text,  -- 'pending', 'approved', 'rejected'
    resolution_time timestamp,
    validator_count int,
    approve_count int,
    reject_count int,

    created_at timestamp
);
```

---

## API Usage

### Creating a Custom Script Job

**Endpoint:** `POST /api/jobs`

**Request Body:**
```json
{
  "jobs": [
    {
      "job_id": "123456789",
      "user_address": "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb",
      "ether_balance": "1000000000000000000",
      "token_balance": "5000000000000000000",

      "job_title": "Automated Token Swap Based on Price",
      "task_definition_id": 7,
      "custom": true,
      "language": "typescript",
      "time_frame": 604800,
      "recurring": true,
      "job_cost_prediction": 0.5,
      "timezone": "UTC",
      "created_chain_id": "11155111",

      "time_interval": 3600,
      "dynamic_arguments_script_url": "ipfs://QmXyZ123...",

      "is_safe": false
    }
  ]
}
```

**Response:**
```json
{
  "user_id": 42,
  "account_balance": "1000000000000000000",
  "token_balance": "5000000000000000000",
  "job_ids": ["123456789"],
  "task_definition_ids": [7],
  "time_frames": [604800]
}
```

### Script Format

**TypeScript Example:**
```typescript
// main.ts - IPFS hosted script
interface TriggerXOutput {
  shouldExecute: boolean;
  targetContract?: string;
  calldata?: string;
  storageUpdates?: Record<string, string>;
  metadata: {
    timestamp: number;
    reason: string;
    gasEstimate?: number;
    apiCalls?: Array<{
      url: string;
      blockNumber?: number;
      response: any;
      statusCode: number;
      timestamp: number;
    }>;
    contractCalls?: Array<{
      contract: string;
      function: string;
      blockNumber: number;
      response: any;
      chainId: string;
    }>;
  };
}

async function main(): Promise<TriggerXOutput> {
  // Fetch current ETH price
  const response = await fetch('https://api.coinbase.com/v2/prices/ETH-USD/spot');
  const data = await response.json();
  const currentPrice = parseFloat(data.data.amount);

  // Check if we should execute
  const threshold = 2000;
  if (currentPrice > threshold) {
    // Build swap calldata
    const swapCalldata = buildSwapCalldata(currentPrice);

    return {
      shouldExecute: true,
      targetContract: "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb",
      calldata: swapCalldata,
      storageUpdates: {
        "lastPrice": currentPrice.toString(),
        "lastExecutionTime": Date.now().toString(),
        "executionCount": "1"  // Phase 2: will increment existing value
      },
      metadata: {
        timestamp: Date.now(),
        reason: `Price ${currentPrice} exceeded threshold ${threshold}`,
        gasEstimate: 150000,
        apiCalls: [{
          url: "https://api.coinbase.com/v2/prices/ETH-USD/spot",
          response: data,
          statusCode: 200,
          timestamp: Date.now()
        }]
      }
    };
  }

  // Don't execute
  return {
    shouldExecute: false,
    metadata: {
      timestamp: Date.now(),
      reason: `Price ${currentPrice} below threshold ${threshold}`
    }
  };
}

// Output to stdout as JSON
console.log(JSON.stringify(await main()));
```

**Go Example:**
```go
package main

import (
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "os"
)

type Output struct {
    ShouldExecute  bool              `json:"shouldExecute"`
    TargetContract string            `json:"targetContract,omitempty"`
    Calldata       string            `json:"calldata,omitempty"`
    StorageUpdates map[string]string `json:"storageUpdates,omitempty"`
    Metadata       Metadata          `json:"metadata"`
}

type Metadata struct {
    Timestamp int64  `json:"timestamp"`
    Reason    string `json:"reason"`
}

func main() {
    // Fetch data
    resp, _ := http.Get("https://api.example.com/data")
    defer resp.Body.Close()
    body, _ := io.ReadAll(resp.Body)

    // Process and decide
    output := Output{
        ShouldExecute:  true,
        TargetContract: "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb",
        Calldata:       "0x...",
        StorageUpdates: map[string]string{
            "lastUpdate": "1234567890",
        },
        Metadata: Metadata{
            Timestamp: 1234567890,
            Reason:    "Condition met",
        },
    }

    // Output JSON to stdout
    json.NewEncoder(os.Stdout).Encode(output)
}
```

---

## Code Locations

### Phase 1 Implementation Files

#### Database Layer
```
internal/dbserver/repository/schema/
  └── custom_jobs_schema.cql                    # Schema definitions

internal/dbserver/repository/queries/
  └── custom_job_queries.go                     # CQL query constants

internal/dbserver/repository/
  ├── custom_job_repository.go                  # Custom job CRUD
  ├── custom_execution_repository.go            # Execution tracking (unused)
  └── script_storage_repository.go              # Storage management
```

#### Handler Layer
```
internal/dbserver/handlers/
  ├── handler.go                                # Added repository fields
  ├── job_create.go                             # Case 7 handler
  └── time_jobs.go                              # Scheduler integration
```

#### Type Definitions
```
internal/dbserver/types/
  └── job_types.go                              # Updated validation (max=7)

pkg/types/
  ├── custom_execution.go                       # NEW: Custom types
  ├── database_models.go                        # CustomJobData
  ├── schedulers.go                             # Extended TaskTargetData
  └── keeper.go                                 # Extended PerformerActionData
```

#### Keeper Execution
```
internal/keeper/core/execution/
  ├── custom_executor.go                        # NEW: Custom script execution
  └── action.go                                 # Case 7 integration
```

#### TaskMonitor
```
internal/taskmonitor/clients/database/
  ├── task.go                                   # Storage update methods
  └── queries/task.go                           # Storage queries

internal/taskmonitor/events/
  └── task.go                                   # Storage persistence hook
```

### Phase 2 Files to Create/Modify

#### Challenger Component (NEW)
```
internal/challenger/
  ├── executor.go                               # Re-execution logic
  ├── verifier.go                               # Output comparison
  └── client.go                                 # Challenge submission
```

#### Handler Endpoints (NEW)
```
internal/dbserver/handlers/
  └── challenge.go                              # Challenge API endpoints
```

#### Docker Executor Enhancement
```
pkg/dockerexecutor/
  ├── interface.go                              # Add ExecuteWithEnv method
  └── dockerexecutor.go                         # Implement ExecuteWithEnv
```

#### Keeper Enhancements
```
internal/keeper/core/execution/
  └── custom_executor.go                        # Uncomment Phase 2 functions
```

---

## Testing Checklist

### Phase 1 Testing (Completed)
- [x] Custom job creation via API
- [x] Job stored in custom_jobs table
- [x] Scheduler fetches custom jobs
- [x] Storage fetched and included in task data
- [x] Keeper executes script successfully
- [x] JSON output parsed correctly
- [x] Storage updates extracted from JSON
- [x] On-chain transaction executed
- [x] TaskMonitor updates storage in DB
- [x] Storage persists across executions

### Phase 2 Testing (TODO)
- [ ] Environment variables injected correctly
- [ ] Scripts can read previous storage values
- [ ] Execution proofs generated
- [ ] Challenge submission works
- [ ] Validators re-execute correctly
- [ ] Output comparison accurate
- [ ] Slashing mechanism functions
- [ ] Metadata recorded completely
- [ ] Deterministic re-execution works
- [ ] Secrets encrypted/decrypted properly

---

## Migration Guide (Phase 1 → Phase 2)

### Database Migration
```sql
-- Already created in Phase 1, just activate usage:
-- custom_script_executions table
-- execution_challenges table

-- No schema changes required
```

### Code Migration Steps

1. **Docker Executor:**
   ```bash
   # Add ExecuteWithEnv to interface
   vim pkg/dockerexecutor/interface.go

   # Implement the method
   vim pkg/dockerexecutor/dockerexecutor.go
   ```

2. **Custom Executor:**
   ```bash
   # Uncomment Phase 2 functions
   vim internal/keeper/core/execution/custom_executor.go

   # Update ExecuteCustomScript to use ExecuteWithEnv
   # Uncomment: prepareCustomScriptEnv, generateExecutionProof, etc.
   ```

3. **Add Challenger:**
   ```bash
   # Create new component
   mkdir -p internal/challenger
   touch internal/challenger/{executor,verifier,client}.go
   ```

4. **Add Challenge Handler:**
   ```bash
   # Create new endpoint
   touch internal/dbserver/handlers/challenge.go

   # Register routes in router
   vim cmd/dbserver/main.go
   ```

5. **Update Keeper:**
   ```bash
   # Add proof generation after execution
   vim internal/keeper/core/execution/action.go
   ```

---

## Performance Considerations

### Phase 1
- **Storage Fetch:** O(1) per job (primary key lookup)
- **Scheduler Query:** O(N) where N = active custom jobs
- **Storage Update:** O(K) where K = number of updated keys

**Optimizations:**
- Index on `next_execution_time` for efficient scheduling
- Index on `is_active` to filter inactive jobs
- Batch storage updates if many keys modified

### Phase 2
- **Execution Record Storage:** Additional write per execution
- **Challenge Verification:** Requires re-execution (expensive)
- **Validator Consensus:** Network overhead for voting

**Optimizations:**
- Cache frequently accessed execution records
- Parallel validator execution
- Optimize deterministic replay (cache API responses)

---

## Security Considerations

### Phase 1
- ✅ Script validation before job creation
- ✅ IPFS URL validation
- ✅ Script hash stored for verification (not enforced yet)
- ⚠️ No secrets support (user scripts can't access private APIs)
- ⚠️ No execution proofs (trust-based execution)

### Phase 2
- 🔒 Script hash verification (prevent tampering)
- 🔒 Encrypted secrets (secure API key injection)
- 🔒 Execution proofs (fraud prevention)
- 🔒 Challenge mechanism (incorrect executions punished)
- 🔒 Validator consensus (decentralized verification)

**Threat Model:**
- **Malicious Performer:** Submits incorrect execution results
  - *Mitigation (Phase 2):* Challenge + slashing
- **Malicious Challenger:** False challenges to grief performers
  - *Mitigation (Phase 2):* Challenge bond + counter-slashing
- **Malicious User:** Uploads malicious script
  - *Mitigation:* Sandboxed Docker execution, resource limits
- **Script Tampering:** User changes script after job creation
  - *Mitigation (Phase 2):* Script hash verification

---
