# TriggerX Keeper Backend

The **TriggerX Keeper Backend** is a decentralized system designed to automate and manage task execution across blockchain networks. It consists of three core components: **Schedulers**, **Keepers**, and **Aggregator**, each playing a critical role in ensuring efficient, reliable, and scalable task orchestration.

---

<p align="center"><img src="https://github.com/trigg3rX/triggerx-backend/actions/workflows/build.yml/badge.svg" alt="Build" /> <img src="https://github.com/trigg3rX/triggerx-backend/actions/workflows/go_lint.yml/badge.svg" alt="Lint" /></p>

---

## Table of Contents

- [TriggerX Keeper Backend](#triggerx-keeper-backend)
  - [Table of Contents](#table-of-contents)
  - [Introduction](#introduction)
  - [Core Components](#core-components)
    - [Schedulers](#schedulers)
    - [Keepers](#keepers)
    - [Aggregator](#aggregator)
      - [Developer Notes can be found here at devNotes.md](#developer-notes-can-be-found-here-at-devnotesmd)
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

### TaskDispatcher

The **TaskDispatcher** is responsible for:

- Managing the task queue.
- Selecting appropriate Keepers based on availability and load.
- Dispatching tasks to the selected Keeper via the P2P network.
- Handling task reassignment in case of failures.

### TaskMonitor

The **TaskMonitor** ensures reliable execution by:

- Tracking the lifecycle of assigned tasks.
- Monitoring the Attestation Center for task submissions.
- Updating task status in the database.
- Handling timeouts and triggering retries.

### Aggregator

The **Aggregator** is part of the **Othentic Network** and ensures consensus by:

- Aggregating tasks from multiple Keepers.
- Collecting attestations from validator nodes.
- Achieving BFT consensus on task results.
- Submitting validated tasks to the blockchain.

#### Developer Notes can be found here at [devNotes.md](docs/devNotes.md)

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
     just db-setup
     ```

5. Start the database server.

   - ```sh
     just start-db-server
     ```

6. Start the Aggregator.

   - ```sh
     just start-othentic
     ```

7. Start the Redis.

   - ```sh
     just start-redis
     ```

8. Start the Task Dispatcher.

   - ```sh
     just start-taskdispatcher
     ```

9. Start the Task Monitor.

   - ```sh
     just start-taskmonitor
     ```

10. Start the Schedulers.

- ```sh
  just start-time-scheduler
  just start-condition-scheduler
  ```

11. Start the Event Monitor.

- ```sh
  just start-eventmonitor
  ```

12. Start the Keepers.

- Clone the repo:

  - ```sh
    git clone https://github.com/trigg3rX/triggerx-keeper-setup.git
    ```

- Run the Docker image:

  - ```sh
    ./triggerx.sh start
    ```

13. Run the Keeper node without docker.

    - ```sh
      just start-keeper
      ```

**Note:**

- Ideally all these services should be run via the `just run-all` which is a WIP. for now, you can run them individually as mentioned above.
