# Keeper Service Architecture

## Service Overview

The Keeper service is a distributed task execution and validation node that operates as part of a decentralized network. It receives tasks from schedulers, validates triggers, executes actions, generates cryptographic proofs, and coordinates with validators through an aggregator service.

## Core Architecture Layers

### 1. Main Entry Point (`cmd/keeper/main.go`)

#### Initialization Sequence

- **Configuration Loading**: Initializes environment-based configuration via `config.Init()`
- **Logger Setup**: Creates structured logger with process-specific configuration
- **Dependency Injection**: Initializes all service dependencies in sequence:
  1. Metrics collector (`metrics.NewCollector()`)
  2. Aggregator client (`aggregator.NewAggregatorClient()`)
  3. Health client (`health.NewClient()`)
  4. Code executor (`docker.NewCodeExecutor()`)
  5. Task validator (`validation.NewTaskValidator()`)
  6. Task executor (`execution.NewTaskExecutor()`)
  7. API server (`api.NewServer()`)

#### Process Management

- **Health Check Routine**: Periodic 60-second health check-ins with health service
- **API Server**: HTTP server for task execution endpoints
- **Metrics Collection**: Prometheus metrics collection service
- **Graceful Shutdown**: Coordinated shutdown with 10-second timeout

### 2. Configuration Layer (`internal/keeper/config/`)

#### Configuration Management (`config.go`)

- **Environment Variables**: Loads configuration from `.env` file
- **Validation**: Validates all configuration parameters on startup to ensure the user has provided the correct values
- **Cryptographic Keys**: Manages consensus and controller private keys (controller key will be dropped in the future, only controller address is needed)

#### Registration Validation (`registered.go`)

- **Keeper Registration Check**: Validates keeper address registration on L2 chain
- **Startup Validation**: Prevents service startup if keeper is not registered

### 3. API Layer (`internal/keeper/api/`)

#### HTTP Server (`server.go`)

- **Gin Framework**: Uses Gin HTTP framework with release mode configuration
- **Server Configuration**: Configurable timeouts, port, and header limits
- **Middleware Stack**:
  - Recovery middleware for panic handling
  - Trace middleware for request tracing
  - Logger middleware for request logging
  - Error middleware for error handling
- **Route Configuration**:
  - `POST /p2p/message` - Task execution endpoint
  - `POST /task/validate` - Task validation endpoint
  - `GET /metrics` - Prometheus metrics endpoint

#### Request Handlers (`handlers/`)

- **Task Handler** (`handler.go`, `execute.go`, `validate.go`):
  - `ExecuteTask()`: Processes incoming task execution requests
  - `ValidateTask()`: Validates task data and proofs
- **Metrics Handler** (`metrics.go`):
  - `Metrics()`: Serves Prometheus metrics data

#### Middleware (`middleware.go`)

- **TraceMiddleware()**: Generates and propagates request trace IDs, this trace would be Attester Node -> core services ->response to attester. This different from what Scheduler -> (execution and validation) -> registrar -> DB Server will be implemented in future. This would be keeper specific (local).
- **LoggerMiddleware()**: Logs HTTP requests with structured logging
- **ErrorMiddleware()**: Handles and formats error responses

### 4. Core Business Logic (`internal/keeper/core/`)

#### Task Execution (`execution/executor.go`)

- **TaskExecutor Structure**:
  - Alchemy API integration for blockchain data
  - Docker code executor for sandboxed execution
  - Argument converter for dynamic parameter handling
  - Task validator for pre-execution validation
  - Aggregator client for result submission

- **ExecuteTask() Function Flow**:
  1. **Scheduler Signature Validation**: Verifies task authenticity
  2. **Trigger Validation**: Validates trigger conditions
  3. **Blockchain Client Setup**: Establishes RPC connection to target chain
  4. **Action Execution**: Executes task based on TaskDefinitionID:
     - Static args execution (IDs: 1, 3, 5)
     - Dynamic args execution (IDs: 2, 4, 6)
  5. **Proof Generation**: Creates TLS-based cryptographic proof
  6. **Data Signing**: Signs IPFS data with consensus private key
  7. **IPFS Upload**: Uploads proof data to IPFS network
  8. **Aggregator Submission**: Sends results to validator network

#### Task Validation (`validation/validator.go`)

- **TaskValidator Structure**:
  - Alchemy and Etherscan API integration
  - Docker code executor for validation
  - Aggregator client for network coordination

- **ValidateTask() Function Flow**:
  1. **Scheduler Signature Validation**: Verifies task origin
  2. **Blockchain Connection**: Establishes RPC connection
  3. **Trigger Validation**: Validates trigger conditions
  4. **Action Validation**: Validates executed actions
  5. **Proof Validation**: Verifies cryptographic proofs
  6. **Performer Signature Validation**: Validates executor signatures

### 5. Client Layer (`internal/keeper/client/`)

#### Health Client (`health/health.go`)

- **Health Monitoring**:
  - Periodic health check-ins with health service
  - Cryptographic signature-based authentication
  - Encrypted response handling
  - Keeper verification status monitoring

- **Client Configuration**:
  - Health service URL configuration
  - Private key for authentication
  - Keeper address and peer ID
  - Request timeout configuration

- **CheckIn() Function**:
  1. **Key Derivation**: Derives consensus address from private key
  2. **Message Signing**: Signs keeper address for authentication
  3. **Payload Creation**: Creates health check payload with timestamp
  4. **HTTP Request**: Sends POST request to health service
  5. **Response Processing**: Handles encrypted response and error codes
  6. **Verification Handling**: Manages keeper verification status

### 6. Utilities Layer (`internal/keeper/utils/`)

- **Chain RPC Management**: Provides RPC URL resolution for different chains
- **IPFS Integration**: Handles file uploads to IPFS network

### 7. Metrics Layer (`internal/keeper/metrics/`)

#### Metrics Collector (`collector.go`)

- **Prometheus Integration**: Uses Prometheus client for metrics collection
- **HTTP Handler**: Provides `/metrics` endpoint for scraping
- **Collection Management**: Starts and manages metrics collection processes

## Data Flow Architecture

### Task Execution Flow

1. **Task Reception**: API server receives task via `/p2p/message` endpoint
2. **Request Processing**: Handler extracts task data and generates trace ID
3. **Validation Phase**: TaskValidator validates scheduler signature and triggers
4. **Execution Phase**: TaskExecutor performs blockchain actions
5. **Proof Generation**: Creates cryptographic proof of execution
6. **Result Submission**: Uploads to IPFS and submits to aggregator
7. **Response**: Returns execution status to caller

### Task Validation Flow

1. **Validation Request**: API server receives validation request via `/task/validate`
2. **IPFS Data Retrieval**: Fetches task data from IPFS
3. **Comprehensive Validation**: Validates all aspects of task execution
4. **Proof Verification**: Verifies cryptographic proofs
5. **Signature Validation**: Validates all signatures in the chain
6. **Result Return**: Returns validation status

### Health Monitoring Flow

1. **Periodic Check-in**: Health client sends check-in every 60 seconds
2. **Authentication**: Signs keeper address with consensus private key
3. **Status Verification**: Health service verifies keeper registration
4. **Response Processing**: Handles encrypted response and status updates
5. **Error Handling**: Manages verification failures and service errors

## Security Architecture

### Cryptographic Components

- **Consensus Private Key**: Used for task signing and health authentication
- **ECDSA Signatures**: All communications are cryptographically signed
- **TLS Proof Generation**: Creates tamper-proof execution evidence
- **Message Encryption**: Health service responses are encrypted

### Validation Layers

- **Scheduler Signature Validation**: Ensures task authenticity
- **Trigger Validation**: Validates blockchain trigger conditions
- **Action Validation**: Verifies executed actions against expected results
- **Proof Validation**: Validates cryptographic execution proofs
- **Performer Signature Validation**: Ensures executor authenticity

### Network Security

- **Private Key Management**: Secure handling of consensus and controller keys
- **Request Authentication**: All external communications are authenticated
- **Error Handling**: Secure error responses without information leakage
- **Graceful Shutdown**: Secure cleanup of resources and connections

## Integration Points

### External Services

- **Aggregator Service**: Coordinates with validator network
- **Health Service**: Monitors keeper status and verification
- **IPFS Network**: Stores execution proofs and task data
- **Blockchain Networks**: Interacts with L1/L2 chains for validation and execution

### Internal Dependencies

- **Docker Code Executor**: Sandboxed code execution environment
- **Logging System**: Structured logging with trace correlation
- **Metrics Collection**: Prometheus-based monitoring
- **Configuration Management**: Environment-based configuration system

## TODO: Improvements

1. **Health Check In**: The Valid and active Schedulers' Signing Address and rotating TLS certificates for much better security in validation.
2. **Single Function for execution**: Current implementation is 2 functions, with redundant code, we can add a simple if check for dynamic args, and use codeExecutor class if needed.
3. **Update Condition Based Trigger Validation**: Add support for past values from Oracles and APIs, if supported.
4. **Update Action Validation**: Current validation checks for tx success, and block time <= ExpirationTime + Tolerance. No calldata and arguments check. Will be successful after implementation of former point.
5. The TLS server with rotating certificates, as defined in [TLS](tls.md).
