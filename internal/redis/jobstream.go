package redis

import (
	"context"
	"encoding/json"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

const (
	JobsReadyStream = "jobs:ready"
	JobsRetryStream = "jobs:retry"
	StreamMaxLen    = 10000 // Retain only last 10,000 jobs
	StreamTTL       = 24 * time.Hour
)

func AddJobToStream(stream string, job interface{}) error {
	ctx := context.Background()
	logger := logging.GetServiceLogger()
	jobJSON, err := json.Marshal(job)
	if err != nil {
		logger.Errorf("Failed to marshal job: %v", err)
		return err
	}
	res, err := GetClient().XAdd(ctx, &redis.XAddArgs{
		Stream: stream,
		MaxLen: StreamMaxLen,
		Approx: true,
		Values: map[string]interface{}{
			"job":        jobJSON,
			"created_at": time.Now().Unix(),
		},
	}).Result()
	if err != nil {
		logger.Errorf("Failed to add job to stream %s: %v", stream, err)
		return err
	}
	logger.Infof("Job added to stream %s with ID %s", stream, res)

	// Set TTL on the stream key (refreshes on each add)
	expireRes := GetClient().Expire(ctx, stream, StreamTTL)
	if expireRes.Err() != nil {
		logger.Warnf("Failed to set TTL on stream %s: %v", stream, expireRes.Err())
	}

	// Check stream length for overflow
	lenRes := GetClient().XLen(ctx, stream)
	if lenRes.Err() == nil && lenRes.Val() >= StreamMaxLen {
		logger.Warnf("Stream %s reached max length (%d)", stream, StreamMaxLen)
	}

	return nil
}
