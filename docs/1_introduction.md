# Introduction to TriggerX Backend

## What is TriggerX?

TriggerX is a decentralized task execution and automation platform built for blockchain ecosystems. It enables developers to schedule, execute, and validate tasks across multiple blockchain networks in a secure, reliable, and fault-tolerant manner.

The TriggerX Backend serves as the core infrastructure that orchestrates task scheduling, execution, and validation through a network of decentralized **Keepers** (node operators) coordinated by intelligent **Schedulers** and validated by an **Aggregator** network.

## Why TriggerX?

Traditional blockchain automation solutions face several challenges:

- **Limited Cross-Chain Support**: Most solutions are confined to a single blockchain
- **Centralized Points of Failure**: Relying on centralized servers compromises decentralization
- **High Operational Costs**: Inefficient resource allocation leads to expensive execution
- **Complex Integration**: Difficult for developers to integrate automation into their dApps
- **Lack of Transparency**: Limited visibility into task execution and validation

TriggerX addresses these challenges by providing:

- **Multi-Chain Support**: Execute tasks across multiple blockchain networks seamlessly  
- **Decentralized Architecture**: No single point of failure with distributed Keeper network  
- **Cost-Efficient**: Intelligent load balancing and resource optimization  
- **Developer-Friendly**: Simple SDK and API for easy integration  
- **Full Transparency**: Complete audit trail with OpenTelemetry tracing  
- **Secure Execution**: Isolated Docker containers with seccomp profiles  

## Core Concepts

### Jobs and Tasks

- **Job**: A user-defined automation workflow with triggers and actions
- **Task**: A single execution instance of a job, performed by a Keeper
- **Trigger**: The condition that initiates job execution (time-based, event-based, or condition-based)
- **Action**: The operation to be performed when a trigger fires (smart contract call, API request, etc.)

### Roles in the System

1. **Users/Developers**: Create and manage jobs through the SDK or web interface
2. **Keepers**: Decentralized node operators who execute and validate tasks
3. **Schedulers**: Backend services that monitor triggers and assign tasks to Keepers
4. **Aggregator**: Consensus layer that validates task execution and submits results to the blockchain

### Task Lifecycle

```bash
Job Creation → Trigger Monitoring → Task Assignment → Execution → Validation → Consensus → Blockchain Submission -> Rewards/Slashing
```

## Architecture at a Glance

TriggerX Backend is built with a microservices architecture consisting of:

- **DBServer**: Centralized API for data persistence (users, jobs, tasks, keepers) for SDK
- **Schedulers**: Two specialized scheduler types for different trigger mechanisms
  - Time Scheduler (cron-like time-based triggers, where we know the exact time of the trigger)
  - Condition Scheduler (blockchain event triggers, on-chain or off-chain state condition triggers, where we don't know the exact time of the trigger)
- **TaskDispatcher**: Assigns tasks to appropriate Keepers
- **TaskMonitor**: Tracks task execution status and handles failures
- **Health Service**: Monitors Keeper availability and alerts operators
- **Keeper Nodes**: Decentralized executors that perform tasks in isolated environments, or validate them when not performing tasks
- **Aggregator Network**: Othentic stack for consensus layer for task validation

All services communicate using **gRPC** with **OpenTelemetry** tracing for end-to-end observability.

## Technology Stack

TriggerX leverages cutting-edge technologies:

- **Go 1.25+**: High-performance backend services
- **ScyllaDB**: Low-latency, highly scalable database
- **Redis**: Fast caching and job streaming
- **Docker**: Secure, isolated task execution environments
- **gRPC**: Efficient inter-service communication
- **OpenTelemetry**: Distributed tracing and metrics
- **Prometheus & Grafana**: Monitoring and alerting
- **Othentic Network**: Decentralized consensus and validation
- **EigenLayer**: Re-staking infrastructure for economic security
- **IPFS**: Decentralized storage for task scripts and results

## Key Features

### For Developers

- **Simple SDK**: Easy-to-use TypeScript SDK for job creation
- **Flexible Triggers**: Time-based, event-based, and condition-based triggers
- **Multi-Chain**: Support for Arbitrum One; and Ethereum, Base, Optimism and Arbitrum testnets are suported (Sepolia)
- **Dynamic Arguments**: Pass dynamic data to your smart contract functions
- **Job Chaining**: Link multiple jobs for complex workflows
- **Real-Time Monitoring**: Track job execution status in real-time

### For Keeper Operators

- **Easy Setup**: One-line Docker installation
- **Automated Operations**: Self-managing node with health monitoring
- **Competitive Rewards**: Earn rewards for task execution and validation
- **Configurable Resources**: Adjust CPU/memory allocation based on your hardware
- **Community Support**: Active Telegram community and comprehensive documentation

### For the Ecosystem

- **Economic Security**: Built on EigenLayer with re-staked ETH
- **Transparent Execution**: All task executions are traceable and auditable
- **Fault Tolerance**: Automatic failover and retry mechanisms
- **Scalability**: Horizontal scaling with multiple microservices and keeper instances
- **Open Source**: MIT licensed

## Getting Started

Ready to start building with TriggerX? Here's your next steps:

1. **Read the [Architecture](./2_architecture.md)** → Understand the system design
2. **Explore the [Backbones](./3_backbones.md)** → Learn about the technology stack
3. **Study the [Services](./4_services.md)** → Deep dive into each service
4. **Review the [Data Types](./5_datatypes.md)** → Understand data structures
5. **Follow the [Data Flow](./6_dataflow.md)** → Trace how data moves through the system
6. **Run the [Tests](./7_tests.md)** → Learn about testing strategies

### Quick Links

- **Main README**: [../README.md](../README.md) - Installation and setup instructions
- **Developer Notes**: [devNotes.md](./devNotes.md) - Internal architecture notes
- **Diagrams**: [diagrams.md](./diagrams.md) - Visual architecture representations
- **ToDo**: [ToDo](./todo.md) list, which we are working on, and you can help us with

## Community & Support

- **GitHub**: Check out our [repositories](https://github.com/orgs/trigg3rX/repositories)
- **Telegram**: [Join our community](https://t.me/triggerxnetwork) for support and discussions
- **Documentation**: Read our [docs](https://triggerx.gitbook.io/triggerx-docs)
- **Contributing**: We welcome contributions! Check out our [contribution guidelines](../.github/CONTRIBUTING.md)

---
