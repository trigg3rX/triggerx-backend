package loadbalancer

import (
	"context"
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

// updateTaskManagerMetrics updates metrics for all task managers
func (lb *LoadBalancer) updateTaskManagerMetrics(ctx context.Context) {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	for _, tm := range lb.taskManagers {
		// Update metrics from task manager
		// This is where you would implement the actual metrics collection
		// For now, we'll just update the last ping time
		tm.LastPing = time.Now()
	}
}
