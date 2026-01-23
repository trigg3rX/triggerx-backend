# ScyllaDB Infrastructure

[ScyllaDB](https://www.scylladb.com/) is the primary persistent data store for TriggerX - a high-performance NoSQL database compatible with Apache Cassandra.

## Why ScyllaDB?

- **Low Latency**: <1ms p50, <10ms p99 read/write operations
- **High Throughput**: 1M+ ops/sec per node
- **No GC Pauses**: Written in C++ (vs Java for Cassandra)
- **Horizontal Scaling**: Add nodes without downtime

## Data Model

### Tables

| Table                | Purpose                     | Primary Key      |
| -------------------- | --------------------------- | ---------------- |
| `user_data`          | User profiles, settings     | `user_id`        |
| `job_data`           | Job definitions             | `job_id`         |
| `time_job_data`      | Time-based job configs      | `job_id`         |
| `event_job_data`     | Event-based job configs     | `job_id`         |
| `condition_job_data` | Condition-based job configs | `job_id`         |
| `task_data`          | Task execution records      | `task_id`        |
| `keeper_data`        | Keeper registry             | `keeper_address` |
| `apikeys`            | API key authentication      | `api_key`        |

## Datastore Package

The `pkg/database/` package provides a service-oriented ScyllaDB client:

```go
import "github.com/trigg3rX/triggerx-backend/pkg/database"

// Initialize
config := connection.NewConfig("scylla-node", "9042")
config.WithKeyspace("triggerx")
service, err := database.NewService(config, logger)

// Use repositories
user, err := service.User().GetByID(ctx, userID)
err = service.Job().Create(ctx, &jobData)
```

### Features

- **Generic Repository Pattern**: Type-safe operations for all entities
- **Connection Pooling**: Managed sessions with health checks
- **gocqlx Integration**: Efficient query building with prepared statements
- **Thread-Safe**: Repository factory and operations are thread-safe

## Configuration

```bash
# Environment variables
DATABASE_HOST_ADDRESS=scylla-node
DATABASE_HOST_PORT=9042
DATABASE_KEYSPACE=triggerx

# Optional TLS
DATABASE_SSL_ENABLED=true
DATABASE_SSL_CERT_PATH=/path/to/client.crt
DATABASE_SSL_KEY_PATH=/path/to/client.key
DATABASE_SSL_CA_PATH=/path/to/ca.crt
```

### Connection Defaults

- **Consistency**: Quorum
- **Timeout**: 30 seconds
- **Retries**: 5
- **Health Check Interval**: 15 seconds

## TLS Setup

For production, enable TLS and authentication. See `database/setup.md` for:

1. Certificate generation (CA, server, client)
2. ScyllaDB configuration with `PasswordAuthenticator`
3. User creation and permission grants
4. Client certificate distribution

## Resources

- [ScyllaDB Documentation](https://docs.scylladb.com/)
- [gocqlx Documentation](https://github.com/scylladb/gocqlx)
- [CQL Reference](https://docs.scylladb.com/stable/cql/)
