# Data Flow

This document traces how data moves through the TriggerX Backend system, from user input to task execution and result recording. Understanding these flows is crucial for debugging, optimization, and extending the platform.

---

## Table of Contents

1. [Overview](#overview)
2. [Job Creation Flow](#job-creation-flow)
3. [Time-Based Execution Flow](#time-based-execution-flow)
4. [Event-Based Execution Flow](#event-based-execution-flow)
5. [Condition-Based Execution Flow](#condition-based-execution-flow)
6. [Task Execution Flow](#task-execution-flow)
7. [Task Validation Flow](#task-validation-flow)
8. [Consensus and Blockchain Submission](#consensus-and-blockchain-submission)
9. [Cost Calculation and Reward Distribution](#cost-calculation-and-reward-distribution)
10. [Error Handling and Retry Flow](#error-handling-and-retry-flow)
11. [Keeper Health Monitoring Flow](#keeper-health-monitoring-flow)
12. [Trace ID Propagation](#trace-id-propagation)

---

## Data Flow Architecture

### Job Creation Flow

```bash
User → SDK/UI → DBServer → Database
                    ↓
              Schedulers (subscribe to new jobs)
```

### Task Execution Flow

```bash
Scheduler → Trigger Detection → TaskDispatcher → Keeper Selection
                                      ↓
                              Assign Performer Role
                                      ↓
                                  Send Task → Keeper (Performer)
                                      ↓
                              Docker Execution
                                      ↓
                              Submit to P2P Network
                                      ↓
                              Keepers (Attesters) → Validate Task
                                      ↓
                              Submit Attestations
                                      ↓
                          Aggregator → Consensus → Blockchain
                                      ↓
                          TaskMonitor → Update Status → Database
```

## Overview

TriggerX follows an event-driven, asynchronous data flow architecture:

```bash
User Input → API Gateway → Database → Schedulers → Task Queue
    ↓
Task Dispatcher → Keeper Selection → Task Assignment
    ↓
Keeper Execution → Result Broadcast → Attestation
    ↓
Aggregator Consensus → Blockchain Submission → Cost Settlement
```

Each stage involves multiple services communicating via HTTP, gRPC, or P2P protocols with OpenTelemetry tracing for observability.

---

## Job Creation Flow

### Step-by-Step Flow

```bash
┌──────────┐
│   User   │
│ (SDK/UI) │
└────┬─────┘
     │ 1. POST /api/jobs
     │    with job definition
     ▼
┌─────────────────┐
│    DBServer     │ 2. Generate X-Trace-ID (if not present)
│   API Gateway   │ 3. Validate API Key
└────┬────────────┘ 4. Validate job data
     │ 5. Convert DTO to Entity
     │
     ▼
┌─────────────────┐
│   ScyllaDB      │ 6. Insert JobDataEntity
│                 │ 7. Insert TimeJobData/EventJobData/ConditionJobData
│                 │ 8. Update UserDataEntity (increment total_jobs)
└────┬────────────┘
     │ 9. Return job_id to user
     │
     ▼
┌─────────────────┐
│  Schedulers     │ 10. Polling: Detect new jobs
│ (all 3 types)   │ 11. Load job into memory
└─────────────────┘ 12. Start monitoring triggers
```

### Data Transformation

**Input (JSON via HTTP)**:

```json
{
  "job_title": "Daily Reward Claim",
  "job_type": "time",
  "timezone": "America/New_York",
  "schedule_type": "cron",
  "cron_expression": "0 0 * * *",
  "target_chain_id": "42161",
  "target_contract_address": "0x1234...",
  "target_function": "claimRewards",
  "abi": "[...]",
  "arg_type": 0
}
```

**Processing**:

1. DBServer validates and converts to `CreateJobRequest` (HTTP type)
2. Converts to `JobDataDTO` and `TimeJobDataDTO`
3. Converts to `JobDataEntity` and `TimeJobDataEntity`
4. Inserts into ScyllaDB

**Output (Response)**:

```json
{
  "job_id": "1234567890",
  "message": "Job created successfully",
  "status": "running"
}
```

### Trace ID Flow

- **User provides**: `X-Trace-ID: tgrx-frnt-<uuid>` (optional)
- **DBServer generates**: If not provided, generates `tgrx-frnt-<uuid>`
- **Propagates**: Trace ID included in logs, database records, and forwarded to schedulers

---

## Time-Based Execution Flow

### Trigger Detection

```bash
┌──────────────────┐
│  Time Scheduler  │ Every 10 seconds (configurable)
└────┬─────────────┘
     │ 1. Fetch all active time jobs from database
     │    WHERE status='running' AND expiration_time > now()
     ▼
┌─────────────────────────────────────────────────────────┐
│  For Each Job:                                          │
│  2. Parse cron expression or calculate next interval    │
│  3. Check: Is current_time >= next_execution_timestamp? │
│     ├─ Yes → Trigger job (create task)                  │
│     └─ No → Skip                                        │
└────┬────────────────────────────────────────────────────┘
     │ 4. Task creation
     ▼
┌─────────────────┐
│ Task Creation   │ 5. Generate task_id
│                 │ 6. Create TaskDataEntity
│                 │ 7. Insert into ScyllaDB
└────┬────────────┘ 8. Publish to Redis Stream
     │
     ▼
┌─────────────────┐
│  Redis Stream   │ 9. Task queued in "time_scheduler_stream"
│  (Task Queue)   │
└─────────────────┘
```

### Cron Expression Parsing

Example: `0 0 * * *` (daily at midnight)

```go
// Pseudocode
nextExecution := cronParser.Next(lastExecuted, timezone)
if now >= nextExecution {
    createTask(job)
    job.NextExecutionTimestamp = cronParser.Next(now, timezone)
    updateDatabase(job)
}
```

### Redis Stream Format

```bash
XADD time_scheduler_stream * \
  task_id 1001 \
  job_id 1234567890 \
  scheduled_at 2025-01-15T00:00:00Z \
  trace_id tgrx-frnt-uuid
```

---

## Event-Based Execution Flow

### WebSocket Subscription

```bash
┌──────────────────┐
│  Event Scheduler │ On startup
└────┬─────────────┘
     │ 1. Fetch all active event jobs from database
     │    WHERE status='running' AND expiration_time > now()
     ▼
┌─────────────────────────────────────────────────────────┐
│  For Each Job:                                          │
│  2. Parse trigger_chain_id, contract_address, event     │
│  3. Create WebSocket subscription to RPC node           │
│     eth_subscribe("logs", {                             │
│       "address": "0x1234...",                           │
│       "topics": ["Transfer(address,address,uint256)"]   │
│     })                                                  │
└────┬────────────────────────────────────────────────────┘
     │ 4. Persistent connection maintained
     ▼
┌─────────────────┐
│  RPC Node       │ 5. Event emitted on blockchain
│  (WebSocket)    │
└────┬────────────┘
     │ 6. Event notification received
     ▼
┌─────────────────┐
│  Event Filter   │ 7. Check trigger_event_filter_para_name
│                 │ 8. Compare trigger_event_filter_value
│                 │    ├─ Match → Proceed to task creation
└────┬────────────┘    └─ No match → Discard event
     │ 9. Task creation
     ▼
┌─────────────────┐
│  Redis Stream   │ 10. Task queued in "event_scheduler_stream"
│                 │
└─────────────────┘
```

### Event Filtering Example

**Job Configuration**:

```json
{
  "trigger_event": "Transfer(address,address,uint256)",
  "trigger_event_filter_para_name": "to",
  "trigger_event_filter_value": "0x5678..."
}
```

**Event Received**:

```json
{
  "event": "Transfer",
  "topics": [
    "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef",
    "0x000000000000000000000000abcd...",
    "0x0000000000000000000000005678..."  // ← Matches filter!
  ],
  "data": "0x000000000000000000000000000000000000000000000000000000000000000a"
}
```

**Result**: Task created because `to` address matches `0x5678...`

### Block Confirmations

```bash
Event Detected → Wait N Blocks → Confirm No Reorg → Create Task
```

Example: If `required_confirmations = 12`:

- Event seen at block 1000
- Wait until block 1012
- Verify event still exists (no chain reorg)
- Create task

---

## Condition-Based Execution Flow

### Polling Mechanism

```bash
┌──────────────────┐
│Condition Scheduler│ Every 60 seconds (default, per-job configurable)
└────┬─────────────┘
     │ 1. Fetch all active condition jobs
     ▼
┌─────────────────────────────────────────────────────────┐
│  For Each Job:                                          │
│  2. Determine value_source_type                         │
│     ├─ "contract" → Query contract state                │
│     ├─ "api" → Fetch from API endpoint                  │
│     └─ "oracle" → Query oracle contract                 │
└────┬────────────────────────────────────────────────────┘
     │ 3. Extract value
     ▼
┌─────────────────────────────────────────────────────────┐
│  Value Extraction:                                      │
│  • Contract: eth_call to read state variable            │
│  • API: HTTP GET + JSON path extraction                 │
│  • Oracle: Call oracle contract's read function         │
└────┬────────────────────────────────────────────────────┘
     │ 4. Condition evaluation
     ▼
┌─────────────────────────────────────────────────────────┐
│  Evaluate Condition:                                    │
│  5. Compare value with upper_limit and lower_limit      │
│     condition_type = "above" → value > upper_limit      │
│     condition_type = "below" → value < lower_limit      │
│     condition_type = "range" → lower < value < upper    │
│     condition_type = "outside" → value < lower OR > upper│
└────┬────────────────────────────────────────────────────┘
     │ 6. Condition met?
     ▼
┌─────────────────┐
│ Edge Detection  │ 7. Check last_state_value
│                 │ 8. If state changed from false → true
│                 │    ├─ Yes → Create task
└────┬────────────┘    └─ No → Skip (prevent duplicate triggers)
     │
     ▼
┌─────────────────┐
│  Redis Stream   │ 9. Task queued in "condition_scheduler_stream"
│                 │
└─────────────────┘
```

### API Value Extraction

**Job Configuration**:

```json
{
  "value_source_type": "api",
  "value_source_url": "https://api.coingecko.com/api/v3/simple/price?ids=ethereum&vs_currencies=usd",
  "selected_key_route": "ethereum.usd",
  "condition_type": "above",
  "upper_limit": 3000
}
```

**API Response**:

```json
{
  "ethereum": {
    "usd": 3150
  }
}
```

**Processing**:

1. Fetch URL
2. Parse JSON
3. Navigate path: `ethereum.usd` → 3150
4. Evaluate: 3150 > 3000 → **true** → Create task

### Contract State Reading

**Job Configuration**:

```json
{
  "value_source_type": "contract",
  "value_source_url": "0xContractAddress",
  "selected_key_route": "balanceOf(0xUserAddress)",
  "condition_type": "above",
  "upper_limit": 100
}
```

**Processing**:

1. Encode function call: `balanceOf(address)`
2. `eth_call` to contract
3. Decode result: 150 tokens
4. Evaluate: 150 > 100 → **true** → Create task

---

## Task Execution Flow

### Task Assignment to Keeper

```bash
┌─────────────────┐
│ TaskDispatcher  │ Consumes from Redis Streams (Consumer Group)
└────┬────────────┘
     │ 1. Read task from stream (XREADGROUP)
     ▼
┌─────────────────────────────────────────────────────────┐
│  Keeper Selection:                                      │
│  2. Fetch all online Keepers from database              │
│     WHERE online=true AND registered=true               │
│  3. Filter by requirements (version, whitelist, etc.)   │
│  4. Score each Keeper:                                  │
│     - Performance score (40%)                           │
│     - Load score (30%)                                  │
│     - Reputation score (20%)                            │
│     - Specialization score (10%)                        │
│  5. Select Keeper with highest score                    │
└────┬────────────────────────────────────────────────────┘
     │ 6. Assign performer role
     ▼
┌─────────────────┐
│  P2P Network    │ 7. Send task to Keeper via libp2p
│                 │    Message: SendTaskDataToKeeper
└────┬────────────┘ 8. Wait for acknowledgment (30s timeout)
     │
     ▼
┌─────────────────┐
│  Keeper Node    │ 9. Receive task assignment
│                 │
└─────────────────┘
```

### Keeper Task Processing

```bash
┌─────────────────┐
│  Keeper Node    │
└────┬────────────┘
     │ 1. Validate task signature
     │ 2. Check task parameters
     ▼
┌─────────────────┐
│ IPFS Download   │ 3. If arg_type=2 (dynamic):
│                 │    - Fetch script from IPFS (dynamic_arguments_script_url)
└────┬────────────┘    - Store locally
     │
     ▼
┌─────────────────────────────────────────────────────────┐
│  Docker Container Setup:                                │
│  4. Create container with:                              │
│     - Resource limits (CPU: 1 core, Memory: 512MB)      │
│     - Seccomp profile (restricted syscalls)             │
│     - Network mode (limited or none)                    │
│     - Read-only filesystem                              │
│  5. Mount script and inject arguments                   │
└────┬────────────────────────────────────────────────────┘
     │ 6. Start container
     ▼
┌─────────────────────────────────────────────────────────┐
│  Script Execution:                                      │
│  7. Execute script inside container                     │
│     - For dynamic args: Run script to generate args     │
│     - For static args: Use provided args                │
│  8. Call target contract function with args             │
│  9. Capture transaction hash, gas used, status          │
│  10. Collect resource metrics (CPU, memory, network)    │
└────┬────────────────────────────────────────────────────┘
     │ 11. Container cleanup
     ▼
┌─────────────────┐
│  Result Storage │ 12. Store execution results locally
│                 │ 13. Calculate cost (gas + resources)
└────┬────────────┘ 14. Generate proof
     │
     ▼
┌─────────────────┐
│ IPFS Upload     │ 15. Package IPFSData:
│                 │     - Task data
│                 │     - Action data (PerformerActionData)
│                 │     - Proof data (ProofData)
│                 │     - Signature (PerformerSignatureData)
└────┬────────────┘ 16. Upload to IPFS → Get CID
     │
     ▼
┌─────────────────┐
│  P2P Broadcast  │ 17. Broadcast BroadcastDataForValidators
│                 │     to attester Keepers via Gossipsub
└─────────────────┘
```

### Resource Metric Collection

During execution, Keeper monitors Docker container:

```go
// Pseudocode
stats := docker.ContainerStats(containerID)

metrics := PerformerActionData{
    MemoryUsage:   stats.MemoryStats.Usage,        // Bytes
    CPUPercentage: calculateCPUPercent(stats),     // 0-100%
    NetworkRx:     stats.Networks.Eth0.RxBytes,    // Bytes received
    NetworkTx:     stats.Networks.Eth0.TxBytes,    // Bytes transmitted
    BlockRead:     stats.BlkioStats.IoServiceBytesRecursive.Read,
    BlockWrite:    stats.BlkioStats.IoServiceBytesRecursive.Write,
}
```

---

## Task Validation Flow

### Attester Selection and Validation

```bash
┌─────────────────┐
│ Attester Keeper │ Listening on P2P Gossipsub
└────┬────────────┘
     │ 1. Receive BroadcastDataForValidators
     ▼
┌─────────────────────────────────────────────────────────┐
│  Selection:                                             │
│  2. Aggregator randomly selects N attesters (e.g., 5)   │
│  3. Selected attesters notified via P2P                 │
│  4. Non-selected attesters ignore                       │
└────┬────────────────────────────────────────────────────┘
     │ 5. Attester validates
     ▼
┌─────────────────────────────────────────────────────────┐
│  Validation Process:                                    │
│  6. Download task data from IPFS (using CID)            │
│  7. Verify performer's signature                        │
│  8. Re-execute task with same inputs:                   │
│     - Same contract, function, arguments                │
│     - Same chain state (or close enough)                │
│  9. Compare results:                                    │
│     - Transaction hash (should match)                   │
│     - Gas used (should be similar)                      │
│     - Status (success/failure should match)             │
└────┬────────────────────────────────────────────────────┘
     │ 10. Attestation decision
     ▼
┌─────────────────┐
│  Sign Attestation│ 11. If results match:
│                 │     - Sign positive attestation
│                 │     Else:
│                 │     - Sign negative attestation
└────┬────────────┘ 12. Include attester's signature
     │
     ▼
┌─────────────────┐
│  Submit to      │ 13. Send attestation to Aggregator
│  Aggregator     │     via P2P or HTTP
└─────────────────┘
```

### Attestation Message Format

```json
{
  "task_id": 1001,
  "attester_address": "0xAttester...",
  "attestation_result": true,  // true = valid, false = invalid
  "signature": "0x...",
  "timestamp": "2025-01-15T10:30:00Z"
}
```

---

## Consensus and Blockchain Submission

### Aggregator Consensus Flow

```bash
┌─────────────────┐
│   Aggregator    │
└────┬────────────┘
     │ 1. Collect performer result
     │ 2. Collect N attester attestations
     ▼
┌─────────────────────────────────────────────────────────┐
│  Consensus Algorithm (BFT):                             │
│  3. Count attestations:                                 │
│     - Positive: attestation_result = true               │
│     - Negative: attestation_result = false              │
│  4. Calculate threshold (2/3+ majority):                │
│     required = ceil(N * 2 / 3)                          │
│  5. Check consensus:                                    │
│     - If positive >= required → Task VALID              │
│     - Else → Task INVALID                               │
└────┬────────────────────────────────────────────────────┘
     │ 6. Consensus decision
     ▼
┌─────────────────────────────────────────────────────────┐
│  If Task VALID:                                         │
│  7. Aggregate signatures (BLS signature aggregation)    │
│  8. Generate Merkle proof for batch                     │
│  9. Prepare blockchain submission payload:              │
│     - Task ID                                           │
│     - Proof of task (Merkle root)                       │
│     - Aggregated signature                              │
│     - Performer and attester addresses                  │
└────┬────────────────────────────────────────────────────┘
     │ 10. Submit to blockchain
     ▼
┌─────────────────┐
│  L2 Blockchain  │ 11. Call TriggerX Validation Contract
│  (Arbitrum/Base)│     function submitTaskValidation(
│                 │       uint256 taskId,
│                 │       bytes32 proofOfTask,
│                 │       bytes aggregatedSignature,
│                 │       address[] attesters
└────┬────────────┘     )
     │ 12. Transaction confirmed
     ▼
┌─────────────────┐
│  DBServer       │ 13. Update TaskDataEntity:
│  (Update DB)    │     - submission_tx_hash
│                 │     - is_accepted = true
│                 │     - proof_of_task
└─────────────────┘ 14. Trigger reward distribution
```

### Batch Submission Optimization

For efficiency, tasks are batched:

```bash
Tasks: [1001, 1002, 1003, 1004, 1005]
    ↓
Build Merkle Tree:
         Root
        /    \
      H12    H34
      / \    / \
    H1  H2 H3  H4
    ↓
Submit to Blockchain:
  - Merkle root
  - Task IDs
  - Aggregated signatures
```

Single transaction validates multiple tasks → Lower gas costs

---

## Cost Calculation and Reward Distribution

### Task Cost Calculation

```bash
┌─────────────────┐
│  Keeper Node    │ After execution
└────┬────────────┘
     │ 1. Collect metrics
     ▼
┌─────────────────────────────────────────────────────────┐
│  Cost Components:                                       │
│                                                          │
│  Gas Cost (Wei):                                        │
│    gas_used * gas_price                                 │
│                                                          │
│  Resource Cost (Wei):                                   │
│    cpu_cost = cpu_percentage * cpu_price_per_second     │
│    memory_cost = memory_mb * memory_price_per_mb        │
│    network_cost = (network_rx + network_tx) * price_per_gb│
│    disk_cost = (block_read + block_write) * price_per_gb│
│                                                          │
│  Complexity Cost (Wei):                                 │
│    complexity_index * base_complexity_fee               │
│                                                          │
│  Total Cost:                                            │
│    total = gas_cost + resource_cost + complexity_cost   │
└────┬────────────────────────────────────────────────────┘
     │ 2. Store as task_opx_actual_cost (Wei string)
     ▼
┌─────────────────┐
│  Database       │ 3. Update TaskDataEntity
│                 │ 4. Aggregate for JobDataEntity.job_cost_actual
└─────────────────┘ 5. Aggregate for UserDataEntity.user_points
```

### Reward Distribution

After blockchain submission and cost settlement:

```bash
Total Task Fee (from job_cost_actual)
    ↓
┌───────────────────────────────────────────────────┐
│  Reward Split:                                    │
│                                                    │
│  Performer Keeper: 70%                            │
│    reward = total_fee * 0.70                      │
│                                                    │
│  Attester Keepers: 20% (split equally)            │
│    each = (total_fee * 0.20) / num_attesters      │
│                                                    │
│  Protocol Treasury: 10%                           │
│    treasury = total_fee * 0.10                    │
└───────────────────────────────────────────────────┘
    ↓
Smart Contract Execution:
    - Transfer rewards to keeper reward_addresses
    - Update keeper_points for each Keeper
    - Log reward event on blockchain
```

### Keeper Points Update

```sql
-- Performer Keeper
UPDATE keeper_data 
SET keeper_points = keeper_points + task_opx_actual_cost,
    no_executed_tasks = no_executed_tasks + 1
WHERE keeper_address = performer_address;

-- Attester Keepers
UPDATE keeper_data
SET keeper_points = keeper_points + attestation_cost,
    no_attested_tasks = no_attested_tasks + 1
WHERE keeper_address IN (attester_addresses);
```

**Note**: `keeper_points` and `user_points` are Wei-based, stored as strings.

---

## Error Handling and Retry Flow

### Task Failure Scenarios

```bash
┌─────────────────────────────────────────────────────────┐
│  Failure Points:                                        │
│                                                          │
│  1. Keeper Acknowledgment Timeout                       │
│     - Keeper doesn't respond within 30s                 │
│     → TaskDispatcher reassigns to different Keeper      │
│                                                          │
│  2. Execution Failure                                   │
│     - Container crashes or script fails                 │
│     → Keeper reports failure status                     │
│     → TaskMonitor triggers retry                        │
│                                                          │
│  3. Keeper Offline During Execution                     │
│     - Keeper goes offline mid-execution                 │
│     → TaskMonitor timeout detection                     │
│     → Reassign to new Keeper                            │
│                                                          │
│  4. Consensus Failure                                   │
│     - Attesters disagree (no 2/3 majority)              │
│     → Task marked invalid                               │
│     → Performer may be slashed                          │
│                                                          │
│  5. Blockchain Submission Failure                       │
│     - Transaction reverts or times out                  │
│     → Aggregator retries submission (exponential backoff)│
└─────────────────────────────────────────────────────────┘
```

### Retry Flow

```bash
Task Failed
    ↓
┌───────────────────────────────────────┐
│  Check Retry Count:                   │
│  - If retry_count < MAX_RETRY (3):    │
│    ├─ Increment retry_count           │
│    ├─ Calculate backoff delay         │
│    │   delay = base_delay * 2^retry   │
│    └─ Re-queue task after delay       │
│  - Else:                              │
│    └─ Mark as permanently failed      │
└───────────────────────────────────────┘
    ↓
TaskDispatcher reassigns to different Keeper
    ↓
Retry execution...
```

### Exponential Backoff

```bash
Retry 1: 1 second delay
Retry 2: 2 seconds delay
Retry 3: 4 seconds delay
Retry 4 (max): Permanent failure
```

---

## Keeper Health Monitoring Flow

### Heartbeat Flow

```bash
┌─────────────────┐
│  Keeper Node    │ Every 60 seconds
└────┬────────────┘
     │ 1. POST /keeper/checkin
     │    to Health Service
     ▼
┌─────────────────────────────────────────────────────────┐
│  Health Service:                                        │
│  2. Receive check-in request with:                      │
│     - keeper_address                                    │
│     - version                                           │
│     - uptime                                            │
│  3. Update database:                                    │
│     UPDATE keeper_data                                  │
│     SET last_checked_in = now(),                        │
│         online = true,                                  │
│         uptime = uptime + 60                            │
│     WHERE keeper_address = address                      │
└─────────────────────────────────────────────────────────┘
```

### Offline Detection Flow

```bash
┌─────────────────┐
│  Health Service │ Every 60 seconds (background job)
└────┬────────────┘
     │ 1. Scan database
     ▼
┌─────────────────────────────────────────────────────────┐
│  Query:                                                 │
│  SELECT * FROM keeper_data                              │
│  WHERE online = true                                    │
│    AND last_checked_in < now() - INTERVAL '10 minutes' │
└────┬────────────────────────────────────────────────────┘
     │ 2. For each offline Keeper:
     ▼
┌─────────────────────────────────────────────────────────┐
│  Alert Generation:                                      │
│  3. Mark as offline:                                    │
│     UPDATE keeper_data SET online = false               │
│  4. Fetch contact info (chat_id, email_id)              │
│  5. Send alerts:                                        │
│     - Telegram message to chat_id                       │
│     - Email to email_id                                 │
│  6. Log alert event                                     │
└─────────────────────────────────────────────────────────┘
```

---

## Trace ID Propagation

End-to-end tracing with OpenTelemetry ensures observability across all services.

### Trace ID Lifecycle

```bash
┌──────────┐
│   User   │ 1. Request with X-Trace-ID: tgrx-frnt-<uuid>
└────┬─────┘    (or generated by DBServer if not provided)
     │
     ▼
┌─────────────────┐
│    DBServer     │ 2. Extract or generate trace ID
│   (Root Span)   │ 3. Create root span with trace ID
└────┬────────────┘ 4. Log: [tgrx-frnt-uuid] Request received
     │ 5. Forward trace ID via gRPC metadata
     ▼
┌─────────────────┐
│   Scheduler     │ 6. Extract trace ID from gRPC context
│  (Child Span)   │ 7. Create child span
└────┬────────────┘ 8. Log: [tgrx-frnt-uuid] Job scheduled
     │ 9. Publish to Redis with trace ID
     ▼
┌─────────────────┐
│ TaskDispatcher  │ 10. Extract trace ID from Redis message
│  (Child Span)   │ 11. Create child span
└────┬────────────┘ 12. Log: [tgrx-frnt-uuid] Task assigned
     │ 13. Send trace ID to Keeper via P2P
     ▼
┌─────────────────┐
│  Keeper Node    │ 14. Extract trace ID from P2P message
│  (Child Span)   │ 15. Create child span
└────┬────────────┘ 16. Log: [tgrx-frnt-uuid] Task executed
     │ 17. Include trace ID in broadcast
     ▼
┌─────────────────┐
│  Aggregator     │ 18. Extract trace ID
│  (Child Span)   │ 19. Create child span
└────┬────────────┘ 20. Log: [tgrx-frnt-uuid] Consensus achieved
     │ 21. Store trace ID in database
     ▼
┌─────────────────┐
│   ScyllaDB      │ 22. Task record includes trace_id field
│                 │     (for debugging and audit)
└─────────────────┘
```

### Trace Visualization in Grafana Tempo

Search by `tgrx-frnt-<uuid>` to see:

- Complete request timeline
- Service boundaries and latencies
- Errors and retry attempts
- Database query times
- P2P communication delays

**Example Span Tree**:

```bash
tgrx-frnt-550e8400...
├─ DBServer: POST /api/jobs (10ms)
├─ TimeScheduler: Fetch job (5ms)
│  └─ ScyllaDB Query (3ms)
├─ TaskDispatcher: Assign task (50ms)
│  ├─ Keeper Selection (20ms)
│  └─ P2P Send (30ms)
├─ Keeper: Execute task (2000ms)
│  ├─ IPFS Download (100ms)
│  ├─ Docker Execution (1800ms)
│  └─ IPFS Upload (100ms)
├─ Aggregator: Consensus (500ms)
│  ├─ Collect Attestations (300ms)
│  └─ Blockchain Submit (200ms)
└─ Total: 2565ms
```

---

## Summary

TriggerX's data flow architecture provides:

✅ **Asynchronous Processing**: Non-blocking operations via Redis Streams  
✅ **Distributed Execution**: Decentralized Keepers with P2P coordination  
✅ **Fault Tolerance**: Retry mechanisms and failover strategies  
✅ **Transparency**: End-to-end tracing with OpenTelemetry  
✅ **Cost Efficiency**: Batched blockchain submissions and optimized resource usage  
✅ **Security**: Multi-layer validation with BFT consensus  

Understanding these flows enables you to:

- Debug issues by tracing data through the system
- Optimize performance by identifying bottlenecks
- Extend functionality by adding new data transformation steps
- Monitor system health with observability tools

---

For testing strategies related to these flows, see [7_tests.md](./7_tests.md).
