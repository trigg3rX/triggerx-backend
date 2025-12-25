package scheduler

import (
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// convertCustomJobToScheduleTimeTaskData converts a CustomJobData to ScheduleTimeTaskData format.
// This allows custom jobs to be processed using the same scheduling pipeline as time-based jobs.
func (s *TimeBasedScheduler) convertCustomJobToScheduleTimeTaskData(customJob *types.CustomJobData) types.ScheduleTimeTaskData {
	// For custom jobs, we don't have script storage repository in scheduler
	// The storage will be fetched by the executor when needed
	storage := make(map[string]string)

	return types.ScheduleTimeTaskData{
		TaskID:                 0, // Will be assigned during task creation
		TaskDefinitionID:       7, // Custom job task definition ID
		LastExecutedAt:         customJob.LastExecutedAt,
		ExpirationTime:         customJob.ExpirationTime,
		NextExecutionTimestamp: customJob.NextExecutionTime,
		ScheduleType:           "interval",
		TimeInterval:           customJob.TimeInterval,
		CronExpression:         "",
		SpecificSchedule:       "",
		TaskTargetData: types.TaskTargetData{
			JobID:                     customJob.JobID,
			TaskID:                    0, // Will be assigned later
			TaskDefinitionID:          7,
			TargetChainID:             customJob.TargetChainID, // Will be filled by script output
			TargetContractAddress:     "",                      // Will be filled by script output
			TargetFunction:            "",                      // Will be filled by script output
			ABI:                       "",
			ArgType:                   0,
			Arguments:                 []string{},
			DynamicArgumentsScriptUrl: customJob.CustomScriptUrl, // Use this field to pass script URL
			IsImua:                    false,
			// Custom script fields
			ScriptStorage:  storage, // Storage from database
			ScriptLanguage: customJob.ScriptLanguage,
		},
		IsImua: false,
	}
}
