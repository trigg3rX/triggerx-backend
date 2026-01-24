# Development Overview

This document describes the details, practises and architecture used in the codebase, for understanding and contributing to the project.

## Services

The backend is built with a microservices architecture, with each service having a specific responsibility.

| Service                | Description                                          |
|------------------------|------------------------------------------------------|
| `dbserver`             | Main API server with DB access for SDK (& web app)   |
| `eventmonitor`         | Monitors blockchain events                           |
| `health`               | Health monitoring & keeper state management          |
| `keeper`               | Executes tasks & validates its peers' executed tasks |
| `schedulers/condition` | Schedules event-based and condition-based jobs       |
| `schedulers/time`      | Schedules time-based jobs                            |
| `taskdispatcher`       | Dispatches tasks to keepers                          |
| `taskmonitor`          | Monitors task execution & completion                 |

See [Services Structure](./b_services_structure.md) for detailed service-specific information.
