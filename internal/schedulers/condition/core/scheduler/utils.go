package scheduler

import (
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/condition/core/scheduler/worker"
)

// Helper functions
func isValidConditionType(conditionType string) bool {
	validTypes := []string{
		worker.ConditionGreaterThan, worker.ConditionLessThan, worker.ConditionBetween,
		worker.ConditionEquals, worker.ConditionNotEquals, worker.ConditionGreaterEqual, worker.ConditionLessEqual,
	}
	for _, valid := range validTypes {
		if conditionType == valid {
			return true
		}
	}
	return false
}

func isValidSourceType(sourceType string) bool {
	validTypes := []string{worker.SourceTypeAPI, worker.SourceTypeOracle, worker.SourceTypeStatic, worker.SourceTypeWebSocket}
	for _, valid := range validTypes {
		if sourceType == valid {
			return true
		}
	}
	return false
}
