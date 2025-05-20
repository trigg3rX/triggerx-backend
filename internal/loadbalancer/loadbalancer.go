package loadbalancer

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// TaskManager represents a task manager instance
type TaskManager struct {
	ID           string
	Address      string
	Status       string
	LastPing     time.Time
	CPUUsage     float64
	MemoryUsage  float64
	ActiveTasks  int
	MaxTasks     int
	Availability float64
}

// LoadBalancer represents the main load balancer service
type LoadBalancer struct {
	redisClient  *redis.Client
	taskManagers map[string]*TaskManager
	mu           sync.RWMutex
	isLeader     bool
	leaderKey    string
	leaderTTL    time.Duration
	healthCheck  *HealthChecker
	selector     TaskManagerSelector
}

// TaskManagerSelector interface for different selection strategies
type TaskManagerSelector interface {
	SelectTaskManager(taskManagers map[string]*TaskManager) (*TaskManager, error)
}

// NewLoadBalancer creates a new load balancer instance
func NewLoadBalancer(redisAddr string) (*LoadBalancer, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	lb := &LoadBalancer{
		redisClient:  rdb,
		taskManagers: make(map[string]*TaskManager),
		leaderKey:    "loadbalancer:leader",
		leaderTTL:    30 * time.Second,
		healthCheck:  NewHealthChecker(),
		selector:     NewWeightedSelector(),
	}

	return lb, nil
}

// Start begins the load balancer operations
func (lb *LoadBalancer) Start(ctx context.Context) error {
	// Start leader election
	go lb.runLeaderElection(ctx)

	// Start health checks
	go lb.healthCheck.Start(ctx, lb.taskManagers)

	// Start metrics collection
	go lb.collectMetrics(ctx)

	return nil
}

// SelectTaskManager selects the best task manager based on the current strategy
func (lb *LoadBalancer) SelectTaskManager() (*TaskManager, error) {
	lb.mu.RLock()
	defer lb.mu.RUnlock()

	return lb.selector.SelectTaskManager(lb.taskManagers)
}

// runLeaderElection implements the leader election mechanism
func (lb *LoadBalancer) runLeaderElection(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			lb.tryAcquireLeadership(ctx)
		}
	}
}

// tryAcquireLeadership attempts to become the leader
func (lb *LoadBalancer) tryAcquireLeadership(ctx context.Context) {
	// Implementation of leader election using Redis
	// This is a simplified version - you might want to add more robust error handling
	success, err := lb.redisClient.SetNX(ctx, lb.leaderKey, "leader", lb.leaderTTL).Result()
	if err != nil {
		return
	}

	lb.mu.Lock()
	lb.isLeader = success
	lb.mu.Unlock()

	if success {
		// Start leader-specific tasks
		go lb.refreshLeadership(ctx)
	}
}

// refreshLeadership periodically refreshes the leader's TTL
func (lb *LoadBalancer) refreshLeadership(ctx context.Context) {
	ticker := time.NewTicker(lb.leaderTTL / 2)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			lb.mu.RLock()
			if !lb.isLeader {
				lb.mu.RUnlock()
				return
			}
			lb.mu.RUnlock()

			// Refresh the leader key
			lb.redisClient.Expire(ctx, lb.leaderKey, lb.leaderTTL)
		}
	}
}

// collectMetrics periodically collects metrics from task managers
func (lb *LoadBalancer) collectMetrics(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			lb.updateTaskManagerMetrics(ctx)
		}
	}
}

// RegisterTaskManager registers a new task manager with the load balancer
func (lb *LoadBalancer) RegisterTaskManager(ctx context.Context, tm *TaskManager) error {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	// Store task manager info in Redis for persistence
	tmKey := fmt.Sprintf("taskmanager:%s", tm.ID)
	tmData, err := json.Marshal(tm)
	if err != nil {
		return fmt.Errorf("failed to marshal task manager data: %v", err)
	}

	if err := lb.redisClient.Set(ctx, tmKey, tmData, 0).Err(); err != nil {
		return fmt.Errorf("failed to store task manager in Redis: %v", err)
	}

	// Add to local map
	lb.taskManagers[tm.ID] = tm
	return nil
}

// UnregisterTaskManager removes a task manager from the load balancer
func (lb *LoadBalancer) UnregisterTaskManager(ctx context.Context, tmID string) error {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	// Remove from Redis
	tmKey := fmt.Sprintf("taskmanager:%s", tmID)
	if err := lb.redisClient.Del(ctx, tmKey).Err(); err != nil {
		return fmt.Errorf("failed to remove task manager from Redis: %v", err)
	}

	// Remove from local map
	delete(lb.taskManagers, tmID)
	return nil
}

// updateTaskManagerMetrics updates metrics for all task managers
func (lb *LoadBalancer) updateTaskManagerMetrics(ctx context.Context) {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	for _, tm := range lb.taskManagers {
		// Make HTTP request to task manager's metrics endpoint
		metricsURL := fmt.Sprintf("http://%s/metrics", tm.Address)
		resp, err := http.Get(metricsURL)
		if err != nil {
			tm.Status = "unhealthy"
			tm.Availability = 0
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			tm.Status = "unhealthy"
			tm.Availability = 0
			continue
		}

		// Parse metrics response
		var metrics struct {
			CPUUsage    float64 `json:"cpu_usage"`
			MemoryUsage float64 `json:"memory_usage"`
			ActiveTasks int     `json:"active_tasks"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&metrics); err != nil {
			tm.Status = "unhealthy"
			tm.Availability = 0
			continue
		}

		// Update task manager metrics
		tm.CPUUsage = metrics.CPUUsage
		tm.MemoryUsage = metrics.MemoryUsage
		tm.ActiveTasks = metrics.ActiveTasks
		tm.Status = "healthy"
		tm.Availability = 1.0
		tm.LastPing = time.Now()

		// Store updated metrics in Redis
		tmKey := fmt.Sprintf("taskmanager:%s", tm.ID)
		tmData, err := json.Marshal(tm)
		if err != nil {
			continue
		}

		lb.redisClient.Set(ctx, tmKey, tmData, 0)
	}
}

// GetTaskManagerStats returns statistics about all task managers
func (lb *LoadBalancer) GetTaskManagerStats() map[string]interface{} {
	lb.mu.RLock()
	defer lb.mu.RUnlock()

	stats := make(map[string]interface{})
	stats["total_managers"] = len(lb.taskManagers)
	stats["healthy_managers"] = 0
	stats["unhealthy_managers"] = 0
	stats["total_active_tasks"] = 0

	for _, tm := range lb.taskManagers {
		if tm.Status == "healthy" {
			stats["healthy_managers"] = stats["healthy_managers"].(int) + 1
		} else {
			stats["unhealthy_managers"] = stats["unhealthy_managers"].(int) + 1
		}
		stats["total_active_tasks"] = stats["total_active_tasks"].(int) + tm.ActiveTasks
	}

	return stats
}
