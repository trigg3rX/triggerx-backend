package loadbalancer

import (
	"context"
	"sync"
	"time"

	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

type ManagerInfo struct {
	Address       string
	LastHeartbeat time.Time
	JobCount      int
	IsHealthy     bool
}

type LoadBalancer struct {
	managers map[string]*ManagerInfo
	mu       sync.RWMutex
	logger   logging.Logger
}

func NewLoadBalancer() *LoadBalancer {
	return &LoadBalancer{
		managers: make(map[string]*ManagerInfo),
		logger:   logging.GetLogger(logging.Development, logging.ManagerProcess),
	}
}

func (lb *LoadBalancer) AddManager(managerID, address string) error {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	lb.managers[managerID] = &ManagerInfo{
		Address:       address,
		LastHeartbeat: time.Now(),
		JobCount:      0,
		IsHealthy:     true,
	}

	lb.logger.Infof("Added manager %s with address %s", managerID, address)
	return nil
}

func (lb *LoadBalancer) RemoveManager(managerID string) error {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	delete(lb.managers, managerID)
	lb.logger.Infof("Removed manager %s", managerID)
	return nil
}

func (lb *LoadBalancer) UpdateManagerHeartbeat(managerID string) error {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	if manager, exists := lb.managers[managerID]; exists {
		manager.LastHeartbeat = time.Now()
		lb.logger.Debugf("Updated heartbeat for manager %s", managerID)
		return nil
	}

	lb.logger.Warnf("Manager %s not found for heartbeat update", managerID)
	return nil
}

func (lb *LoadBalancer) GetLeastLoadedManager() (string, error) {
	lb.mu.RLock()
	defer lb.mu.RUnlock()

	var leastLoadedID string
	minJobs := -1

	for id, manager := range lb.managers {
		if minJobs == -1 || manager.JobCount < minJobs {
			leastLoadedID = id
			minJobs = manager.JobCount
		}
	}

	if leastLoadedID == "" {
		lb.logger.Warn("No managers available for job assignment")
		return "", nil
	}

	lb.logger.Debugf("Selected manager %s with %d jobs", leastLoadedID, minJobs)
	return leastLoadedID, nil
}

func (lb *LoadBalancer) IncrementJobCount(managerID string) {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	if manager, exists := lb.managers[managerID]; exists {
		manager.JobCount++
		lb.logger.Debugf("Incremented job count for manager %s to %d", managerID, manager.JobCount)
	}
}

func (lb *LoadBalancer) DecrementJobCount(managerID string) {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	if manager, exists := lb.managers[managerID]; exists {
		manager.JobCount--
		lb.logger.Debugf("Decremented job count for manager %s to %d", managerID, manager.JobCount)
	}
}

func (lb *LoadBalancer) CleanupInactiveManagers(timeout time.Duration) {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	now := time.Now()
	for id, manager := range lb.managers {
		if now.Sub(manager.LastHeartbeat) > timeout {
			delete(lb.managers, id)
			lb.logger.Infof("Removed inactive manager %s", id)
		}
	}
}

func (lb *LoadBalancer) StartHealthChecks(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			lb.checkManagerHealth()
		}
	}
}

func (lb *LoadBalancer) checkManagerHealth() {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	now := time.Now()
	for id, manager := range lb.managers {
		if now.Sub(manager.LastHeartbeat) > 30*time.Second {
			manager.IsHealthy = false
			lb.logger.Warnf("Manager %s marked as unhealthy due to missed heartbeats", id)
		}
	}
}

func (lb *LoadBalancer) UpdateManagerLoad(managerID string, load int) {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	if manager, exists := lb.managers[managerID]; exists {
		manager.JobCount = load
		lb.logger.Debugf("Updated load for manager %s to %d", managerID, load)
	}
}

// CheckResourceAvailability verifies if the system has enough resources to handle new jobs
func (lb *LoadBalancer) CheckResourceAvailability() bool {
	lb.mu.RLock()
	defer lb.mu.RUnlock()

	// Check if any manager is available and healthy
	for _, manager := range lb.managers {
		if manager.IsHealthy {
			return true
		}
	}

	return false
}
