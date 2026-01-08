package metrics

import (
	"context"
	"time"

	"github.com/trigg3rX/triggerx-backend/pkg/observability"
)

var (
	startTime = time.Now()
	ctx       = context.Background()

	// UptimeSeconds tracks the service uptime in seconds
	uptimeSeconds observability.Gauge

	// Total task processing requests on keeper API server
	tasksReceivedTotal observability.Counter

	// Total task completed successfully, type: executed, validated
	tasksCompletedTotal *observability.CounterVec

	// Time taken for task completion, type: executed, validated
	taskDurationSeconds *observability.HistogramVec

	// Total tasks validated by type/id: 1-6
	tasksByDefinitionIDTotal *observability.CounterVec

	// Transaction metrics
	transactionsSentTotal *observability.CounterVec
	gasUsedTotal          *observability.CounterVec
	transactionFeesTotal  *observability.CounterVec

	// IPFS metrics
	ipfsDownloadSizeBytes observability.Counter
	ipfsUploadSizeBytes   observability.Counter

	// Health metrics
	failedHealthCheckinsTotal observability.Counter

	// System metrics
	memoryUsageBytes  observability.Gauge
	cpuUsagePercent   observability.Gauge
	goroutinesActive  observability.Gauge
	gcDurationSeconds observability.Gauge

	// Docker metrics
	dockerContainersCreatedTotal   *observability.CounterVec
	dockerContainerDurationSeconds *observability.GaugeVec

	// Aggregate metrics
	taskSuccessRate *observability.GaugeVec
	tasksPerDay     *observability.CounterVec

	// Total service restarts
	restartsTotal observability.Counter
)

// Exported variables for use in other packages
var (
	UptimeSeconds                  observability.Gauge
	TasksReceivedTotal             observability.Counter
	TasksCompletedTotal            *observability.CounterVec
	TaskDurationSeconds            *observability.HistogramVec
	TasksByDefinitionIDTotal       *observability.CounterVec
	TransactionsSentTotal          *observability.CounterVec
	GasUsedTotal                   *observability.CounterVec
	TransactionFeesTotal           *observability.CounterVec
	IPFSDownloadSizeBytes          observability.Counter
	IPFSUploadSizeBytes            observability.Counter
	FailedHealthCheckinsTotal      observability.Counter
	MemoryUsageBytes               observability.Gauge
	CPUUsagePercent                observability.Gauge
	GoroutinesActive               observability.Gauge
	GCDurationSeconds              observability.Gauge
	DockerContainersCreatedTotal   *observability.CounterVec
	DockerContainerDurationSeconds *observability.GaugeVec
	TaskSuccessRate                *observability.GaugeVec
	TasksPerDay                    *observability.CounterVec
	RestartsTotal                  observability.Counter
)

// InitializeMetrics initializes all metrics using the observability metrics instance
func InitializeMetrics(obsMetrics observability.Metrics) {
	// UptimeSeconds tracks the service uptime in seconds
	uptimeSeconds = obsMetrics.Gauge(
		"triggerx.keeper.uptime_seconds",
		observability.WithDescription("The uptime of the keeper service in seconds"),
		observability.WithUnit("s"),
	)
	UptimeSeconds = uptimeSeconds

	// Total task processing requests on keeper API server
	tasksReceivedTotal = obsMetrics.Counter(
		"triggerx.keeper.tasks_received_total",
		observability.WithDescription("Total tasks received"),
	)
	TasksReceivedTotal = tasksReceivedTotal

	// Total task completed successfully, type: executed, validated
	tasksCompletedTotal = observability.NewCounterVec(
		obsMetrics,
		"triggerx.keeper.tasks_completed_total",
		[]string{"type"},
		observability.WithDescription("Total tasks completed"),
	)
	TasksCompletedTotal = tasksCompletedTotal

	// Time taken for task completion, type: executed, validated
	taskDurationSeconds = observability.NewHistogramVec(
		obsMetrics,
		"triggerx.keeper.task_duration_seconds",
		[]string{"type"},
		observability.WithDescription("Time taken for task completion"),
		observability.WithUnit("s"),
	)
	TaskDurationSeconds = taskDurationSeconds

	// Total tasks validated by type/id: 1-6
	tasksByDefinitionIDTotal = observability.NewCounterVec(
		obsMetrics,
		"triggerx.keeper.tasks_by_definition_id_total",
		[]string{"id"},
		observability.WithDescription("Tasks validated by type/id"),
	)
	TasksByDefinitionIDTotal = tasksByDefinitionIDTotal

	// Transaction metrics
	transactionsSentTotal = observability.NewCounterVec(
		obsMetrics,
		"triggerx.keeper.transactions_sent_total",
		[]string{"chain_id", "status"},
		observability.WithDescription("Total transactions done for task executions"),
	)
	TransactionsSentTotal = transactionsSentTotal

	gasUsedTotal = observability.NewCounterVec(
		obsMetrics,
		"triggerx.keeper.gas_used_total",
		[]string{"chain_id"},
		observability.WithDescription("Total gas used in transactions"),
	)
	GasUsedTotal = gasUsedTotal

	transactionFeesTotal = observability.NewCounterVec(
		obsMetrics,
		"triggerx.keeper.transaction_fees_total",
		[]string{"chain_id"},
		observability.WithDescription("Total transaction fee incurred in transactions"),
	)
	TransactionFeesTotal = transactionFeesTotal

	// IPFS metrics
	ipfsDownloadSizeBytes = obsMetrics.Counter(
		"triggerx.keeper.ipfs_download_size_bytes",
		observability.WithDescription("Total IPFS content downloaded"),
		observability.WithUnit("By"),
	)
	IPFSDownloadSizeBytes = ipfsDownloadSizeBytes

	ipfsUploadSizeBytes = obsMetrics.Counter(
		"triggerx.keeper.ipfs_upload_size_bytes",
		observability.WithDescription("Total IPFS content uploaded"),
		observability.WithUnit("By"),
	)
	IPFSUploadSizeBytes = ipfsUploadSizeBytes

	// Health metrics
	failedHealthCheckinsTotal = obsMetrics.Counter(
		"triggerx.keeper.failed_health_checkins_total",
		observability.WithDescription("Total failed health checkins"),
	)
	FailedHealthCheckinsTotal = failedHealthCheckinsTotal

	// System metrics
	memoryUsageBytes = obsMetrics.Gauge(
		"triggerx.keeper.memory_usage_bytes",
		observability.WithDescription("Memory consumption"),
		observability.WithUnit("By"),
	)
	MemoryUsageBytes = memoryUsageBytes

	cpuUsagePercent = obsMetrics.Gauge(
		"triggerx.keeper.cpu_usage_percent",
		observability.WithDescription("CPU utilization"),
		observability.WithUnit("%"),
	)
	CPUUsagePercent = cpuUsagePercent

	goroutinesActive = obsMetrics.Gauge(
		"triggerx.keeper.goroutines_active",
		observability.WithDescription("Active Go routines"),
	)
	GoroutinesActive = goroutinesActive

	gcDurationSeconds = obsMetrics.Gauge(
		"triggerx.keeper.gc_duration_seconds",
		observability.WithDescription("Garbage collection time"),
		observability.WithUnit("s"),
	)
	GCDurationSeconds = gcDurationSeconds

	// Docker metrics
	dockerContainersCreatedTotal = observability.NewCounterVec(
		obsMetrics,
		"triggerx.keeper.docker_containers_created_total",
		[]string{"language"},
		observability.WithDescription("Docker container creation count"),
	)
	DockerContainersCreatedTotal = dockerContainersCreatedTotal

	dockerContainerDurationSeconds = observability.NewGaugeVec(
		obsMetrics,
		"triggerx.keeper.docker_container_duration_seconds",
		[]string{"language"},
		observability.WithDescription("Container execution time"),
		observability.WithUnit("s"),
	)
	DockerContainerDurationSeconds = dockerContainerDurationSeconds

	// Aggregate metrics
	taskSuccessRate = observability.NewGaugeVec(
		obsMetrics,
		"triggerx.keeper.task_success_rate",
		[]string{"type"},
		observability.WithDescription("Overall task success percentage"),
		observability.WithUnit("%"),
	)
	TaskSuccessRate = taskSuccessRate

	tasksPerDay = observability.NewCounterVec(
		obsMetrics,
		"triggerx.keeper.tasks_per_day",
		[]string{"type"},
		observability.WithDescription("Task throughput rate"),
	)
	TasksPerDay = tasksPerDay

	// Total service restarts
	restartsTotal = obsMetrics.Counter(
		"triggerx.keeper.restarts_total",
		observability.WithDescription("Service restart count"),
	)
	RestartsTotal = restartsTotal
}
