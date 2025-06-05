package scheduler

// GetStats returns current scheduler statistics
func (s *TimeBasedScheduler) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"manager_id":   s.managerID,
		"active_jobs":  len(s.activeJobs),
		"queue_length": len(s.jobQueue),
		"max_workers":  s.maxWorkers,
	}
}
