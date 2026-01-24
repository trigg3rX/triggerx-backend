# Services Structure

Detailed breakdown of each service's structure and components.

## Architecture Principles

1. **Layered Architecture**: `api/` or `rpc/` → `core/` → `database/` (if needed) → `rpc/clients/` (if needed)
2. **Separation of Concerns**: Business logic in `core/`, infrastructure in `database/`, `redis/`, `rpc/`
3. **Service Communication**: Inter-service calls via `rpc/clients/`
4. **Consistent Naming**: Repositories in `database/repository/`, clients in `rpc/clients/{service}/`
5. **Configuration Centralization**: All services have `config/` at the root

## Standard Service Structure

```bash
internal/{service}/
├── api/              # HTTP API layer
├── config/           # Configuration
├── core/             # Business logic
├── database/         # Database access layer
├── metrics/          # Metrics collection
├── redis/            # Redis client
└── rpc/              # gRPC layer
```

## Directory Purposes

### `api/` — HTTP API Layer

Handles HTTP requests, middleware, and server setup.

```bash
api/
├── handlers/         # HTTP request handlers
├── middleware/       # HTTP middleware (auth, metrics, rate limiting, etc.)
└── server.go         # HTTP server setup
```

OR

```bash
api/
└── status.go     # Status endpoint for Nginx and Pulsetic
```

### `config/` — Configuration

Service configuration loading and validation.

```bash
config/
└── config.go         # Configuration struct and loading
```

### `core/` — Business Logic

Core domain logic organized by feature/domain. Subdirectories group related functionality.

```bash
core/
├── {domain}/         # Domain-specific logic
└── {feature}/        # Feature-specific logic
```

### `database/` — Database Access

Database connections and repository implementations.

```bash
database/
├── connection.go     # Database connection setup
└── repository/       # Repository implementations
    ├── {entity}_repository.go      # Repository for specific entity
    └── {entity}_queries.go          # Query definitions for specific entity
```

### `metrics/` — Metrics Collection

Service-specific metrics instrumentation.

```bash
metrics/
├── metrics.go      # Metric definitions and collectors
└── collectors.go   # Metric collectors
```

### `redis/` — Redis Client

Redis client setup and operations.

```bash
redis/
├── client.go         # Redis client configuration
├── index.go          # Redis index operations
├── stream.go         # Redis stream operations
└── expiration.go     # Redis expiration operations
```

### `rpc/` — gRPC Layer

gRPC server and client implementations for inter-service communication.

```bash
rpc/
├── clients/          # gRPC clients to other services
│   └── {service}/    # Client for specific service
├── handler.go        # gRPC request handlers
└── server.go         # gRPC server setup
```

## dbserver

Main API server handling user requests, database operations, and WebSocket connections.

```bash
internal/dbserver/
├── api, config, database, metrics, redis
├── rpc/
│   └── clients/            # Clients to other services
│       └── conditionscheduler/
├── events/                 # Event publishing
└── websocket/              # WebSocket hub for real-time updates
```

**Notes**: No `core/` folder — business logic resides in `api/handlers/` and `database/repository/`.

## eventmonitor

Monitors blockchain events for condition-based and event-based jobs.

```bash
internal/eventmonitor/
├── api, config, metrics, redis
├── core/
│   ├── attestation/        # Attestation logic
│   ├── registry/           # Job registry
│   ├── service/            # Main service logic
│   ├── types/              # Domain types
│   └── worker/             # Event monitoring workers
└── rpc/
    ├── clients/
    │   ├── conditionscheduler/
    │   └── taskmonitor/
    ├── handler.go          # gRPC handlers
    └── server.go           # gRPC server
```

## health

Health monitoring service managing keeper state and performing cleanup operations.

```bash
internal/health/
├── api, config, database,metrics, redis
├── core/
│   ├── keeper/             # Keeper state management
│   │   ├── cleanup.go
│   │   ├── next_performer_selection.go
│   │   ├── persistence.go
│   │   ├── state_manager.go
│   │   └── update.go
│   └── telegram/           # Telegram bot integration
│       └── bot.go
└── rpc/
    ├── handler.go
    └── server.go
```

## keeper

Executes tasks by running validation, execution, and blockchain transactions.

```bash
internal/keeper/
├── aggregator/             # Aggregator client
│   └── client.go
├── api, config, metrics
├── core/
│   ├── checkin/            # Keeper check-in logic
│   │   └── checkin.go
│   ├── execution/          # Task execution
│   │   └── executor.go
│   ├── nonce_manager/      # Transaction nonce management
│   │   └── nonce_manager.go
│   └── validation/         # Task validation
│       └── validator.go
├── rpc/
│   └── clients/
│       └── taskmonitor/    # Task monitor client
└── utils/
    ├── chainids.go
    └── fetch.go
```

**Notes**: No `redis/` or `database/` folder — uses external clients.

## schedulers/condition

Schedules and monitors condition-based jobs.

```bash
internal/schedulers/condition/
├── api, config, database, metrics, redis
├── core/
│   └── scheduler/
│       ├── job_status_checker.go
│       ├── notification.go
│       ├── schedule.go
│       ├── scheduler.go
│       ├── utils.go
│       └── worker/         # Condition monitoring workers
│           ├── init_condition.go
│           ├── monitor_condition.go
│           ├── monitor_websocket.go
│           └── types.go
└── rpc/
    ├── clients/
    │   ├── eventmonitor/
    │   └── taskdispatcher/
    ├── handler.go
    └── server.go
```

## schedulers/time

Schedules time-based jobs using batch processing and polling.

```bash
internal/schedulers/time/
├── api, config, database, metrics, redis
├── core/
│   └── scheduler/
│       ├── batch.go        # Batch processing
│       ├── dispatcher.go   # Task dispatching
│       └── poll.go         # Job polling
└── rpc/
    └── clients/
        └── taskdispatcher/
```

**Notes**: No gRPC server — only acts as a client to other services.

## taskdispatcher

Dispatches tasks to available keepers for execution.

```bash
internal/taskdispatcher/
├── aggregator/
│   └── client.go
├── api, config, database, metrics, redis
├── core/
│   └── dispatcher/         # Dispatch logic
└── rpc/
    ├── clients/
    │   └── health/
    ├── handler.go
    └── server.go
```

## taskmonitor

Monitors task execution status and handles completion/failure events.

```bash
internal/taskmonitor/
├── api, config, database, metrics, redis
├── core/
│   ├── events/             # Event handling
│   └── tasks/              # Task state management
└── rpc/
    ├── clients/
    │   └── dbserver/
    ├── handler.go
    └── server.go
```

## Service Variations Summary

| Service              | api     | core | database | redis    | rpc server | rpc clients |
|----------------------|---------|------|----------|----------|------------|-------------|
| dbserver             | Full    | No   | Full     | Yes      | No         | Yes         |
| eventmonitor         | Minimal | Yes  | No       | Yes      | Yes        | Yes         |
| health               | Full    | Yes  | Yes      | Yes      | Yes        | No          |
| keeper               | Full    | Yes  | No       | No       | No         | Yes         |
| schedulers/condition | Minimal | Yes  | Yes      | Yes      | Yes        | Yes         |
| schedulers/time      | Minimal | Yes  | Yes      | Yes      | No         | Yes         |
| taskdispatcher       | Minimal | Yes  | Yes      | Yes      | Yes        | Yes         |
| taskmonitor          | Minimal | Yes  | Yes      | Yes      | Yes        | Yes         |

**Legend**:

- **Full**: Complete implementation with handlers, middleware, server
- **Minimal**: Only status endpoint
- **Yes/No**: Present or absent
