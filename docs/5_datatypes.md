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

### 1. UserDataEntity

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
- `JobIDs`: Array of job IDs for quick lookup of user's jobs

### 2. JobDataEntity

Represents the `job_data` table for all job types.

```go
type JobDataEntity struct {
    JobID             string    `cql:"job_id"`              // Unique job identifier
    JobTitle          string    `cql:"job_title"`           // User-defined job name
    TaskDefinitionID  int       `cql:"task_definition_id"`  // Maps to task type
    CreatedChainID    string    `cql:"created_chain_id"`    // Chain where job was created
    UserAddress       string    `cql:"user_address"`        // Job owner
    LinkJobID         string    `cql:"link_job_id"`         // Linked job ID (for chaining)
    ChainStatus       int       `cql:"chain_status"`        // 0=None, 1=Chain Head, 2=Chain Block
    Timezone          string    `cql:"timezone"`            // User's timezone (e.g., "America/New_York")
    IsImua            bool      `cql:"is_imua"`             // Special IMUA job type
    JobType           string    `cql:"job_type"`            // "time", "event", "condition"
    TimeFrame         int64     `cql:"time_frame"`          // Job validity duration (seconds)
    Recurring         bool      `cql:"recurring"`           // Recurring job or one-time
    Status            string    `cql:"status"`              // "created", "running", "completed", "failed", "expired"
    JobCostPrediction string    `cql:"job_cost_prediction"` // Estimated cost (Wei)
    JobCostActual     string    `cql:"job_cost_actual"`     // Actual cost (Wei, sum of task costs)
    TaskIDs           []int64   `cql:"task_ids"`            // Tasks executed for this job
    CreatedAt         time.Time `cql:"created_at"`          // Job creation time
    UpdatedAt         time.Time `cql:"updated_at"`          // Last update time
    LastExecutedAt    time.Time `cql:"last_executed_at"`    // Last task execution time
}
```

**Primary Key**: `job_id`

**Field Notes**:

- `JobCostPrediction` and `JobCostActual`: Wei-based monetary values (string for arbitrary-precision integer)
- `ChainStatus`: Used for job chaining (execute job B after job A completes)
- `TaskDefinitionID`: References the task type (see constants)

### 3. TimeJobDataEntity

Represents `time_job_data` table for time-based scheduled jobs.

```go
type TimeJobDataEntity struct {
    JobID                     string    `cql:"job_id"`                      // Reference to JobDataEntity
    TaskDefinitionID          int       `cql:"task_definition_id"`          // Task type
    ScheduleType              string    `cql:"schedule_type"`               // "cron", "interval", "specific"
    TimeInterval              int64     `cql:"time_interval"`               // Interval in seconds (for interval type)
    CronExpression            string    `cql:"cron_expression"`             // Cron expression (for cron type)
    SpecificSchedule          string    `cql:"specific_schedule"`           // Specific timestamp (for specific type)
    NextExecutionTimestamp    time.Time `cql:"next_execution_timestamp"`    // Calculated next run time
    TargetChainID             string    `cql:"target_chain_id"`             // Chain to execute on
    TargetContractAddress     string    `cql:"target_contract_address"`     // Contract to call
    TargetFunction            string    `cql:"target_function"`             // Function to invoke
    ABI                       string    `cql:"abi"`                         // Contract ABI (JSON)
    ArgType                   int       `cql:"arg_type"`                    // 0=None, 1=Static, 2=Dynamic
    Arguments                 []string  `cql:"arguments"`                   // Static arguments (if ArgType=1)
    DynamicArgumentsScriptURL string    `cql:"dynamic_arguments_script_url"`// IPFS CID for dynamic arg script
    IsCompleted               bool      `cql:"is_completed"`                // Job finished
    LastExecutedAt            time.Time `cql:"last_executed_at"`            // Last execution time
    ExpirationTime            time.Time `cql:"expiration_time"`             // Job expiration
}
```

**Primary Key**: `job_id`

**Field Notes**:

- `ScheduleType`: Determines which schedule field to use
- `NextExecutionTimestamp`: Updated after each execution for recurring jobs
- `DynamicArgumentsScriptURL`: IPFS CID for script that generates arguments at runtime

### 4. EventJobDataEntity

Represents `event_job_data` table for blockchain event-triggered jobs.

```go
type EventJobDataEntity struct {
    JobID                      string    `cql:"job_id"`                       // Reference to JobDataEntity
    TaskDefinitionID           int       `cql:"task_definition_id"`           // Task type
    Recurring                  bool      `cql:"recurring"`                    // Trigger on every event or once
    TriggerChainID             string    `cql:"trigger_chain_id"`             // Chain to monitor
    TriggerContractAddress     string    `cql:"trigger_contract_address"`     // Contract to watch
    TriggerEvent               string    `cql:"trigger_event"`                // Event signature (e.g., "Transfer(address,address,uint256)")
    TriggerEventFilterParaName string    `cql:"trigger_event_filter_para_name"` // Indexed parameter to filter (e.g., "to")
    TriggerEventFilterValue    string    `cql:"trigger_event_filter_value"`   // Filter value (e.g., specific address)
    TargetChainID              string    `cql:"target_chain_id"`              // Chain to execute action
    TargetContractAddress      string    `cql:"target_contract_address"`      // Action contract
    TargetFunction             string    `cql:"target_function"`              // Action function
    ABI                        string    `cql:"abi"`                          // Contract ABI
    ArgType                    int       `cql:"arg_type"`                     // Argument type
    Arguments                  []string  `cql:"arguments"`                    // Static arguments
    DynamicArgumentsScriptURL  string    `cql:"dynamic_arguments_script_url"` // Dynamic arg script CID
    IsCompleted                bool      `cql:"is_completed"`                 // Job finished
    LastExecutedAt             time.Time `cql:"last_executed_at"`             // Last trigger time
    ExpirationTime             time.Time `cql:"expiration_time"`              // Job expiration
}
```

**Primary Key**: `job_id`

**Field Notes**:

- `TriggerEvent`: Event signature in Solidity format
- `TriggerEventFilterParaName` and `TriggerEventFilterValue`: Filter events by indexed parameters
- `Recurring`: If false, job is executed once and marked completed

### 5. ConditionJobDataEntity

Represents `condition_job_data` table for condition-based jobs.

```go
type ConditionJobDataEntity struct {
    JobID                     string    `cql:"job_id"`                      // Reference to JobDataEntity
    TaskDefinitionID          int       `cql:"task_definition_id"`          // Task type
    Recurring                 bool      `cql:"recurring"`                   // Check continuously or once
    ConditionType             string    `cql:"condition_type"`              // "balance", "state", "oracle"
    UpperLimit                float64   `cql:"upper_limit"`                 // Upper threshold
    LowerLimit                float64   `cql:"lower_limit"`                 // Lower threshold
    ValueSourceType           string    `cql:"value_source_type"`           // "contract", "api", "oracle"
    ValueSourceURL            string    `cql:"value_source_url"`            // RPC endpoint or API URL
    SelectedKeyRoute          string    `cql:"selected_key_route"`          // JSON path for API responses
    TargetChainID             string    `cql:"target_chain_id"`             // Action chain
    TargetContractAddress     string    `cql:"target_contract_address"`     // Action contract
    TargetFunction            string    `cql:"target_function"`             // Action function
    ABI                       string    `cql:"abi"`                         // Contract ABI
    ArgType                   int       `cql:"arg_type"`                    // Argument type
    Arguments                 []string  `cql:"arguments"`                   // Static arguments
    DynamicArgumentsScriptURL string    `cql:"dynamic_arguments_script_url"`// Dynamic arg script CID
    IsCompleted               bool      `cql:"is_completed"`                // Job finished
    LastExecutedAt            time.Time `cql:"last_executed_at"`            // Last check time
    ExpirationTime            time.Time `cql:"expiration_time"`             // Job expiration
}
```

**Primary Key**: `job_id`

**Field Notes**:

- `ConditionType`: Type of condition to monitor
- `UpperLimit` and `LowerLimit`: Trigger when value is outside or inside range
- `SelectedKeyRoute`: JSON path (e.g., `data.price.usd`) for extracting value from API response

### 6. TaskDataEntity

Represents `task_data` table for individual task executions.

```go
type TaskDataEntity struct {
    TaskID               int64     `cql:"task_id"`                // Unique task ID (auto-increment)
    TaskNumber           int64     `cql:"task_number"`            // Task sequence number for job
    JobID                string    `cql:"job_id"`                 // Reference to job
    TaskDefinitionID     int       `cql:"task_definition_id"`     // Task type
    CreatedAt            time.Time `cql:"created_at"`             // Task creation time
    TaskOpxPredictedCost string    `cql:"task_opx_predicted_cost"`// Predicted cost (Wei)
    TaskOpxActualCost    string    `cql:"task_opx_actual_cost"`   // Actual cost including tx gas (Wei)
    ExecutionTimestamp   time.Time `cql:"execution_timestamp"`    // Execution start time
    ExecutionTxHash      string    `cql:"execution_tx_hash"`      // Transaction hash for action
    TaskPerformerID      int64     `cql:"task_performer_id"`      // Keeper who executed
    TaskAttesterIDs      []int64   `cql:"task_attester_ids"`      // Keepers who attested
    ConvertedArguments   string    `cql:"converted_arguments"`    // Arguments used (JSON)
    ProofOfTask          string    `cql:"proof_of_task"`          // Cryptographic proof
    SubmissionTxHash     string    `cql:"submission_tx_hash"`     // Aggregator submission tx hash
    IsSuccessful         bool      `cql:"is_successful"`          // Execution success
    IsAccepted           bool      `cql:"is_accepted"`            // Consensus accepted
    IsImua               bool      `cql:"is_imua"`                // IMUA task type
}
```

**Primary Key**: `task_id`

**Field Notes**:

- `TaskOpxActualCost`: Total cost in Wei (includes gas used + resource utilization)
- `TaskPerformerID`: Reference to KeeperDataEntity.OperatorID
- `TaskAttesterIDs`: Array of attester operator IDs

### 7. KeeperDataEntity

Represents `keeper_data` table for Keeper node registry.

```go
type KeeperDataEntity struct {
    KeeperName       string    `cql:"keeper_name"`        // Human-readable keeper name
    KeeperAddress    string    `cql:"keeper_address"`     // Wallet address (primary key)
    RewardsAddress   string    `cql:"rewards_address"`    // Address for reward payouts
    ConsensusAddress string    `cql:"consensus_address"`  // Consensus layer address
    RegisteredTx     string    `cql:"registered_tx"`      // Registration transaction hash
    OperatorID       int64     `cql:"operator_id"`        // Unique operator ID
    VotingPower      string    `cql:"voting_power"`       // Voting power (Wei-based stake)
    Whitelisted      bool      `cql:"whitelisted"`        // Whitelisted for special tasks
    Registered       bool      `cql:"registered"`         // Registered on-chain
    Online           bool      `cql:"online"`             // Currently online
    Version          string    `cql:"version"`            // Keeper software version
    OnImua           bool      `cql:"on_imua"`            // Supports IMUA tasks
    PublicIP         string    `cql:"public_ip"`          // Public IP for P2P
    ChatID           int64     `cql:"chat_id"`            // Telegram chat ID for alerts
    EmailID          string    `cql:"email_id"`           // Email for alerts
    RewardsBooster   string    `cql:"rewards_booster"`    // Reward multiplier (Wei-based)
    NoExecutedTasks  int64     `cql:"no_executed_tasks"`  // Total tasks executed
    NoAttestedTasks  int64     `cql:"no_attested_tasks"`  // Total tasks attested
    Uptime           int64     `cql:"uptime"`             // Total uptime (seconds)
    KeeperPoints     string    `cql:"keeper_points"`      // Total points (Wei, sum of TaskOpxCost executed/attested)
    LastCheckedIn    time.Time `cql:"last_checked_in"`    // Last heartbeat time
}
```

**Primary Key**: `keeper_address`

**Field Notes**:

- `VotingPower`: Staked amount determining consensus weight
- `KeeperPoints`: Cumulative resource utilization (performance metric)
- `Online`: Set to false if `LastCheckedIn` > 10 minutes ago

### 8. ApiKeyDataEntity

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

#### HealthCheckResponse

Used by all services for health endpoints:

```go
type HealthCheckResponse struct {
    Status    string    `json:"status"`    // "ok", "degraded", "error"
    Timestamp time.Time `json:"timestamp"` // Current time
    Service   string    `json:"service"`   // Service name
    Version   string    `json:"version"`   // Software version
    Error     string    `json:"error,omitempty"` // Error message if unhealthy
}
```

**Example Response**:

```json
{
  "status": "ok",
  "timestamp": "2025-01-15T10:30:00Z",
  "service": "dbserver",
  "version": "1.0.0"
}
```

### DBServer-Specific Types

**Location**: `pkg/types/http_dbserver.go`

Examples (typical patterns):

```go
// Request to create a new job
type CreateJobRequest struct {
    JobTitle  string `json:"job_title" binding:"required"`
    JobType   string `json:"job_type" binding:"required,oneof=time event condition"`
    // ... other fields with validation tags
}

// Response after creating a job
type CreateJobResponse struct {
    JobID   string `json:"job_id"`
    Message string `json:"message"`
}

// Generic error response
type ErrorResponse struct {
    Error   string `json:"error"`
    Code    string `json:"code"`
    Message string `json:"message"`
    TraceID string `json:"trace_id,omitempty"`
}
```

### Health Service Types

**Location**: `pkg/types/http_health.go`

```go
// Keeper check-in request
type KeeperCheckInRequest struct {
    KeeperAddress string `json:"keeper_address" binding:"required"`
    Version       string `json:"version"`
    Uptime        int64  `json:"uptime"`
}
```

### Keeper Types

**Location**: `pkg/types/http_keeper.go`

```go
// Task assignment from TaskDispatcher
type SendTaskDataToKeeper struct {
    TaskID           int64       `json:"task_id"`
    JobID            string      `json:"job_id"`
    TargetChainID    string      `json:"target_chain_id"`
    ContractAddress  string      `json:"contract_address"`
    FunctionName     string      `json:"function_name"`
    ABI              string      `json:"abi"`
    Arguments        []string    `json:"arguments"`
    ScriptURL        string      `json:"script_url"` // IPFS CID if dynamic args
}
```

---

## RPC Message Types

Used for gRPC communication between internal services.

**Location**: `pkg/types/rpc_*.go`

### Keeper RPC Types

**Location**: `pkg/types/rpc_keeper.go`

#### PerformerActionData

Data sent by Keeper after executing a task:

```go
type PerformerActionData struct {
    TaskID       int64  `json:"task_id"`
    ActionTxHash string `json:"action_tx_hash"`      // Transaction hash
    GasUsed      string `json:"gas_used"`            // Gas consumed (Wei)
    Status       bool   `json:"status"`              // Success/failure

    // Resource utilization metrics
    MemoryUsage   uint64  `json:"memory_usage"`      // Bytes
    CPUPercentage float64 `json:"cpu_percentage"`    // 0-100%
    NetworkRx     uint64  `json:"network_rx"`        // Bytes received
    NetworkTx     uint64  `json:"network_tx"`        // Bytes transmitted
    BlockRead     uint64  `json:"block_read"`        // Disk bytes read
    BlockWrite    uint64  `json:"block_write"`       // Disk bytes written
    BandwidthRate float64 `json:"bandwidth_rate"`    // Mbps

    // Cost calculation fields
    TotalFee           string       `json:"total_fee"`            // Total cost (Wei)
    StaticComplexity   float64      `json:"static_complexity"`    // Pre-execution complexity
    DynamicComplexity  float64      `json:"dynamic_complexity"`   // Runtime complexity
    ComplexityIndex    float64      `json:"complexity_index"`     // Combined complexity
    ExecutionTimestamp time.Time    `json:"execution_timestamp"`  // When executed
    ConvertedArguments []interface{} `json:"converted_arguments"` // Arguments used
}
```

**Usage**: Keeper sends this to Aggregator after executing a task.

#### ProofData

Cryptographic proof generated by Keeper:

```go
type ProofData struct {
    TaskID               int64     `json:"task_id"`
    ProofOfTask          string    `json:"proof_of_task"`          // Merkle proof or signature
    CertificateHash      string    `json:"certificate_hash"`       // Hash of execution certificate
    CertificateTimestamp time.Time `json:"certificate_timestamp"`  // Proof generation time
}
```

#### PerformerSignatureData

Keeper's signature on task execution:

```go
type PerformerSignatureData struct {
    TaskID                  int64  `json:"task_id"`
    PerformerSigningAddress string `json:"performer_signing_address"` // Keeper's address
    PerformerSignature      string `json:"performer_signature"`       // ECDSA signature
}
```

#### IPFSData

Complete data package uploaded to IPFS:

```go
type IPFSData struct {
    TaskData           *SendTaskDataToKeeper   `json:"task_data"`              // Task definition
    ActionData         *PerformerActionData    `json:"action_data"`            // Execution results
    ProofData          *ProofData              `json:"proof_data"`             // Cryptographic proof
    PerformerSignature *PerformerSignatureData `json:"performer_signature_data"` // Signature
}
```

#### BroadcastDataForValidators

Data broadcast to attester Keepers for validation:

```go
type BroadcastDataForValidators struct {
    ProofOfTask        string `json:"proof_of_task"`        // Proof to verify
    Data               []byte `json:"data"`                 // Serialized task data
    TaskDefinitionID   int    `json:"task_definition_id"`   // Task type
    PerformerAddress   string `json:"performer_address"`    // Executor's address
    PerformerSignature string `json:"performer_signature"`  // Executor's signature
    SignatureType      string `json:"signature_type"`       // "ECDSA", "BLS", etc.
    TargetChainID      int    `json:"target_chain_id"`      // Chain ID
}
```

**Usage**: Performer broadcasts this via P2P for attesters to validate.

### Scheduler RPC Types

**Location**: `pkg/types/rpc_schedulers.go`

Examples (typical patterns):

```go
// Request to fetch jobs from DBServer
type FetchJobsRequest struct {
    JobType string `json:"job_type"` // "time", "event", "condition"
    Limit   int    `json:"limit"`
    Offset  int    `json:"offset"`
}

// Task pushed to Redis stream
type ScheduledTask struct {
    TaskID      int64     `json:"task_id"`
    JobID       string    `json:"job_id"`
    ScheduledAt time.Time `json:"scheduled_at"`
}
```

---

## Constants and Enums

**Location**: `pkg/types/constants.go`

### Task Definition IDs

```go
const (
    TaskDefinitionSimple       = 0  // Simple contract call
    TaskDefinitionComplex      = 1  // Complex contract call with validation
    TaskDefinitionCrossChain   = 2  // Cross-chain execution
    TaskDefinitionDataFetch    = 3  // Fetch external data and execute
)
```

### Job Types

```go
const (
    JobTypeTime      = "time"       // Time-based scheduler
    JobTypeEvent     = "event"      // Event-based scheduler
    JobTypeCondition = "condition"  // Condition-based scheduler
)
```

### Job Status

```go
const (
    JobStatusCreated   = "created"    // Job created but not yet active
    JobStatusRunning   = "running"    // Job actively monitored by scheduler
    JobStatusCompleted = "completed"  // Job finished (non-recurring)
    JobStatusFailed    = "failed"     // Job failed permanently
    JobStatusExpired   = "expired"    // Job expired (past expiration time)
    JobStatusPaused    = "paused"     // Job temporarily paused by user
)
```

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
- `KeeperDataEntity.RewardsBooster`
- `KeeperDataEntity.KeeperPoints`

### Calculation Examples

**Location**: `pkg/types/bigint_operations.go`

```go
// Convert TG to Wei
func TGToWei(tg float64) *big.Int

// Convert Wei to TG
func WeiToTG(wei *big.Int) float64

// Add two Wei values
func AddWei(a, b *big.Int) *big.Int

// Calculate percentage
func PercentageOfWei(amount *big.Int, percentage float64) *big.Int
```

### Usage Example

```go
import "math/big"

// Task cost calculation
gasCost := big.NewInt(500000)  // 500,000 Wei
resourceCost := big.NewInt(250000) // 250,000 Wei
totalCost := AddWei(gasCost, resourceCost) // 750,000 Wei

// Store as string
taskEntity.TaskOpxActualCost = totalCost.String() // "750000"

// User points update
userPoints, _ := new(big.Int).SetString(userEntity.UserPoints, 10)
newPoints := AddWei(userPoints, totalCost)
userEntity.UserPoints = newPoints.String()
```

### Why Wei?

✅ **No Precision Loss**: Integer arithmetic, no floating-point errors  
✅ **Consistency**: Same unit used in smart contracts  
✅ **Deterministic**: Exact calculations, reproducible results  
✅ **Ethereum Compatible**: Direct mapping to blockchain values  

---

## Validation Rules

Data validation is handled by `go-playground/validator` with struct tags.

**Location**: `pkg/types/validators.go`

### Common Validation Tags

```go
type ExampleRequest struct {
    Email       string `json:"email" binding:"required,email"`
    Address     string `json:"address" binding:"required,eth_addr"`
    Amount      string `json:"amount" binding:"required,numeric"`
    JobType     string `json:"job_type" binding:"required,oneof=time event condition"`
    URL         string `json:"url" binding:"omitempty,url"`
    ChainID     string `json:"chain_id" binding:"required,min=1"`
}
```

### Custom Validators

```go
// Ethereum address validator
func ValidateEthAddress(fl validator.FieldLevel) bool

// IPFS CID validator
func ValidateIPFSCID(fl validator.FieldLevel) bool

// Cron expression validator
func ValidateCronExpression(fl validator.FieldLevel) bool
```

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

For understanding how data flows through the system using these types, see [6_dataflow.md](./6_dataflow.md).
