# Agent Jobs Database Schema (TDI 7, 8, 9)

**Date:** 2026-01-16  
**Schema File:** `scripts/database/init-db.cql`  
**Architecture:** Unified tables with nullable fields (ScyllaDB optimized)

---

## Overview

Agent jobs (TDI 7, 8, 9) extend the TriggerX system with **programmable execution logic**. Instead of predefined target contracts and functions, agent jobs use custom scripts that:

1. **Decide** whether to execute (`shouldExecute`)
2. **Determine** the target contract
3. **Build** the full calldata
4. **Maintain** persistent state across executions

### Key Design Decision: Unified Tables ✨

**Instead of creating separate tables for agent jobs, we extend the existing tables:**

- `time_job_data` → TDI 1, 2, **7** (6 agent fields added)
- `event_job_data` → TDI 3, 4, **8** (6 agent fields added)
- `condition_job_data` → TDI 5, 6, **9** (6 agent fields added)

**Rationale:**

- Same trigger mechanism = same scheduler = same table
- Only 6 additional nullable columns per table
- `task_definition_id` distinguishes traditional vs agent
- Simpler codebase: one repository, one query pattern per trigger type

---

## Table Structure

### Job Tables (Unified Traditional + Agent)

| Table                | TDIs        | Description                                |
| -------------------- | ----------- | ------------------------------------------ |
| `time_job_data`      | 1, 2, **7** | Time/cron-based jobs (traditional + agent) |
| `event_job_data`     | 3, 4, **8** | Event-based jobs (traditional + agent)     |
| `condition_job_data` | 5, 6, **9** | Condition-based jobs (traditional + agent) |

### Supporting Tables (Agent-Specific)

| Table                     | Purpose                      | Used By     |
| ------------------------- | ---------------------------- | ----------- |
| `agent_script_executions` | Fraud-proof tracking         | TDI 7, 8, 9 |
| `script_storage`          | Persistent key-value storage | TDI 7, 8, 9 |
| `execution_challenges`    | Challenge tracking           | TDI 7, 8, 9 |

---

## Schema Details

### `time_job_data` (TDI 1, 2, 7)

```sql
CREATE TABLE time_job_data (
    job_id varint PRIMARY KEY,
    task_definition_id int,            -- 1, 2, or 7

    -- Time scheduling (common for all)
    schedule_type text,
    time_interval bigint,
    cron_expression text,
    next_execution_timestamp timestamp,
    timezone text,

    -- Traditional job fields (TDI 1, 2) - NULL for agent jobs
    target_chain_id text,
    target_contract_address text,
    target_function text,
    abi text,
    arg_type int,
    arguments list<text>,
    dynamic_arguments_script_url text,

    -- Agent job fields (TDI 7) - NULL for traditional jobs
    custom_script_url text,            -- IPFS URL
    script_language text,              -- 'ts', 'go', 'python', 'javascript'
    script_hash text,                  -- keccak256(scriptCode)
    agent_target_chain_id text,        -- Default chain
    max_execution_time int,            -- Timeout (default 60s)
    challenge_period bigint,           -- Challenge period (default 6 hours)

    -- Common status fields
    is_active boolean,
    is_completed boolean,
    expiration_time timestamp,
    ...
);
```

**Field Usage by TDI:**

| Field Group               | TDI 1, 2 | TDI 7   |
| ------------------------- | -------- | ------- |
| Scheduling fields         | ✅ Set   | ✅ Set  |
| Traditional target fields | ✅ Set   | ❌ NULL |
| Agent script fields       | ❌ NULL  | ✅ Set  |

### `event_job_data` (TDI 3, 4, 8)

```sql
CREATE TABLE event_job_data (
    job_id varint PRIMARY KEY,
    task_definition_id int,            -- 3, 4, or 8
    recurring boolean,

    -- Event trigger (common for all)
    trigger_chain_id text,
    trigger_contract_address text,
    trigger_event text,
    event_filter_para_name text,
    event_filter_value text,

    -- Traditional job fields (TDI 3, 4) - NULL for agent jobs
    target_chain_id text,
    target_contract_address text,
    target_function text,
    abi text,
    arg_type int,
    arguments list<text>,
    dynamic_arguments_script_url text,

    -- Agent job fields (TDI 8) - NULL for traditional jobs
    custom_script_url text,
    script_language text,
    script_hash text,
    agent_target_chain_id text,
    max_execution_time int,
    challenge_period bigint,

    -- Common status fields
    is_active boolean,
    is_completed boolean,
    ...
);
```

### `condition_job_data` (TDI 5, 6, 9)

```sql
CREATE TABLE condition_job_data (
    job_id varint PRIMARY KEY,
    task_definition_id int,            -- 5, 6, or 9
    recurring boolean,

    -- Condition trigger (common for all)
    condition_type text,
    upper_limit double,
    lower_limit double,
    value_source_type text,
    value_source_url text,
    selected_key_route text,

    -- Traditional job fields (TDI 5, 6) - NULL for agent jobs
    target_chain_id text,
    target_contract_address text,
    target_function text,
    abi text,
    arg_type int,
    arguments list<text>,
    dynamic_arguments_script_url text,

    -- Agent job fields (TDI 9) - NULL for traditional jobs
    custom_script_url text,
    script_language text,
    script_hash text,
    agent_target_chain_id text,
    max_execution_time int,
    challenge_period bigint,

    -- Common status fields
    is_active boolean,
    is_completed boolean,
    ...
);
```

### `agent_script_executions`

Tracks every execution for fraud-proof verification (TDI 7, 8, 9 only):

```sql
CREATE TABLE agent_script_executions (
    execution_id text PRIMARY KEY,
    job_id varint,
    task_id bigint,
    task_definition_id int,            -- 7, 8, or 9

    -- Input (for deterministic re-execution)
    input_timestamp bigint,
    input_storage text,                -- JSON snapshot
    input_hash text,
    trigger_data text,                 -- Trigger-specific data

    -- Output
    should_execute boolean,
    target_contract text,
    calldata text,
    output_hash text,

    -- Metadata (for verification)
    execution_metadata text,
    script_hash text,
    signature text,

    -- Result
    tx_hash text,
    execution_status text,

    -- Challenge tracking
    verification_status text,
    challenge_deadline timestamp,
    is_challenged boolean,
    challenge_count int,
    ...
);
```

**`trigger_data` format by TDI:**

```json
// TDI 7 (Time)
{"scheduled_time": "2026-01-14T10:00:00Z"}

// TDI 8 (Event)
{
  "event_data": {"from": "0x...", "amount": "1000000"},
  "block_number": 12345678,
  "tx_hash": "0xabc...",
  "event_signature": "Transfer(address,address,uint256)"
}

// TDI 9 (Condition)
{
  "condition_value": 2050.0,
  "condition_type": "price",
  "timestamp": "2026-01-14T10:05:00Z"
}
```

---

## Execution Flow

### Unified Scheduler Pattern

**Key Insight:** Same scheduler handles traditional AND agent jobs!

```
┌──────────────────────────────────────────────────────────┐
│ TIME SCHEDULER                                           │
│ SELECT * FROM time_job_data                              │
│ WHERE is_active = true AND next_execution_timestamp <= NOW() │
│                                                          │
│ Returns: TDI 1, 2, AND 7 jobs                           │
└──────────────────────────────────────────────────────────┘
                          ↓
┌──────────────────────────────────────────────────────────┐
│ KEEPER RECEIVES TASK                                     │
│ Check task_definition_id:                                │
│   IF TDI == 1 or 2 → Execute traditional flow            │
│   IF TDI == 7 → Execute agent flow                       │
└──────────────────────────────────────────────────────────┘
```

### Agent Flow (TDI 7, 8, 9)

```
Keeper receives task with TDI 7/8/9
  ↓
Fetch job data (time_job_data / event_job_data / condition_job_data)
  ↓
Fetch script from IPFS (custom_script_url)
  ↓
Fetch persistent storage (script_storage)
  ↓
Run script in Docker sandbox
  ↓
Script output: {shouldExecute, targetContract, calldata, storageUpdates}
  ↓
If shouldExecute = true:
  - Submit transaction
  - Record in agent_script_executions
  - Update script_storage
Else:
  - Skip execution (still log in agent_script_executions)
  ↓
Challenge period begins (6 hours default)
```

---

## Query Patterns

### Scheduler Polling (Hot Path)

**The beauty: ONE query per trigger type, covers all TDIs!**

```sql
-- Time Scheduler (gets TDI 1, 2, AND 7)
SELECT * FROM time_job_data
WHERE is_active = true
  AND next_execution_timestamp <= currentTimestamp()
ALLOW FILTERING;

-- Event Monitor (gets TDI 3, 4, AND 8)
SELECT * FROM event_job_data
WHERE is_active = true
  AND trigger_chain_id = '8453'
ALLOW FILTERING;

-- Condition Monitor (gets TDI 5, 6, AND 9)
SELECT * FROM condition_job_data
WHERE is_active = true
ALLOW FILTERING;
```

### Execution History (Agent Jobs Only)

```sql
-- Get all executions for a job
SELECT * FROM agent_script_executions
WHERE job_id = 123
ALLOW FILTERING;

-- Get pending verifications
SELECT * FROM agent_script_executions
WHERE verification_status = 'pending'
ALLOW FILTERING;
```

### Persistent Storage (Agent Jobs Only)

```sql
-- Get all storage for a job
SELECT * FROM script_storage WHERE job_id = 123;

-- Update storage
UPDATE script_storage
SET storage_value = '2050.0', updated_at = currentTimestamp()
WHERE job_id = 123 AND storage_key = 'lastPrice';
```

---

## Key Differences: Traditional vs Agent Jobs

| Aspect                 | Traditional (TDI 1-6)                     | Agent (TDI 7-9)                           |
| ---------------------- | ----------------------------------------- | ----------------------------------------- |
| **Table**              | Same (time/event/condition_job_data)      | Same (time/event/condition_job_data)      |
| **Target Contract**    | Pre-defined (`target_contract_address`)   | Script decides                            |
| **Function Call**      | Pre-defined (`target_function`, `abi`)    | Script decides                            |
| **Arguments**          | Static or script-generated (`arguments`)  | Script builds full calldata               |
| **Execution Decision** | Always execute                            | Script decides (`shouldExecute`)          |
| **Populated Fields**   | Traditional fields set, agent fields NULL | Agent fields set, traditional fields NULL |
| **Persistent Storage** | None                                      | `script_storage`                          |
| **Fraud Proofs**       | No                                        | `agent_script_executions`                 |
| **Challenge Period**   | No                                        | Yes (`challenge_period`)                  |

---

## Design Rationale

### Why Unified Tables?

✅ **Same trigger mechanism** - Time-based jobs (TDI 1, 2, 7) all use same scheduler  
✅ **Single query** - One scheduler query returns all relevant jobs  
✅ **Minimal overhead** - Only 6 nullable columns added  
✅ **Simpler codebase** - One repository interface, not three  
✅ **Backward compatible** - Existing TDI 1-6 unaffected  
✅ **Consistent pattern** - Matches existing TDI 1-6 architecture

### Why NOT Separate Tables?

❌ **Duplicated schedulers** - Would need separate scheduler for agent jobs  
❌ **Duplicated repositories** - Would need separate code for same trigger logic  
❌ **More complex** - Extra tables, extra maintenance  
❌ **No performance benefit** - ScyllaDB handles NULL columns efficiently

### Storage Overhead Analysis

**Per job, 6 additional columns:**

- `custom_script_url` (text) - ~100 bytes
- `script_language` (text) - ~10 bytes
- `script_hash` (text) - ~66 bytes
- `agent_target_chain_id` (text) - ~10 bytes
- `max_execution_time` (int) - 4 bytes
- `challenge_period` (bigint) - 8 bytes

**Total:** ~200 bytes per job (NULL if traditional)

**For 1M jobs:**

- Traditional jobs (TDI 1-6): 200 bytes × 1M = 200 MB (wasted, but negligible)
- Agent jobs (TDI 7-9): Populated as needed

**Verdict:** Storage cost is trivial compared to operational simplicity.

---

## Implementation Checklist

### Database

- [x] Schema updated (`init-db.cql`)
- [x] Agent fields added to existing tables
- [x] Execution tracking tables
- [x] Indexes maintained

### Go Code

- [ ] Update type definitions
  - [ ] Add agent fields to `TimeJobData`, `EventJobData`, `ConditionJobData`
  - [ ] Add `AgentScriptExecution` type
  - [ ] Add `ScriptStorage` type
- [ ] Update repositories
  - [ ] Extend existing repositories (no new repos needed!)
  - [ ] Add agent field population methods
- [ ] Update schedulers
  - [ ] Time scheduler: Handle TDI 7 in same query as TDI 1, 2
  - [ ] Event monitor: Handle TDI 8 in same query as TDI 3, 4
  - [ ] Condition monitor: Handle TDI 9 in same query as TDI 5, 6
- [ ] Update keeper
  - [ ] Check `task_definition_id` to route to agent flow
  - [ ] Implement agent execution logic
- [ ] Implement agent infrastructure
  - [ ] Script execution in Docker
  - [ ] Script storage management
  - [ ] Challenge infrastructure

### Testing

- [ ] Unit tests for agent fields
- [ ] Integration tests (traditional + agent in same table)
- [ ] Performance tests (query latency unchanged)
- [ ] Challenge flow tests

---

## Migration Notes

### From Old Schema

**If you had separate `agent_time_job_data`, `agent_event_job_data`, `agent_condition_job_data` tables:**

```sql
-- Migrate TDI 7 to unified time_job_data
INSERT INTO time_job_data (
    job_id, task_definition_id,
    schedule_type, time_interval, next_execution_timestamp, timezone,
    custom_script_url, script_language, script_hash,
    agent_target_chain_id, max_execution_time, challenge_period,
    is_active, is_completed, ...
)
SELECT * FROM agent_time_job_data;

-- Similar for TDI 8, 9...

-- Drop old tables
DROP TABLE agent_time_job_data;
DROP TABLE agent_event_job_data;
DROP TABLE agent_condition_job_data;
```

### Adding New Agent Jobs

```sql
-- TDI 7 (Agent Time)
INSERT INTO time_job_data (
    job_id, task_definition_id,
    schedule_type, next_execution_timestamp,
    custom_script_url, script_language, script_hash,
    agent_target_chain_id, max_execution_time, challenge_period,
    is_active, ...
) VALUES (
    123, 7,
    'interval', '2026-01-14T10:00:00Z',
    'ipfs://bafkrei...', 'typescript', '0xabc...',
    '8453', 60, 21600,
    true, ...
);
-- Note: target_contract_address, target_function, etc. are NULL
```

---

## Summary

### Unified Architecture Benefits

✅ **Separation by trigger, not execution** - Time/Event/Condition (not traditional/agent)  
✅ **Single scheduler per trigger** - No duplication  
✅ **Minimal schema changes** - 6 columns per table  
✅ **ScyllaDB optimized** - NULL columns have minimal overhead  
✅ **Consistent with existing design** - Matches TDI 1-6 pattern  
✅ **Simpler codebase** - One repository, one query pattern

### The 6 Magic Columns

Every agent job needs exactly these 6 fields:

1. `custom_script_url` - Where the script lives
2. `script_language` - How to execute it
3. `script_hash` - Fraud-proof verification
4. `agent_target_chain_id` - Default chain
5. `max_execution_time` - Timeout
6. `challenge_period` - Fraud-proof window

**Result:** Clean, simple, unified architecture! 🎉

---

**Next Steps:** Update Go types and repositories to populate agent fields when `task_definition_id ∈ {7, 8, 9}`.
