package redis

import (
    "context"
    "os"
    "sync"
    "time"

    "github.com/go-redis/redis/v8"
)

var (
    client *redis.Client
    once   sync.Once
)

// GetClient returns a singleton Redis client
func GetClient() *redis.Client {
    once.Do(func() {
        addr := os.Getenv("REDIS_ADDR")
        if addr == "" {
            addr = "localhost:6379"
        }
        password := os.Getenv("REDIS_PASSWORD") // "" if no password set
        db := 0 // use default DB

        client = redis.NewClient(&redis.Options{
            Addr:     addr,
            Password: password,
            DB:       db,
        })
    })
    return client
}

// Ping checks if Redis is reachable
func Ping() error {
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
    defer cancel()
    return GetClient().Ping(ctx).Err()
}