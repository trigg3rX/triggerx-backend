# Introduction to TriggerX Backend

## What is TriggerX?

TriggerX is a decentralized automation platform built for blockchain ecosystems. It enables developers to schedule, execute, and validate tasks across multiple blockchain networks in a secure, reliable, and fault-tolerant manner.

The TriggerX Backend serves as the core infrastructure that orchestrates task scheduling, execution, and validation through a network of decentralized **Keepers** (node operators) coordinated by **Schedulers** and **Monitors**.

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
4. **Aggregator**: Node that orchestrates the consensus layer that validates task execution and submits results to the blockchain

### Task Lifecycle

```bash
Job Creation → Trigger Monitoring → Task Assignment → Execution → Validation → Consensus → Blockchain Submission -> Rewards/Slashing
```

## Architecture at a Glance

TriggerX Backend is built with a microservices architecture. For detailed architecture information, see [System Architecture](./01-architecture/a_system_overview.md).

All services communicate using **gRPC** with **OpenTelemetry** tracing for end-to-end observability.

## Technology Stack

TriggerX leverages cutting-edge technologies. For detailed technology information, see [Technology Stack](./01-architecture/b_technology_stack.md).

## Key Features

### For Developers

- [x] **Simple SDK**: Easy-to-use TypeScript SDK for job creation, and task monitoring
- [x] **Flexible Triggers**: Time-based, event-based, and condition-based triggers with custom script support
- [x] **Multi-Chain**: Support for Arbitrum One; and Ethereum, Base, Optimism and Arbitrum Sepolia testnets are suported
- [x] **Dynamic Arguments**: Pass dynamic data to your smart contract functions with custom script support
- [x] **Job Chaining**: Link multiple jobs for complex workflows
- [x] **Real-Time Monitoring**: Track job execution status on dashboard
  - [ ] Real-time using webhooks (WIP)

### For Keeper Operators

- [x] **Easy Setup**: One script to rule them all (just run `./triggerx.sh start`)
- [x] **Automated Operations**: Self-managing node with health monitoring
- [ ] **Competitive Rewards**: Earn rewards for task execution and validation (current rewards are from EigenCloud Programmatic Incentive program)
- [ ] **Configurable Resources**: Adjust CPU/memory allocation based on your hardware resources
- [x] **Community Support**: Active Telegram community and comprehensive documentation

### For the Ecosystem

- [x] **Economic Security**: Built on EigenCloud with re-staked ETH
- [x] **Transparent Execution**: All task executions are traceable and auditable
- [x] **Fault Tolerance**: Automatic failover and retry mechanisms
- [x] **Scalability**: Horizontal scaling with multiple microservices and keeper instances
- [x] **Open Source**: MIT licensed

## Documentation Structure

### [01-architecture/](./01-architecture/)

- System overview and design
- Technology stack
- Data types and structures
- Data flow through the system

### [02-services/](./02-services/)

- Individual service documentation
- Service architecture
- API endpoints
- Configuration options

### [03-infrastructure/](./03-infrastructure/)

- Infrastructure components overview
- Database setup and configuration
- Storage configuration
- Observability (metrics, logging, tracing)

### [04-development/](./04-development/)

- Development setup guides
- Development best practices
- Testing strategies
- Tracing implementation
- Design documents

### [05-features/](./05-features/)

- Platform features
- Rewards system
