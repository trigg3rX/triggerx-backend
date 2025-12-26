package metrics

import (
	"context"
	"runtime"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
)

var (
	startTime = time.Now()
	ctx       = context.Background()

	// System Metrics
	uptimeSeconds     observability.Gauge
	memoryUsageBytes  observability.Gauge
	cpuUsagePercent   observability.Gauge
	goroutinesActive  observability.Gauge
	gcDurationSeconds observability.Gauge

	// Job and Worker Metrics
	jobsScheduled observability.Counter
	jobsCompleted *observability.CounterVec
	activeWorkers observability.Gauge

	// Condition Metrics
	conditionsByTypeTotal       *observability.CounterVec
	conditionsBySourceTotal     *observability.CounterVec
	conditionEvaluationDuration observability.Histogram

	// HTTP Metrics
	httpRequestsTotal     observability.Counter
	httpClientErrorsTotal *observability.CounterVec // Error codes

	// Database Metrics
	dbRequestsTotal       *observability.CounterVec // Read and Write
	dbRequestsErrorsTotal *observability.CounterVec // Read and Write

	// Error Metrics
	timeoutsTotal       *observability.CounterVec
	criticalErrorsTotal *observability.CounterVec

	// API and Value Metrics
	apiResponseStatusTotal  *observability.CounterVec
	valueParsingErrorsTotal *observability.CounterVec
	invalidValuesTotal      *observability.CounterVec

	// Internal tracking variables for performance calculations
	conditionStatsLock sync.RWMutex
	workerStartTimes   map[string]time.Time // job_id -> start_time
)

func init() {
	workerStartTimes = make(map[string]time.Time)
}

// InitializeMetrics initializes all metrics using the observability metrics instance.
// This must be called before using any metrics.
func InitializeMetrics(obsMetrics observability.Metrics) {
	// System metrics
	uptimeSeconds = obsMetrics.Gauge(
		"triggerx.condition_scheduler.uptime_seconds",
		observability.WithDescription("Time passed since Condition Scheduler started in seconds"),
		observability.WithUnit("s"),
	)

	memoryUsageBytes = obsMetrics.Gauge(
		"triggerx.condition_scheduler.memory_usage_bytes",
		observability.WithDescription("Memory consumption"),
		observability.WithUnit("By"),
	)

	cpuUsagePercent = obsMetrics.Gauge(
		"triggerx.condition_scheduler.cpu_usage_percent",
		observability.WithDescription("CPU utilization percentage"),
		observability.WithUnit("%"),
	)

	goroutinesActive = obsMetrics.Gauge(
		"triggerx.condition_scheduler.goroutines_active",
		observability.WithDescription("Number of active goroutines"),
	)

	gcDurationSeconds = obsMetrics.Gauge(
		"triggerx.condition_scheduler.gc_duration_seconds",
		observability.WithDescription("Garbage collection time"),
		observability.WithUnit("s"),
	)

	// Job and worker metrics
	jobsScheduled = obsMetrics.Counter(
		"triggerx.condition_scheduler.jobs_scheduled_total",
		observability.WithDescription("Total number of jobs scheduled"),
	)

	jobsCompleted = observability.NewCounterVec(
		obsMetrics,
		"triggerx.condition_scheduler.jobs_completed_total",
		[]string{"status"},
		observability.WithDescription("Total number of jobs completed"),
	)

	activeWorkers = obsMetrics.Gauge(
		"triggerx.condition_scheduler.active_workers",
		observability.WithDescription("Number of active job workers currently running"),
	)

	// Condition metrics
	conditionsByTypeTotal = observability.NewCounterVec(
		obsMetrics,
		"triggerx.condition_scheduler.conditions_by_type_total",
		[]string{"condition_type"},
		observability.WithDescription("Conditions monitored by type"),
	)

	conditionsBySourceTotal = observability.NewCounterVec(
		obsMetrics,
		"triggerx.condition_scheduler.conditions_by_source_total",
		[]string{"source_type"},
		observability.WithDescription("Conditions monitored by source type"),
	)

	conditionEvaluationDuration = obsMetrics.Histogram(
		"triggerx.condition_scheduler.condition_evaluation_duration_seconds",
		observability.WithDescription("Time taken to evaluate conditions"),
		observability.WithUnit("s"),
	)

	// HTTP metrics
	httpRequestsTotal = obsMetrics.Counter(
		"triggerx.condition_scheduler.http_requests_total",
		observability.WithDescription("HTTP API requests received"),
	)

	httpClientErrorsTotal = observability.NewCounterVec(
		obsMetrics,
		"triggerx.condition_scheduler.http_client_errors_total",
		[]string{"error_code"},
		observability.WithDescription("HTTP client errors"),
	)

	// Database metrics
	dbRequestsTotal = observability.NewCounterVec(
		obsMetrics,
		"triggerx.condition_scheduler.db_requests_total",
		[]string{"method", "endpoint", "status"},
		observability.WithDescription("Database client requests (read/write)"),
	)

	dbRequestsErrorsTotal = observability.NewCounterVec(
		obsMetrics,
		"triggerx.condition_scheduler.db_requests_errors_total",
		[]string{"method", "endpoint", "error_type"},
		observability.WithDescription("Database client request errors"),
	)

	// Error metrics
	timeoutsTotal = observability.NewCounterVec(
		obsMetrics,
		"triggerx.condition_scheduler.timeouts_total",
		[]string{"operation"},
		observability.WithDescription("Operation timeouts"),
	)

	criticalErrorsTotal = observability.NewCounterVec(
		obsMetrics,
		"triggerx.condition_scheduler.critical_errors_total",
		[]string{"error_type"},
		observability.WithDescription("Critical system errors"),
	)

	// API and value metrics
	apiResponseStatusTotal = observability.NewCounterVec(
		obsMetrics,
		"triggerx.condition_scheduler.api_response_status_total",
		[]string{"status_code"},
		observability.WithDescription("API response status codes"),
	)

	valueParsingErrorsTotal = observability.NewCounterVec(
		obsMetrics,
		"triggerx.condition_scheduler.value_parsing_errors_total",
		[]string{"source_type"},
		observability.WithDescription("Value parsing errors by source type"),
	)

	invalidValuesTotal = observability.NewCounterVec(
		obsMetrics,
		"triggerx.condition_scheduler.invalid_values_total",
		[]string{"source"},
		observability.WithDescription("Invalid/unparseable values received"),
	)
}

// Collector manages metrics collection for the condition scheduler.
type Collector struct {
	metrics observability.Metrics
}

// NewCollector creates a new metrics collector.
// The metrics are managed through the observability package, which handles
// Prometheus export internally if configured.
func NewCollector(obsMetrics observability.Metrics) *Collector {
	return &Collector{
		metrics: obsMetrics,
	}
}

// Start starts the background metrics collection goroutines.
// This begins collecting system metrics, performance metrics, and configuration metrics.
func (c *Collector) Start() {
	// Update metrics every 15 seconds
	go func() {
		ticker := time.NewTicker(15 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			if uptimeSeconds != nil {
				uptimeSeconds.Set(ctx, time.Since(startTime).Seconds())
			}
			collectSystemMetrics()
			collectConfigurationMetrics()
			collectPerformanceMetrics()
		}
	}()

	// Reset daily metrics every day at midnight
	go func() {
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()

		for range ticker.C {
			resetDailyMetrics()
		}
	}()
}

// collectSystemMetrics collects system resource metrics
func collectSystemMetrics() {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	// Update memory usage (current allocated bytes)
	if memoryUsageBytes != nil {
		memoryUsageBytes.Set(ctx, float64(memStats.Alloc))
	}

	// Update CPU usage
	cpuPercent, err := cpu.Percent(0, false)
	if err == nil && len(cpuPercent) > 0 {
		if cpuUsagePercent != nil {
			cpuUsagePercent.Set(ctx, cpuPercent[0])
		}
	} else {
		// Fallback to 0.0 if CPU monitoring fails
		if cpuUsagePercent != nil {
			cpuUsagePercent.Set(ctx, 0.0)
		}
	}

	// Update active goroutines count
	if goroutinesActive != nil {
		goroutinesActive.Set(ctx, float64(runtime.NumGoroutine()))
	}

	// Update garbage collection duration (total pause time in seconds)
	if gcDurationSeconds != nil {
		gcDurationSeconds.Set(ctx, float64(memStats.PauseTotalNs)/1e9)
	}
}

// collectConfigurationMetrics collects configuration-based metrics
func collectConfigurationMetrics() {
	// Configuration metrics can be added here if needed
	// For now, we don't have config-based metrics that need periodic updates
}

// collectPerformanceMetrics collects performance-related metrics
func collectPerformanceMetrics() {
	conditionStatsLock.RLock()
	defer conditionStatsLock.RUnlock()

	// Calculate average worker uptime if needed
	// This can be expanded based on requirements
}

// resetDailyMetrics resets metrics that should be reset daily
func resetDailyMetrics() {
	conditionStatsLock.Lock()
	defer conditionStatsLock.Unlock()

	// Reset tracking variables
	workerStartTimes = make(map[string]time.Time)
}

// Tracking functions

// TrackDBRequest tracks database request metrics
func TrackDBRequest(method, endpoint, status string) {
	if dbRequestsTotal != nil {
		dbRequestsTotal.WithLabelValues(method, endpoint, status).Inc(ctx)
	}
}

// TrackDBRequestError tracks database request errors
func TrackDBRequestError(method, endpoint, errorType string) {
	if dbRequestsErrorsTotal != nil {
		dbRequestsErrorsTotal.WithLabelValues(method, endpoint, errorType).Inc(ctx)
	}
}

// TrackDBConnectionError tracks database connection errors (legacy function name)
func TrackDBConnectionError() {
	// Track as a database request error
	TrackDBRequestError("connection", "database", "connection_error")
}

// TrackHTTPRequest tracks HTTP request metrics
func TrackHTTPRequest(method, endpoint, statusCode string) {
	if httpRequestsTotal != nil {
		httpRequestsTotal.Inc(ctx)
	}
}

// TrackHTTPClientConnectionError tracks HTTP client connection errors
func TrackHTTPClientConnectionError() {
	if httpClientErrorsTotal != nil {
		httpClientErrorsTotal.WithLabelValues("connection_error").Inc(ctx)
	}
}

// TrackJobScheduled tracks when a job is scheduled
func TrackJobScheduled() {
	if jobsScheduled != nil {
		jobsScheduled.Inc(ctx)
	}
}

// TrackJobCompleted tracks when a job completes
func TrackJobCompleted(status string) {
	if jobsCompleted != nil {
		jobsCompleted.WithLabelValues(status).Inc(ctx)
	}
}

// UpdateActiveWorkers updates the count of active workers
func UpdateActiveWorkers(count int) {
	if activeWorkers != nil {
		activeWorkers.Set(ctx, float64(count))
	}
}

// TrackWorkerStart tracks when a worker starts
func TrackWorkerStart(jobID string) {
	conditionStatsLock.Lock()
	defer conditionStatsLock.Unlock()
	workerStartTimes[jobID] = time.Now()
}

// TrackWorkerStop tracks when a worker stops
func TrackWorkerStop(jobID string) {
	conditionStatsLock.Lock()
	defer conditionStatsLock.Unlock()
	delete(workerStartTimes, jobID)
}

// TrackConditionByType tracks conditions by their type
func TrackConditionByType(conditionType string) {
	if conditionsByTypeTotal != nil {
		conditionsByTypeTotal.WithLabelValues(conditionType).Inc(ctx)
	}
}

// TrackConditionBySource tracks conditions by their source type
func TrackConditionBySource(sourceType string) {
	if conditionsBySourceTotal != nil {
		conditionsBySourceTotal.WithLabelValues(sourceType).Inc(ctx)
	}
}

// TrackConditionEvaluation tracks condition evaluation metrics
func TrackConditionEvaluation(duration time.Duration) {
	if conditionEvaluationDuration != nil {
		conditionEvaluationDuration.Record(ctx, duration.Seconds())
	}
}

// TrackTimeout tracks operation timeouts
func TrackTimeout(operation string) {
	if timeoutsTotal != nil {
		timeoutsTotal.WithLabelValues(operation).Inc(ctx)
	}
}

// TrackCriticalError tracks critical system errors
func TrackCriticalError(errorType string) {
	if criticalErrorsTotal != nil {
		criticalErrorsTotal.WithLabelValues(errorType).Inc(ctx)
	}
}

// TrackAPIResponse tracks API response status
func TrackAPIResponse(statusCode string) {
	if apiResponseStatusTotal != nil {
		apiResponseStatusTotal.WithLabelValues(statusCode).Inc(ctx)
	}
}

// TrackValueParsingError tracks value parsing errors
func TrackValueParsingError(sourceType string) {
	if valueParsingErrorsTotal != nil {
		valueParsingErrorsTotal.WithLabelValues(sourceType).Inc(ctx)
	}
}

// TrackInvalidValue tracks invalid values received
func TrackInvalidValue(source string) {
	if invalidValuesTotal != nil {
		invalidValuesTotal.WithLabelValues(source).Inc(ctx)
	}
}

// TrackConditionCheck tracks condition checks with duration and success
// This is a simplified version that tracks condition evaluation
func TrackConditionCheck(chainID string, duration time.Duration, success bool) {
	// Track condition evaluation duration
	TrackConditionEvaluation(duration)

	// Track success/failure through condition evaluation
	// Additional tracking can be added here if needed
}
