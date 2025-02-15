package manager

import (
	"sync"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
)

type SystemResources struct {
	CpuUsage    float64
	MemUsage    float64
	MaxCpuUsage float64
	MaxMemUsage float64
}

type JobQueue struct {
	jobIDs   []int64
	jobTypes []int
	maxSize  int
}

type LoadBalancer struct {
	resources       SystemResources
	jobQueue        JobQueue
	jobQueueMutex   sync.RWMutex
	metricsInterval time.Duration
}

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

func (lb *LoadBalancer) CheckResourceAvailability() bool {
	return lb.resources.CpuUsage < lb.resources.MaxCpuUsage &&
		lb.resources.MemUsage < lb.resources.MaxMemUsage
}

func (lb *LoadBalancer) AddJobToQueue(jobID int64, jobType int) {
	lb.jobQueueMutex.Lock()
	defer lb.jobQueueMutex.Unlock()

	lb.jobQueue.jobIDs = append(lb.jobQueue.jobIDs, jobID)
	lb.jobQueue.jobTypes = append(lb.jobQueue.jobTypes, jobType)
}

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

func (lb *LoadBalancer) GetQueueStatus() int {
	lb.jobQueueMutex.RLock()
	defer lb.jobQueueMutex.RUnlock()
	return len(lb.jobQueue.jobIDs)
}
