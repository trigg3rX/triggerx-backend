# Notes for Developers

## Database Schema

### User Data

- `user_id`: [bigint] Auto Incremented ID of the User
- `user_address`: [text] Wallet Address of the User
- `created_at`: [timestamp] Timestamp of the User Creation
- `job_ids`: [set<bigint>] Set of Job IDs Created by the User
- `account_balance`: [varint] Points Balance of the User
- `token_balance`: [varint] Staked Token Balance of the User
- `last_updated_at`: [timestamp] Timestamp of the Last Update

### Job Data

- `job_id`: [bigint] Auto Incremented ID of the Job
- `task_definition_id`: [int] ID of the Task Definition
- `user_id`: [bigint] ID of the User who created the Job
- `priority`: [int] Priority for the Job (Not Implemented)
  - 1 = High, 2 = Medium, 3 = Low
  - Priority is used to determine the base fee for the task. Not implemented as of now.
- `security`: [int] Security of the Job (Not Implemented)
  - 1 = High, 2 = Medium, 3 = Low, 4 = Lowest
  - Security is used to determine the voting power for the job. Not implemented as of now.
- `link_job_id`: [bigint] ID of the Linked Job
  - -1 = None, n = Linked Job ID
  - If the job is linked to another job, scheduler will schedule the linked job only when the current job is completed.
- `chain_status`: [int] Status of the Job being part of interlinked jobs
  - 0 = None, 1 = Chain Head, 2 = Chain Block
  - If the job is head of a chain, we schedule only head. Rest is handled by scheduler.
- `time_frame`: [bigint] Time Frame of the Job will be valid for
- `recurring`: [boolean] Recurring (true) means the job (Event / Condition) will be scheduled again and again until TimeFrame is reached.
- `time_interval`: [bigint] Time Interval of the Time Based Job
- `trigger_chain_id`: [text] Chain ID of the Trigger to look for (Event / Condition)
- `trigger_contract_address`: [text] Contract Address where the Trigger Event is located
- `trigger_event`: [text] Trigger Event Signature
- `script_ipfs_url`: [text] IPFS URL of the Script User Submitted
- `script_trigger_function`: [text] Trigger Function in the Script, ran by Manager to check for Trigger
- `script_target_function`: [text] Target Function in the Script, ran by Keeper when Trigger is detected, the Action to be taken
- `status`: [int] Status of the Job
  - 0 = Created, 1 = Scheduled, 2 = Completed, 3 = Failed, 4 = Paused / Deleted
- `job_cost_prediction`: [double] Cost Prediction of the Job (only approximation, not used in resources calculation)
- `created_at`: [timestamp] Timestamp of the Job Creation
- `last_executed_at`: [timestamp] Timestamp of the Last Execution
- `task_ids`: [set<bigint>] Set of Task IDs executed for the Job

### Task Data

- `task_id`: [bigint] Auto Incremented ID of the Task (on database)
- `task_number`: [int] Auto Incremented Number of the Task (on Attestation Center contract)
- `job_id`: [bigint] ID of the Job
- `task_definition_id`: [int] Task Definition ID
  - 1 = StartTimeBasedJob (Static Execution)
  - 2 = StartTimeBasedJob (Dynamic Execution)
  - 3 = StartEventBasedJob (Static Execution)
  - 4 = StartEventBasedJob (Dynamic Execution)
  - 5 = StartConditionBasedJob (Static Execution)
  - 6 = StartConditionBasedJob (Dynamic Execution)
  - n = Custom Task (Not Implemented)
  -   Static means we are not running any off chain computation, we are just calling a function with static arguments (like a normal function call)
  -   Dynamic means we are running some off chain computation, we are calling a function with dynamic arguments.
- `created_at`: [timestamp] Timestamp of the Task Creation
- `task_cost`: [double] Cost of the Task, incurred by the Keeper for executing the Task
- `execution_timestamp`: [timestamp] Timestamp of the Action contract call
- `execution_tx_hash`: [text] Transaction Hash of the Action contract call
- `task_performer_id`: [bigint] ID of the Keeper who performed the Task
- `proof_of_task`: [text] TLS Proof of Task
- `action_data_cid`: [text] CID of the Action Data (IPFS) (Not useful, will be removed in future)
- `task_attester_ids`: [list<bigint>] List of Task Attester IDs
- `is_approved`: [boolean] Was the Task Approved by the Attesters
- `tp_signature`: [blob] Task Performer Signature
- `ta_signature`: [blob] Task Attesters Aggregate Signature
- `task_submission_tx_hash`: [text] Transaction Hash of the Task Submission
- `is_successful`: [boolean] Is the Task Successfully Executed

### Keeper Data

- `keeper_id`: [bigint] Auto Incremented ID of the Keeper (added when a new keeper is encountered, from registration, form, or peer discovery)
- `keeper_name`: [text] Name of the Keeper from the Form
- `keeper_address`: [text] Wallet Address of the Keeper (on EigenLayer and TriggerX)
- `registered_tx`: [text] Transaction Hash of the Keeper Registration to TriggerX
- `operator_id`: [text] Index of the Operator (Keeper) on the Attestation Center contract
- `rewards_address`: [text] Address to receive rewards (from TriggerX Treasury, Not Implemented)
- `rewards_booster`: [double] Multiplier for Rewards (2 for those whitelisted and registered before 15/04/2025, 1 for others, may introduce more multipliers in future)
- `voting_power`: [text] Voting Power of the Keeper
- `keeper_points`: [double] Sum of all the `task_cost` of the Tasks performed by the Keeper
- `connection_address`: [text] Public IP Address of the Keeper
- `peer_id`: [text] Peer ID of the Keeper on Othentic P2P Network
- `strategies`: [list<text>] Strategies in which Keeper has Stakes
- `verified`: [boolean] Is the Keeper Whitelisted (form)
- `status`: [boolean] Is the Keeper Registered (AVSG and Attestation Center)
- `online`: [boolean] Is the Keeper Online (Health Check)
- `version`: [text] Version of the Keeper Execution Binary
- `no_exctask`: [int] Number of Tasks Performed
- `chat_id`: [bigint] Telegram Chat ID of the Keeper for Notifications
- `email_id`: [text] Email ID of the Keeper for Notifications

### ApiKey Data

- `key`: [text] Key of the API
- `owner`: [text] Owner of the API
- `isActive`: [boolean] Is the API Active
- `rateLimit`: [int] Rate Limit of the API
- `lastUsed`: [timestamp] Last Used of the API
- `createdAt`: [timestamp] Timestamp of the API Creation


## Rename the Services ?

I think we should rename the services to make it more intuitive and exciting for developer users.

- `manager` -> `engine` [Handles scheduling, load balancing, and task assignment, "Engine" appropriately conveys its central, driving role in the system]
- `registrar` -> `sentinel` [A sentinel watches over something important, which aligns with tracking keeper registration and actions]
- `health` -> `pulse` [This service actively monitors keeper health and online status, this name conveys the heartbeat-like monitoring functions]
- `dbserver` -> `librarian` [Don't know, couldn't think of a good name]
- `attester` -> `nexus` [This service connects to the keeper p2p network and relays events to other services, it serves as a crucial interconnection point]