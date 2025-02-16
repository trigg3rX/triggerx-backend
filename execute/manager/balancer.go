package manager

import (
	"log"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
)

// SystemResources tracks system resource usage
type SystemResources struct {
	CPUUsage    float64
	MemoryUsage float64
	MaxCPU      float64
	MaxMemory   float64
}

// WaitingJob represents a job waiting in queue
type WaitingJob struct {
	Job           *Job
	EstimatedTime time.Time
}

// LoadBalancer handles system resources and job queuing
type LoadBalancer struct {
	resources       SystemResources
	waitingQueue    []WaitingJob
	waitingQueueMu  sync.RWMutex
	metricsInterval time.Duration
}

// NewLoadBalancer creates a new load balancer instance
func NewLoadBalancer() *LoadBalancer {
	lb := &LoadBalancer{
		resources: SystemResources{
			MaxCPU:    10.0, // 10% CPU threshold
			MaxMemory: 80.0, // 80% Memory threshold
		},
		waitingQueue:    make([]WaitingJob, 0),
		metricsInterval: 5 * time.Second,
	}
	return lb
}

// MonitorResources continuously monitors system resources
func (lb *LoadBalancer) MonitorResources() {
	ticker := time.NewTicker(lb.metricsInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			cpuPercent, err := cpu.Percent(time.Second, false)
			if err == nil && len(cpuPercent) > 0 {
				lb.resources.CPUUsage = cpuPercent[0]
			}

			memInfo, err := mem.VirtualMemory()
			if err == nil {
				lb.resources.MemoryUsage = memInfo.UsedPercent
			}
		}
	}
}

// CheckResourceAvailability verifies if system can handle new jobs
func (lb *LoadBalancer) CheckResourceAvailability() bool {
	return lb.resources.CPUUsage < lb.resources.MaxCPU &&
		lb.resources.MemoryUsage < lb.resources.MaxMemory
}

// CalculateEstimatedWaitTime estimates when resources might be available
func (lb *LoadBalancer) CalculateEstimatedWaitTime(jobs map[string]*Job) time.Time {
	var earliestCompletion time.Time
	now := time.Now()
	earliestCompletion = now.Add(30 * time.Second) // Default wait time

	for _, job := range jobs {
		if job.Status == "processing" {
			expectedCompletion := job.CreatedAt.Add(time.Duration(job.TimeFrame) * time.Second)
			if earliestCompletion.After(expectedCompletion) {
				earliestCompletion = expectedCompletion
			}
		}
	}

	return earliestCompletion
}

// AddToWaitingQueue adds a job to the waiting queue
func (lb *LoadBalancer) AddToWaitingQueue(job *Job, estimatedTime time.Time) {
	lb.waitingQueueMu.Lock()
	defer lb.waitingQueueMu.Unlock()

	lb.waitingQueue = append(lb.waitingQueue, WaitingJob{
		Job:           job,
		EstimatedTime: estimatedTime,
	})

	log.Printf("Job %s added to waiting queue. Estimated start time: %v",
		job.JobID, estimatedTime)
}

// GetNextWaitingJob returns and removes the next job from the waiting queue
func (lb *LoadBalancer) GetNextWaitingJob() (*WaitingJob, bool) {
	lb.waitingQueueMu.Lock()
	defer lb.waitingQueueMu.Unlock()

	if len(lb.waitingQueue) == 0 {
		return nil, false
	}

	nextJob := lb.waitingQueue[0]
	lb.waitingQueue = lb.waitingQueue[1:]
	return &nextJob, true
}

// SetResourceLimits updates the resource thresholds
func (lb *LoadBalancer) SetResourceLimits(cpuThreshold, memoryThreshold float64) {
	lb.resources.MaxCPU = cpuThreshold
	lb.resources.MaxMemory = memoryThreshold
}

// GetSystemMetrics returns current system metrics
func (lb *LoadBalancer) GetSystemMetrics() SystemResources {
	return lb.resources
}

// GetQueueStatus returns the current status of waiting queue
func (lb *LoadBalancer) GetQueueStatus() int {
	lb.waitingQueueMu.RLock()
	defer lb.waitingQueueMu.RUnlock()
	return len(lb.waitingQueue)
}