package scheduler

import (
	"bytes"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/trigg3rX/triggerx-backend/internal/cache"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

func TestJobCacheDeduplication(t *testing.T) {
	logConfig := logging.LoggerConfig{
		LogDir:          "",
		ProcessName:     "test",
		Environment:     logging.Development,
		UseColors:       false,
		MinStdoutLevel:  logging.DebugLevel,
		MinFileLogLevel: logging.DebugLevel,
	}
	if err := logging.InitServiceLogger(logConfig); err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logging.Shutdown()

	// Initialize cache
	err := cache.Init()
	if err != nil {
		t.Fatalf("Failed to initialize cache: %v", err)
	}
	c, err := cache.GetCache()
	if err != nil {
		t.Fatalf("Failed to get cache: %v", err)
	}

	jobID := int64(12345)
	jobKey := "timejob:processing:" + strconv.FormatInt(jobID, 10)

	// Ensure key does not exist
	_ = c.Delete(jobKey)

	// Simulate first scheduling: should not find the key
	_, err = c.Get(jobKey)
	if err == nil {
		t.Fatalf("Expected cache miss for new job, but got a hit")
	}

	// Set the key (simulate scheduling)
	err = c.Set(jobKey, "1", 2*time.Second)
	if err != nil {
		t.Fatalf("Failed to set cache key: %v", err)
	}

	// Simulate second scheduling: should find the key
	_, err = c.Get(jobKey)
	if err != nil {
		t.Fatalf("Expected cache hit after setting key, but got miss: %v", err)
	}

	// Wait for TTL to expire
	time.Sleep(3 * time.Second)

	// Key should be expired
	_, err = c.Get(jobKey)
	if err == nil {
		t.Fatalf("Expected cache miss after TTL expiry, but got a hit")
	}

	// Clean up
	_ = c.Delete(jobKey)
}

func TestCacheTTLAccuracy(t *testing.T) {
	logConfig := logging.LoggerConfig{
		LogDir:          "",
		ProcessName:     "test",
		Environment:     logging.Development,
		UseColors:       false,
		MinStdoutLevel:  logging.DebugLevel,
		MinFileLogLevel: logging.DebugLevel,
	}
	if err := logging.InitServiceLogger(logConfig); err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logging.Shutdown()

	// Initialize cache
	err := cache.Init()
	if err != nil {
		t.Fatalf("Failed to initialize cache: %v", err)
	}
	c, err := cache.GetCache()
	if err != nil {
		t.Fatalf("Failed to get cache: %v", err)
	}

	jobID := int64(54321)
	jobKey := "timejob:processing:" + strconv.FormatInt(jobID, 10)
	_ = c.Delete(jobKey)

	ttl := 2 * time.Second
	err = c.Set(jobKey, "1", ttl)
	if err != nil {
		t.Fatalf("Failed to set cache key: %v", err)
	}

	// Should exist immediately
	_, err = c.Get(jobKey)
	if err != nil {
		t.Fatalf("Expected cache hit right after set, got miss: %v", err)
	}

	// Poll until key expires, measure time
	start := time.Now()
	for {
		time.Sleep(100 * time.Millisecond)
		_, err = c.Get(jobKey)
		if err != nil {
			break // Key expired
		}
		if time.Since(start) > ttl+1*time.Second {
			t.Fatalf("Key did not expire within expected TTL window")
		}
	}
	elapsed := time.Since(start)
	t.Logf("Key expired after %v (expected ~%v)", elapsed, ttl)
	if elapsed < ttl || elapsed > ttl+500*time.Millisecond {
		t.Errorf("TTL expiry was not accurate: got %v, expected %v", elapsed, ttl)
	}
}

func TestCacheReadAndSendTime(t *testing.T) {
	logConfig := logging.LoggerConfig{
		LogDir:          "",
		ProcessName:     "test",
		Environment:     logging.Development,
		UseColors:       false,
		MinStdoutLevel:  logging.DebugLevel,
		MinFileLogLevel: logging.DebugLevel,
	}
	if err := logging.InitServiceLogger(logConfig); err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logging.Shutdown()

	// Initialize cache
	err := cache.Init()
	if err != nil {
		t.Fatalf("Failed to initialize cache: %v", err)
	}
	c, err := cache.GetCache()
	if err != nil {
		t.Fatalf("Failed to get cache: %v", err)
	}

	jobID := int64(67890)
	jobKey := "timejob:processing:" + strconv.FormatInt(jobID, 10)
	value := "test-payload"
	_ = c.Set(jobKey, value, 5*time.Second)

	// Measure cache read
	start := time.Now()
	val, err := c.Get(jobKey)
	if err != nil {
		t.Fatalf("Failed to get cache key: %v", err)
	}
	readDuration := time.Since(start)
	t.Logf("Cache read took %v", readDuration)

	// Simulate sending to a service (replace with your actual endpoint)
	serviceURL := "http://localhost:8080/echo" // Example endpoint
	req, _ := http.NewRequest("POST", serviceURL, bytes.NewBuffer([]byte(val)))
	req.Header.Set("Content-Type", "application/json")

	startSend := time.Now()
	resp, err := http.DefaultClient.Do(req)
	sendDuration := time.Since(startSend)
	if err != nil {
		t.Logf("Failed to send to service (is it running?): %v", err)
	} else {
		t.Logf("Send to service took %v", sendDuration)
		resp.Body.Close()
	}
}

func TestCacheReadTimeFor100Jobs(t *testing.T) {
	logConfig := logging.LoggerConfig{
		LogDir:          "",
		ProcessName:     "test",
		Environment:     logging.Development,
		UseColors:       false,
		MinStdoutLevel:  logging.DebugLevel,
		MinFileLogLevel: logging.DebugLevel,
	}
	if err := logging.InitServiceLogger(logConfig); err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logging.Shutdown()

	// Initialize cache
	err := cache.Init()
	if err != nil {
		t.Fatalf("Failed to initialize cache: %v", err)
	}
	c, err := cache.GetCache()
	if err != nil {
		t.Fatalf("Failed to get cache: %v", err)
	}

	// Insert 100 jobs into cache
	jobKeys := make([]string, 100)
	for i := 0; i < 100; i++ {
		jobID := int64(10000 + i)
		jobKey := "timejob:processing:" + strconv.FormatInt(jobID, 10)
		jobKeys[i] = jobKey
		err := c.Set(jobKey, "payload", 10*time.Second)
		if err != nil {
			t.Fatalf("Failed to set cache key %s: %v", jobKey, err)
		}
	}

	// Measure read time for all 100 jobs
	var totalReadTime time.Duration
	for _, jobKey := range jobKeys {
		start := time.Now()
		_, err := c.Get(jobKey)
		readDuration := time.Since(start)
		if err != nil {
			t.Fatalf("Failed to get cache key %s: %v", jobKey, err)
		}
		totalReadTime += readDuration
	}

	avgReadTime := totalReadTime / 100
	t.Logf("Average cache read time for 100 jobs: %v", avgReadTime)

	// Clean up
	for _, jobKey := range jobKeys {
		_ = c.Delete(jobKey)
	}
}
