package main

import (
	"context"
	"fmt"
	"log"
	"os"

	dbconfig "github.com/trigg3rX/triggerx-backend/internal/dbserver/config"
	redisclient "github.com/trigg3rX/triggerx-backend/internal/dbserver/redis"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

// clear-job-cache logs all job data currently matching the given job ID in Redis, then deletes it.
// Default job ID is 999888777666555, but you can override it by passing a CLI arg:
//
//	go run ./cmd/tools/clear-job-cache 123456
func main() {
	if err := dbconfig.Init(); err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	logger, err := logging.NewZapLogger(logging.LoggerConfig{
		ProcessName:   logging.DatabaseProcess,
		IsDevelopment: true,
	})
	if err != nil {
		log.Fatalf("failed to initialise logger: %v", err)
	}

	redisClient, err := redisclient.NewClient(logger)
	if err != nil {
		log.Fatalf("failed to connect to redis: %v", err)
	}
	defer func() { _ = redisClient.Close() }()

	jobID := "999888777666555"
	if len(os.Args) > 1 && os.Args[1] != "" {
		jobID = os.Args[1]
	}

	pattern := "*" + jobID + "*"
	ctx := context.Background()

	iter := redisClient.Client().Scan(ctx, 0, pattern, 0).Iterator()

	var keys []string
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	if err := iter.Err(); err != nil {
		log.Fatalf("failed to scan redis keys: %v", err)
	}

	if len(keys) == 0 {
		fmt.Printf("No Redis keys found matching pattern %q for job ID %s\n", pattern, jobID)
		return
	}

	fmt.Printf("Found %d Redis keys matching job ID %s (pattern %q):\n", len(keys), jobID, pattern)
	// Log all current job data before deletion
	for _, key := range keys {
		val, err := redisClient.Client().Get(ctx, key).Result()
		if err != nil {
			fmt.Printf("  Key: %s (error fetching value: %v)\n", key, err)
			continue
		}
		fmt.Printf("  Key: %s\n    Value: %s\n", key, val)
	}

	if err := redisClient.Del(ctx, keys...); err != nil {
		log.Fatalf("failed to delete keys: %v", err)
	}

	fmt.Printf("Deleted %d Redis keys matching job ID %s (pattern %q)\n", len(keys), jobID, pattern)
}

