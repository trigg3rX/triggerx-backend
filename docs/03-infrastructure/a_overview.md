# Infrastructure Overview

The TriggerX Backend infrastructure consists of several key components that work together to provide a decentralized, scalable, and fault-tolerant task execution platform.

## Infrastructure Components

### 1. EigenCloud

**Purpose**: Economic security and validator network  
**Location**: External service integration  
**Details**: [EigenCloud Documentation](./b_eigencloud.md)

EigenCloud provides the cryptoeconomic security layer for TriggerX. Keepers and validators stake ETH through EigenCloud, ensuring honest behavior through slashing mechanisms.

### 2. Redis (Upstash)

**Purpose**: Task lifecycle management, caching, and distributed coordination  
**Location**: External service (Upstash Redis)  
**Details**: [Redis Documentation](./c_redis.md)

Redis is used for:

- Task lifecycle streams (dispatched, executed, validated, completed, failed)
- Distributed locking between scheduler instances
- Rate limiting for API keys
- Service discovery and registry

### 3. Docker

**Purpose**: Sandboxed code execution environment  
**Location**: Local Docker daemon  
**Details**: [Docker Documentation](./d_docker.md)

Docker provides isolated execution environments for:

- Dynamic argument generation scripts
- Agent scripts (custom logic)
- Code validation

### 4. IPFS (Pinata)

**Purpose**: Distributed storage for scripts and execution proofs  
**Location**: External service (Pinata IPFS gateway)  
**Details**: [IPFS Documentation](./e_ipfs.md)

IPFS is used for:

- Storing agent scripts
- Storing execution proofs and task data
- Distributed content addressing

### 5. ScyllaDB

**Purpose**: Primary database for jobs, tasks, and keeper data  
**Location**: External database cluster  
**Details**: [ScyllaDB Documentation](./f_scylla.md)

ScyllaDB stores all persistent data including jobs, tasks, keeper information, and user data.

### 6. Observability Stack

**Purpose**: Monitoring, logging, and tracing  
**Location**: External services (OpenTelemetry, Prometheus)  
**Details**: [Observability Documentation](./g_observability.md)

Observability infrastructure includes:

- OpenTelemetry for distributed tracing
- Prometheus for metrics collection
- Structured logging

## Infrastructure Requirements

### Production Environment

- **ScyllaDB**: Managed cluster with replication factor 3+
- **Redis**: Upstash Redis (serverless or dedicated)
- **Docker**: Docker daemon with sufficient resources for container pools
- **IPFS**: Pinata account with JWT authentication
- **EigenCloud**: Registered AVS with validator network
- **Observability**: OpenTelemetry collector endpoint

### Development Environment

- **ScyllaDB**: Local instance or Docker container
- **Redis**: Local Redis instance or Upstash free tier
- **Docker**: Local Docker Desktop or Docker Engine
- **IPFS**: Pinata free tier or local IPFS node
- **EigenCloud**: Testnet deployment
- **Observability**: Local OpenTelemetry collector

## Infrastructure Configuration

All infrastructure components are configured via environment variables and YAML configuration files:

- **Database**: `config/services/dbserver.yaml`
- **Redis**: Environment variables (`UPSTASH_REDIS_URL`, `UPSTASH_REDIS_REST_TOKEN`)
- **Docker**: `config/services/docker-executor.yaml`
- **IPFS**: Environment variables (`IPFS_HOST`, `PINATA_JWT`)
- **EigenCloud**: Environment variables for network and contract addresses
- **Observability**: Environment variables (`OTEL_EXPORTER_ENDPOINT`)

## High Availability

### Redis

- Upstash provides automatic failover and high availability
- Streams are replicated across Redis cluster
- Consumer groups ensure message delivery even with instance failures

### Docker

- Container pools provide redundancy
- Failed containers are automatically replaced
- Health checks ensure container availability

### IPFS

- Content is distributed across IPFS network
- Pinata provides reliable gateway access
- Multiple gateway fallbacks (Pinata, ipfs.io)

### ScyllaDB

- Multi-node cluster with replication
- Automatic failover and load balancing
- Consistent hashing for data distribution

## Security Considerations

1. **Network Isolation**: Docker containers run in isolated networks
2. **Resource Limits**: CPU and memory limits prevent resource exhaustion
3. **Access Control**: All external services require authentication
4. **Encryption**: TLS for all external communications
5. **Key Management**: Private keys stored securely, never in code
