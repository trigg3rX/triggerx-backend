# TODO List

## Current State Analysis

The codebase has tight coupling between services with direct instantiation of external dependencies. Key issues identified:

**Hard-coded Dependencies:**
- Database connections are directly instantiated in service constructors
- Redis clients created without interface abstraction
- Ethereum clients and blockchain connections hard-coded
- IPFS clients are directly instantiated
- Docker manager is tightly coupled to handlers

**Singleton Patterns:**
- Database managers using the singleton pattern (health service)
- State managers with global instances
- No dependency injection for external services

**Missing Interfaces:**
- Database operations not abstracted
- Blockchain client operations are not interface-based
- Notification services (Telegram, Email) are directly coupled
- Repository patterns are inconsistent across services

## Refactoring Strategy

### 1. Database Layer Abstraction

**Create Database Interface:**
```go
type DatabaseInterface interface {
    // Connection management
    Connect(ctx context.Context) error
    Close() error
    IsConnected() bool
    
    // Session management
    Session() SessionInterface
    
    // Health checks
    Ping(ctx context.Context) error
}

type SessionInterface interface {
    Query(stmt string, values ...interface{}) QueryInterface
    Batch() BatchInterface
    Close() error
}

type QueryInterface interface {
    Scan(dest ...interface{}) error
    Exec() error
    Iter() IterInterface
}
```

### 2. External Service Interfaces

**Blockchain Client Interface:**
```go
type BlockchainClientInterface interface {
    // Chain operations
    GetBlockNumber(ctx context.Context) (uint64, error)
    GetChainID(ctx context.Context) (*big.Int, error)
    
    // Contract operations
    CallContract(ctx context.Context, call ethereum.CallMsg) ([]byte, error)
    SendTransaction(ctx context.Context, tx *types.Transaction) error
    
    // Event subscription
    SubscribeFilterLogs(ctx context.Context, q ethereum.FilterQuery, ch chan<- types.Log) (ethereum.Subscription, error)
}
```

**Redis Client Interface (Already exists, extend):**
```go
type RedisClientInterface interface {
    // Add missing methods for complete coverage
    SetWithExpiration(ctx context.Context, key string, value interface{}, expiration time.Duration) error
    GetWithDefault(ctx context.Context, key string, defaultValue interface{}) (interface{}, error)
    DeletePattern(ctx context.Context, pattern string) error
    Exists(ctx context.Context, keys ...string) (int64, error)
}
```

**Notification Service Interface:**
```go
// pkg/notifications/interface.go
type NotificationServiceInterface interface {
    SendEmail(to, subject, body string) error
    SendTelegramMessage(chatID string, message string) error
    SendSlackMessage(channel, message string) error
}
```

### 3. Repository Pattern Standardisation

**Create Base Repository Interface:**
```go
// pkg/repository/interface.go
type RepositoryInterface[T any] interface {
    Create(ctx context.Context, entity T) error
    GetByID(ctx context.Context, id string) (T, error)
    Update(ctx context.Context, entity T) error
    Delete(ctx context.Context, id string) error
    List(ctx context.Context, filter Filter) ([]T, error)
}

type Filter interface {
    ToQuery() (string, []interface{})
}
```

**Standardise All Repositories:**
- EventJobRepository
- ConditionJobRepository  
- TimeJobRepository
- TaskRepository
- UserRepository
- KeeperRepository
- ApiKeysRepository

### 4. Mock Implementations

Mock External Services:

- MockDatabase
- MockBlockchainClient
- MockRedisClient  
- MockNotificationService
- MockDockerManager
- MockIPFSClient

## 1. Internal Services

Switch to GRPC from API interactions between internal services.

- [ ] Start the trace ID at DB Server - Schedulers interaction
- [ ] Carry trace ID across all services, add trace ID to DB
- [ ] Include Keepers in the trace ID loop, with error update back to backend
  - [ ] Health check-in can be expanded to cover this base, or create a new interface to another service

## 2. Keepers

- [ ] Use of seccomp profiles in container executions - restricts misuse of keepers
  - [ ] Can we pass secrets for API keys from the user in a secure method?
- [ ] Switch to keystore file reading for wallet keys
- [ ] CLI: add methods to cover all operations, update the install script

## 3. Users (Developer) UX

- [ ] Add script language, IPFS doesn't store the extension if the name is not added upon upload
- [ ] Simple dynamic price fetch via API / oracle should be migrated from user defined script to our provided options

## 4. Developer UX

- [ ] Logger: Add function signatures and trace IDs in log structure. Add logger.Info() and logger.Infow()
- [ ] Persistent storage: Migrate all data saving to `~/.triggerx/` folder - logs, cache, peerstore
