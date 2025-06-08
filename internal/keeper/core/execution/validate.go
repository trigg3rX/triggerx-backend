package execution

import (
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

func (e *TaskExecutor) validateTrigger(triggerData *types.TaskTriggerData) bool {
	e.logger.Infof("Validating trigger data: %+v", triggerData)

	return true
	
}