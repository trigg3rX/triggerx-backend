package scheduler

import (
	"fmt"

	"github.com/trigg3rX/triggerx-backend/internal/schedulers/event/metrics"
)

// GetStats returns current scheduler statistics
func (s *EventBasedScheduler) GetStats() map[string]interface{} {
	s.workersMutex.RLock()
	activeWorkers := len(s.workers)

	// Create a slice of worker details
	workerDetails := make([]map[string]interface{}, 0, activeWorkers)
	for jobID, worker := range s.workers {
		workerDetails = append(workerDetails, map[string]interface{}{
			"job_id":               jobID,
			"is_running":           worker.IsRunning(),
			"trigger_chain_id":     worker.Job.TriggerChainID,
			"trigger_contract":     worker.Job.TriggerContractAddress,
			"trigger_event":        worker.Job.TriggerEvent,
			"target_chain_id":      worker.Job.TaskTargetData.TargetChainID,
			"target_contract":      worker.Job.TaskTargetData.TargetContractAddress,
			"target_function":      worker.Job.TaskTargetData.TargetFunction,
			"last_processed_block": worker.LastBlock,
		})
	}
	s.workersMutex.RUnlock()

	s.clientsMutex.RLock()
	connectedChains := len(s.chainClients)
	chainIDs := make([]string, 0, connectedChains)
	for chainID := range s.chainClients {
		chainIDs = append(chainIDs, chainID)
	}
	s.clientsMutex.RUnlock()

	// Get event and action statistics
	totalEvents, successfulEvents, avgProcessingTime := metrics.GetEventStats()
	totalActions, successfulActions := metrics.GetActionStats()
	workerCount := metrics.GetWorkerCount()

	return map[string]interface{}{
		"manager_id":       s.managerID,
		"total_workers":    len(s.workers),
		"active_workers":   activeWorkers,
		"max_workers":      s.maxWorkers,
		"supported_chains": []string{"11155420", "84532", "11155111"}, // OP Sepolia, Base Sepolia, Ethereum Sepolia
		"connected_chains": connectedChains,
		"chain_ids":        chainIDs,
		"worker_details":   workerDetails,

		// Performance metrics
		"event_stats": map[string]interface{}{
			"total_events":      totalEvents,
			"successful_events": successfulEvents,
			"failed_events":     totalEvents - successfulEvents,
			"success_rate_percent": func() float64 {
				if totalEvents > 0 {
					return (float64(successfulEvents) / float64(totalEvents)) * 100
				}
				return 0
			}(),
			"average_processing_time_seconds": avgProcessingTime,
		},

		"action_stats": map[string]interface{}{
			"total_actions":      totalActions,
			"successful_actions": successfulActions,
			"failed_actions":     totalActions - successfulActions,
			"action_success_rate_percent": func() float64 {
				if totalActions > 0 {
					return (float64(successfulActions) / float64(totalActions)) * 100
				}
				return 0
			}(),
		},

		"system_stats": map[string]interface{}{
			"metric_tracked_workers": workerCount,
			"uptime_tracked":         true,
			"metrics_enabled":        true,
		},
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
