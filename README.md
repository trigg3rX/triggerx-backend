# TriggerX Keeper Backend

The **TriggerX Keeper Backend** is a decentralized system designed to automate and manage task execution across blockchain networks. It consists of three core components: **Schedulers**, **Keepers**, and **Aggregator**, each playing a critical role in ensuring efficient, reliable, and scalable task orchestration.

---

## Table of Contents

- [TriggerX Keeper Backend](#triggerx-keeper-backend)
  - [Table of Contents](#table-of-contents)
  - [Introduction](#introduction)
  - [Core Components](#core-components)
    - [Schedulers](#schedulers)
    - [Keepers](#keepers)
    - [Aggregator](#aggregator)
  - [Steps to Run the Keeper Backend](#steps-to-run-the-keeper-backend)

---

## Introduction

The TriggerX Keeper Backend simplifies task management and automation in blockchain ecosystems. By leveraging decentralized technologies, it ensures fault-tolerant and secure task orchestration while enabling efficient cross-chain operations. Designed for scalability and reliability, the system provides a flexible and extensible platform for blockchain automation.

---

## Core Components

### Schedulers

The **Schedulers** serve as the backbone for decentralized job scheduling and execution. They:

- Automate task scheduling and optimizes resource usage.
- Monitor progress and persist results.
- Common Features:
  - Load balancing by having multiple scheduler instances for each job type.
  - State persistence by storing the state of the job in the database.
  - Cache management by storing the state in redis.
  - Metrics collection by storing the metrics in Prometheus.
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

7. Start the Redis.

   - ```sh
     make start-redis
     ```

8. Start the Schedulers.

   - ```sh
     make start-time-scheduler
     make start-event-schedulers
     make start-condition-scheduler
     ```

9. Start the Keepers.
   - Clone the repo:

     - ```sh
       git clone https://github.com/trigg3rX/triggerx-keeper-setup.git
       ```

   - Run the Docker image:

     - ```sh
       ./triggerx.sh start
       ```

10. Run the Keeper node without docker.

     - ```sh
       make start-keeper
       ```
