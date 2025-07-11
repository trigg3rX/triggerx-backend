package scheduler

import "github.com/trigg3rX/triggerx-backend-imua/internal/schedulers/time/metrics"

// GetStats returns current scheduler statistics
func (s *TimeBasedScheduler) GetStats() map[string]interface{} {
	// Get task performance stats
	totalTasks, successfulTasks, avgTime := metrics.GetTaskStats()

	return map[string]interface{}{
		"scheduler_id": s.schedulerID,
		"active_tasks":              len(s.activeTasks),
		"performer_lock_ttl":        s.performerLockTTL,
		"task_cache_ttl":            s.taskCacheTTL,
		"duplicate_task_window":     s.duplicateTaskWindow,
		"polling_interval":          s.pollingInterval,
		"polling_look_ahead":        s.pollingLookAhead,

		// Performance metrics
		"task_stats": map[string]interface{}{
			"total_tasks":      totalTasks,
			"successful_tasks": successfulTasks,
			"failed_tasks":     totalTasks - successfulTasks,
			"success_rate_percent": func() float64 {
				if totalTasks > 0 {
					return (float64(successfulTasks) / float64(totalTasks)) * 100
				}
				return 0
			}(),
			"average_completion_time_seconds": avgTime,
		},
	}
}
