# Services

TriggerX Backend consists of multiple microservices, each with a specific responsibility. This document provides a comprehensive overview of each service, its architecture, responsibilities, and integration points.

---

## Table of Contents

1. [DBServer](#1-dbserver)
2. [Schedulers](#2-schedulers)
   - [Time Scheduler](#time-scheduler)
   - [Condition Scheduler](#condition-scheduler)
3. [TaskDispatcher](#3-taskdispatcher)
4. [TaskMonitor](#4-taskmonitor)
5. [Keeper](#5-keeper)
6. [Health](#6-health)
7. [Aggregator](#7-aggregator)

---

## 1. DBServer

The **DBServer** is the central API gateway that provides a RESTful HTTP interface for all data operations. It serves as the single entry point for external clients (UI, SDK, Forms) and manages all CRUD operations for users, jobs, tasks, keepers, and API keys.

**Location**: `internal/dbserver/`

**Note**: No DELETE operations are performed on the database, we only update the status of the data to deleted.

### Responsibilities

- **User Management**: Create, read, update user profiles
- **Job Management**: CRUD operations for all job types
- **Task Tracking**: Record and query task execution history
- **Keeper Registry**: Manage keeper registration and metadata
- **API Key Management**: Create, validate, and manage API keys
- **Authentication**: Validate API keys and enforce rate limits
- **Request Routing**: Forward validated requests to appropriate services
- **Trace Initialization**: Generate and propagate `X-Trace-ID` for distributed tracing

### API Endpoints

#### User Endpoints

- `POST /api/users` - Create new user
- `GET /api/users/address/:address` - Get user by wallet address
- `PUT /api/users/email` - Update user information
- `GET /api/users/address/:address/jobs` - List user's jobs

#### Job Endpoints

- `POST /api/jobs` - Create new job (all types)
- `GET /api/jobs/id/:id` - Get job details
- `PUT /api/jobs/update/:id` - Update job
- `DELETE /api/jobs/delete/:id` - Delete/cancel job
- `GET /api/jobs` - List jobs (with filters)

#### Task Endpoints

- `GET /api/tasks/id/:id` - Get task details
- `GET /api/tasks/job/:job_id` - List tasks for a job
- `GET /api/tasks` - List all tasks (paginated)

#### Keeper Endpoints

- `POST /api/keepers/form` - Register new keeper from Google Form

#### Leaderboard Endpoints

- `GET /api/leaderboard/keepers` - Get keeper leaderboard
- `GET /api/leaderboard/keepers/search` - Get keeper leaderboard by identifier
- `GET /api/leaderboard/users` - Get user leaderboard
- `GET /api/leaderboard/users/search` - Get user leaderboard by address

#### Fee Endpoints

- `GET /api/jobs/fees` - Get estimated fees for a job
- `POST /api/claim-fund` - Claim fund from faucet

#### WebSocket Endpoints

- `GET /api/ws/tasks` - Get WebSocket connection for tasks
- `GET /api/ws/stats` - Get WebSocket stats
- `GET /api/ws/health` - Get WebSocket health

#### API Key Endpoints (Admin Only)

- `POST /api/admin/api-keys` - Generate new API key
- `GET /api/admin/api-keys/:owner` - Get API key details by owner
- `PUT /api/admin/api-keys/:key` - Revoke API key

### Technology Stack

- **Framework**: Gin (Go web framework)
- **Database**: ScyllaDB with `gocqlx` ORM
- **Cache**: Redis for hot data
- **Validation**: `go-playground/validator`
- **Logging**: Zap (structured logging)
- **Metrics**: Prometheus client

---

## 2. Schedulers

TriggerX has **two specialized scheduler types**, each handling different trigger mechanisms. While time scheduler is quite simple, condition scheduler is more complex and requires more resources.

---

### Time Scheduler

Handles **time-based triggers** like cron expressions, specific timestamps, and recurring schedules.

**Location**: `internal/schedulers/time/`

#### Trigger Types

- **Cron Schedules**: Standard cron expressions (e.g., `0 0 * * *` for daily at midnight)
- **Specific Times**: One-time execution at a specific timestamp
- **Intervals**: Recurring execution every N seconds/minutes/hours
- **Timezone Support**: User-specified timezone conversion

#### Working

1. Each job is polled from the database (`time_job_data` table)
2. If `next_execution_timestamp` is within the next 40 seconds, it picks up the job and processes it
   a. Create a new task (`task_id`)
   b. Update the `next_execution_timestamp` to the next execution time
3. Pass the the tasks in batches to the task dispatcher

---

### Condition Scheduler

Monitors conditional triggers: **blockchain events** and/or **off-chain state conditions**. It triggers tasks when specific events are emitted by smart contracts or conditions are met (e.g., token balance > threshold).

**Location**: `internal/schedulers/condition/`

#### Trigger Types Supported

- **Contract Events**: Listen for specific events (e.g., `Transfer`, `Approval`)
- **Event Filters**: Filter by indexed parameters (e.g., only transfers to a specific address)
- **Multi-Chain**: Monitor events across multiple blockchain networks
- **Block Confirmations**: Wait for N confirmations before triggering
- **Balance Conditions**: Token or ETH balance comparisons
- **Contract State**: Monitor specific contract state variables
- **Price Oracles**: Trigger on price thresholds
- **Composite Conditions**: Multiple conditions with AND/OR logic

### Features

1. **Load Balancing**: Multiple scheduler instances can run simultaneously
2. **Redis Streams**: Running jobs are pushed to Redis streams for consumption
3. **Metrics Collection**: Prometheus metrics for monitoring
4. **Worker Pools**: Concurrent goroutines for parallel processing
5. **Failure Handling**: Retry logic with exponential backoff
6. **Graceful Shutdown**: Clean resource cleanup on termination

#### Working

1. Scheduler gets gRPC call from DBServer to schedule a new job
2. Scheduler initiates a worker to monitor the trigger
3. If the trigger is met, worker notifies the scheduler
4. Scheduler creates a new task (`task_id`) and passes it to the task dispatcher

#### Event Subscription Strategy

1. **WebSocket Connection**: Maintain persistent connection to RPC node
2. **Subscription**: Subscribe to filtered logs for target contract
3. **Buffering**: Queue events in memory for processing
4. **Batch Processing**: Process multiple events in parallel
5. **Checkpointing**: Persist last processed block for recovery

#### Condition Evaluation

1. **Polling**: Query blockchain at specified intervals
2. **State Extraction**: Read contract state or account balance
3. **Comparison**: Evaluate condition expression (>, <, ==, !=)
4. **Edge Detection**: Trigger only on state transitions (prevent duplicate triggers)

#### Supported Comparisons

- `>`, `<`, `>=`, `<=`, `==`, `!=` - Numeric comparisons

---

## 3. TaskDispatcher

The **TaskDispatcher** is responsible for assigning tasks to appropriate Keepers based on availability, performance, and load balancing strategies.

**Location**: `internal/taskdispatcher/`

### Responsibilities

- **Task Queue Addition**: Add tasks to the queue
- **Keeper Selection**: Choose Keeper for task execution based on list of online keepers from health service
- **Task Transmission**: Send task details to Keeper via P2P network
- **Assignment Tracking**: Record which Keeper is assigned to which task
- **Failure Handling**: Reassign tasks if Keeper fails to acknowledge

---

## 4. TaskMonitor

The **TaskMonitor** tracks the lifecycle of tasks after assignment, monitors execution status, and handles failures or timeouts.

**Location**: `internal/taskmonitor/`

### Responsibilities

- **Status Tracking**: Listen for events on Attestation Center contract for task submissions from aggregator node
- **Timeout Detection**: Identify tasks that exceed execution time limits
- **Failure Handling**: Trigger retries for failed tasks
- **Database Updates**: Persist task status and results in Redis Streams
- **Alert Generation**: Notify on critical failures

### Retry Logic

#### Retry Conditions (Only for failed tasks)

- Task failed with transient error (network, timeout)
- Keeper went offline during execution
- Execution timeout

#### Retry Strategy

1. **Exponential Backoff**: Delay between retries increases (1s, 2s, 4s, ...)
2. **Max Retries**: Default 3 attempts
3. **Different Keeper**: Retry on a different Keeper
4. **Permanent Failure**: After max retries, mark as permanently failed

---

## 5. Keeper

The **Keeper** is a decentralized node that executes assigned tasks in isolated Docker containers and validates tasks executed by peer Keepers.

**Location**: `internal/keeper/`

### Responsibilities

- **Task Reception**: Listen for task assignments from TaskDispatcher
- **Script Download**: Fetch task scripts from IPFS
- **Task Execution**: Run scripts in isolated Docker containers
- **Proof Generation**: Generate proof of execution
- **Result Submission**: Send execution results to P2P network
- **Task Validation**: Attest to tasks executed by other Keepers
- **Health Reporting**: Regular check-ins with Health service

### Task Execution Flow

1. Receive Task Assignment via P2P
2. Validate Task Signature and Parameters
3. Fetch Script CID from IPFS
4. Create Docker Container with:
   - Resource Limits (CPU, Memory)
   - Seccomp Profile (restricted syscalls)
   - Network Mode (limited or none)
   - Read-Only Filesystem
5. Inject Task Arguments
6. Execute Script
7. Capture stdout, stderr, exit code
8. Generate Proof of Execution
9. Store Results
10. Broadcast Results to P2P Network
11. Cleanup Container

### Validation (Attestation) Flow

1. Listen for Peer Task Executions on P2P
2. Re-execute Task with Same Inputs
3. Compare Outputs:
   - If Match → Sign Positive Attestation
   - If Mismatch → Sign Negative Attestation
4. Submit Attestation to Aggregator Network

---

## 6. Health

The **Health** service monitors the availability and performance of all Keeper nodes, alerting operators when Keepers go offline or experience issues.

**Location**: `internal/health/`

### Responsibilities

- **Keeper Heartbeat Monitoring**: Track regular check-ins from Keepers
- **Offline Detection**: Identify Keepers offline for >10 minutes
- **Alert Generation**: Send notifications via Telegram or Email
- **Uptime Tracking**: Calculate and store Keeper uptime statistics
- **Service Health**: Monitor health of other backend services

---

## 7. Aggregator

The **Aggregator** is built on the Othentic Network and serves as the consensus layer for TriggerX. It collects executed tasks, aggregates attestations, achieves consensus, and submits validated tasks to the blockchain.

**Location**: `othentic/`

### Responsibilities

- **Task Collection**: Gather executed tasks from Keepers via P2P network
- **Attestation Aggregation**: Collect validation signatures from attester Keepers
- **Consensus**: Achieve Byzantine Fault Tolerant (BFT) consensus on task validity
- **Blockchain Submission**: Submit validated tasks to L2 smart contracts
- **P2P Bootstrap**: Act as entry point for Keeper network discovery

### Consensus Mechanism

#### BFT Consensus Flow

1. **Task Submission**: Performer Keeper submits executed task result
2. **Attestation Request**: Aggregator requests attestations from N random attesters
3. **Attestation Collection**: Attesters validate and sign (agree/disagree)
4. **Threshold Check**: Require 1/3+ attesters to disagree (Byzantine Fault Tolerant)
5. **Consensus Decision**:
   - If 2/3+ agree → Task is valid
   - If 1/3+ disagree → Task is invalid (Performer may be slashed)

### Configuration

The Aggregator is configured via `.othentic/` directory:

- **Network Config**: Chain IDs, RPC endpoints, contract addresses
- **P2P Config**: Listen addresses, bootstrap peers
- **Keys**: Aggregator private keys for signing

---

## Service Communication Matrix

Coming soon...

---

For detailed data flow, see [6_dataflow.md](./6_dataflow.md).
