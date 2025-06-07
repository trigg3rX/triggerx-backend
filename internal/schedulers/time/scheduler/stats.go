package scheduler

// GetStats returns current scheduler statistics
func (s *TimeBasedScheduler) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"scheduler_signing_address":   s.schedulerSigningAddress,
		"active_tasks":  len(s.activeTasks),
		"queue_length": len(s.taskQueue),
		"performer_lock_ttl":  s.performerLockTTL,
		"task_cache_ttl":  s.taskCacheTTL,
		"duplicate_task_window":  s.duplicateTaskWindow,
		"polling_interval":  s.pollingInterval,
		"polling_look_ahead":  s.pollingLookAhead,
	}
}
