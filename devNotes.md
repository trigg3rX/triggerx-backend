# Notes for Developers

## Database

### Job Data

- `TaskDefinitionID`: 
  - 1 = StartTimeBasedJob (Static Execution)
  - 2 = StartTimeBasedJob (Dynamic Execution)
  - 3 = StartEventBasedJob (Static Execution)
  - 4 = StartEventBasedJob (Dynamic Execution)
  - 5 = StartConditionBasedJob (Static Execution)
  - 6 = StartConditionBasedJob (Dynamic Execution)

  Static means we are not running any off chain computation, we are just calling a function with static arguments.
  Dynamic means we are running some off chain computation, we are calling a function with dynamic arguments.

- `chain_status`: 0 = None, 1 = Chain Head, 2 = Chain Block 

    If the job is head of a chain, we schedule only head. Rest is handled by scheduler.

- `link_job_id`: -1 = None, n = Linked Job ID

    If the job is linked to another job, scheduler will schedule the linked job only when the current job is completed.

- `priority`: 1 = High, 2 = Medium, 3 = Low

    Priority is used to determine the base fee for the task. Not implemented as of now.

- `security`: 1 = High, 2 = Medium, 3 = Low, 4 = Lowest

    Security is used to determine the voting power for the job. Not implemented as of now.

- `recurring`: true = Recurring, false = Non-Recurring

    Recurring means the job will be scheduled again and again until TimeFrame is reached.

## Rename the Services ?

I think we should rename the services to make it more intuitive and exciting for developer users.

- `manager` -> `engine` [Handles scheduling, load balancing, and task assignment, "Engine" appropriately conveys its central, driving role in the system]
- `registrar` -> `sentinel` [A sentinel watches over something important, which aligns with tracking keeper registration and actions]
- `health` -> `pulse` [This service actively monitors keeper health and online status, this name conveys the heartbeat-like monitoring functions]
- `dbserver` -> `librarian` [Don't know, couldn't think of a good name]
- `attester` -> `nexus` [This service connects to the keeper p2p network and relays events to other services, it serves as a crucial interconnection point]