package cache

import (
	"context"
	"time"

	redisx "github.com/trigg3rX/triggerx-backend/internal/redis"
)

// RedisCache implements the Cache interface using Redis as the backend
type RedisCache struct {
	client *redisx.Client
}

// Get retrieves a value from Redis cache
func (r *RedisCache) Get(key string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	return r.client.Get(ctx, key)
}

// Set stores a value in Redis cache with TTL
func (r *RedisCache) Set(key string, value string, ttl time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	return r.client.Set(ctx, key, value, ttl)
}

// Delete removes a key from Redis cache
func (r *RedisCache) Delete(key string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	return r.client.Del(ctx, key)
}

// AcquirePerformerLock acquires a distributed lock for a performer using Redis SetNX
func (r *RedisCache) AcquirePerformerLock(performerID string, ttl time.Duration) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	key := "performer:busy:" + performerID
	return r.client.SetNX(ctx, key, "1", ttl)
}

// ReleasePerformerLock releases a performer lock by deleting the key
func (r *RedisCache) ReleasePerformerLock(performerID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	key := "performer:busy:" + performerID
	return r.client.Del(ctx, key)
}

// Close closes the Redis connection
func (r *RedisCache) Close() error {
	return r.client.Close()
}
