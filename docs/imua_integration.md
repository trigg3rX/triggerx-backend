# 🔗 TriggerX + Imua Integration Plan

This document outlines the architecture of **TriggerX**, a decentralized, multi-chain automation protocol, and how it integrates with **Imua’s Shared Security Protocol** using the `TriggerX AVS Contract`. This integration enhances task execution security, enables native restaking, and aligns with the modular AVS vision.

---

## 🎯 TriggerX's Purpose

TriggerX provides a truly decentralized and autonomous Keeper Network for developers and users across any blockchain network.

It leverages an **Actively Validated Service (AVS)** to secure job execution. The AVS layer ensures that automation tasks are executed by a decentralized set of incentivized operators, enforced through staking and slashing mechanisms. Integrating with AVS frameworks like Imua brings tamper-resistance, fault tolerance, and trustless execution—eliminating centralized assumptions.

---

## 🏗️ TriggerX Architecture

### Core Components

- **Jobs**: Created by users and composed of one or more **tasks**.
- **Tasks**: Executable units of work within a job, validated by AVS operators.
- **TriggerGas (TG)**: Represents resource consumption; deposited and tracked per job.

### Core Smart Contracts

| Contract       | Purpose                                                  |
|----------------|----------------------------------------------------------|
| `JobRegistry`  | Tracks jobs, their statuses, and associated task IDs     |
| `GasRegistry`  | Tracks TG balances and usage per job creator             |
| `TriggerXAvs`  | Interfaces with Imua’s AVS-Manager for task validation   |
| `ProxyHub`     | L2 job coordinator that validates keeper identities      |
| `ProxySpoke`   | L2 mirror of `ProxyHub`, deployed across other chains    |

---

## 🔄 Job Execution Flow

1. **Job Creation**
   - User creates a job and deposits TG.
   - Job and TG data are stored in `JobRegistry`, `GasRegistry`, and off-chain DB.

2. **Scheduling**
   - Off-chain DB forwards job metadata to appropriate schedulers (time/event/condition).
   - Upon trigger condition, scheduler sends execution payload to performer.

3. **Execution & Task Submission**
   - Performer executes the task, generates IPFS link with data for validators.
   - Calls `createTask()` on `TriggerXAvs`.
   - Task is registered with AVS; validators receive task details.

4. **Validation**
   - Validators verify trigger and action correctness.
   - Submit vote via `operatorSubmitTask()`.
   - On quorum:
     - Task is finalized.
     - `JobRegistry` is updated with task ID.
     - `GasRegistry` deducts consumed TG.

---

## 🛠️ Keeper Lifecycle (Imua Integrated)

1. **Registration**
   - Keeper restakes on Imua.
   - Registers via `registerOperatorToAVS()`.
   - Submits BLS public key.

2. **Authorization**
   - TriggerX admin whitelists the keeper.

3. **Cross-Chain Execution**
   - `ProxyHub`/`ProxySpoke` contracts are deployed across L2s.
   - User authorizes proxy contract execution.
   - Keeper submits calldata to proxy.
   - If validated, proxy calls target function.

4. **Validation & Voting**
   - Validators confirm execution integrity.
   - Vote via `operatorSubmitTask()`.
   - Task status resolved by consensus.

5. **Registry Updates**
   - On task finalization:
     - `JobRegistry` records task ID.
     - `GasRegistry` deducts TG used.

---

## 🔐 AVS Security Model with Imua

### 🔗 Integration Point: `TriggerX AVS Contract`

The `TriggerX AVS Contract` is a UUPS upgradeable proxy that interfaces with Imua’s AVS-Manager precompile (`0x000...0901`). It handles task execution, validator coordination, and operator management in a modular and extensible way.

---

### ⚙️ AVS Flow with Imua

```mermaid
sequenceDiagram
  participant Developer
  participant TriggerX Backend
  participant Operator
  participant ProxyHub.sol
  participant TriggerXAVS.sol
  participant Imua/AVSManager
  participant Job/GasRegistry

  Developer->>TriggerX Backend: createJob(params)
  TriggerX Backend->>Operator: executeTask(params)
  Operator->>ProxyHub.sol: call executeFunction(calldata)
  Operator->>TriggerXAVS.sol: createNewTask(taskInfo)
  TriggerXAVS.sol->>Operator: via listenForNewTasks()
  Operator->>TriggerXAVS.sol: operatorSubmitTask(...)
  TriggerXAVS.sol->>Imua/AVSManager: forward for finalization
  TriggerXAVS.sol->>Job/GasRegistry: addTaskId(), deductTG()
