# Service Refactoring for Testability

## Current State Analysis

The TriggerX backend consists of multiple services that handle different aspects of the decentralized task execution system. These services include:

- **Core Services**: `dbserver`, `health`, `schedulers/time`, `schedulers/condition`, `taskdispatcher`, `taskmonitor`, `keeper`
- **Package Services**: `pkg/client/aggregator`, `pkg/client/docker`, `pkg/client/redis`, `pkg/dockerexecutor`, `pkg/http`, `pkg/proof`, `pkg/retry`, `pkg/rpc`
- **CLI Tools**: Command-line interface for operator management and key generation

### Current Issues for Testability

1. **Tight coupling with external dependencies**: Direct dependencies on databases, HTTP clients, file systems, and blockchain networks
2. **Hard-coded external dependencies**: HTTP clients, file system operations, database connections
3. **Insufficient interfaces**: While some interfaces exist, they're not consistently used across all services
4. **Missing dependency injection**: Components create their own dependencies instead of receiving them
5. **Limited test coverage**: Most services lack comprehensive unit, integration, and API tests
6. **Complex state management**: Goroutines and shared state without proper abstractions
7. **Resource cleanup**: Not testable due to direct external API calls and resource dependencies

## Refactoring Phases

### Phase 1: Interface Extraction and Basic Dependency Injection

#### Goals

- Extract interfaces for all external dependencies
- Implement dependency injection in constructors
- Create mock implementations for testing

#### Tasks

1. **External Dependency Interfaces**

   - Create interfaces for database clients (ScyllaDB, Redis)
   - Create interfaces for HTTP clients and blockchain RPC clients
   - Create interfaces for file system operations
   - Create interfaces for container management (Docker)
   - Create interfaces for cryptographic operations

2. **Service-Specific Interfaces**

   - Create interfaces for each service's core functionality
   - Define clear boundaries between service layers
   - Implement wrapper around external dependencies
   - Create mock implementations for all interfaces

3. **Update Constructors**
   - Modify all constructors to accept interfaces
   - Use dependency injection pattern consistently
   - Remove hard-coded dependency creation

#### Deliverables:

- `interfaces/` packages with all service interfaces
- `mocks/` packages with mock implementations
- Updated constructors with dependency injection
- Basic unit tests for core functions

### Phase 2: Repository Pattern Implementation

#### Goals

- Implement repository pattern for data access
- Separate business logic from data storage
- Enable in-memory testing

#### Tasks

1. **Data Repository Interfaces**

   - Implement repository interfaces for all data access
   - Add methods for CRUD operations
   - Create mock repositories for testing

2. **Service-Specific Repositories**

   - **Task Repository**: Store and manage task execution data
   - **Keeper Repository**: Manage keeper registration and status
   - **User Repository**: Handle user and API key management
   - **Cache Repository**: Implement caching with TTL and eviction policies
   - **Metrics Repository**: Store and aggregate performance metrics

3. **Storage Adapters**

   - Support both persistent and in-memory storage
   - Implement database-specific adapters (ScyllaDB, Redis)
   - Add transaction support where needed

4. **Update Existing Code**
   - Replace direct database/API calls with repository calls
   - Update service layers to use repositories
   - Maintain backward compatibility

#### Deliverables

- Repository implementations with interfaces
- In-memory and persistent storage adapters
- Unit tests for repository operations
- Integration tests with real storage

### Phase 3: Service Layer Refactoring

#### Goals

- Extract business logic into service layer
- Implement command and query separation
- Add comprehensive error handling

#### Tasks

1. **Core Service Layer**

   - **Keeper Service**: Extract keeper management and task execution logic
   - **Scheduler Service**: Implement job scheduling and load balancing
   - **Aggregator Service**: Handle task aggregation and validation
   - **Health Service**: Implement health monitoring and status reporting

2. **Supporting Services**

   - **Task Management Service**: Extract task lifecycle management
   - **Validation Service**: Implement task and execution validation
   - **Metrics Service**: Handle performance monitoring and aggregation
   - **Notification Service**: Manage alerts and status updates

3. **Cross-Cutting Services**

   - **Authentication Service**: Handle API key and user authentication
   - **Configuration Service**: Manage service configuration and validation
   - **Logging Service**: Centralized logging and audit trails
   - **Retry Service**: Implement retry mechanisms and circuit breakers

4. **Orchestration Service**
   - Coordinate between services
   - Implement transaction-like operations
   - Add retry mechanisms and error recovery

#### Deliverables

- Service layer with clear boundaries
- Business logic tests without external dependencies
- Error handling and recovery mechanisms
- Performance optimizations

### Phase 4: Event-Driven Architecture

#### Goals

- Implement event-driven communication
- Add comprehensive monitoring
- Enable loose coupling between components

#### Tasks

1. **Event System**

   - Implement event publisher/subscriber for all services
   - Define event schemas for task execution, keeper status, and system events
   - Add event persistence and replay capabilities

2. **Monitoring Service**

   - Subscribe to service events across the system
   - Aggregate metrics and performance statistics
   - Generate alerts and notifications for critical events

3. **Audit Service**

   - Track all operations across services
   - Store execution history and system state changes
   - Enable debugging and troubleshooting capabilities

4. **Health Monitoring**
   - Implement health checks for all services
   - Monitor resource usage and performance metrics
   - Detect and handle failures with automatic recovery

#### Deliverables

- Event-driven architecture across all services
- Comprehensive monitoring system
- Audit trail and debugging tools
- Proactive health monitoring and alerting

### Phase 5: Testing Implementation

#### Goals

- Achieve 90%+ test coverage across all services
- Implement all test categories
- Set up CI/CD integration

#### Tasks

1. **Unit Tests**

   - Test all service methods in isolation
   - Mock all external dependencies (databases, HTTP clients, blockchain RPC)
   - Cover edge cases and error conditions
   - Target: 95% coverage per service

2. **Integration Tests**

   - Test service interactions and workflows
   - Use testcontainers for database and Docker testing
   - Test with real external dependencies in controlled environments
   - Cover end-to-end task execution scenarios

3. **API Tests**

   - Test all public service interfaces
   - Validate input/output contracts
   - Test authentication and authorization
   - Performance and load testing for all endpoints

4. **Database Tests**

   - Test repository implementations across all services
   - Validate data integrity and consistency
   - Test transaction scenarios and rollback mechanisms
   - Migration testing for schema changes

5. **Benchmark Tests**
   - Performance-critical operations across all services
   - Resource usage profiling and optimization
   - Scalability testing for high-load scenarios
   - Network and blockchain interaction performance

#### Deliverables

- Comprehensive test suite for all services
- CI/CD pipeline integration with automated testing
- Performance benchmarks and monitoring
- Test documentation and guidelines

## Test Architecture

### Testing Strategy

#### 1. Unit Tests Structure

```bash
# Core Services
cmd/keeper/
├── keeper_test.go # Main keeper service tests
├── api/
│ ├── server_test.go # API server tests
│ └── handlers_test.go # HTTP handler tests
├── core/
│ ├── execution_test.go # Task execution tests
│ └── validation_test.go # Task validation tests
└── config/
    └── config_test.go # Configuration tests

cmd/schedulers/
├── condition/
│ ├── scheduler_test.go # Condition scheduler tests
│ ├── api/
│ │ └── handlers_test.go # API handler tests
│ └── client/
│     └── client_test.go # Client tests
└── time/
    ├── scheduler_test.go # Time scheduler tests
    └── api/
        └── handlers_test.go # API handler tests

# Package Services
pkg/database/
├── connection_test.go # Database connection tests
├── retry_test.go # Retry mechanism tests
└── interfaces_test.go # Interface tests

pkg/redis/
├── client_test.go # Redis client tests
├── jobstream_test.go # Job stream tests
└── cache_test.go # Cache tests

pkg/http/
├── client_test.go # HTTP client tests
└── middleware_test.go # Middleware tests

# Service-Specific Tests
internal/keeper/
├── keeper_test.go # Keeper service tests
├── api/
│ └── server_test.go # API server tests
└── core/
    ├── execution_test.go # Execution logic tests
    └── validation_test.go # Validation logic tests
```

#### 2. Mock Strategy

- Use `testify/mock` for interface mocking across all services
- Create realistic mock behaviors for external dependencies
- Support both success and failure scenarios
- Enable deterministic testing with controlled mock responses

#### 3. Test Data Management

- Use table-driven tests for multiple scenarios across all services
- Create test fixtures for complex data structures
- Implement test data builders for consistent test setup
- Clean up resources after tests (database connections, file systems, etc.)

#### 4. Integration Testing

- Use `testcontainers-go` for Docker and database integration
- Set up test databases (ScyllaDB, Redis) in containers
- Test real HTTP endpoints and blockchain RPC connections
- Validate end-to-end workflows across service boundaries

### Performance Testing

#### 1. Benchmarks

- Service startup and shutdown times
- Task execution performance across all services
- Database query performance and cache hit/miss ratios
- Resource utilization (CPU, memory, network)

#### 2. Load Testing

- Concurrent task executions across multiple keepers
- High-frequency scheduler operations
- Database connection pool exhaustion scenarios
- Memory pressure testing under load

#### 3. Scalability Testing

- Multiple service instances and load balancing
- Large-scale task processing and aggregation
- High-frequency blockchain interactions
- Resource cleanup efficiency under stress

## Implementation Guidelines

### 1. Code Organization

- Follow service-per-domain pattern with clear boundaries
- Keep interfaces in separate files for all services
- Use dependency injection consistently across all components
- Implement builder patterns for complex service configurations

### 2. Error Handling

- Define custom error types for each service domain
- Implement error wrapping with service context
- Add correlation IDs for error tracking across services
- Create error classification system for different failure types

### 3. Logging and Monitoring

- Use structured logging with consistent format across all services
- Add trace IDs for correlation across service boundaries
- Monitor resource usage and performance metrics for all services
- Track service health and availability metrics

### 4. Configuration

- Use environment-based configuration for all services
- Validate configuration at startup for each service
- Support configuration reloading where appropriate
- Document all configuration options with service-specific examples

## Success Criteria

### 1. Test Coverage

- Unit tests: 95% coverage across all services
- Integration tests: 80% coverage for service interactions
- API tests: 100% endpoint coverage for all services
- Critical path: 100% coverage for task execution workflows

### 2. Performance

- Service startup: <5 seconds for all services
- Task execution: <30 seconds for complex tasks
- Database query performance: <100ms for 95th percentile
- Memory usage: <2GB per service instance

### 3. Reliability

- Zero memory leaks across all services
- Proper resource cleanup and connection management
- Graceful error handling and recovery mechanisms
- Service availability: >99.9% uptime

### 4. Maintainability

- Clear separation of concerns between services
- Minimal cyclomatic complexity (<10 per function)
- Comprehensive documentation for all services
- Easy to extend and modify individual services
