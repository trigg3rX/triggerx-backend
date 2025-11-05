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
9. [Keeper Health Monitoring Flow](#keeper-health-monitoring-flow)
10=. [Trace ID Propagation](#trace-id-propagation)

---

## Overview

TriggerX follows an event-driven, asynchronous data flow architecture:

```bash
User Input → API Gateway → Database → Schedulers → Task Queue
    ↓
Task Dispatcher → Keeper Selection → Task Assignment
    ↓
Keeper Execution → Result Broadcast → Attestation
    ↓
Aggregator Consensus → Blockchain Submission → Points Settlement
```

Each stage involves multiple services communicating via HTTP, gRPC, or P2P protocols with OpenTelemetry tracing for observability.

---

## Job Creation Flow

```bash
┌────────────────┐
│      User      │ 1. POST /api/jobs with job definition
│    (SDK/UI)    │
└────────┬───────┘
         │
         ▼
┌────────────────┐
│    DBServer    │ 2. Pass X-Trace-ID
│   API Gateway  │ 3. Validate API Key
└────────┬───────┘ 4. Validate job data
         │         5. Convert DTO to Entity
         ▼
┌────────────────┐
│    ScyllaDB    │ 6. Insert JobDataEntity
│                │ 7. Insert TimeJobData/EventJobData/ConditionJobData
└────────┬───────┘ 8. Update UserDataEntity
         │         9. Return job_id to user
         ▼
┌────────────────┐
│   Schedulers   │ 10. Polling: Detect new jobs
│  (all 3 types) │ 11. Load job into memory
└────────────────┘ 12. Start monitoring triggers
```

---

## Time-Based Execution Flow

```bash
┌─────────────────┐
│  Time Scheduler │ Poll every 30 seconds
└────────┬────────┘
         │          1. Fetch all active time jobs from database
         ▼               WHERE status='running' AND next_execution_timestamp < now() + 40s
┌─────────────────┐ For Each Job:
│  Task Creation  │ 2. Parse cron expression / calculate next interval, set next_execution_timestamp
│    (ScyllaDB)   │ 3. Task creation
└────────┬────────┘
         │         
         ▼
┌─────────────────┐ 4. Publish to Redis Stream
│   Redis Stream  │ 5. Dispatch to aggregator
└─────────────────┘
```

---

## Event-Based Execution Flow

```bash
┌─────────────────┐
│ Event Scheduler │ On startup
└────────┬────────┘
         │          1. DBServer calls SendTaskToScheduler with ScheduleConditionJobData
         ▼
┌─────────────────┐
│     Worker      │ 2. Parse trigger_chain_id, contract_address, event
└────────┬────────┘ 3. Create WebSocket subscription to RPC node
         |               eth_subscribe("logs", {
         │               "address": "0x1234...",
         ▼               "topics": ["Transfer(address,address,uint256)"] })
┌─────────────────┐
│  Task Creation  │ 4. Trigger notification from worker
│    (ScyllaDB)   │ 5. Task creation
└────────┬────────┘ 6. Stop Worker if recurring = false, or expiration_time > now()
         │
         ▼
┌─────────────────┐ 7. Publish to Redis Stream
│   Redis Stream  │ 8. Dispatch to aggregator
└─────────────────┘
```

---

## Condition-Based Execution Flow

```bash
┌───────────────────┐
│Condition Scheduler│ On startup
└────────┬──────────┘
         │          1. DBServer calls SendTaskToScheduler with ScheduleConditionJobData
         ▼
┌─────────────────┐
│     Worker      │ 2. Subscribe to WebSocket, Oracle
└────────┬────────┘ 3. OR Poll API / Oracle endpoint each second
         │          └─ Check the value from SelectedKeyRoute against the limits and condition type
         ▼
┌─────────────────┐
│  Task Creation  │ 4. Trigger notification from worker
│    (ScyllaDB)   │ 5. Task creation
└────────┬────────┘ 6. Stop Worker if recurring = false, or expiration_time > now()
         │
         ▼
┌─────────────────┐ 7. Publish to Redis Stream
│   Redis Stream  │ 8. Dispatch to aggregator
└─────────────────┘
```

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

For testing strategies related to these flows, see [7_tests.md](./7_tests.md).
