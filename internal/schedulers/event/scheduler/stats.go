package scheduler

import (
	"fmt"
)

// GetStats returns current scheduler statistics
func (s *EventBasedScheduler) GetStats() map[string]interface{} {
	s.workersMutex.RLock()
	defer s.workersMutex.RUnlock()

	s.clientsMutex.RLock()
	defer s.clientsMutex.RUnlock()

	runningWorkers := 0
	for _, worker := range s.workers {
		if worker.IsRunning() {
			runningWorkers++
		}
	}

	return map[string]interface{}{
		"manager_id":       s.managerID,
		"total_workers":    len(s.workers),
		"running_workers":  runningWorkers,
		"max_workers":      s.maxWorkers,
		"connected_chains": len(s.chainClients),
		"supported_chains": []string{"11155420", "84532", "11155111"}, // OP Sepolia, Base Sepolia, Ethereum Sepolia
		"cache_available":  s.cache != nil,
	}
}

// GetJobWorkerStats returns statistics for a specific job worker
func (s *EventBasedScheduler) GetJobWorkerStats(jobID int64) (map[string]interface{}, error) {
	s.workersMutex.RLock()
	defer s.workersMutex.RUnlock()

	worker, exists := s.workers[jobID]
	if !exists {
		return nil, fmt.Errorf("job %d not found", jobID)
	}

	return map[string]interface{}{
		"job_id":           worker.Job.JobID,
		"is_running":       worker.IsRunning(),
		"trigger_chain_id": worker.Job.TriggerChainID,
		"contract_address": worker.Job.TriggerContractAddress,
		"trigger_event":    worker.Job.TriggerEvent,
		"last_block":       worker.LastBlock,
		"manager_id":       worker.ManagerID,
	}, nil
}
