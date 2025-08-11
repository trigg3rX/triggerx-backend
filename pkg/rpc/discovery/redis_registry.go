package discovery

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/trigg3rX/triggerx-backend/pkg/client/redis"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/rpc"
)

// RedisRegistry implements the ServiceRegistry interface using Redis as the backend
type RedisRegistry struct {
	client     *redis.Client
	logger     logging.Logger
	config     RedisRegistryConfig
	watchers   map[string][]chan rpc.ServiceInfo
	watcherMu  sync.RWMutex
	stopChan   chan struct{}
	processMap map[string]logging.ProcessName
}

// RedisRegistryConfig holds configuration for the Redis registry
type RedisRegistryConfig struct {
	// Redis configuration
	RedisConfig redis.RedisConfig
	// Registry-specific settings
	KeyPrefix       string
	TTL             time.Duration
	RefreshInterval time.Duration
	WatchBufferSize int
	MaxRetries      int
	RetryDelay      time.Duration
}

// DefaultRedisRegistryConfig returns default configuration for Redis registry
func DefaultRedisRegistryConfig() RedisRegistryConfig {
	return RedisRegistryConfig{
		RedisConfig: redis.RedisConfig{
			UpstashConfig: redis.UpstashConfig{
				URL:   "", // Must be set by caller
				Token: "", // Must be set by caller
			},
			ConnectionSettings: redis.ConnectionSettings{
				PoolSize:     10,
				MinIdleConns: 2,
				MaxRetries:   3,
				DialTimeout:  5 * time.Second,
				ReadTimeout:  3 * time.Second,
				WriteTimeout: 3 * time.Second,
				PoolTimeout:  4 * time.Second,
			},
		},
		KeyPrefix:       "triggerx:registry:",
		TTL:             30 * time.Second,
		RefreshInterval: 15 * time.Second,
		WatchBufferSize: 100,
		MaxRetries:      3,
		RetryDelay:      1 * time.Second,
	}
}

// NewRedisRegistry creates a new Redis-based service registry
func NewRedisRegistry(logger logging.Logger, config RedisRegistryConfig) (*RedisRegistry, error) {
	if config.RedisConfig.UpstashConfig.URL == "" {
		return nil, fmt.Errorf("redis URL is required")
	}

	client, err := redis.NewRedisClient(logger, config.RedisConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Redis client: %w", err)
	}

	registry := &RedisRegistry{
		client:   client,
		logger:   logger,
		config:   config,
		watchers: make(map[string][]chan rpc.ServiceInfo),
		stopChan: make(chan struct{}),
		processMap: map[string]logging.ProcessName{
			"aggregator":           logging.AggregatorProcess,
			"dbserver":             logging.DatabaseProcess,
			"keeper":               logging.KeeperProcess,
			"registrar":            logging.RegistrarProcess,
			"health":               logging.HealthProcess,
			"taskdispatcher":       logging.TaskDispatcherProcess,
			"taskmonitor":          logging.TaskMonitorProcess,
			"schedulers-time":      logging.TimeSchedulerProcess,
			"schedulers-condition": logging.ConditionSchedulerProcess,
			"test":                 logging.TestProcess,
		},
	}

	// Start the background refresh goroutine
	go registry.refreshLoop()

	logger.Infof("Redis registry initialized with prefix: %s", config.KeyPrefix)
	return registry, nil
}

// Register registers a service in the Redis registry
func (r *RedisRegistry) Register(ctx context.Context, info rpc.ServiceInfo) error {
	key := r.serviceKey(info.Name)

	// Update the last seen timestamp
	info.LastSeen = time.Now()

	// Marshal the service info to JSON
	data, err := json.Marshal(info)
	if err != nil {
		return fmt.Errorf("failed to marshal service info: %w", err)
	}

	// Store in Redis with TTL
	err = r.client.Set(ctx, key, string(data), r.config.TTL)
	if err != nil {
		return fmt.Errorf("failed to register service in Redis: %w", err)
	}

	r.logger.Infof("Registered service: %s at %s:%d", info.Name, info.Address, info.Port)

	// Notify watchers
	r.notifyWatchers(info.Name, info)

	return nil
}

// Deregister removes a service from the Redis registry
func (r *RedisRegistry) Deregister(ctx context.Context, name string) error {
	key := r.serviceKey(name)

	err := r.client.Del(ctx, key)
	if err != nil {
		return fmt.Errorf("failed to deregister service from Redis: %w", err)
	}

	r.logger.Infof("Deregistered service: %s", name)

	// Notify watchers with empty service info to indicate deregistration
	r.notifyWatchers(name, rpc.ServiceInfo{Name: name})

	return nil
}

// GetService retrieves a specific service from the registry
func (r *RedisRegistry) GetService(ctx context.Context, name string) (*rpc.ServiceInfo, error) {
	key := r.serviceKey(name)

	value, exists, err := r.client.GetWithExists(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("failed to get service from Redis: %w", err)
	}

	if !exists {
		return nil, fmt.Errorf("service not found: %s", name)
	}

	var info rpc.ServiceInfo
	err = json.Unmarshal([]byte(value), &info)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal service info: %w", err)
	}

	return &info, nil
}

// ListServices retrieves all registered services
func (r *RedisRegistry) ListServices(ctx context.Context) ([]rpc.ServiceInfo, error) {
	pattern := r.config.KeyPrefix + "*"

	// Use SCAN to get all keys matching the pattern
	var services []rpc.ServiceInfo
	var cursor uint64

	for {
		// Note: This is a simplified implementation. In production, you might want to use
		// Redis SCAN command directly for better performance with large datasets
		keys, err := r.scanKeys(ctx, pattern, cursor, 100)
		if err != nil {
			return nil, fmt.Errorf("failed to scan Redis keys: %w", err)
		}

		for _, key := range keys {
			value, exists, err := r.client.GetWithExists(ctx, key)
			if err != nil {
				r.logger.Warnf("Failed to get value for key %s: %v", key, err)
				continue
			}

			if !exists {
				continue
			}

			var info rpc.ServiceInfo
			err = json.Unmarshal([]byte(value), &info)
			if err != nil {
				r.logger.Warnf("Failed to unmarshal service info for key %s: %v", key, err)
				continue
			}

			services = append(services, info)
		}

		if cursor == 0 {
			break
		}
	}

	return services, nil
}

// Watch watches for changes to a specific service
func (r *RedisRegistry) Watch(ctx context.Context, name string) (<-chan rpc.ServiceInfo, error) {
	r.watcherMu.Lock()
	defer r.watcherMu.Unlock()

	// Create a new channel for this watcher
	ch := make(chan rpc.ServiceInfo, r.config.WatchBufferSize)

	if r.watchers[name] == nil {
		r.watchers[name] = make([]chan rpc.ServiceInfo, 0)
	}
	r.watchers[name] = append(r.watchers[name], ch)

	// Send initial state if service exists
	if info, err := r.GetService(ctx, name); err == nil {
		select {
		case ch <- *info:
		default:
			r.logger.Warnf("Failed to send initial state to watcher for service: %s", name)
		}
	}

	r.logger.Infof("Started watching service: %s", name)
	return ch, nil
}

// Close closes the registry and cleans up resources
func (r *RedisRegistry) Close() error {
	close(r.stopChan)

	// Close all watcher channels
	r.watcherMu.Lock()
	for serviceName, channels := range r.watchers {
		for _, ch := range channels {
			close(ch)
		}
		delete(r.watchers, serviceName)
	}
	r.watcherMu.Unlock()

	// Close Redis client
	return r.client.Close()
}

// GetProcessName returns the ProcessName for a given service name
func (r *RedisRegistry) GetProcessName(serviceName string) (logging.ProcessName, bool) {
	processName, exists := r.processMap[serviceName]
	return processName, exists
}

// GetServiceHealth retrieves the health status of a service
func (r *RedisRegistry) GetServiceHealth(ctx context.Context, name string) (*rpc.HealthStatus, error) {
	info, err := r.GetService(ctx, name)
	if err != nil {
		return nil, err
	}
	return &info.Health, nil
}

// UpdateServiceHealth updates the health status of a service
func (r *RedisRegistry) UpdateServiceHealth(ctx context.Context, name string, health rpc.HealthStatus) error {
	info, err := r.GetService(ctx, name)
	if err != nil {
		return err
	}

	info.Health = health
	info.LastSeen = time.Now()

	return r.Register(ctx, *info)
}

// serviceKey generates the Redis key for a service
func (r *RedisRegistry) serviceKey(name string) string {
	return r.config.KeyPrefix + name
}

// notifyWatchers notifies all watchers of a service about changes
func (r *RedisRegistry) notifyWatchers(serviceName string, info rpc.ServiceInfo) {
	r.watcherMu.RLock()
	defer r.watcherMu.RUnlock()

	channels, exists := r.watchers[serviceName]
	if !exists {
		return
	}

	for _, ch := range channels {
		select {
		case ch <- info:
		default:
			r.logger.Warnf("Failed to notify watcher for service: %s", serviceName)
		}
	}
}

// refreshLoop runs in the background to refresh service registrations
func (r *RedisRegistry) refreshLoop() {
	ticker := time.NewTicker(r.config.RefreshInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			r.refreshServices()
		case <-r.stopChan:
			return
		}
	}
}

// refreshServices refreshes all service registrations to extend their TTL
func (r *RedisRegistry) refreshServices() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	services, err := r.ListServices(ctx)
	if err != nil {
		r.logger.Errorf("Failed to list services for refresh: %v", err)
		return
	}

	for _, service := range services {
		// Only refresh if the service is still healthy
		if service.Health.Status == "healthy" {
			err := r.Register(ctx, service)
			if err != nil {
				r.logger.Warnf("Failed to refresh service %s: %v", service.Name, err)
			}
		}
	}
}

// scanKeys is a helper method to scan Redis keys (simplified implementation)
func (r *RedisRegistry) scanKeys(ctx context.Context, pattern string, cursor uint64, count int64) ([]string, error) {
	// This is a simplified implementation. In production, you would use the actual Redis SCAN command
	// For now, we'll return an empty slice as this would require direct Redis client access
	// The actual implementation would use r.client.Client().Scan(ctx, cursor, pattern, count)
	return []string{}, nil
}

// HealthCheck performs a health check on the registry
func (r *RedisRegistry) HealthCheck(ctx context.Context) error {
	return r.client.Ping(ctx)
}

// GetRegistryStats returns statistics about the registry
func (r *RedisRegistry) GetRegistryStats(ctx context.Context) (map[string]interface{}, error) {
	services, err := r.ListServices(ctx)
	if err != nil {
		return nil, err
	}

	stats := map[string]interface{}{
		"total_services": len(services),
		"services":       make(map[string]interface{}),
	}

	// Group services by health status
	healthCounts := make(map[string]int)
	for _, service := range services {
		healthCounts[service.Health.Status]++
		stats["services"].(map[string]interface{})[service.Name] = map[string]interface{}{
			"address":   service.Address,
			"port":      service.Port,
			"health":    service.Health.Status,
			"last_seen": service.LastSeen,
		}
	}

	stats["health_counts"] = healthCounts
	return stats, nil
}
