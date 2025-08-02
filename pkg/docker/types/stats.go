package types

import "time"

type CacheStats struct {
	HitCount      int64     `json:"hit_count"`
	MissCount     int64     `json:"miss_count"`
	HitRate       float64   `json:"hit_rate"`
	Size          int64     `json:"size"`
	MaxSize       int64     `json:"max_size"`
	ItemCount     int       `json:"item_count"`
	EvictionCount int64     `json:"eviction_count"`
	LastCleanup   time.Time `json:"last_cleanup"`
}

type PoolStats struct {
	Language          Language      `json:"language"`
	TotalContainers   int           `json:"total_containers"`
	ReadyContainers   int           `json:"ready_containers"`
	BusyContainers    int           `json:"busy_containers"`
	ErrorContainers   int           `json:"error_containers"`
	UtilizationRate   float64       `json:"utilization_rate"`
	AverageWaitTime   time.Duration `json:"average_wait_time"`
	MaxWaitTime       time.Duration `json:"max_wait_time"`
	ContainerLifetime time.Duration `json:"container_lifetime"`
	CreatedCount      int64         `json:"created_count"`
	DestroyedCount    int64         `json:"destroyed_count"`
	LastCleanup       time.Time     `json:"last_cleanup"`
}

type PerformanceMetrics struct {
	TotalExecutions      int64         `json:"total_executions"`
	SuccessfulExecutions int64         `json:"successful_executions"`
	FailedExecutions     int64         `json:"failed_executions"`
	AverageExecutionTime time.Duration `json:"average_execution_time"`
	MinExecutionTime     time.Duration `json:"min_execution_time"`
	MaxExecutionTime     time.Duration `json:"max_execution_time"`
	TotalCost            float64       `json:"total_cost"`
	AverageCost          float64       `json:"average_cost"`
	LastExecution        time.Time     `json:"last_execution"`
}
