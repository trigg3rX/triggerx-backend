package middleware

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/redis"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

type RateLimitInfo struct {
	Remaining int   `json:"remaining"`
	Limit     int   `json:"limit"`
	Reset     int64 `json:"reset"`
}

type RateLimiter struct {
	redis  *redis.Client
	logger observability.Logger
}

func NewRateLimiterWithClient(redisClient *redis.Client, logger observability.Logger) (*RateLimiter, error) {
	if redisClient == nil {
		return nil, fmt.Errorf("redis client is nil")
	}

	return &RateLimiter{
		redis:  redisClient,
		logger: logger,
	}, nil
}

const rateLimitScript = `
local key = KEYS[1]
local limit = tonumber(ARGV[1])
local window = tonumber(ARGV[2])

local current = redis.call("INCR", key)
if current == 1 then
    redis.call("EXPIRE", key, window)
end

local ttl = redis.call("TTL", key)

if current > limit then
    return {current, 0, ttl}
else
    return {current, limit - current, ttl}
end
`

func (rl *RateLimiter) ApplyGinRateLimit(c *gin.Context, apiKey *types.ApiKey) error {
	key := fmt.Sprintf("rate_limit:%s", apiKey.Key)
	window := 60 // 1 minute window
	limit := apiKey.RateLimit

	ctx := context.Background()
	result, err := rl.redis.EvalScript(ctx, rateLimitScript, []string{key}, []interface{}{limit, window})
	if err != nil {
		rl.logger.Error(ctx, "Failed to evaluate rate limit script", observability.Error(err))
		return err
	}

	values, ok := result.([]interface{})
	if !ok || len(values) != 3 {
		rl.logger.Error(ctx, "Invalid response from rate limit script")
		return fmt.Errorf("invalid response from rate limit script")
	}

	current := values[0].(int64)
	remaining := values[1].(int64)
	reset := values[2].(int64)

	c.Header("X-RateLimit-Limit", strconv.Itoa(limit))
	c.Header("X-RateLimit-Remaining", strconv.FormatInt(remaining, 10))
	c.Header("X-RateLimit-Reset", strconv.FormatInt(time.Now().Unix()+reset, 10))

	if current > int64(limit) {
		return fmt.Errorf("rate limit exceeded")
	}

	return nil
}

func (rl *RateLimiter) ApplyRateLimit(r *http.Request, apiKey *types.ApiKey) (*http.Response, error) {
	ctx := r.Context()
	rateLimitKey := fmt.Sprintf("rate-limit:%s", apiKey.Key)
	windowSeconds := 60
	currentTimestamp := time.Now().UTC().Unix()

	result, err := rl.redis.Eval(ctx, rateLimitScript, []string{rateLimitKey},
		apiKey.RateLimit, windowSeconds)
	if err != nil {
		rl.logger.Error(ctx, "Rate limiting error", observability.Error(err))

		if true {
			return nil, nil
		}

		return &http.Response{
			StatusCode: http.StatusInternalServerError,
			Body:       nil,
		}, fmt.Errorf("rate limiting service unavailable: %w", err)
	}

	results, ok := result.([]interface{})
	if !ok || len(results) != 3 {
		rl.logger.Error(ctx, "Invalid result from Redis", observability.Any("result", result))
		return nil, fmt.Errorf("invalid response from rate limiter")
	}

	count := int(results[0].(int64))
	remaining := int(results[1].(int64))
	ttl := int(results[2].(int64))
	reset := currentTimestamp + int64(ttl)

	rl.logger.Info(ctx, "Rate Limit Debug", observability.String("api_key", apiKey.Key), observability.String("owner", apiKey.Owner), observability.Int("rate_limit", apiKey.RateLimit), observability.Int("current_count", count), observability.Int("remaining", remaining), observability.Int("ttl", ttl))

	r.Header.Set("X-RateLimit-Limit", strconv.Itoa(apiKey.RateLimit))
	r.Header.Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
	r.Header.Set("X-RateLimit-Reset", strconv.FormatInt(reset, 10))

	if count > apiKey.RateLimit {
		rl.logger.Warn(ctx, "Rate limit exceeded", observability.String("api_key", apiKey.Key), observability.String("owner", apiKey.Owner), observability.Int("count", count), observability.Int("limit", apiKey.RateLimit))

		resp := &http.Response{
			StatusCode: http.StatusTooManyRequests,
			Header:     http.Header{},
			Body:       nil,
		}

		resp.Header.Set("X-RateLimit-Limit", strconv.Itoa(apiKey.RateLimit))
		resp.Header.Set("X-RateLimit-Remaining", "0")
		resp.Header.Set("X-RateLimit-Reset", strconv.FormatInt(reset, 10))
		resp.Header.Set("Retry-After", strconv.Itoa(ttl))

		return resp, fmt.Errorf("rate limit exceeded")
	}

	return nil, nil
}

func (rl *RateLimiter) GetRateLimitStatus(ctx context.Context, apiKey *types.ApiKey) (*RateLimitInfo, error) {
	rateLimitKey := fmt.Sprintf("rate-limit:%s", apiKey.Key)
	currentTimestamp := time.Now().UTC().Unix()

	countStr, err := rl.redis.Get(ctx, rateLimitKey)
	count := 0
	if err == nil && countStr != "" {
		count, _ = strconv.Atoi(countStr)
	}

	ttl, err := rl.redis.TTL(ctx, rateLimitKey)
	if err != nil || ttl < 0 {
		ttl = 60 * time.Second
	}

	return &RateLimitInfo{
		Remaining: int(math.Max(0, float64(apiKey.RateLimit-count))),
		Limit:     apiKey.RateLimit,
		Reset:     currentTimestamp + int64(ttl.Seconds()),
	}, nil
}

// CheckRateLimitForKey checks rate limit for a given API key without using gin.Context
func (rl *RateLimiter) CheckRateLimitForKey(apiKey *types.ApiKey) error {
	key := fmt.Sprintf("rate_limit:%s", apiKey.Key)
	window := 60 // 1 minute window
	limit := apiKey.RateLimit

	ctx := context.Background()
	result, err := rl.redis.EvalScript(ctx, rateLimitScript, []string{key}, []interface{}{limit, window})
	if err != nil {
		rl.logger.Error(ctx, "Failed to evaluate rate limit script", observability.Error(err))
		return err
	}

	values, ok := result.([]interface{})
	if !ok || len(values) != 3 {
		rl.logger.Error(ctx, "Invalid response from rate limit script")
		return fmt.Errorf("invalid response from rate limit script")
	}

	current := values[0].(int64)
	if current > int64(limit) {
		return fmt.Errorf("rate limit exceeded")
	}
	return nil
}
