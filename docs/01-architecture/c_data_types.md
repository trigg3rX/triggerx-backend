# Data Types

This document provides a comprehensive reference for all data types, entities, and DTOs (Data Transfer Objects) used throughout the TriggerX Backend. Understanding these structures is essential for working with the codebase, API integration, and database schema.

---

## Table of Contents

1. [Data Type Hierarchy](#data-type-hierarchy)
2. [Database Entities](#database-entities)
3. [Data Transfer Objects (DTOs)](#data-transfer-objects-dtos)
4. [HTTP Request/Response Types](#http-requestresponse-types)
5. [RPC Message Types](#rpc-message-types)
6. [Constants and Enums](#constants-and-enums)
7. [Monetary Calculations (Wei-Based)](#monetary-calculations-wei-based)

---

## Data Type Hierarchy

TriggerX uses a clear separation between different data representations:

```bash
┌─────────────────────────────────────────────────────────────────┐
│                      Data Type Layers                           │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  Database Entities (pkg/types/db_entity.go)                     │
│  ├── Direct mapping to ScyllaDB tables                          │
│  ├── CQL struct tags (e.g., `cql:"user_address"`)               │
│  └── Used by database layer only                                │
│                            ↕                                    │
│            Conversion Layer (pkg/types/db_converters.go)        │
│                            ↕                                    │
│  Data Transfer Objects - DTOs (pkg/types/db_dto.go)             │
│  ├── JSON struct tags (e.g., `json:"user_address"`)             │
│  ├── Exposed via HTTP/REST APIs                                 │
│  └── Used by services and frontend                              │
│                            ↕                                    │
│  HTTP Request/Response Types (pkg/types/http_*.go)              │
│  ├── API-specific request/response wrappers                     │
│  └── Include validation rules                                   │
│                            ↕                                    │
│  RPC Message Types (pkg/types/rpc_*.go)                         │
│  ├── gRPC / Protocol Buffer equivalents                         │
│  └── Used for internal service communication                    │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

**Important**: The source of truth is `pkg/types/db_entity.go` and `pkg/types/db_dto.go`. If any other data structure uses different names or data types, they should be corrected to match these files.

---

## Database Entities

All entities map directly to ScyllaDB tables. They use `cql` struct tags for column mapping.

**Location**: `pkg/types/db_entity.go`

### 1. ApiKeyDataEntity

Represents `apikeys` table for API key management.

```go
type ApiKeyDataEntity struct {
    Key          string    `cql:"key"`           // API key (primary key, UUID)
    Owner        string    `cql:"owner"`         // Wallet address of owner
    IsActive     bool      `cql:"is_active"`     // Active or revoked
    RateLimit    int       `cql:"rate_limit"`    // Requests per minute
    SuccessCount int64     `cql:"success_count"` // Successful requests
    FailedCount  int64     `cql:"failed_count"`  // Failed requests
    LastUsed     time.Time `cql:"last_used"`     // Last request time
    CreatedAt    time.Time `cql:"created_at"`    // Key creation time
}
```

**Primary Key**: `key`

**Field Notes**:

- `RateLimit`: Enforced by API gateway middleware
- `SuccessCount` and `FailedCount`: Used for analytics and abuse detection

### 2. UserDataEntity

Represents the `user_data` table for user profiles.

```go
type UserDataEntity struct {
    UserAddress   string    `cql:"user_address"`   // Wallet address (primary key)
    EmailID       string    `cql:"email_id"`       // Email for notifications
    JobIDs        []string  `cql:"job_ids"`        // List of job IDs owned by user
    UserPoints    string    `cql:"user_points"`    // Total points (Wei-based, sum of JobCostActual)
    TotalJobs     int64     `cql:"total_jobs"`     // Total jobs created
    TotalTasks    int64     `cql:"total_tasks"`    // Total tasks executed
    CreatedAt     time.Time `cql:"created_at"`     // Account creation timestamp
    LastUpdatedAt time.Time `cql:"last_updated_at"`// Last profile update
}
```

**Primary Key**: `user_address`

**Field Notes**:

- `UserPoints`: Stored as string to represent Wei (big integer)
- `JobIDs`: Array of job IDs (`set<text>` in CQL, []string in Go) for quick lookup of user's jobs. Stored as string because varint can exceed int64 limits.

### 3. SafeAddressDataEntity

Represents `safe_addresses` table for Safe address management.

```go
type SafeAddressDataEntity struct {
    UserAddress string    `cql:"user_address"` // Wallet address of owner
    SafeAddress string    `cql:"safe_address"` // Safe address
    SafeName    string    `cql:"safe_name"`    // Safe name
    CreatedAt   time.Time `cql:"created_at"`   // Creation time
}
```

**Primary Key**: `user_address, safe_address`

### 4. JobDataEntity

Represents the `job_data` table for all job types.

```go
type JobDataEntity struct {
    JobID             string    `cql:"job_id"`              // Unique job identifier (same as Job Registry)
    JobTitle          string    `cql:"job_title"`           // User-defined job name
    TaskDefinitionID  int       `cql:"task_definition_id"`  // Maps to task type (1-9)
    CreatedChainID    string    `cql:"created_chain_id"`    // Chain where job was created
    UserAddress       string    `cql:"user_address"`        // Job owner
    LinkJobID         string    `cql:"link_job_id"`         // Linked job ID (for chaining)
    ChainStatus       int       `cql:"chain_status"`        // 0=None, 1=Chain Head, 2=Chain Block
    SafeAddress       string    `cql:"safe_address"`        // Safe address
    Timezone          string    `cql:"timezone"`            // User's timezone (e.g., "America/New_York")
    IsImua            bool      `cql:"is_imua"`             // Special IMUA job type
    JobType           string    `cql:"job_type"`            // "frontend", "sdk", "template", "contract"
    TimeFrame         int64     `cql:"time_frame"`          // Job validity duration (seconds)
    Recurring         bool      `cql:"recurring"`           // Recurring job or one-time
    Status            string    `cql:"status"`              // "created", "running", "completed", "failed", "expired", "deleted"
    JobCostPrediction string    `cql:"job_cost_prediction"` // Estimated cost (Wei, for a single task)
    JobCostActual     string    `cql:"job_cost_actual"`     // Actual cost (Wei, sum of task costs)
    TaskIDs           []int64   `cql:"task_ids"`            // Tasks executed for this job
    CreatedAt         time.Time `cql:"created_at"`          // Job creation time
    UpdatedAt         time.Time `cql:"updated_at"`          // Last update time
    LastExecutedAt    time.Time `cql:"last_executed_at"`    // Last task execution time
}
```

**Primary Key**: `job_id`

**Indexes**:

- `status` - For filtering jobs by status (created, running, completed, failed, expired, deleted)
- `created_at` - For querying jobs by creation time
- `updated_at` - For querying jobs by last update time
- `last_executed_at` - For querying jobs by last execution time
- `safe_address` - For querying all jobs associated with a Safe address

**Field Notes**:

- `JobID` and `LinkJobID`: Stored as varint in CQL, mapped to `string` in Go because varint can exceed Go's `int64` maximum value. This allows for arbitrarily large job identifiers.
- `JobCostPrediction` and `JobCostActual`: Wei-based monetary values (string for arbitrary-precision integer)
- `ChainStatus`: Used for job chaining (execute job B after job A completes)
- `TaskDefinitionID`: References the task type (1-6 for traditional jobs, 7-9 for agent jobs, see constants)
- `TaskIDs`: Set of task IDs (bigint in CQL, int64 in Go) representing all tasks executed for this job

### 5. TimeJobDataEntity

Represents `time_job_data` table for time-based scheduled jobs. Supports both traditional jobs (TDI 1, 2) and agent jobs (TDI 7).

```go
type TimeJobDataEntity struct {
    JobID                     string    `cql:"job_id"`                      // Unique job identifier (varint in CQL, string in Go)
    TaskDefinitionID          int       `cql:"task_definition_id"`          // Task type (1, 2, or 7)
    ScheduleType              string    `cql:"schedule_type"`               // "cron", "interval", "specific"
    TimeInterval              int64     `cql:"time_interval"`               // Interval in seconds (for interval type)
    CronExpression            string    `cql:"cron_expression"`             // Cron expression (for cron type)
    SpecificSchedule          string    `cql:"specific_schedule"`           // Specific timestamp (for specific type)
    Timezone                  string    `cql:"timezone"`                    // Timezone selected by user
    NextExecutionTimestamp    time.Time `cql:"next_execution_timestamp"`    // Calculated next run time

    // Traditional job fields (TDI 1, 2) - nullable for agent jobs (TDI 7)
    TargetChainID             string    `cql:"target_chain_id"`             // Chain to execute on
    TargetContractAddress     string    `cql:"target_contract_address"`     // Contract to call
    TargetFunction            string    `cql:"target_function"`             // Function to invoke
    ABI                       string    `cql:"abi"`                         // Contract ABI (JSON)
    ArgType                   int       `cql:"arg_type"`                    // 0=None, 1=Static, 2=Dynamic
    Arguments                 []string  `cql:"arguments"`                   // Static arguments (if ArgType=1, list<text> in CQL)
    DynamicArgumentsScriptURL string    `cql:"dynamic_arguments_script_url"`// IPFS CID for dynamic arg script

    // Agent job fields (TDI 7) - nullable for traditional jobs (TDI 1, 2)
    AgentScriptURL            string    `cql:"agent_script_url"`            // IPFS URL of the agent script
    AgentScriptLanguage       string    `cql:"agent_script_language"`       // Script language: 'ts', 'go', 'python', 'javascript'
    AgentScriptHash           string    `cql:"agent_script_hash"`           // keccak256(scriptCode) for verification
    AgentTargetChainID        int       `cql:"agent_target_chain_id"`       // Default chain ID (script can override)
    MaxExecutionTime          int       `cql:"max_execution_time"`          // Script timeout in seconds (default 60)
    ChallengePeriod           int64     `cql:"challenge_period"`            // Challenge period in seconds (default 21600 = 6 hours)

    // Common status fields
    IsActive       bool      `cql:"is_active"`        // Job is active (true on creation, false when expiration time is reached)
    LastExecutedAt time.Time `cql:"last_executed_at"` // Last execution time
    ExpirationTime time.Time `cql:"expiration_time"`  // Job expiration
}
```

**Primary Key**: `job_id`

**Indexes**:

- `last_executed_at` - For querying jobs by last execution time
- `next_execution_timestamp` - **Critical**: Used by time scheduler to find jobs due for execution (WHERE next_execution_timestamp >= ? AND next_execution_timestamp <= ?)

**Field Notes**:

- **Traditional Jobs (TDI 1, 2)**: Traditional target fields (`target_chain_id`, `target_contract_address`, `target_function`, `abi`, `arg_type`, `arguments`, `dynamic_arguments_script_url`) are populated. Agent fields are NULL.
- **Agent Jobs (TDI 7)**: Agent fields (`agent_script_url`, `agent_script_language`, `agent_script_hash`, `agent_target_chain_id`, `max_execution_time`, `challenge_period`) are populated. Traditional target fields are NULL.
- `TaskDefinitionID`: Determines execution model - 1/2 for traditional, 7 for agent
- `Arguments`: List of text in CQL, converted to []string in Go

### 6. EventJobDataEntity

Represents `event_job_data` table for blockchain event-triggered jobs. Supports both traditional jobs (TDI 3, 4) and agent jobs (TDI 8).

```go
type EventJobDataEntity struct {
    JobID                      string    `cql:"job_id"`                       // Unique job identifier (varint in CQL, string in Go)
    TaskDefinitionID           int       `cql:"task_definition_id"`           // Task type (3, 4, or 8)
    Recurring                  bool      `cql:"recurring"`                    // Trigger on every event in time frame or once

    // Event trigger fields (common for all)
    TriggerChainID             string    `cql:"trigger_chain_id"`             // Chain to monitor
    TriggerContractAddress     string    `cql:"trigger_contract_address"`     // Contract to watch
    TriggerEvent               string    `cql:"trigger_event"`                // Event signature (e.g., "Transfer(address,address,uint256)")
    EventFilterParaName        string    `cql:"event_filter_para_name"`       // Indexed parameter to filter (e.g., "to")
    EventFilterValue           string    `cql:"event_filter_value"`           // Filter value (e.g., specific address)

    // Traditional job fields (TDI 3, 4) - nullable for agent jobs (TDI 8)
    TargetChainID              string    `cql:"target_chain_id"`              // Chain to execute on
    TargetContractAddress      string    `cql:"target_contract_address"`      // Contract to call
    TargetFunction             string    `cql:"target_function"`              // Function to invoke
    ABI                        string    `cql:"abi"`                          // Contract ABI (JSON)
    ArgType                    int       `cql:"arg_type"`                     // 0=None, 1=Static, 2=Dynamic
    Arguments                  []string  `cql:"arguments"`                    // Static arguments (if ArgType=1, list<text> in CQL)
    DynamicArgumentsScriptURL  string    `cql:"dynamic_arguments_script_url"` // IPFS CID for dynamic arg script

    // Agent job fields (TDI 8) - nullable for traditional jobs (TDI 3, 4)
    AgentScriptURL             string    `cql:"agent_script_url"`             // IPFS URL of the agent script
    AgentScriptLanguage        string    `cql:"agent_script_language"`        // Script language: 'ts', 'go', 'python', 'javascript'
    AgentScriptHash            string    `cql:"agent_script_hash"`            // keccak256(scriptCode) for verification
    AgentTargetChainID         int       `cql:"agent_target_chain_id"`        // Default chain ID (script can override)
    MaxExecutionTime           int       `cql:"max_execution_time"`           // Script timeout in seconds (default 60)
    ChallengePeriod            int64     `cql:"challenge_period"`             // Challenge period in seconds (default 21600 = 6 hours)

    // Common status fields
    IsActive       bool      `cql:"is_active"`        // Job is active (true on creation, false when expiration time is reached)
    LastExecutedAt time.Time `cql:"last_executed_at"` // Last execution time
    ExpirationTime time.Time `cql:"expiration_time"`  // Job expiration
}
```

**Primary Key**: `job_id`

**Indexes**:

- `last_executed_at` - For querying jobs by last execution time

**Field Notes**:

- **Traditional Jobs (TDI 3, 4)**: Traditional target fields are populated. Agent fields are NULL.
- **Agent Jobs (TDI 8)**: Agent fields are populated. Traditional target fields are NULL.
- `TriggerEvent`: Event signature in Solidity format
- `EventFilterParaName` and `EventFilterValue`: Filter events by indexed parameters (allows monitoring specific event instances)
- `Recurring`: If false, job is executed once when triggered and marked completed
- `TaskDefinitionID`: Determines execution model - 3/4 for traditional, 8 for agent

### 7. ConditionJobDataEntity

Represents `condition_job_data` table for condition-based jobs. Supports both traditional jobs (TDI 5, 6) and agent jobs (TDI 9).

```go
type ConditionJobDataEntity struct {
    JobID                     string    `cql:"job_id"`                      // Unique job identifier (varint in CQL, string in Go)
    TaskDefinitionID          int       `cql:"task_definition_id"`          // Task type (5, 6, or 9)
    Recurring                 bool      `cql:"recurring"`                   // Trigger on every condition match or once

    // Condition trigger fields (common for all)
    ConditionType             string    `cql:"condition_type"`              // "balance", "state", "oracle"
    UpperLimit                float64   `cql:"upper_limit"`                 // Upper threshold (double in CQL)
    LowerLimit                float64   `cql:"lower_limit"`                 // Lower threshold (double in CQL)
    ValueSourceType           string    `cql:"value_source_type"`           // "api", "oracle", "websocket"
    ValueSourceURL            string    `cql:"value_source_url"`            // API URL or Websocket URL
    SelectedKeyRoute          string    `cql:"selected_key_route"`          // JSON path for API responses

    // Traditional job fields (TDI 5, 6) - nullable for agent jobs (TDI 9)
    TargetChainID             string    `cql:"target_chain_id"`             // Chain to execute on
    TargetContractAddress     string    `cql:"target_contract_address"`     // Contract to call
    TargetFunction            string    `cql:"target_function"`             // Function to invoke
    ABI                       string    `cql:"abi"`                         // Contract ABI (JSON)
    ArgType                   int       `cql:"arg_type"`                    // 0=None, 1=Static, 2=Dynamic
    Arguments                 []string  `cql:"arguments"`                   // Static arguments (if ArgType=1, list<text> in CQL)
    DynamicArgumentsScriptURL string    `cql:"dynamic_arguments_script_url"`// IPFS CID for dynamic arg script

    // Agent job fields (TDI 9) - nullable for traditional jobs (TDI 5, 6)
    AgentScriptURL            string    `cql:"agent_script_url"`            // IPFS URL of the agent script
    AgentScriptLanguage       string    `cql:"agent_script_language"`       // Script language: 'ts', 'go', 'python', 'javascript'
    AgentScriptHash           string    `cql:"agent_script_hash"`           // keccak256(scriptCode) for verification
    AgentTargetChainID        int       `cql:"agent_target_chain_id"`       // Default chain ID (script can override)
    MaxExecutionTime          int       `cql:"max_execution_time"`          // Script timeout in seconds (default 60)
    ChallengePeriod           int64     `cql:"challenge_period"`            // Challenge period in seconds (default 21600 = 6 hours)

    // Common status fields
    IsActive       bool      `cql:"is_active"`        // Job is active (true on creation, false when expiration time is reached)
    LastExecutedAt time.Time `cql:"last_executed_at"` // Last execution time
    ExpirationTime time.Time `cql:"expiration_time"`  // Job expiration
}
```

**Primary Key**: `job_id`

**Indexes**:

- `last_executed_at` - For querying jobs by last execution time

**Field Notes**:

- **Traditional Jobs (TDI 5, 6)**: Traditional target fields are populated. Agent fields are NULL.
- **Agent Jobs (TDI 9)**: Agent fields are populated. Traditional target fields are NULL.
- `ConditionType`: Type of condition to monitor (e.g., "balance", "state", "oracle")
- `UpperLimit` and `LowerLimit`: Trigger when value crosses thresholds. Only upper limit is used for single value conditions.
- `SelectedKeyRoute`: JSON path (e.g., `data.price.usd`) for extracting value from API response
- `TaskDefinitionID`: Determines execution model - 5/6 for traditional, 9 for agent

### 8. TaskDataEntity

Represents `task_data` table for individual task executions.

```go
type TaskDataEntity struct {
    TaskID               int64     `cql:"task_id"`                // Unique task ID (bigint in CQL, auto-increment)
    TaskNumber           int64     `cql:"task_number"`            // Task sequence number on the contract
    TaskStatus           string    `cql:"task_status"`            // "pending", "in-queue", "running", "completed", "failed", "expired", "deleted"
    TaskError            string    `cql:"task_error"`             // Error message if task failed
    JobID                string    `cql:"job_id"`                 // Reference to job (varint in CQL, string in Go)
    TaskDefinitionID     int       `cql:"task_definition_id"`     // Task definition ID (1-9)
    CreatedAt            time.Time `cql:"created_at"`             // Task creation timestamp
    ExecutedAt           time.Time `cql:"executed_at"`            // Execution timestamp
    SubmittedAt          time.Time `cql:"submitted_at"`           // Submission timestamp
    TaskOpxPredictedCost string    `cql:"task_opx_predicted_cost"`// From JobCostPrediction (Wei)
    TaskOpxActualCost    string    `cql:"task_opx_actual_cost"`   // Actual cost including tx gas (Wei)
    ExecutionTxHash      string    `cql:"execution_tx_hash"`      // Transaction hash for action
    SubmissionTxHash     string    `cql:"submission_tx_hash"`     // Aggregator submission tx hash
    ConvertedArguments   []string  `cql:"converted_arguments"`    // Arguments used (JSON, list<text> in CQL)
    TaskPerformerAddress []string  `cql:"task_performer_address"` // Keeper who executed (list<text> in CQL, multiple if first keeper failed)
    TaskAttesterAddress  []string  `cql:"task_attester_address"`  // Keepers who attested (list<text> in CQL)
    ProofOfTask          string    `cql:"proof_of_task"`          // Cryptographic proof (IPFS CID)
    IsSuccessful         bool      `cql:"is_successful"`          // Execution success
    IsAccepted           bool      `cql:"is_accepted"`            // Consensus accepted
    IsImua               bool      `cql:"is_imua"`                // IMUA task type
}
```

**Primary Key**: `task_id`

**Indexes**:

- `job_id` - For querying all tasks for a specific job (WHERE job_id = ? ALLOW FILTERING)

**Field Notes**:

- `JobID`: References the parent job (varint in CQL, string in Go to support values exceeding int64 limits)
- `TaskDefinitionID`: Can be 1-9, indicating both traditional (1-6) and agent (7-9) task types
- `TaskStatus`: Can be "pending", "in-queue", "running", "completed", "failed", "expired", "deleted" (as documented in entity comment). The constants file defines: "created", "dispatched", "executed", "completed", "failed" for task lifecycle management.
- `ConvertedArguments`, `TaskPerformerAddress`, `TaskAttesterAddress`: Lists of text in CQL, converted to []string in Go

### 9. KeeperDataEntity

Represents `keeper_data` table for Keeper node registry.

```go
type KeeperDataEntity struct {
    KeeperAddress     string      `cql:"keeper_address"`     // Wallet address (primary key)
    KeeperName        string      `cql:"keeper_name"`        // Human-readable keeper name
    RewardsAddress    string      `cql:"rewards_address"`    // Address for reward payouts
    ConsensusAddress  string      `cql:"consensus_address"`  // Consensus layer address
    OperatorID        int          `cql:"operator_id"`        // Unique operator ID on the contract (int in CQL)
    VotingPower       string      `cql:"voting_power"`       // Voting power (Wei-based stake)
    Registered        bool        `cql:"registered"`         // Registered on-chain
    RegisteredAt      []time.Time `cql:"registered_at"`      // List of registration timestamps (list<timestamp> in CQL)
    Online            bool        `cql:"online"`             // Currently online
    Version           string      `cql:"version"`            // Keeper software version
    Network           string      `cql:"network"`            // Network the keeper is running on
    ConnectionAddress string      `cql:"connection_address"` // Connection address for P2P
    PeerID            string      `cql:"peer_id"`            // Peer ID for P2P
    Uptime            int64       `cql:"uptime"`             // Total uptime (seconds, bigint in CQL)
    LastCheckedIn     time.Time   `cql:"last_checked_in"`    // Last heartbeat time
    RewardsBooster    string      `cql:"rewards_booster"`    // Reward multiplier (for testnets, obsolete for mainnet, text in CQL)
    NoExecutedTasks   int64       `cql:"no_executed_tasks"`  // Total tasks executed (bigint in CQL)
    NoAttestedTasks   int64       `cql:"no_attested_tasks"`  // Total tasks attested (bigint in CQL)
    KeeperPoints      string      `cql:"keeper_points"`      // Total points (Wei, sum of TaskOpxCost executed/attested, daily rewards)
    ChatID            int64       `cql:"chat_id"`            // Telegram chat ID for alerts (int64 in CQL)
    EmailID           string      `cql:"email_id"`           // Email for alerts
}
```

**Primary Key**: `keeper_address`

**Field Notes**:

- `OperatorID`: Stored as `int` in CQL (32-bit), matching `int` in Go
- `RegisteredAt`: List of timestamps in CQL, allowing tracking of multiple registration events
- `RewardsBooster`: Stored as `text` in CQL, `string` in Go (for testnets, obsolete for mainnet)

### 10. AgentScriptExecutionsEntity

Represents `agent_script_executions` table for tracking agent script executions (TDI 7, 8, 9) with fraud-proof data.

```go
type AgentScriptExecutionsEntity struct {
    ExecutionID        string    `cql:"execution_id"`         // Unique execution identifier (UUID, primary key)
    JobID              string    `cql:"job_id"`               // Reference to job (varint in CQL, string in Go)
    TaskID             int64     `cql:"task_id"`              // Link to task_data table (bigint in CQL)
    TaskDefinitionID   int       `cql:"task_definition_id"`   // 7, 8, or 9
    ScheduledTime      time.Time `cql:"scheduled_time"`       // When it was supposed to execute
    ActualTime         time.Time `cql:"actual_time"`          // When it actually executed
    PerformerAddress   string    `cql:"performer_address"`    // Keeper who executed the script

    // Input data (for deterministic re-execution)
    InputTimestamp     int64     `cql:"input_timestamp"`      // Timestamp used as input (bigint in CQL)
    InputStorage       string    `cql:"input_storage"`        // JSON snapshot of storage at execution time
    InputHash          string    `cql:"input_hash"`           // Hash of inputs for verification

    // Trigger-specific input data (JSON)
    TriggerData        string    `cql:"trigger_data"`         // TDI 7: {scheduled_time}
                                                               // TDI 8: {event_data, block_number, tx_hash, event_params}
                                                               // TDI 9: {condition_value, timestamp, condition_type}

    // Output data
    ShouldExecute      bool      `cql:"should_execute"`       // Whether to submit on-chain transaction
    TargetContract     string    `cql:"target_contract"`      // Contract address to call
    Calldata           string    `cql:"calldata"`             // Encoded function call (hex string)
    OutputHash         string    `cql:"output_hash"`          // Hash of outputs for verification

    // Metadata (CRITICAL: API calls, contract calls, block numbers)
    ExecutionMetadata  string    `cql:"execution_metadata"`   // JSON containing:
                                                               // - API calls made: [{url, blockNumber, response}, ...]
                                                               // - Contract calls: [{contract, blockNumber, function, response}, ...]
                                                               // - Oracle calls: [{oracle, blockNumber, data}, ...]

    // Proof
    ScriptHash         string    `cql:"script_hash"`          // Hash of the script code
    Signature          string    `cql:"signature"`            // Performer's signature

    // Execution result
    TxHash             string    `cql:"tx_hash"`              // Transaction hash if executed
    ExecutionStatus    string    `cql:"execution_status"`     // 'success', 'failed', 'no_execution'
    ExecutionError     string    `cql:"execution_error"`      // Error message if failed

    // Verification status
    VerificationStatus string    `cql:"verification_status"`  // 'pending', 'verified', 'challenged', 'slashed'
    ChallengeDeadline  time.Time `cql:"challenge_deadline"`   // Deadline for challenges
    IsChallenged       bool      `cql:"is_challenged"`        // Whether execution has been challenged
    ChallengeCount     int       `cql:"challenge_count"`      // Number of challenges raised
    CreatedAt          time.Time `cql:"created_at"`           // Record creation timestamp
}
```

**Primary Key**: `execution_id`

**Indexes**:

- `job_id` - For querying all executions for a job
- `task_id` - For linking to task_data
- `verification_status` - For finding pending/verified/challenged executions
- `performer_address` - For querying by keeper
- `task_definition_id` - For filtering by task type

**Field Notes**:

- **Purpose**: Tracks every agent script execution with complete fraud-proof data for verification and challenge resolution
- **Trigger Data Format**: Varies by TDI:
  - **TDI 7 (Time)**: `{"scheduledTime": "2026-01-14T10:00:00Z", "storage": {...}, "jobId": "123"}`
  - **TDI 8 (Event)**: `{"eventData": {...}, "blockNumber": 12345678, "txHash": "0x...", "eventSignature": "...", "storage": {...}, "jobId": "124"}`
  - **TDI 9 (Condition)**: `{"conditionValue": 2050.0, "conditionType": "price", "timestamp": "...", "storage": {...}, "jobId": "125"}`
- **Execution Metadata**: Critical for fraud-proof verification - contains all external API calls, contract calls, and oracle queries with block numbers for deterministic re-execution
- **Challenge Period**: Default 21600 seconds (6 hours) during which any validator can challenge incorrect executions
- **Verification Flow**: Execution starts as 'pending', can transition to 'verified', 'challenged', or 'slashed' based on consensus

### 11. ScriptStorageEntity

Represents `script_storage` table for persistent key-value storage per agent job. Used by TDI 7, 8, 9 for maintaining state across executions.

```go
type ScriptStorageEntity struct {
    JobID       string    `cql:"job_id"`       // Job identifier (varint in CQL, string in Go, part of composite primary key)
    StorageKey  string    `cql:"storage_key"`  // Storage key (text, part of composite primary key)
    StorageValue string   `cql:"storage_value"`// Storage value (text, JSON or string)
    UpdatedAt   time.Time `cql:"updated_at"`   // Last update timestamp
}
```

**Primary Key**: `(job_id, storage_key)`

**Field Notes**:

- **Purpose**: Provides persistent state storage for agent scripts across multiple executions
- **Usage**: Agent scripts can store and retrieve key-value pairs to maintain state between runs
- **Example**: A price monitoring script might store `{"lastPrice": "1850.50", "counter": 42}` as separate key-value pairs
- **Isolation**: Each job has its own isolated storage namespace

### 12. ExecutionChallengesEntity

Represents `execution_challenges` table for fraud-proof challenges against agent script executions.

```go
type ExecutionChallengesEntity struct {
    ChallengeID            string    `cql:"challenge_id"`              // Unique challenge identifier (UUID, primary key)
    ExecutionID            string    `cql:"execution_id"`              // Reference to agent_script_executions
    ChallengerAddress      string    `cql:"challenger_address"`        // Address of the challenger
    ChallengeReason        string    `cql:"challenge_reason"`          // 'wrong_output', 'missing_execution', 'invalid_calldata', 'wrong_trigger'

    // Challenger's claimed output
    ChallengerOutputHash   string    `cql:"challenger_output_hash"`    // Hash of challenger's claimed output
    ChallengerShouldExecute bool     `cql:"challenger_should_execute"` // Challenger's claim: should execute?
    ChallengerTargetContract string  `cql:"challenger_target_contract"`// Challenger's claimed target contract
    ChallengerCalldata     string    `cql:"challenger_calldata"`       // Challenger's claimed calldata
    ChallengerSignature    string    `cql:"challenger_signature"`      // Challenger's signature

    // Challenge bond (prevents spam)
    BondAmount             string    `cql:"bond_amount"`               // Wei-based bond amount

    // Resolution
    ResolutionStatus       string    `cql:"resolution_status"`         // 'pending', 'approved', 'rejected', 'inconclusive'
    ResolutionTime         time.Time `cql:"resolution_time"`           // When challenge was resolved
    ValidatorCount         int       `cql:"validator_count"`           // Number of validators who voted
    ApproveCount           int       `cql:"approve_count"`             // Number of approving votes
    RejectCount            int       `cql:"reject_count"`              // Number of rejecting votes
    CreatedAt              time.Time `cql:"created_at"`                // Challenge creation timestamp
}
```

**Primary Key**: `challenge_id`

**Indexes**:

- `execution_id` - For finding all challenges for an execution
- `resolution_status` - For filtering by resolution state
- `challenger_address` - For querying challenges by challenger

**Field Notes**:

- **Purpose**: Tracks fraud-proof challenges where validators dispute incorrect agent script executions
- **Challenge Reasons**:
  - `wrong_output`: Output hash doesn't match expected result
  - `missing_execution`: Execution should have happened but didn't
  - `invalid_calldata`: Calldata is malformed or invalid
  - `wrong_trigger`: Trigger condition was incorrectly evaluated
- **Bond System**: Challengers must deposit a bond to prevent spam attacks. Bond is slashed if challenge is rejected.
- **Resolution**: Validators vote on challenge validity. If approved, performer may be slashed. If rejected, challenger loses bond.
- **Security**: Enables trustless verification of agent script executions without requiring all validators to execute scripts

---

## Agent Jobs Architecture

### Task Definition ID Mapping

The system supports 9 task definition IDs, grouped by trigger type and execution model:

**Time-Based Jobs**:

- **TDI 1** (Time Static): `job_data` + `time_job_data` (traditional fields populated)
- **TDI 2** (Time Dynamic): `job_data` + `time_job_data` (traditional fields populated)
- **TDI 7** (Agent Time): `job_data` + `time_job_data` (agent fields populated) + `agent_script_executions` + `script_storage`

**Event-Based Jobs**:

- **TDI 3** (Event Static): `job_data` + `event_job_data` (traditional fields populated)
- **TDI 4** (Event Dynamic): `job_data` + `event_job_data` (traditional fields populated)
- **TDI 8** (Agent Event): `job_data` + `event_job_data` (agent fields populated) + `agent_script_executions` + `script_storage`

**Condition-Based Jobs**:

- **TDI 5** (Condition Static): `job_data` + `condition_job_data` (traditional fields populated)
- **TDI 6** (Condition Dynamic): `job_data` + `condition_job_data` (traditional fields populated)
- **TDI 9** (Agent Condition): `job_data` + `condition_job_data` (agent fields populated) + `agent_script_executions` + `script_storage`

### Traditional Jobs (TDI 1-6) Architecture

- `job_data`: Common metadata (title, user, status, costs)
- `{time|event|condition}_job_data`: Scheduling + Target execution (target_contract_address, target_function, abi, arguments)
- Agent fields (agent_script_url, agent_script_language, etc.) are NULL
- Keeper receives task and executes directly on target contract
- Script (if arg_type=2) only generates arguments, not full calldata

### Agent Jobs (TDI 7-9) Architecture

- `job_data`: Common metadata
- `{time|event|condition}_job_data`: Scheduling + Agent script details (agent_script_url, agent_script_language, etc.)
- Traditional target fields (target_contract_address, target_function, abi, arguments) are NULL
- Keeper receives task, runs agent script in Docker container
- Script decides dynamically:
  1. `shouldExecute` (bool) - whether to submit on-chain transaction
  2. `targetContract` (address) - which contract to call
  3. `calldata` (bytes) - full encoded function call
  4. `storageUpdates` (map) - persistent state updates
- Execution recorded in `agent_script_executions` with fraud-proof data
- Challenge period (default 6 hours) allows validators to verify correct execution

### Execution Flow Comparison

**Traditional (TDI 1-6)**:

```bash
Scheduler → Creates Task → Keeper → Executes Target Contract
```

**Agent (TDI 7-9)**:

```bash
Scheduler → Creates Task → Keeper → Runs Script → Script Decides
                                              ↓
                                 shouldExecute = true/false
                                              ↓
                                 Submit Tx or Skip (both logged)
                                              ↓
                                 Challenge Period (6 hours default)
```

### Agent Script Input/Output

**Input Format** (varies by TDI):

- **TDI 7 (Time)**: `{"scheduledTime": "...", "storage": {...}, "jobId": "..."}`
- **TDI 8 (Event)**: `{"eventData": {...}, "blockNumber": ..., "txHash": "...", "eventSignature": "...", "storage": {...}, "jobId": "..."}`
- **TDI 9 (Condition)**: `{"conditionValue": ..., "conditionType": "...", "timestamp": "...", "storage": {...}, "jobId": "..."}`

**Output Format**:

```json
{
  "shouldExecute": true,
  "targetContract": "0xDeFiProtocol...",
  "calldata": "0x...",
  "storageUpdates": { "lastPrice": "2050.0", "counter": 43 }
}
```

---

## Data Transfer Objects (DTOs)

DTOs mirror the database entities but use `json` tags for API serialization. They are the external representation of data.

**Location**: `pkg/types/db_dto.go`

All DTOs have identical field names and types to their corresponding entities, but use JSON struct tags:

```go
type UserDataDTO struct {
    UserAddress   string    `json:"user_address"`
    EmailID       string    `json:"email_id"`
    JobIDs        []string  `json:"job_ids"`
    UserPoints    string    `json:"user_points"`
    TotalJobs     int64     `json:"total_jobs"`
    TotalTasks    int64     `json:"total_tasks"`
    CreatedAt     time.Time `json:"created_at"`
    LastUpdatedAt time.Time `json:"last_updated_at"`
}
```

### Complete Job Data DTO

Special composite DTO that combines job data with job-type-specific data:

```go
type CompleteJobDataDTO struct {
    JobDataDTO           JobDataDTO           `json:"job_data"`
    TimeJobDataDTO       *TimeJobDataDTO      `json:"time_job_data,omitempty"`
    EventJobDataDTO      *EventJobDataDTO     `json:"event_job_data,omitempty"`
    ConditionJobDataDTO  *ConditionJobDataDTO `json:"condition_job_data,omitempty"`
}
```

**Usage**: API returns this for full job details. Only one of the job-type-specific fields is populated based on `JobType`.

### Conversion Between Entities and DTOs

**Location**: `pkg/types/db_converters.go`

Example converter functions:

```go
func UserEntityToDTO(entity *UserDataEntity) *UserDataDTO
func UserDTOToEntity(dto *UserDataDTO) *UserDataEntity
```

---

## HTTP Request/Response Types

API-specific types for request validation and response formatting.

**Location**: `pkg/types/http_*.go`

### Common Types

```go
type HealthCheckResponse struct {
    Status    string    `json:"status"`    // "ok", "degraded", "error"
    Timestamp time.Time `json:"timestamp"` // Current time
    Service   string    `json:"service"`   // Service name
    Version   string    `json:"version"`   // Software version
    Error     string    `json:"error,omitempty"` // Error message if unhealthy
}

// Generic error response
type ErrorResponse struct {
    Error   string `json:"error"`
    Code    string `json:"code"`
    Message string `json:"message"`
    TraceID string `json:"trace_id,omitempty"`
}
```

---

## RPC Message Types

Used for gRPC communication between internal services.

**Location**: `pkg/types/rpc_*.go`

---

## Constants and Enums

**Location**: `pkg/types/constants.go`

### Task Definition IDs

```go
const (
    // Traditional Jobs (1-6)
    TaskDefTimeBasedStatic        = 1  // Time based job with static arguments
    TaskDefTimeBasedDynamic       = 2  // Time based job with dynamic arguments
    TaskDefEventBasedStatic       = 3  // Event based job with static arguments
    TaskDefEventBasedDynamic      = 4  // Event based job with dynamic arguments
    TaskDefConditionBasedStatic   = 5  // Condition based job with static arguments
    TaskDefConditionBasedDynamic  = 6  // Condition based job with dynamic arguments

    // Agent Jobs (7-9)
    TaskDefTimeBasedAgent         = 7  // Time based job with agentic script
    TaskDefEventBasedAgent        = 8  // Event based job with agentic script
    TaskDefConditionBasedAgent    = 9  // Condition based job with agentic script
)
```

### Job Types

```go
const (
    JobTypeFrontend = "frontend"  // Scheduled from Web UI
    JobTypeSdk      = "sdk"       // Scheduled from SDK
    JobTypeTemplate = "template"  // Template from frontend
    JobTypeContract = "contract"  // Created from contract call
)
```

### Job Status

```go
const (
    JobStatusRunning   = "running"    // Job created and actively monitored by scheduler
    JobStatusCompleted = "completed"  // Job finished (atleast 1 task completed)
    JobStatusExpired   = "expired"    // Job past expiration time with no tasks completed
    JobStatusFailed    = "failed"     // Job failed permanently (≥1/3rd tasks failed)
    JobStatusDeleted   = "deleted"    // Job deleted by user (was active when deleted)
)
```

**Note**: The `status` field in `JobDataEntity` can also contain "created" (as documented in the entity comment), but this is not defined as a constant. Jobs typically transition directly to "running" status.

### Argument Types

```go
const (
    ArgTypeNone    = 0  // No arguments
    ArgTypeStatic  = 1  // Static arguments provided at job creation
    ArgTypeDynamic = 2  // Dynamic arguments generated at execution time
)
```

### Chain Status (for job chaining)

```go
const (
    ChainStatusNone  = 0  // Not part of a chain
    ChainStatusHead  = 1  // First job in chain
    ChainStatusBlock = 2  // Middle or last job in chain
)
```

---

## Monetary Calculations (Wei-Based)

TriggerX uses **Wei** (the smallest unit of ETH) for all monetary calculations to avoid floating-point precision errors.

### Wei Conversion Units

```bash
1 ETH  = 1,000,000,000,000,000,000 Wei (1e18)
1 TG   = 1,000,000,000,000,000 Wei (1e15)  // TriggerGas (platform unit)
0.001 TG = 1,000,000,000,000 Wei (1e12)    // Static job fee per execution
```

### Wei-Based Fields

All cost-related fields are stored as **strings** to represent arbitrary-precision integers (big.Int in Go):

- `UserDataEntity.UserPoints`
- `JobDataEntity.JobCostPrediction`
- `JobDataEntity.JobCostActual`
- `TaskDataEntity.TaskOpxPredictedCost`
- `TaskDataEntity.TaskOpxActualCost`
- `KeeperDataEntity.VotingPower`
- `KeeperDataEntity.KeeperPoints`
- `ExecutionChallengesEntity.BondAmount`

**Note**: `KeeperDataEntity.RewardsBooster` is stored as `text` in CQL (string in Go, not Wei-based).

### Calculation Examples

**Location**: `pkg/types/bigint_operations.go`

```go
// Parse string to big.Int
func ParseBigInt(s string) (*big.Int, error)

// Mathematical operations
func Add(x, y string) string
func Sub(x, y string) string
func Mul(x, y string) string
func Div(x, y string) string
func IsEqual(x, y string) bool
func IsLess(x, y string) bool
func IsGreater(x, y string) bool
func IsZero(x string) bool
func IsPositive(x string) bool
func IsNegative(x string) bool
```

### Why Wei?

✅ **No Precision Loss**: Integer arithmetic, no floating-point errors  
✅ **Consistency**: Same unit used in smart contracts  
✅ **Deterministic**: Exact calculations, reproducible results  
✅ **Ethereum Compatible**: Direct mapping to blockchain values

---

## Summary

TriggerX's data type system provides:

✅ **Clear Separation**: Entities (DB) ↔ DTOs (API) ↔ RPC Messages (Internal)  
✅ **Type Safety**: Strong typing with Go structs  
✅ **Validation**: Built-in validation rules  
✅ **Wei-Based Precision**: No floating-point errors in monetary calculations  
✅ **Consistency**: Single source of truth in `pkg/types/`  
✅ **Extensibility**: Easy to add new fields or types

**Remember**: `pkg/types/db_entity.go` and `pkg/types/db_dto.go` are the authoritative sources. All other data structures should conform to these definitions.

---

For understanding how data flows through the system using these types, see [data-flow.md](./d_data_flow.md).
