package execution

import (
	"fmt"

	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// TaskExecutor is the default implementation of TaskExecutor
type TaskExecutor struct {
	ethClient       *ethclient.Client
	etherscanAPIKey string
	argConverter    *ArgumentConverter
	logger logging.Logger
}

// NewTaskExecutor creates a new instance of TaskExecutor
func NewTaskExecutor(ethClient *ethclient.Client, etherscanAPIKey string, logger logging.Logger) *TaskExecutor {
	return &TaskExecutor{
		ethClient:       ethClient,
		etherscanAPIKey: etherscanAPIKey,
		argConverter:    &ArgumentConverter{},
		// validator:       validation.NewJobValidator(logger, ethClient),
		logger: logger,
	}
}

// Execute implements the TaskExecutor interface
func (e *TaskExecutor) Execute(job *types.HandleCreateJobData) (types.ActionData, error) {
	e.logger.Info("Executing task", "jobID", job.JobID)

	switch job.TaskDefinitionID {
	case 1, 3, 5:
		return e.executeActionWithStaticArgs(job)
	case 2, 4, 6:
		return e.executeActionWithDynamicArgs(job)
	default:
		return types.ActionData{}, fmt.Errorf("unsupported task definition id: %d", job.TaskDefinitionID)
	}
}