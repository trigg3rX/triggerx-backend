package redis

import (
    "context"
    "encoding/json"
    "time"

    "github.com/go-redis/redis/v8"
)

const (
    JobsReadyStream = "jobs:ready"
    JobsRetryStream = "jobs:retry"
    StreamMaxLen    = 10000 // Retain only last 10,000 jobs
    StreamTTL       = 24 * time.Hour
)

func AddJobToStream(stream string, job interface{}) error {
    ctx := context.Background()
    jobJSON, err := json.Marshal(job)
    if err != nil {
        return err
    }
    _, err = GetClient().XAdd(ctx, &redis.XAddArgs{
        Stream: stream,
        MaxLen: StreamMaxLen,
        Approx: true,
        Values: map[string]interface{}{
            "job":       jobJSON,
            "created_at": time.Now().Unix(),
        },
    }).Result()
    if err != nil {
        return err
    }
    // Set TTL on the stream key (refreshes on each add)
    GetClient().Expire(ctx, stream, StreamTTL)
    return nil
}