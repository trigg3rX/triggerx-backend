package workers

import (
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

type JobScheduler interface {
	RemoveJob(jobID int64)
	GetJobDetails(jobID int64) (*types.HandleCreateJobData, error)
	StartTimeBasedJob(jobData types.HandleCreateJobData) error
	StartEventBasedJob(jobData types.HandleCreateJobData) error
	StartConditionBasedJob(jobData types.HandleCreateJobData) error
	UpdateJobChainStatus(jobID int64, status string)
	Logger() logging.Logger
}
