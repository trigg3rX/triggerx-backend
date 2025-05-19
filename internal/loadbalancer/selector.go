package loadbalancer

import (
	"errors"
	"math"
	"sort"
)

// WeightedSelector implements a weighted selection strategy
type WeightedSelector struct {
	weights map[string]float64
}

// NewWeightedSelector creates a new weighted selector
func NewWeightedSelector() *WeightedSelector {
	return &WeightedSelector{
		weights: make(map[string]float64),
	}
}

// SelectTaskManager implements the TaskManagerSelector interface
func (ws *WeightedSelector) SelectTaskManager(taskManagers map[string]*TaskManager) (*TaskManager, error) {
	if len(taskManagers) == 0 {
		return nil, errors.New("no task managers available")
	}

	// Filter out unhealthy task managers
	healthyManagers := make([]*TaskManager, 0)
	for _, tm := range taskManagers {
		if tm.Status == "healthy" {
			healthyManagers = append(healthyManagers, tm)
		}
	}

	if len(healthyManagers) == 0 {
		return nil, errors.New("no healthy task managers available")
	}

	// Calculate scores for each task manager
	scores := make([]struct {
		tm    *TaskManager
		score float64
	}, len(healthyManagers))

	for i, tm := range healthyManagers {
		// Calculate resource utilization score (lower is better)
		cpuScore := 1.0 - tm.CPUUsage
		memScore := 1.0 - tm.MemoryUsage

		// Calculate capacity score
		capacityScore := float64(tm.MaxTasks-tm.ActiveTasks) / float64(tm.MaxTasks)

		// Calculate availability score
		availabilityScore := tm.Availability

		// Combine scores with weights
		// You can adjust these weights based on your priorities
		totalScore := (cpuScore * 0.3) + (memScore * 0.3) + (capacityScore * 0.2) + (availabilityScore * 0.2)

		scores[i] = struct {
			tm    *TaskManager
			score float64
		}{
			tm:    tm,
			score: totalScore,
		}
	}

	// Sort by score in descending order
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].score > scores[j].score
	})

	// Return the task manager with the highest score
	return scores[0].tm, nil
}

// RoundRobinSelector implements a simple round-robin selection strategy
type RoundRobinSelector struct {
	lastIndex int
}

// NewRoundRobinSelector creates a new round-robin selector
func NewRoundRobinSelector() *RoundRobinSelector {
	return &RoundRobinSelector{
		lastIndex: -1,
	}
}

// SelectTaskManager implements the TaskManagerSelector interface
func (rr *RoundRobinSelector) SelectTaskManager(taskManagers map[string]*TaskManager) (*TaskManager, error) {
	if len(taskManagers) == 0 {
		return nil, errors.New("no task managers available")
	}

	// Convert map to slice for easier indexing
	managers := make([]*TaskManager, 0, len(taskManagers))
	for _, tm := range taskManagers {
		if tm.Status == "healthy" {
			managers = append(managers, tm)
		}
	}

	if len(managers) == 0 {
		return nil, errors.New("no healthy task managers available")
	}

	// Increment and wrap around
	rr.lastIndex = (rr.lastIndex + 1) % len(managers)
	return managers[rr.lastIndex], nil
}

// LeastConnectionsSelector implements a least connections selection strategy
type LeastConnectionsSelector struct{}

// NewLeastConnectionsSelector creates a new least connections selector
func NewLeastConnectionsSelector() *LeastConnectionsSelector {
	return &LeastConnectionsSelector{}
}

// SelectTaskManager implements the TaskManagerSelector interface
func (lc *LeastConnectionsSelector) SelectTaskManager(taskManagers map[string]*TaskManager) (*TaskManager, error) {
	if len(taskManagers) == 0 {
		return nil, errors.New("no task managers available")
	}

	var selected *TaskManager
	minConnections := math.MaxInt32

	for _, tm := range taskManagers {
		if tm.Status != "healthy" {
			continue
		}

		if tm.ActiveTasks < minConnections {
			minConnections = tm.ActiveTasks
			selected = tm
		}
	}

	if selected == nil {
		return nil, errors.New("no healthy task managers available")
	}

	return selected, nil
}
