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
       npm i -g @othentic/othentic-cli@1.8.1  # (Node v22.6.0 is required)
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
     make start-aggregator
     ```

7. Start the Task Manager.
   - ```sh
     make start-manager
     ```

7. Start the Keepers.
   - Pull the Docker image for executing tasks:
     - ```sh
       docker pull trigg3rx/triggerx-keeper:latest
       ```
   - Run the Docker image:
     - ```sh
       docker run --env .env --name triggerx_keeper -d trigg3rx/ triggerx-keeper:latest
       ```

8. Run the Keeper node without docker.
   - ```sh
     make start-keeper
     ```
