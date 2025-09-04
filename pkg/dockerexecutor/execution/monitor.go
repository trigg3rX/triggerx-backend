package execution

import (
	"fmt"
	"sync"
	"time"

	"github.com/trigg3rX/triggerx-backend/pkg/dockerexecutor/config"
	"github.com/trigg3rX/triggerx-backend/pkg/dockerexecutor/types"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

type Alert struct {
	Type      string    `json:"type"`
	Message   string    `json:"message"`
	Severity  string    `json:"severity"`
	Timestamp time.Time `json:"timestamp"`
}

type HealthStatus struct {
	Status    string                    `json:"status"`
	Score     float64                   `json:"score"`
	LastCheck time.Time                 `json:"last_check"`
	Alerts    []Alert                   `json:"alerts"`
	Metrics   *types.PerformanceMetrics `json:"metrics"`
}

type executionMonitor struct {
	pipeline         *executionPipeline
	config           config.ConfigProviderInterface
	logger           logging.Logger
	mutex            sync.RWMutex
	alerts           []Alert
	metrics          *types.PerformanceMetrics
	monitoringTicker *time.Ticker
	stopMonitoring   chan struct{}
}

func newExecutionMonitor(pipeline *executionPipeline, cfg config.ConfigProviderInterface, logger logging.Logger) *executionMonitor {
	monitor := &executionMonitor{
		pipeline: pipeline,
		config:   cfg,
		logger:   logger,
		alerts:   make([]Alert, 0),
		metrics: &types.PerformanceMetrics{
			TotalExecutions:      0,
			SuccessfulExecutions: 0,
			FailedExecutions:     0,
			AverageExecutionTime: 0,
			MinExecutionTime:     0,
			MaxExecutionTime:     0,
			TotalCost:            0.0,
			AverageCost:          0.0,
			LastExecution:        time.Time{},
		},
		stopMonitoring: make(chan struct{}),
	}

	// Start monitoring routine
	monitor.startMonitoring()

	return monitor
}

func (em *executionMonitor) startMonitoring() {
	// Get monitoring configuration
	monitoringConfig := em.config.GetMonitoringConfig()
	interval := monitoringConfig.HealthCheckInterval
	em.monitoringTicker = time.NewTicker(interval)

	go func() {
		for {
			select {
			case <-em.monitoringTicker.C:
				em.performHealthCheck()
			case <-em.stopMonitoring:
				em.monitoringTicker.Stop()
				return
			}
		}
	}()
}

func (em *executionMonitor) performHealthCheck() {
	em.logger.Debugf("Performing health check")

	// Get monitoring configuration
	monitoringConfig := em.config.GetMonitoringConfig()

	// Check active executions
	activeExecutions := em.pipeline.getActiveExecutions()

	// Check for stuck executions
	maxExecutionTime := monitoringConfig.MaxExecutionTime
	for _, exec := range activeExecutions {
		duration := time.Since(exec.StartedAt)
		if duration > maxExecutionTime {
			em.createAlert("execution_timeout", fmt.Sprintf("Execution %s has exceeded max time: %v", exec.TraceID, duration), "warning")
		}
	}

	// Check pipeline statistics
	stats := em.pipeline.getStats()

	// Check success rate
	minSuccessRate := monitoringConfig.MinSuccessRate
	if stats.TotalExecutions > 0 {
		successRate := float64(stats.SuccessfulExecutions) / float64(stats.TotalExecutions)
		if successRate < minSuccessRate {
			em.createAlert("low_success_rate", fmt.Sprintf("Success rate is low: %.2f%%", successRate*100), "error")
		}
	}

	// Check average execution time
	maxAverageTime := monitoringConfig.MaxAverageExecutionTime
	if stats.AverageExecutionTime > maxAverageTime {
		em.createAlert("high_execution_time", fmt.Sprintf("Average execution time is high: %v", stats.AverageExecutionTime), "warning")
	}

	// Update metrics
	em.updateMetrics(stats)
}

func (em *executionMonitor) createAlert(alertType, message, severity string) {
	alert := Alert{
		Type:      alertType,
		Message:   message,
		Severity:  severity,
		Timestamp: time.Now(),
	}

	em.mutex.Lock()
	em.alerts = append(em.alerts, alert)

	// Keep only recent alerts based on configuration
	monitoringConfig := em.config.GetMonitoringConfig()
	maxAlerts := monitoringConfig.MaxAlerts
	if len(em.alerts) > maxAlerts {
		em.alerts = em.alerts[len(em.alerts)-maxAlerts:]
	}
	em.mutex.Unlock()

	em.logger.Warnf("Alert: %s - %s", alertType, message)
}

func (em *executionMonitor) updateMetrics(stats *types.PerformanceMetrics) {
	em.mutex.Lock()
	defer em.mutex.Unlock()

	em.metrics = stats
}

func (em *executionMonitor) getHealthStatus() *HealthStatus {
	em.mutex.RLock()
	defer em.mutex.RUnlock()

	// Get monitoring configuration
	monitoringConfig := em.config.GetMonitoringConfig()

	// Calculate overall health score
	healthScore := 100.0

	// Reduce score based on alerts
	criticalAlerts := 0
	warningAlerts := 0

	for _, alert := range em.alerts {
		if time.Since(alert.Timestamp) < monitoringConfig.AlertRetentionTime { // Only consider recent alerts
			switch alert.Severity {
			case "error":
				criticalAlerts++
			case "warning":
				warningAlerts++
			}
		}
	}

	healthScore -= float64(criticalAlerts) * monitoringConfig.CriticalAlertPenalty
	healthScore -= float64(warningAlerts) * monitoringConfig.WarningAlertPenalty

	if healthScore < 0 {
		healthScore = 0
	}

	status := "healthy"
	if healthScore < monitoringConfig.HealthScoreThresholds.Critical {
		status = "critical"
	} else if healthScore < monitoringConfig.HealthScoreThresholds.Warning {
		status = "warning"
	}

	return &HealthStatus{
		Status:    status,
		Score:     healthScore,
		LastCheck: time.Now(),
		Alerts:    em.getRecentAlerts(),
		Metrics:   em.metrics,
	}
}

func (em *executionMonitor) getRecentAlerts() []Alert {
	recentAlerts := make([]Alert, 0)
	monitoringConfig := em.config.GetMonitoringConfig()
	cutoff := time.Now().Add(-monitoringConfig.AlertRetentionTime)

	for _, alert := range em.alerts {
		if alert.Timestamp.After(cutoff) {
			recentAlerts = append(recentAlerts, alert)
		}
	}

	return recentAlerts
}

func (em *executionMonitor) getAlerts(severity string, limit int) []Alert {
	em.mutex.RLock()
	defer em.mutex.RUnlock()

	filteredAlerts := make([]Alert, 0)

	for _, alert := range em.alerts {
		if severity == "" || alert.Severity == severity {
			filteredAlerts = append(filteredAlerts, alert)
		}
	}

	// Sort by timestamp (newest first)
	for i := 0; i < len(filteredAlerts)-1; i++ {
		for j := i + 1; j < len(filteredAlerts); j++ {
			if filteredAlerts[i].Timestamp.Before(filteredAlerts[j].Timestamp) {
				filteredAlerts[i], filteredAlerts[j] = filteredAlerts[j], filteredAlerts[i]
			}
		}
	}

	// Apply limit
	if limit > 0 && len(filteredAlerts) > limit {
		filteredAlerts = filteredAlerts[:limit]
	}

	return filteredAlerts
}

func (em *executionMonitor) clearAlerts() {
	em.mutex.Lock()
	defer em.mutex.Unlock()

	em.alerts = make([]Alert, 0)
	em.logger.Info("All alerts cleared")
}

func (em *executionMonitor) getActiveExecutions() []*types.ExecutionContext {
	return em.pipeline.getActiveExecutions()
}

func (em *executionMonitor) cancelExecution(executionID string) error {
	return em.pipeline.cancelExecution(executionID)
}

func (em *executionMonitor) close() error {
	em.logger.Info("Closing execution monitor")

	if em.monitoringTicker != nil {
		close(em.stopMonitoring)
	}

	em.logger.Info("Execution monitor closed")
	return nil
}
