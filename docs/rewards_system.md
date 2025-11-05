# Keeper Rewards System

## Overview

The TriggerX rewards system distributes daily points to keepers based on their uptime. This incentivizes keepers to maintain high availability and ensures reliable task execution across the network.

## Architecture

### Components

1. **Redis Client** (`pkg/client/redis`) - Shared Redis client for all services
2. **Rewards Cache** (`internal/health/cache`) - Wrapper providing rewards-specific Redis operations
3. **Rewards Service** - Orchestrates daily reward distribution
4. **State Manager** - Tracks keeper health and increments daily uptime
5. **Database** - Persistent storage for accumulated keeper points

### Data Flow

```bash
Keeper Health Check (every 60s)
    ↓
State Manager increments daily uptime in Redis (+60 seconds)
    ↓
Daily Distribution (06:30 UTC)
    ↓
Rewards Service reads all daily uptimes from Redis
    ↓
Calculate rewards based on uptime tiers
    ↓
Update keeper_points in Database
    ↓
Reset Redis daily uptime counters
```

## Reward Tiers

Daily rewards are calculated based on 24-hour uptime:

| Uptime Range | Reward | Points |
|--------------|--------|--------|
| 20+ hours | Full reward (100%) | 1000 |
| 15-20 hours | 2/3 reward (~67%) | 667 |
| 10-15 hours | 1/3 reward (~33%) | 333 |
| 6-10 hours | Fractional (linear) | 0-333 |
| < 6 hours | No reward | 0 |

### Calculation Logic

```go
// Minimum threshold
if dailyUptime < 6 hours → 0 points

// Tier 1: Full reward
if dailyUptime >= 20 hours → 1000 points

// Tier 2: Two-thirds reward
if dailyUptime >= 15 hours → 667 points

// Tier 3: One-third reward
if dailyUptime >= 10 hours → 333 points

// Between 6-10 hours: Linear interpolation
if 6 hours <= dailyUptime < 10 hours:
    fraction = (dailyUptime - 6h) / (10h - 6h)
    points = fraction * 333
```

## Redis Keys

### Daily Uptime Tracking

- **Key**: `keeper:daily_uptime:{keeper_address}`
- **Type**: String (integer)
- **Value**: Cumulative seconds of uptime today
- **TTL**: 48 hours (allows for one missed distribution)
- **Updated**: Every 60 seconds when keeper is active

### Rewards Metadata

- **Key**: `rewards:last_distribution`
- **Type**: String (RFC3339 timestamp)
- **Value**: Timestamp of last successful distribution
- **TTL**: None (persistent)

- **Key**: `rewards:current_period_start`
- **Type**: String (RFC3339 timestamp)
- **Value**: Start of current 24-hour reward period
- **TTL**: None (persistent)

## Distribution Schedule

- **Time**: 06:30 UTC daily
- **Process**:
  1. Read all daily uptimes from Redis
  2. Calculate reward points per keeper
  3. Add points to keeper accounts in database
  4. Update distribution timestamp
  5. Reset all daily uptime counters
  6. Start new 24-hour period

### Missed Distributions

If the service is down during distribution time:

- The system logs missed distributions but does not backfill
- Historical uptime data is not retained beyond the current period
- Next distribution runs on schedule
- This is by design to incentivize consistent keeper operation

## Database Schema

### Keeper Points

Points are stored in the `keeper_data` table:

```sql
keeper_address: string (primary key)
keeper_points: string (big integer as string)
```

Points are accumulated over time and can be used for:

- Governance voting power
- Rewards distribution
- Keeper reputation scoring

## API Endpoints

### Get Rewards Health Status

```http
GET /rewards/health
```

**Response:**

```json
{
  "service": "rewards",
  "status": "running",
  "last_distribution": "2025-10-16T06:30:00Z",
  "time_since_last_distribution": "18h30m45s",
  "current_period_start": "2025-10-16T06:30:00Z",
  "current_period_duration": "18h30m45s",
  "next_distribution": "2025-10-17T06:30:00Z",
  "time_until_next_distribution": "5h29m15s",
  "keepers_tracked": 42
}
```

**Status Values:**

- `running`: Normal operation
- `degraded`: Errors retrieving data (still operational)
- `overdue`: Last distribution was > 25 hours ago

### Get Keeper Daily Uptime

```http
GET /rewards/uptime?keeper_address=0x123...
```

**Response:**

```json
{
  "keeper_address": "0x123...",
  "daily_uptime_seconds": 72000,
  "daily_uptime_hours": 20.0
}
```

## Configuration

### Environment Variables

Add to your `.env` file:

```bash
# Redis connection (Upstash or any Redis provider)
UPSTASH_REDIS_REST_URL=redis://...
UPSTASH_REDIS_REST_TOKEN=your_token
```

### Service Configuration

The rewards service is automatically initialized and started in the health service:

```go
// In cmd/health/main.go
cacheClient, err := cache.NewClient(
    config.GetRedisURL(),
    config.GetRedisPassword(),
    logger,
)

rewardsService := rewards.NewService(logger, cacheClient, dbManager)
rewardsService.Start()
```

## Monitoring

### Health Checks

Monitor rewards service health through:

1. HTTP endpoint: `GET /rewards/health`
2. Logs: Search for "rewards" component
3. Metrics: Check keeper point updates in database

### Key Metrics to Monitor

- **Time since last distribution**: Should be < 25 hours
- **Keepers tracked**: Should match active keeper count
- **Distribution success**: Check logs at 06:30 UTC daily
- **Redis connection**: Monitor cache client connectivity

### Alerts

Set up alerts for:

- Distribution overdue (> 25 hours)
- Redis connection failures
- Failed point updates in database
- Zero keepers tracked

## Troubleshooting

### Rewards Not Distributing

1. **Check Redis Connection**

   ```bash
   # Check Redis availability
   redis-cli ping
   ```

2. **Verify Service Started**

   ```bash
   # Check logs for:
   "Rewards service started successfully"
   ```

3. **Check Scheduler**

   ```bash
   # Look for log entry:
   "Next reward distribution scheduled"
   ```

### Keeper Not Receiving Rewards

1. **Verify Keeper is Active**
   - Keeper must be registered and whitelisted
   - Keeper must be sending health check-ins

2. **Check Daily Uptime**

   ```bash
   curl "http://localhost:8080/rewards/uptime?keeper_address=0x..."
   ```

3. **Verify Minimum Threshold**
   - Keeper must have >= 6 hours uptime in 24h period

### Redis Data Loss

If Redis data is lost:

- Current day's uptime tracking is reset to zero
- Keepers start accumulating uptime from zero
- This effectively means no rewards for that day
- Normal operation resumes on next distribution

## Security Considerations

### Redis Access

- Use authentication (password/token)
- Restrict network access to health service only
- Enable TLS for production environments
- Use Upstash or managed Redis for production

### Point Manipulation

- Points are calculated server-side only
- Keepers cannot directly modify their points
- Redis cache is not exposed to keepers
- Database updates use atomic operations

## Performance

### Redis Operations

- **Writes**: ~1 per keeper per minute (60s health check)
- **Reads**: Bulk read once per day at distribution
- **Memory**: ~50 bytes per active keeper

### Database Operations

- **Updates**: 1 per keeper per day (point addition)
- **Queries**: Minimal - only for point retrieval

### Scalability

Current architecture supports:

- 1000+ concurrent active keepers
- Sub-millisecond uptime increments
- Horizontal scaling through Redis clustering

## Future Enhancements

### Planned Features

1. **Historical Rewards Tracking**
   - Store daily reward history per keeper
   - Enable reward analytics and reporting

2. **Dynamic Reward Multipliers**
   - Bonus rewards for task execution
   - Penalty for SLA violations

3. **Configurable Reward Tiers**
   - Admin API to adjust point values
   - Custom thresholds per environment

4. **Reward Notifications**
   - Telegram/Email notifications on reward distribution
   - Daily uptime summaries for keepers

## Testing

### Manual Testing

1. **Start Health Service**

   ```bash
   go run cmd/health/main.go
   ```

2. **Simulate Keeper Check-in**

   ```bash
   curl -X POST http://localhost:8080/health \
     -H "Content-Type: application/json" \
     -d '{
       "keeper_address": "0x123...",
       "consensus_pub_key": "...",
       "consensus_address": "0x456...",
       "version": "v1.0.0",
       "timestamp": "2025-10-16T12:00:00Z",
       "signature": "...",
       "is_imua": false
     }'
   ```

3. **Check Daily Uptime**

   ```bash
   curl "http://localhost:8080/rewards/uptime?keeper_address=0x123..."
   ```

4. **Check Rewards Health**

   ```bash
   curl http://localhost:8080/rewards/health
   ```

### Integration Testing

See `internal/health/rewards/rewards_test.go` for comprehensive test suite.

## References

- [Redis Best Practices](https://redis.io/docs/manual/patterns/)
- [TriggerX Architecture](./2_architecture.md)
- [Database Schema](./5_datatypes.md)
- [Health Service Documentation](./4_services.md)
