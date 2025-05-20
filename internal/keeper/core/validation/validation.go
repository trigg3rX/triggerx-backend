package validation

import (
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

type TaskValidator interface {
	ValidateTask(task *types.TaskData) (bool, error)
	ValidateTimeBasedJob(job *types.HandleCreateJobData) (bool, error)
	ValidateEventBasedJob(job *types.HandleCreateJobData) (bool, error)
	ValidateConditionBasedJob(job *types.HandleCreateJobData) (bool, error)
}

type TaskValidatorImpl struct {
	logger logging.Logger

}

func NewTaskValidator(logger logging.Logger) TaskValidator {
	return &TaskValidatorImpl{
		logger: logger,
	}
}

func (v *TaskValidatorImpl) ValidateTask(task *types.TaskData) (bool, error) {
	return true, nil
}

func (v *TaskValidatorImpl) ValidateTimeBasedJob(job *types.HandleCreateJobData) (bool, error) {
	return true, nil
}

func (v *TaskValidatorImpl) ValidateEventBasedJob(job *types.HandleCreateJobData) (bool, error) {
	return true, nil
}

func (v *TaskValidatorImpl) ValidateConditionBasedJob(job *types.HandleCreateJobData) (bool, error) {
	return true, nil
}
