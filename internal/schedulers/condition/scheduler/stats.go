package scheduler

import (
	"fmt"

	schedulerTypes "github.com/trigg3rX/triggerx-backend/internal/schedulers/condition/scheduler/types"
)

// GetStats returns current scheduler statistics
func (s *ConditionBasedScheduler) GetStats() map[string]interface{} {
	s.workersMutex.RLock()
	defer s.workersMutex.RUnlock()

	runningWorkers := 0
	for _, worker := range s.workers {
		if worker.IsRunning() {
			runningWorkers++
		}
	}

	return map[string]interface{}{
		"manager_id":        s.managerID,
		"total_workers":     len(s.workers),
		"running_workers":   runningWorkers,
		"max_workers":       s.maxWorkers,
		"supported_sources": []string{"api", "oracle", "static"},
		"supported_conditions": []string{
			"greater_than", "less_than", "between",
			"equals", "not_equals", "greater_equal", "less_equal",
		},
		"poll_interval_seconds": schedulerTypes.PollInterval.Seconds(),
	}
}

// GetJobWorkerStats returns statistics for a specific condition worker
func (s *ConditionBasedScheduler) GetJobWorkerStats(jobID int64) (map[string]interface{}, error) {
	s.workersMutex.RLock()
	defer s.workersMutex.RUnlock()

	worker, exists := s.workers[jobID]
	if !exists {
		return nil, fmt.Errorf("job %d not found", jobID)
	}

	return map[string]interface{}{
		"job_id":              worker.Job.JobID,
		"is_running":          worker.IsRunning(),
		"condition_type":      worker.Job.ConditionType,
		"upper_limit":         worker.Job.UpperLimit,
		"lower_limit":         worker.Job.LowerLimit,
		"value_source":        worker.Job.ValueSourceUrl,
		"last_value":          worker.LastValue,
		"last_check":          worker.LastCheck,
		"condition_met_count": worker.ConditionMet,
		"manager_id":          worker.ManagerID,
	}, nil
}
