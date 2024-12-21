# TriggerX Go Backend

Backend Utilities and APIs for TriggerX

## Architecture Overview

### Directory Structure

```
triggerx-backend/
├── cmd/                    # Application entry points
│   ├── keeper/            # Keeper service entry point
│   ├── quorum/     # Quorum creation service
│   ├── manager/       # Task management service
│   └── validator/     # Task validation service
│
├── internal/              # Private application code
│   ├── keeper/           # Keeper service implementation
│   ├── quorum/    # Quorum creation logic
│   ├── manager/      # Task management implementation
│   └── validator/    # Task validation logic
│
├── pkg/                  # Public shared libraries
│   ├── communication/    # Network communication utilities
│   └── database/        # Database interactions
│
└── scripts/               # Utility scripts
    └── start-*.sh      # Service startup scripts
```

### Component Overview

#### Core Services

1. **Keeper Service** (`cmd/keeper/`)
   - Manages distributed keeper nodes
   - Handles leader and worker coordination
   - Aggregates data from multiple sources

2. **Quorum Creator** (`cmd/quorumcreator/`)
   - Establishes and maintains quorum requirements
   - Handles consensus mechanisms
   - Manages node participation

3. **Manager** (`cmd/manager/`)
   - Orchestrates job and task distribution
   - Implements load balancing
   - Manages thread pools and resources

4. **Validator** (`cmd/validator/`)
   - Validates execution results
   - Ensures data integrity
   - Submits results on-chain

#### Shared Packages

1. **Communication** (`pkg/communication/`)
   - P2P Network setup and configuration
   - Peer discovery mechanisms
   - Data transmission protocols

2. **Database** (`pkg/database/`)
   - ScyllaDB interactions
   - Query operations

### Technology Stack

- **Language:** Go
- **Database:** ScyllaDB

### Getting Started

[Add installation and setup instructions here]
