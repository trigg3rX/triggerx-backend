# TriggerX Datastore Architecture

## Overview

This document outlines the architecture for the TriggerX datastore package, designed as a service-oriented ScyllaDB client that can be imported and used across different microservices. The architecture follows clean architecture principles with proper separation of concerns, dependency injection, and service-oriented design.

## Key Design Principles

- **Service-Oriented**: Importable service that connects to an already initialized database
- **No Validation**: Data is pre-validated before reaching the datastore layer
- **Per-Service Connection**: Each microservice maintains its own connection to the shared ScyllaDB instance
- **Generic Repository Pattern**: Type-safe generic repositories for all data entities
- **Interface-Based Design**: All components implement interfaces for testability and flexibility
- **Query Builder Integration**: Uses gocqlx for efficient query building and execution
- **Thread-Safe Operations**: Repository factory and operations are thread-safe

## Architecture Layers

### 1. Interface Layer (`pkg/datastore/interfaces/`)
Contains all interface definitions for the datastore components.

```
interfaces/
├── connection.go          # Connection and session interfaces
├── query_builder.go       # Query builder interface
└── repository.go          # Repository and factory interfaces
```

### 2. Infrastructure Layer (`pkg/datastore/infrastructure/`)
Contains concrete implementations of interfaces.

```
infrastructure/
├── connection/            # Database connection management
│   ├── config.go         # Connection configuration
│   ├── connection.go     # Connection implementation
│   └── *_test.go         # Connection tests
├── query_builder/        # Query builder implementation
│   ├── query_builder.go  # Gocqlx query builder
│   └── *_test.go         # Query builder tests
└── repository/           # Repository implementations
    ├── factory.go        # Repository factory
    ├── repository.go     # Generic repository implementation
    └── *_test.go         # Repository tests
```

### 3. Service Layer (`pkg/datastore/`)
Main service entry point and public API.

```
├── datastore.go          # Main service interface and implementation
└── datastore_test.go     # Service tests
```

### 4. Mocks Layer (`pkg/datastore/mocks/`)
Mock implementations for testing.

```
mocks/
├── connection_mock.go           # Connection mocks
├── query_builder_mock.go        # Query builder mocks
├── repository_mocks.go          # Repository mocks
└── *_test.go                    # Mock tests
```

## Core Components

### 1. Connection Manager (`infrastructure/connection/`)

**Purpose**: Connection manager that handles ScyllaDB connections with health checking and session management.

**Key Features**:
- Per-service connection management (each microservice has its own connection)
- Health checking with configurable intervals
- Connection pooling and session management
- Graceful shutdown handling
- Configuration validation and defaults

**Files**:
- `connection.go`: Core connection implementation with gocql session
- `config.go`: Connection configuration with validation
- `*_test.go`: Comprehensive connection tests

### 2. Query Builder (`infrastructure/query_builder/`)

**Purpose**: Gocqlx-based query builder for efficient ScyllaDB operations with type safety.

**Key Features**:
- Generic type-safe operations (Insert, Update, Delete, Select, Get)
- Prepared statement caching
- Batch operations support
- Gocqlx integration for better performance
- Error handling and logging

**Files**:
- `query_builder.go`: Gocqlx query builder implementation
- `*_test.go`: Query builder tests

### 3. Generic Repository (`infrastructure/repository/`)

**Purpose**: Generic repository implementation using reflection and query builder for all data entities.

**Key Features**:
- Generic type-safe repository for all entities
- CRUD operations with reflection-based field mapping
- Batch operations support
- Custom query execution
- Field-based filtering and searching
- Thread-safe operations

**Repository Structure**:
```go
type GenericRepository[T any] interface {
    Create(ctx context.Context, data *T) error
    Update(ctx context.Context, data *T) error
    GetByID(ctx context.Context, id interface{}) (*T, error)
    GetByNonID(ctx context.Context, field string, value interface{}) (*T, error)
    List(ctx context.Context) ([]*T, error)
    ExecuteQuery(ctx context.Context, query string, values ...interface{}) ([]*T, error)
    BatchCreate(ctx context.Context, data []*T) error
    GetByField(ctx context.Context, field string, value interface{}) ([]*T, error)
    Count(ctx context.Context) (int64, error)
    Exists(ctx context.Context, id interface{}) (bool, error)
    Close()
}
```

### 4. Repository Factory (`infrastructure/repository/`)

**Purpose**: Factory for creating type-specific repositories with proper table and primary key configuration.

**Key Features**:
- Thread-safe repository creation
- Pre-configured table names and primary keys
- Repository caching (reuses instances)
- Support for all data entities

**Supported Entities**:
- UserDataEntity → user_data table
- JobDataEntity → job_data table
- TimeJobDataEntity → time_job_data table
- EventJobDataEntity → event_job_data table
- ConditionJobDataEntity → condition_job_data table
- TaskDataEntity → task_data table
- KeeperDataEntity → keeper_data table
- ApiKeyDataEntity → apikeys table

### 5. Main Service (`datastore.go`)

**Purpose**: Main entry point for the datastore service with dependency injection.

**Key Features**:
- Service factory for dependency injection
- Direct repository access methods
- Health check functionality
- Connection lifecycle management
- Interface-based design for testability

## Data Models

The datastore package uses entities from `pkg/types` package. These entities are used directly without DTOs since data validation is handled at the application layer.

### Entity Structure
Entities are defined in `pkg/types` and follow ScyllaDB CQL field mapping patterns:

```go
// Example from pkg/types (actual structure may vary)
type UserDataEntity struct {
    UserID         int64     `cql:"user_id"`
    UserAddress    string    `cql:"user_address"`
    EmailID        string    `cql:"email_id"`
    JobIDs         []int64   `cql:"job_ids"`
    TgConsumed     int64     `cql:"tg_consumed"`
    TotalJobs      int64     `cql:"total_jobs"`
    TotalTasks     int64     `cql:"total_tasks"`
    CreatedAt      time.Time `cql:"created_at"`
    LastUpdatedAt  time.Time `cql:"last_updated_at"`
}
```

### Supported Entities
The datastore package supports the following entities from `pkg/types`:

- `types.UserDataEntity` → user_data table
- `types.JobDataEntity` → job_data table
- `types.TimeJobDataEntity` → time_job_data table
- `types.EventJobDataEntity` → event_job_data table
- `types.ConditionJobDataEntity` → condition_job_data table
- `types.TaskDataEntity` → task_data table
- `types.KeeperDataEntity` → keeper_data table
- `types.ApiKeyDataEntity` → apikeys table

### Field Mapping
The generic repository uses reflection to map struct fields to database columns:
- CQL tags determine the database column names
- Primary keys are configured per entity in the repository factory
- Field types are automatically handled by gocqlx

## Service Usage

### Initialization
```go
import (
    "github.com/trigg3rX/triggerx-backend/pkg/datastore"
    "github.com/trigg3rX/triggerx-backend/pkg/datastore/infrastructure/connection"
    "github.com/trigg3rX/triggerx-backend/pkg/logging"
)

// Initialize the datastore service in each microservice
config := connection.NewConfig("scylla-node", "9042")
config.WithKeyspace("triggerx")
config.WithTimeout(30 * time.Second)

logger := logging.NewLogger() // or your preferred logger
service, err := datastore.NewService(config, logger)
if err != nil {
    log.Fatal(err)
}
defer service.Close()
```

### Usage in Microservices
```go
import (
    "github.com/trigg3rX/triggerx-backend/pkg/datastore"
    "github.com/trigg3rX/triggerx-backend/pkg/types"
)

// In your microservice (e.g., dbserver, keeper, etc.)
type MyService struct {
    datastore datastore.DatastoreService
}

func NewMyService(datastore datastore.DatastoreService) *MyService {
    return &MyService{
        datastore: datastore,
    }
}

func (s *MyService) CreateUser(ctx context.Context, userData *types.UserDataEntity) error {
    return s.datastore.User().Create(ctx, userData)
}

func (s *MyService) GetUserByID(ctx context.Context, userID int64) (*types.UserDataEntity, error) {
    return s.datastore.User().GetByID(ctx, userID)
}

func (s *MyService) ListAllUsers(ctx context.Context) ([]*types.UserDataEntity, error) {
    return s.datastore.User().List(ctx)
}

// Each microservice initializes its own datastore connection
func main() {
    // Initialize datastore service for this microservice
    config := connection.NewConfig("scylla-node", "9042")
    datastoreService, err := datastore.NewService(config, logger)
    if err != nil {
        log.Fatal(err)
    }
    defer datastoreService.Close()
    
    // Use in your service
    myService := NewMyService(datastoreService)
    // ... rest of your service logic
}
```

### Repository Access Methods
```go
// Direct repository access
userRepo := service.User()
jobRepo := service.Job()
timeJobRepo := service.TimeJob()
eventJobRepo := service.EventJob()
conditionJobRepo := service.ConditionJob()
taskRepo := service.Task()
keeperRepo := service.Keeper()
apiKeyRepo := service.ApiKey()

// Factory access for custom operations
factory := service.GetFactory()
customUserRepo := factory.CreateUserRepository()

// Health check
err := service.HealthCheck(ctx)
if err != nil {
    log.Printf("Database health check failed: %v", err)
}
```

## Configuration

### Connection Configuration
```go
type Config struct {
    Hosts               []string
    Keyspace            string
    Timeout             time.Duration
    Retries             int
    ConnectWait         time.Duration
    Consistency         gocql.Consistency
    HealthCheckInterval time.Duration
    ProtoVersion        int
    SocketKeepalive     time.Duration
    MaxPreparedStmts    int
    DefaultIdempotence  bool
    RetryConfig         *retry.RetryConfig
}
```

### Configuration Creation
```go
import (
    "github.com/trigg3rX/triggerx-backend/pkg/datastore/infrastructure/connection"
    "time"
)

// Create config with sensible defaults
config := connection.NewConfig("scylla-node", "9042")

// Customize configuration
config.WithKeyspace("triggerx")
config.WithTimeout(30 * time.Second)
config.WithRetries(5)
config.WithConnectWait(10 * time.Second)
config.WithHealthCheckInterval(15 * time.Second)

// Validate configuration
err := config.Validate()
if err != nil {
    log.Fatal(err)
}
```

### Default Configuration Values
- **Hosts**: `["scylla-node:9042"]` (from parameters)
- **Keyspace**: `"triggerx"`
- **Timeout**: `30 seconds`
- **Retries**: `5`
- **ConnectWait**: `10 seconds`
- **Consistency**: `gocql.Quorum`
- **HealthCheckInterval**: `15 seconds`
- **ProtoVersion**: `4`
- **SocketKeepalive**: `15 seconds`
- **MaxPreparedStmts**: `1000`
- **DefaultIdempotence**: `true`
- **RetryConfig**: `retry.DefaultRetryConfig()`

## Error Handling

### Error Types
- `ConnectionError`: Database connection issues
- `QueryError`: Query execution errors
- `ValidationError`: Data validation errors (minimal since data is pre-validated)
- `TimeoutError`: Query timeout errors
- `NotFoundError`: Entity not found errors

## Migration Strategy

### From Current Implementation
1. **Phase 1**: Create new architecture alongside existing
2. **Phase 2**: Migrate one service at a time
3. **Phase 3**: Remove old implementation
4. **Phase 4**: Optimize and refactor

### Migration Steps
1. Implement new domain entities
2. Create repository interfaces
3. Implement infrastructure layer
4. Create service layer
5. Update microservices to use new service
6. Remove old datastore code

## Dependencies

### External Dependencies
- `github.com/gocql/gocql`: ScyllaDB/Cassandra driver
- `github.com/scylladb/gocqlx/v2`: High-level ScyllaDB/Cassandra toolkit
- `github.com/scylladb/gocqlx/v2/qb`: Query builder for gocqlx

### Internal Dependencies
- `github.com/trigg3rX/triggerx-backend/pkg/logging`: Logging package
- `github.com/trigg3rX/triggerx-backend/pkg/retry`: Retry package  
- `github.com/trigg3rX/triggerx-backend/pkg/types`: Type definitions and data entities

### Testing Dependencies
- `github.com/golang/mock/gomock`: Mock generation and testing
- `github.com/stretchr/testify`: Testing assertions and mocks
- `github.com/stretchr/testify/assert`: Assertion library

## File Structure Summary

```
pkg/datastore/
├── datastore.go                    # Main service interface and implementation
├── datastore_test.go              # Service tests
├── interfaces/                    # Interface definitions
│   ├── connection.go             # Connection and session interfaces
│   ├── query_builder.go          # Query builder interface
│   └── repository.go             # Repository and factory interfaces
├── infrastructure/               # Concrete implementations
│   ├── connection/               # Connection management
│   │   ├── config.go            # Configuration with validation
│   │   ├── connection.go        # Connection implementation
│   │   └── *_test.go            # Connection tests
│   ├── query_builder/           # Query builder implementation
│   │   ├── query_builder.go     # Gocqlx query builder
│   │   └── *_test.go            # Query builder tests
│   └── repository/              # Repository implementations
│       ├── factory.go           # Repository factory
│       ├── repository.go        # Generic repository
│       └── *_test.go            # Repository tests
└── mocks/                       # Mock implementations for testing
    ├── connection_mock.go       # Connection mocks
    ├── query_builder_mock.go    # Query builder mocks
    ├── repository_mocks.go      # Repository mocks
    └── *_test.go               # Mock tests
```

This architecture provides a robust, scalable, and maintainable datastore service that can be easily imported and used across your microservices while maintaining clean separation of concerns and following Go best practices. The generic repository pattern ensures type safety while the interface-based design enables comprehensive testing and flexibility.
