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
│   │   └── registrar.go
│   ├── manager/                 # Job management internals
│   │   ├── config/
│   │   ├── client/
│   │   ├── cache/
│   │   ├── scheduler/
│   │   │   ├── workers/
│   │   │   ├── scheduler.go
│   │   └── manager.go
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
│   ├── converter/
│   ├── database/
│   ├── logging/
│   ├── proof/
│   ├── redis/
│   ├── resources/
│   ├── types/
│   └── validator/
├── scripts/                     # Utility scripts and tools
├── Dockerfile                   # Keeper container build configuration
├── docker-compose.yaml          # Multi-container orchestration
├── go.mod                       # Go module definition
├── Makefile                     # Build and development commands
```

### Key Components

1. **Core Services** (in `cmd/` and `internal/`)
   - `manager/`: Handles job scheduling, load balancing, and task assignment
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

1. **Manager Service**
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

## Database Schema

### User Data

- `user_id`: Auto Incremented ID of the User
- `user_address`: Wallet Address of the User
- `created_at`: Timestamp of the User Creation
- `job_ids`: Set of Job IDs Created by the User
- `account_balance`: Points Balance of the User
- `token_balance`: Staked Token Balance of the User
- `last_updated_at`: Timestamp of the Last Update

### Job Data

- `job_id`: Auto Incremented ID of the Job
- `task_definition_id`: ID of the Task Definition
- `user_id`: ID of the User who created the Job
- `priority`: Priority for the Job (Not Implemented)
  - 1 = High, 2 = Medium, 3 = Low
  - Priority is used to determine the base fee for the task. Not implemented as of now.
- `security`: Security of the Job (Not Implemented)
  - 1 = High, 2 = Medium, 3 = Low, 4 = Lowest
  - Security is used to determine the voting power for the job. Not implemented as of now.
- `link_job_id`: ID of the Linked Job
  - -1 = None, n = Linked Job ID
  - If the job is linked to another job, scheduler will schedule the linked job only when the current job is completed.
- `chain_status`: Status of the Job being part of interlinked jobs
  - 0 = None, 1 = Chain Head, 2 = Chain Block
  - If the job is head of a chain, we schedule only head. Rest is handled by scheduler.
- `time_frame`: Time Frame of the Job will be valid for
- `recurring`: Recurring (true) means the job (Event / Condition) will be scheduled again and again until TimeFrame is reached.
- `time_interval`: Time Interval of the Time Based Job
- `trigger_chain_id`: Chain ID of the Trigger to look for (Event / Condition)
- `trigger_contract_address`: Contract Address where the Trigger Event is located
- `trigger_event`: Trigger Event Signature
- `script_ipfs_url`: IPFS URL of the Script User Submitted
- `script_trigger_function`: Trigger Function in the Script, ran by Manager to check for Trigger
- `script_target_function`: Target Function in the Script, ran by Keeper when Trigger is detected, the Action to be taken
- `status`: Status of the Job
  - 0 = Created, 1 = Scheduled, 2 = Completed, 3 = Failed, 4 = Paused / Deleted
- `job_cost_prediction`: Cost Prediction of the Job (only approximation, not used in resources calculation)
- `created_at`: Timestamp of the Job Creation
- `last_executed_at`: Timestamp of the Last Execution
- `task_ids`: Set of Task IDs executed for the Job

### Task Data

- `task_id`: Auto Incremented ID of the Task (on database)
- `task_number`: Auto Incremented Number of the Task (on Attestation Center contract)
- `job_id`: ID of the Job
- `task_definition_id`: Task Definition ID
  - 1 = StartTimeBasedJob (Static Execution)
  - 2 = StartTimeBasedJob (Dynamic Execution)
  - 3 = StartEventBasedJob (Static Execution)
  - 4 = StartEventBasedJob (Dynamic Execution)
  - 5 = StartConditionBasedJob (Static Execution)
  - 6 = StartConditionBasedJob (Dynamic Execution)
  - n = Custom Task (Not Implemented)
  - Static means we are not running any off chain computation, we are just calling a function with static arguments (like a normal function call)
  - Dynamic means we are running some off chain computation, we are calling a function with dynamic arguments.
- `created_at`: Timestamp of the Task Creation
- `task_cost`: Cost of the Task, incurred by the Keeper for executing the Task
- `execution_timestamp`: Timestamp of the Action contract call
- `execution_tx_hash`: Transaction Hash of the Action contract call
- `task_performer_id`: ID of the Keeper who performed the Task
- `proof_of_task`: TLS Proof of Task
- `action_data_cid`: CID of the Action Data (IPFS) (Not useful, will be removed in future)
- `task_attester_ids`: List of Task Attester IDs
- `is_approved`: Was the Task Approved by the Attesters
- `tp_signature`: Task Performer Signature
- `ta_signature`: Task Attesters Aggregate Signature
- `task_submission_tx_hash`: Transaction Hash of the Task Submission
- `is_successful`: Is the Task Successfully Executed

### Keeper Data

- `keeper_id`: Auto Incremented ID of the Keeper (added when a new keeper is encountered, from registration, form, or peer discovery)
- `keeper_name`: Name of the Keeper from the Form
- `keeper_address`: Wallet Address of the Keeper (on EigenLayer and TriggerX)
- `registered_tx`: Transaction Hash of the Keeper Registration to TriggerX
- `operator_id`: Index of the Operator (Keeper) on the Attestation Center contract
- `rewards_address`: Address to receive rewards (from TriggerX Treasury, Not Implemented)
- `rewards_booster`: Multiplier for Rewards (2 for those whitelisted and registered before 15/04/2025, 1 for others, may introduce more multipliers in future)
- `voting_power`: Voting Power of the Keeper
- `keeper_points`: Sum of all the `task_cost` of the Tasks performed by the Keeper
- `connection_address`: Public IP Address of the Keeper
- `peer_id`: Peer ID of the Keeper on Othentic P2P Network
- `strategies`: Strategies in which Keeper has Stakes
- `verified`: Is the Keeper Whitelisted (form)
- `status`: Is the Keeper Registered (AVSG and Attestation Center)
- `online`: Is the Keeper Online (Health Check)
- `version`: Version of the Keeper Execution Binary
- `no_exctask`: Number of Tasks Performed
- `chat_id`: Telegram Chat ID of the Keeper for Notifications
- `email_id`: Email ID of the Keeper for Notifications

### ApiKey Data

- `key`: Key of the API
- `owner`: Owner of the API
- `isActive`: Is the API Active
- `rateLimit`: Rate Limit of the API
- `lastUsed`: Last Used of the API
- `createdAt`: Timestamp of the API Creation

## Rename the Services ?

I think we should rename the services to make it more intuitive and exciting for developer users.

- `manager` -> `engine` [Handles scheduling, load balancing, and task assignment, "Engine" appropriately conveys its central, driving role in the system]
- `registrar` -> `sentinel` [A sentinel watches over something important, which aligns with tracking keeper registration and actions]
- `health` -> `pulse` [This service actively monitors keeper health and online status, this name conveys the heartbeat-like monitoring functions]
- `dbserver` -> `librarian` [Don't know, couldn't think of a good name]
- `attester` -> `nexus` [This service connects to the keeper p2p network and relays events to other services, it serves as a crucial interconnection point]
