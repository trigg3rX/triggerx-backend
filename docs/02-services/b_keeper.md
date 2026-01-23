# Keeper Service Architecture

## Service Overview

The Keeper service is a distributed task execution and validation node that operates as part of a decentralized network. It receives tasks from TaskDispatcher via HTTP, validates triggers, executes actions (including agent scripts), generates cryptographic proofs, and coordinates with validators through an aggregator service. It also validates tasks executed by peer Keepers for attestation.

## Core Architecture Layers

### 1. Main Entry Point (`cmd/keeper/main.go`)

#### Initialization Sequence

- **Configuration Loading**: Initializes environment-based configuration via `config.Init()`
- **Logger Setup**: Creates structured logger with process-specific configuration
- **Dependency Injection**: Initializes all service dependencies in sequence:
  1. Observability (Logger, Tracer, Metrics)
  2. Metrics collector (`metrics.NewCollector()`)
  3. Health client (`health.NewClient()`) - performs initial check-in
  4. Aggregator client (`aggregator.NewAggregatorClient()`)
  5. Code executor (`docker.NewCodeExecutor()`) - Docker executor with language pools
  6. IPFS client (`ipfs.NewClient()`)
  7. TaskMonitor client (`taskmonitor.NewClient()`) - optional, may not be configured
  8. Task validator (`validation.NewTaskValidator()`)
  9. Task executor (`execution.NewTaskExecutor()`)
  10. API server (`api.NewServer()`)

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
  - `POST /p2p/message` - Task execution endpoint (receives hex-encoded JSON, returns 202 Accepted for async processing)
  - `POST /task/validate` - Task validation endpoint (for attestation)
  - `POST /task/rebroadcast` - Rebroadcast task execution results
  - `GET /status` - Health check endpoint
  - `GET /metrics` - Prometheus metrics endpoint

#### Request Handlers (`handlers/`)

- **Task Handler** (`handler.go`, `execute.go`, `validate.go`, `rebroadcast.go`):
  - `ExecuteTask()`: Processes incoming task execution requests (async, returns 202 Accepted)
  - `ValidateTask()`: Validates task data and proofs from peer keepers
  - `RebroadcastTask()`: Rebroadcasts task execution results to aggregator
- **Metrics Handler** (`metrics.go`):
  - `Metrics()`: Serves Prometheus metrics data

#### Middleware (`middleware.go`)

- **TraceMiddleware()**: Generates and propagates request trace IDs, this trace would be Attester Node -> core services -> response to attester. This is different from the flow: Scheduler -> (execution and validation) -> DB Server, which will be implemented in the future. This is currently keeper-specific (local).
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
  - TaskMonitor client for status reporting
  - Nonce managers for transaction management per chain
  - Broadcast data storage for rebroadcast capability

- **ExecuteTask() Function Flow**:
  1. **Network Validation**: Verifies task network matches keeper network
  2. **Manager Signature Validation**: Verifies TaskDispatcher signature
  3. **Parallel Execution**: Executes all target data in parallel goroutines
  4. **For Each Target**:
     a. **Trigger Validation**: Validates trigger conditions
     b. **Blockchain Client Setup**: Establishes RPC connection to target chain
     c. **Action Execution**: Executes task based on TaskDefinitionID:
     - **Traditional Jobs (TDI 1-6)**:
       - Static args (1, 3, 5): Use provided arguments
       - Dynamic args (2, 4, 6): Execute script to generate arguments
     - **Agent Jobs (TDI 7, 8, 9)**: - Execute agent script in Docker - Parse JSON output: `{shouldExecute, targetContract, calldata, storageUpdates}` - If `shouldExecute=false`, skip transaction submission - Use script-provided `targetContract` and `calldata`
       d. **Transaction Submission**: Submit on-chain transaction (if applicable)
       e. **Proof Generation**: Creates TLS-based cryptographic proof
       f. **Data Signing**: Signs IPFS data with consensus private key
       g. **IPFS Upload**: Uploads proof data to IPFS network
       h. **Aggregator Submission**: Broadcasts results to aggregator via P2P
       i. **Status Reporting**: Reports execution status to TaskMonitor
  5. **Error Handling**: Reports failures to TaskMonitor
  6. **Broadcast Storage**: Stores broadcast data for rebroadcast capability

#### Task Validation (`validation/validator.go`)

- **TaskValidator Structure**:

  - Alchemy and Etherscan API integration
  - Docker code executor for validation (re-execution if needed)
  - IPFS client for fetching task data
  - Aggregator client for network coordination

- **ValidateTask() Function Flow**:
  1. **IPFS Data Retrieval**: Decode hex data and fetch from IPFS (using CID)
  2. **Trace Continuation**: Continue trace from IPFS data if available
  3. **Network Validation**: Verify task network matches keeper network
  4. **Manager Signature Validation**: Verify TaskDispatcher signature
  5. **Blockchain Connection**: Establish RPC connection to target chain
  6. **Trigger Validation**: Validate trigger conditions
  7. **Action Validation**:
     - Fetch transaction receipt from blockchain
     - Verify transaction success (status = 1)
     - Verify transaction timestamp <= expiration_time + tolerance
  8. **Proof Validation**: Verify TLS certificate proof
  9. **Performer Signature Validation**: Verify performer's signature on IPFS data
  10. **Return Validation Result**: Return true if all validations pass

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

1. **Task Reception**: API server receives HTTP POST to `/p2p/message` endpoint
   - Request body: `{"data": "0x<hex-encoded-json>"}`
   - Extract trace context from HTTP headers
2. **Request Processing**:
   - Decode hex data to JSON
   - Parse `SendTaskDataToKeeper` structure
   - Verify this keeper is the assigned performer
   - Return 202 Accepted immediately (async processing)
3. **Async Execution** (in goroutine):
   a. **Network Validation**: Verify task network matches keeper network
   b. **Manager Signature Validation**: Verify TaskDispatcher signature
   c. **Parallel Target Execution**: Execute all targets concurrently
   d. **For Each Target**:
   - **Trigger Validation**: Validate trigger conditions
   - **Action Execution**:
     - Traditional (TDI 1-6): Execute contract call with args
     - Agent (TDI 7-9): Execute script, parse output, submit if `shouldExecute=true`
   - **Proof Generation**: Generate TLS certificate proof
   - **IPFS Upload**: Package and upload execution data
   - **Aggregator Broadcast**: Send to aggregator via P2P
   - **Status Report**: Report to TaskMonitor
4. **Error Handling**: Report failures to TaskMonitor with error details

### Task Validation Flow (Attestation)

1. **Validation Request**: API server receives HTTP POST to `/task/validate`
   - Request body contains IPFS CID (hex-encoded)
2. **IPFS Data Retrieval**:
   - Decode hex data
   - Fetch complete task data from IPFS using CID
   - Continue trace from IPFS data if available
3. **Network Validation**: Verify task network matches keeper network
4. **Manager Signature Validation**: Verify TaskDispatcher signature
5. **Blockchain Connection**: Establish RPC connection to target chain
6. **Trigger Validation**: Validate trigger conditions
7. **Action Validation**:
   - Fetch transaction receipt from blockchain
   - Verify transaction success
   - Verify transaction was executed before expiration time + tolerance
8. **Proof Validation**: Verify TLS certificate proof
9. **Performer Signature Validation**: Verify performer's signature
10. **Result Return**: Return validation status (true/false)
11. **Attestation**: Sign attestation and submit to aggregator (if validation passes)

### Health Monitoring Flow

1. **Initial Check-in**: On startup, perform health check-in to verify keeper registration
   - If keeper not verified, service shuts down
2. **Periodic Check-in**: Health client sends check-in every 60 seconds
   - Derive consensus address from private key
   - Sign keeper address with consensus private key
   - Create payload: `{keeper_address, consensus_address, version, peer_id, network, timestamp, signature}`
3. **HTTP Request**: POST to `/health` endpoint
4. **Response Processing**:
   - Health service verifies signature
   - Updates keeper status in database
   - Returns encrypted task execution address
5. **Error Handling**:
   - If keeper not verified, initiate graceful shutdown
   - Log errors but continue operation for transient failures

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

- **TaskDispatcher**: Sends tasks via HTTP POST to `/p2p/message`
- **Aggregator Service**: Receives execution results via P2P network, coordinates consensus
- **Health Service**: Monitors keeper status, provides keeper selection for TaskDispatcher
- **TaskMonitor Service**: Receives execution status reports (success/failure)
- **IPFS Network**: Stores execution proofs and task data
- **Blockchain Networks**: Interacts with L1/L2 chains for validation and execution

### Internal Dependencies

- **Docker Code Executor**: Sandboxed code execution environment with resource limits
- **Task Validator**: Pre-execution and post-execution validation
- **Argument Converter**: Handles static and dynamic argument processing
- **Nonce Managers**: Per-chain transaction nonce management
- **IPFS Client**: File upload and retrieval
- **Logging System**: Structured logging with trace correlation
- **Metrics Collection**: Prometheus-based monitoring
- **Configuration Management**: Environment-based configuration system

## TODO: Improvements

1. **Health Check In**: The Valid and active Schedulers' Signing Address and rotating TLS certificates for much better security in validation.
2. **Single Function for execution**: Current implementation is 2 functions, with redundant code, we can add a simple if check for dynamic args, and use codeExecutor class if needed.
3. **Update Condition Based Trigger Validation**: Add support for past values from Oracles and APIs, if supported.
4. **Update Action Validation**: Current validation checks for tx success, and block time <= ExpirationTime + Tolerance. No calldata and arguments check. Will be successful after implementation of former point.
5. The TLS server with rotating certificates (documentation pending).
