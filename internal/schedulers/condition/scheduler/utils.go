package scheduler

import (
	schedulerTypes "github.com/trigg3rX/triggerx-backend/internal/schedulers/condition/scheduler/types"
)

// Helper functions
func isValidConditionType(conditionType string) bool {
	validTypes := []string{
		schedulerTypes.ConditionGreaterThan, schedulerTypes.ConditionLessThan, schedulerTypes.ConditionBetween,
		schedulerTypes.ConditionEquals, schedulerTypes.ConditionNotEquals, schedulerTypes.ConditionGreaterEqual, schedulerTypes.ConditionLessEqual,
	}
	for _, valid := range validTypes {
		if conditionType == valid {
			return true
		}
	}
	return false
}

func isValidSourceType(sourceType string) bool {
	validTypes := []string{schedulerTypes.SourceTypeAPI, schedulerTypes.SourceTypeOracle, schedulerTypes.SourceTypeStatic}
	for _, valid := range validTypes {
		if sourceType == valid {
			return true
		}
	}
	return false
}
