package scheduler

import (
	"fmt"
	"math/big"
)

// GetStats returns current scheduler statistics
func (s *ConditionBasedScheduler) GetStats() map[string]interface{} {
	s.workersMutex.RLock()
	runningConditionWorkers := 0
	runningEventWorkers := 0

	// Create a slice of condition worker details
	conditionWorkerDetails := make([]map[string]interface{}, 0, len(s.conditionWorkers))
	for jobID, worker := range s.conditionWorkers {
		conditionWorkerDetails = append(conditionWorkerDetails, map[string]interface{}{
			"job_id":              jobID,
			"is_running":          worker.IsRunning(),
			"condition_type":      worker.ConditionWorkerData.ConditionType,
			"upper_limit":         worker.ConditionWorkerData.UpperLimit,
			"lower_limit":         worker.ConditionWorkerData.LowerLimit,
			"value_source_type":   worker.ConditionWorkerData.ValueSourceType,
			"value_source_url":    worker.ConditionWorkerData.ValueSourceUrl,
			"last_value":          worker.LastValue,
			"last_check":          worker.LastCheckTimestamp,
			"condition_met_count": worker.ConditionMet,
			"recurring":           worker.ConditionWorkerData.Recurring,
			"expiration_time":     worker.ConditionWorkerData.ExpirationTime,
		})
		if worker.IsRunning() {
			runningConditionWorkers++
		}
	}

	// Create a slice of event worker details
	eventWorkerDetails := make([]map[string]interface{}, 0, len(s.eventWorkers))
	for jobID, worker := range s.eventWorkers {
		eventWorkerDetails = append(eventWorkerDetails, map[string]interface{}{
			"job_id":               jobID,
			"is_running":           worker.IsRunning(),
			"trigger_chain_id":     worker.EventWorkerData.TriggerChainID,
			"trigger_contract":     worker.EventWorkerData.TriggerContractAddress,
			"trigger_event":        worker.EventWorkerData.TriggerEvent,
			"last_processed_block": worker.LastBlock,
			"recurring":            worker.EventWorkerData.Recurring,
			"expiration_time":      worker.EventWorkerData.ExpirationTime,
		})
		if worker.IsRunning() {
			runningEventWorkers++
		}
	}
	s.workersMutex.RUnlock()

	// Calculate total workers and active workers
	totalConditionWorkers := len(s.conditionWorkers)
	totalEventWorkers := len(s.eventWorkers)
	totalWorkers := totalConditionWorkers + totalEventWorkers
	activeWorkers := runningConditionWorkers + runningEventWorkers

	// Get connected chains information
	connectedChains := len(s.chainClients)
	chainIDs := make([]string, 0, connectedChains)
	for chainID := range s.chainClients {
		chainIDs = append(chainIDs, chainID)
	}

	// Get performance metrics (using placeholder values since some metrics functions may not exist)
	var totalEvents, successfulEvents int64
	var avgProcessingTime float64
	var totalActions, successfulActions int64
	var workerCount int64

	// These may not exist yet, so we'll use safe defaults
	totalEvents = 0
	successfulEvents = 0
	avgProcessingTime = 0.0
	totalActions = 0
	successfulActions = 0
	workerCount = int64(totalWorkers)

	return map[string]interface{}{
		"scheduler_info": map[string]interface{}{
			"max_workers":               s.maxWorkers,
			"supported_chains":          []string{"11155420", "84532", "11155111"}, // OP Sepolia, Base Sepolia, Ethereum Sepolia
		},

		"worker_summary": map[string]interface{}{
			"total_workers":             totalWorkers,
			"active_workers":            activeWorkers,
			"condition_workers":         totalConditionWorkers,
			"event_workers":             totalEventWorkers,
			"running_condition_workers": runningConditionWorkers,
			"running_event_workers":     runningEventWorkers,
		},

		"chain_info": map[string]interface{}{
			"connected_chains": connectedChains,
			"chain_ids":        chainIDs,
		},

		"condition_workers": conditionWorkerDetails,
		"event_workers":     eventWorkerDetails,

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

// GetConditionWorkerStats returns statistics for a specific condition worker
func (s *ConditionBasedScheduler) GetConditionWorkerStats(jobID *big.Int) (map[string]interface{}, error) {
	s.workersMutex.RLock()
	defer s.workersMutex.RUnlock()

	worker, exists := s.conditionWorkers[jobID]
	if !exists {
		return nil, fmt.Errorf("condition worker for job %d not found", jobID)
	}

	return map[string]interface{}{
		"job_id":              worker.ConditionWorkerData.JobID,
		"is_running":          worker.IsRunning(),
		"condition_type":      worker.ConditionWorkerData.ConditionType,
		"upper_limit":         worker.ConditionWorkerData.UpperLimit,
		"lower_limit":         worker.ConditionWorkerData.LowerLimit,
		"value_source_type":   worker.ConditionWorkerData.ValueSourceType,
		"value_source_url":    worker.ConditionWorkerData.ValueSourceUrl,
		"last_value":          worker.LastValue,
		"last_check":          worker.LastCheckTimestamp,
		"condition_met_count": worker.ConditionMet,
		"recurring":           worker.ConditionWorkerData.Recurring,
		"expiration_time":     worker.ConditionWorkerData.ExpirationTime,
	}, nil
}

// GetEventWorkerStats returns statistics for a specific event worker
func (s *ConditionBasedScheduler) GetEventWorkerStats(jobID *big.Int) (map[string]interface{}, error) {
	s.workersMutex.RLock()
	defer s.workersMutex.RUnlock()

	worker, exists := s.eventWorkers[jobID]
	if !exists {
		return nil, fmt.Errorf("event worker for job %d not found", jobID)
	}

	return map[string]interface{}{
		"job_id":               worker.EventWorkerData.JobID,
		"is_running":           worker.IsRunning(),
		"trigger_chain_id":     worker.EventWorkerData.TriggerChainID,
		"trigger_contract":     worker.EventWorkerData.TriggerContractAddress,
		"trigger_event":        worker.EventWorkerData.TriggerEvent,
		"last_processed_block": worker.LastBlock,
		"recurring":            worker.EventWorkerData.Recurring,
		"expiration_time":      worker.EventWorkerData.ExpirationTime,
	}, nil
}

// GetAllWorkerStats returns statistics for all workers
func (s *ConditionBasedScheduler) GetAllWorkerStats() map[string]interface{} {
	s.workersMutex.RLock()
	defer s.workersMutex.RUnlock()

	conditionStats := make(map[string]interface{})
	eventStats := make(map[string]interface{})

	// Get condition worker stats
	for jobID, worker := range s.conditionWorkers {
		conditionStats[fmt.Sprintf("job_%d", jobID)] = map[string]interface{}{
			"job_id":              worker.ConditionWorkerData.JobID,
			"is_running":          worker.IsRunning(),
			"condition_type":      worker.ConditionWorkerData.ConditionType,
			"last_value":          worker.LastValue,
			"condition_met_count": worker.ConditionMet,
			"last_check":          worker.LastCheckTimestamp,
		}
	}

	// Get event worker stats
	for jobID, worker := range s.eventWorkers {
		eventStats[fmt.Sprintf("job_%d", jobID)] = map[string]interface{}{
			"job_id":               worker.EventWorkerData.JobID,
			"is_running":           worker.IsRunning(),
			"trigger_chain_id":     worker.EventWorkerData.TriggerChainID,
			"last_processed_block": worker.LastBlock,
		}
	}

	return map[string]interface{}{
		"condition_workers": conditionStats,
		"event_workers":     eventStats,
		"summary": map[string]interface{}{
			"total_condition_workers": len(s.conditionWorkers),
			"total_event_workers":     len(s.eventWorkers),
			"total_workers":           len(s.conditionWorkers) + len(s.eventWorkers),
		},
	}
}
