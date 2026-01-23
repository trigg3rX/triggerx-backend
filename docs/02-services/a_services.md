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
5. [Eventmonitor](#5-eventmonitor)
6. [Keeper](#6-keeper)
7. [Health](#7-health)
8. [Aggregator](#8-aggregator)

---

## 1. DBServer

The **DBServer** is the central API gateway that provides a RESTful HTTP interface for all data operations. It serves as the single entry point for SDK and manages all CRUD operations for users, jobs, tasks, and API keys.

**Location**: `internal/dbserver/`

**Note**: No DELETE operations are performed on the database, we only update the status of the data to deleted.

### Responsibilities

- **User Management**: Create, read, update user profiles
- **Job Management**: CRUD operations for all job types
- **Task Tracking**: Query task execution history
- **Leaderboard**: Query leaderboard data for users and keepers
- **API Key Management**: Create, validate, and manage API keys
- **Authentication**: Validate API keys and enforce rate limits
- **Request Routing**: Forward validated requests to appropriate services

### API Endpoints

#### User Endpoints

- `GET /api/users/:address` - Get user by wallet address (protected)
- `POST /api/users/email` - Add email to user profile for email notifications (protected)
- `GET /api/users/safe-addresses/:user_address` - Get safe addresses for user (protected)

#### Job Endpoints

- `POST /api/jobs` - Create new job (all types, validation middleware)
- `PUT /api/jobs/update/:id` - Update job
- `PUT /api/jobs/delete/:id` - Delete/cancel job (protected)
- `GET /api/jobs/by-apikey` - List jobs by API key (protected)
- `GET /api/jobs/user/:user_address` - List user's jobs (protected)
- `GET /api/jobs/user/:user_address/chain/:created_chain_id` - List user's jobs by chain (protected)
- `GET /api/jobs/user/:user_address/:job_id` - Get job details for user (protected)
- `GET /api/jobs/safe-address/:safe_address` - Get jobs by safe address (protected)

#### Task Endpoints

- `GET /api/tasks/:id` - Get task details
- `GET /api/tasks/job/:job_id` - List tasks for a job
- `GET /api/tasks/recent` - Get recent tasks (protected)
- `GET /api/tasks/user/:user_address` - Get tasks by user address (protected)
- `GET /api/tasks/by-apikey/:api_key` - Get tasks by API key (protected)
- `GET /api/tasks/safe-address/:safe_address` - Get tasks by safe address (protected)

#### Leaderboard Endpoints

- `GET /api/leaderboard/keepers` - Get keeper leaderboard (protected)
- `GET /api/leaderboard/users` - Get user leaderboard (protected)

#### Fee Endpoints

- `GET /api/fees` - Get estimated fees for a job
- `POST /api/claim-fund` - Claim fund from faucet

#### Code Validation Endpoints

- `POST /api/code/validate` - Validate code executable (raw source)

#### WebSocket Endpoints

- `GET /api/ws/tasks` - Get WebSocket connection for tasks
- `GET /api/ws/stats` - Get WebSocket stats
- `GET /api/ws/health` - Get WebSocket health

#### API Key Endpoints (Admin Only)

- `POST /api/admin/api-keys` - Generate new API key (protected, validation middleware)
- `GET /api/admin/api-keys/:owner` - Get API key details by owner (protected)
- `PUT /api/admin/api-keys/:key` - Update API key (protected)
- `DELETE /api/admin/api-keys/:key` - Delete API key (protected)

### Technology Stack

- **Framework**: Gin (Go web framework)
- **Database**: ScyllaDB with `gocqlx` ORM
- **Cache**: Redis for hot data
- **Validation**: `go-playground/validator`
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

1. Polls database every 30 seconds (`time_job_data` table)
2. Fetches jobs where `next_execution_timestamp <= now() + 40s` AND `is_active = true` AND `expiration_time >= now()`
3. For each job:
   a. Create a new task (`task_id`) in database
   b. Add task_id to job's `task_ids` set
   c. Update the `next_execution_timestamp` to the next execution time
   d. For agent jobs (TDI 7), fetch script storage from database
4. Process tasks in batches (default batch size: 15)
5. Send batches to TaskDispatcher via gRPC (RPC method: `submit-task`)
6. Separate IMUA and non-IMUA tasks into different batches

---

### Condition Scheduler

Handles **condition-based triggers** (TDI 5, 6, 9) and **event-based triggers** (TDI 3, 4, 8). For condition-based jobs, it monitors off-chain state conditions. For event-based jobs, it delegates to EventMonitor service for blockchain event monitoring.

**Location**: `internal/schedulers/condition/`

#### Job Types Supported

**Condition-Based Jobs (TDI 5, 6, 9)**:
- **API Conditions**: Monitor API endpoints for value thresholds
- **Oracle Conditions**: Monitor oracle data feeds
- **WebSocket Conditions**: Monitor WebSocket streams for value changes
- **Static Conditions**: Test conditions with static values

**Event-Based Jobs (TDI 3, 4, 8)**:
- **Contract Events**: Listen for specific blockchain events (e.g., `Transfer`, `Approval`)
- **Event Filters**: Filter by indexed parameters (e.g., only transfers to a specific address)
- **Multi-Chain**: Monitor events across multiple blockchain networks
- Delegates actual event monitoring to EventMonitor service via gRPC

### Features

1. **Dual Job Type Support**: Handles both condition-based and event-based jobs
2. **Worker Management**: Manages condition workers and event job registrations
3. **EventMonitor Integration**: Delegates blockchain event monitoring to EventMonitor service
4. **Cooldown Period**: Prevents duplicate triggers with 30-second cooldown for recurring jobs
5. **Metrics Collection**: Prometheus metrics for monitoring
6. **Job Status Checking**: Background job to check for expired jobs and clean up
7. **Graceful Shutdown**: Clean resource cleanup, unregister event jobs from EventMonitor

#### Working

**For Condition-Based Jobs (TDI 5, 6, 9)**:
1. Scheduler gets gRPC call from DBServer to schedule a new job
2. Scheduler creates a condition worker to monitor the trigger (API/Oracle/WebSocket)
3. Worker polls every 1 second and checks condition
4. If condition is satisfied, worker notifies the scheduler
5. Scheduler creates a new task (`task_id`) and passes it to the task dispatcher via gRPC

**For Event-Based Jobs (TDI 3, 4, 8)**:
1. Scheduler gets gRPC call from DBServer to schedule a new job
2. Scheduler registers the event monitoring request with EventMonitor service via gRPC
3. EventMonitor service handles WebSocket subscriptions and event monitoring
4. When event matches filter, EventMonitor sends notification to scheduler via webhook
5. Scheduler creates a new task (`task_id`) and passes it to the task dispatcher via gRPC

#### Condition Evaluation (for Condition-Based Jobs)

1. **Polling**: Query API/Oracle/WebSocket endpoint every 1 second
2. **Value Extraction**: Extract value from response using JSON path (SelectedKeyRoute)
3. **Comparison**: Evaluate condition expression against limits:
   - `greater_than` - value > upper_limit
   - `less_than` - value < lower_limit
   - `between` - lower_limit < value < upper_limit
   - `equals` - value == upper_limit
   - `not_equals` - value != upper_limit
   - `greater_equal` - value >= upper_limit
   - `less_equal` - value <= upper_limit
4. **Edge Detection**: Trigger only on condition state transitions (prevent duplicate triggers)
5. **Value Sources**: Support for API, Oracle, Static, and WebSocket sources

#### Event Monitoring (for Event-Based Jobs)

1. **Registration**: Register event monitoring request with EventMonitor service
2. **Delegation**: EventMonitor handles WebSocket subscriptions to RPC nodes
3. **Notification**: EventMonitor sends webhook notification when event matches
4. **Task Creation**: Scheduler creates task upon receiving event notification

---

## 3. TaskDispatcher

The **TaskDispatcher** is responsible for assigning tasks to appropriate Keepers based on availability, performance, and load balancing strategies.

**Location**: `internal/taskdispatcher/`

### Responsibilities

- **Task Reception**: Receive tasks from schedulers via gRPC (RPC method: `submit-task`)
- **Keeper Selection**: Call Health Service to get performer keeper based on network
- **Task Signing**: Sign task data with manager signature for authentication
- **Task Transmission**: Send task details to Keeper via HTTP POST to `/p2p/message` endpoint
- **Assignment Tracking**: Record task dispatch in Redis Streams (`task:dispatched`)
- **Failure Handling**: Track dispatch failures and update task status

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

## 5. Eventmonitor

The **EventMonitor** service monitors blockchain events across multiple chains and triggers tasks when events match registered filters. It manages WebSocket subscriptions to RPC nodes and handles event processing.

**Location**: `internal/eventmonitor/`

### Responsibilities

- **Event Monitoring**: Monitor blockchain events via WebSocket subscriptions
- **Multi-Chain Support**: Monitor events across multiple blockchain networks
- **Event Registration**: Register/unregister event monitoring requests via gRPC
- **Worker Management**: Manage workers for each unique event subscription
- **Task Triggering**: Trigger task creation when events match filters
- **Permanent Polling**: Poll Base networks for attestation events
- **IPFS Integration**: Fetch task data from IPFS for consensus event reporting

### Working

1. Receives monitoring requests via gRPC (`register` method) from DBServer
2. Registers event subscription in registry
3. Creates worker for WebSocket subscription if not exists
4. Worker subscribes to blockchain events via `eth_subscribe("logs", ...)`
5. When event matches filter, worker triggers task creation
6. Reports consensus events to TaskMonitor service

### Features

- **WebSocket Subscriptions**: Persistent connections to RPC nodes
- **Event Filtering**: Filter by contract address and event signature
- **Registry Management**: Track active subscriptions
- **Worker Pool**: One worker per unique event subscription
- **Permanent Poller**: Background polling for Base network attestation events

---

## 6. Keeper

The **Keeper** is a decentralized node that executes assigned tasks in isolated Docker containers and validates tasks executed by peer Keepers.

**Location**: `internal/keeper/`

### Responsibilities

- **Task Reception**: Receive task assignments from TaskDispatcher via HTTP POST to `/p2p/message` endpoint
- **Task Execution**: Execute traditional jobs (TDI 1-6) and agent jobs (TDI 7-9) in isolated Docker containers
- **Proof Generation**: Generate cryptographic proof of execution (TLS certificate proof)
- **Result Submission**: Send execution results to Aggregator via P2P network
- **Task Validation (Attestation)**: Validate tasks executed by other Keepers via HTTP POST to `/task/validate`
- **Status Reporting**: Report execution status to TaskMonitor service
- **Health Reporting**: Regular check-ins with Health service (every 60 seconds)

### API Endpoints

- `POST /p2p/message` - Receive task execution requests (returns 202 Accepted for async processing)
- `POST /task/validate` - Validate task execution from peer keeper
- `POST /task/rebroadcast` - Rebroadcast task execution results
- `GET /status` - Health check endpoint
- `GET /metrics` - Prometheus metrics endpoint

### Key Features

- **Async Processing**: Tasks are processed asynchronously after returning 202 Accepted
- **Parallel Execution**: Multiple targets in a task are executed in parallel
- **Agent Job Support**: Executes agent scripts (TDI 7, 8, 9) that decide execution dynamically
- **TLS Proof Generation**: Creates tamper-proof execution evidence
- **Nonce Management**: Per-chain transaction nonce management
- **Rebroadcast Capability**: Can rebroadcast execution results if needed

For detailed architecture and implementation details, see [Keeper Service Architecture](./b_keeper.md).

---

## 7. Health

The **Health** service monitors the availability and performance of all Keeper nodes, alerting operators when Keepers go offline or experience issues. It also provides keeper selection for TaskDispatcher.

**Location**: `internal/health/`

### Responsibilities

- **Keeper Heartbeat Monitoring**: Track regular check-ins from Keepers (every 60 seconds)
- **Keeper Selection**: Provide performer keeper selection via RPC for TaskDispatcher
- **Offline Detection**: Identify Keepers offline for >10 minutes (background job)
- **Alert Generation**: Send notifications via Telegram or Email
- **Uptime Tracking**: Calculate and store Keeper uptime statistics
- **In-Memory State Management**: Maintain keeper state in memory for fast lookups

### API Endpoints

- `POST /health` - Keeper health check-in endpoint
- `GET /operators` - Get detailed keeper status
- `GET /status` - Service health check endpoint

### RPC Methods

- `get-performer` - Get a performer keeper based on network (called by TaskDispatcher)
- `health` - Health check RPC method

---

## 8. Aggregator

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

For detailed data flow, see [data-flow.md](./data-flow.md).
