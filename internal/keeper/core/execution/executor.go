package execution

import (
	"fmt"

	"github.com/trigg3rX/triggerx-backend/pkg/client/aggregator"
	"github.com/trigg3rX/triggerx-backend/pkg/docker"
	"github.com/trigg3rX/triggerx-backend/pkg/proof"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// TaskExecutor is the default implementation of TaskExecutor
type TaskExecutor struct {
	alchemyAPIKey   string
	etherscanAPIKey string
	codeExecutor    *docker.CodeExecutor
	argConverter    *ArgumentConverter
	aggregatorClient *aggregator.AggregatorClient
	logger          logging.Logger
}

// NewTaskExecutor creates a new instance of TaskExecutor
func NewTaskExecutor(alchemyAPIKey string, etherscanAPIKey string, codeExecutor *docker.CodeExecutor, aggregatorClient *aggregator.AggregatorClient, logger logging.Logger) *TaskExecutor {
	return &TaskExecutor{
		alchemyAPIKey:   alchemyAPIKey,
		etherscanAPIKey: etherscanAPIKey,
		codeExecutor:    codeExecutor,
		argConverter:    &ArgumentConverter{},
		aggregatorClient: aggregatorClient,
		logger: logger,
	}
}

func (e *TaskExecutor) ExecuteTask(taskTargetData *types.TaskTargetData, triggerData *types.TaskTriggerData) (bool, error) {
	e.logger.Info("Executing task", "jobID", taskTargetData.TaskID)

	isTriggerTrue := e.validateTrigger(triggerData)
	if !isTriggerTrue {
		return false, nil
	} else {
		var actionData types.PerformerActionData
		var err error
		switch taskTargetData.TaskDefinitionID {
		case 1, 3, 5:
			actionData, err = e.executeActionWithStaticArgs(taskTargetData)
			if err != nil {
				return false, err
			}
		case 2, 4, 6:
			actionData, err = e.executeActionWithDynamicArgs(taskTargetData)
			if err != nil {
				return false, err
			}
		default:
			return false, fmt.Errorf("unsupported task definition id: %d", taskTargetData.TaskDefinitionID)
		}

		e.logger.Infof("Action data: %+v", actionData)

		ipfsData := types.IPFSData{
			TargetData: taskTargetData,
			TriggerData: triggerData,
			ActionData: &actionData,
		}

		proofData, err := proof.GenerateProof(ipfsData)
		if err != nil {
			return false, err
		}
		e.logger.Infof("Proof data: %+v", proofData)

		// cid, err := e.generateIPFSData(taskTargetData, triggerData, actionData, proofData)
		// if err != nil {
		// 	return false, err
		// }
		// e.logger.Infof("IPFS data: %+v", cid)

		return true, nil
	}
}
