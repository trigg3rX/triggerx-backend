# Technology Backbones

TriggerX Backend is built on a foundation of industry-leading technologies, each selected for specific capabilities that enable decentralized, scalable, and fault-tolerant task execution. This document details the core technologies that form the backbone of the platform.

---

## 1. EigenCloud

### What is EigenCloud?

EigenCloud (previously known as EigenLayer) is a protocol that enables **restaking** of ETH to secure multiple services simultaneously. It allows Ethereum validators to opt-in to validating additional protocols (called Actively Validated Services or AVS) while earning additional rewards.

### EigenCloud in TriggerX

TriggerX leverages EigenCloud for:

- **Economic Security**: Keepers and validators are backed by restaked ETH, providing strong economic guarantees
- **Slashing Mechanism**: Malicious or faulty behavior can result in slashing of staked assets
- **Validator Network**: Access to a large pool of professional validators
- **Trustless Execution**: Cryptoeconomic incentives ensure honest behavior

### Integration Points

- **Keeper Registration**: Keepers register through EigenCloud with staked ETH
- **Reward Distribution**: Execution rewards distributed through EigenCloud contracts
- **Slashing Conditions**: Invalid task execution or attestation results in slashing
- **Delegation**: Users can delegate their stake to professional Keeper operators

### Benefits

✅ Shared security with Ethereum  
✅ Lower bootstrapping costs for new services  
✅ Access to professional validator infrastructure  
✅ Proven cryptoeconomic security model  

**Learn More**: [EigenCloud Documentation](https://docs.eigencloud.xyz/)

---

## 2. Othentic

### What is Othentic?

Othentic is a modular AVS (Actively Validated Service) framework built on EigenCloud. It provides infrastructure for building decentralized validation networks with built-in consensus, P2P networking, and aggregation mechanisms.

### Othentic in TriggerX

Othentic powers the **Aggregator Network**, TriggerX's consensus layer:

- **Task Aggregation**: Collects executed tasks from Keepers
- **Consensus Mechanism**: Byzantine Fault Tolerant (BFT) consensus on task validity
- **Blockchain Submission**: Submits validated tasks to L2 chains
- **P2P Networking**: Bootstrap nodes for Keeper network discovery

### Components Used

1. **Aggregator Nodes**: Collect and validate task submissions
2. **Attester Nodes**: Keepers that validate peer executions
3. **CLI**: Command line interface for Keeper Registration

### Configuration

Othentic is configured via `./othentic` directory with:

- Network configurations
- Aggregator endpoints
- Cryptographic keys

**Learn More**: [Othentic Documentation](https://docs.othentic.xyz/)

---

## 3. ScyllaDB

### What is ScyllaDB?

ScyllaDB is a high-performance NoSQL database compatible with Apache Cassandra, written in C++ for maximum efficiency. It's designed for low-latency, high-throughput workloads at scale.

### Role in TriggerX

ScyllaDB serves as the **primary persistent data store** for all TriggerX data:

- **User Data**: User profiles, API keys, authentication
- **Job Data**: Job definitions, status, metadata
- **Task Data**: Task execution records, results, logs
- **Keeper Data**: Keeper registry, performance metrics, reputation

### Why ScyllaDB?

#### Performance Characteristics

- **Low Latency**: <1ms p50, <10ms p99 read/write operations
- **High Throughput**: 1M+ ops/sec per node
- **Predictable Performance**: No garbage collection pauses (C++ vs Java)
- **Efficient Resource Usage**: Better CPU and memory utilization than Cassandra

#### Scalability

- **Horizontal Scaling**: Add nodes without downtime
- **No Single Point of Failure**: Fully distributed architecture
- **Automatic Sharding**: Data distributed across cluster
- **Tunable Consistency**: Choose between consistency and availability (CAP theorem)

### Schema Design

TriggerX uses ScyllaDB with the following design principles:

#### Partition Keys

Optimized for query patterns:

- `user_address` for user data (single-user queries)
- `job_id` for job data (per-job queries)
- `task_id` for task data (task lookup)
- `keeper_address` for keeper data (keeper queries)

#### Clustering Keys

For ordered data retrieval:

- `created_at` for time-ordered job listings
- `task_id` for task execution sequences

#### Materialized Views

For alternate query patterns:

- Jobs by user
- Tasks by job
- Keeper performance rankings

**Learn More**: [ScyllaDB Documentation](https://docs.scylladb.com/)

---

## 4. Redis

### What is Redis?

Redis (Remote Dictionary Server) is an in-memory data structure store used as a database, cache, message broker, and streaming engine.

### Redis in TriggerX

We use Upstash Redis as our Redis provider. It provides **high-speed caching and messaging** across TriggerX services:

#### Use Cases

1. **Job Streaming**: Redis Streams for failover and recovery
2. **Caching Layer**: Task execution results and keeper performance metrics
3. **Distributed Locking**: Coordination between multiple scheduler instances
4. **Rate Limiting**: API key request counters
5. **Session Storage**: Temporary state management
6. **Pub/Sub**: Real-time event notifications to SDKs

### Redis Data Structures Used

#### Streams (Job Queues)

```bash
time_scheduler_stream → [task1, task2, task3, ...]
condition_scheduler_stream → [task1, task2, ...]
```

**Consumer Groups**: Multiple scheduler instances consume from the same stream

#### Hashes (Cached Data)

```bash
job:{job_id} → {status, user_address, created_at, ...}
keeper:{keeper_address} → {online, uptime, tasks_executed, ...}
```

#### Sorted Sets (Priority Queues)

```bash
pending_tasks → [(priority1, task1), (priority2, task2), ...]
keeper_rankings → [(score1, keeper1), (score2, keeper2), ...]
```

#### Strings (Counters & Locks)

```bash
rate_limit:{api_key} → request_count
lock:scheduler:{job_id} → locked_by_instance_id
```

**Learn More**: [Upstash Redis Documentation](https://upstash.com/docs/introduction)

---

## 5. Pinata

### What is IPFS?

Pinata is a decentralized storage provider for storing and sharing data in a distributed file system. It uses content-addressing (CIDs) instead of location-addressing (URLs).

### Pinata in TriggerX

We use Pinata as our IPFS provider. It provides **decentralized storage** for task execution artifacts:

#### Stored Content

1. **Task Scripts**: JavaScript, Python, Shell scripts for execution
2. **Task Results**: Execution outputs
3. **Proof Artifacts**: Cryptographic proofs and attestations

### Why Pinata?

#### Decentralization

- No single point of failure
- Censorship-resistant
- Content remains available as long as at least one node pins it

#### Content Addressing

- CID (Content Identifier) based on file hash
- Immutable: Content hash ensures data integrity
- Deduplication: Same content = same CID (saved only once)

**Learn More**: [Pinata Documentation](https://docs.pinata.cloud/quickstart)

---

## Supporting Technologies

### Docker

**Purpose**: Container runtime for isolated task execution

**Features**:

- Isolated execution environments
- Resource limits (CPU, memory, network)
- Seccomp profiles for security
- Image caching for faster startup

**Usage**: Keepers run each task in an ephemeral Docker container

---

### gRPC (Google Remote Procedure Call)

**Purpose**: High-performance inter-service communication

**Features**:

- Protocol Buffers for efficient serialization
- HTTP/2 multiplexing
- Bi-directional streaming
- Built-in load balancing
- Language-agnostic

**Usage**: All internal service-to-service communication

---

### OpenTelemetry

**Purpose**: Distributed tracing and observability

**Features**:

- Standardized tracing API
- Context propagation across services
- Span creation and management
- Integration with Grafana Tempo

**Usage**: End-to-end request tracing with `X-Trace-ID`

---

### Prometheus & Grafana

**Purpose**: Metrics collection and visualization

**Features**:

- Time-series metric storage
- PromQL query language
- Pre-built dashboards
- Alerting on thresholds

**Usage**: Service health monitoring and performance tracking

---

## Technology Selection Rationale

| Requirement | Technology | Reason |
|------------|-----------|--------|
| Economic Security | EigenCloud | Proven restaking protocol with strong cryptoeconomic guarantees |
| Consensus & Validation | Othentic | Modular AVS framework with BFT consensus and proof generation |
| High-Performance Database | ScyllaDB | Low-latency, high-throughput, horizontally scalable NoSQL |
| Fast Caching & Messaging | Redis | In-memory speed, rich data structures, Streams for queues |
| Decentralized Storage | Pinata | Content-addressed, immutable, censorship-resistant storage |
| Service Communication | gRPC | High performance, type-safe, streaming support |
| Container Runtime | Docker | Industry standard, secure isolation, rich ecosystem |
| Observability | OpenTelemetry | Vendor-neutral, standardized tracing and metrics |

---

For details on how these technologies are used in each service, see [4_services.md](./4_services.md).
