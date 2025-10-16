# Architecture

## System Overview

TriggerX Backend follows a **microservices architecture** with clear separation of concerns, enabling horizontal scalability, fault tolerance, and independent service deployment. The system is designed around the principle of decentralized task execution with centralized coordination.

## High-Level Architecture

```sh
┌──────────────────────────────────────────────────────────────┐
│                          User Layer                          │
│  ┌───────────────────┐                ┌───────────────────┐  │
│  │ Web UI (Uses SDK) │                │        SDK        │  │
│  └──────────┬────────┘                └─────────┬─────────┘  │
└─────────────┼───────────────────────────────────┼────────────┘
              │                                   │
              └───────────────────────────────────┘
                             │ HTTP/REST API
                             ▼
┌──────────────────────────────────────────────────────────────┐
│                     API Gateway Layer                        │
│  ┌──────────────────────────────────────────────────────────┐│
│  │                DBServer (API Server)                     ││
│  │  • User Management  • Job CRUD  • API Key Auth           ││
│  │  • Task Tracking    • Keeper Registry                    ││
│  └──────────────────────────────────────────────────────────┘│
└──────────────────────────────────────────────────────────────┘
                             │
                             │ gRPC with OpenTelemetry
                             ▼
┌──────────────────────────────────────────────────────────────┐
│                      Scheduling Layer                        │
│  ┌──────────────┐     ┌──────────────┐     ┌──────────────┐  │
│  │    Time      │     │    Event     │     │  Condition   │  │
│  │  Scheduler   │     │  Scheduler   │     │  Scheduler   │  │
│  └──────┬───────┘     └──────┬───────┘     └──────┬───────┘  │
│         │                    │                    │          │
│         └────────────────────┼────────────────────┘          │
│                              │                               │
│         ┌────────────────────┴────────────────────┐          │
│         ▼                                         ▼          │
│  ┌──────────────┐                          ┌──────────────┐  │
│  │     Task     │     ┌──────────────┐     │     Task     │  │
│  │  Dispatcher  │ ◄─► │    Redis     │ ◄─► │    Monitor   │  │
│  └──────────────┘     └──────────────┘     └──────────────┘  │
└──────────────────────────────────────────────────────────────┘
            │ ▲                                    ▲
    Task    │ │       ┌──────────┐                 │                              
 Assignment │ └─────► │  Health  │                 │
            │         └──────────┘                 │
            │               ▲                      │ Status
            │               │ Health Check-ins     │ Updates
            ▼               ▼                      │
┌──────────────────────────────────────────────────────────────┐
│                       Execution Layer                        │
│     • Task Execution by assigned  • Proof Generation         │
│     • Task Validation by rest of peers                       │
│  ┌──────────────┐  ┌──────────────┐        ┌──────────────┐  │
│  │   Keeper 1   │  │   Keeper 2   │  ...   │   Keeper N   │  │
│  │              │  │              │        │              │  │
│  │ ┌──────────┐ │  │ ┌──────────┐ │        │ ┌──────────┐ │  │
│  │ │  Docker  │ │  │ │  Docker  │ │        │ │  Docker  │ │  │
│  │ │Container │ │  │ │Container │ │        │ │Container │ │  │
│  │ └──────────┘ │  │ └──────────┘ │        │ └──────────┘ │  │
│  │ ┌──────────┐ │  │ ┌──────────┐ │        │ ┌──────────┐ │  │
│  │ │ Othentic │ │  │ │ Othentic │ │        │ │ Othentic │ │  │
│  │ │ Attester │ │  │ │ Attester │ │        │ │ Attester │ │  │
│  │ └──────────┘ │  │ └──────────┘ │        │ └──────────┘ │  │
│  └──────┬───────┘  └──────┬───────┘        └──────┬───────┘  │
└─────────┼─────────────────┼───────────────────────┼──────────┘
          │                 │                       │
          └─────────────────┼───────────────────────┘
                            │ P2P Network
                            ▼
┌──────────────────────────────────────────────────────────────┐
│                     Consensus Layer                          │
│  ┌──────────────────────────────────────────────────────────┐│
│  │              Othentic Aggregator Network                 ││
│  │  • Operator BLS Signature Aggregation  • Consensus       ││
│  └──────────────────────────────────────────────────────────┘│
└──────────────────────────────────────────────────────────────┘
                             │
                             │ L2 Transactions
                             ▼
┌──────────────────────────────────────────────────────────────┐
│                     Blockchain Layer                         │
│  ┌──────────────────────────────────────────────────────────┐│
│  │              AVS Task Attestation Center                 ││
│  │  • BLS Signature Verification  • Onchain storage         ││
│  └──────────────────────────────────────────────────────────┘│
└──────────────────────────────────────────────────────────────┘
```

## Architectural Layers

### 1. User Layer

The user-facing interfaces that allow interaction with the TriggerX platform:

- **SDK**: TypeScript SDK for programmatic job management and monitoring
- **Web UI**: Next.js-based frontend using SDK for non-technical users with templates

### 2. API Gateway Layer (DBServer)

The central API server that handles all external requests and data persistence:

- **Authentication**: API key validation using JWT tokens
- **Data Management**: CRUD operations for users, jobs, tasks, and keepers
- **Request Routing**: Forwards validated requests to appropriate services
- **Trace ID Management**: Initiates and propagates X-Trace-ID for distributed tracing

**Technology**: Go with Gin framework, ScyllaDB (gocql) for persistence

### 3. Scheduling Layer

Two specialized scheduler types handle different trigger mechanisms:

#### Time Scheduler

- **Purpose**: Handles cron-like time-based job execution
- **Triggers**: Specific times, intervals, recurring schedules
- **Features**: Polls jobs, and passes them in batches

#### Condition Scheduler

- **Purpose**: Monitors unpredictable state conditions and 'triggers' upon conditions being met
- **Workers**: Task specific workers for each condition (Blockchain, API, WebSocket)
- **Triggers**: Smart contract events, transaction confirmations, Token balances, contract states, oracle data
- **Features**: On-chain event monitoring, off-chain polling-based monitoring, state comparison, threshold detection

Each condition scheduler instance include:

- **Worker Pools**: Concurrent trigger monitoring
- **Load Balancing**: Schedules upto a set number of workers to run in parallel
- **Redis Job Streams**: Scheduled jobs are picked/assigned to other services in case of failure

**Technology**: Go with gRPC, Redis Streams, OpenTelemetry

### 4. Task Management Layer

#### TaskDispatcher

- Receives tasks from schedulers
- Dynamic selection of Keepers based availability
- Assigns performer role to selected Keeper
- Sends task execution requests via P2P network (aggregator node)

#### TaskMonitor

- Listens for events on Attestation Center contract for task submissions from aggregator node
- Tracks task execution status in Redis Streams
- Updates DB with execution results, and keeper points
- Handles task failures and retries
- Triggers alerts for critical failures

**Technology**: Go with gRPC, OpenTelemetry, Upstash Redis Go SDK

### 5. Execution Layer (Keepers)

Decentralized nodes that execute and validate tasks:

#### Keeper Node Components

1. **API Server**: Receives task assignments from accompanying Attester Node (which is connected to the P2P network)
2. **Core Execution Engine**:
   - Downloads task scripts from IPFS
   - Validates task parameters
   - Spins up isolated Docker containers
   - Executes tasks with resource limits (CPU, memory, network)
3. **Validation Engine**: Attests to tasks executed by peer Keepers
4. **Health Check Service**: Regular heartbeat to Health Monitor

**Technology**: Go (Docker API, go-ethereum)

### 6. Consensus Layer (Aggregator)

Built by Othentic Team:

#### Aggregator Roles

1. **Task Collection**: Gathers executed tasks from Keepers
2. **Attestation Aggregation**: Collects validation signatures from attester Keepers
3. **Consensus Mechanism**: Achieves agreement on task validity (BFT consensus)
4. **Blockchain Submission**: Submits validated tasks to L2 blockchain
5. **P2P Bootstrap**: Acts as entry point for Keeper network discovery

### 7. Blockchain Layer

Multi-chain support for task validation:

- **Layer 1 (Ethereum)**: Keeper registration, staking, EigenLayer integration
- **Layer 2 (Base)**: Task validation, result recording
- **Smart Contracts** (On each chain TriggerX is funcitonal):
  - TriggerX Job Registry: Job data storage
  - Trigger Gas Registry: Track User's Trigger Gas balance
  - Task Execution Hub: Proxy contract for task execution (user's can allow this contract to execute tasks on their behalf)

## Communication Protocols

Coming soon...

## Data Storage Architecture

### Primary Database: ScyllaDB

**Why ScyllaDB?**

- Low-latency read/write operations (<10ms p99)
- Horizontal scalability with no single point of failure
- High write throughput for task logging

**Tables**:

- `user_data`: User profiles and metadata
- `job_data`: Job definitions and status
- `time_job_data`, `event_job_data`, `condition_job_data`: Scheduler-specific data
- `task_data`: Task execution records
- `keeper_data`: Keeper registry and performance metrics
- `apikeys`: API key management

### Cache Layer: Redis

- **Job Streaming**: Redis Streams for task queue management
- **Session Cache**: Temporary data and state management
- **Rate Limiting**: Counter-based request throttling
- **Distributed Locks**: Coordination between scheduler instances
- **Task Streams**: Stream to monitor task execution status across all services

### File Storage: IPFS

**Stored Content**:

- Task execution scripts (JavaScript, Python, etc.)
- Task execution results
- Proof artifacts

**Benefits**:

- Decentralized storage
- Content-addressable (CID-based)
- Permanent data availability
- Reduced database load

## Observability Architecture

### Distributed Tracing

**OpenTelemetry Integration**:

- Trace ID format: `tgrx-frnt-<uuid>` (frontend) or `tgrx-sdk-<uuid>` (SDK)
- Span creation at every service boundary
- Context propagation via gRPC metadata
- Exported to Grafana Tempo

**Trace Lifecycle**:

1. User request generates trace ID
2. DBServer initiates root span
3. Each service creates child spans
4. Trace propagated through gRPC context
5. Keeper execution adds spans
6. Complete trace viewable in Grafana Tempo

### Metrics Collection

**Prometheus Metrics** (per service):

- **Common Metrics**: Uptime, CPU, memory, goroutines, GC duration
- **HTTP Metrics**: Request count, duration, RPS, status codes
- **Business Metrics**: Tasks executed, jobs scheduled, keeper performance
- **Custom Metrics**: Service-specific KPIs

**Collection Architecture**:

- Each service exposes `/metrics` endpoint
- Prometheus scrapes all services periodically
- Grafana visualizes metrics with pre-built dashboards
- Alertmanager triggers notifications on threshold breaches

### Logging

**Structured Logging** (Zap):

- JSON-formatted logs
- Log levels: DEBUG, INFO, WARN, ERROR, FATAL
- Contextual fields: trace_id, service, timestamp, function
- Log aggregation in Grafana Loki

### Health Monitoring

**Health Service**:

- Monitors Keeper heartbeats (1-minute window)
- Tracks service availability
- Sends alerts via Telegram/Email to operators

## Security Architecture

### Authentication & Authorization

1. **API Key Authentication**: Users authenticate with API keys
2. **JWT Tokens**: Service-to-service authentication with signed tokens
3. **Rate Limiting**: Per-key request throttling

### Network Security

1. **TLS/SSL**: All external communications encrypted
2. **gRPC Security**: mTLS for service-to-service communication
3. **Firewall Rules**: Restricted service access by IP/port

### Execution Security

1. **Container Isolation**: Tasks run in ephemeral Docker containers
2. **Seccomp Profiles**: System call restrictions
3. **Resource Quotas**: CPU/memory/network limits
4. **Network Policies**: Restricted internet access for containers
5. **Code Validation**: Script validation before execution

### Data Security

1. **Encryption at Rest**: ScyllaDB encryption
2. **Encryption in Transit**: TLS for all communications
3. **Secret Management**: Environment-based secrets, no hardcoding
4. **Access Control**: Role-based access to database and services

## Scalability Considerations

### Horizontal Scaling

- **Schedulers**: Multiple instances with Redis-based coordination
- **Keepers**: Add more nodes to increase execution capacity
- **DBServer**: Stateless design allows multiple replicas

### Load Balancing

- **API Gateway**: Nginx for DBServer load balancing
- **Task Distribution**: Round-robin Keeper selection

### Performance Optimizations

- **Connection Pooling**: Reuse database and Redis connections
- **Caching Strategy**: Redis for hot data, database for cold data
- **Asynchronous Processing**: Non-blocking task assignment and monitoring

## Fault Tolerance & Reliability

### Redundancy

- **Multiple Scheduler Instances**: Failover if one crashes
- **Keeper Redundancy**: Tasks can be reassigned on failure
- **Database Replication**: ScyllaDB multi-node cluster

### Retry Mechanisms

- **Task Execution Retry**: Automatic retry on Keeper failure
- **Network Retry**: Exponential backoff for transient failures
- **Database Retry**: Connection retry with circuit breaker

### Monitoring & Alerting

- **Service Health Checks**: Liveness and readiness probes
- **Keeper Monitoring**: 10-minute offline alert
- **Error Rate Tracking**: Alert on high error rates
- **Performance Degradation**: Alert on slow response times

## Deployment Architecture

### Development Environment

- Docker Compose for local development
- All services in single-host deployment
- In-memory Redis and single-node ScyllaDB

### Production Environment

- Docker Compose
- Multi-node ScyllaDB cluster
- Load-balanced API servers
- Distributed Keepers across multiple operators

---

For more details on each service, see [4_services.md](./4_services.md).
