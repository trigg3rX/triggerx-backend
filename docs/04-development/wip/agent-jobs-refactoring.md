# Agent Jobs Architecture Refactoring Plan (TDI 7 → 7/8/9)

**Document Version:** 1.0  
**Date:** 2026-01-13  
**Status:** Planning Phase  
**Target:** Make TDI 7 infrastructure reusable for TDI 8 & 9

---

## Table of Contents

1. [Executive Summary](#1-executive-summary)
2. [Current State Analysis](#2-current-state-analysis)
3. [Refactoring Goals](#3-refactoring-goals)
4. [Code Audit](#4-code-audit)
5. [Refactoring Plan](#5-refactoring-plan)
6. [Implementation Phases](#6-implementation-phases)
7. [Testing Strategy](#7-testing-strategy)
8. [Success Criteria](#8-success-criteria)
9. [Risk Assessment](#9-risk-assessment)
10. [Rollback Plan](#10-rollback-plan)
11. [Appendix](#11-appendix)

---

## 1. Executive Summary

### Problem

The current TDI 7 (custom time-based jobs) implementation has hardcoded references and naming that assumes only time-based triggers. To support TDI 8 (agent event-based) and TDI 9 (agent condition-based), we need to refactor the code to be trigger-agnostic.

### Solution

Refactor TDI 7 infrastructure into a generic "Agent Jobs" system that can handle any trigger type (time, event, condition) while maintaining the same execution model:

- Script decides `shouldExecute`
- Script builds full `calldata`
- Persistent storage support
- Challenge/fraud-proof infrastructure

### Timeline

- **Phase 1:** Rename & Refactor (2 weeks)
- **Phase 2:** Add TDI 8 Support (1 week)
- **Phase 3:** Add TDI 9 Support (1 week)
- **Phase 4:** Testing & Documentation (1 week)

---

## 2. Current State Analysis

### 2.1 What Works ✅

| Component                | Status     | Notes                                        |
|--------------------------|------------|----------------------------------------------|
| Storage System           | ✅ Generic | `script_storage` table is job-agnostic       |
| Docker Execution         | ✅ Generic | Already supports multiple TDIs via metadata  |
| Script Output Format     | ✅ Generic | `CustomScriptOutput` works for all triggers  |
| Fee Calculation          | ✅ Generic | Works with any TDI via `task_definition_id`  |
| Challenge Infrastructure | ✅ Generic | `custom_script_executions` can track any TDI |

### 2.2 What Needs Refactoring ❌

#### A. Naming Issues

**Files:**

```bash
internal/keeper/core/execution/custom_executor.go
  → Should be: agent_executor.go
```

**Functions:**

```bash
ExecuteCustomScript()
  → Should be: ExecuteAgentScript()

CreateCustomJob()
  → Should be: CreateAgentJob()

GetCustomJobByID()
  → Should be: GetAgentJobByID()
```

**Types:**

```go
CustomJobData
  → Should be: AgentJobData

CustomJobRepository
  → Should be: AgentJobRepository
```

**Database Tables:**

```sql
custom_jobs
  → Should be: agent_jobs

custom_script_executions
  → Should be: agent_script_executions
```

#### B. Hardcoded TDI 7 Checks

**File: `internal/keeper/core/execution/action.go`**

❌ **Line 25:**

```go
if targetData.TaskDefinitionID != 7 && targetData.TargetContractAddress == "" {
```

Should be:

```go
if !isAgentJob(targetData.TaskDefinitionID) && targetData.TargetContractAddress == "" {
```

❌ **Line 52-55:**

```go
// Skip ABI parsing for custom scripts (TaskDefinitionID 7)
var contractABI *abi.ABI
var method *abi.Method
if targetData.TaskDefinitionID != 7 {
```

Should be:

```go
// Skip ABI parsing for agent jobs (TDI 7, 8, 9)
var contractABI *abi.ABI
var method *abi.Method
if !isAgentJob(targetData.TaskDefinitionID) {
```

❌ **Line 68:**

```go
case 7:
    // Custom script execution (TaskDefinitionID = 7)
```

Should be:

```go
case 7, 8, 9:
    // Agent job execution (TDI 7=time, 8=event, 9=condition)
```

#### C. Missing Trigger Type Abstraction

**Current:** TDI 7 is time-based only  
**Needed:** Generic trigger type field

```go
// Add to AgentJobData
type AgentJobData struct {
    JobID            string `json:"job_id"`
    TaskDefinitionID int     `json:"task_definition_id"` // 7, 8, or 9
    TriggerType      string  `json:"trigger_type"`       // "time", "event", "condition"
    // ... rest
}
```

#### D. Scheduler Integration

**Current:** Time scheduler has custom job support  
**Needed:** Event & Condition schedulers need agent job support

```go
// internal/schedulers/time/scheduler/poll.go (✅ Already exists)
if s.customJobRepository != nil {
    customJobs, err := s.customJobRepository.GetCustomJobsDueForExecution(lookAheadTime)
    // ...
}

// internal/schedulers/event/scheduler/poll.go (❌ Needs adding)
if s.agentJobRepository != nil {
    agentEventJobs := s.agentJobRepository.GetAgentJobsByTriggerType("event")
    // ...
}

// internal/schedulers/condition/scheduler/poll.go (❌ Needs adding)
if s.agentJobRepository != nil {
    agentConditionJobs := s.agentJobRepository.GetAgentJobsByTriggerType("condition")
    // ...
}
```

---

## 3. Refactoring Goals

### 3.1 Primary Goals

1. **Zero Breaking Changes** - Existing TDI 7 jobs continue working
2. **Generic Naming** - No "custom" references, use "agent" instead
3. **Trigger Agnostic** - Code doesn't assume time-based triggers
4. **Unified Table** - Single `agent_jobs` table for TDI 7, 8, 9
5. **Helper Functions** - `isAgentJob(tdi)` instead of `tdi == 7`

### 3.2 Secondary Goals

1. **Better Logging** - Log trigger type explicitly
2. **Metrics Separation** - Track TDI 7/8/9 separately
3. **Documentation** - Clear distinctions between trigger types
4. **Testing** - Unit tests for each TDI independently

---

## 4. Code Audit

### 4.1 Files Requiring Changes

#### Tier 1: Critical (Core Execution)

```bash
internal/keeper/core/execution/
  ├── action.go                    [REFACTOR - Add TDI 8/9 support]
  ├── custom_executor.go           [RENAME → agent_executor.go]
  └── utils.go                     [ADD - Helper functions]
```

#### Tier 2: Database Layer (Implemented in the database schema file)

```bash
scripts/database/init-db.cql
```

#### Tier 3: Scheduler Integration

```bash
internal/schedulers/
  ├── time/scheduler/
  │   ├── poll.go                  [UPDATE - Use new naming]
  │   └── converter.go             [UPDATE - Generic conversion]
  └── condition/scheduler/
      └── core/scheduler
```

#### Tier 4: Task Monitor

```bash
internal/taskmonitor/
  ├── core/events/task.go               [UPDATE - Handle TDI 8/9]
  └── database/repository/     [UPDATE - Storage for all agents]
```

#### Tier 5: Types & Models

```bash
pkg/types/
  ├── custom_execution.go          [RENAME → agent_execution.go]
  └── schedulers.go                [UPDATE - Generic task data]
```

### 4.2 Files NOT Requiring Changes ✅

```bash
pkg/dockerexecutor/               ✅ Already supports any TDI
```
