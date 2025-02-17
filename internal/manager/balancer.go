package manager

import (
	"sync"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
)

// SystemResources tracks CPU and memory usage against configured thresholds
type SystemResources struct {
	CpuUsage    float64
	MemUsage    float64
	MaxCpuUsage float64
	MaxMemUsage float64
}

// JobQueue maintains ordered lists of job IDs and their types with a maximum capacity
type JobQueue struct {
	jobIDs   []int64
	jobTypes []int
	maxSize  int
}

// LoadBalancer manages system resources and job queuing to prevent overload
type LoadBalancer struct {
	resources       SystemResources
	jobQueue        JobQueue
	jobQueueMutex   sync.RWMutex
	metricsInterval time.Duration
}

// NewLoadBalancer creates a load balancer with default resource thresholds
// CPU max 10%, Memory max 80%, Queue size 1000 jobs
func NewLoadBalancer() *LoadBalancer {
	lb := &LoadBalancer{
		resources: SystemResources{
			MaxCpuUsage: 10.0,
			MaxMemUsage: 80.0,
		},
		jobQueue: JobQueue{
			maxSize: 1000,
		},
		metricsInterval: 5 * time.Second,
	}
	return lb
}

// MonitorResources continuously samples system metrics on an interval
// Updates current CPU and memory usage percentages
func (lb *LoadBalancer) MonitorResources() {
	ticker := time.NewTicker(lb.metricsInterval)
	defer ticker.Stop()

	for range ticker.C {
		cpuPercent, err := cpu.Percent(time.Second, false)
		if err == nil && len(cpuPercent) > 0 {
			lb.resources.CpuUsage = cpuPercent[0]
		}

		memInfo, err := mem.VirtualMemory()
		if err == nil {
			lb.resources.MemUsage = memInfo.UsedPercent
		}
	}
}

// CheckResourceAvailability verifies system has capacity for new jobs
// Returns true if both CPU and memory are below configured thresholds
func (lb *LoadBalancer) CheckResourceAvailability() bool {
	return lb.resources.CpuUsage < lb.resources.MaxCpuUsage &&
		lb.resources.MemUsage < lb.resources.MaxMemUsage
}

// AddJobToQueue appends a new job to the queue when system is at capacity
// Thread-safe via mutex protection
func (lb *LoadBalancer) AddJobToQueue(jobID int64, jobType int) {
	lb.jobQueueMutex.Lock()
	defer lb.jobQueueMutex.Unlock()

	lb.jobQueue.jobIDs = append(lb.jobQueue.jobIDs, jobID)
	lb.jobQueue.jobTypes = append(lb.jobQueue.jobTypes, jobType)
}

// GetNextJob removes and returns the next job from the queue
// Returns the job and true if queue has items, nil and false if empty
func (lb *LoadBalancer) GetNextJob() (*JobQueue, bool) {
	lb.jobQueueMutex.Lock()
	defer lb.jobQueueMutex.Unlock()

	if len(lb.jobQueue.jobIDs) == 0 {
		return nil, false
	}

	jobID := lb.jobQueue.jobIDs[0]
	jobType := lb.jobQueue.jobTypes[0]
	lb.jobQueue.jobIDs = lb.jobQueue.jobIDs[1:]
	lb.jobQueue.jobTypes = lb.jobQueue.jobTypes[1:]

	return &JobQueue{
		jobIDs:   []int64{jobID},
		jobTypes: []int{jobType},
	}, true
}

// GetQueueStatus returns the current number of jobs in the queue
func (lb *LoadBalancer) GetQueueStatus() int {
	lb.jobQueueMutex.RLock()
	defer lb.jobQueueMutex.RUnlock()
	return len(lb.jobQueue.jobIDs)
}
