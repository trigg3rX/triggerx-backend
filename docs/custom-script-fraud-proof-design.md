# Custom Script Execution with Fraud-Proof Verification - Design Document

## Executive Summary

This document outlines the design for TriggerX's **time-interval based custom script execution** with **fraud-proof verification**. Unlike the per-block execution model, this design uses:

1. **Time-based scheduling** (every N seconds) instead of block-based triggers
2. **Optimistic execution** with fraud proofs instead of immediate multi-validator verification
3. **Custom script output format** that returns `contractAddress` and `calldata` instead of just parameters
4. **Challenge-based verification** to reduce costs while maintaining security

---

## Table of Contents

1. [Architecture Overview](#1-architecture-overview)
2. [Job Type & Data Models](#2-job-type--data-models)
3. [Scheduler Integration](#3-scheduler-integration)
4. [Script Execution Engine](#4-script-execution-engine)
5. [Fraud-Proof System](#5-fraud-proof-system)
6. [Verification & Challenge Mechanism](#6-verification--challenge-mechanism)
7. [Smart Contract Integration](#7-smart-contract-integration)
8. [Implementation Roadmap](#8-implementation-roadmap)

---

## 1. Architecture Overview

### 1.1 High-Level Flow

```
┌──────────────────────────────────────────────────────────────────┐
│                    EVERY N SECONDS (User Defined)                 │
└──────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌──────────────────────────────────────────────────────────────────┐
│  TIME-BASED SCHEDULER (Already Exists)                           │
│                                                                    │
│  1. Checks NextExecutionTimestamp for custom jobs                 │
│  2. If current_time >= NextExecutionTimestamp:                    │
│     - Selects performer keeper (via Othentic)                     │
│     - Sends task to performer via TaskDispatcher                  │
│     - Updates NextExecutionTimestamp += TimeInterval              │
└──────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌──────────────────────────────────────────────────────────────────┐
│  PERFORMER KEEPER (Optimistic Execution)                          │
│                                                                    │
│  1. Receives custom job task from scheduler                       │
│  2. Fetches script from IPFS (CustomScriptUrl)                    │
│  3. Executes script in Docker sandbox with context:               │
│     Environment Variables:                                         │
│     - TRIGGERX_TIMESTAMP: Execution timestamp                     │
│     - TRIGGERX_JOB_ID: Job ID                                     │
│     - TRIGGERX_EXECUTION_ID: Unique execution ID                  │
│     - SECRET_*: User secrets (encrypted storage)                  │
│     - TRIGGERX_STORAGE_*: Persistent storage values               │
│                                                                    │
│  4. Parse script output (JSON):                                   │
│     {                                                              │
│       "shouldExecute": true,                                       │
│       "targetContract": "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb", │
│       "calldata": "0xabcdef...",                                   │
│       "metadata": {                                                │
│         "reason": "Condition met",                                 │
│         "gasEstimate": 150000                                      │
│       }                                                            │
│     }                                                              │
│                                                                    │
│  5. Generate Execution Proof:                                     │
│     - executionId: unique ID                                       │
│     - timestamp: execution time                                    │
│     - scriptHash: keccak256(scriptCode)                           │
│     - inputHash: keccak256(timestamp, jobId, storage)             │
│     - outputHash: keccak256(shouldExecute, targetContract, calldata) │
│     - signature: sign(inputHash + outputHash)                     │
│                                                                    │
│  6. If shouldExecute == true:                                     │
│     - Submit transaction to targetContract with calldata          │
│     - Record execution proof in database                           │
│     - Broadcast proof to Othentic for potential challenges        │
│  7. If shouldExecute == false:                                    │
│     - Record no-execution proof                                    │
│     - Update storage if needed                                     │
└──────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌──────────────────────────────────────────────────────────────────┐
│  FRAUD-PROOF CHALLENGE PERIOD (6-24 hours)                        │
│                                                                    │
│  CASE A: Random Sampling (Proactive Verification)                │
│  - 10% of executions randomly selected for verification           │
│  - Validator re-executes script with same inputs                  │
│  - Compares outputs (outputHash)                                  │
│  - If mismatch → challenge submitted                              │
│                                                                    │
│  CASE B: User Challenge (Reactive Verification)                   │
│  - Anyone can challenge execution by:                              │
│    1. Re-running script with same inputs                          │
│    2. Submitting different output to AttestationCenter            │
│    3. Providing proof of mismatch                                 │
│  - Requires small bond (prevent spam)                             │
│                                                                    │
│  CASE C: Validator Challenge (Slashing Event)                     │
│  - If validator detects wrong execution                           │
│  - Submits challenge with proof                                   │
│  - Triggers re-execution by multiple validators                   │
└──────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌──────────────────────────────────────────────────────────────────┐
│  CHALLENGE RESOLUTION (Via Othentic + AttestationCenter)         │
│                                                                    │
│  1. Challenge submitted to AttestationCenter on L2                │
│  2. Multiple validators (5-7) re-execute script                   │
│  3. Each validator submits attestation:                           │
│     - APPROVE: Performer's output matches my output               │
│     - REJECT: Performer's output differs from my output           │
│  4. Othentic aggregates BLS signatures                            │
│  5. Decision:                                                      │
│     - If >67% APPROVE → Performer rewarded, challenger slashed    │
│     - If >67% REJECT → Performer slashed, challenger rewarded     │
│                                                                    │
│  Slashing Scenarios:                                              │
│  - Wrong output: Slash 10% of performer stake                     │
│  - Missing execution: Slash 5% of performer stake                 │
│  - Invalid calldata: Slash 10% of performer stake                 │
│  - Late execution (outside tolerance): Slash 2% of stake          │
└──────────────────────────────────────────────────────────────────┘
```

### 1.2 Key Design Principles

| Principle | Rationale |
|-----------|-----------|
| **Optimistic Execution** | Execute immediately, verify only when challenged or sampled |
| **Deterministic Re-execution** | Validators can reproduce exact script output given same inputs |
| **Time-based Scheduling** | Use existing time scheduler, just add custom job support |
| **Lazy Verification** | Cheaper than immediate multi-validator (block-based approach) |
| **Economic Security** | Slashing makes fraud unprofitable |
| **Challenge Period** | Enough time for validators to detect issues (6-24 hours) |

---

## 2. Job Type & Data Models

### 2.1 Existing CustomJobData (Already in database_models.go)

```go
type CustomJobData struct {
    JobID            *BigInt   `json:"job_id"`
    TaskDefinitionID int       `json:"task_definition_id"`
    Recurring        bool      `json:"recurring"`
    CustomScriptUrl  string    `json:"custom_script_url"`  // IPFS URL
    TimeInterval     int64     `json:"time_interval"`      // Seconds
    IsCompleted      bool      `json:"is_completed"`
    IsActive         bool      `json:"is_active"`
    CreatedAt        time.Time `json:"created_at"`
    UpdatedAt        time.Time `json:"updated_at"`
    LastExecutedAt   time.Time `json:"last_executed_at"`
    ExpirationTime   time.Time `json:"expiration_time"`
}
```

### 2.2 New Fields Needed for CustomJobData

We need to extend the existing model to support fraud-proof verification:

```go
type CustomJobData struct {
    JobID            *BigInt   `json:"job_id"`
    TaskDefinitionID int       `json:"task_definition_id"`
    Recurring        bool      `json:"recurring"`
    CustomScriptUrl  string    `json:"custom_script_url"`
    TimeInterval     int64     `json:"time_interval"`
    IsCompleted      bool      `json:"is_completed"`
    IsActive         bool      `json:"is_active"`
    CreatedAt        time.Time `json:"created_at"`
    UpdatedAt        time.Time `json:"updated_at"`
    LastExecutedAt   time.Time `json:"last_executed_at"`
    ExpirationTime   time.Time `json:"expiration_time"`

    // NEW FIELDS for fraud-proof system
    ScriptLanguage      string    `json:"script_language"`      // "go", "typescript", "python"
    ScriptHash          string    `json:"script_hash"`          // keccak256(scriptCode) for verification
    NextExecutionTime   time.Time `json:"next_execution_time"`  // Scheduled next execution
    MaxExecutionTime    int64     `json:"max_execution_time"`   // Timeout in seconds
    ChallengePeriod     int64     `json:"challenge_period"`     // Seconds (default 21600 = 6 hours)
    RequiresVerification bool     `json:"requires_verification"` // If true, always verify (premium tier)
}
```

### 2.3 New Database Tables

#### A. custom_jobs Table (Update Existing)

```sql
ALTER TABLE custom_jobs ADD COLUMN script_language VARCHAR(20) DEFAULT 'typescript';
ALTER TABLE custom_jobs ADD COLUMN script_hash VARCHAR(66);
ALTER TABLE custom_jobs ADD COLUMN next_execution_time TIMESTAMP;
ALTER TABLE custom_jobs ADD COLUMN max_execution_time INT DEFAULT 60;
ALTER TABLE custom_jobs ADD COLUMN challenge_period BIGINT DEFAULT 21600; -- 6 hours
ALTER TABLE custom_jobs ADD COLUMN requires_verification BOOLEAN DEFAULT FALSE;

CREATE INDEX idx_custom_jobs_next_execution ON custom_jobs(next_execution_time) WHERE is_active = TRUE;
```

#### B. custom_script_executions Table (New)

Track every execution with proof:

```sql
CREATE TABLE custom_script_executions (
    id BIGSERIAL PRIMARY KEY,
    execution_id VARCHAR(66) NOT NULL UNIQUE,  -- Unique execution identifier
    job_id NUMERIC(78, 0) NOT NULL,
    scheduled_time TIMESTAMP NOT NULL,          -- When should execute
    actual_time TIMESTAMP NOT NULL,             -- When actually executed
    performer_address VARCHAR(42) NOT NULL,

    -- Input data (for deterministic re-execution)
    input_timestamp BIGINT NOT NULL,
    input_storage JSONB,                        -- Snapshot of storage at execution time
    input_hash VARCHAR(66) NOT NULL,

    -- Output data
    should_execute BOOLEAN NOT NULL,
    target_contract VARCHAR(42),                -- Only if shouldExecute=true
    calldata TEXT,                              -- Only if shouldExecute=true
    output_hash VARCHAR(66) NOT NULL,

    -- Proof
    script_hash VARCHAR(66) NOT NULL,
    signature TEXT NOT NULL,

    -- Execution result
    tx_hash VARCHAR(66),                        -- Transaction hash if executed
    execution_status VARCHAR(20) NOT NULL,      -- 'success', 'failed', 'no_execution'
    execution_error TEXT,

    -- Verification
    verification_status VARCHAR(20) DEFAULT 'pending',  -- 'pending', 'verified', 'challenged', 'slashed'
    challenge_deadline TIMESTAMP NOT NULL,
    is_challenged BOOLEAN DEFAULT FALSE,
    challenge_count INT DEFAULT 0,

    created_at TIMESTAMP DEFAULT NOW(),
    FOREIGN KEY (job_id) REFERENCES jobs(job_id)
);

CREATE INDEX idx_custom_executions_job_id ON custom_script_executions(job_id);
CREATE INDEX idx_custom_executions_status ON custom_script_executions(verification_status);
CREATE INDEX idx_custom_executions_deadline ON custom_script_executions(challenge_deadline);
CREATE INDEX idx_custom_executions_performer ON custom_script_executions(performer_address);
```

#### C. execution_challenges Table (New)

```sql
CREATE TABLE execution_challenges (
    id BIGSERIAL PRIMARY KEY,
    execution_id VARCHAR(66) NOT NULL,
    challenger_address VARCHAR(42) NOT NULL,
    challenge_reason VARCHAR(50) NOT NULL,      -- 'wrong_output', 'missing_execution', 'invalid_calldata'

    -- Challenger's claimed output
    challenger_output_hash VARCHAR(66) NOT NULL,
    challenger_should_execute BOOLEAN,
    challenger_target_contract VARCHAR(42),
    challenger_calldata TEXT,
    challenger_signature TEXT NOT NULL,

    -- Challenge bond
    bond_amount NUMERIC(78, 0) NOT NULL,

    -- Resolution
    resolution_status VARCHAR(20) DEFAULT 'pending',  -- 'pending', 'approved', 'rejected'
    resolution_time TIMESTAMP,
    validator_count INT DEFAULT 0,
    approve_count INT DEFAULT 0,
    reject_count INT DEFAULT 0,

    created_at TIMESTAMP DEFAULT NOW(),
    FOREIGN KEY (execution_id) REFERENCES custom_script_executions(execution_id)
);

CREATE INDEX idx_challenges_execution ON execution_challenges(execution_id);
CREATE INDEX idx_challenges_status ON execution_challenges(resolution_status);
```

#### D. script_secrets Table (Reuse from block-based design)

```sql
CREATE TABLE IF NOT EXISTS script_secrets (
    id BIGSERIAL PRIMARY KEY,
    job_id NUMERIC(78, 0) NOT NULL,
    secret_key VARCHAR(255) NOT NULL,
    secret_value_encrypted TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    FOREIGN KEY (job_id) REFERENCES jobs(job_id),
    UNIQUE(job_id, secret_key)
);
```

#### E. script_storage Table (Reuse from block-based design)

```sql
CREATE TABLE IF NOT EXISTS script_storage (
    id BIGSERIAL PRIMARY KEY,
    job_id NUMERIC(78, 0) NOT NULL,
    storage_key VARCHAR(255) NOT NULL,
    storage_value TEXT NOT NULL,
    updated_at TIMESTAMP DEFAULT NOW(),
    FOREIGN KEY (job_id) REFERENCES jobs(job_id),
    UNIQUE(job_id, storage_key)
);
```

### 2.4 Go Type Definitions

**New file: `pkg/types/custom_execution.go`**

```go
package types

import "time"

type CustomScriptExecution struct {
    ExecutionID         string    `json:"execution_id" db:"execution_id"`
    JobID               *BigInt   `json:"job_id" db:"job_id"`
    ScheduledTime       time.Time `json:"scheduled_time" db:"scheduled_time"`
    ActualTime          time.Time `json:"actual_time" db:"actual_time"`
    PerformerAddress    string    `json:"performer_address" db:"performer_address"`

    // Inputs
    InputTimestamp      int64     `json:"input_timestamp" db:"input_timestamp"`
    InputStorage        string    `json:"input_storage" db:"input_storage"` // JSONB as string
    InputHash           string    `json:"input_hash" db:"input_hash"`

    // Outputs
    ShouldExecute       bool      `json:"should_execute" db:"should_execute"`
    TargetContract      string    `json:"target_contract" db:"target_contract"`
    Calldata            string    `json:"calldata" db:"calldata"`
    OutputHash          string    `json:"output_hash" db:"output_hash"`

    // Proof
    ScriptHash          string    `json:"script_hash" db:"script_hash"`
    Signature           string    `json:"signature" db:"signature"`

    // Result
    TxHash              string    `json:"tx_hash" db:"tx_hash"`
    ExecutionStatus     string    `json:"execution_status" db:"execution_status"`
    ExecutionError      string    `json:"execution_error" db:"execution_error"`

    // Verification
    VerificationStatus  string    `json:"verification_status" db:"verification_status"`
    ChallengeDeadline   time.Time `json:"challenge_deadline" db:"challenge_deadline"`
    IsChallenged        bool      `json:"is_challenged" db:"is_challenged"`
    ChallengeCount      int       `json:"challenge_count" db:"challenge_count"`

    CreatedAt           time.Time `json:"created_at" db:"created_at"`
}

type ExecutionChallenge struct {
    ID                      int64     `json:"id" db:"id"`
    ExecutionID             string    `json:"execution_id" db:"execution_id"`
    ChallengerAddress       string    `json:"challenger_address" db:"challenger_address"`
    ChallengeReason         string    `json:"challenge_reason" db:"challenge_reason"`

    // Challenger's output
    ChallengerOutputHash    string    `json:"challenger_output_hash" db:"challenger_output_hash"`
    ChallengerShouldExecute bool      `json:"challenger_should_execute" db:"challenger_should_execute"`
    ChallengerTargetContract string   `json:"challenger_target_contract" db:"challenger_target_contract"`
    ChallengerCalldata      string    `json:"challenger_calldata" db:"challenger_calldata"`
    ChallengerSignature     string    `json:"challenger_signature" db:"challenger_signature"`

    // Bond
    BondAmount              *BigInt   `json:"bond_amount" db:"bond_amount"`

    // Resolution
    ResolutionStatus        string    `json:"resolution_status" db:"resolution_status"`
    ResolutionTime          time.Time `json:"resolution_time" db:"resolution_time"`
    ValidatorCount          int       `json:"validator_count" db:"validator_count"`
    ApproveCount            int       `json:"approve_count" db:"approve_count"`
    RejectCount             int       `json:"reject_count" db:"reject_count"`

    CreatedAt               time.Time `json:"created_at" db:"created_at"`
}

type CustomScriptOutput struct {
    ShouldExecute   bool                        `json:"shouldExecute"`
    TargetContract  string                      `json:"targetContract"`
    Calldata        string                      `json:"calldata"`
    Metadata        CustomScriptOutputMetadata  `json:"metadata"`
}

type CustomScriptOutputMetadata struct {
    Timestamp   int64    `json:"timestamp"`
    Reason      string   `json:"reason"`
    GasEstimate uint64   `json:"gasEstimate,omitempty"`
    ApiCalls    []string `json:"apiCalls,omitempty"`
}
```

---

## 3. Scheduler Integration

### 3.1 Extend Existing Time Scheduler

The time-based scheduler already exists (`internal/schedulers/time/`). We need to add support for custom jobs.

**File: `internal/schedulers/time/scheduler/custom_scheduler.go`** (New)

```go
package scheduler

import (
    "context"
    "fmt"
    "time"

    "github.com/trigg3rX/triggerx-backend/pkg/types"
    "github.com/trigg3rX/triggerx-backend/pkg/logging"
)

type CustomJobScheduler struct {
    logger          logging.Logger
    dbClient        CustomJobRepository
    taskDispatcher  TaskDispatcher
    ticker          *time.Ticker
    stopChan        chan struct{}
}

func NewCustomJobScheduler(
    logger logging.Logger,
    dbClient CustomJobRepository,
    taskDispatcher TaskDispatcher,
) *CustomJobScheduler {
    return &CustomJobScheduler{
        logger:         logger,
        dbClient:       dbClient,
        taskDispatcher: taskDispatcher,
        stopChan:       make(chan struct{}),
    }
}

func (s *CustomJobScheduler) Start(ctx context.Context) error {
    s.logger.Info("Starting custom job scheduler")

    // Check every second for jobs to execute
    s.ticker = time.NewTicker(1 * time.Second)
    defer s.ticker.Stop()

    for {
        select {
        case <-s.ticker.C:
            s.processCustomJobs(ctx)

        case <-s.stopChan:
            s.logger.Info("Custom job scheduler stopped")
            return nil

        case <-ctx.Done():
            return ctx.Err()
        }
    }
}

func (s *CustomJobScheduler) processCustomJobs(ctx context.Context) {
    // Get all custom jobs that should execute now
    jobs, err := s.dbClient.GetCustomJobsDueForExecution(time.Now())
    if err != nil {
        s.logger.Errorf("Failed to get custom jobs: %v", err)
        return
    }

    if len(jobs) == 0 {
        return
    }

    s.logger.Infof("Processing %d custom jobs", len(jobs))

    for _, job := range jobs {
        go s.scheduleCustomJob(ctx, &job)
    }
}

func (s *CustomJobScheduler) scheduleCustomJob(ctx context.Context, job *types.CustomJobData) {
    s.logger.Infof("Scheduling custom job %s for execution", job.JobID.String())

    // Create task data
    taskData := types.ScheduleCustomTaskData{
        TaskID:              generateTaskID(),
        TaskDefinitionID:    job.TaskDefinitionID,
        JobID:               job.JobID,
        CustomScriptUrl:     job.CustomScriptUrl,
        ScriptLanguage:      job.ScriptLanguage,
        ScriptHash:          job.ScriptHash,
        ScheduledTime:       job.NextExecutionTime,
        TimeInterval:        job.TimeInterval,
        ChallengePeriod:     job.ChallengePeriod,
        LastExecutedAt:      job.LastExecutedAt,
        ExpirationTime:      job.ExpirationTime,
    }

    // Send to task dispatcher (which will assign to performer)
    err := s.taskDispatcher.DispatchCustomTask(ctx, taskData)
    if err != nil {
        s.logger.Errorf("Failed to dispatch custom task %s: %v", job.JobID.String(), err)
        return
    }

    // Update next execution time
    nextExecution := job.NextExecutionTime.Add(time.Duration(job.TimeInterval) * time.Second)
    err = s.dbClient.UpdateNextExecutionTime(job.JobID, nextExecution)
    if err != nil {
        s.logger.Errorf("Failed to update next execution time for job %s: %v", job.JobID.String(), err)
    }

    s.logger.Infof("Custom job %s dispatched successfully", job.JobID.String())
}

func (s *CustomJobScheduler) Stop() {
    close(s.stopChan)
}
```

### 3.2 New Type for Custom Task Scheduling

Add to `pkg/types/schedulers.go`:

```go
type ScheduleCustomTaskData struct {
    TaskID              int64          `json:"task_id"`
    TaskDefinitionID    int            `json:"task_definition_id"`
    JobID               *BigInt        `json:"job_id"`
    CustomScriptUrl     string         `json:"custom_script_url"`
    ScriptLanguage      string         `json:"script_language"`
    ScriptHash          string         `json:"script_hash"`
    ScheduledTime       time.Time      `json:"scheduled_time"`
    TimeInterval        int64          `json:"time_interval"`
    ChallengePeriod     int64          `json:"challenge_period"`
    LastExecutedAt      time.Time      `json:"last_executed_at"`
    ExpirationTime      time.Time      `json:"expiration_time"`
}
```

---

## 4. Script Execution Engine

### 4.1 Script Format & Output

**User's Script must return JSON in this format:**

```json
{
  "shouldExecute": true,
  "targetContract": "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb",
  "calldata": "0xa9059cbb000000000000000000000000...",
  "metadata": {
    "timestamp": 1709876543,
    "reason": "Price threshold exceeded",
    "gasEstimate": 150000,
    "apiCalls": ["https://api.coingecko.com/..."]
  }
}
```

**Key Difference from Block-Based Design:**
- Block-based: Returns `shouldExecute + params` (keeper builds calldata)
- Custom script: Returns `shouldExecute + targetContract + calldata` (script builds calldata)

### 4.2 Script Execution in Keeper

**File: `internal/keeper/core/execution/custom_executor.go`** (New)

```go
package execution

import (
    "context"
    "crypto/sha256"
    "encoding/hex"
    "encoding/json"
    "fmt"
    "time"

    "github.com/ethereum/go-ethereum/crypto"
    "github.com/trigg3rX/triggerx-backend/pkg/types"
    dockertypes "github.com/trigg3rX/triggerx-backend/pkg/dockerexecutor/types"
)

type CustomScriptExecutor struct {
    dockerExecutor  *DockerExecutor
    logger          logging.Logger
    dbClient        CustomExecutionRepository
    secretsManager  SecretsManager
    storageManager  StorageManager
    txSubmitter     TransactionSubmitter
    proofBroadcaster ProofBroadcaster
}

func (cse *CustomScriptExecutor) ExecuteCustomTask(
    ctx context.Context,
    task *types.ScheduleCustomTaskData,
) error {
    executionID := generateExecutionID(task.JobID, task.ScheduledTime)

    cse.logger.Infof("Executing custom script for job %s, execution %s",
        task.JobID.String(), executionID)

    // 1. Prepare execution context
    execCtx, err := cse.prepareExecutionContext(task)
    if err != nil {
        return fmt.Errorf("failed to prepare context: %w", err)
    }

    // 2. Execute script in Docker
    scriptOutput, err := cse.executeScript(ctx, task, execCtx)
    if err != nil {
        cse.recordFailedExecution(executionID, task, err)
        return fmt.Errorf("script execution failed: %w", err)
    }

    // 3. Generate execution proof
    proof := cse.generateExecutionProof(task, execCtx, scriptOutput)

    // 4. Record execution in database
    execution := &types.CustomScriptExecution{
        ExecutionID:        executionID,
        JobID:              task.JobID,
        ScheduledTime:      task.ScheduledTime,
        ActualTime:         time.Now(),
        PerformerAddress:   cse.getPerformerAddress(),
        InputTimestamp:     execCtx.Timestamp,
        InputStorage:       cse.serializeStorage(execCtx.Storage),
        InputHash:          proof.InputHash,
        ShouldExecute:      scriptOutput.ShouldExecute,
        TargetContract:     scriptOutput.TargetContract,
        Calldata:           scriptOutput.Calldata,
        OutputHash:         proof.OutputHash,
        ScriptHash:         task.ScriptHash,
        Signature:          proof.Signature,
        ExecutionStatus:    "pending",
        VerificationStatus: "pending",
        ChallengeDeadline:  time.Now().Add(time.Duration(task.ChallengePeriod) * time.Second),
    }

    err = cse.dbClient.RecordExecution(execution)
    if err != nil {
        return fmt.Errorf("failed to record execution: %w", err)
    }

    // 5. If shouldExecute, submit transaction
    if scriptOutput.ShouldExecute {
        txHash, err := cse.submitTransaction(ctx, scriptOutput, task)
        if err != nil {
            cse.logger.Errorf("Failed to submit transaction: %v", err)
            cse.updateExecutionStatus(executionID, "failed", err.Error())
            return err
        }

        cse.updateExecutionTxHash(executionID, txHash)
        cse.logger.Infof("Transaction submitted: %s", txHash)
    }

    // 6. Broadcast proof to Othentic for potential challenges
    err = cse.proofBroadcaster.BroadcastExecutionProof(proof)
    if err != nil {
        cse.logger.Warnf("Failed to broadcast proof: %v", err)
        // Non-critical, continue
    }

    // 7. Update storage if script modified it
    cse.updateStorage(task.JobID, execCtx.StorageUpdates)

    cse.logger.Infof("Custom script execution completed for %s", executionID)
    return nil
}

type ExecutionContext struct {
    Timestamp      int64
    JobID          string
    ExecutionID    string
    Secrets        map[string]string
    Storage        map[string]string
    StorageUpdates map[string]string  // Populated after script execution
}

func (cse *CustomScriptExecutor) prepareExecutionContext(
    task *types.ScheduleCustomTaskData,
) (*ExecutionContext, error) {
    // Get secrets
    secrets, err := cse.secretsManager.GetSecrets(task.JobID)
    if err != nil {
        cse.logger.Warnf("Failed to get secrets: %v", err)
        secrets = make(map[string]string)
    }

    // Get storage
    storage, err := cse.storageManager.GetStorage(task.JobID)
    if err != nil {
        cse.logger.Warnf("Failed to get storage: %v", err)
        storage = make(map[string]string)
    }

    return &ExecutionContext{
        Timestamp:      time.Now().Unix(),
        JobID:          task.JobID.String(),
        ExecutionID:    generateExecutionID(task.JobID, task.ScheduledTime),
        Secrets:        secrets,
        Storage:        storage,
        StorageUpdates: make(map[string]string),
    }, nil
}

func (cse *CustomScriptExecutor) executeScript(
    ctx context.Context,
    task *types.ScheduleCustomTaskData,
    execCtx *ExecutionContext,
) (*types.CustomScriptOutput, error) {
    // Prepare environment variables
    envVars := map[string]string{
        "TRIGGERX_TIMESTAMP":    fmt.Sprintf("%d", execCtx.Timestamp),
        "TRIGGERX_JOB_ID":       execCtx.JobID,
        "TRIGGERX_EXECUTION_ID": execCtx.ExecutionID,
    }

    // Add secrets
    for key, value := range execCtx.Secrets {
        envVars["SECRET_"+key] = value
    }

    // Add storage
    for key, value := range execCtx.Storage {
        envVars["TRIGGERX_STORAGE_"+key] = value
    }

    // Execute script in Docker
    result, err := cse.dockerExecutor.ExecuteWithEnv(
        ctx,
        task.CustomScriptUrl,
        task.ScriptLanguage,
        envVars,
    )
    if err != nil {
        return nil, fmt.Errorf("docker execution failed: %w", err)
    }

    // Parse storage updates from stderr
    execCtx.StorageUpdates = parseStorageUpdates(result.Stderr)

    // Parse output JSON
    var output types.CustomScriptOutput
    err = json.Unmarshal([]byte(result.Output), &output)
    if err != nil {
        return nil, fmt.Errorf("failed to parse script output: %w", err)
    }

    // Validate output
    if err := cse.validateOutput(&output); err != nil {
        return nil, fmt.Errorf("invalid output: %w", err)
    }

    return &output, nil
}

func (cse *CustomScriptExecutor) validateOutput(output *types.CustomScriptOutput) error {
    if output.ShouldExecute {
        if output.TargetContract == "" {
            return fmt.Errorf("targetContract required when shouldExecute=true")
        }
        if output.Calldata == "" {
            return fmt.Errorf("calldata required when shouldExecute=true")
        }
        // Validate address format
        if !isValidAddress(output.TargetContract) {
            return fmt.Errorf("invalid targetContract address")
        }
        // Validate calldata format
        if !isValidCalldata(output.Calldata) {
            return fmt.Errorf("invalid calldata format")
        }
    }
    return nil
}

type ExecutionProof struct {
    ExecutionID     string
    JobID           string
    Timestamp       int64
    ScriptHash      string
    InputHash       string
    OutputHash      string
    Signature       string
    PerformerAddress string
}

func (cse *CustomScriptExecutor) generateExecutionProof(
    task *types.ScheduleCustomTaskData,
    execCtx *ExecutionContext,
    output *types.CustomScriptOutput,
) *ExecutionProof {
    // Calculate input hash
    inputData := fmt.Sprintf("%d:%s:%s",
        execCtx.Timestamp,
        execCtx.JobID,
        cse.hashStorage(execCtx.Storage),
    )
    inputHash := crypto.Keccak256Hash([]byte(inputData)).Hex()

    // Calculate output hash
    outputData := fmt.Sprintf("%t:%s:%s",
        output.ShouldExecute,
        output.TargetContract,
        output.Calldata,
    )
    outputHash := crypto.Keccak256Hash([]byte(outputData)).Hex()

    // Sign (inputHash + outputHash)
    signature := cse.signProof(inputHash, outputHash)

    return &ExecutionProof{
        ExecutionID:      execCtx.ExecutionID,
        JobID:            task.JobID.String(),
        Timestamp:        execCtx.Timestamp,
        ScriptHash:       task.ScriptHash,
        InputHash:        inputHash,
        OutputHash:       outputHash,
        Signature:        signature,
        PerformerAddress: cse.getPerformerAddress(),
    }
}

func (cse *CustomScriptExecutor) submitTransaction(
    ctx context.Context,
    output *types.CustomScriptOutput,
    task *types.ScheduleCustomTaskData,
) (string, error) {
    // Submit transaction to targetContract with calldata
    // This uses the existing transaction submission infrastructure

    txHash, err := cse.txSubmitter.SubmitTransaction(ctx, &TransactionRequest{
        To:       output.TargetContract,
        Data:     output.Calldata,
        JobID:    task.JobID,
        GasLimit: output.Metadata.GasEstimate,
    })

    return txHash, err
}
```

---

## 5. Fraud-Proof System

### 5.1 Core Fraud-Proof Concepts

**Execution is "Optimistic":**
- Performer executes immediately without waiting for validators
- Lower latency and cost compared to immediate consensus
- Security comes from:
  1. Economic penalties (slashing)
  2. Challenge period (6-24 hours)
  3. Verifiable re-execution

**Challenge Window:**
```
Execution happens → Challenge period starts (6-24h) → Window closes
    |                         |                            |
    |                         |                            |
    v                         v                            v
Performer                 Anyone can                  Execution
executes                  challenge by                finalized
                          re-running script
```

### 5.2 Challenge Types

| Challenge Type | Description | Challenger | Bond Required |
|----------------|-------------|------------|---------------|
| **Random Sampling** | 10% of executions randomly verified | Validator | No bond (duty) |
| **User Challenge** | User disputes execution | Any address | Yes (prevent spam) |
| **Validator Challenge** | Validator detects fraud | Validator | No bond (duty) |
| **Missing Execution** | Should have executed but didn't | Anyone | Small bond |

### 5.3 Challenge Resolution Flow

**File: `internal/keeper/core/validation/challenge_handler.go`** (New)

```go
package validation

import (
    "context"
    "fmt"
    "time"

    "github.com/trigg3rX/triggerx-backend/pkg/types"
)

type ChallengeHandler struct {
    logger              logging.Logger
    dbClient            ChallengeRepository
    scriptExecutor      *CustomScriptExecutor
    attestationClient   *AttestationClient
    slashingManager     *SlashingManager
}

func (ch *ChallengeHandler) HandleChallenge(
    ctx context.Context,
    challenge *types.ExecutionChallenge,
) error {
    ch.logger.Infof("Handling challenge %d for execution %s",
        challenge.ID, challenge.ExecutionID)

    // 1. Get original execution data
    execution, err := ch.dbClient.GetExecution(challenge.ExecutionID)
    if err != nil {
        return fmt.Errorf("failed to get execution: %w", err)
    }

    // 2. Check if still in challenge period
    if time.Now().After(execution.ChallengeDeadline) {
        return fmt.Errorf("challenge period expired")
    }

    // 3. Trigger multi-validator re-execution
    validators, err := ch.selectValidators(5) // Select 5 validators
    if err != nil {
        return fmt.Errorf("failed to select validators: %w", err)
    }

    // 4. Send re-execution request to validators
    attestations := make([]*Attestation, 0)
    for _, validator := range validators {
        attestation, err := ch.requestValidatorAttestation(ctx, validator, execution)
        if err != nil {
            ch.logger.Errorf("Validator %s failed to attest: %v", validator, err)
            continue
        }
        attestations = append(attestations, attestation)
    }

    // 5. Aggregate attestations
    approveCount := 0
    rejectCount := 0

    for _, att := range attestations {
        if att.Approved {
            approveCount++
        } else {
            rejectCount++
        }
    }

    ch.logger.Infof("Attestation result: %d approve, %d reject", approveCount, rejectCount)

    // 6. Resolve challenge (>67% threshold)
    threshold := len(attestations) * 2 / 3

    if approveCount > threshold {
        // Performer was correct
        ch.approveChallenge(challenge, execution, approveCount, rejectCount)
        ch.slashChallenger(challenge) // Slash false challenger
    } else if rejectCount > threshold {
        // Performer was wrong
        ch.rejectChallenge(challenge, execution, approveCount, rejectCount)
        ch.slashPerformer(execution, challenge.ChallengeReason) // Slash performer
    } else {
        // No consensus - inconclusive
        ch.markChallengeInconclusive(challenge)
    }

    return nil
}

func (ch *ChallengeHandler) requestValidatorAttestation(
    ctx context.Context,
    validatorAddress string,
    execution *types.CustomScriptExecution,
) (*Attestation, error) {
    // 1. Get validator client
    validator := ch.getValidatorClient(validatorAddress)

    // 2. Request re-execution
    request := &ValidationRequest{
        ExecutionID:     execution.ExecutionID,
        JobID:           execution.JobID,
        ScriptHash:      execution.ScriptHash,
        InputTimestamp:  execution.InputTimestamp,
        InputStorage:    execution.InputStorage,
        PerformerOutput: PerformerOutput{
            ShouldExecute:  execution.ShouldExecute,
            TargetContract: execution.TargetContract,
            Calldata:       execution.Calldata,
            OutputHash:     execution.OutputHash,
        },
    }

    attestation, err := validator.ValidateExecution(ctx, request)
    if err != nil {
        return nil, err
    }

    return attestation, nil
}

type Attestation struct {
    ValidatorAddress string
    ExecutionID      string
    Approved         bool
    Reason           string
    OutputHash       string
    Signature        string
}
```

### 5.4 Validator Re-Execution

**File: `internal/keeper/core/validation/custom_validator.go`** (New)

```go
package validation

type CustomScriptValidator struct {
    scriptExecutor *CustomScriptExecutor
    logger         logging.Logger
}

func (csv *CustomScriptValidator) ValidateExecution(
    ctx context.Context,
    req *ValidationRequest,
) (*Attestation, error) {
    csv.logger.Infof("Validating execution %s", req.ExecutionID)

    // 1. Get script from IPFS
    scriptURL := csv.getScriptURL(req.JobID)

    // 2. Reconstruct execution context with SAME inputs
    execCtx := &ExecutionContext{
        Timestamp:   req.InputTimestamp,
        JobID:       req.JobID.String(),
        ExecutionID: req.ExecutionID,
        Storage:     csv.deserializeStorage(req.InputStorage),
        Secrets:     csv.getSecrets(req.JobID), // Same secrets
    }

    // 3. Re-execute script
    output, err := csv.scriptExecutor.executeScript(ctx, &types.ScheduleCustomTaskData{
        JobID:           req.JobID,
        CustomScriptUrl: scriptURL,
        ScriptLanguage:  csv.getScriptLanguage(req.JobID),
    }, execCtx)

    if err != nil {
        csv.logger.Errorf("Re-execution failed: %v", err)
        return csv.createRejectionAttestation(req, "re-execution failed"), nil
    }

    // 4. Calculate output hash
    validatorOutputHash := csv.calculateOutputHash(
        output.ShouldExecute,
        output.TargetContract,
        output.Calldata,
    )

    // 5. Compare with performer's output hash
    if validatorOutputHash != req.PerformerOutput.OutputHash {
        csv.logger.Warnf("Output mismatch! Performer: %s, Validator: %s",
            req.PerformerOutput.OutputHash, validatorOutputHash)

        return csv.createRejectionAttestation(req, fmt.Sprintf(
            "output hash mismatch: performer=%s, validator=%s",
            req.PerformerOutput.OutputHash, validatorOutputHash,
        )), nil
    }

    // 6. Verify on-chain transaction (if shouldExecute=true)
    if output.ShouldExecute {
        valid, err := csv.verifyOnChainTransaction(req, output)
        if err != nil || !valid {
            return csv.createRejectionAttestation(req, "transaction verification failed"), nil
        }
    }

    // 7. All checks passed - approve
    csv.logger.Infof("Validation passed for execution %s", req.ExecutionID)
    return csv.createApprovalAttestation(req, validatorOutputHash), nil
}

func (csv *CustomScriptValidator) createApprovalAttestation(
    req *ValidationRequest,
    outputHash string,
) *Attestation {
    attestation := &Attestation{
        ValidatorAddress: csv.getValidatorAddress(),
        ExecutionID:      req.ExecutionID,
        Approved:         true,
        Reason:           "validation passed",
        OutputHash:       outputHash,
    }

    // Sign attestation
    attestation.Signature = csv.signAttestation(attestation)

    return attestation
}

func (csv *CustomScriptValidator) createRejectionAttestation(
    req *ValidationRequest,
    reason string,
) *Attestation {
    attestation := &Attestation{
        ValidatorAddress: csv.getValidatorAddress(),
        ExecutionID:      req.ExecutionID,
        Approved:         false,
        Reason:           reason,
    }

    attestation.Signature = csv.signAttestation(attestation)

    return attestation
}

func (csv *CustomScriptValidator) verifyOnChainTransaction(
    req *ValidationRequest,
    output *types.CustomScriptOutput,
) (bool, error) {
    // 1. Get execution record
    execution, err := csv.dbClient.GetExecution(req.ExecutionID)
    if err != nil {
        return false, err
    }

    if execution.TxHash == "" {
        return false, fmt.Errorf("no transaction hash recorded")
    }

    // 2. Fetch transaction from chain
    tx, err := csv.getTransaction(execution.TxHash)
    if err != nil {
        return false, err
    }

    // 3. Verify transaction details
    if tx.To.Hex() != output.TargetContract {
        return false, fmt.Errorf("target contract mismatch")
    }

    if hex.EncodeToString(tx.Data) != output.Calldata {
        return false, fmt.Errorf("calldata mismatch")
    }

    // 4. Verify transaction was included in block
    if tx.BlockNumber == nil {
        return false, fmt.Errorf("transaction not mined")
    }

    return true, nil
}
```

---

## 6. Verification & Challenge Mechanism

### 6.1 Proactive Verification (Random Sampling)

**File: `internal/keeper/core/validation/sampler.go`** (New)

```go
package validation

type ExecutionSampler struct {
    logger          logging.Logger
    dbClient        ExecutionRepository
    validator       *CustomScriptValidator
    samplingRate    float64 // 0.1 = 10%
}

func (es *ExecutionSampler) Start(ctx context.Context) error {
    es.logger.Info("Starting execution sampler")

    ticker := time.NewTicker(1 * time.Minute)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            es.sampleExecutions(ctx)

        case <-ctx.Done():
            return ctx.Err()
        }
    }
}

func (es *ExecutionSampler) sampleExecutions(ctx context.Context) {
    // Get recent unverified executions
    executions, err := es.dbClient.GetUnverifiedExecutions(100)
    if err != nil {
        es.logger.Errorf("Failed to get unverified executions: %v", err)
        return
    }

    // Randomly sample
    for _, exec := range executions {
        // 10% chance of verification
        if rand.Float64() < es.samplingRate {
            go es.verifyExecution(ctx, &exec)
        }
    }
}

func (es *ExecutionSampler) verifyExecution(
    ctx context.Context,
    execution *types.CustomScriptExecution,
) {
    es.logger.Infof("Sampling execution %s for verification", execution.ExecutionID)

    // Re-execute and verify
    attestation, err := es.validator.ValidateExecution(ctx, &ValidationRequest{
        ExecutionID:     execution.ExecutionID,
        JobID:           execution.JobID,
        ScriptHash:      execution.ScriptHash,
        InputTimestamp:  execution.InputTimestamp,
        InputStorage:    execution.InputStorage,
        PerformerOutput: PerformerOutput{
            ShouldExecute:  execution.ShouldExecute,
            TargetContract: execution.TargetContract,
            Calldata:       execution.Calldata,
            OutputHash:     execution.OutputHash,
        },
    })

    if err != nil {
        es.logger.Errorf("Validation failed: %v", err)
        return
    }

    // If rejected, submit challenge
    if !attestation.Approved {
        es.submitChallenge(execution, attestation)
    } else {
        // Mark as verified
        es.dbClient.UpdateVerificationStatus(execution.ExecutionID, "verified")
    }
}

func (es *ExecutionSampler) submitChallenge(
    execution *types.CustomScriptExecution,
    attestation *Attestation,
) {
    challenge := &types.ExecutionChallenge{
        ExecutionID:       execution.ExecutionID,
        ChallengerAddress: es.getValidatorAddress(),
        ChallengeReason:   "wrong_output",
        ChallengerOutputHash: attestation.OutputHash,
        ResolutionStatus:  "pending",
    }

    err := es.dbClient.SubmitChallenge(challenge)
    if err != nil {
        es.logger.Errorf("Failed to submit challenge: %v", err)
    } else {
        es.logger.Infof("Challenge submitted for execution %s", execution.ExecutionID)
    }
}
```

### 6.2 Challenge Deadline Monitor

**File: `internal/keeper/core/validation/deadline_monitor.go`** (New)

```go
package validation

type ChallengeDeadlineMonitor struct {
    logger   logging.Logger
    dbClient ExecutionRepository
}

func (cdm *ChallengeDeadlineMonitor) Start(ctx context.Context) error {
    cdm.logger.Info("Starting challenge deadline monitor")

    ticker := time.NewTicker(5 * time.Minute)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            cdm.checkDeadlines(ctx)

        case <-ctx.Done():
            return ctx.Err()
        }
    }
}

func (cdm *ChallengeDeadlineMonitor) checkDeadlines(ctx context.Context) {
    // Get executions with expired challenge periods
    executions, err := cdm.dbClient.GetExecutionsWithExpiredChallenges()
    if err != nil {
        cdm.logger.Errorf("Failed to get expired challenges: %v", err)
        return
    }

    for _, exec := range executions {
        // If no challenges and deadline passed → mark as finalized
        if exec.ChallengeCount == 0 {
            err := cdm.dbClient.FinalizeExecution(exec.ExecutionID)
            if err != nil {
                cdm.logger.Errorf("Failed to finalize execution %s: %v", exec.ExecutionID, err)
            } else {
                cdm.logger.Infof("Execution %s finalized (no challenges)", exec.ExecutionID)
            }
        }
    }
}
```

---

## 7. Smart Contract Integration

### 7.1 AttestationCenter Extension for Custom Scripts

**Update: `contracts/AttestationCenter.sol`**

```solidity
// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

contract AttestationCenter {
    // ... existing code ...

    // NEW: Custom script execution attestations
    struct CustomExecutionAttestation {
        bytes32 executionId;
        uint256 jobId;
        uint64 timestamp;
        bytes32 outputHash;
        bool approved;
        address[] validators;
        bytes aggregatedSignature;
        uint256 createdAt;
    }

    mapping(bytes32 => CustomExecutionAttestation) public customExecutionAttestations;
    mapping(bytes32 => bool) public isExecutionChallenged;

    event CustomExecutionAttested(
        bytes32 indexed executionId,
        uint256 indexed jobId,
        bytes32 outputHash,
        bool approved,
        uint8 validatorCount
    );

    event CustomExecutionChallenged(
        bytes32 indexed executionId,
        address indexed challenger,
        string reason
    );

    event ChallengeResolved(
        bytes32 indexed executionId,
        bool performerCorrect,
        address slashedParty
    );

    // Submit aggregated attestation for custom script execution
    function submitCustomExecutionAttestation(
        bytes32 executionId,
        uint256 jobId,
        uint64 timestamp,
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

        // Store attestation
        customExecutionAttestations[executionId] = CustomExecutionAttestation({
            executionId: executionId,
            jobId: jobId,
            timestamp: timestamp,
            outputHash: outputHash,
            approved: approved,
            validators: validators,
            aggregatedSignature: aggregatedSignature,
            createdAt: block.timestamp
        });

        emit CustomExecutionAttested(
            executionId,
            jobId,
            outputHash,
            approved,
            uint8(validators.length)
        );

        // If rejected, trigger slashing
        if (!approved) {
            _slashPerformer(executionId);
        }
    }

    // Submit challenge to custom execution
    function challengeCustomExecution(
        bytes32 executionId,
        bytes32 challengerOutputHash,
        string calldata reason
    ) external payable {
        require(msg.value >= CHALLENGE_BOND, "Insufficient challenge bond");
        require(!isExecutionChallenged[executionId], "Already challenged");

        isExecutionChallenged[executionId] = true;

        emit CustomExecutionChallenged(executionId, msg.sender, reason);

        // Trigger re-validation by validators
        _requestReValidation(executionId);
    }

    // Query if execution was validated
    function isCustomExecutionValid(bytes32 executionId) external view returns (bool) {
        CustomExecutionAttestation memory attestation = customExecutionAttestations[executionId];
        return attestation.approved && attestation.validators.length >= minValidatorThreshold;
    }

    function _slashPerformer(bytes32 executionId) internal {
        // TODO: Implement slashing logic
        // This will integrate with the existing slashing contract
    }

    function _requestReValidation(bytes32 executionId) internal {
        // TODO: Request validators to re-execute and attest
    }
}
```

### 7.2 Transaction Execution Contract

Since custom scripts return `targetContract` and `calldata`, we need a way to execute these safely.

**New Contract: `contracts/CustomScriptExecutor.sol`**

```solidity
// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import "./AttestationCenter.sol";
import "./TriggerGasRegistry.sol";

contract CustomScriptExecutor {
    AttestationCenter public attestationCenter;
    TriggerGasRegistry public triggerGasRegistry;

    mapping(bytes32 => bool) public usedExecutionIds;

    event CustomScriptExecuted(
        bytes32 indexed executionId,
        uint256 indexed jobId,
        address targetContract,
        bool success
    );

    modifier onlyKeeper() {
        // TODO: Implement keeper authorization
        _;
    }

    // Execute custom script result
    function executeCustomScript(
        bytes32 executionId,
        uint256 jobId,
        address targetContract,
        bytes calldata data,
        bytes32 outputHash,
        bytes calldata signature
    ) external onlyKeeper {
        // 1. Verify execution ID not used
        require(!usedExecutionIds[executionId], "Execution already used");

        // 2. Verify signature
        require(_verifySignature(executionId, outputHash, signature), "Invalid signature");

        // 3. Mark as used
        usedExecutionIds[executionId] = true;

        // 4. Deduct TG balance (gas prepayment)
        address jobOwner = _getJobOwner(jobId);
        uint256 tgAmount = _estimateTGCost(data);
        triggerGasRegistry.deductTGBalance(jobOwner, tgAmount);

        // 5. Execute call to target contract
        (bool success, ) = targetContract.call(data);

        emit CustomScriptExecuted(executionId, jobId, targetContract, success);

        require(success, "Target contract execution failed");
    }

    function _verifySignature(
        bytes32 executionId,
        bytes32 outputHash,
        bytes calldata signature
    ) internal view returns (bool) {
        // Verify performer's signature
        // TODO: Implement signature verification
        return true;
    }

    function _getJobOwner(uint256 jobId) internal view returns (address) {
        // TODO: Get job owner from JobRegistry
        return address(0);
    }

    function _estimateTGCost(bytes calldata data) internal pure returns (uint256) {
        // Estimate TG cost based on calldata size
        return data.length * 10; // Simple estimate
    }
}
```

---

## 8. Implementation Roadmap

### Phase 1: Core Infrastructure (Weeks 1-3)

**Week 1: Database & Data Models**
- [ ] Update `custom_jobs` table schema
- [ ] Create `custom_script_executions` table
- [ ] Create `execution_challenges` table
- [ ] Create `script_secrets` and `script_storage` tables
- [ ] Implement Go types in `pkg/types/custom_execution.go`
- [ ] Create repository interfaces and implementations

**Week 2: Scheduler Integration**
- [ ] Implement `CustomJobScheduler` in `internal/schedulers/time/scheduler/custom_scheduler.go`
- [ ] Add `ScheduleCustomTaskData` type
- [ ] Integrate with existing time-based scheduler
- [ ] Add database queries for jobs due for execution
- [ ] Test scheduling flow

**Week 3: Script Execution Engine**
- [ ] Implement `CustomScriptExecutor` in `internal/keeper/core/execution/custom_executor.go`
- [ ] Add script execution context preparation
- [ ] Integrate with Docker executor (add env vars support)
- [ ] Implement script output parsing and validation
- [ ] Add execution proof generation
- [ ] Test script execution with TypeScript/Go/Python samples

### Phase 2: Fraud-Proof System (Weeks 4-6)

**Week 4: Execution Recording & Proofs**
- [ ] Implement execution recording in database
- [ ] Add proof generation logic
- [ ] Implement transaction submission (if shouldExecute=true)
- [ ] Add proof broadcasting to Othentic
- [ ] Create execution monitoring dashboard

**Week 5: Validation & Re-execution**
- [ ] Implement `CustomScriptValidator` in `internal/keeper/core/validation/custom_validator.go`
- [ ] Add deterministic re-execution logic
- [ ] Implement output comparison
- [ ] Add on-chain transaction verification
- [ ] Test validation flow with sample executions

**Week 6: Challenge System**
- [ ] Implement `ChallengeHandler` in `internal/keeper/core/validation/challenge_handler.go`
- [ ] Add multi-validator attestation aggregation
- [ ] Implement challenge resolution logic
- [ ] Add slashing triggers
- [ ] Create challenge submission API

### Phase 3: Smart Contracts (Weeks 7-8)

**Week 7: AttestationCenter Extension**
- [ ] Extend `AttestationCenter.sol` with custom execution attestations
- [ ] Add challenge submission functions
- [ ] Implement BLS signature verification for custom executions
- [ ] Add slashing logic
- [ ] Write unit tests

**Week 8: CustomScriptExecutor Contract**
- [ ] Deploy `CustomScriptExecutor.sol`
- [ ] Integrate with `TriggerGasRegistry`
- [ ] Add signature verification
- [ ] Implement safe call execution
- [ ] Test on Base Sepolia testnet

### Phase 4: Monitoring & Optimization (Weeks 9-10)

**Week 9: Proactive Verification**
- [ ] Implement `ExecutionSampler` (10% random sampling)
- [ ] Add `ChallengeDeadlineMonitor`
- [ ] Create automated challenge submission
- [ ] Add metrics and logging
- [ ] Test sampling and challenge flows

**Week 10: Testing & Optimization**
- [ ] End-to-end testing (100+ custom jobs)
- [ ] Load testing (concurrent executions)
- [ ] Security audit (internal)
- [ ] Performance optimization
- [ ] Documentation

### Success Criteria

| Metric | Target |
|--------|--------|
| **Execution Latency** | <10s from scheduled time |
| **Validation Accuracy** | >99.9% |
| **False Positive Rate** | <0.1% |
| **Challenge Resolution Time** | <1 hour |
| **Slashing Accuracy** | 100% correct |
| **Uptime** | >99.5% |

---

## 9. Cost Model & Economics

### 9.1 Cost Components

| Component | Cost per Execution | Notes |
|-----------|-------------------|-------|
| Script Execution | $0.0005 | Docker container |
| Storage Ops | $0.0001 | Read/write persistent storage |
| Proof Generation | $0.0001 | Hashing + signing |
| Transaction Gas | Variable | User pays (deducted from TG) |
| Random Sampling (10%) | $0.0005 | Validator re-execution |
| Challenge Resolution | $0.0025 | 5 validators re-execute |
| Protocol Fee | 3% | TriggerX margin |

**Total Cost (No Challenge): ~$0.0007 + gas**
**Total Cost (With Challenge): ~$0.0032 + gas**

### 9.2 Comparison with Block-Based Design

| Aspect | Block-Based (Docs) | Custom Script (This Design) |
|--------|-------------------|----------------------------|
| Execution Frequency | Every N blocks | Every N seconds |
| Verification Cost | High (3-5 validators always) | Low (10% sampling + challenges) |
| Latency | ~2-12s (block time) | <10s (scheduled) |
| Cost per Execution | $0.004 + gas | $0.0007 + gas |
| Security | Immediate consensus | Optimistic + fraud proofs |
| Best For | High-value, frequent | Medium-value, scheduled |

---

## 10. Security Considerations

### 10.1 Attack Vectors & Mitigations

| Attack | Description | Mitigation |
|--------|-------------|-----------|
| **False Execution** | Performer executes wrong calldata | Challenge + multi-validator verification + slashing |
| **Missing Execution** | Performer doesn't execute when should | Monitoring + challenge + slashing |
| **Non-deterministic Script** | Script gives different outputs | Static analysis + determinism checks |
| **Replay Attack** | Reuse old execution proof | Track used execution IDs on-chain |
| **Challenge Spam** | Flood with fake challenges | Challenge bond requirement |
| **Validator Collusion** | Validators collude with performer | BLS threshold (>67% needed) + random sampling |
| **Script Exploitation** | Malicious script code | Docker sandbox + resource limits + seccomp |

### 10.2 Slashing Schedule

| Offense | First Time | Second Time | Third Time |
|---------|-----------|-------------|-----------|
| Wrong Output | 10% stake | 25% stake | 100% stake + ban |
| Missing Execution | 5% stake | 15% stake | 50% stake + ban |
| Invalid Calldata | 10% stake | 25% stake | 100% stake + ban |
| Late Execution (>5min) | 2% stake | 5% stake | 15% stake |

### 10.3 Determinism Enforcement

**Static Analysis (Pre-deployment):**
```go
func validateScriptDeterminism(scriptCode string) error {
    // Detect non-deterministic patterns
    forbiddenPatterns := []string{
        "time.Now()",
        "Date.now()",
        "random()",
        "Math.random()",
        "os.urandom",
    }

    for _, pattern := range forbiddenPatterns {
        if strings.Contains(scriptCode, pattern) {
            return fmt.Errorf("non-deterministic pattern detected: %s", pattern)
        }
    }

    return nil
}
```

**Runtime Sandbox:**
- Disable network access (unless TLS proof mode)
- Override time functions to use `TRIGGERX_TIMESTAMP`
- Limit entropy sources

---

## 11. Example Use Cases

### Example 1: Automated Liquidation Bot

**Script (TypeScript):**
```typescript
// TRIGGERX_TIMESTAMP, TRIGGERX_JOB_ID provided as env vars

import { ethers } from 'ethers';

async function main() {
  const provider = new ethers.JsonRpcProvider(process.env.RPC_URL);
  const lendingContract = new ethers.Contract(LENDING_ADDRESS, ABI, provider);

  // Check user's collateral ratio
  const position = await lendingContract.getPosition(USER_ADDRESS);
  const collateralValue = position.collateral * await getPrice('ETH');
  const debtValue = position.debt * await getPrice('USDC');
  const ratio = collateralValue / debtValue;

  if (ratio < 1.2) {
    // Build liquidation calldata
    const iface = new ethers.Interface(LENDING_ABI);
    const calldata = iface.encodeFunctionData('liquidate', [USER_ADDRESS, position.debt]);

    console.log(JSON.stringify({
      shouldExecute: true,
      targetContract: LENDING_ADDRESS,
      calldata: calldata,
      metadata: {
        timestamp: parseInt(process.env.TRIGGERX_TIMESTAMP),
        reason: `Collateral ratio ${ratio} below 1.2`,
        gasEstimate: 150000
      }
    }));
  } else {
    console.log(JSON.stringify({
      shouldExecute: false,
      targetContract: "",
      calldata: "",
      metadata: {
        timestamp: parseInt(process.env.TRIGGERX_TIMESTAMP),
        reason: `Collateral ratio ${ratio} healthy`
      }
    }));
  }
}

main();
```

**Job Configuration:**
- TimeInterval: 300 (every 5 minutes)
- ChallengePeriod: 21600 (6 hours)
- RequiresVerification: false (use sampling)

### Example 2: DAO Proposal Execution

**Script (Go):**
```go
package main

import (
    "encoding/json"
    "fmt"
    "os"
)

func main() {
    timestamp := os.Getenv("TRIGGERX_TIMESTAMP")

    // Check if proposal timelock has expired
    proposal := getProposal(PROPOSAL_ID)

    if proposal.State == "Queued" && proposal.ETA <= timestamp {
        // Build execution calldata
        calldata := encodeExecuteProposal(PROPOSAL_ID)

        output := map[string]interface{}{
            "shouldExecute": true,
            "targetContract": GOV_CONTRACT,
            "calldata": calldata,
            "metadata": map[string]interface{}{
                "timestamp": timestamp,
                "reason": "Proposal timelock expired",
                "gasEstimate": 200000,
            },
        }

        json.NewEncoder(os.Stdout).Encode(output)
    } else {
        output := map[string]interface{}{
            "shouldExecute": false,
            "targetContract": "",
            "calldata": "",
            "metadata": map[string]interface{}{
                "timestamp": timestamp,
                "reason": fmt.Sprintf("Proposal state: %s, ETA: %d", proposal.State, proposal.ETA),
            },
        }
        json.NewEncoder(os.Stdout).Encode(output)
    }
}
```

---

## 12. Summary

### Key Design Decisions

1. **Time-based scheduling** instead of block-based (leverage existing scheduler)
2. **Optimistic execution** with fraud proofs instead of immediate consensus
3. **Custom script output format**: `shouldExecute + targetContract + calldata`
4. **10% random sampling** + challenge-based verification
5. **6-24 hour challenge period** for fraud detection
6. **Economic security** through slashing (10% stake for wrong execution)

### Benefits Over Block-Based Design

| Benefit | Impact |
|---------|--------|
| **Lower Cost** | 80% cheaper ($0.0007 vs $0.004) |
| **Flexible Timing** | Seconds instead of block granularity |
| **Simpler Scripts** | Script builds calldata, not keeper |
| **Scalable** | Less validator overhead |

### Trade-offs

| Trade-off | Block-Based | Custom Script (This) |
|-----------|-------------|---------------------|
| **Security Guarantee** | Immediate (consensus) | Delayed (challenge period) |
| **Finality Time** | ~1 minute | 6-24 hours |
| **Cost** | Higher | Lower |
| **Best For** | High-value MEV | Scheduled automation |

---

## 13. Next Steps

1. **Review this design** with team
2. **Approve architecture** decisions
3. **Start Phase 1** implementation (database + scheduler)
4. **Build MVP** with 1-2 sample scripts
5. **Test on Base Sepolia** testnet
6. **Iterate** based on findings
7. **Launch beta** with selected users

---

**Document Version:** 1.0
**Date:** 2025-11-16
**Author:** TriggerX Engineering Team
**Status:** Design Ready for Review
