package execution

// import (
// 	"fmt"
// 	"sync"
// 	"time"

// 	"github.com/trigg3rX/triggerx-backend/pkg/docker/config"
// 	"github.com/trigg3rX/triggerx-backend/pkg/docker/types"
// 	"github.com/trigg3rX/triggerx-backend/pkg/logging"
// )

// type Alert struct {
// 	Type      string    `json:"type"`
// 	Message   string    `json:"message"`
// 	Severity  string    `json:"severity"`
// 	Timestamp time.Time `json:"timestamp"`
// }

// type HealthStatus struct {
// 	Status    string                    `json:"status"`
// 	Score     float64                   `json:"score"`
// 	LastCheck time.Time                 `json:"last_check"`
// 	Alerts    []Alert                   `json:"alerts"`
// 	Metrics   *types.PerformanceMetrics `json:"metrics"`
// }

// type ExecutionMonitor struct {
// 	pipeline         *ExecutionPipeline
// 	config           config.ExecutorConfig
// 	logger           logging.Logger
// 	mutex            sync.RWMutex
// 	alerts           []Alert
// 	metrics          *types.PerformanceMetrics
// 	monitoringTicker *time.Ticker
// 	stopMonitoring   chan struct{}
// }

// func NewExecutionMonitor(pipeline *ExecutionPipeline, cfg config.ExecutorConfig, logger logging.Logger) *ExecutionMonitor {
// 	monitor := &ExecutionMonitor{
// 		pipeline: pipeline,
// 		config:   cfg,
// 		logger:   logger,
// 		alerts:   make([]Alert, 0),
// 		metrics: &types.PerformanceMetrics{
// 			TotalExecutions:      0,
// 			SuccessfulExecutions: 0,
// 			FailedExecutions:     0,
// 			AverageExecutionTime: 0,
// 			MinExecutionTime:     0,
// 			MaxExecutionTime:     0,
// 			TotalCost:            0.0,
// 			AverageCost:          0.0,
// 			LastExecution:        time.Time{},
// 		},
// 		stopMonitoring: make(chan struct{}),
// 	}

// 	// Start monitoring routine
// 	monitor.startMonitoring()

// 	return monitor
// }

// func (em *ExecutionMonitor) startMonitoring() {
// 	// Use a default interval since Monitoring config doesn't exist
// 	interval := 30 * time.Second
// 	em.monitoringTicker = time.NewTicker(interval)

// 	go func() {
// 		for {
// 			select {
// 			case <-em.monitoringTicker.C:
// 				em.performHealthCheck()
// 			case <-em.stopMonitoring:
// 				em.monitoringTicker.Stop()
// 				return
// 			}
// 		}
// 	}()
// }

// func (em *ExecutionMonitor) performHealthCheck() {
// 	em.logger.Debugf("Performing health check")

// 	// Check active executions
// 	activeExecutions := em.pipeline.GetActiveExecutions()

// 	// Check for stuck executions (default max time: 5 minutes)
// 	maxExecutionTime := 5 * time.Minute
// 	for _, exec := range activeExecutions {
// 		duration := time.Since(exec.StartedAt)
// 		if duration > maxExecutionTime {
// 			em.createAlert("execution_timeout", fmt.Sprintf("Execution %s has exceeded max time: %v", exec.TraceID, duration), "warning")
// 		}
// 	}

// 	// Check pipeline statistics
// 	stats := em.pipeline.GetStats()

// 	// Check success rate (default min: 80%)
// 	minSuccessRate := 0.8
// 	if stats.TotalExecutions > 0 {
// 		successRate := float64(stats.SuccessfulExecutions) / float64(stats.TotalExecutions)
// 		if successRate < minSuccessRate {
// 			em.createAlert("low_success_rate", fmt.Sprintf("Success rate is low: %.2f%%", successRate*100), "error")
// 		}
// 	}

// 	// Check average execution time (default max: 2 minutes)
// 	maxAverageTime := 2 * time.Minute
// 	if stats.AverageExecutionTime > maxAverageTime {
// 		em.createAlert("high_execution_time", fmt.Sprintf("Average execution time is high: %v", stats.AverageExecutionTime), "warning")
// 	}

// 	// Update metrics
// 	em.updateMetrics(stats)
// }

// func (em *ExecutionMonitor) createAlert(alertType, message, severity string) {
// 	alert := Alert{
// 		Type:      alertType,
// 		Message:   message,
// 		Severity:  severity,
// 		Timestamp: time.Now(),
// 	}

// 	em.mutex.Lock()
// 	em.alerts = append(em.alerts, alert)

// 	// Keep only recent alerts (max 100)
// 	maxAlerts := 100
// 	if len(em.alerts) > maxAlerts {
// 		em.alerts = em.alerts[len(em.alerts)-maxAlerts:]
// 	}
// 	em.mutex.Unlock()

// 	em.logger.Warnf("Alert: %s - %s", alertType, message)
// }

// func (em *ExecutionMonitor) updateMetrics(stats *types.PerformanceMetrics) {
// 	em.mutex.Lock()
// 	defer em.mutex.Unlock()

// 	em.metrics = stats
// }

// func (em *ExecutionMonitor) GetHealthStatus() *HealthStatus {
// 	em.mutex.RLock()
// 	defer em.mutex.RUnlock()

// 	// Calculate overall health score
// 	healthScore := 100.0

// 	// Reduce score based on alerts
// 	criticalAlerts := 0
// 	warningAlerts := 0

// 	for _, alert := range em.alerts {
// 		if time.Since(alert.Timestamp) < time.Hour { // Only consider recent alerts
// 			switch alert.Severity {
// 			case "error":
// 				criticalAlerts++
// 			case "warning":
// 				warningAlerts++
// 			}
// 		}
// 	}

// 	healthScore -= float64(criticalAlerts) * 20 // Each critical alert reduces score by 20
// 	healthScore -= float64(warningAlerts) * 5   // Each warning alert reduces score by 5

// 	if healthScore < 0 {
// 		healthScore = 0
// 	}

// 	status := "healthy"
// 	if healthScore < 50 {
// 		status = "critical"
// 	} else if healthScore < 80 {
// 		status = "warning"
// 	}

// 	return &HealthStatus{
// 		Status:    status,
// 		Score:     healthScore,
// 		LastCheck: time.Now(),
// 		Alerts:    em.getRecentAlerts(),
// 		Metrics:   em.metrics,
// 	}
// }

// func (em *ExecutionMonitor) getRecentAlerts() []Alert {
// 	recentAlerts := make([]Alert, 0)
// 	cutoff := time.Now().Add(-time.Hour) // Last hour

// 	for _, alert := range em.alerts {
// 		if alert.Timestamp.After(cutoff) {
// 			recentAlerts = append(recentAlerts, alert)
// 		}
// 	}

// 	return recentAlerts
// }

// func (em *ExecutionMonitor) GetAlerts(severity string, limit int) []Alert {
// 	em.mutex.RLock()
// 	defer em.mutex.RUnlock()

// 	filteredAlerts := make([]Alert, 0)

// 	for _, alert := range em.alerts {
// 		if severity == "" || alert.Severity == severity {
// 			filteredAlerts = append(filteredAlerts, alert)
// 		}
// 	}

// 	// Sort by timestamp (newest first)
// 	for i := 0; i < len(filteredAlerts)-1; i++ {
// 		for j := i + 1; j < len(filteredAlerts); j++ {
// 			if filteredAlerts[i].Timestamp.Before(filteredAlerts[j].Timestamp) {
// 				filteredAlerts[i], filteredAlerts[j] = filteredAlerts[j], filteredAlerts[i]
// 			}
// 		}
// 	}

// 	// Apply limit
// 	if limit > 0 && len(filteredAlerts) > limit {
// 		filteredAlerts = filteredAlerts[:limit]
// 	}

// 	return filteredAlerts
// }

// func (em *ExecutionMonitor) ClearAlerts() {
// 	em.mutex.Lock()
// 	defer em.mutex.Unlock()

// 	em.alerts = make([]Alert, 0)
// 	em.logger.Info("All alerts cleared")
// }

// func (em *ExecutionMonitor) GetMetrics() *types.PerformanceMetrics {
// 	em.mutex.RLock()
// 	defer em.mutex.RUnlock()

// 	// Create a copy to avoid race conditions
// 	metrics := *em.metrics
// 	return &metrics
// }

// func (em *ExecutionMonitor) GetActiveExecutions() []*types.ExecutionContext {
// 	return em.pipeline.GetActiveExecutions()
// }

// func (em *ExecutionMonitor) GetExecutionByID(executionID string) (*types.ExecutionContext, bool) {
// 	return em.pipeline.GetExecutionByID(executionID)
// }

// func (em *ExecutionMonitor) CancelExecution(executionID string) error {
// 	return em.pipeline.CancelExecution(executionID)
// }

// func (em *ExecutionMonitor) Close() error {
// 	em.logger.Info("Closing execution monitor")

// 	if em.monitoringTicker != nil {
// 		close(em.stopMonitoring)
// 	}

// 	em.logger.Info("Execution monitor closed")
// 	return nil
// }
