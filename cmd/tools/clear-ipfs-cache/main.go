package main

import (
	"context"
	"fmt"
	"log"

	dbconfig "github.com/trigg3rX/triggerx-backend/internal/dbserver/config"
	redisclient "github.com/trigg3rX/triggerx-backend/internal/dbserver/redis"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

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

	ctx := context.Background()
	iter := redisClient.Client().Scan(ctx, 0, "ipfs:validation:*", 0).Iterator()

	var keys []string
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	if err := iter.Err(); err != nil {
		log.Fatalf("failed to scan redis keys: %v", err)
	}

	if len(keys) == 0 {
		fmt.Println("No ipfs validation cache entries found")
		return
	}

	if err := redisClient.Del(ctx, keys...); err != nil {
		log.Fatalf("failed to delete cache entries: %v", err)
	}

	fmt.Printf("Deleted %d ipfs validation cache entries\n", len(keys))
}

