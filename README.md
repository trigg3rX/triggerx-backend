# TriggerX Keeper Backend

The **TriggerX Keeper Backend** is a decentralized system designed to automate and manage task execution across blockchain networks. It consists of three core components: **Task Manager**, **Keepers**, and **Aggregator**, each playing a critical role in ensuring efficient, reliable, and scalable task orchestration.

---

## Table of Contents

- [TriggerX Keeper Backend](#triggerx-keeper-backend)
  - [Table of Contents](#table-of-contents)
  - [Introduction](#introduction)
  - [Core Components](#core-components)
    - [Task Manager](#task-manager)
    - [Keepers](#keepers)
    - [Aggregator](#aggregator)
  - [Steps to Run the Keeper Backend](#steps-to-run-the-keeper-backend)
  - [High Availability Setup](#high-availability-setup)
    - [Manager HA Configuration](#manager-ha-configuration)
    - [How It Works](#how-it-works)

---

## Introduction

The TriggerX Keeper Backend simplifies task management and automation in blockchain ecosystems. By leveraging decentralized technologies, it ensures fault-tolerant and secure task orchestration while enabling efficient cross-chain operations. Designed for scalability and reliability, the system provides a flexible and extensible platform for blockchain automation.

---

## Core Components

### Task Manager

The **Task Manager** serves as the backbone for decentralized job scheduling and execution. It:
- Automates task scheduling and optimizes resource usage.
- Monitors progress and persists results.
- Features:
  - Load balancing.
  - State persistence.
  - Support for various execution triggers.
  
### Keepers

The **Keepers** are responsible for executing and validating tasks by:
- Interacting with smart contracts securely.
- Processing arguments (static, dynamic, or none).
- Integrating with external data sources for real-time inputs.
- Validating Triggers and Actions of Tasks executed by Peers.

Operating in a decentralized architecture, Keepers ensure:
- Fault tolerance.
- Efficient resource usage.
- Secure and reliable contract interactions.

### Aggregator

The **Aggregator** ensures the consensus of tasks by:
- Aggregating tasks from multiple Keepers.
- Submitting the tasks to the blockchain.
- Acting as a bootstrap for the p2p network.

---

## Steps to Run the Keeper Backend

1. Clone the repository.
     - ```sh
       git clone https://github.com/trigg3rX/triggerx-backend.git
       ```
2. Install the dependencies.
     - ```sh
       go mod tidy
       ```
     - ```sh
       npm i -g @othentic/othentic-cli  # (Node v22.6.0 is required)
       ```
3. Copy the `.env.example` file to `.env` and set the environment variables.

4. Set up the database.
   - ```sh
     make db-setup
     ```

5. Start the database server.
   - ```sh
     make start-db-server
     ```

6. Start the Aggregator.
   - ```sh
     make start-othentic
     ```

7. Start the Task Manager.
   - ```sh
     make start-manager
     ```

7. Start the Keepers.
   - Clone the repo:
     - ```sh
       git clone https://github.com/trigg3rX/triggerx-keeper-setup.git
       ```
   - Run the Docker image:
     - ```sh
       ./triggerx.sh start
       ```

8. Run the Keeper node without docker.
   - ```sh
     make start-keeper
     ```

## High Availability Setup

The TriggerX platform supports high availability for the task manager component, allowing for redundancy and automatic failover if one instance goes down.

### Manager HA Configuration

The manager service now supports a leader-follower architecture with automatic leader election:

- Multiple manager instances can be configured to run in a cluster
- A distributed lock using Redis ensures only one manager is the active leader
- Automatic failover happens if the leader instance goes down
- All instances accept task validation and p2p messages for task execution
- Only the leader processes job creation and management requests

To enable high availability mode, set the following environment variables:

```
MANAGER_HA_ENABLED=true
REDIS_ADDRESS=redis:6379
REDIS_PASSWORD=
OTHER_MANAGER_ADDRESSES=manager2:8082
```

The docker-compose.yaml file includes a configuration for running two manager instances with Redis for leader election.

### How It Works

- Manager instances automatically elect a leader using distributed locking via Redis
- If the leader goes down, another instance will automatically become the new leader
- Job scheduling is only performed by the leader to avoid duplicate jobs
- All instances can process task validation and execution, ensuring tasks are still processed if one manager is down
- Health endpoints are available at `/health` and `/status` to monitor manager instances
