# Notes for Developers

## Repository Architecture

```bash
triggerx-backend/
├── .othentic/                  # Othentic Contracts network configurations
├── bin/                        # Compiled binaries
├── cmd/                        # Main application entry points
├── config/                     # Configuration files
├── data/                       # Data storage and persistence
├── docker/                     # Docker configuration files
├── docs/                       # Documentation
├── internal/                   # Internal packages and shared code
│   ├── dbserver/               # Database server internals
│   ├── eventmonitor/           # Event monitor internals
│   ├── health/                 # Health monitoring internals
│   ├── keeper/                 # Keeper node internals
│   ├── schedulers/             # Schedulers (Time, Condition, Event)
│   ├── taskdispatcher/         # Task Dispatcher internals
│   └── taskmonitor/            # Task Monitor internals
├── othentic/                   # Othentic Aggregator implementation
├── pkg/                        # Public packages and shared libraries
│   ├── client/                 # External service clients
│   ├── database/               # Database utilities
│   ├── dockerexecutor/         # Docker execution logic
│   ├── observability/          # Monitoring and tracing
│   └── ...                     # Other shared utilities
├── scripts/                    # Utility scripts and tools
├── v_collect_data/             # Data collection scripts/tools
├── w_keeper_setup/             # Keeper setup scripts
├── x_frontend/                 # Frontend application
├── y_contracts/                # Smart contracts
├── z_sdk/                      # TypeScript SDK
├── justfile                    # Task runner configuration
├── Dockerfile                  # Keeper container build configuration
├── docker-compose.yaml         # Multi-container orchestration
└── go.mod                      # Go module definition
```

### Key Components

1. **Core Services** (in `cmd/` and `internal/`)

   - `schedulers/`: Handles job scheduling (Time, Condition, Event based)
   - `taskdispatcher/`: Manages task queue and assignment to Keepers
   - `taskmonitor/`: Tracks task execution and handles failures
   - `eventmonitor/`: Monitors blockchain events
   - `health/`: Monitors keeper health and online status
   - `dbserver/`: Central API server for data persistence
   - `keeper/`: Keeper node implementation for task execution and validation

2. **Network Layer** (`othentic/`)

   - Implements the Aggregator functionality using Othentic Stack
   - Handles consensus and blockchain submission

3. **Build and Deployment**

   - `Dockerfile`: Container configuration
   - `docker/`: Docker Compose files
   - `justfile`: Task runner commands
   - `.github/`: CI/CD workflows
   - `scripts/`: Deployment and maintenance scripts

4. **Documentation**
   - `README.md`: Project overview and setup instructions
   - `docs/`: Comprehensive documentation suite
