# Redis Infrastructure

TriggerX uses [Upstash Redis](https://upstash.com/redis) for task lifecycle management, distributed coordination, and caching.

## Task Lifecycle Streams

Redis Streams manage task state transitions:

| Stream            | Purpose                                   |
| ----------------- | ----------------------------------------- |
| `task:dispatched` | Tasks sent to Keepers (pending execution) |
| `task:executed`   | Tasks executed, awaiting validation       |
| `task:validated`  | Tasks validated on-chain                  |
| `task:completed`  | Final completed state                     |
| `task:failed`     | Failed tasks                              |
| `task:retry`      | Tasks queued for retry                    |

All streams have a **30-minute TTL** to prevent unbounded growth.

## Other Use Cases

- **Distributed Locking**: Coordination between scheduler instances
- **Task Timeout Tracking**: Sorted sets for tracking task execution timeouts
- **Caching**: Temporary storage of frequently accessed data

## Configuration

```bash
# Environment variables
UPSTASH_REDIS_URL=redis://default:password@host:port
UPSTASH_REDIS_REST_TOKEN=token  # Optional for REST API
```

### Connection Settings

Default configuration in `pkg/client/redis/`:

- Pool Size: 10 connections
- Min Idle: 2 connections
- Max Retries: 3
- Dial/Read/Write Timeout: 3-5 seconds
- Automatic connection recovery

## Implementation

The Redis client (`pkg/client/redis/client.go`) wraps `github.com/redis/go-redis/v9` with:

- Retry logic with exponential backoff
- Connection health monitoring
- Automatic recovery on connection loss

## Resources

- [Upstash Redis Documentation](https://docs.upstash.com/redis)
- [Redis Streams](https://redis.io/docs/data-types/streams/)
- [go-redis Documentation](https://redis.uptrace.dev/)
