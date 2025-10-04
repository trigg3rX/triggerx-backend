# Notes for Developers

## Repository Architecture

```bash
triggerx-backend/
├── .othentic/                   # Othentic Contracts network configurations
├── checker/                     # Testing and validation tools
├── cmd/                         # Main application entry points
├── data/                        # Data storage and persistence
├── internal/                    # Internal packages and shared code
│   ├── dbserver/                # Database server internals
│   │   ├── config/
│   │   ├── handlers/
│   │   ├── middleware/
│   │   └── server.go
│   ├── health/                  # Health monitoring internals
│   │   ├── config/
│   │   ├── client/
│   │   ├── keeper/
│   │   ├── telegram/
│   │   └── health.go
│   ├── registrar/               # Registration service internals
│   │   ├── config/
│   │   ├── client/
│   │   ├── events/
│   │   ├── rewards/
│   │   ├── config/
│   ├── redis/                   # Redis internals
│   │   ├── config/
│   │   ├── redis.go
│   │   └── jobstream.go
│   ├── cache/                   # Cache mechanism using Redis
│   ├── schedulers/              # Schedulers internals
│   │   ├── condition/           # Condition Based Schedulers internals
│   │   │   ├── config/          # Config values from env
│   │   │   └── api/             # API server with status, metrics and job scheduling routes
│   │   │       ├── handlers/
│   │   │       └── server.go
│   │   │   ├── client/          # DB server client
│   │   │   ├── metrics/         # Prometheus Metrics Collector
│   │   │   └── scheduler/       # Scheduler logic with Worker Pools
│   │   ├── event/               # Same structure as condition scheduler
│   │   └── time/                # Same structure as condition scheduler
│   └── keeper/                  # Keeper node internals
│       ├── config/
│       ├── client/
│       ├── api/
│       ├── core/
│       │   ├── execution/
│       │   └── validation/
│       └── keeper.go
├── othentic/                    # Othentic Nodes implementation
├── pkg/                         # Public packages and shared libraries
│   ├── bindings/
│   ├── client/
│   │   ├── aggregator/
│   │   └── ipfs/
│   ├── converter/
│   ├── database/
│   ├── env/
│   ├── logging/
│   ├── proof/
│   ├── redis/
│   ├── resources/
│   ├── retry/
│   ├── parser/
│   ├── types/
│   └── validator/
├── scripts/                     # Utility scripts and tools
├── Dockerfile                   # Keeper container build configuration
├── docker-compose.yaml          # Multi-container orchestration
├── go.mod                       # Go module definition
└── Makefile                     # Build and development commands
```

### Key Components

1. **Core Services** (in `cmd/` and `internal/`)
   - `schedulers/`: Handles job scheduling, load balancing, and task assignment
   - `registrar/`: Manages keeper registration and task submission persistence
   - `health/`: Monitors keeper health and online status
   - `dbserver/`: API server for data persistence regarding jobs, tasks, users, keepers and api keys
   - `keeper/`: Keeper node implementation for task execution and validation

2. **Network Layer** (`othentic/`)
   - Implements the P2P network functionality
   - Handles peer discovery and communication
   - Manages network security and encryption (implemented soon)

3. **Build and Deployment**
   - `Dockerfile`: Container configuration
   - `docker-compose.yaml`: Service orchestration
   - `Makefile`: Build automation
   - `.github/`: CI/CD workflows

4. **Documentation**
   - `README.md`: Project overview and setup instructions
   - `devNotes.md`: Developer documentation and notes
   - `LICENSE`: Project license information

### Service Interactions

1. **Scheduler Services**
   - Schedules new jobs, and monitors triggers for them
   - Assigns performer role to one among the pool of keepers
   - Sends task to performer
   - Monitors P2P network for that status of the tasks assigned

2. **Registrar Service**
   - Listens for Keeper and Task events on L1 and L2 respectively
   - Updates their status in the database

3. **Health Service**
   - Monitors keeper online status
   - Alerts operators if Keeper goes offline for >10m

4. **Database API Server**
   - Provides data persistence for all services

5. **Keeper Node**
   - API Server for Attester Node:
     - Executes assigned tasks
     - Validates executed tasks
   - Regular health check ins

## Rename the Services ?

I think we should rename the services to make it more intuitive and exciting for developer users.

- `manager` -> `engine` [Handles scheduling, load balancing, and task assignment, "Engine" appropriately conveys its central, driving role in the system]
- `registrar` -> `sentinel` [A sentinel watches over something important, which aligns with tracking keeper registration and actions]
- `health` -> `pulse` [This service actively monitors keeper health and online status, this name conveys the heartbeat-like monitoring functions]
- `dbserver` -> `librarian` [Don't know, couldn't think of a good name]
- `attester` -> `nexus` [This service connects to the keeper p2p network and relays events to other services, it serves as a crucial interconnection point]
